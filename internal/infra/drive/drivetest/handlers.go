package drivetest

import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"slices"
	"strconv"

	"github.com/MathTrail/mathtrail-standalone/internal/infra/drive"
)

// defaultPage is how many files a search answers with when it does not say.
const defaultPage = 100

// metadata is what a call that makes or changes a file says about it. A
// property set to null in a change is taken off the file.
type metadata struct {
	Name          string             `json:"name"`
	MimeType      string             `json:"mimeType"`
	Parents       []string           `json:"parents"`
	AppProperties map[string]*string `json:"appProperties"`
}

// list answers a search: the service's own files that match the query, the
// most recently modified first when it asks for that order, with the fields
// it asks for.
func (d *Drive) list(w http.ResponseWriter, r *http.Request, token string) {
	query := r.URL.Query()
	matches, err := parseQuery(query.Get("q"))
	if err != nil {
		d.malformed(w, "q %q: %v", query.Get("q"), err)
		return
	}
	order := query.Get("orderBy")
	if order != "" && order != "modifiedTime desc" {
		d.malformed(w, "orderBy %q: the stand-in orders by modifiedTime desc alone", order)
		return
	}
	if spaces := query.Get("spaces"); spaces != "" && spaces != "drive" {
		d.malformed(w, "spaces %q: the service's files are in drive", spaces)
		return
	}
	page, err := pageSize(query.Get("pageSize"))
	if err != nil {
		d.malformed(w, "pageSize: %v", err)
		return
	}
	fields, err := listSelection(query.Get("fields"))
	if err != nil {
		d.malformed(w, "fields: %v", err)
		return
	}

	d.mu.Lock()
	var found []*File
	for _, file := range d.drives[token] {
		if !file.Own && matches(file) {
			found = append(found, file.copied())
		}
	}
	d.mu.Unlock()

	if order != "" {
		slices.SortStableFunc(found, func(a, b *File) int { return b.ModifiedTime.Compare(a.ModifiedTime) })
	}
	listed := make([]map[string]any, 0, min(page, len(found)))
	for _, file := range found[:min(page, len(found))] {
		listed = append(listed, file.pick(fields))
	}
	answer := map[string]any{"files": listed}
	if len(found) > page {
		answer["nextPageToken"] = "the-next-page"
	}
	writeJSON(w, http.StatusOK, answer)
}

// get answers what a file is: the fields asked for.
func (d *Drive) get(w http.ResponseWriter, r *http.Request, token string) {
	fields, err := fileSelection(r.URL.Query().Get("fields"))
	if err != nil {
		d.malformed(w, "fields: %v", err)
		return
	}
	id := r.PathValue("id")
	d.mu.Lock()
	var answer map[string]any
	if file := d.visible(token, id); file != nil {
		answer = file.pick(fields)
	}
	d.mu.Unlock()

	if answer == nil {
		notFound(w, id)
		return
	}
	writeJSON(w, http.StatusOK, answer)
}

// download answers what a file holds.
func (d *Drive) download(w http.ResponseWriter, r *http.Request, token string) {
	id := r.PathValue("id")
	d.mu.Lock()
	var file *File
	if found := d.visible(token, id); found != nil {
		file = found.copied()
	}
	d.mu.Unlock()

	switch {
	case file == nil:
		notFound(w, id)
	case file.MimeType == drive.FolderType:
		refuse(w, http.StatusForbidden, "fileNotDownloadable", "Only files with binary content can be downloaded.")
	default:
		w.Header().Set("Content-Type", file.MimeType)
		w.WriteHeader(http.StatusOK)
		// The bytes the service uploaded, or a test put there, served back to the
		// service's own client as Drive serves a file: no browser reads them.
		_, _ = w.Write(file.Content) //nolint:gosec // a file served back to the client that stored it
	}
}

// createEmpty makes a file of metadata alone, as a folder is made.
func (d *Drive) createEmpty(w http.ResponseWriter, r *http.Request, token string) {
	fields, err := fileSelection(r.URL.Query().Get("fields"))
	if err != nil {
		d.malformed(w, "fields: %v", err)
		return
	}
	if !isJSON(r.Header.Get("Content-Type")) {
		d.malformed(w, "the metadata of a file is JSON, got %q", r.Header.Get("Content-Type"))
		return
	}
	meta, err := readMetadata(r.Body)
	if err != nil {
		d.malformed(w, "%v", err)
		return
	}
	d.make(w, token, meta, nil, fields)
}

// createFile makes a file of metadata and content, uploaded together.
func (d *Drive) createFile(w http.ResponseWriter, r *http.Request, token string) {
	fields, meta, content, err := readUpload(r)
	if err != nil {
		d.malformed(w, "%v", err)
		return
	}
	d.make(w, token, meta, content, fields)
}

// make keeps a new file of the service's own in the parent's Drive, in the
// folder the metadata names, and answers the fields asked for.
func (d *Drive) make(w http.ResponseWriter, token string, meta metadata, content []byte, fields []string) {
	if len(meta.Parents) > 1 {
		d.malformed(w, "a file can only have one parent folder")
		return
	}
	d.mu.Lock()
	for _, parent := range meta.Parents {
		if folder := d.visible(token, parent); folder == nil || folder.MimeType != drive.FolderType {
			d.mu.Unlock()
			notFound(w, parent)
			return
		}
	}
	file := &File{
		ID:            rand.Text(),
		Name:          meta.Name,
		MimeType:      meta.MimeType,
		Parents:       slices.Clone(meta.Parents),
		AppProperties: map[string]string{},
		Content:       content,
		ModifiedTime:  d.tick(),
	}
	for key, value := range meta.AppProperties {
		if value != nil {
			file.AppProperties[key] = *value
		}
	}
	d.drives[token] = append(d.drives[token], file)
	answer := file.pick(fields)
	d.mu.Unlock()

	writeJSON(w, http.StatusOK, answer)
}

// update replaces what a file holds and sets the properties the metadata
// names, keeping the others and taking off the ones set to null. The folder a
// file is in is not changed this way.
func (d *Drive) update(w http.ResponseWriter, r *http.Request, token string) {
	fields, meta, content, err := readUpload(r)
	if err != nil {
		d.malformed(w, "%v", err)
		return
	}
	if meta.Parents != nil {
		refuse(w, http.StatusForbidden, "fieldNotWritable", "The resource body includes fields which are not directly writable.")
		return
	}
	id := r.PathValue("id")
	d.mu.Lock()
	file := d.visible(token, id)
	if file == nil {
		d.mu.Unlock()
		notFound(w, id)
		return
	}
	if meta.Name != "" {
		file.Name = meta.Name
	}
	if meta.MimeType != "" {
		file.MimeType = meta.MimeType
	}
	if file.AppProperties == nil {
		file.AppProperties = map[string]string{}
	}
	for key, value := range meta.AppProperties {
		if value == nil {
			delete(file.AppProperties, key)
		} else {
			file.AppProperties[key] = *value
		}
	}
	file.Content = content
	file.ModifiedTime = d.tick()
	answer := file.pick(fields)
	d.mu.Unlock()

	writeJSON(w, http.StatusOK, answer)
}

// readUpload reads an upload the way Drive takes one: multipart/related of a
// known length, the metadata first as JSON, then the content as what it is,
// and nothing after them.
func readUpload(r *http.Request) (fields []string, meta metadata, content []byte, err error) {
	if kind := r.URL.Query().Get("uploadType"); kind != "multipart" {
		return nil, metadata{}, nil, fmt.Errorf("uploadType %q: the service uploads multipart", kind)
	}
	fields, err = fileSelection(r.URL.Query().Get("fields"))
	if err != nil {
		return nil, metadata{}, nil, fmt.Errorf("fields: %w", err)
	}
	if r.ContentLength < 0 {
		return nil, metadata{}, nil, errors.New("an upload says how long it is")
	}
	mediaType, params, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil || mediaType != "multipart/related" || params["boundary"] == "" {
		return nil, metadata{}, nil, fmt.Errorf("an upload is multipart/related with a boundary, got %q", r.Header.Get("Content-Type"))
	}
	parts := multipart.NewReader(r.Body, params["boundary"])

	first, err := parts.NextPart()
	if err != nil {
		return nil, metadata{}, nil, fmt.Errorf("an upload has no metadata: %w", err)
	}
	if !isJSON(first.Header.Get("Content-Type")) {
		return nil, metadata{}, nil, fmt.Errorf("the metadata comes first, as JSON, got %q", first.Header.Get("Content-Type"))
	}
	meta, err = readMetadata(first)
	if err != nil {
		return nil, metadata{}, nil, err
	}
	second, err := parts.NextPart()
	if err != nil {
		return nil, metadata{}, nil, fmt.Errorf("an upload has no content: %w", err)
	}
	if second.Header.Get("Content-Type") == "" {
		return nil, metadata{}, nil, errors.New("the content of an upload says what it is")
	}
	content, err = io.ReadAll(second)
	if err != nil {
		return nil, metadata{}, nil, fmt.Errorf("the content of an upload: %w", err)
	}
	if _, err = parts.NextPart(); !errors.Is(err, io.EOF) {
		return nil, metadata{}, nil, errors.New("an upload has two parts and no more")
	}
	return fields, meta, content, nil
}

// readMetadata reads the metadata of a file, and refuses a field the stand-in
// does not know: a name spelt wrong is a field Drive would drop without a word.
func readMetadata(body io.Reader) (metadata, error) {
	var meta metadata
	decoder := json.NewDecoder(body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&meta); err != nil {
		return metadata{}, fmt.Errorf("the metadata does not read: %w", err)
	}
	return meta, nil
}

// isJSON reports whether a content type is JSON's.
func isJSON(contentType string) bool {
	mediaType, _, err := mime.ParseMediaType(contentType)
	return err == nil && mediaType == "application/json"
}

// pageSize is how many files a search asks for, which Drive holds to 1–1000.
func pageSize(given string) (int, error) {
	if given == "" {
		return defaultPage, nil
	}
	size, err := strconv.Atoi(given)
	if err != nil || size < 1 || size > 1000 {
		return 0, fmt.Errorf("%q is not from 1 to 1000", given)
	}
	return size, nil
}

// notFound answers a call about a file the service may not see, the way Drive
// does: naming the file in its sentence, which a caller must never repeat.
func notFound(w http.ResponseWriter, id string) {
	refuse(w, http.StatusNotFound, "notFound", "File not found: "+id+".")
}

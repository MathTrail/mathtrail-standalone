package drive

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"net/url"
	"strings"
	"time"
)

const (
	// maxAnswer is the most of an answer about files that is read: a search
	// answers with a few, every other call with one.
	maxAnswer = 1 << 20
	// listed is how many files one search answers with. The service looks
	// for one file or one folder, and a handful more is an accident to
	// survive, not a page to read through.
	listed = "10"
	// jsonType is how the metadata of a file travels.
	jsonType = "application/json; charset=UTF-8"
)

// The fields each call asks Drive for. Whatever else a file carries stays
// with Drive.
const (
	listFields    = "files(id,name,parents,modifiedTime,webViewLink)"
	getFields     = "id,name"
	madeFields    = "id"
	changedFields = "id,trashed,appProperties"
)

// files makes each call as one request of its own, with its own deadline.
type files struct {
	client  *http.Client
	root    string
	timeout time.Duration
}

// NewFiles builds the calls, or refuses settings they could not be made with.
func NewFiles(settings *Settings) (Files, error) {
	return newFiles(settings)
}

// newFiles builds the calls every interface of the package is made with.
func newFiles(settings *Settings) (*files, error) {
	if settings == nil {
		return nil, fmt.Errorf("%w: the calls need their settings", ErrSettings)
	}
	if err := settings.validate(); err != nil {
		return nil, err
	}
	return &files{
		// A redirect is never followed: Drive's API answers where it is asked,
		// and a request that carries a parent's token goes nowhere else.
		client: &http.Client{CheckRedirect: func(*http.Request, []*http.Request) error {
			return http.ErrUseLastResponse
		}},
		root:    strings.TrimSuffix(settings.Root, "/"),
		timeout: settings.Timeout,
	}, nil
}

// request is one call before it is made: what it is called in an error, and
// what goes over the wire. The token and the deadline are added when it is
// sent.
type request struct {
	op          string
	method      string
	address     string
	contentType string
	body        []byte
}

func (f *files) List(ctx context.Context, token string, query Query) ([]File, error) {
	var found struct {
		Files []File `json:"files"`
	}
	err := f.send(ctx, token, &request{
		op:     "list",
		method: http.MethodGet,
		address: f.root + "/drive/v3/files?" + url.Values{
			"q":        {query.String()},
			"orderBy":  {"modifiedTime desc"},
			"pageSize": {listed},
			"spaces":   {"drive"},
			"fields":   {listFields},
		}.Encode(),
	}, func(body io.Reader) error {
		if err := decoded(&found)(body); err != nil {
			return err
		}
		for i := range found.Files {
			if found.Files[i].ID == "" {
				return errNoID
			}
		}
		return nil
	})
	return found.Files, err
}

func (f *files) Get(ctx context.Context, token, id string) (File, error) {
	if id == "" {
		return File{}, fmt.Errorf("drive: get: %w", errNoFile)
	}
	var file File
	err := f.send(ctx, token, &request{
		op:      "get",
		method:  http.MethodGet,
		address: f.file(id) + "?" + url.Values{"fields": {getFields}}.Encode(),
	}, decodedFile(&file))
	return file, err
}

func (f *files) Download(ctx context.Context, token, id string, limit int64) ([]byte, error) {
	if id == "" {
		return nil, fmt.Errorf("drive: download: %w", errNoFile)
	}
	var content []byte
	err := f.send(ctx, token, &request{
		op:      "download",
		method:  http.MethodGet,
		address: f.file(id) + "?alt=media",
	}, limited(limit, &content))
	return content, err
}

// limited reads content of at most limit bytes into what a call answers with,
// and refuses more with ErrTooLarge.
func limited(limit int64, into *[]byte) func(io.Reader) error {
	return func(body io.Reader) error {
		// One byte past the limit is enough to know the content is past it.
		read, err := io.ReadAll(io.LimitReader(body, limit+1))
		switch {
		case err != nil:
			return err
		case int64(len(read)) > limit:
			return ErrTooLarge
		}
		*into = read
		return nil
	}
}

func (f *files) Create(ctx context.Context, token string, meta *File, content []byte) (File, error) {
	// A file of metadata alone — a folder — goes to the address of the files,
	// and one with content to the address uploads go to, both in one request.
	made := request{
		op:      "create",
		method:  http.MethodPost,
		address: f.root + "/drive/v3/files?" + url.Values{"fields": {madeFields}}.Encode(),
	}
	var err error
	if content == nil {
		made.contentType = jsonType
		made.body, err = json.Marshal(meta)
	} else {
		made.address = f.root + "/upload/drive/v3/files?" + uploaded()
		err = made.carry(meta, content)
	}
	if err != nil {
		return File{}, fmt.Errorf("drive: create: %w", err)
	}
	var file File
	err = f.send(ctx, token, &made, decodedFile(&file))
	return file, err
}

func (f *files) Update(ctx context.Context, token, id string, change *Change) (File, error) {
	switch {
	case id == "":
		return File{}, fmt.Errorf("drive: update: %w", errNoFile)
	case change == nil || change.Meta == nil:
		return File{}, errors.New("drive: update: the change sets nothing")
	case change.Keep && change.Content == nil:
		return File{}, errors.New("drive: update: a change with no content makes no revision to keep")
	}
	query := url.Values{"fields": {changedFields}}
	updated := request{op: "update", method: http.MethodPatch}
	var err error
	if change.Content == nil {
		// Metadata alone goes to the address of the file, and makes no
		// revision.
		updated.address = f.file(id) + "?" + query.Encode()
		updated.contentType = jsonType
		updated.body, err = json.Marshal(change.Meta)
	} else {
		query.Set("uploadType", "multipart")
		if change.Keep {
			query.Set("keepRevisionForever", "true")
		}
		updated.address = f.root + "/upload/drive/v3/files/" + url.PathEscape(id) + "?" + query.Encode()
		err = updated.carry(change.Meta, change.Content)
	}
	if err != nil {
		return File{}, fmt.Errorf("drive: update: %w", err)
	}
	var file File
	err = f.send(ctx, token, &updated, decodedFile(&file))
	return file, err
}

// file is the address of one file.
func (f *files) file(id string) string {
	return f.root + "/drive/v3/files/" + url.PathEscape(id)
}

// uploaded is the query of an upload: the metadata and the content in one
// request, and the file's ID in the answer.
func uploaded() string {
	return url.Values{"uploadType": {"multipart"}, "fields": {madeFields}}.Encode()
}

// send makes one request with the parent's token, within the time a call has,
// and hands the answer to read when Drive took the request. What went wrong is
// told in the call's own words and Drive's refusal — never with the address,
// which names the file.
func (f *files) send(ctx context.Context, token string, call *request, read func(io.Reader) error) error {
	ctx, cancel := context.WithTimeout(ctx, f.timeout)
	defer cancel()

	var body io.Reader
	if call.body != nil {
		body = bytes.NewReader(call.body)
	}
	req, err := http.NewRequestWithContext(ctx, call.method, call.address, body)
	if err != nil {
		return fmt.Errorf("drive: %s: %w", call.op, withoutAddress(err))
	}
	req.Header.Set("Authorization", "Bearer "+token)
	if call.contentType != "" {
		req.Header.Set("Content-Type", call.contentType)
	}

	resp, err := f.client.Do(req)
	if err != nil {
		return fmt.Errorf("drive: %s: %w", call.op, unanswered(err))
	}
	defer func() {
		drain(resp.Body)
		_ = resp.Body.Close()
	}()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf("drive: %s: %w", call.op, RefusalOf(resp))
	}
	switch err := read(resp.Body); {
	case errors.Is(err, ErrTooLarge):
		return fmt.Errorf("drive: %s: %w", call.op, err)
	case err != nil:
		return fmt.Errorf("drive: %s: %w", call.op, unanswered(err))
	}
	return nil
}

// unanswered is a call Drive gave no whole answer to: it could not be reached,
// the connection failed on the way, or the answer broke off or did not read as
// one of Drive's. That is Drive not answering — unless the caller gave up on
// the call first, which is nobody's failure and stays the caller's own.
func unanswered(err error) error {
	err = withoutAddress(err)
	if errors.Is(err, context.Canceled) {
		return err
	}
	return fmt.Errorf("%w: %w", ErrUnavailable, err)
}

// drain reads what is left of an answer, as far as any answer about files
// runs: a connection whose answer was not read to its end cannot carry the
// next call, and a new one to Drive costs a handshake.
func drain(body io.Reader) {
	_, _ = io.Copy(io.Discard, io.LimitReader(body, maxAnswer))
}

// withoutAddress is what went wrong with a request, less the address it was
// sent to: the address names the file, and an error may be written where a
// person reads it.
func withoutAddress(err error) error {
	var failed *url.Error
	if errors.As(err, &failed) {
		return failed.Err
	}
	return err
}

// decoded reads an answer of Drive's into what a call answers with, as far as
// an answer about files ever runs.
func decoded(into any) func(io.Reader) error {
	return func(body io.Reader) error {
		return json.NewDecoder(io.LimitReader(body, maxAnswer)).Decode(into)
	}
}

// decodedFile reads an answer about one file, which says which file it is.
func decodedFile(into *File) func(io.Reader) error {
	return func(body io.Reader) error {
		if err := decoded(into)(body); err != nil {
			return err
		}
		if into.ID == "" {
			return errNoID
		}
		return nil
	}
}

var (
	// errNoFile refuses a call about a file that names none: sent as it
	// stands, it would reach the address of the search instead.
	errNoFile = errors.New("the call names no file")
	// errNoID is an answer about a file that does not say which file it is,
	// which no answer of Drive's API leaves out.
	errNoID = errors.New("the answer names no file")
)

// carry makes a file's metadata and its content the one body of an upload:
// the metadata first, as JSON, and then the content, as the kind of file it
// is.
func (call *request) carry(meta *File, content []byte) error {
	if meta.MimeType == "" {
		return errors.New("the file's type is not named, and its content needs one")
	}
	metadata, err := json.Marshal(meta)
	if err != nil {
		return err
	}
	var buffer bytes.Buffer
	parts := multipart.NewWriter(&buffer)
	for _, part := range []struct {
		contentType string
		data        []byte
	}{{jsonType, metadata}, {meta.MimeType, content}} {
		if err := writePart(parts, part.contentType, part.data); err != nil {
			return err
		}
	}
	if err := parts.Close(); err != nil {
		return err
	}
	call.body = buffer.Bytes()
	call.contentType = mime.FormatMediaType("multipart/related", map[string]string{"boundary": parts.Boundary()})
	return nil
}

// writePart adds one part of the given type to an upload's body.
func writePart(parts *multipart.Writer, contentType string, data []byte) error {
	part, err := parts.CreatePart(textproto.MIMEHeader{"Content-Type": {contentType}})
	if err != nil {
		return err
	}
	_, err = part.Write(data)
	return err
}

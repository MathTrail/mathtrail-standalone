package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

// Deposition is a draft or a published version of a record, as Zenodo's
// deposit API gives it: only what the steps read of it.
type Deposition struct {
	ID           int64        `json:"id"`
	ConceptRecID string       `json:"conceptrecid"`
	DOI          string       `json:"doi"`
	ConceptDOI   string       `json:"conceptdoi"`
	Metadata     depositMeta  `json:"metadata"`
	Links        depositLinks `json:"links"`
}

type depositMeta struct {
	Title   string `json:"title"`
	Version string `json:"version"`
	// PrereserveDOI is the DOI Zenodo keeps for a draft, as an object, or
	// false before one is asked for.
	PrereserveDOI json.RawMessage `json:"prereserve_doi"`
}

type depositLinks struct {
	Bucket      string `json:"bucket"`
	HTML        string `json:"html"`
	LatestDraft string `json:"latest_draft"`
}

// ReservedDOI is the DOI Zenodo keeps for the deposition, or nothing when it
// keeps none.
func (d *Deposition) ReservedDOI() string {
	var reserved struct {
		DOI string `json:"doi"`
	}
	if json.Unmarshal(d.Metadata.PrereserveDOI, &reserved) != nil {
		return ""
	}
	return reserved.DOI
}

// File is a file of a deposition.
type File struct {
	ID       string `json:"id"`
	Filename string `json:"filename"`
}

// Client speaks to Zenodo's deposit API as the owner of the token.
type Client struct {
	HTTP  *http.Client
	URL   string // the base: https://zenodo.org, or a sandbox's
	Token string
}

// jsonType is the type of what the API takes and gives.
const jsonType = "application/json"

// pageSize is the most depositions a page is asked to hold; Zenodo may give
// fewer, so the pages are read until one comes empty.
const pageSize = 100

// mostAnswer is the most bytes an answer is read to; one that runs longer is an
// error rather than JSON cut short.
const mostAnswer = 16 << 20

// mostPages is as many pages as an owner's depositions could fill. A server
// that gave the same page whatever was asked would otherwise be read forever.
const mostPages = 100

func (c Client) depositions() string { return c.URL + "/api/deposit/depositions" }

func (c Client) deposition(id int64) string {
	return c.depositions() + "/" + strconv.FormatInt(id, 10)
}

// List gives the owner's depositions in one state, draft or published, every
// version of every record among them.
func (c Client) List(ctx context.Context, status string) ([]Deposition, error) {
	var all []Deposition
	for page := 1; page <= mostPages; page++ {
		query := url.Values{
			"status": {status}, "all_versions": {"true"},
			"size": {strconv.Itoa(pageSize)}, "page": {strconv.Itoa(page)},
		}
		var batch []Deposition
		if err := c.call(ctx, http.MethodGet, c.depositions()+"?"+query.Encode(), nil, "", &batch); err != nil {
			return nil, err
		}
		if len(batch) == 0 {
			return all, nil
		}
		all = append(all, batch...)
	}
	return nil, fmt.Errorf("the %s depositions run past %d pages", status, mostPages)
}

// Create starts a record: its first draft, with nothing in it.
func (c Client) Create(ctx context.Context) (Deposition, error) {
	var d Deposition
	err := c.call(ctx, http.MethodPost, c.depositions(), strings.NewReader("{}"), jsonType, &d)
	return d, err
}

// Get reads a deposition from its link.
func (c Client) Get(ctx context.Context, link string) (Deposition, error) {
	var d Deposition
	err := c.call(ctx, http.MethodGet, link, nil, "", &d)
	return d, err
}

// NewVersion asks for a draft of the next version of a record, from its latest
// published version. The answer is that version, whose link latest_draft leads
// to the new draft.
func (c Client) NewVersion(ctx context.Context, latest int64) (Deposition, error) {
	var d Deposition
	err := c.call(ctx, http.MethodPost, c.deposition(latest)+"/actions/newversion", nil, "", &d)
	return d, err
}

// Files lists the files of a deposition.
func (c Client) Files(ctx context.Context, id int64) ([]File, error) {
	var files []File
	err := c.call(ctx, http.MethodGet, c.deposition(id)+"/files", nil, "", &files)
	return files, err
}

// DeleteFile takes a file out of a draft.
func (c Client) DeleteFile(ctx context.Context, id int64, file string) error {
	return c.call(ctx, http.MethodDelete, c.deposition(id)+"/files/"+url.PathEscape(file), nil, "", nil)
}

// Upload puts a file of a size into a draft's bucket under a name, and says
// the checksum Zenodo computed of what it received. The size goes with it: a
// bucket takes no upload of a length it is not told.
func (c Client) Upload(ctx context.Context, bucket, name string, body io.Reader, size int64) (string, error) {
	link := bucket + "/" + url.PathEscape(name)
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, link, body)
	if err != nil {
		return "", fmt.Errorf("PUT %s: %w", link, err)
	}
	req.ContentLength = size
	req.Header.Set("Content-Type", "application/octet-stream")
	var stored struct {
		Checksum string `json:"checksum"`
	}
	err = c.send(req, &stored)
	return stored.Checksum, err
}

// Update replaces what a draft says of itself.
func (c Client) Update(ctx context.Context, id int64, m *Metadata) (Deposition, error) {
	body, err := json.Marshal(struct {
		Metadata *Metadata `json:"metadata"`
	}{m})
	if err != nil {
		return Deposition{}, fmt.Errorf("update %d: %w", id, err)
	}
	var d Deposition
	err = c.call(ctx, http.MethodPut, c.deposition(id), bytes.NewReader(body), jsonType, &d)
	return d, err
}

// Publish publishes a draft. A published version cannot be taken back.
func (c Client) Publish(ctx context.Context, id int64) (Deposition, error) {
	var d Deposition
	err := c.call(ctx, http.MethodPost, c.deposition(id)+"/actions/publish", nil, "", &d)
	return d, err
}

// call makes one request and reads its answer into into, when there is one
// to read.
func (c Client) call(ctx context.Context, method, link string, body io.Reader, contentType string, into any) error {
	req, err := http.NewRequestWithContext(ctx, method, link, body)
	if err != nil {
		return fmt.Errorf("%s %s: %w", method, link, err)
	}
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	return c.send(req, into)
}

// send sends a request as the owner of the token and reads the answer into
// into. An answer other than a success is an error that quotes what Zenodo
// said, which never holds the token.
func (c Client) send(req *http.Request, into any) error {
	where := req.Method + " " + req.URL.String()
	// The links Zenodo answers with are followed, and none may lead elsewhere
	// with the token.
	if !strings.HasPrefix(req.URL.String(), c.URL+"/") {
		return fmt.Errorf("%s: the address is not under %s", where, c.URL)
	}
	req.Header.Set("Authorization", "Bearer "+c.Token)
	req.Header.Set("Accept", jsonType)
	resp, err := c.HTTP.Do(req) //nolint:gosec // the address is under Zenodo's own, held to it just above
	if err != nil {
		return fmt.Errorf("%s: %w", where, err)
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(io.LimitReader(resp.Body, mostAnswer+1))
	if err != nil {
		return fmt.Errorf("%s: %w", where, err)
	}
	if len(data) > mostAnswer {
		return fmt.Errorf("%s: the answer runs past %d bytes", where, mostAnswer)
	}
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return fmt.Errorf("%s: %s: %s", where, resp.Status, strings.TrimSpace(string(data)))
	}
	if into == nil || len(bytes.TrimSpace(data)) == 0 {
		return nil
	}
	if err := json.Unmarshal(data, into); err != nil {
		return fmt.Errorf("%s: the answer is not what the API gives: %w", where, err)
	}
	return nil
}

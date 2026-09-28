package drive

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

const (
	// revisionListFields are the fields a list of revisions asks for.
	revisionListFields = "revisions(id,modifiedTime,keepForever)"
	// revisionsListed is how many revisions one list answers with: every one
	// a file can have, since Drive keeps at most a hundred it may purge and
	// two hundred kept forever.
	revisionsListed = "1000"
	// keptFields is what the answer to keeping a revision is asked to say:
	// which revision it is.
	keptFields = "id"
)

// revisions makes each call to a file's history as one request of its own,
// the way the calls to the files are made.
type revisions struct {
	calls *files
}

// NewRevisions builds the calls to the history of files, or refuses settings
// they could not be made with.
func NewRevisions(settings *Settings) (Revisions, error) {
	calls, err := newFiles(settings)
	if err != nil {
		return nil, err
	}
	return &revisions{calls: calls}, nil
}

func (r *revisions) List(ctx context.Context, token, file string) ([]Revision, error) {
	if file == "" {
		return nil, fmt.Errorf("drive: list revisions: %w", errNoFile)
	}
	var listed struct {
		Revisions []Revision `json:"revisions"`
	}
	err := r.calls.send(ctx, token, &request{
		op:     "list revisions",
		method: http.MethodGet,
		address: r.calls.file(file) + "/revisions?" + url.Values{
			"pageSize": {revisionsListed},
			"fields":   {revisionListFields},
		}.Encode(),
	}, func(body io.Reader) error {
		if err := decoded(&listed)(body); err != nil {
			return err
		}
		for i := range listed.Revisions {
			if listed.Revisions[i].ID == "" {
				return errNoID
			}
		}
		return nil
	})
	return listed.Revisions, err
}

func (r *revisions) Download(ctx context.Context, token, file, revision string, limit int64) ([]byte, error) {
	if file == "" || revision == "" {
		return nil, fmt.Errorf("drive: download revision: %w", errNoFile)
	}
	var content []byte
	err := r.calls.send(ctx, token, &request{
		op:      "download revision",
		method:  http.MethodGet,
		address: r.address(file, revision) + "?alt=media",
	}, limited(limit, &content))
	return content, err
}

func (r *revisions) Keep(ctx context.Context, token, file, revision string) error {
	if file == "" || revision == "" {
		return fmt.Errorf("drive: keep revision: %w", errNoFile)
	}
	body, err := json.Marshal(map[string]bool{"keepForever": true})
	if err != nil {
		return fmt.Errorf("drive: keep revision: %w", err)
	}
	return r.calls.send(ctx, token, &request{
		op:          "keep revision",
		method:      http.MethodPatch,
		address:     r.address(file, revision) + "?" + url.Values{"fields": {keptFields}}.Encode(),
		contentType: jsonType,
		body:        body,
	}, func(answer io.Reader) error {
		var kept Revision
		if err := decoded(&kept)(answer); err != nil {
			return err
		}
		if kept.ID == "" {
			return errNoID
		}
		return nil
	})
}

func (r *revisions) Delete(ctx context.Context, token, file, revision string) error {
	if file == "" || revision == "" {
		return fmt.Errorf("drive: delete revision: %w", errNoFile)
	}
	return r.calls.send(ctx, token, &request{
		op:      "delete revision",
		method:  http.MethodDelete,
		address: r.address(file, revision),
	}, func(io.Reader) error { return nil })
}

// address is the address of one revision of a file.
func (r *revisions) address(file, revision string) string {
	return r.calls.file(file) + "/revisions/" + url.PathEscape(revision)
}

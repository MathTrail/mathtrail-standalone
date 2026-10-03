package main

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// maxBody bounds what one request may carry; a client's request for a reply,
// with its instructions and tools, is well under it.
const maxBody = 64 << 20

// refusal is the body of every answer, each sent with status 400: a client does
// not send an invalid request again, so it stops at the first answer instead of
// retrying. The body carries the fields each service's clients read an error
// from.
const refusal = `{"type":"error","error":{"type":"invalid_request_error","code":400,"status":"INVALID_ARGUMENT","message":"no model behind this endpoint"}}`

// Request is one request the client sent. Body is the request's body when it
// is JSON, and the body as a JSON string otherwise. Encoding names the
// Content-Encoding the body is still in, when the endpoint could not undo it:
// such a body cannot be read, and neither can what it offered.
type Request struct {
	Method   string          `json:"method"`
	Target   string          `json:"target"`
	Encoding string          `json:"encoding,omitempty"`
	Body     json.RawMessage `json:"body,omitempty"`
}

// recorder is the endpoint: it keeps every request, writes each to a file of
// its own for a person to read, and refuses them all.
type recorder struct {
	dir      string
	mu       sync.Mutex
	requests []Request
	err      error
}

func newRecorder(dir string) *recorder { return &recorder{dir: dir} }

func (r *recorder) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	body, err := io.ReadAll(http.MaxBytesReader(w, req.Body, maxBody))
	request := Request{Method: req.Method, Target: req.URL.RequestURI()}
	if err == nil {
		body, request.Encoding, err = decoded(body, req.Header.Get("Content-Encoding"))
		request.Body = asJSON(body)
	}

	r.mu.Lock()
	r.requests = append(r.requests, request)
	n := len(r.requests)
	if err == nil {
		err = writeRequest(filepath.Join(r.dir, fmt.Sprintf("request-%03d.json", n)), request)
	}
	if err != nil && r.err == nil {
		r.err = err
	}
	r.mu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(refusalStatus(req))
	_, _ = io.WriteString(w, refusal)
}

// decoded undoes the body's Content-Encoding where the standard library can:
// gzip. Any other encoding is kept, and named, so the body is judged
// unreadable rather than empty of tools.
func decoded(body []byte, encoding string) (plain []byte, kept string, err error) {
	switch strings.ToLower(strings.TrimSpace(encoding)) {
	case "", "identity":
		return body, "", nil
	case "gzip", "x-gzip":
		reader, err := gzip.NewReader(bytes.NewReader(body))
		if err != nil {
			return nil, "", fmt.Errorf("gzip body: %w", err)
		}
		plain, err = io.ReadAll(io.LimitReader(reader, maxBody))
		if err != nil {
			return nil, "", fmt.Errorf("gzip body: %w", err)
		}
		return plain, "", nil
	default:
		return body, encoding, nil
	}
}

// refusalStatus is 400 for every request but one that asks to become a
// WebSocket: that one is told the endpoint does not speak it, so the client
// sends its request as plain HTTP instead, with a body that can be read.
func refusalStatus(req *http.Request) int {
	if strings.EqualFold(req.Header.Get("Upgrade"), "websocket") {
		return http.StatusUpgradeRequired
	}
	return http.StatusBadRequest
}

// recorded is every request so far, in the order they came; it fails if one
// could not be read or kept.
func (r *recorder) recorded() ([]Request, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.err != nil {
		return nil, fmt.Errorf("record a request: %w", r.err)
	}
	return append([]Request(nil), r.requests...), nil
}

func asJSON(body []byte) json.RawMessage {
	switch {
	case len(body) == 0:
		return nil
	case json.Valid(body):
		return body
	default:
		text, _ := json.Marshal(string(body))
		return text
	}
}

func writeRequest(path string, request Request) error {
	data, err := json.MarshalIndent(request, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(data, '\n'), 0o600)
}

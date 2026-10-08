package main

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRecorderKeepsEveryRequestAndRefusesIt(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	endpoint := newRecorder(dir)
	server := httptest.NewServer(endpoint)
	defer server.Close()

	response, err := post(t, server.URL+"/v1/messages?beta=true", `{"messages":[]}`)
	if err != nil {
		t.Fatal(err)
	}
	answer, err := io.ReadAll(response.Body)
	_ = response.Body.Close()
	if err != nil {
		t.Fatal(err)
	}
	if response.StatusCode != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", response.StatusCode)
	}
	if string(answer) != refusal {
		t.Errorf("answer = %s, want the refusal", answer)
	}

	requests, err := endpoint.recorded()
	if err != nil {
		t.Fatalf("recorded: %v", err)
	}
	want := Request{Method: "POST", Target: "/v1/messages?beta=true", Body: []byte(`{"messages":[]}`)}
	if len(requests) != 1 || !sameRequest(requests[0], want) {
		t.Fatalf("recorded %+v, want only %+v", requests, want)
	}
	var written Request
	readJSON(t, filepath.Join(dir, "request-001.json"), &written)
	if !sameRequest(written, want) {
		t.Errorf("request-001.json holds %+v, want %+v", written, want)
	}
}

func TestRecorderKeepsABodyThatIsNotJSONAsText(t *testing.T) {
	t.Parallel()
	endpoint := newRecorder(t.TempDir())
	server := httptest.NewServer(endpoint)
	defer server.Close()

	for _, body := range []string{"plain text", ""} {
		response, err := post(t, server.URL+"/upload", body)
		if err != nil {
			t.Fatal(err)
		}
		_ = response.Body.Close()
	}

	requests, err := endpoint.recorded()
	if err != nil {
		t.Fatalf("recorded: %v", err)
	}
	if len(requests) != 2 {
		t.Fatalf("recorded %d requests, want 2", len(requests))
	}
	if got := string(requests[0].Body); got != `"plain text"` {
		t.Errorf("text body kept as %s, want the JSON string \"plain text\"", got)
	}
	if requests[1].Body != nil {
		t.Errorf("empty body kept as %s, want none", requests[1].Body)
	}
}

func TestRecorderFailsWhenItCannotKeepARequest(t *testing.T) {
	t.Parallel()
	endpoint := newRecorder(filepath.Join(t.TempDir(), "missing"))
	server := httptest.NewServer(endpoint)
	defer server.Close()

	response, err := post(t, server.URL+"/v1/messages", `{}`)
	if err != nil {
		t.Fatal(err)
	}
	_ = response.Body.Close()

	if _, err := endpoint.recorded(); err == nil {
		t.Error("recorded succeeded though no request could be written, want an error")
	}
}

// sameRequest compares two requests' bodies as JSON: the file a request is
// written to is indented for a person to read.
func sameRequest(a, b Request) bool {
	var bodyA, bodyB bytes.Buffer
	if json.Compact(&bodyA, a.Body) != nil || json.Compact(&bodyB, b.Body) != nil {
		return false
	}
	return a.Method == b.Method && a.Target == b.Target && bodyA.String() == bodyB.String()
}

func readJSON(t *testing.T, path string, v any) {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(data, v); err != nil {
		t.Fatalf("%s: %v", path, err)
	}
}

func post(t *testing.T, url, body string) (*http.Response, error) {
	t.Helper()
	request, err := http.NewRequestWithContext(t.Context(), http.MethodPost, url, strings.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	return http.DefaultClient.Do(request)
}

func TestRecorderUndoesGzipAndNamesAnEncodingItCannotUndo(t *testing.T) {
	t.Parallel()
	endpoint := newRecorder(t.TempDir())
	server := httptest.NewServer(endpoint)
	defer server.Close()
	plain := `{"messages":[],"tools":[{"name":"Bash"}]}`

	sendEncoded(t, server.URL+"/gzip", "gzip", gzipped(t, plain))
	sendEncoded(t, server.URL+"/zstd", "zstd", []byte("\x28\xb5\x2f\xfd"))

	requests, err := endpoint.recorded()
	if err != nil {
		t.Fatalf("recorded: %v", err)
	}
	if len(requests) != 2 {
		t.Fatalf("recorded %d requests, want 2", len(requests))
	}
	if want := (Request{Method: "POST", Target: "/gzip", Body: []byte(plain)}); requests[0].Encoding != "" || !sameRequest(requests[0], want) {
		t.Errorf("gzip body kept as %+v, want it undone", requests[0])
	}
	if requests[1].Encoding != "zstd" {
		t.Errorf("zstd body kept with encoding %q, want it named", requests[1].Encoding)
	}
}

func gzipped(t *testing.T, text string) []byte {
	t.Helper()
	var zipped bytes.Buffer
	writer := gzip.NewWriter(&zipped)
	if _, err := writer.Write([]byte(text)); err != nil {
		t.Fatal(err)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	return zipped.Bytes()
}

func sendEncoded(t *testing.T, url, encoding string, body []byte) {
	t.Helper()
	request, err := http.NewRequestWithContext(t.Context(), http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	request.Header.Set("Content-Encoding", encoding)
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		t.Fatal(err)
	}
	_ = response.Body.Close()
}

// A body that says it is gzipped and cannot be undone could hold any tool at
// all: the recorder keeps no file for it and fails, so the check cannot pass
// on a request nobody read.
func TestRecorderFailsOnABodyItCannotUndo(t *testing.T) {
	t.Parallel()
	plain := `{"messages":[],"tools":[{"name":"Bash"}]}`
	zipped := gzipped(t, plain)
	for _, tc := range []struct {
		name string
		body []byte
	}{
		{"a body that is not gzip", []byte(plain)},
		{"a gzip body cut short", zipped[:len(zipped)/2]},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			dir := t.TempDir()
			endpoint := newRecorder(dir)
			server := httptest.NewServer(endpoint)
			defer server.Close()

			sendEncoded(t, server.URL+"/v1/messages", "gzip", tc.body)

			if _, err := endpoint.recorded(); err == nil || !strings.Contains(err.Error(), "gzip body") {
				t.Errorf("recorded() error = %v, want one about the gzip body", err)
			}
			if _, err := os.Stat(filepath.Join(dir, "request-001.json")); !errors.Is(err, os.ErrNotExist) {
				t.Errorf("request-001.json: %v, want no file kept for a body nobody read", err)
			}
		})
	}
}

// A client that asks to speak WebSocket is told the endpoint does not, so it
// sends its request as plain HTTP instead; the request that asked is kept like
// any other.
func TestRecorderTurnsAWebSocketUpgradeAway(t *testing.T) {
	t.Parallel()
	endpoint := newRecorder(t.TempDir())
	server := httptest.NewServer(endpoint)
	defer server.Close()
	request, err := http.NewRequestWithContext(t.Context(), http.MethodGet, server.URL+"/v1/responses", http.NoBody)
	if err != nil {
		t.Fatal(err)
	}
	request.Header.Set("Connection", "Upgrade")
	request.Header.Set("Upgrade", "websocket")

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		t.Fatal(err)
	}
	_ = response.Body.Close()

	if response.StatusCode != http.StatusUpgradeRequired {
		t.Errorf("status = %d, want 426", response.StatusCode)
	}
	requests, err := endpoint.recorded()
	if err != nil {
		t.Fatalf("recorded: %v", err)
	}
	if len(requests) != 1 || requests[0].Method != http.MethodGet || requests[0].Target != "/v1/responses" || requests[0].Body != nil {
		t.Errorf("recorded %+v, want only the GET of /v1/responses, with no body", requests)
	}
}

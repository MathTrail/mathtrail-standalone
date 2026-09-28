package oauthserver

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"go.uber.org/zap/zapcore"
)

// brokenSealer seals nothing, as a key ring that failed would.
type brokenSealer struct{}

func (brokenSealer) Seal([]byte, ...string) (string, error) { return "", errors.New("sealing failed") }

func (brokenSealer) Open(string, ...string) ([]byte, error) { return nil, errors.New("opening failed") }

// A registration that fails on this side is answered in the protocol's words
// as a failure of the server's, issues no identifier, and leaves an error line
// that says so without repeating what the client sent.
func TestARegistrationThatFailsHereIsToldSo(t *testing.T) {
	t.Parallel()

	known, logs := knownClients(t, ringOf(t, 'k'), &documents{})
	known.registrations = brokenSealer{}
	registrar := &registrar{clients: known, events: known.events, now: known.now}

	request := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/oauth/register",
		strings.NewReader(`{"redirect_uris":["https://claude.ai/api/mcp/auth_callback"],"client_name":"Claude"}`))
	request.Header.Set("Content-Type", "application/json")
	answer := httptest.NewRecorder()
	registrar.ServeHTTP(answer, request)

	var body map[string]any
	if err := json.Unmarshal(answer.Body.Bytes(), &body); err != nil {
		t.Fatalf("the answer is not JSON: %v", err)
	}
	if answer.Code != http.StatusInternalServerError || body["error"] != "server_error" || body["error_description"] == "" {
		t.Errorf("status %d, answer %v; want 500 and server_error with a description", answer.Code, body)
	}
	if _, issued := body["client_id"]; issued {
		t.Errorf("a failure carries a client_id: %v", body)
	}
	if got := answer.Header().Get("Cache-Control"); got != "no-store" {
		t.Errorf("Cache-Control = %q, want no-store", got)
	}

	lines := logs.FilterMessage(eventAuthRegister).All()
	if len(lines) != 1 || lines[0].Level != zapcore.ErrorLevel || lines[0].ContextMap()["outcome"] != "failed" {
		t.Fatalf("the lines are %v, want one error that the registration failed", lines)
	}
	for key, value := range lines[0].ContextMap() {
		if text, isText := value.(string); isText && (strings.Contains(text, "Claude") || strings.Contains(text, "auth_callback")) {
			t.Errorf("the line carries what the client sent in %s: %q", key, text)
		}
	}
}

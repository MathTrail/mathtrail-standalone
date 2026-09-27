package oauthserver

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"testing"

	"github.com/MathTrail/mathtrail-standalone/internal/infra/cimd"
)

// Whatever a registration sends, it is either registered or refused in the
// protocol's words — never a failure of the server — and a registration made
// opens into exactly what the answer said was registered.
func FuzzRegister(f *testing.F) {
	for _, seed := range []struct{ contentType, body string }{
		{"application/json", `{"redirect_uris":["https://claude.ai/api/mcp/auth_callback"],"client_name":"Claude"}`},
		{"application/json; charset=utf-8", `{"redirect_uris":["http://localhost:3000/cb","https://a.example/cb"]}`},
		{"application/json", `{"redirect_uris":["javascript:alert(1)"]}`},
		{"application/json", `{"redirect_uris":"https://a.example/cb"}`},
		{"application/json", `null`},
		{"text/plain", `{}`},
		{"", ``},
	} {
		f.Add(seed.contentType, seed.body)
	}

	f.Fuzz(func(t *testing.T, contentType, body string) {
		server, _ := knownServer(t)
		request := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/oauth/register", strings.NewReader(body))
		request.Header.Set("Content-Type", contentType)
		answer := httptest.NewRecorder()
		server.Register.ServeHTTP(answer, request)

		switch answer.Code {
		case http.StatusBadRequest, http.StatusRequestEntityTooLarge:
			return
		case http.StatusCreated:
		default:
			t.Fatalf("status = %d for %q, want 201, 400 or 413", answer.Code, body)
		}
		var registered struct {
			ClientID     string   `json:"client_id"`
			RedirectURIs []string `json:"redirect_uris"`
			ClientName   string   `json:"client_name"`
		}
		if err := json.Unmarshal(answer.Body.Bytes(), &registered); err != nil {
			t.Fatalf("the answer to %q is not JSON: %v", body, err)
		}
		client, err := server.clients.resolve(t.Context(), registered.ClientID)
		if err != nil {
			t.Fatalf("the client_id registered for %q does not open: %v", body, err)
		}
		if !slices.Equal(client.RedirectURIs, registered.RedirectURIs) || client.Name != registered.ClientName {
			t.Errorf("the client_id opens into %+v, want what the answer said: %+v", client, registered)
		}
	})
}

// Whatever a client presents as its identifier, it is either no client or a
// client a parent may be sent back to — and it never takes the server down.
func FuzzClientID(f *testing.F) {
	server, _ := knownServer(f)
	issued, err := server.clients.register(registration{RedirectURIs: []string{"https://a.example/cb"}})
	if err != nil {
		f.Fatalf("register() error = %v, want nil", err)
	}
	for _, seed := range []string{
		issued,
		"https://claude.ai/oauth/mcp-oauth-client-metadata",
		"HTTPS:",
		"mt1.d.x.y",
		"mt1.d..",
		"",
	} {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, clientID string) {
		client, err := server.clients.resolve(t.Context(), clientID)
		if err != nil {
			return
		}
		if client.ID != clientID || len(client.RedirectURIs) == 0 ||
			slices.ContainsFunc(client.RedirectURIs, func(uri string) bool { return !redirectable(uri) }) {
			t.Errorf("resolve(%q) = %+v, want no client or one with somewhere a parent may be sent", clientID, client)
		}
	})
}

// knownServer is a server whose every client document is one the fetcher
// refuses: what a fuzzed identifier reaches is the registrations alone.
func knownServer(t testing.TB) (*Server, *documents) {
	t.Helper()

	fetcher := &documents{err: cimd.ErrUnreachable}
	known, _ := knownClients(t, ringOf(t, 'k'), fetcher)
	return &Server{Register: &registrar{clients: known, events: known.events, now: known.now}, clients: known}, fetcher
}

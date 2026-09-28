package httpserver_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// The sign-in's documents and its endpoint are served under the service's own
// name alone, each at its one path: the resource's metadata under the path of
// the resource and not at the root, nothing under a name that is not the
// service's, and nothing at a path the sign-in does not have — the ones that
// arrive later among them.
func TestTheSignInIsServedAtItsOwnPaths(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		method  string
		host    string
		path    string
		status  int
		reached string
	}{
		{http.MethodGet, "example.com", "/.well-known/oauth-protected-resource/mcp", http.StatusOK, reachedResourceMetadata},
		{http.MethodGet, "example.com", "/.well-known/oauth-authorization-server", http.StatusOK, reachedServerMetadata},
		{http.MethodPost, "example.com", "/oauth/register", http.StatusOK, reachedRegister},
		{http.MethodGet, "example.com", "/oauth/authorize", http.StatusOK, reachedAuthorize},
		{http.MethodPost, "example.com", "/oauth/consent", http.StatusOK, reachedConsent},
		{http.MethodGet, "example.com", "/oauth/callback", http.StatusOK, reachedCallback},

		{http.MethodGet, "service-abc.a.run.app", "/.well-known/oauth-protected-resource/mcp", http.StatusNotFound, ""},
		{http.MethodGet, "service-abc.a.run.app", "/.well-known/oauth-authorization-server", http.StatusNotFound, ""},
		{http.MethodPost, "service-abc.a.run.app", "/oauth/register", http.StatusNotFound, ""},
		{http.MethodGet, "service-abc.a.run.app", "/oauth/authorize", http.StatusNotFound, ""},
		{http.MethodGet, "service-abc.a.run.app", "/oauth/callback", http.StatusNotFound, ""},
		{http.MethodGet, "example.com", "/.well-known/oauth-authorization-server/", http.StatusNotFound, ""},
		{http.MethodGet, "example.com", "/.well-known/oauth-protected-resource", http.StatusNotFound, ""},
		{http.MethodGet, "example.com", "/.well-known/openid-configuration", http.StatusNotFound, ""},
		{http.MethodGet, "example.com", "/oauth/authorize/", http.StatusNotFound, ""},
		{http.MethodGet, "example.com", "/oauth/token", http.StatusNotFound, ""},
	} {
		t.Run(tc.method+" "+tc.host+tc.path, func(t *testing.T) {
			t.Parallel()

			req := httptest.NewRequestWithContext(t.Context(), tc.method, tc.path, http.NoBody)
			req.Host = tc.host
			rec := send(t, req)

			if rec.Code != tc.status {
				t.Fatalf("status = %d, want %d", rec.Code, tc.status)
			}
			if tc.reached != "" && rec.Body.String() != tc.reached {
				t.Errorf("body = %q, want %q", rec.Body.String(), tc.reached)
			}
		})
	}
}

// A path of the sign-in asked with a method it does not take is told which
// one it does.
func TestASignInPathSaysWhichMethodItTakes(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		method string
		path   string
		allow  string
	}{
		{http.MethodPost, "/.well-known/oauth-protected-resource/mcp", "GET"},
		{http.MethodPost, "/.well-known/oauth-authorization-server", "GET"},
		{http.MethodGet, "/oauth/register", "POST"},
		{http.MethodPost, "/oauth/authorize", "GET"},
		{http.MethodGet, "/oauth/consent", "POST"},
		{http.MethodPost, "/oauth/callback", "GET"},
	} {
		rec := call(t, tc.method, tc.path)
		if rec.Code != http.StatusMethodNotAllowed {
			t.Errorf("%s %s: status = %d, want %d", tc.method, tc.path, rec.Code, http.StatusMethodNotAllowed)
		}
		if got := rec.Header().Get("Allow"); got != tc.allow {
			t.Errorf("%s %s: Allow = %q, want %q", tc.method, tc.path, got, tc.allow)
		}
	}
}

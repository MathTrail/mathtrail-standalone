package app_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"regexp"
	"strings"
	"testing"

	oauthserver "github.com/MathTrail/mathtrail-standalone/internal/transport/oauth"
)

// With a Google client configured, a parent is shown the consent screen and,
// on allowing, sent on to Google's own sign-in, which is to send them back to
// the service's callback.
func TestContainerSendsAParentOnToGoogle(t *testing.T) {
	t.Parallel()

	cfg := testConfig()
	cfg.GoogleClientID, cfg.GoogleClientSecret = "mathtrail.apps.googleusercontent.com", "a-secret-for-tests"
	container := containerFrom(t, cfg)

	var client struct {
		ClientID string `json:"client_id"`
	}
	registered := serveOne(t, container, http.MethodPost, "/oauth/register",
		`{"redirect_uris":["https://claude.ai/api/mcp/auth_callback"],"client_name":"Claude"}`)
	if err := json.Unmarshal(registered.Body.Bytes(), &client); err != nil || client.ClientID == "" {
		t.Fatalf("the registration answered %s, want a client", registered.Body.String())
	}

	screen := serveOne(t, container, http.MethodGet, "/oauth/authorize?"+url.Values{
		"response_type": {"code"}, "client_id": {client.ClientID},
		"redirect_uri":   {"https://claude.ai/api/mcp/auth_callback"},
		"code_challenge": {"E9Melhoa2OwvFrEMTJguCHaoeK1t8URWbuGJSstw-cM"}, "code_challenge_method": {"S256"},
		"state": {"a-state"}, "resource": {"http://localhost/mcp"}, "scope": {"mcp"},
	}.Encode(), "")
	if screen.Code != http.StatusOK || !strings.Contains(screen.Header().Get("Content-Security-Policy"), "https://accounts.google.com") {
		t.Fatalf("GET /oauth/authorize = %d with policy %q, want the consent screen leading on to Google",
			screen.Code, screen.Header().Get("Content-Security-Policy"))
	}
	request := regexp.MustCompile(`name="request" value="([^"]+)"`).FindStringSubmatch(screen.Body.String())
	if request == nil {
		t.Fatal("the consent screen holds no request to post back")
	}

	allow := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/oauth/consent",
		strings.NewReader(url.Values{"request": {request[1]}, "decision": {"allow"}}.Encode()))
	allow.Host = "localhost"
	allow.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	for _, line := range screen.Header().Values("Set-Cookie") {
		cookie, err := http.ParseSetCookie(line)
		if err != nil {
			t.Fatalf("the screen set %q, which is no cookie: %v", line, err)
		}
		allow.AddCookie(cookie)
	}
	allowed := httptest.NewRecorder()
	container.Router.ServeHTTP(allowed, allow)

	google, err := url.Parse(allowed.Header().Get("Location"))
	if err != nil || allowed.Code != http.StatusSeeOther || google.Host != "accounts.google.com" {
		t.Fatalf("POST /oauth/consent = %d to %q, want 303 to Google", allowed.Code, allowed.Header().Get("Location"))
	}
	if got := google.Query().Get("redirect_uri"); got != "http://localhost/oauth/callback" {
		t.Errorf("Google is to send the parent back to %q, want the service's callback", got)
	}
	if got := google.Query().Get("client_id"); got != cfg.GoogleClientID {
		t.Errorf("Google is asked for the client %q, want %q", got, cfg.GoogleClientID)
	}
}

// The page that asks a parent back to Google for the Drive is served where the
// callback sends them: asked for with no sign-in's request, it is refused at a
// page of the sign-in's own rather than at a path nobody serves.
func TestContainerServesThePageForTheDrive(t *testing.T) {
	t.Parallel()

	container := containerFrom(t, testConfig())
	refused := serveOne(t, container, http.MethodGet, oauthserver.DrivePath+"?request=not-a-request", "")
	if refused.Code != http.StatusBadRequest || !strings.Contains(refused.Body.String(), "invalid_request") {
		t.Errorf("GET %s = %d, want the sign-in's page of refusal", oauthserver.DrivePath, refused.Code)
	}
}

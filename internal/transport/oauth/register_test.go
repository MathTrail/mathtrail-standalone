package oauthserver_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/oauthex"
	"go.uber.org/zap/zaptest/observer"
)

// registrationAnswer is what a registration was answered with: its status,
// its headers and the fields of its body.
type registrationAnswer struct {
	status int
	header http.Header
	fields map[string]any
}

// register sends a registration with the content type and body given.
func register(t *testing.T, served, contentType, body string) registrationAnswer {
	t.Helper()

	request, err := http.NewRequestWithContext(t.Context(), http.MethodPost, served+"/oauth/register", strings.NewReader(body))
	if err != nil {
		t.Fatalf("a registration request: %v", err)
	}
	request.Header.Set("Content-Type", contentType)
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		t.Fatalf("POST /oauth/register: %v", err)
	}
	defer func() { _ = response.Body.Close() }()
	answered := registrationAnswer{status: response.StatusCode, header: response.Header}
	if err := json.NewDecoder(response.Body).Decode(&answered.fields); err != nil {
		t.Fatalf("POST /oauth/register: the answer is not JSON: %v", err)
	}
	return answered
}

// A client registers with the protocol library's own client, as a host would,
// and is registered as the only kind of client this server has: a public one,
// with no secret, for the code and refresh grants — whatever it asked for.
// The identifier it gets back is the registration, sealed.
func TestAClientRegistersAsAPublicClient(t *testing.T) {
	t.Parallel()

	served, logs := serve(t)
	registered, err := oauthex.RegisterClient(t.Context(), served.URL+"/oauth/register", &oauthex.ClientRegistrationMetadata{
		RedirectURIs:            []string{"https://claude.ai/api/mcp/auth_callback", "http://localhost:3000/callback"},
		ClientName:              "Claude",
		TokenEndpointAuthMethod: "client_secret_basic",
		GrantTypes:              []string{"authorization_code", "client_credentials"},
		Scope:                   "mcp offline_access",
	}, served.Client())
	if err != nil {
		t.Fatalf("RegisterClient() error = %v, want nil", err)
	}

	switch {
	case !strings.HasPrefix(registered.ClientID, "mt1.d."):
		t.Errorf("client_id = %q, want a sealed registration", registered.ClientID)
	case registered.ClientSecret != "" || !registered.ClientSecretExpiresAt.IsZero():
		t.Errorf("a secret was issued: %+v", registered)
	case registered.TokenEndpointAuthMethod != "none":
		t.Errorf("token_endpoint_auth_method = %q, want none", registered.TokenEndpointAuthMethod)
	case !slices.Equal(registered.GrantTypes, []string{"authorization_code", "refresh_token"}),
		!slices.Equal(registered.ResponseTypes, []string{"code"}):
		t.Errorf("grants %v and responses %v, want the code and refresh grants, and code", registered.GrantTypes, registered.ResponseTypes)
	case !slices.Equal(registered.RedirectURIs, []string{"https://claude.ai/api/mcp/auth_callback", "http://localhost:3000/callback"}),
		registered.ClientName != "Claude":
		t.Errorf("registered %v as %q, want what was asked for", registered.RedirectURIs, registered.ClientName)
	case !registered.ClientIDIssuedAt.Equal(someDay.Truncate(time.Second)):
		t.Errorf("client_id_issued_at = %v, want %v", registered.ClientIDIssuedAt, someDay)
	}

	lines := logs.FilterMessage("auth_register").All()
	if len(lines) != 1 {
		t.Fatalf("%d auth_register lines, want 1", len(lines))
	}
	if fields := lines[0].ContextMap(); fields["registration"] != "dcr" || fields["outcome"] != "ok" || fields["redirect_host"] != "claude.ai" {
		t.Errorf("the line is %v, want a registration made for claude.ai", fields)
	}
	wantNothingTheClientWrote(t, logs, "Claude", "auth_callback", "localhost:3000")
}

// A registration that cannot be made is refused in the protocol's own words,
// is not stored anywhere to be refused again later, and leaves a line that says
// why without repeating what was sent.
func TestARegistrationIsRefusedInTheProtocolsWords(t *testing.T) {
	t.Parallel()

	uris := func(uris ...string) string {
		encoded, err := json.Marshal(map[string]any{"redirect_uris": uris})
		if err != nil {
			t.Fatalf("encoding %v: %v", uris, err)
		}
		return string(encoded)
	}
	const (
		asJSON  = "application/json"
		badURI  = "invalid_redirect_uri"
		badMeta = "invalid_client_metadata"
	)
	for _, tc := range []struct {
		name        string
		contentType string
		body        string
		status      int
		code        string
		reason      string
	}{
		{"no redirect", asJSON, `{"client_name":"A"}`, 400, badURI, badURI},
		{"an empty list", asJSON, uris(), 400, badURI, badURI},
		{"six redirects", asJSON, uris("https://a.example/1", "https://a.example/2", "https://a.example/3",
			"https://a.example/4", "https://a.example/5", "https://a.example/6"), 400, badURI, badURI},
		{"plain http elsewhere", asJSON, uris("http://a.example/cb"), 400, badURI, badURI},
		{"another scheme", asJSON, uris("com.example.app:/oauth"), 400, badURI, badURI},
		{"a script", asJSON, uris("javascript:alert(1)"), 400, badURI, badURI},
		{"a fragment", asJSON, uris("https://a.example/cb#x"), 400, badURI, badURI},
		{"credentials", asJSON, uris("https://user:secret@a.example/cb"), 400, badURI, badURI},
		{"a relative path", asJSON, uris("/cb"), 400, badURI, badURI},
		{"one good and one bad", asJSON, uris("https://a.example/cb", "http://a.example/cb"), 400, badURI, badURI},
		{"a redirect too long", asJSON, uris("https://a.example/" + strings.Repeat("a", 512)), 400, badURI, badURI},
		{"a name too long", asJSON, `{"redirect_uris":["https://a.example/cb"],"client_name":"` +
			strings.Repeat("я", 101) + `"}`, 400, badMeta, badMeta},
		{"not JSON", asJSON, `redirect_uris=https://a.example/cb`, 400, badMeta, badMeta},
		{"a list", asJSON, `["https://a.example/cb"]`, 400, badMeta, badMeta},
		{"two objects", asJSON, uris("https://a.example/cb") + uris("https://a.example/cb"), 400, badMeta, badMeta},
		{"an object and more", asJSON, uris("https://a.example/cb") + " and more", 400, badMeta, badMeta},
		{"a form", "application/x-www-form-urlencoded", `redirect_uris=https://a.example/cb`, 400, badMeta, badMeta},
		{"no content type", "", uris("https://a.example/cb"), 400, badMeta, badMeta},
		{"past the size of any registration", asJSON, `{"redirect_uris":["https://a.example/cb"],"software_statement":"` +
			strings.Repeat("a", 64<<10) + `"}`, 413, badMeta, "too_large"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			served, logs := serve(t)
			wantRefused(t, register(t, served.URL, tc.contentType, tc.body), tc.status, tc.code)
			lines := logs.FilterMessage("auth_register").All()
			if len(lines) != 1 || lines[0].ContextMap()["outcome"] != "refused" || lines[0].ContextMap()["reason"] != tc.reason {
				t.Errorf("the lines are %v, want one refusal for %s", lines, tc.reason)
			}
			wantNothingTheClientWrote(t, logs, "a.example", "javascript", "secret")
		})
	}
}

// wantRefused holds an answer to a refusal of RFC 7591: the status, the code
// and a description, no identifier, and nothing a cache may keep.
func wantRefused(t *testing.T, got registrationAnswer, status int, code string) {
	t.Helper()

	if got.status != status || got.fields["error"] != code || got.fields["error_description"] == "" {
		t.Errorf("status %d, answer %v; want %d and %s with a description", got.status, got.fields, status, code)
	}
	if _, issued := got.fields["client_id"]; issued {
		t.Errorf("a refusal carries a client_id: %v", got.fields)
	}
	if value := got.header.Get("Cache-Control"); value != "no-store" {
		t.Errorf("Cache-Control = %q, want no-store", value)
	}
}

// wantNothingTheClientWrote holds every line to carrying none of the words a
// client sent in its own request.
func wantNothingTheClientWrote(t *testing.T, logs *observer.ObservedLogs, words ...string) {
	t.Helper()

	lines := logs.All()
	for i := range lines {
		for key, value := range lines[i].ContextMap() {
			text, isText := value.(string)
			if key == "redirect_host" || !isText {
				continue
			}
			for _, word := range words {
				if strings.Contains(text, word) {
					t.Errorf("the %s line carries %q in %s: %q", lines[i].Message, word, key, text)
				}
			}
		}
	}
}

// A registration is one JSON object, and the white space a client may end its
// body with is not more.
func TestARegistrationMayEndInWhiteSpace(t *testing.T) {
	t.Parallel()

	served, _ := serve(t)
	got := register(t, served.URL, "application/json; charset=utf-8", `{"redirect_uris":["https://a.example/cb"]}`+"\r\n\t ")
	if got.status != http.StatusCreated || !strings.HasPrefix(fmt.Sprint(got.fields["client_id"]), "mt1.d.") {
		t.Errorf("status %d, answer %v; want the client registered", got.status, got.fields)
	}
}

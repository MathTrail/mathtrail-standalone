package cimd

import (
	"errors"
	"strings"
	"testing"
)

// clientAt is the address a document in these cases is fetched from.
const clientAt = "https://client.example.com/oauth/metadata"

// claudeLike is a client's metadata document shaped like the one Claude
// publishes — as it was fetched on 2026-09-27, with its grant type this
// service does not offer — naming itself by the address it is served at.
func claudeLike(clientID string) string {
	return `{"client_id":"` + clientID + `","client_name":"Claude","client_uri":"https://claude.ai",` +
		`"redirect_uris":["https://claude.ai/api/mcp/auth_callback"],` +
		`"grant_types":["authorization_code","refresh_token","urn:ietf:params:oauth:grant-type:jwt-bearer"],` +
		`"response_types":["code"],"token_endpoint_auth_method":"none"}`
}

// A document is read as the client's only when it keeps every rule a client
// document keeps.
func TestADocumentIsTheClientsOnlyWhenItKeepsTheRules(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name string
		body string
		read bool
	}{
		{"claude's own shape", claudeLike(clientAt), true},
		{"a secret left out as null", `{"client_id":"` + clientAt + `","client_name":"A","redirect_uris":["https://a.example/cb"],` +
			`"client_secret":null,"client_secret_expires_at":null}`, true},
		{"an empty secret", `{"client_id":"` + clientAt + `","client_name":"A","redirect_uris":["https://a.example/cb"],"client_secret":""}`, false},
		{"signed in with a key", `{"client_id":"` + clientAt + `","client_name":"A","redirect_uris":["https://a.example/cb"],` +
			`"token_endpoint_auth_method":"private_key_jwt","jwks_uri":"https://a.example/jwks"}`, true},

		{"another client's", claudeLike("https://client.example.com/oauth/other"), false},
		{"a client named in another case", claudeLike(strings.ToUpper(clientAt)), false},
		{"no name", `{"client_id":"` + clientAt + `","redirect_uris":["https://a.example/cb"]}`, false},
		{"a blank name", `{"client_id":"` + clientAt + `","client_name":" ","redirect_uris":["https://a.example/cb"]}`, false},
		{"no redirect", `{"client_id":"` + clientAt + `","client_name":"A"}`, false},
		{"an empty redirect list", `{"client_id":"` + clientAt + `","client_name":"A","redirect_uris":[]}`, false},
		{"a redirect that is not a string", `{"client_id":"` + clientAt + `","client_name":"A","redirect_uris":[1]}`, false},
		{"a secret", `{"client_id":"` + clientAt + `","client_name":"A","redirect_uris":["https://a.example/cb"],"client_secret":"s"}`, false},
		{"a secret's end", `{"client_id":"` + clientAt + `","client_name":"A","redirect_uris":["https://a.example/cb"],"client_secret_expires_at":0}`, false},
		{"basic", `{"client_id":"` + clientAt + `","client_name":"A","redirect_uris":["https://a.example/cb"],"token_endpoint_auth_method":"client_secret_basic"}`, false},
		{"post", `{"client_id":"` + clientAt + `","client_name":"A","redirect_uris":["https://a.example/cb"],"token_endpoint_auth_method":"client_secret_post"}`, false},
		{"jwt", `{"client_id":"` + clientAt + `","client_name":"A","redirect_uris":["https://a.example/cb"],"token_endpoint_auth_method":"client_secret_jwt"}`, false},
		{"a secret of its own kind", `{"client_id":"` + clientAt + `","client_name":"A","redirect_uris":["https://a.example/cb"],"token_endpoint_auth_method":"shared_SECRET"}`, false},
		{"not JSON", `client_id=` + clientAt, false},
		{"a list", `[]`, false},
		{"null", `null`, false},
		{"nothing", ``, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			document, err := parseDocument([]byte(tc.body), clientAt)
			if read := err == nil; read != tc.read {
				t.Fatalf("parseDocument() = %+v, %v; want it read: %v", document, err, tc.read)
			}
			if err != nil && !errors.Is(err, ErrDocument) {
				t.Errorf("parseDocument() error = %v, want ErrDocument", err)
			}
		})
	}
}

// Claude's document is read for what the sign-in needs of it: its name and
// the one address it is sent back to.
func TestClaudesDocumentIsReadForItsNameAndRedirect(t *testing.T) {
	t.Parallel()

	document, err := parseDocument([]byte(claudeLike(clientAt)), clientAt)
	if err != nil {
		t.Fatalf("parseDocument() error = %v, want nil", err)
	}
	if document.ClientID != clientAt || document.ClientName != "Claude" ||
		len(document.RedirectURIs) != 1 || document.RedirectURIs[0] != "https://claude.ai/api/mcp/auth_callback" {
		t.Errorf("parseDocument() = %+v, want Claude with its one redirect", document)
	}
}

// Whatever arrives at a client's address, it is either refused or a document
// that names the client, calls it something and says where to send the parent
// back to — and it never takes the reading down.
func FuzzDocument(f *testing.F) {
	for _, seed := range []string{
		claudeLike(clientAt),
		`{"client_id":"` + clientAt + `","client_name":"A","redirect_uris":["x"],"client_secret":"s"}`,
		`{"client_id":1}`,
		`[]`,
		`null`,
		``,
	} {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, body string) {
		document, err := parseDocument([]byte(body), clientAt)
		if err != nil {
			return
		}
		if document.ClientID != clientAt || strings.TrimSpace(document.ClientName) == "" || len(document.RedirectURIs) == 0 {
			t.Errorf("parseDocument(%q) = %+v, want it refused", body, document)
		}
	})
}

package oauthserver_test

import (
	"encoding/json"
	"io"
	"net/http"
	"slices"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/oauthex"
)

// answer is what the served server answered: its status, its headers and its
// body.
type answer struct {
	status int
	header http.Header
	body   []byte
}

// get asks the served server for one of its documents.
func get(t *testing.T, served, path string) answer {
	t.Helper()

	request, err := http.NewRequestWithContext(t.Context(), http.MethodGet, served+path, http.NoBody)
	if err != nil {
		t.Fatalf("a request for %s: %v", path, err)
	}
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		t.Fatalf("GET %s: %v", path, err)
	}
	defer func() { _ = response.Body.Close() }()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatalf("GET %s: reading the body: %v", path, err)
	}
	return answer{status: response.StatusCode, header: response.Header, body: body}
}

// The resource's metadata names the resource, the one authorization server,
// the one scope and a token in the header — kept for an hour, and readable
// from any origin.
func TestTheResourcesMetadataNamesTheResource(t *testing.T) {
	t.Parallel()

	served, _ := serve(t)
	got := get(t, served.URL, "/.well-known/oauth-protected-resource/mcp")
	wantPublicMetadata(t, "the resource's metadata", got)

	var document oauthex.ProtectedResourceMetadata
	if err := json.Unmarshal(got.body, &document); err != nil {
		t.Fatalf("the document is not JSON: %v", err)
	}
	if document.Resource != issuer+"/mcp" || !slices.Equal(document.AuthorizationServers, []string{issuer}) ||
		!slices.Equal(document.ScopesSupported, []string{"mcp"}) || !slices.Equal(document.BearerMethodsSupported, []string{"header"}) {
		t.Errorf("the document is %+v, want the resource, this server, the scope and the header", document)
	}
}

// The authorization server's metadata says where every endpoint is and what
// the server does — and leaves out what it has none of, a jwks_uri among them.
func TestTheServersMetadataSaysWhatItDoes(t *testing.T) {
	t.Parallel()

	served, _ := serve(t)
	got := get(t, served.URL, "/.well-known/oauth-authorization-server")
	wantPublicMetadata(t, "the server's metadata", got)

	var document map[string]any
	if err := json.Unmarshal(got.body, &document); err != nil {
		t.Fatalf("the document is not JSON: %v", err)
	}
	want := map[string]any{
		"issuer":                                         issuer,
		"authorization_endpoint":                         issuer + "/oauth/authorize",
		"token_endpoint":                                 issuer + "/oauth/token",
		"registration_endpoint":                          issuer + "/oauth/register",
		"revocation_endpoint":                            issuer + "/oauth/revoke",
		"scopes_supported":                               []any{"mcp"},
		"response_types_supported":                       []any{"code"},
		"response_modes_supported":                       []any{"query"},
		"grant_types_supported":                          []any{"authorization_code", "refresh_token"},
		"token_endpoint_auth_methods_supported":          []any{"none"},
		"revocation_endpoint_auth_methods_supported":     []any{"none"},
		"code_challenge_methods_supported":               []any{"S256"},
		"client_id_metadata_document_supported":          true,
		"authorization_response_iss_parameter_supported": true,
	}
	for key, value := range want {
		if encoded, err := json.Marshal(document[key]); err != nil || string(encoded) != mustJSON(t, value) {
			t.Errorf("%s = %s, want %s", key, encoded, mustJSON(t, value))
		}
	}
	for key := range document {
		if _, expected := want[key]; !expected {
			t.Errorf("the document carries %q, want only the fields above", key)
		}
	}
}

// The documents are what the protocol library's own client, as a host would
// use it, accepts: it finds the resource and the server's issuer where it
// expects them, and PKCE among what the server does.
func TestAClientOfTheProtocolAcceptsTheDocuments(t *testing.T) {
	t.Parallel()

	served, _ := serve(t)
	resource, err := oauthex.GetProtectedResourceMetadata(t.Context(),
		served.URL+"/.well-known/oauth-protected-resource/mcp", issuer+"/mcp", served.Client())
	if err != nil {
		t.Fatalf("GetProtectedResourceMetadata() error = %v, want nil", err)
	}
	if !slices.Equal(resource.AuthorizationServers, []string{issuer}) {
		t.Errorf("authorization servers = %v, want %s", resource.AuthorizationServers, issuer)
	}
	server, err := oauthex.GetAuthServerMeta(t.Context(),
		served.URL+"/.well-known/oauth-authorization-server", issuer, served.Client())
	if err != nil || server == nil {
		t.Fatalf("GetAuthServerMeta() = %v, %v; want the metadata", server, err)
	}
	if !server.ClientIDMetadataDocumentSupported || !slices.Contains(server.CodeChallengeMethodsSupported, "S256") {
		t.Errorf("the metadata read is %+v, want client documents and S256", server)
	}
}

// wantPublicMetadata holds an answer to the way public metadata is served:
// JSON, kept for an hour, readable from any origin.
func wantPublicMetadata(t *testing.T, what string, got answer) {
	t.Helper()

	if got.status != http.StatusOK {
		t.Errorf("%s: status = %d, want %d", what, got.status, http.StatusOK)
	}
	for header, want := range map[string]string{
		"Content-Type":                "application/json",
		"Cache-Control":               "max-age=3600",
		"Access-Control-Allow-Origin": "*",
	} {
		if value := got.header.Get(header); value != want {
			t.Errorf("%s: %s = %q, want %q", what, header, value, want)
		}
	}
}

func mustJSON(t *testing.T, value any) string {
	t.Helper()

	encoded, err := json.Marshal(value)
	if err != nil {
		t.Fatalf("encoding %v: %v", value, err)
	}
	return string(encoded)
}

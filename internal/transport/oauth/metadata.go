package oauthserver

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/modelcontextprotocol/go-sdk/auth"
	"github.com/modelcontextprotocol/go-sdk/oauthex"
)

// The paths of the sign-in, under the issuer.
const (
	// resourcePath is where the MCP endpoint is served: the resource the
	// tokens are for.
	resourcePath = "/mcp"
	// resourceMetadataPath is where the resource's metadata is, at the path
	// RFC 9728 derives from the resource's own.
	resourceMetadataPath = "/.well-known/oauth-protected-resource" + resourcePath
	authorizePath        = "/oauth/authorize"
	tokenPath            = "/oauth/token" //nolint:gosec // the path of the token endpoint, not a credential
	registerPath         = "/oauth/register"
	revokePath           = "/oauth/revoke"
)

// metadataCacheControl is how long a host may keep a metadata document: an
// hour, since nothing in it changes between releases.
const metadataCacheControl = "max-age=3600"

// resourceMetadataHandler serves the protected resource's metadata with the
// protocol library's own handler, which lets any origin read it, as public
// metadata is read. The resource is the MCP endpoint, the one authorization
// server is this one, and a token is presented in the Authorization header
// alone.
func resourceMetadataHandler(issuer, scope string) http.Handler {
	document := auth.ProtectedResourceMetadataHandler(&oauthex.ProtectedResourceMetadata{
		Resource:               issuer + resourcePath,
		AuthorizationServers:   []string{issuer},
		ScopesSupported:        []string{scope},
		BearerMethodsSupported: []string{"header"},
	})
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", metadataCacheControl)
		document.ServeHTTP(w, r)
	})
}

// serverMetadata is the authorization server's metadata document (RFC 8414):
// where its endpoints are, and what it does. It is written out field by field
// rather than with the protocol library's type, which always writes a
// jwks_uri, empty when there is none, where the RFC leaves an unused field out:
// the tokens here are opaque, and there are no keys to publish.
//
// Every client is public: it proves itself with PKCE rather than a secret, at
// the token endpoint and at revocation alike. A client may name itself by the
// address of its own metadata document, and every answer to an authorization
// request says which server gave it.
type serverMetadata struct {
	Issuer                                     string   `json:"issuer"`
	AuthorizationEndpoint                      string   `json:"authorization_endpoint"`
	TokenEndpoint                              string   `json:"token_endpoint"`
	RegistrationEndpoint                       string   `json:"registration_endpoint"`
	RevocationEndpoint                         string   `json:"revocation_endpoint"`
	ScopesSupported                            []string `json:"scopes_supported"`
	ResponseTypesSupported                     []string `json:"response_types_supported"`
	ResponseModesSupported                     []string `json:"response_modes_supported"`
	GrantTypesSupported                        []string `json:"grant_types_supported"`
	TokenEndpointAuthMethodsSupported          []string `json:"token_endpoint_auth_methods_supported"`
	RevocationEndpointAuthMethodsSupported     []string `json:"revocation_endpoint_auth_methods_supported"`
	CodeChallengeMethodsSupported              []string `json:"code_challenge_methods_supported"`
	ClientIDMetadataDocumentSupported          bool     `json:"client_id_metadata_document_supported"`
	AuthorizationResponseIssParameterSupported bool     `json:"authorization_response_iss_parameter_supported"`
}

// serverMetadataHandler serves the authorization server's metadata, encoded
// once: nothing in it changes while the process runs.
func serverMetadataHandler(issuer, scope string) (http.Handler, error) {
	body, err := json.Marshal(serverMetadata{
		Issuer:                                     issuer,
		AuthorizationEndpoint:                      issuer + authorizePath,
		TokenEndpoint:                              issuer + tokenPath,
		RegistrationEndpoint:                       issuer + registerPath,
		RevocationEndpoint:                         issuer + revokePath,
		ScopesSupported:                            []string{scope},
		ResponseTypesSupported:                     []string{"code"},
		ResponseModesSupported:                     []string{"query"},
		GrantTypesSupported:                        registeredGrants,
		TokenEndpointAuthMethodsSupported:          []string{publicClient},
		RevocationEndpointAuthMethodsSupported:     []string{publicClient},
		CodeChallengeMethodsSupported:              []string{"S256"},
		ClientIDMetadataDocumentSupported:          true,
		AuthorizationResponseIssParameterSupported: true,
	})
	if err != nil {
		return nil, fmt.Errorf("oauth: encode the server's metadata: %w", err)
	}
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		header := w.Header()
		header.Set("Content-Type", "application/json")
		header.Set("Cache-Control", metadataCacheControl)
		header.Set("Access-Control-Allow-Origin", "*")
		_, _ = w.Write(body)
	}), nil
}

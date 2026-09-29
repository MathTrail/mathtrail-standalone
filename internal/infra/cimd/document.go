package cimd

import (
	"encoding/json"
	"fmt"
	"strings"
)

// Document is what the service reads of a client's metadata document.
type Document struct {
	// ClientID is the address the document was fetched from, which the
	// document names itself by.
	ClientID string
	// ClientName is what the client calls itself.
	ClientName string
	// RedirectURIs are where a sign-in may send the parent back to, as the
	// document lists them.
	RedirectURIs []string
}

// mostRedirectURIs is how many addresses a document may list for a sign-in to
// send the parent back to. The clients known list one to a few. A document is
// kept in memory once read, and a list of thousands of short addresses fits
// in its size while each of them costs more to hold than it took to send.
const mostRedirectURIs = 32

// parseDocument reads a client's metadata document by the rules every such
// document keeps. It is a JSON object that names itself by the address it was
// fetched from, exactly; it says what the client is called and where a sign-in
// may send the parent back to, in no more addresses than a client needs; and
// it carries no shared secret, since nothing could have been shared with a
// client nobody registered — neither a secret of its own nor a way of signing
// in to the token endpoint that needs one.
func parseDocument(body []byte, clientID string) (Document, error) {
	var fields struct {
		ClientID                string          `json:"client_id"`
		ClientName              string          `json:"client_name"`
		RedirectURIs            []string        `json:"redirect_uris"`
		TokenEndpointAuthMethod string          `json:"token_endpoint_auth_method"`
		ClientSecret            json.RawMessage `json:"client_secret"`
		ClientSecretExpiresAt   json.RawMessage `json:"client_secret_expires_at"`
	}
	if err := json.Unmarshal(body, &fields); err != nil {
		return Document{}, fmt.Errorf("%w: not a JSON object of client metadata", ErrDocument)
	}
	switch {
	case fields.ClientID != clientID:
		return Document{}, fmt.Errorf("%w: it names another client", ErrDocument)
	case strings.TrimSpace(fields.ClientName) == "":
		return Document{}, fmt.Errorf("%w: it has no client_name", ErrDocument)
	case len(fields.RedirectURIs) == 0:
		return Document{}, fmt.Errorf("%w: it lists no redirect_uris", ErrDocument)
	case len(fields.RedirectURIs) > mostRedirectURIs:
		return Document{}, fmt.Errorf("%w: it lists more than %d redirect_uris", ErrDocument, mostRedirectURIs)
	case given(fields.ClientSecret) || given(fields.ClientSecretExpiresAt):
		return Document{}, fmt.Errorf("%w: it carries a client secret", ErrDocument)
	case strings.Contains(strings.ToLower(fields.TokenEndpointAuthMethod), "secret"):
		return Document{}, fmt.Errorf("%w: it signs in with a shared secret", ErrDocument)
	}
	return Document{
		ClientID:     fields.ClientID,
		ClientName:   fields.ClientName,
		RedirectURIs: fields.RedirectURIs,
	}, nil
}

// given reports whether a field a document may leave out holds a value. A null
// is the field left out, as a serialiser writes a field it has nothing for.
func given(field json.RawMessage) bool {
	return len(field) > 0 && string(field) != "null"
}

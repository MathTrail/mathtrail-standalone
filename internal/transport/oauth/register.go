package oauthserver

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"slices"
	"time"
	"unicode/utf8"

	"github.com/modelcontextprotocol/go-sdk/oauthex"
)

const (
	// maxRedirectURIs is how many addresses one registration may name. A
	// client needs one, or a few for its several surfaces; the bound keeps
	// the identifier the registration is sealed into short.
	maxRedirectURIs = 5
	// maxClientName is the longest name a client may give itself, here or in
	// its own document: the consent screen shows it to the parent.
	maxClientName = 100
	// maxRegistration is the largest registration read.
	maxRegistration = 64 << 10
	// publicClient is the one way a client proves itself at the token
	// endpoint: with nothing but PKCE, since no secret is ever issued.
	publicClient = "none"
)

// The grants and the answers every registered client gets.
var (
	registeredGrants    = []string{"authorization_code", "refresh_token"}
	registeredResponses = []string{"code"}
)

// registration is a client registered here: where a sign-in may send the
// parent back to, what it calls itself, and when it registered. It is sealed
// into the identifier the client is given, and kept nowhere else.
type registration struct {
	RedirectURIs []string `json:"redirect_uris"`
	ClientName   string   `json:"client_name,omitempty"`
	IssuedAt     int64    `json:"issued_at"`
}

// refusal is a registration the server did not make, in the words of
// RFC 7591, with the reason its line gives.
type refusal struct {
	status      int
	code        string
	description string
	reason      string
}

// registrar registers clients (RFC 7591) without keeping anything: the record
// of a registration is the identifier it hands back.
type registrar struct {
	clients *clients
	events  *signInLog
	now     func() time.Time
}

// ServeHTTP registers a public client from the metadata it sends, or refuses
// it in the protocol's words. Whatever the client asks for, it is registered
// as the only kind this server has: a public client that proves itself with
// PKCE, for the code and refresh grants — RFC 7591 lets a server register
// other values than the ones asked for, and the answer says which.
func (reg *registrar) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")

	asked, refused := readRegistration(w, r)
	var record registration
	if refused == nil {
		record, refused = registrationOf(asked, reg.now())
	}
	if refused != nil {
		reg.events.registrationRefused(r.Context(), refused.reason)
		writeJSON(w, refused.status, &oauthex.ClientRegistrationError{
			ErrorCode:        refused.code,
			ErrorDescription: refused.description,
		})
		return
	}

	clientID, err := reg.clients.register(record)
	if err != nil {
		reg.events.registrationFailed(r.Context(), err)
		writeJSON(w, http.StatusInternalServerError, &oauthex.ClientRegistrationError{
			ErrorCode:        "server_error",
			ErrorDescription: "the registration could not be made; try again",
		})
		return
	}
	reg.events.registered(r.Context(), hostOf(record.RedirectURIs[0]))
	writeJSON(w, http.StatusCreated, &oauthex.ClientRegistrationResponse{
		ClientRegistrationMetadata: oauthex.ClientRegistrationMetadata{
			RedirectURIs:            record.RedirectURIs,
			TokenEndpointAuthMethod: publicClient,
			GrantTypes:              registeredGrants,
			ResponseTypes:           registeredResponses,
			ClientName:              record.ClientName,
		},
		ClientID:         clientID,
		ClientIDIssuedAt: time.Unix(record.IssuedAt, 0),
	})
}

// readRegistration reads the client metadata a registration sends: one JSON
// object, and no larger than a registration ever needs to be.
func readRegistration(w http.ResponseWriter, r *http.Request) (*oauthex.ClientRegistrationMetadata, *refusal) {
	media, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil || media != "application/json" {
		return nil, &refusal{http.StatusBadRequest, "invalid_client_metadata",
			"a registration is sent as application/json", "invalid_client_metadata"}
	}
	var asked oauthex.ClientRegistrationMetadata
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, maxRegistration))
	err = decoder.Decode(&asked)
	if err == nil {
		err = nothingAfter(decoder)
	}
	var tooLarge *http.MaxBytesError
	switch {
	case errors.As(err, &tooLarge):
		return nil, &refusal{http.StatusRequestEntityTooLarge, "invalid_client_metadata",
			fmt.Sprintf("a registration is at most %d bytes", maxRegistration), "too_large"}
	case err != nil:
		return nil, &refusal{http.StatusBadRequest, "invalid_client_metadata",
			"a registration is one JSON object of client metadata", "invalid_client_metadata"}
	}
	return &asked, nil
}

// registrationOf is the registration a client's metadata asks for, when it
// can be made: a few addresses a sign-in may send the parent back to, and a
// name short enough to show.
func registrationOf(asked *oauthex.ClientRegistrationMetadata, now time.Time) (registration, *refusal) {
	switch uris := asked.RedirectURIs; {
	case len(uris) == 0:
		return registration{}, &refusal{http.StatusBadRequest, "invalid_redirect_uri",
			"a registration names at least one redirect URI", "invalid_redirect_uri"}
	case len(uris) > maxRedirectURIs:
		return registration{}, &refusal{http.StatusBadRequest, "invalid_redirect_uri",
			fmt.Sprintf("a registration names at most %d redirect URIs", maxRedirectURIs), "invalid_redirect_uri"}
	case !allRedirectable(uris):
		return registration{}, &refusal{http.StatusBadRequest, "invalid_redirect_uri",
			fmt.Sprintf("a redirect URI is an absolute https address, or http to localhost, 127.0.0.1 or [::1], "+
				"with no fragment and no credentials, of at most %d characters", maxRedirectURI), "invalid_redirect_uri"}
	case utf8.RuneCountInString(asked.ClientName) > maxClientName:
		return registration{}, &refusal{http.StatusBadRequest, "invalid_client_metadata",
			fmt.Sprintf("client_name is at most %d characters", maxClientName), "invalid_client_metadata"}
	}
	return registration{
		RedirectURIs: slices.Clone(asked.RedirectURIs),
		ClientName:   asked.ClientName,
		IssuedAt:     now.Unix(),
	}, nil
}

func allRedirectable(uris []string) bool {
	return !slices.ContainsFunc(uris, func(uri string) bool { return !redirectable(uri) })
}

// errMoreThanOne is a registration that goes on past its one object.
var errMoreThanOne = errors.New("oauth: a registration of more than one JSON value")

// nothingAfter holds a body to the one JSON value already read from it: what
// follows is nothing but white space.
func nothingAfter(decoder *json.Decoder) error {
	_, err := decoder.Token()
	switch {
	case errors.Is(err, io.EOF):
		return nil
	case err != nil:
		return err
	}
	return errMoreThanOne
}

// writeJSON answers with a JSON body.
func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

package oauthserver

import (
	"errors"
	"fmt"
	"mime"
	"net/http"
	"net/url"
)

// maxTokenForm is the largest request read at the token and revocation
// endpoints: a code or a token and a few words take a few kilobytes.
const maxTokenForm = 64 << 10

// oauthError is a refusal as the token and revocation endpoints write it
// (RFC 6749 5.2).
type oauthError struct {
	Code        string `json:"error"`
	Description string `json:"error_description,omitempty"`
}

// readForm reads what a host posts to the token or the revocation endpoint: a
// form, as the protocol has it, no larger than such a request ever needs to
// be, and with each parameter given once — but for the resource, which a
// request may name more than once and is then told the one there is. Clients
// here are public and have no secret, so a client that tries to prove itself
// in a header is refused as a client whose proof was not accepted: it may
// then name itself in the body, as a public client does. A form that was read
// is returned beside a refusal of it too, so that its line can say what it
// asked for.
func readForm(w http.ResponseWriter, r *http.Request) (url.Values, *refusal) {
	media, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil || media != "application/x-www-form-urlencoded" {
		return nil, &refusal{http.StatusBadRequest, "invalid_request",
			"the request is sent as application/x-www-form-urlencoded", "invalid_request"}
	}
	r.Body = http.MaxBytesReader(w, r.Body, maxTokenForm)
	if err := r.ParseForm(); err != nil {
		var tooLarge *http.MaxBytesError
		if errors.As(err, &tooLarge) {
			return nil, &refusal{http.StatusRequestEntityTooLarge, "invalid_request",
				fmt.Sprintf("the request is at most %d bytes", maxTokenForm), "too_large"}
		}
		return nil, &refusal{http.StatusBadRequest, "invalid_request", "the request is not a form that reads", "invalid_request"}
	}
	for name, values := range r.PostForm {
		if len(values) > 1 && name != "resource" {
			return r.PostForm, &refusal{http.StatusBadRequest, "invalid_request",
				"each parameter is given once at most", "invalid_request"}
		}
	}
	if r.Header.Get("Authorization") != "" {
		return r.PostForm, &refusal{http.StatusUnauthorized, "invalid_client",
			"clients here are public: they name themselves with client_id in the body, and prove nothing in a header",
			"invalid_client"}
	}
	return r.PostForm, nil
}

// writeRefusal answers with a refusal in the protocol's words. A client that
// tried to prove itself in a header is answered 401 with a challenge, as the
// protocol asks, of the scheme a client with a secret would use.
func writeRefusal(w http.ResponseWriter, issuer string, how *refusal) {
	if how.status == http.StatusUnauthorized {
		w.Header().Set("WWW-Authenticate", `Basic realm="`+issuer+`"`)
	}
	writeJSON(w, how.status, &oauthError{Code: how.code, Description: how.description})
}

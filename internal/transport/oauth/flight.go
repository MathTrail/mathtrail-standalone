package oauthserver

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"
)

const (
	// flightLifetime is how long a sign-in may take from the host's request
	// to Google's answer, the screens of both included.
	flightLifetime = 10 * time.Minute
	// clockSkew is how far the clocks of two instances, or of this one and
	// Google, may disagree.
	clockSkew = 60 * time.Second

	// csrfCookie ties a sign-in to the browser that began it.
	csrfCookie = "mt_csrf"
	// cookiePath is where the sign-in's cookies are sent: its own endpoints,
	// and nowhere else.
	cookiePath = "/oauth"
	// cookieBytes is how many random bytes the cookie is made of.
	cookieBytes = 32
)

// flight is an authorization request between the host asking for it and
// Google answering. What a stateful server would keep in a session is sealed
// into the request instead, which travels with the parent: in the consent
// screen's form, and as the state Google carries back.
//
// The client is kept as a digest of its identifier, which is all a later step
// compares it with: an identifier this server issued carries its whole
// registration, and the request rides in the address Google is sent.
type flight struct {
	Client       string `json:"client"`
	Registration string `json:"registration"`
	RedirectURI  string `json:"redirect_uri"`
	State        string `json:"state,omitempty"`
	Challenge    string `json:"code_challenge"`
	Resource     string `json:"resource"`
	Scope        string `json:"scope"`
	// Verifier and Nonce are what this server proves and expects at Google.
	Verifier string `json:"verifier"`
	Nonce    string `json:"nonce"`
	// Cookie is the digest of the cookie the browser was given, which is the
	// only place the cookie is compared with.
	Cookie    string `json:"cookie"`
	StartedAt int64  `json:"started_at"`
}

// The ways a sealed request is refused.
var (
	// errForeignFlight is a request this server did not seal, for itself, as
	// a request — or one changed since.
	errForeignFlight = errors.New("oauth: not a request of this server's")
	// errLateFlight is a request older than a sign-in may take.
	errLateFlight = errors.New("oauth: the sign-in took too long")
)

// sealFlight seals a request, bound to this issuer.
func (f *flow) sealFlight(request *flight) (string, error) {
	plain, err := json.Marshal(request)
	if err != nil {
		return "", fmt.Errorf("oauth: encode the request: %w", err)
	}
	sealed, err := f.flights.Seal(plain, f.issuer)
	if err != nil {
		return "", fmt.Errorf("oauth: seal the request: %w", err)
	}
	return sealed, nil
}

// openFlight opens a request this server sealed, if it is still under way: no
// older than a sign-in may take, and not dated ahead of this clock by more
// than the clocks may disagree.
func (f *flow) openFlight(sealed string) (flight, error) {
	plain, err := f.flights.Open(sealed, f.issuer)
	if err != nil {
		return flight{}, fmt.Errorf("%w: %w", errForeignFlight, err)
	}
	var request flight
	if err := json.Unmarshal(plain, &request); err != nil {
		return flight{}, fmt.Errorf("%w: %w", errForeignFlight, err)
	}
	age := f.now().Sub(time.Unix(request.StartedAt, 0))
	if age > flightLifetime || age < -clockSkew {
		return flight{}, errLateFlight
	}
	return request, nil
}

// fromThisBrowser reports whether a request came back to the browser it was
// begun in: the browser holds the cookie whose digest the request carries.
func fromThisBrowser(r *http.Request, request *flight) bool {
	cookie, err := r.Cookie(csrfCookie)
	if err != nil {
		return false
	}
	return sameDigest(digestOf(cookie.Value), request.Cookie)
}

// setCSRFCookie gives the browser the cookie a sign-in is tied to. Lax is what
// lets it come back with Google's redirect, a top-level navigation from
// another site, and keeps it off anything another site posts.
func setCSRFCookie(w http.ResponseWriter, value string) {
	http.SetCookie(w, &http.Cookie{
		Name:     csrfCookie,
		Value:    value,
		Path:     cookiePath,
		MaxAge:   int(flightLifetime / time.Second),
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
}

// dropCSRFCookie takes the cookie back once its sign-in is over.
func dropCSRFCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     csrfCookie,
		Path:     cookiePath,
		MaxAge:   -1,
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
}

// digestOf is the SHA-256 of the parts as base64url. Each part is written with
// its length in front, so no two different lists of parts have one digest.
func digestOf(parts ...string) string {
	hash := sha256.New()
	for _, part := range parts {
		hash.Write(binary.AppendUvarint(nil, uint64(len(part))))
		hash.Write([]byte(part))
	}
	return base64.RawURLEncoding.EncodeToString(hash.Sum(nil))
}

// randomValue is as many random bytes as asked for, as base64url.
func randomValue(bytes int) string {
	value := make([]byte, bytes)
	// It never fails: a system without randomness stops the process instead.
	_, _ = rand.Read(value)
	return base64.RawURLEncoding.EncodeToString(value)
}

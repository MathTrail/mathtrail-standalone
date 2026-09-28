package oauthserver

import (
	"errors"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/MathTrail/mathtrail-standalone/internal/infra/googleauth"
)

// CallbackPath is where Google sends the parent back to, under the issuer:
// the redirect URI the service's client at Google has registered.
const CallbackPath = "/oauth/callback"

// The steps of a sign-in, as its lines name them.
const (
	stepAuthorize = "authorize"
	stepConsent   = "consent"
	stepCallback  = "callback"
)

// flow is a parent's way through a sign-in: the host's request, the consent
// screen, and Google's answer, which ends in the code the host exchanges for
// its tokens. Nothing of a sign-in under way is kept here: it travels sealed,
// with the parent.
type flow struct {
	issuer   string
	resource string
	scope    string
	clients  *clients
	flights  sealer
	consents sealer
	codes    sealer
	userID   func(account string) string
	google   googleauth.SignIn
	pages    *pages
	events   *signInLog
	now      func() time.Time
}

// sendBack sends the parent's browser back to the client with the answer
// added to its own address, whose query is kept as it is; beside the answer
// go the host's state and this server's name (RFC 9207), so that a client
// signing in at several servers knows which one answered.
func (f *flow) sendBack(w http.ResponseWriter, r *http.Request, status int, request *flight, answer url.Values) {
	if request.State != "" {
		answer.Set("state", request.State)
	}
	answer.Set("iss", f.issuer)
	//nolint:gosec // an address the client registered, matched byte for byte before the request was sealed
	http.Redirect(w, r, withQuery(request.RedirectURI, answer.Encode()), status)
}

// withQuery is an address with a query added after whatever query it has.
func withQuery(address, query string) string {
	switch {
	case !strings.Contains(address, "?"):
		return address + "?" + query
	case strings.HasSuffix(address, "?"), strings.HasSuffix(address, "&"):
		return address + query
	}
	return address + "&" + query
}

// refused is the answer a client hears when the sign-in did not give it a
// code, in the words of RFC 6749.
func refused(code, description string) url.Values {
	return url.Values{"error": {code}, "error_description": {description}}
}

// resume opens the request a step of the sign-in came back with, and holds it
// to the browser it was begun in. When either fails, the step stops at a page
// and the client hears nothing: the address it gave is inside what could not
// be read, or it is not this browser's parent who would be sent there.
func (f *flow) resume(w http.ResponseWriter, r *http.Request, step, sealed string) (flight, bool) {
	request, err := f.openFlight(sealed)
	switch {
	case errors.Is(err, errLateFlight):
		f.stopAt(w, r, step, stoppedExpired, "expired")
		return flight{}, false
	case err != nil:
		f.stopAt(w, r, step, stoppedRequest, "unknown_request")
		return flight{}, false
	case !fromThisBrowser(r, &request):
		f.stopAt(w, r, step, stoppedCookie, "cookie")
		return flight{}, false
	}
	return request, true
}

// stopAt stops a step of the sign-in at a page, and leaves one line of it: the
// refusal and why — an error of ours as well, when the page could not be
// drawn.
func (f *flow) stopAt(w http.ResponseWriter, r *http.Request, step string, how stop, reason string) {
	f.events.rejected(r.Context(), step, reason, f.pages.showRefusal(w, r, how))
}

// fail stops a step of the sign-in at the page of a failure of this server's
// own, and leaves one line of it — with the page's own failure beside, when
// the page could not be drawn either.
func (f *flow) fail(w http.ResponseWriter, r *http.Request, step string, err error) {
	if pageErr := f.pages.showRefusal(w, r, stoppedFailed); pageErr != nil {
		err = errors.Join(err, pageErr)
	}
	f.events.failed(r.Context(), step, err)
}

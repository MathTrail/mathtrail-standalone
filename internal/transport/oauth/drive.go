package oauthserver

import (
	"net/http"
	"net/url"
)

// DrivePath is where a parent who left Google's box for the Drive unticked is
// asked to go back to Google and tick it.
const DrivePath = "/oauth/drive"

// outcomeRetry is how a callback ends that asks the parent back to Google for
// the Drive: its line says so, and the sign-in goes on.
const outcomeRetry = "retry"

// askForDrive sends a parent who left Google's box for the Drive unticked on to
// the page that asks them to go back and tick it. The client hears nothing
// yet: the request goes on, readied for a way through Google of its own, and
// rides to the page in its address. The page is not drawn at the callback's
// own address, which carries Google's code: drawn there again — reloaded, or
// returned to from Google's screen — it would spend that code a second time.
// The browser keeps the cookie the sign-in is tied to for the time the way
// back may take, so that the first way, too, still finishes for a parent who
// returns to Google's screen with the browser's own Back.
func (f *flow) askForDrive(w http.ResponseWriter, r *http.Request, request *flight) {
	sealed, err := f.readyForGoogle(request)
	if err != nil {
		dropCSRFCookie(w)
		f.fail(w, r, stepCallback, err)
		return
	}
	// resume held the request to this cookie, so the browser has it.
	if cookie, err := r.Cookie(csrfCookie); err == nil {
		setCSRFCookie(w, cookie.Value)
	}
	http.Redirect(w, r, DrivePath+"?"+url.Values{"request": {sealed}}.Encode(), http.StatusSeeOther)
}

// drive draws the page that asks for the Drive, for the request its address
// carries, held to the browser it was begun in as at every step. Drawn again,
// it is the same page, and asks Google nothing. Its form is answered with the
// consent screen's: going back, or cancelling.
func (f *flow) drive(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")

	query, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil || len(query["request"]) > 1 {
		f.stopAt(w, r, stepDrive, stoppedRequest, "invalid_request")
		return
	}
	sealed := query.Get("request")
	request, ok := f.resume(w, r, stepDrive, sealed)
	if !ok {
		return
	}
	screen := driveScreen{redirectURI: request.RedirectURI, request: sealed}
	if f.google != nil {
		screen.google = f.google.AuthURL(sealed, request.Verifier, request.Nonce)
	}
	if err := f.pages.showDrive(w, screen); err != nil {
		f.events.failed(r.Context(), stepDrive, err)
	}
}

// backToGoogle sends a parent who pressed "Back to Google" on the page that
// asks for the Drive on to Google, with the request that page readied.
// Nothing is approved: that page names no client, and a client is approved
// on the consent screen alone.
func (f *flow) backToGoogle(w http.ResponseWriter, r *http.Request, request *flight, sealed string) {
	if f.google == nil {
		f.stopAt(w, r, stepConsent, stoppedUnconfigured, "unconfigured")
		return
	}
	f.events.consented(r.Context(), request, &ending{outcome: "again"})
	//nolint:gosec // Google's own address, carrying the request this server sealed
	http.Redirect(w, r, f.google.AuthURL(sealed, request.Verifier, request.Nonce), http.StatusSeeOther)
}

// cancelForDrive ends the sign-in of a parent who cancelled on the page that
// asks for the Drive: the client hears they declined, and the line says it
// was the Drive they declined.
func (f *flow) cancelForDrive(w http.ResponseWriter, r *http.Request, request *flight) {
	dropCSRFCookie(w)
	f.events.consented(r.Context(), request, &ending{outcome: "denied", reason: "no_drive"})
	f.sendBack(w, r, http.StatusSeeOther, request,
		refused("access_denied", "the parent did not allow MathTrail its own file in their Google Drive"))
}

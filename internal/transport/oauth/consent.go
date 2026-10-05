package oauthserver

import (
	"errors"
	"net/http"
)

// maxConsentForm is the largest consent form read: a sealed request and the
// parent's decision take a few kilobytes.
const maxConsentForm = 64 << 10

// consent answers the consent screen. The request comes back sealed in the
// form and is held to the browser it was begun in, so that a page on another
// site cannot post a parent's approval for them: that page has no way to the
// cookie. Allowing is remembered in this browser and sends the parent on to
// Google; declining ends the sign-in, and the client hears it was declined;
// a reviewer's password signs a directory's reviewer in as the demo account.
// The page that asks a parent back to Google for the Drive posts here too:
// going back sends the parent on to Google with nothing approved, and
// cancelling ends the sign-in as declining does.
func (f *flow) consent(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")

	r.Body = http.MaxBytesReader(w, r.Body, maxConsentForm)
	if err := r.ParseForm(); err != nil {
		var tooLarge *http.MaxBytesError
		if errors.As(err, &tooLarge) {
			f.stopAt(w, r, stepConsent, stoppedTooLarge, "too_large")
			return
		}
		f.stopAt(w, r, stepConsent, stoppedRequest, "invalid_request")
		return
	}
	sealed := r.PostForm.Get("request")
	request, ok := f.resume(w, r, stepConsent, sealed)
	if !ok {
		return
	}

	switch r.PostForm.Get("decision") {
	case "allow":
		if f.google == nil {
			f.stopAt(w, r, stepConsent, stoppedUnconfigured, "unconfigured")
			return
		}
		if err := f.approve(w, r, &request); err != nil {
			f.fail(w, r, stepConsent, err)
			return
		}
		f.events.consented(r.Context(), &request, &ending{outcome: "allowed"})
		//nolint:gosec // Google's own address, carrying the request this server sealed
		http.Redirect(w, r, f.google.AuthURL(sealed, request.Verifier, request.Nonce), http.StatusSeeOther)
	case "deny":
		dropCSRFCookie(w)
		f.events.consented(r.Context(), &request, &ending{outcome: "denied"})
		f.sendBack(w, r, http.StatusSeeOther, &request, refused("access_denied", "the parent did not allow the sign-in"))
	case "again":
		f.backToGoogle(w, r, &request, sealed)
	case "cancel":
		f.cancelForDrive(w, r, &request)
	case "reviewer":
		f.signInReviewer(w, r, &request)
	default:
		f.stopAt(w, r, stepConsent, stoppedRequest, "invalid_request")
	}
}

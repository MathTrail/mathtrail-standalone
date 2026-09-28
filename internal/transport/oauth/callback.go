package oauthserver

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"slices"

	"github.com/MathTrail/mathtrail-standalone/internal/infra/googleauth"
)

// ending is how a sign-in Google answered came to its end, as its line tells
// it: the outcome, why, whom the parent signed in as when they did, what went
// wrong when something did, and whether the fault is this server's own.
type ending struct {
	outcome string
	reason  string
	user    string
	cause   error
	ours    bool
}

// callback answers Google's redirect: the parent comes back with Google's code,
// or with Google's refusal. The request comes back sealed in the state, and it
// is held to the browser it was begun in before anything else is read. From
// then on the sign-in is over whatever happens: the cookie is taken back, and
// the client hears how it ended at its own address.
func (f *flow) callback(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")

	query, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil || len(query["state"]) > 1 {
		f.stopAt(w, r, stepCallback, stoppedRequest, "invalid_request")
		return
	}
	request, ok := f.resume(w, r, stepCallback, query.Get("state"))
	if !ok {
		return
	}
	dropCSRFCookie(w)

	answer, end := f.finish(r.Context(), &request, query)
	f.events.calledBack(r.Context(), &request, &end)
	f.sendBack(w, r, http.StatusFound, &request, answer)
}

// finish is how a sign-in Google answered ends: in a code for the client, or in
// the error the client hears instead. A parent who declined at Google, or who
// unticked the one file in their Drive the service keeps, has not allowed the
// sign-in; anything else that stops it is a failure, Google's or this
// server's.
func (f *flow) finish(ctx context.Context, request *flight, query url.Values) (url.Values, ending) {
	switch {
	case query.Get("error") == "access_denied":
		return refused("access_denied", "the parent did not allow the sign-in at Google"),
			ending{outcome: "denied", reason: "access_denied"}
	case query.Has("error"):
		return refused("server_error", "Google could not sign the parent in"),
			ending{outcome: "failed", reason: "google_error"}
	case query.Get("code") == "":
		return refused("server_error", "Google sent the parent back without a code"),
			ending{outcome: "failed", reason: "no_code"}
	case f.google == nil:
		return refused("server_error", "signing in is not set up on this server"),
			ending{outcome: "failed", reason: "unconfigured"}
	}

	grant, err := f.google.Exchange(ctx, query.Get("code"), request.Verifier, request.Nonce)
	if err != nil {
		return exchangeFailed(err)
	}
	if !slices.Contains(grant.Scopes, googleauth.ScopeDriveFile) {
		return refused("access_denied",
				"MathTrail needs to keep its own file in the parent's Google Drive: sign in again and allow it"),
			ending{outcome: "denied", reason: "no_drive"}
	}
	// The account's own identifier at Google goes no further than here.
	user := f.userID("google-sub:" + grant.Subject)
	code, err := f.issueCode(request, &grant, user)
	if err != nil {
		return refused("server_error", "the sign-in could not be finished: try again"),
			ending{outcome: "failed", reason: "internal", cause: err, ours: true}
	}
	return url.Values{"code": {code}}, ending{outcome: "ok", user: user}
}

// exchangeFailed is what the client hears when Google's code could not be
// exchanged for a sign-in, and why, for the line.
func exchangeFailed(err error) (url.Values, ending) {
	end := ending{outcome: "failed", reason: "google_unavailable", cause: err}
	description := "Google did not answer: try again"
	switch {
	case errors.Is(err, googleauth.ErrClient):
		// Google refused the service's own client: every sign-in fails the
		// same way until the deployment is fixed, so the line is an error of
		// ours.
		end.reason, end.ours = "client_refused", true
		description = "the sign-in could not be finished: try again later"
	case errors.Is(err, googleauth.ErrCodeRefused):
		end.reason, description = "code_refused", "Google refused the code of the sign-in: sign in again"
	case errors.Is(err, googleauth.ErrIdentity):
		end.reason, description = "identity", "Google's answer did not prove who signed in: sign in again"
	case errors.Is(err, googleauth.ErrNoRefresh):
		end.reason, description = "no_refresh", "Google gave no lasting access: sign in again"
	}
	return refused("server_error", description), end
}

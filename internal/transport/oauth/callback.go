package oauthserver

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"slices"

	"github.com/MathTrail/mathtrail-standalone/internal/infra/googleauth"
)

// ending is how a step of a sign-in came to its end, as its line tells it: the
// outcome, why, whom the parent signed in as when that is known, what went
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
// then on the sign-in is over whatever happens — the cookie is taken back, and
// the client hears how it ended at its own address — but for a parent who
// left Google's box for the Drive unticked, who is asked to go back and tick
// it.
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

	answer, end := f.finish(r.Context(), &request, query, f.countryOf(r))
	f.events.calledBack(r.Context(), &request, &end)
	if end.outcome == outcomeRetry {
		f.askForDrive(w, r, &request)
		return
	}
	dropCSRFCookie(w)
	f.sendBack(w, r, http.StatusFound, &request, answer)
}

// finish is how a sign-in Google answered ends: in a code for the client, or in
// the error the client hears instead — or in no answer yet and the outcome
// retry, when the parent left Google's box for the one file the service keeps
// in their Drive unticked. A parent who declined at Google has not allowed the
// sign-in; anything else that stops it is a failure, Google's or this
// server's. The country is the one the parent's browser came back from, which
// a finished sign-in carries on into the code.
func (f *flow) finish(ctx context.Context, request *flight, query url.Values, country string) (url.Values, ending) {
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

	grant, err := f.google.Exchange(ctx, query.Get("code"), request.Verifier)
	if err != nil {
		return exchangeFailed(err)
	}
	if !slices.Contains(grant.Scopes, googleauth.ScopeDriveFile) {
		// Asked for alone, the Drive is no box on Google's screen, and a grant
		// without it is not expected. One that comes all the same is a box
		// left unticked: the parent is asked again rather than the client
		// told. The grant is left as Google gave it: ending it at Google would
		// end every grant of the parent's at this service, the chats they
		// have already connected included.
		return nil, ending{outcome: outcomeRetry, reason: "no_drive"}
	}
	// The account's own identifier at Google goes no further than here.
	user := f.userID(googleAccount(grant.Subject))
	code, err := f.issueCode(request, &grant, user, country)
	if err != nil {
		return refused("server_error", "the sign-in could not be finished: try again"),
			ending{outcome: "failed", reason: "internal", cause: err, ours: true}
	}
	return url.Values{"code": {code}}, ending{outcome: "ok", user: user}
}

// googleAccount is an account at Google as its identifier here is derived
// from: the provider, then Google's own identifier for it. A parent's sign-in
// and the reviewers' derive the demo account's alike only while both name it
// so.
func googleAccount(subject string) string { return "google-sub:" + subject }

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

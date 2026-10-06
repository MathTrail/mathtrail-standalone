package oauthserver

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"errors"
	"net/http"
	"net/url"

	"github.com/MathTrail/mathtrail-standalone/internal/infra/googleauth"
)

// Reviewer is the sign-in of a directory's reviewers: on the consent screen, a
// password in place of Google, which signs them in as the demo account — the
// one account whose grant at Google the service holds. Google is not asked to
// sign anybody in, so a reviewer meets none of the checks it may put to a
// sign-in from another country or another device; and the password reaches
// nothing but the demo account's own file.
type Reviewer struct {
	// Password is what the reviewer types. A secret: it is compared, and
	// written nowhere.
	Password string
	// Subject is the demo account's own identifier at Google, which the
	// account the reviewer signs in as is derived from.
	Subject string
	// RefreshToken is the demo account's grant at Google, renewed for each
	// sign-in of a reviewer. A secret like the password.
	RefreshToken string
}

// reviewerSignIn is the reviewer's sign-in as the server holds it: the digest
// of the password, the account it signs in as — the identifier a sign-in is
// given now, first among every one the account's tokens may carry — and that
// account's grant.
type reviewerSignIn struct {
	password     [sha256.Size]byte
	users        []string
	refreshToken string
}

// newReviewerSignIn is the reviewer's sign-in, or nil when there is none. The
// demo account is known by an identifier of each key of the ring, as a
// parent's account is, so that a reviewer signed in before a rotation of the
// keys is still the demo account's.
func newReviewerSignIn(reviewer *Reviewer, userIDs func(account string) []string) *reviewerSignIn {
	if reviewer == nil {
		return nil
	}
	return &reviewerSignIn{
		password:     sha256.Sum256([]byte(reviewer.Password)),
		users:        userIDs(googleAccount(reviewer.Subject)),
		refreshToken: reviewer.RefreshToken,
	}
}

// googleAccount is the demo account as its identifier here is derived from:
// the provider, then Google's own identifier for the account, which the grant
// the reviewers' sign-in renews names. A parent's account is derived from its
// identifier at Drive instead, so the demo account signed in through Google is
// not known as the demo account, and ending that sign-in ends its grants at
// Google, the reviewers' among them.
func googleAccount(subject string) string { return "google-sub:" + subject }

// demoUsers are the identifiers the demo account is known by, or none when
// there is no reviewer's sign-in.
func demoUsers(reviewer *reviewerSignIn) []string {
	if reviewer == nil {
		return nil
	}
	return reviewer.users
}

// admits reports whether a password typed on the consent screen is the
// reviewer's, in the same time whether it is or not: the two are compared as
// digests of one length, so that neither the letters typed nor how many they
// are shows in how long the answer takes.
func (r *reviewerSignIn) admits(typed string) bool {
	digest := sha256.Sum256([]byte(typed))
	return subtle.ConstantTimeCompare(digest[:], r.password[:]) == 1
}

// signInReviewer answers the reviewer's password on the consent screen. A
// password that is not the reviewer's stops at a page: Google is not asked,
// and the client hears nothing. The reviewer's own is worth the code a sign-in
// through Google ends in, for the demo account, with a Google access token
// renewed for it. The code carries no country, since a reviewer's network
// says nothing of where any family is, and the browser remembers no approval,
// since a reviewer who comes back is to be asked for the password again
// rather than sent on to Google.
func (f *flow) signInReviewer(w http.ResponseWriter, r *http.Request, request *flight) {
	switch {
	case f.reviewer == nil:
		f.stopAt(w, r, stepConsent, stoppedRequest, "invalid_request")
		return
	case !f.reviewer.admits(r.PostForm.Get("password")):
		f.stopAt(w, r, stepConsent, stoppedPassword, "reviewer_password")
		return
	}
	dropCSRFCookie(w)
	answer, end := f.reviewerCode(r.Context(), request)
	f.events.consented(r.Context(), request, &end)
	f.sendBack(w, r, http.StatusSeeOther, request, answer)
}

// reviewerCode is the code the reviewer's sign-in ends in, or the error the
// client hears instead, and how the line tells it. The demo account's grant is
// renewed at the pace of the account, as every grant is: the password stands
// in front of it, but a password that went round would otherwise have the
// service call Google with every guess of somebody's that it lets through.
func (f *flow) reviewerCode(ctx context.Context, request *flight) (url.Values, ending) {
	user := f.reviewer.users[0]
	if verdict := f.renewals.Take(user); !verdict.Allowed {
		if verdict.Began {
			f.events.limited(ctx, limitRenewals, user)
		}
		return refused("temporarily_unavailable", "too many sign-ins of the reviewer at once: try again in a minute"),
			ending{outcome: "failed", reason: limitRenewals}
	}
	renewal, err := f.google.Refresh(ctx, f.reviewer.refreshToken)
	if err != nil {
		return reviewerFailed(err)
	}
	grant := googleauth.Grant{AccessToken: renewal.AccessToken, Expiry: renewal.Expiry, RefreshToken: renewal.RefreshToken}
	code, err := f.issueCode(request, &grant, user, "")
	if err != nil {
		return refused("server_error", "the sign-in could not be finished: try again"),
			ending{outcome: "failed", reason: "internal", cause: err, ours: true}
	}
	return url.Values{"code": {code}}, ending{outcome: "reviewer", user: user}
}

// reviewerFailed is what the client hears when the demo account's grant could
// not be renewed, and how the line tells it. A grant Google no longer honours
// is the deployment's to mend — it has to be captured again — so its line is
// an error of ours, as Google refusing the service's own client is.
func reviewerFailed(err error) (url.Values, ending) {
	switch {
	case errors.Is(err, googleauth.ErrGrantEnded):
		return refused("server_error", "the reviewer's account cannot be reached: tell whoever gave you the password"),
			ending{outcome: "failed", reason: "grant_ended", cause: err, ours: true}
	case errors.Is(err, googleauth.ErrClient):
		return refused("server_error", "the sign-in could not be finished: try again later"),
			ending{outcome: "failed", reason: "client_refused", cause: err, ours: true}
	case errors.Is(err, googleauth.ErrUnavailable):
		return refused("temporarily_unavailable", "Google did not answer: try again"),
			ending{outcome: "failed", reason: "google_unavailable", cause: err}
	}
	return refused("server_error", "the sign-in could not be finished: try again later"),
		ending{outcome: "failed", reason: "internal", cause: err, ours: true}
}

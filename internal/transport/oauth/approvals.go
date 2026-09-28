package oauthserver

import (
	"cmp"
	"encoding/json"
	"fmt"
	"net/http"
	"slices"
	"time"
)

const (
	// consentCookie remembers which clients a browser's parent has approved.
	consentCookie = "mt_consent"
	// approvalLifetime is how long an approval is remembered: approving the
	// same host in the same browser twice a year is no burden.
	approvalLifetime = 180 * 24 * time.Hour
	// maxApprovals is how many approvals the cookie keeps, the oldest giving
	// way to a new one: a family uses a few hosts, and a cookie has room for
	// a few kilobytes.
	maxApprovals = 20
)

// approval is one client a parent approved in this browser: the fingerprint
// of the client with the address it sends the parent back to, and when.
type approval struct {
	Fingerprint string `json:"f"`
	ApprovedAt  int64  `json:"t"`
}

// fingerprint is what an approval is of: the client together with the address
// it sends the parent back to, since the same client sending the parent
// somewhere else is somebody else to approve.
func fingerprint(request *flight) string {
	return digestOf(request.Client, request.RedirectURI)
}

// approvals are the approvals the browser holds that are still good. A cookie
// this server did not seal, for itself, as approvals, holds none.
func (f *flow) approvals(r *http.Request) []approval {
	cookie, err := r.Cookie(consentCookie)
	if err != nil {
		return nil
	}
	plain, err := f.consents.Open(cookie.Value, f.issuer)
	if err != nil {
		return nil
	}
	var held []approval
	if err := json.Unmarshal(plain, &held); err != nil {
		return nil
	}
	now := f.now()
	return slices.DeleteFunc(held, func(a approval) bool {
		age := now.Sub(time.Unix(a.ApprovedAt, 0))
		return age > approvalLifetime || age < -clockSkew
	})
}

// approved reports whether the browser's parent has approved this request's
// client, sending the parent back to this request's address.
func (f *flow) approved(r *http.Request, request *flight) bool {
	wanted := fingerprint(request)
	return slices.ContainsFunc(f.approvals(r), func(a approval) bool { return a.Fingerprint == wanted })
}

// approve adds the request's client to the approvals the browser holds, as
// the newest, and keeps the newest ones the cookie has room for.
func (f *flow) approve(w http.ResponseWriter, r *http.Request, request *flight) error {
	added := fingerprint(request)
	held := slices.DeleteFunc(f.approvals(r), func(a approval) bool { return a.Fingerprint == added })
	held = append(held, approval{Fingerprint: added, ApprovedAt: f.now().Unix()})
	slices.SortStableFunc(held, func(a, b approval) int { return cmp.Compare(a.ApprovedAt, b.ApprovedAt) })
	held = held[max(0, len(held)-maxApprovals):]

	plain, err := json.Marshal(held)
	if err != nil {
		return fmt.Errorf("oauth: encode the approvals: %w", err)
	}
	sealed, err := f.consents.Seal(plain, f.issuer)
	if err != nil {
		return fmt.Errorf("oauth: seal the approvals: %w", err)
	}
	http.SetCookie(w, &http.Cookie{
		Name:     consentCookie,
		Value:    sealed,
		Path:     cookiePath,
		MaxAge:   int(approvalLifetime / time.Second),
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
	return nil
}

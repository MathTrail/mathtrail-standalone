package oauthserver

import (
	"encoding/json"
	"fmt"

	"github.com/MathTrail/mathtrail-standalone/internal/infra/googleauth"
)

// grantCode is what an authorization code carries: whom the parent signed in
// as, Google's tokens for them, and everything the exchange of the code has
// to match — the client, the address the code was sent to, the PKCE
// challenge, the resource and the scope. It is sealed as a code, bound to this
// issuer, and kept nowhere: the code is the whole of it.
//
// It carries the digest of the client's identifier, as the request did: the
// exchange compares an identifier with it, and the code rides back to the
// client in an address.
type grantCode struct {
	User               string `json:"user"`
	Client             string `json:"client"`
	RedirectURI        string `json:"redirect_uri"`
	Challenge          string `json:"code_challenge"`
	Resource           string `json:"resource"`
	Scope              string `json:"scope"`
	GoogleAccessToken  string `json:"google_access_token"`
	GoogleAccessExpiry int64  `json:"google_access_expiry"`
	GoogleRefreshToken string `json:"google_refresh_token"`
	IssuedAt           int64  `json:"issued_at"`
}

// issueCode seals the code a finished sign-in hands the client.
func (f *flow) issueCode(request *flight, grant *googleauth.Grant, user string) (string, error) {
	plain, err := json.Marshal(grantCode{
		User:               user,
		Client:             request.Client,
		RedirectURI:        request.RedirectURI,
		Challenge:          request.Challenge,
		Resource:           request.Resource,
		Scope:              request.Scope,
		GoogleAccessToken:  grant.AccessToken,
		GoogleAccessExpiry: grant.Expiry.Unix(),
		GoogleRefreshToken: grant.RefreshToken,
		IssuedAt:           f.now().Unix(),
	})
	if err != nil {
		return "", fmt.Errorf("oauth: encode the code: %w", err)
	}
	code, err := f.codes.Seal(plain, f.issuer)
	if err != nil {
		return "", fmt.Errorf("oauth: seal the code: %w", err)
	}
	return code, nil
}

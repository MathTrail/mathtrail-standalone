package googleauth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"golang.org/x/oauth2"
)

// maxRefusal is the most of a refusal read: an error code, and a sentence
// about it.
const maxRefusal = 4 << 10

// Refresh asks Google for a new access token with the grant's refresh token.
// A refresh token Google no longer honours is the grant ended. Google does not
// usually answer with another refresh token, and when it does, the grant goes
// on with that one.
func (g *google) Refresh(ctx context.Context, refreshToken string) (Renewal, error) {
	ctx, cancel := context.WithTimeout(ctx, callTimeout)
	defer cancel()

	// A token that holds no access token is renewed at once.
	source := g.oauth.TokenSource(context.WithValue(ctx, oauth2.HTTPClient, g.client),
		&oauth2.Token{RefreshToken: refreshToken})
	token, err := source.Token()
	if err != nil {
		return Renewal{}, refusalOf(err, ErrGrantEnded)
	}
	// Counted on this clock from the moment the answer came, as the exchange
	// counts it.
	if token.ExpiresIn <= 0 {
		return Renewal{}, fmt.Errorf("%w: an access token with no lifetime", ErrUnavailable)
	}
	return Renewal{
		AccessToken:  token.AccessToken,
		Expiry:       g.now().Add(time.Duration(token.ExpiresIn) * time.Second),
		RefreshToken: token.RefreshToken,
	}, nil
}

// Revoke ends a grant at Google. Google answers 200 when it ended the grant,
// and refuses a token it no longer honours as invalid_token, which ends
// nothing. Google failing, or asking for a pause, is Google not answering; any
// other refusal is a request Google could not read, and says which.
func (g *google) Revoke(ctx context.Context, token string) error {
	ctx, cancel := context.WithTimeout(ctx, callTimeout)
	defer cancel()

	// The token goes in the body rather than in the address, which the hops on
	// the way may keep.
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, g.revokeURL,
		strings.NewReader(url.Values{"token": {token}}.Encode()))
	if err != nil {
		return fmt.Errorf("googleauth: revoke: %w", err)
	}
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	response, err := g.client.Do(request)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrUnavailable, err)
	}
	defer func() { _ = response.Body.Close() }()

	status := response.StatusCode
	switch {
	case status == http.StatusOK:
		return nil
	case status >= http.StatusInternalServerError, status == http.StatusTooManyRequests:
		return fmt.Errorf("%w: status %d", ErrUnavailable, status)
	}
	var refused struct {
		Error string `json:"error"`
	}
	// An answer that is not JSON leaves the code empty, and is refused below.
	_ = json.NewDecoder(io.LimitReader(response.Body, maxRefusal)).Decode(&refused)
	if status == http.StatusBadRequest && refused.Error == "invalid_token" {
		return ErrNotHonoured
	}
	return fmt.Errorf("googleauth: Google refused to revoke: status %d, %q", status, refused.Error)
}

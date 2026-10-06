package googleauth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"slices"
	"strings"
	"time"

	"golang.org/x/oauth2"

	"github.com/MathTrail/mathtrail-standalone/internal/infra/drive"
)

// maxAbout is the most of Drive's word about an account that is read: one
// short field.
const maxAbout = 4 << 10

// aboutFields is all Drive is asked of the account an access token reaches it
// as: the identifier its permissions know the account by. Drive refuses a
// question of its about resource that does not say what it asks.
const aboutFields = "user(permissionId)"

// Exchange trades Google's code for the grant, and accepts the grant only
// once it reaches the parent's Drive and Drive has said whose it is.
func (g *google) Exchange(ctx context.Context, code, verifier string) (Grant, error) {
	ctx, cancel := context.WithTimeout(ctx, callTimeout)
	defer cancel()

	// The lifetime is counted on this clock from the moment the request is
	// sent. Google counts it from the moment it gave the token, which comes
	// after, so a grant is never taken to last longer than it does.
	sent := g.now()
	token, err := g.oauth.Exchange(context.WithValue(ctx, oauth2.HTTPClient, g.client), code,
		oauth2.VerifierOption(verifier))
	if err != nil {
		return Grant{}, refusalOf(err, ErrCodeRefused)
	}
	if token.ExpiresIn <= 0 {
		return Grant{}, fmt.Errorf("%w: an access token with no lifetime", ErrUnavailable)
	}
	expiry := sent.Add(time.Duration(token.ExpiresIn) * time.Second)
	// Drive is not asked about a grant that would be refused anyway, and
	// cannot be about one that does not reach it. A missing scope is no
	// scope granted.
	if token.RefreshToken == "" {
		return Grant{}, ErrNoRefresh
	}
	scope, _ := token.Extra("scope").(string)
	if !slices.Contains(strings.Fields(scope), ScopeDriveFile) {
		return Grant{}, ErrNoDrive
	}

	permissionID, err := g.driveUser(ctx, token.AccessToken)
	if err != nil {
		return Grant{}, err
	}
	return Grant{
		PermissionID: permissionID,
		AccessToken:  token.AccessToken,
		Expiry:       expiry,
		RefreshToken: token.RefreshToken,
	}, nil
}

// about is what Drive says of the account an access token reaches it as, as
// far as the sign-in reads it.
type about struct {
	User struct {
		PermissionID string `json:"permissionId"`
	} `json:"user"`
}

// driveUser is who signed in: the account Drive says the access token reaches
// it as, by its permission ID. An answer that is not Drive's word about an
// account is Google not answering.
func (g *google) driveUser(ctx context.Context, accessToken string) (string, error) {
	address := g.aboutURL + "?" + url.Values{"fields": {aboutFields}}.Encode()
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, address, http.NoBody)
	if err != nil {
		return "", fmt.Errorf("googleauth: ask whose Drive the token reaches: %w", err)
	}
	// The token goes in a header rather than in the address, which the hops
	// on the way may keep and a failed call repeats in its error.
	request.Header.Set("Authorization", "Bearer "+accessToken)
	response, err := g.client.Do(request)
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrUnavailable, err)
	}
	defer func() { _ = response.Body.Close() }()

	if response.StatusCode != http.StatusOK {
		return "", driveRefused(drive.RefusalOf(response))
	}
	var answer about
	if err := json.NewDecoder(io.LimitReader(response.Body, maxAbout)).Decode(&answer); err != nil {
		return "", fmt.Errorf("%w: %w", ErrUnavailable, err)
	}
	if answer.User.PermissionID == "" {
		return "", fmt.Errorf("%w: no permission ID", ErrIdentity)
	}
	return answer.User.PermissionID, nil
}

// driveRefused is what a refusal of the question of whose Drive a token
// reaches means, told as Drive's refusals are. Drive failing, or asking for a
// pause, is Google not answering. Any other refusal is Drive not saying who
// signed in: a token Google does not honour, one that grants nothing of Drive,
// or an account Drive is kept from, as a domain's administrator may keep it.
func driveRefused(refusal error) error {
	if errors.Is(refusal, drive.ErrRateLimited) || errors.Is(refusal, drive.ErrUnavailable) {
		return fmt.Errorf("%w: %w", ErrUnavailable, refusal)
	}
	return fmt.Errorf("%w: %w", ErrIdentity, refusal)
}

// refusalOf is what a failed call to the token endpoint means. Google failing,
// or asking for a pause, is Google not answering, as is anything that is no
// answer at all. A refusal of the client is the service's own client refused.
// Any other refusal is Google's verdict on what the call was made with — the
// code, or the refresh token — and is the verdict given. What Google explained
// in its own words is left out, since nobody but a developer could use it and
// the error code says what happened.
func refusalOf(err, verdict error) error {
	var refused *oauth2.RetrieveError
	if !errors.As(err, &refused) {
		return fmt.Errorf("%w: %w", ErrUnavailable, err)
	}
	status := 0
	if refused.Response != nil {
		status = refused.Response.StatusCode
	}
	switch {
	case status >= http.StatusInternalServerError, status == http.StatusTooManyRequests:
		return fmt.Errorf("%w: status %d", ErrUnavailable, status)
	case status == http.StatusUnauthorized, refused.ErrorCode == "invalid_client", refused.ErrorCode == "unauthorized_client":
		return fmt.Errorf("%w: status %d", ErrClient, status)
	case refused.ErrorCode == "":
		return verdict
	default:
		return fmt.Errorf("%w: %s", verdict, refused.ErrorCode)
	}
}

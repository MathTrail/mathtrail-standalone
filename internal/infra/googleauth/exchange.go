package googleauth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"golang.org/x/oauth2"
)

// maxTokenInfo is the most of Google's word about an access token that is
// read: a handful of short fields.
const maxTokenInfo = 4 << 10

// Exchange trades Google's code for the grant, and accepts the grant only
// once Google has said whom its access token was issued for.
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
	// Google is not asked about a grant that would be refused anyway.
	if token.RefreshToken == "" {
		return Grant{}, ErrNoRefresh
	}

	subject, err := g.identity(ctx, token.AccessToken)
	if err != nil {
		return Grant{}, err
	}
	// A missing scope is no scope granted, which leaves the caller nothing to
	// accept.
	scope, _ := token.Extra("scope").(string)
	return Grant{
		Subject:      subject,
		AccessToken:  token.AccessToken,
		Expiry:       expiry,
		RefreshToken: token.RefreshToken,
		Scopes:       strings.Fields(scope),
	}, nil
}

// tokenInfo is what Google says of an access token, as far as the sign-in
// reads it: the client the token was issued to, and the account it was issued
// for.
type tokenInfo struct {
	Audience string `json:"aud"`
	Subject  string `json:"sub"`
}

// identity is who signed in: the account Google says the access token was
// issued for, once Google has said it was issued to this client. Google
// failing, or asking for a pause, is Google not answering, and so is an
// answer that is not its word about a token; any other refusal is Google
// calling the token none of its own.
func (g *google) identity(ctx context.Context, accessToken string) (string, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, g.tokenInfoURL, http.NoBody)
	if err != nil {
		return "", fmt.Errorf("googleauth: ask whom the token is for: %w", err)
	}
	// The token goes in a header rather than in the address, which the hops
	// on the way may keep and a failed call repeats in its error. The
	// question is put as Google's own client library puts it: a form posted
	// with nothing in it but that header.
	request.Header.Set("Authorization", "Bearer "+accessToken)
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	response, err := g.client.Do(request)
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrUnavailable, err)
	}
	defer func() { _ = response.Body.Close() }()

	switch status := response.StatusCode; {
	case status >= http.StatusInternalServerError, status == http.StatusTooManyRequests:
		return "", fmt.Errorf("%w: status %d", ErrUnavailable, status)
	case status != http.StatusOK:
		return "", fmt.Errorf("%w: status %d", ErrIdentity, status)
	}
	var info tokenInfo
	if err := json.NewDecoder(io.LimitReader(response.Body, maxTokenInfo)).Decode(&info); err != nil {
		return "", fmt.Errorf("%w: %w", ErrUnavailable, err)
	}
	switch {
	case info.Audience != g.oauth.ClientID:
		return "", fmt.Errorf("%w: the token was issued to another client", ErrIdentity)
	case info.Subject == "":
		return "", fmt.Errorf("%w: no subject", ErrIdentity)
	}
	return info.Subject, nil
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

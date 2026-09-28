package googleauth

import (
	"context"
	"crypto/subtle"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"golang.org/x/oauth2"
)

// Exchange trades Google's code for the grant, and accepts the grant only
// with an ID token that proves who signed in and belongs to this request.
func (g *google) Exchange(ctx context.Context, code, verifier, nonce string) (Grant, error) {
	ctx, cancel := context.WithTimeout(ctx, callTimeout)
	defer cancel()

	token, err := g.oauth.Exchange(context.WithValue(ctx, oauth2.HTTPClient, g.client), code,
		oauth2.VerifierOption(verifier))
	if err != nil {
		return Grant{}, refusalOf(err)
	}
	// The lifetime is counted on this clock from the moment the answer came,
	// as Google counts it from the moment it gave it.
	if token.ExpiresIn <= 0 {
		return Grant{}, fmt.Errorf("%w: an access token with no lifetime", ErrUnavailable)
	}
	expiry := g.now().Add(time.Duration(token.ExpiresIn) * time.Second)

	subject, err := g.identity(ctx, token, nonce)
	if err != nil {
		return Grant{}, err
	}
	if token.RefreshToken == "" {
		return Grant{}, ErrNoRefresh
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

// identity is who signed in, read from the ID token that came with the grant
// once the token is proven Google's, for this client, still good, and issued
// for this very request: its nonce is the one the request was sent with.
func (g *google) identity(ctx context.Context, token *oauth2.Token, nonce string) (string, error) {
	// A missing ID token is read as an empty one, which the verifier refuses
	// like any other that is not a signed token.
	raw, _ := token.Extra("id_token").(string)
	verified, err := g.verifier.Verify(ctx, raw)
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrIdentity, err)
	}
	if subtle.ConstantTimeCompare([]byte(verified.Nonce), []byte(nonce)) != 1 {
		return "", fmt.Errorf("%w: the nonce is not this request's", ErrIdentity)
	}
	if verified.Subject == "" {
		return "", fmt.Errorf("%w: no subject", ErrIdentity)
	}
	return verified.Subject, nil
}

// refusalOf is what a failed exchange means. Google failing, or asking for a
// pause, is Google not answering, as is anything that is no answer at all. A
// refusal of the client is the service's own client refused. Any other refusal
// is Google's verdict on the code. What Google explained in its own words is
// left out, since nobody but a developer could use it and the error code says
// what happened.
func refusalOf(err error) error {
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
		return ErrCodeRefused
	default:
		return fmt.Errorf("%w: %s", ErrCodeRefused, refused.ErrorCode)
	}
}

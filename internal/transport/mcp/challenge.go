package mcpserver

import (
	"context"
	"fmt"
	"net/http"
	"sync/atomic"
)

// challenge is what a sign-in answers a request that has to sign in again
// with — the value of a WWW-Authenticate header — and whether a call of the
// request found the parent's Google access gone and asked for it.
type challenge struct {
	header string
	raised atomic.Bool
}

// raise asks for the request to be answered 401 with the challenge. It
// changes nothing once the answer has started.
func (c *challenge) raise() { c.raised.Store(true) }

// challengeKey holds a request's challenge. It is a type of its own, so that
// nothing else a context carries can be taken for it.
type challengeKey struct{}

// withChallenge is how a sign-in hands a request the challenge a call of it
// raises when it finds the parent's Google access gone.
func withChallenge(ctx context.Context, asked *challenge) context.Context {
	return context.WithValue(ctx, challengeKey{}, asked)
}

// challengeFrom is the challenge a sign-in put in ctx, if one did.
func challengeFrom(ctx context.Context) (*challenge, bool) {
	asked, given := ctx.Value(challengeKey{}).(*challenge)
	return asked, given
}

// signInChallenge is the challenge of a request whose Google access is gone:
// the one a request with no token is answered with, and why the token it
// carried no longer does.
func signInChallenge(resourceMetadataURL string) string {
	return fmt.Sprintf(`Bearer resource_metadata=%q, scope=%q, error="invalid_token", error_description="The access to Google Drive has ended; sign in again."`,
		resourceMetadataURL, Scope)
}

// challenging is a response writer that answers a request 401, with the
// challenge, when a call of it raised the challenge before the answer
// started. The protocol library sends the status of a call with the call's
// result and nothing before it, so the result of the call that found the
// access gone is where the answer starts.
type challenging struct {
	http.ResponseWriter
	challenge *challenge
	started   bool
}

// WriteHeader sends the status: a 401 with the challenge in place of a 200
// when the challenge was raised.
func (w *challenging) WriteHeader(status int) {
	w.ResponseWriter.WriteHeader(w.statusOf(status))
}

// Write sends part of the body, the status first if nothing has been sent.
func (w *challenging) Write(body []byte) (int, error) {
	if !w.started {
		w.WriteHeader(http.StatusOK)
	}
	return w.ResponseWriter.Write(body)
}

// Flush sends what has been written so far: the status too, if nothing has
// been sent.
func (w *challenging) Flush() {
	if !w.started {
		w.WriteHeader(http.StatusOK)
	}
	_ = http.NewResponseController(w.ResponseWriter).Flush()
}

// Unwrap is the writer underneath, for anything that looks past this one.
func (w *challenging) Unwrap() http.ResponseWriter { return w.ResponseWriter }

// statusOf is the status an answer starts with, decided once, as it starts.
func (w *challenging) statusOf(status int) int {
	if w.started {
		return status
	}
	w.started = true
	if status == http.StatusOK && w.challenge.raised.Load() {
		w.Header().Set("WWW-Authenticate", w.challenge.header)
		return http.StatusUnauthorized
	}
	return status
}

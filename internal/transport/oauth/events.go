package oauthserver

import (
	"context"
	"errors"
	"net/url"
	"slices"
	"time"

	"go.uber.org/zap"

	"github.com/MathTrail/mathtrail-standalone/internal/infra/cimd"
	"github.com/MathTrail/mathtrail-standalone/internal/logger"
	"github.com/MathTrail/mathtrail-standalone/internal/telemetry"
)

// The events of a sign-in, as the lines about them are named.
const (
	eventCIMDFetch     = "cimd_fetch"
	eventAuthRegister  = "auth_register"
	eventAuthAuthorize = "auth_authorize"
	eventAuthConsent   = "auth_consent"
	eventAuthCallback  = "auth_callback"
	eventAuthReject    = "auth_reject"
)

// signInLog writes a line for each step of a sign-in, from fields named one by
// one. A line never carries what a client sent in its own words — its name,
// the addresses it listed, the body of its request — and never a token, a
// code, a secret or a sealed value: only which host, which way, and how it
// ended.
type signInLog struct {
	logger    *zap.Logger
	projectID string
}

// fetched leaves the line of a fetch of a client's document: the host it was
// asked of, whether a kept copy answered, how long it took, and how it ended.
func (l *signInLog) fetched(ctx context.Context, clientID string, cached bool, took time.Duration, err error) {
	host := ""
	if !errors.Is(err, cimd.ErrClientURL) {
		host = hostOf(clientID)
	}
	l.write(ctx, eventCIMDFetch,
		zap.String("host", host),
		zap.Bool("cached", cached),
		zap.Int64("duration_ms", took.Milliseconds()),
		zap.String("outcome", fetchOutcome(err)),
	)
}

// registered leaves the line of a registration made.
func (l *signInLog) registered(ctx context.Context, redirectHost string) {
	l.write(ctx, eventAuthRegister,
		zap.String("registration", registrationDCR),
		zap.String("redirect_host", redirectHost),
		zap.String("outcome", "ok"),
	)
}

// registrationRefused leaves the line of a registration refused, and why.
func (l *signInLog) registrationRefused(ctx context.Context, reason string) {
	l.write(ctx, eventAuthRegister,
		zap.String("registration", registrationDCR),
		zap.String("outcome", "refused"),
		zap.String("reason", reason),
	)
}

// registrationFailed leaves the line of a registration that failed on this
// side: an error, since it is ours.
func (l *signInLog) registrationFailed(ctx context.Context, err error) {
	l.logger.Error(eventAuthRegister, slices.Concat(
		[]zap.Field{
			zap.String("registration", registrationDCR),
			zap.String("outcome", "failed"),
			zap.Error(err),
		},
		callerFields(ctx, l.projectID),
	)...)
}

// authorized leaves the line of an authorization request answered: with the
// consent screen, on to Google, or back to the client refused, and why.
func (l *signInLog) authorized(ctx context.Context, request *flight, query url.Values, outcome, reason string) {
	fields := requestFields(request)
	fields = append(fields,
		zap.String("resource", resourceWord(query)),
		zap.String("scope", scopeWord(query, request.Scope)),
		zap.String("outcome", outcome),
	)
	if reason != "" {
		fields = append(fields, zap.String("reason", reason))
	}
	l.write(ctx, eventAuthAuthorize, fields...)
}

// consented leaves the line of the parent's answer on the consent screen.
func (l *signInLog) consented(ctx context.Context, request *flight, outcome string) {
	fields := requestFields(request)
	l.write(ctx, eventAuthConsent, append(fields, zap.String("outcome", outcome))...)
}

// calledBack leaves the line of a sign-in's end, with what went wrong when
// something did: a warning when Google failed it, an error when this server
// did.
func (l *signInLog) calledBack(ctx context.Context, request *flight, end *ending) {
	fields := requestFields(request)
	fields = append(fields, zap.String("outcome", end.outcome))
	if end.reason != "" {
		fields = append(fields, zap.String("reason", end.reason))
	}
	if end.user != "" {
		fields = append(fields, zap.String("user", end.user))
	}
	if end.cause != nil {
		fields = append(fields, zap.Error(end.cause))
	}
	fields = slices.Concat(fields, callerFields(ctx, l.projectID))
	switch {
	case end.ours:
		l.logger.Error(eventAuthCallback, fields...)
	case end.outcome == "failed":
		l.logger.Warn(eventAuthCallback, fields...)
	default:
		l.logger.Info(eventAuthCallback, fields...)
	}
}

// rejected leaves the line of a step of a sign-in stopped at a page, and why:
// an error, telling what failed, when the page could not be drawn either.
func (l *signInLog) rejected(ctx context.Context, step, reason string, err error) {
	fields := []zap.Field{zap.String("step", step), zap.String("reason", reason)}
	if err != nil {
		l.logger.Error(eventAuthReject, slices.Concat(fields, []zap.Field{zap.Error(err)}, callerFields(ctx, l.projectID))...)
		return
	}
	l.write(ctx, eventAuthReject, fields...)
}

// failed leaves the line of a step of a sign-in this server failed: an error,
// since it is ours.
func (l *signInLog) failed(ctx context.Context, step string, err error) {
	l.rejected(ctx, step, "internal", err)
}

func (l *signInLog) write(ctx context.Context, event string, fields ...zap.Field) {
	l.logger.Info(event, slices.Concat(fields, callerFields(ctx, l.projectID))...)
}

// callerFields are what every line of a request carries: its id, and the
// trace it belongs to when there is one.
func callerFields(ctx context.Context, projectID string) []zap.Field {
	return append([]zap.Field{zap.String("request_id", logger.RequestID(ctx))}, telemetry.LogFields(ctx, projectID)...)
}

// fetchOutcome is how a fetch of a client's document ended, in the closed list
// of words its line uses.
func fetchOutcome(err error) string {
	switch {
	case err == nil:
		return "ok"
	case errors.Is(err, cimd.ErrClientURL):
		return "refused_url"
	case errors.Is(err, cimd.ErrAddress):
		return "refused_address"
	case errors.Is(err, cimd.ErrRedirect):
		return "redirect"
	case errors.Is(err, cimd.ErrStatus):
		return "status"
	case errors.Is(err, cimd.ErrTooLarge):
		return "too_large"
	case errors.Is(err, cimd.ErrDocument):
		return "invalid_document"
	default:
		return "unreachable"
	}
}

// requestFields are what every line of a sign-in under way says of it: how
// the client is known, and the host the parent goes back to.
func requestFields(request *flight) []zap.Field {
	return []zap.Field{
		zap.String("registration", request.Registration),
		zap.String("redirect_host", hostOf(request.RedirectURI)),
	}
}

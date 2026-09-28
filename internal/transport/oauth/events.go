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
	eventCIMDFetch    = "cimd_fetch"
	eventAuthRegister = "auth_register"
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

// hostOf is the host of an address, for a line that names where a client is
// or where it sends the parent back to, and nothing else of it.
func hostOf(uri string) string {
	address, err := url.Parse(uri)
	if err != nil {
		return ""
	}
	return address.Hostname()
}

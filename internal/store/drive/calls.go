package drivestore

import (
	"context"
	"errors"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"

	"github.com/MathTrail/mathtrail-standalone/internal/infra/drive"
	"github.com/MathTrail/mathtrail-standalone/internal/logger"
	"github.com/MathTrail/mathtrail-standalone/internal/store"
	"github.com/MathTrail/mathtrail-standalone/internal/telemetry"
)

// scope is what the spans of the calls to Drive are recorded under, so that a
// reader can tell them from the ones a library produced.
const scope = "github.com/MathTrail/mathtrail-standalone/internal/store/drive"

// eventDriveCall is the line every call to Drive leaves.
const eventDriveCall = "drive_call"

// The calls made to Drive, as a span and a line name them.
const (
	opList     = "list"
	opGet      = "get"
	opDownload = "download"
	opCreate   = "create"
	opUpdate   = "update"
)

// How a call to Drive ended, as a span and a line say it: a closed list, never
// the text of an error.
const (
	outcomeOK          = "ok"
	outcomeNotFound    = "not_found"
	outcomeRevoked     = "revoked"
	outcomeRateLimited = "rate_limited"
	outcomeStorageFull = "storage_full"
	outcomeTooLarge    = "too_large"
	outcomeUnavailable = "unavailable"
	outcomeTimeout     = "timeout"
	outcomeCanceled    = "canceled"
	outcomeFailed      = "failed"
)

// parentsDrive is the Drive of the parent an account belongs to, as the store
// reaches it: every call made with the account's token, and every one watched.
type parentsDrive struct {
	files   drive.Files
	account store.Account
	watch   *watch
}

// driveOf is the Drive the account's token reaches.
func (s *driveStore) driveOf(account store.Account) parentsDrive {
	return parentsDrive{files: s.files, account: account, watch: s.watch}
}

// find is the service's own files in the Drive that carry the marker.
func (p parentsDrive) find(ctx context.Context, marker string) ([]drive.File, error) {
	var found []drive.File
	err := p.watch.call(ctx, p.account, opList, func(ctx context.Context) (err error) {
		found, err = p.files.List(ctx, p.account.Token(), drive.Query{Key: markerKey, Value: marker})
		return err
	})
	return found, err
}

// get is what a file is called.
func (p parentsDrive) get(ctx context.Context, id string) (drive.File, error) {
	var file drive.File
	err := p.watch.call(ctx, p.account, opGet, func(ctx context.Context) (err error) {
		file, err = p.files.Get(ctx, p.account.Token(), id)
		return err
	})
	return file, err
}

// download is what a file holds, as far as a profile ever runs.
func (p parentsDrive) download(ctx context.Context, id string) ([]byte, error) {
	var raw []byte
	err := p.watch.call(ctx, p.account, opDownload, func(ctx context.Context) (err error) {
		raw, err = p.files.Download(ctx, p.account.Token(), id, maxFile)
		return err
	})
	return raw, err
}

// create makes a file, or a folder when there is no content.
func (p parentsDrive) create(ctx context.Context, meta *drive.File, content []byte) (drive.File, error) {
	var made drive.File
	err := p.watch.call(ctx, p.account, opCreate, func(ctx context.Context) (err error) {
		made, err = p.files.Create(ctx, p.account.Token(), meta, content)
		return err
	})
	return made, err
}

// update replaces what a file holds, and the properties meta names.
func (p parentsDrive) update(ctx context.Context, id string, meta *drive.File, content []byte) error {
	return p.watch.call(ctx, p.account, opUpdate, func(ctx context.Context) error {
		_, err := p.files.Update(ctx, p.account.Token(), id, meta, content)
		return err
	})
}

// watch records every call to Drive: a span inside the span of whatever asked
// for it, and a line.
type watch struct {
	tracer    trace.Tracer
	logger    *zap.Logger
	projectID string
}

// call makes one call to Drive on behalf of the account and records how it
// ended. Neither the span nor the line names the file or carries anything it
// holds: the outcome is a word of a closed list, and a failure is marked with
// a sentence of ours rather than the error's own.
func (w *watch) call(ctx context.Context, account store.Account, op string, do func(context.Context) error) error {
	started := time.Now()
	ctx, span := w.tracer.Start(ctx, "drive "+op, trace.WithSpanKind(trace.SpanKindClient))
	err := do(ctx)
	outcome := outcomeOf(err)
	span.SetAttributes(attribute.String("mathtrail.drive.outcome", outcome))
	failed := isFailure(outcome)
	if failed {
		span.SetStatus(codes.Error, "drive did not do what was asked")
	}
	span.End()

	fields := append([]zap.Field{
		zap.String("op", op),
		zap.Int64("duration_ms", time.Since(started).Milliseconds()),
		// No call is made twice yet: every one of them is the only attempt.
		zap.Int("retries", 0),
		zap.String("outcome", outcome),
		zap.String("request_id", logger.RequestID(ctx)),
		zap.String("user", account.ID),
	}, telemetry.LogFields(ctx, w.projectID)...)
	if failed {
		w.logger.Warn(eventDriveCall, fields...)
	} else {
		w.logger.Info(eventDriveCall, fields...)
	}
	return err
}

// outcomeOf is how a call ended, in the words a span and a line use. A call
// that ran out of time, or that its caller gave up on, is told as that before
// anything it wraps.
func outcomeOf(err error) string {
	switch {
	case err == nil:
		return outcomeOK
	case errors.Is(err, context.DeadlineExceeded):
		return outcomeTimeout
	case errors.Is(err, context.Canceled):
		return outcomeCanceled
	case errors.Is(err, drive.ErrNotFound):
		return outcomeNotFound
	case errors.Is(err, drive.ErrUnauthorized):
		return outcomeRevoked
	case errors.Is(err, drive.ErrRateLimited):
		return outcomeRateLimited
	case errors.Is(err, drive.ErrStorageFull):
		return outcomeStorageFull
	case errors.Is(err, drive.ErrTooLarge):
		return outcomeTooLarge
	case errors.Is(err, drive.ErrUnavailable):
		return outcomeUnavailable
	}
	return outcomeFailed
}

// isFailure reports whether a call ended in a failure, of Drive's or of ours,
// rather than in an answer about the parent's Drive: a file that is not
// there, access taken back, a full Drive and a file too large are all things
// Drive says about the parent's side, and nothing went wrong in saying them.
// Nor did anything go wrong in a call its caller gave up on.
func isFailure(outcome string) bool {
	switch outcome {
	case outcomeRateLimited, outcomeUnavailable, outcomeTimeout, outcomeFailed:
		return true
	}
	return false
}

package mcpserver

import (
	"context"
	"slices"
	"strings"

	"go.uber.org/zap"

	"github.com/MathTrail/mathtrail-standalone/internal/store"
)

// The events of a lesson, as the lines about them are named.
const (
	eventTaskRequested  = "task_requested"
	eventTaskSkipped    = "task_skipped"
	eventTaskSubmitted  = "task_submitted"
	eventTaskAccepted   = "task_accepted"
	eventAnswerRecorded = "answer_recorded"
	eventSolverRun      = "solver_run"
)

// lessonLog writes a line for each thing that happens in a lesson — a task
// asked for, skipped, handed in, accepted, answered, a run of its solver —
// which is what every count of how the service is used is taken from.
//
// A line is built from fields the caller names one by one, and never from a
// profile, a task or anything a person or a model wrote: those hold what a
// child could be recognised by, and the answer.
type lessonLog struct {
	logger              *zap.Logger
	projectID           string
	instructionsVersion string
}

// write leaves the line of one event: its own fields, the version of the
// instructions the model has now, and what every line of a call carries.
func (l *lessonLog) write(ctx context.Context, account store.Account, event string, fields ...zap.Field) {
	l.writeFor(ctx, account, l.instructionsVersion, event, fields...)
}

// writeFor leaves the line of an event about a task written to a version of
// the instructions of its own — an answer to a task handed out before the
// service took newer ones — so that its result is counted against the
// instructions that produced it. That version is read from the profile, which
// a person can edit, so a line keeps it only while it has the shape of a
// version the service writes, and names it other when it does not.
func (l *lessonLog) writeFor(ctx context.Context, account store.Account, instructionsVersion, event string, fields ...zap.Field) {
	if !versionShaped(instructionsVersion) {
		instructionsVersion = other
	}
	l.logger.Info(event, slices.Concat(
		fields,
		[]zap.Field{zap.String("instructions_version", instructionsVersion)},
		callerFields(ctx, account.ID, l.projectID),
	)...)
}

// versionShaped reports whether a text has the shape of a version of the
// instructions: lowercase hexadecimal digits alone, as a digest writes them,
// and no longer than a whole one. The length of the versions a build writes is
// not held to, so that a task written before a build that changed it is still
// counted against its own.
func versionShaped(version string) bool {
	return version != "" && len(version) <= maxVersion && strings.Trim(version, "0123456789abcdef") == ""
}

// maxVersion is the length of a whole digest written in hexadecimal, the
// longest a version of the instructions can be.
const maxVersion = 64

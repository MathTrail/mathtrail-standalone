package mcpserver

import (
	"context"
	"slices"

	"go.uber.org/zap"

	"github.com/MathTrail/mathtrail-standalone/internal/store"
)

// The events of a lesson, as the lines about them are named.
const (
	eventTaskRequested = "task_requested"
	eventTaskSkipped   = "task_skipped"
	eventTaskSubmitted = "task_submitted"
	eventTaskAccepted  = "task_accepted"
	eventSolverRun     = "solver_run"
)

// lessonLog writes a line for each thing that happens in a lesson — a task
// asked for, skipped, handed in, accepted, a run of its solver — which is what
// every count of how the service is used is taken from.
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
// instructions the model had, and what every line of a call carries.
func (l *lessonLog) write(ctx context.Context, account store.Account, event string, fields ...zap.Field) {
	l.logger.Info(event, slices.Concat(
		fields,
		[]zap.Field{zap.String("instructions_version", l.instructionsVersion)},
		callerFields(ctx, account.ID, l.projectID),
	)...)
}

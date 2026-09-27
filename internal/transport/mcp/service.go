package mcpserver

import (
	"fmt"
	"time"

	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"

	"github.com/MathTrail/mathtrail-standalone/content"
	"github.com/MathTrail/mathtrail-standalone/internal/domain/checks"
	"github.com/MathTrail/mathtrail-standalone/internal/domain/profile"
	"github.com/MathTrail/mathtrail-standalone/internal/store"
)

// Service is what the tools of the lesson work with: where the child's profile
// is kept, the content the service ships, the checks a written task passes and
// the seal its answer is kept under, how long a task being written is waited
// for, and the clock and the build that stamp every profile they write. Every
// tool reads the profile, computes and writes it back through these and
// nothing else.
type Service struct {
	store    store.Storage
	content  *content.Content
	reviewer checks.Reviewer
	sealer   profile.Sealer
	window   time.Duration
	now      func() time.Time
	version  string
	events   *lessonLog
	tracer   trace.Tracer
}

// Parts are what the tools of the lesson are built from.
type Parts struct {
	// Store keeps the child's profile.
	Store store.Storage
	// Content is what the service ships: the catalogs, the reference tasks and
	// what the model is told.
	Content *content.Content
	// Reviewer takes a task the model wrote through the checks.
	Reviewer checks.Reviewer
	// Sealer keeps what gives a task's answer away out of reach until the
	// child has answered.
	Sealer profile.Sealer
	// Window is how long a task being written is waited for before its request
	// counts as abandoned.
	Window time.Duration
	// Now is the clock every write is stamped with.
	Now func() time.Time
	// Version is the build that writes the profile.
	Version string
	// Logger writes the lines of what happens in a lesson.
	Logger *zap.Logger
	// Traces records the stages of a task's review inside the call's span.
	Traces trace.TracerProvider
	// ProjectID names the project a line's trace is filed under.
	ProjectID string
}

// NewService builds what the tools work with, and refuses to build it without
// any part of it: a tool called on a part that is missing would fail in the
// middle of a child's lesson, and the start of the process is the time to say
// so.
func NewService(parts *Parts) (*Service, error) {
	switch {
	case parts == nil:
		return nil, fmt.Errorf("%w: the tools need their parts", ErrSettings)
	case parts.Store == nil:
		return nil, fmt.Errorf("%w: the tools need a store", ErrSettings)
	case parts.Content == nil:
		return nil, fmt.Errorf("%w: the tools need the content", ErrSettings)
	case parts.Reviewer == nil:
		return nil, fmt.Errorf("%w: the tools need the checks a task passes", ErrSettings)
	case parts.Sealer == nil:
		return nil, fmt.Errorf("%w: the tools need a seal for the answers", ErrSettings)
	case parts.Window <= 0:
		return nil, fmt.Errorf("%w: the tools need a window to wait for a task in", ErrSettings)
	case parts.Now == nil:
		return nil, fmt.Errorf("%w: the tools need a clock", ErrSettings)
	case parts.Version == "":
		return nil, fmt.Errorf("%w: the tools need the version of the build", ErrSettings)
	case parts.Logger == nil:
		return nil, fmt.Errorf("%w: the tools need a logger", ErrSettings)
	case parts.Traces == nil:
		return nil, fmt.Errorf("%w: the tools need traces", ErrSettings)
	}
	return &Service{
		store:    parts.Store,
		content:  parts.Content,
		reviewer: parts.Reviewer,
		sealer:   parts.Sealer,
		window:   parts.Window,
		now:      parts.Now,
		version:  parts.Version,
		events: &lessonLog{
			logger:              parts.Logger,
			projectID:           parts.ProjectID,
			instructionsVersion: parts.Content.InstructionsVersion(),
		},
		tracer: parts.Traces.Tracer(tracerScope),
	}, nil
}

// ProfileTools are the tools that read and write the child's details and read
// the progress: the profile for the model and its card, and the progress
// twice — once for the model, with a card, and once for a card that opens the
// progress inside itself.
func (s *Service) ProfileTools() []Tool {
	return []Tool{
		s.getProfileTool(),
		s.saveProfileTool(),
		s.getProgressTool(),
		s.readProgressTool(),
	}
}

// TaskTools are the tools that take a task from the rule to the child and back:
// one asks for it and hands the model what to write it from, one takes what the
// model wrote through the checks and puts it on the child's card, and one
// records the child's answer and tells how it went.
func (s *Service) TaskTools() []Tool {
	return []Tool{
		s.nextTaskTool(),
		s.submitTaskTool(),
		s.submitAnswerTool(),
	}
}

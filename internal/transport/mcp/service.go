package mcpserver

import (
	"fmt"
	"time"

	"github.com/MathTrail/mathtrail-standalone/content"
	"github.com/MathTrail/mathtrail-standalone/internal/store"
)

// Service is what the tools of the lesson work with: where the child's profile
// is kept, the content the service ships, and the clock and the build that
// stamp every profile they write. Every tool reads the profile, computes and
// writes it back through these and nothing else.
type Service struct {
	store   store.Storage
	content *content.Content
	now     func() time.Time
	version string
}

// NewService builds what the tools work with, and refuses to build it without
// any part of it: a tool called on a part that is missing would fail in the
// middle of a child's lesson, and the start of the process is the time to say
// so.
func NewService(kept store.Storage, shipped *content.Content, now func() time.Time, version string) (*Service, error) {
	switch {
	case kept == nil:
		return nil, fmt.Errorf("%w: the tools need a store", ErrSettings)
	case shipped == nil:
		return nil, fmt.Errorf("%w: the tools need the content", ErrSettings)
	case now == nil:
		return nil, fmt.Errorf("%w: the tools need a clock", ErrSettings)
	case version == "":
		return nil, fmt.Errorf("%w: the tools need the version of the build", ErrSettings)
	}
	return &Service{store: kept, content: shipped, now: now, version: version}, nil
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

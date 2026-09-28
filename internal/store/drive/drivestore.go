// Package drivestore keeps the child's profile in the parent's own Google
// Drive: one JSON file, in a folder the parent can open, reached with the
// parent's own token and nothing of the service's.
//
// The file is found by a private property the service marks it with, and
// never by its name or its place: the parent may rename it, move it or tidy it
// away, and the next search still finds it. The folder is a courtesy to
// whoever opens Drive, marked the same way and made again when a new file
// needs one. Which file holds whose profile is remembered while the instance
// runs, for a few minutes at a time, so that a warm instance reads with one
// call; nothing durable is kept.
//
// Drive has no conditional write. So a save reads the file once more before it
// writes, and refuses when what it finds is not what the profile was computed
// from: a change made in between — another tab, another instance, the parent
// by hand — is never written over from here. The writes of one account go one
// at a time within an instance, so two tabs served by the same instance never
// meet in between; what is left is two instances writing within the moment of
// one upload, which nothing in Drive's API lets a caller close.
//
// A file that goes wrong is dealt with where it lies. A damaged one is put
// back to its last state that reads, from the history Drive keeps of it — a
// state of each day is kept forever for that — or, when the parent asks for a
// new start, set aside rather than deleted. One in the bin is left there for
// the parent to restore. A call Drive asks to pause is tried again, twice.
//
// Every call to Drive leaves a span and a line: what was asked and how it
// ended, and never which file, nor anything it holds.
package drivestore

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"

	"github.com/MathTrail/mathtrail-standalone/internal/infra/drive"
	"github.com/MathTrail/mathtrail-standalone/internal/store"
)

// The layout in Drive. The markers are what the folder and the file are
// found by, and they are written into parents' Drives for good: a marker
// changed here is a profile nobody finds again.
const (
	folderName    = "MathTrail"
	fileName      = "mathtrail-profile.json"
	fileType      = "application/json"
	markerKey     = "mathtrail"
	profileMarker = "profile"
	folderMarker  = "folder"
	// setAsideMarker marks a file a profile was started over from: no search
	// for the profile finds it again, and it is still there to be found.
	setAsideMarker = "set-aside"
	// schemaKey carries a copy of the version of the file's shape, so that a
	// later edition can tell what a file is before it downloads it. The
	// file's own is the one that counts, and every write keeps the two alike.
	schemaKey = "schema"
)

// setAsideName is what a file set aside is called: the profile's name and
// the day it was set aside, so that the parent can tell it from the profile.
func setAsideName(day time.Time) string {
	return "mathtrail-profile set aside " + day.UTC().Format(time.DateOnly) + ".json"
}

// maxFile is the most of a file that is read as a profile. A profile is kept
// under 256 KB; four times that leaves room for a file edited by hand, and
// keeps a file that is no profile at all out of memory.
const maxFile = 1 << 20

// Settings are what the store is built from.
type Settings struct {
	// Files are the calls to the parents' Drives.
	Files drive.Files
	// Revisions are the calls to the history Drive keeps of a file.
	Revisions drive.Revisions
	// Timeout is how long one call to Drive may take: the time the calls
	// above were built with, which the store cannot read from them. No call
	// is started that could outlive the token of the account it is made for.
	Timeout time.Duration
	// Logger writes a line for every call to Drive.
	Logger *zap.Logger
	// Traces records a span for every call to Drive, inside the span of
	// whatever asked for it.
	Traces trace.TracerProvider
	// ProjectID names the project a line's trace is filed under.
	ProjectID string
	// Now is the store's clock; the system's when it is nil.
	Now func() time.Time
	// Sleep waits for as long as given, or until ctx ends, whichever comes
	// first; a timer when it is nil.
	Sleep func(ctx context.Context, d time.Duration) error
}

// ErrSettings is returned when the store is given settings it could not work
// with; callers branch on it with errors.Is.
var ErrSettings = errors.New("drivestore: settings")

// driveStore keeps each account's profile in the Drive of the parent it
// belongs to.
type driveStore struct {
	files     drive.Files
	revisions drive.Revisions
	ids       *fileIDs
	turns     *turns
	watch     *watch
	now       func() time.Time
	sleep     func(ctx context.Context, d time.Duration) error
}

// New builds the store, and refuses to build it without any part of it: a
// store missing a part would fail in the middle of a child's lesson, and the
// start of the process is the time to say so.
func New(settings *Settings) (store.Storage, error) {
	switch {
	case settings == nil:
		return nil, fmt.Errorf("%w: the store needs its settings", ErrSettings)
	case settings.Files == nil:
		return nil, fmt.Errorf("%w: the store needs the calls to Drive", ErrSettings)
	case settings.Revisions == nil:
		return nil, fmt.Errorf("%w: the store needs the calls to the history of files", ErrSettings)
	case settings.Timeout <= 0:
		return nil, fmt.Errorf("%w: the store needs the time a call to Drive may take", ErrSettings)
	case settings.Logger == nil:
		return nil, fmt.Errorf("%w: the store needs a logger", ErrSettings)
	case settings.Traces == nil:
		return nil, fmt.Errorf("%w: the store needs traces", ErrSettings)
	}
	now, sleep := settings.Now, settings.Sleep
	if now == nil {
		now = time.Now
	}
	if sleep == nil {
		sleep = timer
	}
	return &driveStore{
		files:     settings.Files,
		revisions: settings.Revisions,
		ids:       newFileIDs(),
		turns:     newTurns(),
		watch: &watch{
			tracer:    settings.Traces.Tracer(scope),
			logger:    settings.Logger,
			projectID: settings.ProjectID,
			timeout:   settings.Timeout,
			now:       now,
			sleep:     sleep,
		},
		now:   now,
		sleep: sleep,
	}, nil
}

// timer waits for as long as given, or until ctx ends.
func timer(ctx context.Context, d time.Duration) error {
	waiting := time.NewTimer(d)
	defer waiting.Stop()
	select {
	case <-waiting.C:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// The refusals that come before anything is asked of Drive.
var (
	// errNoOne refuses an account that names no one: whatever its token would
	// open, it would open on behalf of nobody the service signed in.
	errNoOne = errors.New("an account that names no one")
	// errNoToken refuses an account that carries nothing to reach a Drive
	// with.
	errNoToken = errors.New("an account with no token to reach its Drive with")
)

// ready is what every operation needs before it asks Drive anything: a
// context that has not ended, and an account that names someone and carries
// the token of their Drive.
func ready(ctx context.Context, account store.Account) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	switch {
	case account.ID == "":
		return errNoOne
	case account.Token() == "":
		return errNoToken
	}
	return nil
}

// refusalOf is Drive's refusal in the store's words, where the store has words
// for it: a token Drive no longer takes is access taken back, a file too large
// for any profile is one nothing can read, a full Drive is no room left, and
// Drive asking for a pause, or failing, past the tries a call is given, is
// Drive out of reach. A call that ran out of its time, or that its caller
// gave up on, stays that. Anything else stays Drive's.
func refusalOf(err error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, context.DeadlineExceeded), errors.Is(err, context.Canceled):
		return err
	case errors.Is(err, drive.ErrUnauthorized):
		return fmt.Errorf("%w: %w", store.ErrAccessRevoked, err)
	case errors.Is(err, drive.ErrTooLarge):
		return fmt.Errorf("%w: %w", store.ErrCorrupted, err)
	case errors.Is(err, drive.ErrStorageFull):
		return fmt.Errorf("%w: %w", store.ErrStorageFull, err)
	case errors.Is(err, drive.ErrRateLimited), errors.Is(err, drive.ErrUnavailable):
		return fmt.Errorf("%w: %w", store.ErrUnavailable, err)
	}
	return err
}

// Package profile is the file that holds everything the service knows about
// one child, and the only thing it keeps at all.
//
// There is no database and no session: a JSON file in the parent's Drive is
// read at the start of a tool call, computed over, and written back at the
// end. So this is not a cache of some truth held elsewhere — it is the truth,
// and the shape below is what every other part of the service agrees on.
//
// Two properties run through the whole layout. It holds everything the rule
// needs to choose the next task, in a form that cannot grow without bound: a
// summary per topic that is never pruned, a short window of recent answers,
// and sketches of past tasks rather than their texts. And it gives nothing
// away about the task the child is working on: the answer, the explanations
// behind the wrong options, the solution and the solver live in one sealed
// string, so a parent who opens the file finds the question and not the answer.
//
// This package is the shape of the file and nothing more — reading it, writing
// it and refusing what does not belong in it. Moving a rating, applying an
// answer and unsealing a task are the operations on top of it.
package profile

import (
	"time"

	"github.com/google/uuid"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/rating"
)

// Version is the shape this package reads and writes. It is raised when a
// field is removed, renamed, or comes to mean something else; a new optional
// field is not a new version.
const Version = 1

// Profile is the whole file. Fields are declared in the order they are
// written, which is alphabetical by name: the file is read by a parent and
// diffed by Drive, and a stable order is worth more than a grouping that only
// the code would enjoy.
type Profile struct {
	// AppVersion is the build that last wrote the file, so that a file which
	// later turns out to be broken can be read against the code that made it.
	AppVersion string `json:"app_version"`
	// CreatedAt is when the profile was made.
	CreatedAt Time `json:"created_at"`
	// CurrentTask is the task the child is working on, or nil when there is
	// none.
	CurrentTask *CurrentTask `json:"current_task"`
	// Daily counts what has to be counted across every instance of the
	// service, which is why it is in the file rather than in memory.
	Daily Daily `json:"daily"`
	// OpenRequest is the task being written right now, or nil. It is what
	// makes "three attempts" survive a restart.
	OpenRequest *OpenRequest `json:"open_request"`
	// Ratings is the child's level across every topic.
	Ratings Ratings `json:"ratings"`
	// Recent is the history window, oldest first.
	Recent []Answer `json:"recent"`
	// Revision is raised on every write. Paired with the revision Drive keeps
	// of its own, it is how two tabs writing at once are noticed.
	Revision int `json:"revision"`
	// SchemaVersion is the shape of this file.
	SchemaVersion int `json:"schema_version"`
	// Student is everything the parent set up about the child.
	Student Student `json:"student"`
	// StudentID is a permanent random identifier. It means nothing outside
	// this file and names nobody; it exists so a profile can be carried
	// elsewhere.
	StudentID string `json:"student_id"`
	// TaskFingerprints are sketches of the tasks already given, oldest first
	// and never their texts: enough to notice a repeat, not enough to read
	// back what the child was asked.
	TaskFingerprints []string `json:"task_fingerprints"`
	// Topics is the summary the rule runs on, keyed by the catalog's topic id.
	// It is the part that is never pruned.
	Topics map[string]Topic `json:"topics"`
	// UpdatedAt is when the file was last written.
	UpdatedAt Time `json:"updated_at"`
}

// New starts a profile for a child. Everything that has to be decided once —
// the identifier, the day it was made and where on the ladder the child
// starts — is decided here, and everything else starts empty. The start is the
// one thing the grade decides: the child begins at the level their grade falls
// into, and from there only answers move them.
//
//nolint:gocritic // hugeParam: a student is taken by value on purpose, so that a profile cannot be handed a struct somebody else still holds a pointer to
func New(student Student, appVersion string, now time.Time) *Profile {
	return &Profile{
		AppVersion:       appVersion,
		Student:          student,
		StudentID:        uuid.NewString(),
		CreatedAt:        At(now),
		Daily:            Daily{Date: DateOf(now)},
		Ratings:          Ratings{Start: rating.Start(student.Grade), Theta: rating.Start(student.Grade)},
		Recent:           []Answer{},
		Revision:         1,
		SchemaVersion:    Version,
		TaskFingerprints: []string{},
		Topics:           map[string]Topic{},
		UpdatedAt:        At(now),
	}
}

// Touch records that the file is being written: the revision moves on and the
// build that wrote it is stamped. Every write goes through here, so a caller
// cannot forget the half of a write that nobody would notice missing.
func (p *Profile) Touch(appVersion string, now time.Time) {
	p.Revision++
	p.AppVersion = appVersion
	p.UpdatedAt = At(now)
	p.SchemaVersion = Version
}

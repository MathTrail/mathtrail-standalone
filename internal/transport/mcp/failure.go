package mcpserver

import (
	"context"
	"errors"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/profile"
	"github.com/MathTrail/mathtrail-standalone/internal/store"
)

// The kinds of failure a line can name. They are the only words a failure is
// ever logged or traced by: a closed list, never the text of an error.
const (
	kindInternal    = "internal"
	kindTimeout     = "timeout"
	kindNotSignedIn = "not_signed_in"
	kindPanic       = "panic"
	kindConflict    = "conflict"
	kindSealed      = "sealed"
)

// sentenceInternal is what the model is told when something of ours failed
// and nothing more specific can be said.
const sentenceInternal = "Something went wrong inside MathTrail. " +
	"Try the same request again in a moment; if it keeps failing, tell the adult that MathTrail is having trouble."

// errNotSignedIn is a call that reached a tool with nobody signed in. The
// sign-in in front of the endpoint makes it impossible; the tool frame checks
// anyway, because the alternative is a tool acting for nobody.
var errNotSignedIn = errors.New("mcp: no account signed in")

// errToldNoMore is an answer sent again for a task whose seal can no longer be
// opened. The answer was recorded before the seal was lost and counts; only
// telling it again is gone.
var errToldNoMore = errors.New("mcp: the answer is recorded and can no longer be told again")

// failures are the causes a model is told about in words of their own, looked
// for in this order. Every sentence says what to do next.
var failures = []struct {
	cause    error
	kind     string
	sentence string
}{
	{
		cause:    errNotSignedIn,
		kind:     kindNotSignedIn,
		sentence: "Nobody is signed in to MathTrail. Ask the adult to reconnect MathTrail, then try again.",
	},
	{
		cause:    context.DeadlineExceeded,
		kind:     kindTimeout,
		sentence: "MathTrail took too long to answer. Try the same request again in a moment.",
	},
	{
		// Somebody else wrote the profile between this call's read and its
		// write — another tab, another device. Nothing was written, and the
		// same call made again starts from what they wrote.
		cause:    store.ErrConflict,
		kind:     kindConflict,
		sentence: "The child's profile was changed somewhere else at the same moment, so nothing was saved. Make the same call again.",
	},
	{
		// The seal was lost after the answer was recorded: the answer counts,
		// and the model must not tell the child otherwise. Looked for before
		// the lost seal it wraps.
		cause:    errToldNoMore,
		kind:     kindSealed,
		sentence: "This task was answered already and the answer is recorded, but it can no longer be told again, so the task was taken off the card. Read the last recorded answer in the result of the next tool you call, and ask for a new task with next_task when the child wants another.",
	},
	{
		// The task's answer is sealed under a key that has since been
		// retired: it cannot be checked, so it was taken off the card, and
		// the child is owed another.
		cause:    profile.ErrSealed,
		kind:     kindSealed,
		sentence: "This task can no longer be checked, so no answer was recorded and it was taken off the card. Tell the child, and ask for a new task with next_task.",
	},
}

// failure is a tool call that could not do its work, as it is told to the
// model: one sentence of ours, and never the words of what went wrong.
type failure struct {
	kind     string
	sentence string
	cause    error
}

// Error is the sentence the model reads. The protocol puts an error's text in
// front of the model, so this one is written for the model rather than for a
// log, and it is the only text of a failure anybody ever sees.
func (f *failure) Error() string { return f.sentence }

// Unwrap keeps the cause reachable for errors.Is, and for nothing that prints.
func (f *failure) Unwrap() error { return f.cause }

// explain is the one way an error becomes something to tell the model: the
// first cause of the table it wraps, or the internal failure when it wraps
// none of them.
func explain(err error) *failure {
	for _, known := range failures {
		if errors.Is(err, known.cause) {
			return &failure{kind: known.kind, sentence: known.sentence, cause: err}
		}
	}
	return &failure{kind: kindInternal, sentence: sentenceInternal, cause: err}
}

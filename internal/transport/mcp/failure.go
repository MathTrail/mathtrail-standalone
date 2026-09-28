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
	// The kinds of what happened to the parent's Drive, or to the profile's
	// file in it.
	kindRevoked     = "revoked"
	kindExpired     = "expired"
	kindStorageFull = "storage_full"
	kindUnavailable = "drive_unavailable"
	kindCorrupted   = "corrupted"
	kindInBin       = "in_bin"
	kindBehind      = "behind"
	kindRestored    = "restored"
	kindNewer       = "newer"
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
	{
		// The parent took the service's access to their Drive back. Only a
		// new sign-in mends it; the result also carries the challenge a host
		// that reads one signs in again with.
		cause:    store.ErrAccessRevoked,
		kind:     kindRevoked,
		sentence: "MathTrail can no longer reach the child's profile: the access to the adult's Google Drive was taken back, so nothing was read or saved. Ask the adult to connect MathTrail again and to let it use Google Drive, then make the same call again.",
	},
	{
		// The Google access the call carried ends before a call to Drive
		// could finish. Renewed access mends it, which the host asks for.
		cause:    store.ErrAccessExpired,
		kind:     kindExpired,
		sentence: "MathTrail's access to the adult's Google Drive has to be renewed, so nothing was read or saved. Make the same call again; if it keeps failing, ask the adult to connect MathTrail again.",
	},
	{
		// The one failure where the child did something and it did not stick.
		cause:    store.ErrStorageFull,
		kind:     kindStorageFull,
		sentence: "The adult's Google Drive is full, so nothing was saved: if the child has just answered, that answer was not recorded. Ask the adult to free up space in Google Drive, then make the same call again.",
	},
	{
		cause:    store.ErrUnavailable,
		kind:     kindUnavailable,
		sentence: "Google Drive is not answering at the moment, so nothing was saved. Make the same call again in a minute; if it keeps failing, tell the adult that Google Drive is having trouble.",
	},
	{
		// The file cannot be read, and was not put back: no earlier state of
		// it reads, or it is a profile that breaks a rule, which is never
		// rolled back on the service's say. The two ways out are the parent's.
		cause:    store.ErrCorrupted,
		kind:     kindCorrupted,
		sentence: "The child's profile file in the adult's Google Drive cannot be read, so nothing was done. Tell the adult. They can restore an earlier version of the file from its version history in Google Drive and then make the same call again, or start a new profile: call save_profile with start_over set to true, a pseudonym and the grade. The file that cannot be read is kept in Drive, not deleted.",
	},
	{
		cause:    store.ErrInBin,
		kind:     kindInBin,
		sentence: "The child's profile file is in the adult's Google Drive bin, so nothing was done. Ask the adult to restore it from the bin in Google Drive, then make the same call again. If they would rather start a new profile, call save_profile with start_over set to true, a pseudonym and the grade: the file in the bin is kept, not deleted.",
	},
	{
		cause:    store.ErrBehind,
		kind:     kindBehind,
		sentence: "The child's profile is still being saved in Google Drive, so nothing was done. Make the same call again in a moment.",
	},
	{
		// The call found the file damaged and put it back to a state that
		// reads: that is all it did, and the parent should hear of it.
		cause:    store.ErrRestored,
		kind:     kindRestored,
		sentence: "The child's profile file in the adult's Google Drive was damaged, so MathTrail put it back to its latest earlier version that can be read; anything saved after that version is lost. Tell the adult what happened, then make the same call again.",
	},
	{
		// A rollout under way: the newer build writes a shape this one does
		// not read, and waiting is what helps.
		cause:    profile.ErrNewer,
		kind:     kindNewer,
		sentence: "The child's profile was last saved by a newer version of MathTrail than the one answering now, so nothing was done. Make the same call again in a few minutes.",
	},
}

// signInAgain is whether a failure is mended by the parent signing in again,
// or by the host renewing its tokens: a host that reads a challenge in the
// result is handed one.
func signInAgain(kind string) bool {
	return kind == kindRevoked || kind == kindExpired
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

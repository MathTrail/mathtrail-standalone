package mcpserver

import (
	"context"
	"errors"
)

// The kinds of failure a line can name. They are the only words a failure is
// ever logged or traced by: a closed list, never the text of an error.
const (
	kindInternal    = "internal"
	kindTimeout     = "timeout"
	kindNotSignedIn = "not_signed_in"
	kindPanic       = "panic"
)

// sentenceInternal is what the model is told when something of ours failed
// and nothing more specific can be said.
const sentenceInternal = "Something went wrong inside MathTrail. " +
	"Try the same request again in a moment; if it keeps failing, tell the adult that MathTrail is having trouble."

// errNotSignedIn is a call that reached a tool with nobody signed in. The
// sign-in in front of the endpoint makes it impossible; the tool frame checks
// anyway, because the alternative is a tool acting for nobody.
var errNotSignedIn = errors.New("mcp: no account signed in")

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

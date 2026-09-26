package mcpserver

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

// told is every sentence a model can be told about a failure, and the kind that
// goes with each.
func told() map[string]string {
	sentences := map[string]string{sentenceInternal: kindInternal}
	for _, known := range failures {
		sentences[known.sentence] = known.kind
	}
	return sentences
}

// The property a failure exists for: whatever an error said, the model is told
// one of the sentences written for it, and the kind the line names is the one
// that goes with that sentence. An error's own words never reach the model,
// however deeply they are wrapped and whatever they are joined with.
func TestAFailureIsAlwaysOneOfOurSentences(t *testing.T) {
	t.Parallel()

	sentences := told()
	properties := gopter.NewProperties(nil)

	properties.Property("the sentence is one of ours and the kind is its own", prop.ForAll(
		func(said string, depth int, known int) bool {
			err := errors.New(said)
			if known >= 0 {
				err = errors.Join(err, failures[known].cause)
			}
			for range depth {
				err = fmt.Errorf("layer: %w", err)
			}
			explained := explain(err)
			kind, ours := sentences[explained.Error()]
			return ours && kind == explained.kind
		},
		gen.AnyString(),
		gen.IntRange(0, 4),
		gen.IntRange(-1, len(failures)-1),
	))

	properties.Property("a known cause is named whatever wraps it", prop.ForAll(
		func(said string, depth int, known int) bool {
			err := errors.Join(errors.New(said), failures[known].cause)
			for range depth {
				err = fmt.Errorf("layer: %w", err)
			}
			return explain(err).kind == failures[known].kind
		},
		gen.AnyString(),
		gen.IntRange(0, 4),
		gen.IntRange(0, len(failures)-1),
	))

	properties.TestingRun(t)
}

// Where an error wraps two known causes, the one earlier in the table is the
// one the model is told about: a call that reached a tool with nobody signed in
// is about the sign-in, however long it then took.
func TestTheEarlierCauseWins(t *testing.T) {
	t.Parallel()

	err := errors.Join(context.DeadlineExceeded, errNotSignedIn)
	if got := explain(err).kind; got != kindNotSignedIn {
		t.Errorf("kind = %q, want %q", got, kindNotSignedIn)
	}
}

// An error nobody named is a failure of ours, told in the general sentence,
// and the error stays reachable for errors.Is.
func TestAnUnknownErrorIsInternal(t *testing.T) {
	t.Parallel()

	cause := errors.New("drive: file 1a2b3c: permission denied for masha")
	explained := explain(fmt.Errorf("profile: %w", cause))

	if explained.kind != kindInternal {
		t.Errorf("kind = %q, want %q", explained.kind, kindInternal)
	}
	if explained.Error() != sentenceInternal {
		t.Errorf("sentence = %q, want the internal one", explained.Error())
	}
	if !errors.Is(explained, cause) {
		t.Error("the cause is not reachable with errors.Is, want it kept")
	}
}

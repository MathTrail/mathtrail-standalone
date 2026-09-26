package middleware

import (
	"errors"
	"testing"
	"testing/iotest"

	"github.com/google/uuid"
)

// When the source of randomness stops answering there is still a request to
// correlate. The identifier that follows has to pass the same rule as any
// other, and two of them have to be different.
func TestTheFallbackIdentifierIsUsableAndUnique(t *testing.T) {
	t.Parallel()

	first := fallbackRequestID()
	second := fallbackRequestID()

	for _, id := range []string{first, second} {
		if !usableRequestID(id) {
			t.Errorf("fallbackRequestID() = %q, want an id the middleware would keep", id)
		}
	}
	if first == second {
		t.Errorf("two identifiers came back the same: %q", first)
	}
}

// A request that arrives once the source of randomness has stopped answering
// is still given an identifier: the fallback, under the same rule and just as
// different from the one before it.
func TestAnIdentifierIsMadeWithoutRandomness(t *testing.T) {
	// Not parallel: the source of randomness belongs to the whole process, and
	// it is put back before any test that runs in parallel starts.
	uuid.SetRand(iotest.ErrReader(errors.New("no randomness left")))
	t.Cleanup(func() { uuid.SetRand(nil) })

	first, second := newRequestID(), newRequestID()
	for _, id := range []string{first, second} {
		if !usableRequestID(id) {
			t.Errorf("newRequestID() = %q, want an id the middleware would keep", id)
		}
	}
	if first == second {
		t.Errorf("two requests were given the same id %q, want them told apart", first)
	}
}

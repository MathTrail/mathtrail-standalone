package middleware

import "testing"

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

package ratelimit

import "testing"

// Kept is how many keys a limiter this package built keeps a count for, and it
// fails the test when the order the keys were used in and the counts kept for
// them have come apart.
func Kept(t *testing.T, l Limiter) int {
	t.Helper()

	k, isKeyed := l.(*keyed)
	if !isKeyed {
		t.Fatalf("Kept() was handed a %T, want a limiter New built", l)
	}
	k.mu.Lock()
	defer k.mu.Unlock()
	if k.recent.Len() != len(k.allowances) {
		t.Fatalf("the limiter orders %d keys and counts %d, want the same keys in both", k.recent.Len(), len(k.allowances))
	}
	return len(k.allowances)
}

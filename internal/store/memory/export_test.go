package memory

import (
	"bytes"
	"testing"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/profile"
	"github.com/MathTrail/mathtrail-standalone/internal/store"
)

// Plant writes bytes where an account's profile is kept, the way something
// outside the service would, and hands that to the tests outside the package
// so that they can hold the store to what it does with them. It is a write
// like any other, so what it plants has a revision of its own; the profile's
// own revision is whatever the bytes say, and nothing when they say nothing.
func Plant(t *testing.T, s store.Storage, account store.Account, raw []byte) {
	t.Helper()

	m, ok := s.(*memoryStore)
	if !ok {
		t.Fatalf("Plant() was handed a %T, want a store this package built", s)
	}
	counter := 0
	if p, err := profile.Parse(raw); err == nil {
		counter = p.Revision
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.put(account, bytes.Clone(raw), counter)
}

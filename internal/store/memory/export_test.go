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

// SetAside is every file the account started over from, the earliest first,
// as the store keeps them.
func SetAside(t *testing.T, s store.Storage, account store.Account) [][]byte {
	t.Helper()

	m, ok := s.(*memoryStore)
	if !ok {
		t.Fatalf("SetAside() was handed a %T, want a store this package built", s)
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	kept := make([][]byte, 0, len(m.setAside[account.ID]))
	for _, raw := range m.setAside[account.ID] {
		kept = append(kept, bytes.Clone(raw))
	}
	return kept
}

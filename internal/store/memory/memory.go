// Package memory keeps profiles in the memory of the process.
//
// What it keeps lasts as long as the process does and is seen by no other
// instance of the service. It keeps the bytes a profile is written as rather
// than the profile itself, and reads them back the way every store does: so a
// profile handed to it or taken from it shares nothing with what it holds, and
// a profile no other store would write is not written here either.
package memory

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"sync"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/profile"
	"github.com/MathTrail/mathtrail-standalone/internal/store"
)

// memoryStore holds one file per account, keyed by the account's identifier.
type memoryStore struct {
	// mu makes the check of a revision and the write that follows it one
	// step, so that of two saves made from the same revision exactly one
	// lands.
	mu    sync.Mutex
	files map[string]file
	// writes counts every write the store has taken and numbers the revision
	// of each. A number is never handed out twice, so a revision that was
	// current once cannot come to match a later state.
	writes int
}

// file is what an account's profile is kept as: the bytes, the write they
// came from, and the profile's own revision, which the next write has to move
// on by exactly one.
type file struct {
	raw      []byte
	revision store.Revision
	counter  int
}

// errNoOne refuses an account that names no one: whatever it would reach, it
// would reach on behalf of nobody the service signed in.
var errNoOne = errors.New("an account that names no one")

// New builds a store that holds nothing yet.
func New() store.Storage {
	return &memoryStore{files: map[string]file{}}
}

// ready is what every operation needs before it looks at anything: a context
// that has not ended, and an account that names someone.
func ready(ctx context.Context, account store.Account) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if account.ID == "" {
		return errNoOne
	}
	return nil
}

func (m *memoryStore) Load(ctx context.Context, account store.Account) (*profile.Profile, store.Revision, error) {
	if err := ready(ctx, account); err != nil {
		return nil, "", fmt.Errorf("memory: load: %w", err)
	}
	kept, found := m.find(account)
	if !found {
		return nil, "", fmt.Errorf("memory: load: %w", store.ErrNotFound)
	}
	// Read outside the lock: the bytes of a file are never changed once kept,
	// only replaced.
	p, err := store.Parse(kept.raw)
	if err != nil {
		return nil, "", fmt.Errorf("memory: load: %w", err)
	}
	return p, kept.revision, nil
}

func (m *memoryStore) Create(ctx context.Context, account store.Account, p *profile.Profile) (store.Revision, error) {
	if err := ready(ctx, account); err != nil {
		return "", fmt.Errorf("memory: create: %w", err)
	}
	raw, err := profile.Marshal(p)
	if err != nil {
		return "", fmt.Errorf("memory: create: %w", err)
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	if _, found := m.files[account.ID]; found {
		return "", fmt.Errorf("memory: create: %w: the account has a profile already", store.ErrConflict)
	}
	return m.put(account, raw, p.Revision), nil
}

func (m *memoryStore) Save(ctx context.Context, account store.Account, p *profile.Profile, expected store.Revision) (store.Revision, error) {
	if err := ready(ctx, account); err != nil {
		return "", fmt.Errorf("memory: save: %w", err)
	}
	raw, err := profile.Marshal(p)
	if err != nil {
		return "", fmt.Errorf("memory: save: %w", err)
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	kept, found := m.files[account.ID]
	switch {
	case !found:
		return "", fmt.Errorf("memory: save: %w", store.ErrNotFound)
	case kept.revision != expected:
		return "", fmt.Errorf("memory: save: %w", store.ErrConflict)
	case p.Revision != kept.counter+1:
		return "", fmt.Errorf("memory: save: the profile's revision is %d, want %d: a write moves it on by exactly one",
			p.Revision, kept.counter+1)
	}
	return m.put(account, raw, p.Revision), nil
}

func (m *memoryStore) Export(ctx context.Context, account store.Account) (store.Location, error) {
	if err := ready(ctx, account); err != nil {
		return store.Location{}, fmt.Errorf("memory: export: %w", err)
	}
	if _, found := m.find(account); !found {
		return store.Location{}, fmt.Errorf("memory: export: %w", store.ErrNotFound)
	}
	// Nothing outside the process can open what it keeps.
	return store.Location{}, nil
}

// find is the account's file, when it has one.
func (m *memoryStore) find(account store.Account) (file, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	kept, found := m.files[account.ID]
	return kept, found
}

// put keeps raw as the account's file, under a revision no file has had
// before. The caller holds the lock.
func (m *memoryStore) put(account store.Account, raw []byte, counter int) store.Revision {
	m.writes++
	revision := store.Revision(strconv.Itoa(m.writes))
	m.files[account.ID] = file{raw: raw, revision: revision, counter: counter}
	return revision
}

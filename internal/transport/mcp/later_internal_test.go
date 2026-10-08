package mcpserver

import (
	"context"
	"errors"
	"slices"
	"sync"
	"testing"
	"time"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/profile"
	"github.com/MathTrail/mathtrail-standalone/internal/store"
)

// A call waits for the writes of its account begun before it, and for none of
// another account's; a call whose caller gives up stops waiting.
func TestACallWaitsForTheWritesOfItsAccountBegunBeforeIt(t *testing.T) {
	t.Parallel()

	writes := newPending()
	_, end := writes.begin("masha")
	_, otherEnd := writes.begin("petya")
	defer otherEnd()

	waited := make(chan error, 1)
	go func() { waited <- writes.wait(t.Context(), "masha") }()
	select {
	case err := <-waited:
		t.Fatalf("wait() = %v while the write begun before it was under way, want it to wait", err)
	case <-time.After(50 * time.Millisecond):
	}
	end()
	select {
	case err := <-waited:
		if err != nil {
			t.Errorf("wait() = %v once the write before it ended, want nil: another account's write holds nothing", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("wait() still waits once the write before it ended, want it done")
	}

	_, stillGoing := writes.begin("masha")
	defer stillGoing()
	gaveUp, cancel := context.WithCancel(t.Context())
	cancel()
	if err := writes.wait(gaveUp, "masha"); !errors.Is(err, context.Canceled) {
		t.Errorf("wait() for a caller that gave up = %v, want context.Canceled", err)
	}
}

// A request is held open for its writes no longer than its bound, however
// long a write takes.
func TestARequestIsHeldNoLongerThanItsBound(t *testing.T) {
	t.Parallel()

	held := &heldWrites{}
	held.add(make(chan struct{}))
	const bound = 50 * time.Millisecond
	began := time.Now()
	held.wait(bound)
	if took := time.Since(began); took < bound || took > 5*time.Second {
		t.Errorf("wait() took %v for a write that never ends, want the bound, %v", took, bound)
	}
}

// Two writes of one account under way at once: ending the first leaves the
// second waited for, so a call that comes between them finds what the second
// writes rather than the file before it.
func TestEndingOneWriteOfAnAccountLeavesItsOtherWaitedFor(t *testing.T) {
	t.Parallel()

	writes := newPending()
	_, endFirst := writes.begin("masha")
	_, endSecond := writes.begin("masha")
	endFirst()

	waited := make(chan error, 1)
	go func() { waited <- writes.wait(t.Context(), "masha") }()
	select {
	case err := <-waited:
		t.Fatalf("wait() = %v while the second write was under way, want it to wait", err)
	case <-time.After(50 * time.Millisecond):
	}
	endSecond()
	select {
	case err := <-waited:
		if err != nil {
			t.Errorf("wait() = %v once the second write ended, want nil", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("wait() still waits once both writes ended, want it done")
	}
}

// untouched is a store a case holds to never being reached: it keeps the name
// of every operation called on it.
type untouched struct {
	mu      sync.Mutex
	reached []string
}

func (u *untouched) reach(operation string) {
	u.mu.Lock()
	defer u.mu.Unlock()
	u.reached = append(u.reached, operation)
}

// operations are the operations that reached the store, in the order they
// came.
func (u *untouched) operations() []string {
	u.mu.Lock()
	defer u.mu.Unlock()
	return slices.Clone(u.reached)
}

func (u *untouched) Load(context.Context, store.Account) (*profile.Profile, store.Revision, error) {
	u.reach("Load")
	return nil, "", nil
}

func (u *untouched) Create(context.Context, store.Account, *profile.Profile) (store.Revision, error) {
	u.reach("Create")
	return "", nil
}

func (u *untouched) Save(context.Context, store.Account, *profile.Profile, store.Revision) (store.Revision, error) {
	u.reach("Save")
	return "", nil
}

func (u *untouched) Export(context.Context, store.Account) (store.Location, error) {
	u.reach("Export")
	return store.Location{}, nil
}

func (u *untouched) Restore(context.Context, store.Account) (*profile.Profile, store.Revision, error) {
	u.reach("Restore")
	return nil, "", nil
}

func (u *untouched) StartOver(context.Context, store.Account, *profile.Profile) (store.Revision, error) {
	u.reach("StartOver")
	return "", nil
}

// Every operation on an account waits first for the writes after the answer
// begun for it. One whose caller gives up while it waits stops there, saying
// why, and never reaches the store: it would act on the file as it was before
// the write it was waiting for.
func TestNoOperationReachesTheStoreOnceItsCallerGaveUpWaiting(t *testing.T) {
	t.Parallel()

	account := store.NewAccount("masha", "", time.Time{})
	for _, tc := range []struct {
		name string
		call func(context.Context, settled) error
	}{
		{"Load", func(ctx context.Context, s settled) error { _, _, err := s.Load(ctx, account); return err }},
		{"Create", func(ctx context.Context, s settled) error { _, err := s.Create(ctx, account, nil); return err }},
		{"Save", func(ctx context.Context, s settled) error { _, err := s.Save(ctx, account, nil, "r1"); return err }},
		{"Export", func(ctx context.Context, s settled) error { _, err := s.Export(ctx, account); return err }},
		{"Restore", func(ctx context.Context, s settled) error { _, _, err := s.Restore(ctx, account); return err }},
		{"StartOver", func(ctx context.Context, s settled) error { _, err := s.StartOver(ctx, account, nil); return err }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			kept, writes := &untouched{}, newPending()
			_, end := writes.begin(account.ID)
			defer end()
			ctx, giveUp := context.WithCancel(t.Context())
			defer giveUp()
			returned := make(chan error, 1)
			go func() { returned <- tc.call(ctx, settled{Storage: kept, pending: writes}) }()

			// The call waits while the write is under way, and stops as soon
			// as its caller gives up, in the middle of the wait.
			select {
			case err := <-returned:
				t.Fatalf("%s() = %v while the write was under way, want it waiting for the write", tc.name, err)
			case <-time.After(50 * time.Millisecond):
			}
			giveUp()
			select {
			case err := <-returned:
				if !errors.Is(err, context.Canceled) {
					t.Errorf("%s() error = %v, want context.Canceled", tc.name, err)
				}
			case <-time.After(10 * time.Second):
				t.Fatalf("%s() still waits once its caller gave up, want it stopped", tc.name)
			}
			if reached := kept.operations(); len(reached) != 0 {
				t.Errorf("%s() reached the store with %v, want never", tc.name, reached)
			}
		})
	}
}

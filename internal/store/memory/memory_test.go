package memory_test

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/profile"
	"github.com/MathTrail/mathtrail-standalone/internal/store"
	"github.com/MathTrail/mathtrail-standalone/internal/store/memory"
	"github.com/MathTrail/mathtrail-standalone/internal/store/storetest"
)

func TestTheStoreInMemoryKeepsTheContract(t *testing.T) {
	t.Parallel()

	storetest.Run(t, func(*testing.T) storetest.Harness {
		s := memory.New()
		return storetest.Harness{
			Storage: s,
			Plant: func(t *testing.T, account store.Account, raw []byte) {
				t.Helper()
				memory.Plant(t, s, account, raw)
			},
		}
	})
}

// The contract lets two saves from one revision both land, because a store
// that compares revisions only before it writes cannot promise otherwise. This
// one can: the check and the write are one step, so exactly one lands.
func TestOfSavesRacingFromOneRevisionExactlyOneLands(t *testing.T) {
	t.Parallel()

	s := memory.New()
	mia := store.Account{ID: "mia"}
	p := profile.New(profile.Student{Pseudonym: "Mia", Grade: 3}, "memory_test", time.Now())
	read, err := s.Create(t.Context(), mia, p)
	if err != nil {
		t.Fatalf("Create() error = %v, want nil", err)
	}
	p.Touch("memory_test", time.Now())

	const racers = 16
	start := make(chan struct{})
	var landed atomic.Int32
	var racing sync.WaitGroup
	for range racers {
		racing.Go(func() {
			<-start
			if _, refused := s.Save(t.Context(), mia, p, read); refused == nil {
				landed.Add(1)
			}
		})
	}
	close(start)
	racing.Wait()

	if got := landed.Load(); got != 1 {
		t.Errorf("%d of %d racing saves landed, want exactly one", got, racers)
	}
}

// Nothing outside the process can open a profile kept in its memory, and the
// export says so with the zero location.
func TestAProfileInMemoryIsNowhereAPersonCouldOpen(t *testing.T) {
	t.Parallel()

	s := memory.New()
	mia := store.Account{ID: "mia"}
	p := profile.New(profile.Student{Pseudonym: "Mia", Grade: 3}, "memory_test", time.Now())
	if _, err := s.Create(t.Context(), mia, p); err != nil {
		t.Fatalf("Create() error = %v, want nil", err)
	}

	location, err := s.Export(t.Context(), mia)
	if err != nil {
		t.Fatalf("Export() error = %v, want nil", err)
	}
	if location != (store.Location{}) {
		t.Errorf("Export() = %+v, want the zero location", location)
	}
}

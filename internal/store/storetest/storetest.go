// Package storetest holds a store to the contract every store keeps, so that
// whatever reads and writes the profile can count on the same behaviour
// wherever the profile is kept.
//
// Run builds a fresh store for every case, and the cases run beside each
// other, sharing nothing. What a case cannot do through the store — put into
// it what something outside the service would have written — the harness does
// for it.
package storetest

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"slices"
	"sync"
	"testing"
	"time"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/profile"
	"github.com/MathTrail/mathtrail-standalone/internal/store"
)

// Harness is a store under test, and the one thing a case needs to reach it
// from outside the service.
type Harness struct {
	// Storage is the store under test. It holds nothing yet.
	Storage store.Storage
	// Plant writes raw bytes where the account's profile is kept, the way
	// something other than the service would: a parent editing the file by
	// hand, a newer build, a write cut short. The account need not have a
	// profile yet.
	Plant func(t *testing.T, account store.Account, raw []byte)
}

// Run holds a store to the contract. newHarness is called once for every case,
// with that case's test, so that no case sees what another one left behind.
func Run(t *testing.T, newHarness func(t *testing.T) Harness) {
	t.Helper()

	for _, c := range []struct {
		name string
		run  func(t *testing.T, h Harness)
	}{
		{"an account with no profile has none, and a save does not make one", noProfileIsMadeBySaving},
		{"a created profile comes back as it went in", aCreatedProfileComesBackAsItWentIn},
		{"a profile is created once", aProfileIsCreatedOnce},
		{"a save moves the profile on", aSaveMovesTheProfileOn},
		{"a save from a revision that moved on is refused", aSaveFromARevisionThatMovedOnIsRefused},
		{"a change made outside the service is not written over", aChangeMadeOutsideIsNotWrittenOver},
		{"a profile not moved on by exactly one write is refused", aProfileNotMovedOnByOneWriteIsRefused},
		{"a profile that breaks its own rules is not written", aProfileThatBreaksItsRulesIsNotWritten},
		{"an account that names no one reaches no profile", anAccountThatNamesNoOneReachesNoProfile},
		{"what a caller holds is not what the store keeps", whatACallerHoldsIsNotWhatTheStoreKeeps},
		{"accounts do not see each other's profiles", accountsDoNotSeeEachOther},
		{"a damaged file is reported as damage", aDamagedFileIsReportedAsDamage},
		{"a file from a newer build is not damage", aFileFromANewerBuildIsNotDamage},
		{"an account with a profile can be told where it is", anAccountWithAProfileCanBeToldWhereItIs},
		{"a finished context stops every operation", aFinishedContextStopsEveryOperation},
		{"saves racing from one revision leave one whole profile", savesRacingFromOneRevisionLeaveOneWholeProfile},
	} {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			c.run(t, newHarness(t))
		})
	}
}

// moment is when every profile here is made and moved on: a fixed one, so that
// what a case expects does not depend on the clock.
var moment = time.Date(2026, time.September, 26, 9, 30, 0, 0, time.UTC)

// someone is an account named for a case. The name goes into both the
// identifier and the token, so two accounts never share either.
func someone(name string) store.Account {
	return store.NewAccount("account-"+name, "token-"+name)
}

// child is a profile as the first sign-in makes it. Every field a parent types
// into carries more than one script, so that a store which changes a byte on
// the way in or out is caught doing it.
func child(pseudonym string) *profile.Profile {
	return profile.New(profile.Student{
		ExcludedSkills: []string{},
		Grade:          3,
		Interests:      []string{"космос", "恐竜", "كرة القدم"},
		Notes:          "Loves «головоломки» 🧩 and שחמט.",
		Pseudonym:      pseudonym,
	}, "storetest", moment)
}

// renamed moves a profile on the way a tool does before it writes: a change,
// and the touch every write carries.
func renamed(p *profile.Profile, pseudonym string) *profile.Profile {
	p.Student.Pseudonym = pseudonym
	p.Touch("storetest", moment)
	return p
}

// fileOf is the profile as a store keeps it: what comes back has to match it
// byte for byte.
func fileOf(t *testing.T, p *profile.Profile) []byte {
	t.Helper()

	raw, err := profile.Marshal(p)
	if err != nil {
		t.Fatalf("Marshal() error = %v, want nil", err)
	}
	return raw
}

// create keeps a first profile for the account, and ends the case if it
// cannot.
func create(t *testing.T, h Harness, account store.Account, p *profile.Profile) store.Revision {
	t.Helper()

	revision, err := h.Storage.Create(t.Context(), account, p)
	if err != nil {
		t.Fatalf("Create(%v) error = %v, want nil", account, err)
	}
	return revision
}

// load reads the account's profile, and ends the case if it cannot.
func load(t *testing.T, h Harness, account store.Account) (*profile.Profile, store.Revision) {
	t.Helper()

	p, revision, err := h.Storage.Load(t.Context(), account)
	if err != nil {
		t.Fatalf("Load(%v) error = %v, want nil", account, err)
	}
	return p, revision
}

// holds checks that the account's profile is the file given, byte for byte.
func holds(t *testing.T, h Harness, account store.Account, want []byte) {
	t.Helper()

	p, _ := load(t, h, account)
	if got := fileOf(t, p); !bytes.Equal(got, want) {
		t.Errorf("Load(%v) =\n%s\nwant\n%s", account, got, want)
	}
}

// hasNone checks that the account has no profile.
func hasNone(t *testing.T, h Harness, account store.Account) {
	t.Helper()

	if _, _, err := h.Storage.Load(t.Context(), account); !errors.Is(err, store.ErrNotFound) {
		t.Errorf("Load(%v) error = %v, want %v", account, err, store.ErrNotFound)
	}
}

func noProfileIsMadeBySaving(t *testing.T, h Harness) {
	nobody := someone("nobody")

	hasNone(t, h, nobody)
	if _, err := h.Storage.Export(t.Context(), nobody); !errors.Is(err, store.ErrNotFound) {
		t.Errorf("Export() error = %v, want %v", err, store.ErrNotFound)
	}
	// Whatever revision a save is made from, there is nothing to lay it over.
	var none store.Revision
	if _, err := h.Storage.Save(t.Context(), nobody, child("Mia"), none); !errors.Is(err, store.ErrNotFound) {
		t.Errorf("Save() error = %v, want %v", err, store.ErrNotFound)
	}
	hasNone(t, h, nobody)
}

func aCreatedProfileComesBackAsItWentIn(t *testing.T, h Harness) {
	mia := someone("mia")
	p := child("Миша 🦊")
	want := fileOf(t, p)

	created := create(t, h, mia, p)
	got, revision := load(t, h, mia)
	if file := fileOf(t, got); !bytes.Equal(file, want) {
		t.Errorf("Load() =\n%s\nwant the profile as it was created:\n%s", file, want)
	}
	if revision != created {
		t.Errorf("Load() revision = %q, want %q, the one Create answered", revision, created)
	}
}

func aProfileIsCreatedOnce(t *testing.T, h Harness) {
	mia := someone("mia")
	first := child("Mia")
	want := fileOf(t, first)
	create(t, h, mia, first)

	if _, err := h.Storage.Create(t.Context(), mia, child("A second Mia")); !errors.Is(err, store.ErrConflict) {
		t.Errorf("a second Create() error = %v, want %v", err, store.ErrConflict)
	}
	holds(t, h, mia, want)
}

func aSaveMovesTheProfileOn(t *testing.T, h Harness) {
	mia := someone("mia")
	create(t, h, mia, child("Mia"))

	p, read := load(t, h, mia)
	moved := renamed(p, "Mia the brave")
	want := fileOf(t, moved)
	saved, err := h.Storage.Save(t.Context(), mia, moved, read)
	if err != nil {
		t.Fatalf("Save() error = %v, want nil", err)
	}
	if saved == read {
		t.Errorf("Save() revision = %q, the one it was made from; want a new one", saved)
	}

	got, revision := load(t, h, mia)
	if file := fileOf(t, got); !bytes.Equal(file, want) {
		t.Errorf("Load() =\n%s\nwant the profile as it was saved:\n%s", file, want)
	}
	if revision != saved {
		t.Errorf("Load() revision = %q, want %q, the one Save answered", revision, saved)
	}
}

// Two tabs read the same profile. The first to write wins, and the second is
// told so rather than laying its state over the first one's.
func aSaveFromARevisionThatMovedOnIsRefused(t *testing.T, h Harness) {
	mia := someone("mia")
	create(t, h, mia, child("Mia"))
	first, read := load(t, h, mia)
	second, _ := load(t, h, mia)

	firstTab := renamed(first, "The first tab")
	want := fileOf(t, firstTab)
	if _, err := h.Storage.Save(t.Context(), mia, firstTab, read); err != nil {
		t.Fatalf("the first Save() error = %v, want nil", err)
	}
	if _, err := h.Storage.Save(t.Context(), mia, renamed(second, "The second tab"), read); !errors.Is(err, store.ErrConflict) {
		t.Errorf("the second Save() error = %v, want %v", err, store.ErrConflict)
	}
	holds(t, h, mia, want)
}

// The parent renames the child in the file itself, between the service's read
// and its write. The service's write is refused, and the parent's edit stays.
func aChangeMadeOutsideIsNotWrittenOver(t *testing.T, h Harness) {
	mia := someone("mia")
	create(t, h, mia, child("Mia"))
	ours, read := load(t, h, mia)
	theirs, _ := load(t, h, mia)

	theirs.Student.Pseudonym = "Renamed by hand"
	byHand := fileOf(t, theirs)
	h.Plant(t, mia, byHand)

	if _, err := h.Storage.Save(t.Context(), mia, renamed(ours, "Renamed by the service"), read); !errors.Is(err, store.ErrConflict) {
		t.Errorf("Save() error = %v, want %v", err, store.ErrConflict)
	}
	holds(t, h, mia, byHand)
}

// A write moves the profile's own revision on by exactly one: the history is
// pinned by that number, and an early read is noticed by it. A profile that
// was not touched, or was touched twice, is refused — and as none of the
// refusals anybody branches on, because nothing but fixing the code helps.
func aProfileNotMovedOnByOneWriteIsRefused(t *testing.T, h Harness) {
	mia := someone("mia")
	p := child("Mia")
	want := fileOf(t, p)
	created := create(t, h, mia, p)

	untouched, _ := load(t, h, mia)
	untouched.Student.Pseudonym = "Changed without a touch"
	twice, _ := load(t, h, mia)
	twice.Touch("storetest", moment)
	twice.Touch("storetest", moment)
	branchedOn := []error{
		store.ErrNotFound, store.ErrConflict, store.ErrCorrupted, store.ErrAccessRevoked,
		profile.ErrInvalid, profile.ErrNewer,
	}
	for how, q := range map[string]*profile.Profile{"not touched": untouched, "touched twice": twice} {
		_, err := h.Storage.Save(t.Context(), mia, q, created)
		if err == nil {
			t.Errorf("Save() of a profile %s: error = nil, want a refusal", how)
		}
		for _, refusal := range branchedOn {
			if errors.Is(err, refusal) {
				t.Errorf("Save() of a profile %s: error = %v, want none of the refusals a caller acts on", how, err)
			}
		}
	}
	holds(t, h, mia, want)

	once, _ := load(t, h, mia)
	if _, err := h.Storage.Save(t.Context(), mia, renamed(once, "Touched once"), created); err != nil {
		t.Errorf("Save() of a profile touched once: error = %v, want nil", err)
	}
}

// A profile the service could not read back is not written: not as a first
// profile, not over one that is there, and not from any revision. Its own
// rules are looked at before anything else, so every attempt is refused for
// what it is rather than for where it was aimed — and a refused write moves
// nothing, not even the revision.
func aProfileThatBreaksItsRulesIsNotWritten(t *testing.T, h Harness) {
	mia, leo := someone("mia"), someone("leo")
	p := child("Mia")
	want := fileOf(t, p)
	created := create(t, h, mia, p)
	current, _ := load(t, h, mia)
	nameless := renamed(current, "")
	var notCurrent store.Revision

	for _, attempt := range []struct {
		aimed string
		write func() error
	}{
		{"created for an account with none", func() error {
			_, err := h.Storage.Create(t.Context(), leo, child(""))
			return err
		}},
		{"created over a profile", func() error {
			_, err := h.Storage.Create(t.Context(), mia, child(""))
			return err
		}},
		{"saved to an account with none", func() error {
			_, err := h.Storage.Save(t.Context(), leo, nameless, created)
			return err
		}},
		{"saved from a revision that is not the current one", func() error {
			_, err := h.Storage.Save(t.Context(), mia, nameless, notCurrent)
			return err
		}},
		{"saved from the current revision", func() error {
			_, err := h.Storage.Save(t.Context(), mia, nameless, created)
			return err
		}},
	} {
		if err := attempt.write(); !errors.Is(err, profile.ErrInvalid) {
			t.Errorf("a profile with no pseudonym %s: error = %v, want %v", attempt.aimed, err, profile.ErrInvalid)
		}
	}
	hasNone(t, h, leo)
	holds(t, h, mia, want)

	named, _ := load(t, h, mia)
	if _, err := h.Storage.Save(t.Context(), mia, renamed(named, "Mia again"), created); err != nil {
		t.Errorf("Save() from the revision the refused saves were made from: error = %v, want nil", err)
	}
}

// A call that names no one reaches no one's profile, even with a token that
// opens somebody's: whatever went wrong on the way to the store, what it holds
// for the owner of that token is neither read, nor written, nor pointed to.
func anAccountThatNamesNoOneReachesNoProfile(t *testing.T, h Harness) {
	mia := someone("mia")
	p := child("Mia")
	want := fileOf(t, p)
	created := create(t, h, mia, p)
	current, _ := load(t, h, mia)
	nameless := store.NewAccount("", mia.Token())

	if _, err := h.Storage.Create(t.Context(), nameless, child("Nobody")); err == nil {
		t.Error("Create() for an account that names no one: error = nil, want a refusal")
	}
	if got, _, err := h.Storage.Load(t.Context(), nameless); err == nil || got != nil {
		t.Errorf("Load() for an account that names no one = %v, %v, want no profile and a refusal", got, err)
	}
	if _, err := h.Storage.Save(t.Context(), nameless, renamed(current, "Nobody"), created); err == nil {
		t.Error("Save() for an account that names no one: error = nil, want a refusal")
	}
	if _, err := h.Storage.Export(t.Context(), nameless); err == nil {
		t.Error("Export() for an account that names no one: error = nil, want a refusal")
	}
	holds(t, h, mia, want)
}

// A profile handed to a store, or handed out by one, is the caller's own:
// whatever the caller does to it afterwards never reaches what the store keeps.
func whatACallerHoldsIsNotWhatTheStoreKeeps(t *testing.T, h Harness) {
	mia := someone("mia")
	p := child("Mia")
	want := fileOf(t, p)
	created := create(t, h, mia, p)
	spoil(p)
	holds(t, h, mia, want)

	loaded, _ := load(t, h, mia)
	spoil(loaded)
	holds(t, h, mia, want)

	again, _ := load(t, h, mia)
	moved := renamed(again, "Mia the brave")
	want = fileOf(t, moved)
	if _, err := h.Storage.Save(t.Context(), mia, moved, created); err != nil {
		t.Fatalf("Save() error = %v, want nil", err)
	}
	spoil(moved)
	holds(t, h, mia, want)
}

// spoil changes a profile in place: a field, an element of a list and an
// entry of a map, each of which a store that kept the caller's memory would
// then hand back.
func spoil(p *profile.Profile) {
	p.Student.Pseudonym = "Spoiled"
	p.Student.Interests[0] = "spoiled"
	p.Topics["spoiled"] = profile.Topic{}
}

func accountsDoNotSeeEachOther(t *testing.T, h Harness) {
	mia, leo := someone("mia"), someone("leo")
	mias := child("Mia")
	miasFile := fileOf(t, mias)
	create(t, h, mia, mias)
	hasNone(t, h, leo)

	leos := child("Leo")
	leosFile := fileOf(t, leos)
	create(t, h, leo, leos)
	holds(t, h, mia, miasFile)
	holds(t, h, leo, leosFile)
}

// A file nothing can read is reported as damage, whatever shape the damage
// took, and no profile comes with the report. Nor is it replaced by a new
// profile: starting over is the parent's decision.
func aDamagedFileIsReportedAsDamage(t *testing.T, h Harness) {
	for _, damage := range []struct{ name, raw string }{
		{"an empty file", ""},
		{"a write cut short", `{"schema_version": 1, "student": {"pseud`},
		{"a file that is not a profile", fmt.Sprintf(`{"schema_version": %d}`, profile.Version)},
		{"a shape older than anything this build reads", `{"schema_version": 0}`},
	} {
		t.Run(damage.name, func(t *testing.T) {
			damaged := someone(damage.name)
			h.Plant(t, damaged, []byte(damage.raw))

			p, _, err := h.Storage.Load(t.Context(), damaged)
			if !errors.Is(err, store.ErrCorrupted) {
				t.Errorf("Load() error = %v, want %v", err, store.ErrCorrupted)
			}
			if p != nil {
				t.Error("Load() returned a profile alongside the refusal")
			}
			_, err = h.Storage.Create(t.Context(), damaged, child("Mia"))
			if !errors.Is(err, store.ErrConflict) {
				t.Errorf("Create() over a damaged file: error = %v, want %v", err, store.ErrConflict)
			}
		})
	}
}

// A file from a newer build is readable, only not by this one: it is refused
// as newer and never as damage, and nothing is put in its place.
func aFileFromANewerBuildIsNotDamage(t *testing.T, h Harness) {
	mia := someone("mia")
	h.Plant(t, mia, fmt.Appendf(nil, `{"schema_version": %d}`, profile.Version+1))

	p, _, err := h.Storage.Load(t.Context(), mia)
	if !errors.Is(err, profile.ErrNewer) {
		t.Errorf("Load() error = %v, want %v", err, profile.ErrNewer)
	}
	if errors.Is(err, store.ErrCorrupted) {
		t.Errorf("Load() error = %v, want it not to be %v", err, store.ErrCorrupted)
	}
	if p != nil {
		t.Error("Load() returned a profile alongside the refusal")
	}
	_, err = h.Storage.Create(t.Context(), mia, child("Mia"))
	if !errors.Is(err, store.ErrConflict) {
		t.Errorf("Create() over a newer file: error = %v, want %v", err, store.ErrConflict)
	}
}

// Where the file is depends on the store; that an account which has one is
// answered at all does not.
func anAccountWithAProfileCanBeToldWhereItIs(t *testing.T, h Harness) {
	mia := someone("mia")
	create(t, h, mia, child("Mia"))

	if _, err := h.Storage.Export(t.Context(), mia); err != nil {
		t.Errorf("Export() error = %v, want nil", err)
	}
}

// A call whose context has ended does nothing at all, and says why.
func aFinishedContextStopsEveryOperation(t *testing.T, h Harness) {
	mia, leo := someone("mia"), someone("leo")
	p := child("Mia")
	want := fileOf(t, p)
	created := create(t, h, mia, p)
	loaded, _ := load(t, h, mia)
	moved := renamed(loaded, "Too late")

	finished, cancel := context.WithCancel(t.Context())
	cancel()

	_, err := h.Storage.Create(finished, leo, child("Leo"))
	wantFinished(t, "Create", err)
	_, _, err = h.Storage.Load(finished, mia)
	wantFinished(t, "Load", err)
	_, err = h.Storage.Save(finished, mia, moved, created)
	wantFinished(t, "Save", err)
	_, err = h.Storage.Export(finished, mia)
	wantFinished(t, "Export", err)

	hasNone(t, h, leo)
	holds(t, h, mia, want)
	_, err = h.Storage.Save(t.Context(), mia, moved, created)
	if err != nil {
		t.Errorf("Save() from the revision the finished calls were made from: error = %v, want nil", err)
	}
}

// wantFinished checks that an operation refused because its context ended.
func wantFinished(t *testing.T, operation string, err error) {
	t.Helper()

	if !errors.Is(err, context.Canceled) {
		t.Errorf("%s() with a finished context: error = %v, want %v", operation, err, context.Canceled)
	}
}

// Saves made from one revision at the same moment — the widget and the model
// answering together — leave one whole profile that one of them wrote. Whether
// every other one is refused is more than a store may be able to promise: one
// that can compare revisions only before it writes can let two land, and then
// the last of them is what stays.
func savesRacingFromOneRevisionLeaveOneWholeProfile(t *testing.T, h Harness) {
	mia := someone("mia")
	create(t, h, mia, child("Mia"))
	_, read := load(t, h, mia)

	const racers = 8
	moves := make([]*profile.Profile, racers)
	files := make([][]byte, racers)
	for i := range moves {
		p, _ := load(t, h, mia)
		moves[i] = renamed(p, fmt.Sprintf("Racer %d", i))
		files[i] = fileOf(t, moves[i])
	}

	start := make(chan struct{})
	outcomes := make([]error, racers)
	var racing sync.WaitGroup
	for i, p := range moves {
		racing.Go(func() {
			<-start
			_, outcomes[i] = h.Storage.Save(t.Context(), mia, p, read)
		})
	}
	close(start)
	racing.Wait()

	landed := 0
	for i, err := range outcomes {
		switch {
		case err == nil:
			landed++
		case !errors.Is(err, store.ErrConflict):
			t.Errorf("racer %d: Save() error = %v, want nil or %v", i, err, store.ErrConflict)
		}
	}
	if landed == 0 {
		t.Error("no racing Save() landed, want at least one")
	}

	got, _ := load(t, h, mia)
	final := fileOf(t, got)
	if !slices.ContainsFunc(files, func(racer []byte) bool { return bytes.Equal(racer, final) }) {
		t.Errorf("after the race the store holds\n%s\nwhich no racer wrote", final)
	}
}

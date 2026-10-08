package drivestore_test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"maps"
	"net/http"
	"sync"
	"testing"
	"time"

	"go.uber.org/zap/zapcore"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/profile"
	"github.com/MathTrail/mathtrail-standalone/internal/infra/drive/drivetest"
	"github.com/MathTrail/mathtrail-standalone/internal/store"
)

// loaded is the account's profile and the revision it was read at, or the end
// of the case.
func (f *fixture) loaded(t *testing.T, s store.Storage, account store.Account) (*profile.Profile, store.Revision) {
	t.Helper()

	p, revision, err := s.Load(t.Context(), account)
	if err != nil {
		t.Fatalf("Load() error = %v, want nil", err)
	}
	return p, revision
}

// saved writes the profile moved on over the revision given, or ends the case.
func (f *fixture) saved(t *testing.T, s store.Storage, account store.Account, p *profile.Profile, from store.Revision) store.Revision {
	t.Helper()

	revision, err := s.Save(t.Context(), account, p, from)
	if err != nil {
		t.Fatalf("Save() error = %v, want nil", err)
	}
	return revision
}

// linesOf is every line of an event the store left.
func (f *fixture) linesOf(event string) []map[string]any {
	var fields []map[string]any
	lines := f.logs.FilterMessage(event).All()
	for i := range lines {
		fields = append(fields, lines[i].ContextMap())
	}
	return fields
}

// A pause Drive asks for is ridden out: the call is made again after about
// half a second, and again after about a second and a half, and the line of
// the call says how many times it was made again.
func TestAPauseIsRiddenOut(t *testing.T) {
	t.Parallel()

	f := newFixture(t, 5*time.Second)
	f.create(t, mia, child("Mia"))
	f.fake.Fail(drivetest.Download, http.StatusTooManyRequests, "rateLimitExceeded")
	f.fake.Fail(drivetest.Download, http.StatusForbidden, "userRateLimitExceeded")
	f.fake.ResetCalls()
	f.logs.TakeAll()

	f.loaded(t, f.storage, mia)
	if calls := f.fake.Calls(); !maps.Equal(calls, drivetest.Calls{"download": 3}) {
		t.Errorf("Load() through two pauses cost %v, want the download three times", calls)
	}
	waited := f.waits.taken()
	if len(waited) != 2 {
		t.Fatalf("Load() waited %v, want two pauses", waited)
	}
	for i, around := range []time.Duration{500 * time.Millisecond, 1500 * time.Millisecond} {
		if low, high := around*3/4, around*5/4; waited[i] < low || waited[i] > high {
			t.Errorf("pause %d = %v, want %v give or take a quarter", i+1, waited[i], around)
		}
	}
	lines := f.linesOf("drive_call")
	if len(lines) != 1 || lines[0]["retries"] != int64(2) || lines[0]["outcome"] != "ok" {
		t.Errorf("drive_call lines = %v, want one download made again twice, ok", lines)
	}
}

// wantGivenUp holds the error of a call whose caller gave up to being the
// caller's own, and to none of the store's refusals, which a caller would act
// on.
func wantGivenUp(t *testing.T, call string, err error) {
	t.Helper()

	if !errors.Is(err, context.Canceled) {
		t.Errorf("%s given up on: error = %v, want %v", call, err, context.Canceled)
	}
	for _, refusal := range []error{store.ErrBehind, store.ErrCorrupted, store.ErrNotFound, store.ErrUnavailable} {
		if errors.Is(err, refusal) {
			t.Errorf("%s given up on: error = %v, want it not to be %v", call, err, refusal)
		}
	}
}

// A call whose caller gives up during a pause Drive asked for is not made
// again: it ends with the caller's own error, not as Drive out of reach, and
// its line says it was given up on after no retry — at info, since nothing
// failed.
func TestAPauseGivenUpOnIsNotMadeAgain(t *testing.T) {
	t.Parallel()

	f := newFixture(t, 5*time.Second)
	f.create(t, mia, child("Mia"))
	f.fake.Fail(drivetest.Download, http.StatusTooManyRequests, "rateLimitExceeded")
	f.fake.ResetCalls()
	f.logs.TakeAll()

	_, _, err := f.storage.Load(f.givingUpDuringTheNextWait(t), mia)
	wantGivenUp(t, "Load()", err)
	if calls := f.fake.Calls(); !maps.Equal(calls, drivetest.Calls{"download": 1}) {
		t.Errorf("Load() given up on during a pause cost %v, want the one download Drive asked to pause", calls)
	}
	lines := f.logs.FilterMessage("drive_call").All()
	if len(lines) != 1 || lines[0].Level != zapcore.InfoLevel ||
		lines[0].ContextMap()["outcome"] != "canceled" || lines[0].ContextMap()["retries"] != int64(0) {
		t.Errorf("drive_call lines = %v, want one at info, canceled after no retry", lines)
	}
}

// A write Drive failed on may have landed, so it is not sent again: whoever
// made it reads again instead. A write Drive asked to pause did not land, and
// is sent again.
func TestAWriteDriveFailedOnIsNotSentAgain(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name    string
		status  int
		reason  string
		want    error
		updates int
	}{
		{"a failure of Drive's", http.StatusServiceUnavailable, "backendError", store.ErrUnavailable, 1},
		{"a pause Drive asks for", http.StatusTooManyRequests, "rateLimitExceeded", nil, 2},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			f := newFixture(t, 5*time.Second)
			f.create(t, mia, child("Mia"))
			p, read := f.loaded(t, f.storage, mia)
			f.fake.Fail(drivetest.Update, tc.status, tc.reason)
			f.fake.ResetCalls()

			_, err := f.storage.Save(t.Context(), mia, moved(p, "Mia the brave"), read)
			if !errors.Is(err, tc.want) || (tc.want == nil) != (err == nil) {
				t.Errorf("Save() error = %v, want %v", err, tc.want)
			}
			if got := f.fake.Calls()["update"]; got != tc.updates {
				t.Errorf("Save() sent the upload %d times, want %d", got, tc.updates)
			}
		})
	}
}

// A call that ran out of its time is not made again: it has spent what a call
// is given.
func TestACallThatTookTooLongIsNotMadeAgain(t *testing.T) {
	t.Parallel()

	f := newFixture(t, 200*time.Millisecond)
	f.create(t, mia, child("Mia"))
	f.fake.Stall(drivetest.Download)
	f.fake.ResetCalls()

	if _, _, err := f.storage.Load(t.Context(), mia); err == nil {
		t.Fatal("Load() of a call that took too long: error = nil, want one")
	}
	if calls := f.fake.Calls(); !maps.Equal(calls, drivetest.Calls{"download": 1}) {
		t.Errorf("Load() cost %v, want the one download that took too long", calls)
	}
}

// No call is started that could outlive the token the account reaches its
// Drive with: the operation is refused as access that ends too soon, before
// Drive is asked anything, and the line says so as a failure of the service's
// own. A token with time left reaches the Drive as ever.
func TestNoCallOutlivesTheToken(t *testing.T) {
	t.Parallel()

	const timeout = 10 * time.Second
	f := newFixture(t, timeout)
	f.create(t, mia, child("Mia"))
	f.fake.ResetCalls()
	f.logs.TakeAll()

	ending := store.NewAccount(mia.ID, miaToken, moment.Add(timeout-time.Second))
	if _, _, err := f.storage.Load(t.Context(), ending); !errors.Is(err, store.ErrAccessExpired) {
		t.Errorf("Load() with a token that ends first: error = %v, want %v", err, store.ErrAccessExpired)
	}
	if calls := f.fake.Calls(); len(calls) != 0 {
		t.Errorf("Load() with a token that ends first cost %v, want no call", calls)
	}
	lines := f.logs.FilterMessage("drive_call").All()
	if len(lines) != 1 || lines[0].ContextMap()["outcome"] != "expired" || lines[0].Level != zapcore.WarnLevel {
		t.Errorf("drive_call lines = %v, want one warning that the token ends first", lines)
	}

	lasting := store.NewAccount(mia.ID, miaToken, moment.Add(timeout+time.Second))
	f.loaded(t, f.storage, lasting)
}

// Within one instance the writes of an account take turns: of saves made from
// one revision at the same moment — two tabs on the instance — exactly one
// lands, and every other is refused as a conflict, having read what the first
// one wrote.
func TestWritesOfOneAccountTakeTurnsOnAnInstance(t *testing.T) {
	t.Parallel()

	f := newFixture(t, 5*time.Second)
	f.create(t, mia, child("Mia"))
	_, read := f.loaded(t, f.storage, mia)

	const racers = 8
	start := make(chan struct{})
	outcomes := make([]error, racers)
	var racing sync.WaitGroup
	for i := range racers {
		p, _ := f.loaded(t, f.storage, mia)
		written := moved(p, fmt.Sprintf("Racer %d", i))
		racing.Go(func() {
			<-start
			_, outcomes[i] = f.storage.Save(t.Context(), mia, written, read)
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
	if landed != 1 {
		t.Errorf("%d racing saves landed on one instance, want exactly one", landed)
	}
	if lines := f.linesOf("drive_conflict"); len(lines) != racers-1 {
		t.Errorf("drive_conflict lines = %d, want one for each refused save", len(lines))
	}
}

// writeUnderWay starts a save of the profile whose upload Drive holds, so
// that it keeps the account's turn, and answers what lets the upload land and
// waits for the save to end. The calls, and their lines, are counted from the
// save's start.
func (f *fixture) writeUnderWay(t *testing.T, p *profile.Profile, from store.Revision) (land func() error) {
	t.Helper()

	letGo := f.fake.Hold(drivetest.Update)
	f.fake.ResetCalls()
	f.logs.TakeAll()
	landed := make(chan error, 1)
	go func() {
		_, err := f.storage.Save(t.Context(), mia, p, from)
		landed <- err
	}()
	for deadline := time.Now().Add(10 * time.Second); f.fake.Calls()["update"] == 0; {
		select {
		case err := <-landed:
			t.Fatalf("Save() under way ended before its upload: error = %v, want it held in the upload", err)
		default:
			if time.Now().After(deadline) {
				t.Fatal("Save() under way never reached its upload, want it held there")
			}
			time.Sleep(time.Millisecond)
		}
	}
	return func() error {
		letGo()
		return <-landed
	}
}

// everyCallEndedWell holds every call of Drive's the store made since its
// lines were last taken to having ended well. A call made once its caller had
// given up never reaches Drive, and is told by its line alone.
func (f *fixture) everyCallEndedWell(t *testing.T) {
	t.Helper()

	for _, line := range f.linesOf("drive_call") {
		if line["outcome"] != "ok" {
			t.Errorf("a %v call ended %v, want every call made to have ended well", line["op"], line["outcome"])
		}
	}
}

// wroteNothingBut holds the account's Drive to the write under way alone: no
// other upload, no file made, nothing set aside, and no call of Drive's that
// went any way but well.
func (f *fixture) wroteNothingBut(t *testing.T, underWay []byte) {
	t.Helper()

	if calls := f.fake.Calls(); calls["update"] != 1 || calls["create"] != 0 {
		t.Errorf("the calls made were %v, want the upload under way alone and no file made", calls)
	}
	if got := f.profileFile(t, miaToken).Content; !bytes.Equal(got, underWay) {
		t.Errorf("the profile holds\n%s\nwant the write under way:\n%s", got, underWay)
	}
	if aside := setAside(f.fake, miaToken); len(aside) != 0 {
		t.Errorf("%d files were set aside, want none", len(aside))
	}
	f.everyCallEndedWell(t)
}

// A call that waits for the account's turn while another write holds it — a
// slow upload from another tab — and whose caller gives up meanwhile ends
// with the caller's own error, and is never made: nothing more is asked of
// Drive for it, it writes nothing, and the write under way lands as it would
// have.
func TestACallWhoseCallerGivesUpWaitingForItsTurnIsNeverMade(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name string
		// spoil readies the file for the call, once the write under way holds
		// the turn.
		spoil func(t *testing.T, f *fixture)
		// call makes the call that waits, with a profile and the revision it
		// was read at.
		call func(ctx context.Context, s store.Storage, p *profile.Profile, read store.Revision) error
	}{
		{"Create", nil, func(ctx context.Context, s store.Storage, _ *profile.Profile, _ store.Revision) error {
			_, err := s.Create(ctx, mia, child("A second Mia"))
			return err
		}},
		{"Save", nil, func(ctx context.Context, s store.Storage, p *profile.Profile, read store.Revision) error {
			_, err := s.Save(ctx, mia, moved(p, "Mia from another tab"), read)
			return err
		}},
		{"StartOver", nil, func(ctx context.Context, s store.Storage, _ *profile.Profile, _ store.Revision) error {
			_, err := s.StartOver(ctx, mia, child("Mia anew"))
			return err
		}},
		// A recovery reads the file and its history before it waits for the
		// turn to write, so the file it reads has to be damaged.
		{"Restore", func(t *testing.T, f *fixture) { plant(t, f.fake, miaToken, damage) },
			func(ctx context.Context, s store.Storage, _ *profile.Profile, _ store.Revision) error {
				_, _, err := s.Restore(ctx, mia)
				return err
			}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			f := newFixture(t, 5*time.Second)
			f.create(t, mia, child("Mia"))
			first, read := f.loaded(t, f.storage, mia)
			second, _ := f.loaded(t, f.storage, mia)
			underWay := moved(first, "Mia under way")
			land := f.writeUnderWay(t, underWay, read)
			if tc.spoil != nil {
				tc.spoil(t, f)
			}

			// The caller gives up once the call waits for its turn: well after
			// the few reads of Drive a recovery makes before it does, even on a
			// machine the race detector and the other tests slow down. A read
			// the deadline cut short would be a call of Drive's that did not
			// end well, which this case holds the store to never making.
			ctx, cancel := context.WithTimeout(t.Context(), 2*time.Second)
			defer cancel()
			err := tc.call(ctx, f.storage, second, read)
			if landed := land(); landed != nil {
				t.Fatalf("Save() under way error = %v, want nil", landed)
			}
			if !errors.Is(err, context.DeadlineExceeded) {
				t.Errorf("%s() given up on while it waited for its turn: error = %v, want %v", tc.name, err, context.DeadlineExceeded)
			}
			f.wroteNothingBut(t, fileOf(t, underWay))
		})
	}
}

// The window that remains: two instances share nothing, and two saves from
// one revision whose uploads overlap — each read the file again before the
// other wrote — both land. The later upload is what the file holds, whole.
func TestTwoInstancesCanStillBothLandWithinOneUpload(t *testing.T) {
	t.Parallel()

	f := newFixture(t, 5*time.Second)
	f.create(t, mia, child("Mia"))
	other := f.instance(t, 5*time.Second)
	first, read := f.loaded(t, f.storage, mia)
	second, _ := f.loaded(t, other, mia)
	firstFile := fileOf(t, moved(first, "The first instance"))

	letGo := f.fake.Hold(drivetest.Update)
	f.fake.ResetCalls()
	firstDone := make(chan error, 1)
	go func() {
		_, err := f.storage.Save(t.Context(), mia, first, read)
		firstDone <- err
	}()
	for f.fake.Calls()["update"] == 0 {
		time.Sleep(time.Millisecond)
	}
	// The first upload is under way; the second instance reads the file again,
	// finds it as it was, and writes.
	f.saved(t, other, mia, moved(second, "The second instance"), read)
	letGo()
	if err := <-firstDone; err != nil {
		t.Fatalf("the first Save() error = %v, want nil", err)
	}
	if got := f.profileFile(t, miaToken).Content; !bytes.Equal(got, firstFile) {
		t.Errorf("the file holds\n%s\nwant the later upload:\n%s", got, firstFile)
	}
}

// writtenThenBehind writes the profile once more, and has Drive answer the
// next downloads of it with what it held before, as many times as given:
// Drive does not promise that a read right after a write sees it. It answers
// the profile as written.
func (f *fixture) writtenThenBehind(t *testing.T, downloads int) *profile.Profile {
	t.Helper()

	f.create(t, mia, child("Mia"))
	p, read := f.loaded(t, f.storage, mia)
	written := moved(p, "Mia the brave")
	f.saved(t, f.storage, mia, written, read)
	f.fake.Lag(miaToken, f.profileFile(t, miaToken).ID, downloads)
	f.waits.taken()
	return written
}

// staleReadSaid checks that a read came back behind once, waited a moment and
// read again, and that its line says how that ended.
func (f *fixture) staleReadSaid(t *testing.T, outcome string) {
	t.Helper()

	if waited := f.waits.taken(); len(waited) != 1 || waited[0] != 250*time.Millisecond {
		t.Errorf("Load() waited %v, want one moment of 250ms", waited)
	}
	if lines := f.linesOf("drive_stale_read"); len(lines) != 1 || lines[0]["outcome"] != outcome {
		t.Errorf("drive_stale_read lines = %v, want one, %s", lines, outcome)
	}
}

// A read that comes back with an earlier state than one the instance wrote is
// read once more after a moment, and goes on once it has caught up.
func TestAReadThatCatchesUpGoesOn(t *testing.T) {
	t.Parallel()

	f := newFixture(t, 5*time.Second)
	written := f.writtenThenBehind(t, 1)

	got, _ := f.loaded(t, f.storage, mia)
	if got.Revision != written.Revision {
		t.Errorf("Load() revision %d, want the one written, %d", got.Revision, written.Revision)
	}
	f.staleReadSaid(t, "caught_up")
}

// A read that is still behind after a moment is refused as behind — once: the
// next read takes the file as it is, whatever put it back.
func TestAReadThatStaysBehindIsRefusedOnce(t *testing.T) {
	t.Parallel()

	f := newFixture(t, 5*time.Second)
	f.writtenThenBehind(t, 3)

	if _, _, err := f.storage.Load(t.Context(), mia); !errors.Is(err, store.ErrBehind) {
		t.Fatalf("Load() still behind: error = %v, want %v", err, store.ErrBehind)
	}
	f.staleReadSaid(t, "behind")
	if _, _, err := f.storage.Load(t.Context(), mia); err != nil {
		t.Errorf("Load() after a read refused as behind: error = %v, want the file as it is", err)
	}
}

// A file gone when a read behind a write looks at it again is a file Drive no
// longer has: told as that, and forgotten, so that the next call searches for
// where the profile is now rather than asking for the file again.
func TestAFileGoneWhileAReadWaitedToCatchUpIsForgotten(t *testing.T) {
	t.Parallel()

	f := newFixture(t, 5*time.Second)
	written := f.writtenThenBehind(t, 1)
	letThrough(f.fake, drivetest.Download, 1)
	f.fake.Fail(drivetest.Download, http.StatusNotFound, "notFound")

	if _, _, err := f.storage.Load(t.Context(), mia); !errors.Is(err, store.ErrNotFound) {
		t.Fatalf("Load() of a file gone on the second look: error = %v, want %v", err, store.ErrNotFound)
	}
	f.fake.ResetCalls()
	got, _ := f.loaded(t, f.storage, mia)
	if calls := f.fake.Calls(); !maps.Equal(calls, drivetest.Calls{"list": 1, "download": 1}) {
		t.Errorf("Load() after the file was gone cost %v, want a search and a download", calls)
	}
	if got.Revision != written.Revision {
		t.Errorf("Load() after the file was gone read revision %d, want the one written, %d", got.Revision, written.Revision)
	}
}

// A read that waits for Drive to catch up — with a write of this instance's,
// or with a file it just put back — and whose caller gives up meanwhile ends
// there, with the caller's own error: never told as behind, nor as damage the
// parent would be asked about, and nothing more is asked of Drive. Nor is the
// write waited for let go of: the next read still finds it.
func TestAReadGivenUpOnWhileItWaitsEndsThere(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name string
		// written writes the profile, has Drive answer its next two downloads
		// early, and answers the profile as written.
		written func(t *testing.T, f *fixture) *profile.Profile
		// calls are what the read costs until it waits.
		calls drivetest.Calls
	}{
		{"behind a write", func(t *testing.T, f *fixture) *profile.Profile { return f.writtenThenBehind(t, 2) },
			drivetest.Calls{"download": 1}},
		{"of a file just put back", func(t *testing.T, f *fixture) *profile.Profile { return f.putBackThenEarly(t, 2) },
			drivetest.Calls{"download": 2, "list": 1}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			f := newFixture(t, 5*time.Second)
			written := tc.written(t, f)
			f.fake.ResetCalls()
			f.logs.TakeAll()

			_, _, err := f.storage.Load(f.givingUpDuringTheNextWait(t), mia)
			wantGivenUp(t, "Load()", err)
			if calls := f.fake.Calls(); !maps.Equal(calls, tc.calls) {
				t.Errorf("Load() given up on while it waited cost %v, want %v", calls, tc.calls)
			}
			f.everyCallEndedWell(t)
			if got, _ := f.loaded(t, f.storage, mia); got.Revision != written.Revision {
				t.Errorf("Load() after the read given up on = revision %d, want the one written, %d", got.Revision, written.Revision)
			}
		})
	}
}

// A file found by a search is trusted for ten minutes, and then searched for
// again — writing to it does not make it any younger.
func TestARememberedFileIsSearchedForAgainAfterTenMinutes(t *testing.T) {
	t.Parallel()

	f := newFixture(t, 5*time.Second)
	f.create(t, mia, child("Mia"))
	p, read := f.loaded(t, f.storage, mia)
	f.clock.advance(9 * time.Minute)
	f.saved(t, f.storage, mia, moved(p, "Mia the brave"), read)

	f.fake.ResetCalls()
	f.loaded(t, f.storage, mia)
	if calls := f.fake.Calls(); !maps.Equal(calls, drivetest.Calls{"download": 1}) {
		t.Errorf("Load() nine minutes on cost %v, want the one download", calls)
	}
	f.clock.advance(2 * time.Minute)
	f.fake.ResetCalls()
	f.loaded(t, f.storage, mia)
	if calls := f.fake.Calls(); !maps.Equal(calls, drivetest.Calls{"list": 1, "download": 1}) {
		t.Errorf("Load() eleven minutes on cost %v, want a search and a download", calls)
	}
}

// A save made once the file it was read from is no longer trusted finds the
// file by its marker again, and lands there.
func TestASaveAfterTheFileIsNoLongerTrustedFindsItAndLands(t *testing.T) {
	t.Parallel()

	f := newFixture(t, 5*time.Second)
	f.create(t, mia, child("Mia"))
	p, read := f.loaded(t, f.storage, mia)
	written := moved(p, "Mia the brave")
	f.clock.advance(11 * time.Minute)
	f.fake.ResetCalls()

	f.saved(t, f.storage, mia, written, read)
	if calls := f.fake.Calls(); !maps.Equal(calls, drivetest.Calls{"list": 1, "download": 1, "update": 1}) {
		t.Errorf("Save() eleven minutes on cost %v, want a search, the file read again and the upload", calls)
	}
	if got, want := f.profileFile(t, miaToken).Content, fileOf(t, written); !bytes.Equal(got, want) {
		t.Errorf("the file holds\n%s\nwant the save:\n%s", got, want)
	}
}

// A save is made over the state of the file the profile was read from: a
// revision of another file — even one that holds the same bytes — or one no
// store in Drive hands out is refused as a conflict before anything is read
// again or written, and the caller reads again.
func TestASaveFromARevisionNotOfTheProfilesFileIsAConflict(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name string
		// from is the revision the save is made from, given the one the
		// profile was read at.
		from func(t *testing.T, f *fixture, read store.Revision) store.Revision
	}{
		{"a revision of another file", func(t *testing.T, f *fixture, read store.Revision) store.Revision {
			// A second file carrying the profile, holding the very same bytes,
			// changed last: the profile is in it now.
			f.fake.Put(miaToken, &drivetest.File{
				Name: fileName, MimeType: fileType, AppProperties: maps.Clone(profileMarker), Content: f.profileFile(t, miaToken).Content,
			})
			return read
		}},
		{"what no store in Drive hands out", func(*testing.T, *fixture, store.Revision) store.Revision { return "42" }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			f := newFixture(t, 5*time.Second)
			f.create(t, mia, child("Mia"))
			p, read := f.loaded(t, f.storage, mia)
			from := tc.from(t, f, read)
			before := f.fake.Files(miaToken)
			f.fake.ResetCalls()

			_, err := f.instance(t, 5*time.Second).Save(t.Context(), mia, moved(p, "Mia the brave"), from)
			if !errors.Is(err, store.ErrConflict) {
				t.Errorf("Save() error = %v, want %v", err, store.ErrConflict)
			}
			if calls := f.fake.Calls(); calls["download"] != 0 || calls["update"] != 0 {
				t.Errorf("Save() cost %v, want neither a file read again nor an upload", calls)
			}
			f.leftAsTheyWere(t, before)
		})
	}
}

// leftAsTheyWere holds every file of the account's Drive to holding what it
// held before.
func (f *fixture) leftAsTheyWere(t *testing.T, before []*drivetest.File) {
	t.Helper()

	after := f.fake.Files(miaToken)
	if len(after) != len(before) {
		t.Fatalf("the Drive holds %d files, want the %d it held", len(after), len(before))
	}
	for i, file := range after {
		if !bytes.Equal(file.Content, before[i].Content) {
			t.Errorf("file %q holds\n%s\nwant it as it was:\n%s", file.Name, file.Content, before[i].Content)
		}
	}
}

// A write refused because the file changed since it was read leaves a line
// with the profile's number it was read at and the one the file holds now —
// none, when the file holds no profile — so that a reader can tell two writes
// that raced from a file damaged by hand.
func TestAConflictLineSaysWhatTheFileHeld(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name string
		// change changes the file once the profile was read from it.
		change func(t *testing.T, f *fixture)
		found  int64
	}{
		{"moved on by another write", func(t *testing.T, f *fixture) {
			other := f.instance(t, 5*time.Second)
			p, read := f.loaded(t, other, mia)
			f.saved(t, other, mia, moved(p, "Mia from another tab"), read)
		}, 2},
		{"damaged by hand", func(t *testing.T, f *fixture) { plant(t, f.fake, miaToken, damage) }, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			f := newFixture(t, 5*time.Second)
			f.create(t, mia, child("Mia"))
			p, read := f.loaded(t, f.storage, mia)
			tc.change(t, f)
			changed := f.profileFile(t, miaToken).Content

			if _, err := f.storage.Save(t.Context(), mia, moved(p, "Mia the brave"), read); !errors.Is(err, store.ErrConflict) {
				t.Errorf("Save() over a file changed since it was read: error = %v, want %v", err, store.ErrConflict)
			}
			lines := f.linesOf("drive_conflict")
			if len(lines) != 1 || lines[0]["read_revision"] != int64(1) || lines[0]["found_revision"] != tc.found || lines[0]["reason"] != "changed" {
				t.Errorf("drive_conflict lines = %v, want one, changed, read at 1 and found at %d", lines, tc.found)
			}
			if got := f.profileFile(t, miaToken).Content; !bytes.Equal(got, changed) {
				t.Errorf("the file holds\n%s\nwant it as it was changed:\n%s", got, changed)
			}
		})
	}
}

// A file Drive says the service may no longer reach is, to the service, a file
// that is not there: it is forgotten, and the profile searched for.
func TestAFileOutOfReachIsSearchedForAgain(t *testing.T) {
	t.Parallel()

	f := newFixture(t, 5*time.Second)
	f.create(t, mia, child("Mia"))
	f.fake.Fail(drivetest.Download, http.StatusForbidden, "appNotAuthorizedToFile")
	f.fake.ResetCalls()

	f.loaded(t, f.storage, mia)
	if calls := f.fake.Calls(); !maps.Equal(calls, drivetest.Calls{"download": 2, "list": 1}) {
		t.Errorf("Load() of a file out of reach cost %v, want the refused download, a search and a download", calls)
	}
}

// A search that finds the file written to again — an export, say — does not
// let go of the number written: a read right after it is still held to it.
func TestASearchBetweenTheWriteAndTheReadKeepsTheReadHeld(t *testing.T) {
	t.Parallel()

	f := newFixture(t, 5*time.Second)
	written := f.writtenThenBehind(t, 1)
	if _, err := f.storage.Export(t.Context(), mia); err != nil {
		t.Fatalf("Export() error = %v, want nil", err)
	}

	got, _ := f.loaded(t, f.storage, mia)
	if got.Revision != written.Revision {
		t.Errorf("Load() after an export read revision %d, want the one written, %d", got.Revision, written.Revision)
	}
	f.staleReadSaid(t, "caught_up")
}

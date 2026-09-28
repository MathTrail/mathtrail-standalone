package drivestore_test

import (
	"bytes"
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

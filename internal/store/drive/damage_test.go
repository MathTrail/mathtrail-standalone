package drivestore_test

import (
	"bytes"
	"errors"
	"fmt"
	"maps"
	"net/http"
	"slices"
	"testing"
	"time"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/profile"
	"github.com/MathTrail/mathtrail-standalone/internal/infra/drive/drivetest"
	"github.com/MathTrail/mathtrail-standalone/internal/store"
)

// damage is what a file edited by hand and saved half-way holds.
var damage = []byte(`{"schema_version": 1, "student": {"pseud`)

// movedAt is a profile moved on at a moment of the case's choosing.
func movedAt(p *profile.Profile, pseudonym string, at time.Time) *profile.Profile {
	p.Student.Pseudonym = pseudonym
	p.Touch("drivestore_test", at)
	return p
}

// keptForever is every revision of the file kept forever, the earliest first.
func keptForever(file *drivetest.File) []drivetest.Revision {
	return slices.DeleteFunc(slices.Clone(file.Revisions), func(r drivetest.Revision) bool { return !r.KeepForever })
}

// A damaged file is told as damage, and a read leaves it, and its history, as
// they were: putting an earlier state back loses what came after it, which is
// the parent's to agree to. A read that finds damage again finds the same.
func TestADamagedFileIsToldAndLeftAsItIs(t *testing.T) {
	t.Parallel()

	f := newFixture(t, 5*time.Second)
	f.create(t, mia, child("Mia"))
	p, read := f.loaded(t, f.storage, mia)
	f.saved(t, f.storage, mia, moved(p, "Mia the brave"), read)
	plant(t, f.fake, miaToken, damage)
	before := f.profileFile(t, miaToken)
	f.fake.ResetCalls()

	for range 2 {
		if _, _, err := f.storage.Load(t.Context(), mia); !errors.Is(err, store.ErrDamaged) {
			t.Fatalf("Load() of a damaged file error = %v, want %v", err, store.ErrDamaged)
		}
	}
	// Each read: the remembered file, then the search that makes sure it is
	// still the profile — and nothing of the history, and no write.
	want := drivetest.Calls{"download": 4, "list": 2}
	if calls := f.fake.Calls(); !maps.Equal(calls, want) {
		t.Errorf("two reads of a damaged file cost %v, want %v", calls, want)
	}
	after := f.profileFile(t, miaToken)
	if !bytes.Equal(after.Content, damage) || len(after.Revisions) != len(before.Revisions) || len(keptForever(after)) != len(keptForever(before)) {
		t.Errorf("after the reads the file holds %s with %d revisions, %d kept; want the damage, %d revisions, %d kept",
			after.Content, len(after.Revisions), len(keptForever(after)), len(before.Revisions), len(keptForever(before)))
	}
	if lines := f.linesOf("drive_recovered"); len(lines) != 0 {
		t.Errorf("drive_recovered lines = %v, want none: nothing was put back", lines)
	}
}

// A damaged file is put back to its latest earlier state that reads when the
// parent asks: kept forever first, since Drive gives no other, and written
// forward as a new revision. Restore answers the profile put back, at a
// revision the next write moves on from.
func TestADamagedFileIsPutBackWhenAsked(t *testing.T) {
	t.Parallel()

	f := newFixture(t, 5*time.Second)
	f.create(t, mia, child("Mia"))
	p, read := f.loaded(t, f.storage, mia)
	lastGood := fileOf(t, moved(p, "Mia the brave"))
	f.saved(t, f.storage, mia, p, read)
	plant(t, f.fake, miaToken, damage)
	f.fake.ResetCalls()

	restored, revision, err := f.storage.Restore(t.Context(), mia)
	if err != nil {
		t.Fatalf("Restore() of a damaged file error = %v, want nil", err)
	}
	if got := fileOf(t, restored); !bytes.Equal(got, lastGood) {
		t.Errorf("Restore() =\n%s\nwant its last state that reads:\n%s", got, lastGood)
	}
	// The remembered file, then the search that makes sure it is still the
	// profile, the history, the earlier state kept and read, the file once
	// more before it is written, and the write.
	want := drivetest.Calls{"download": 3, "list": 1, "revisions": 1, "keep": 1, "revision": 1, "update": 1}
	if calls := f.fake.Calls(); !maps.Equal(calls, want) {
		t.Errorf("putting the file back cost %v, want %v", calls, want)
	}
	file := f.profileFile(t, miaToken)
	if !bytes.Equal(file.Content, lastGood) {
		t.Errorf("the file holds\n%s\nwant its last state that reads:\n%s", file.Content, lastGood)
	}
	history := file.Revisions
	if damaged := history[len(history)-2]; !bytes.Equal(damaged.Content, damage) {
		t.Errorf("the revision before the one put back holds %s, want the damage, kept in the history", damaged.Content)
	}
	lines := f.linesOf("drive_recovered")
	if len(lines) != 1 || lines[0]["outcome"] != "restored" || lines[0]["tried"] != int64(1) || lines[0]["restored_revision"] != int64(2) {
		t.Errorf("drive_recovered lines = %v, want one, restored to 2 at the first try", lines)
	}

	f.saved(t, f.storage, mia, moved(restored, "Mia once more"), revision)
	got, _ := f.loaded(t, f.storage, mia)
	if got.Student.Pseudonym != "Mia once more" {
		t.Errorf("Load() after a write from the revision Restore() answered = %q, want the write", got.Student.Pseudonym)
	}
}

// putBackThenEarly puts a damaged file back, and has Drive answer the next
// downloads of it with the damage, as many times as given: Drive does not
// promise that a read right after a write sees it. It answers the profile put
// back.
func (f *fixture) putBackThenEarly(t *testing.T, downloads int) *profile.Profile {
	t.Helper()

	f.create(t, mia, child("Mia"))
	plant(t, f.fake, miaToken, damage)
	restored, _, err := f.storage.Restore(t.Context(), mia)
	if err != nil {
		t.Fatalf("Restore() error = %v, want nil", err)
	}
	f.fake.Lag(miaToken, f.profileFile(t, miaToken).ID, downloads)
	f.waits.taken()
	return restored
}

// A file just put back that reads as the damage it was put back from, as it
// remembers it and as the search finds it, is read once more after a moment
// by the instance that put it back, rather than told as damage the parent
// would be asked about again.
func TestAFileJustPutBackIsReadAgainRatherThanCalledDamaged(t *testing.T) {
	t.Parallel()

	f := newFixture(t, 5*time.Second)
	restored := f.putBackThenEarly(t, 2)

	if got, _ := f.loaded(t, f.storage, mia); got.Revision != restored.Revision {
		t.Errorf("Load() = revision %d, want the profile put back, %d", got.Revision, restored.Revision)
	}
	f.staleReadSaid(t, "caught_up")
}

// A file just put back that still reads as the damage after a moment is told
// as an early read — make the call again — once, and the next read takes the
// file as it is.
func TestAFileJustPutBackStillEarlyIsToldAsBehindOnce(t *testing.T) {
	t.Parallel()

	f := newFixture(t, 5*time.Second)
	restored := f.putBackThenEarly(t, 3)

	if _, _, err := f.storage.Load(t.Context(), mia); !errors.Is(err, store.ErrBehind) {
		t.Fatalf("Load() still early: error = %v, want %v", err, store.ErrBehind)
	}
	f.staleReadSaid(t, "behind")
	if got, _ := f.loaded(t, f.storage, mia); got.Revision != restored.Revision {
		t.Errorf("Load() once the reads caught up = revision %d, want %d", got.Revision, restored.Revision)
	}
}

// A file put back that has since been damaged again and set aside by another
// instance, for a new start, is no early read: the search finds the profile
// started in its place, and nothing waits.
func TestAFilePutBackAndSetAsideSinceGivesWayToTheNewProfile(t *testing.T) {
	t.Parallel()

	f := newFixture(t, 5*time.Second)
	f.create(t, mia, child("Mia"))
	plant(t, f.fake, miaToken, damage)
	if _, _, err := f.storage.Restore(t.Context(), mia); err != nil {
		t.Fatalf("Restore() error = %v, want nil", err)
	}
	plant(t, f.fake, miaToken, damage)
	if _, err := f.instance(t, 5*time.Second).StartOver(t.Context(), mia, child("Mia anew")); err != nil {
		t.Fatalf("StartOver() on another instance error = %v, want nil", err)
	}
	f.waits.taken()

	if got, _ := f.loaded(t, f.storage, mia); got.Student.Pseudonym != "Mia anew" {
		t.Errorf("Load() = %q, want the profile started in place of the one set aside", got.Student.Pseudonym)
	}
	if waited := f.waits.taken(); len(waited) != 0 {
		t.Errorf("Load() waited %v, want no wait for a file set aside", waited)
	}
}

// A run of damage longer than the latest revisions a recovery reads leaves
// the state kept forever for its day, and that is what the file is put back
// to.
func TestARunOfDamageIsPutBackFromTheStateKeptForItsDay(t *testing.T) {
	t.Parallel()

	f := newFixture(t, 5*time.Second)
	kept := fileOf(t, child("Mia of yesterday"))
	id := f.fake.Put(miaToken, &drivetest.File{
		Name: fileName, MimeType: fileType, AppProperties: maps.Clone(profileMarker),
		Revisions: []drivetest.Revision{
			{Content: fileOf(t, child("Mia of the day before"))},
			{Content: kept, KeepForever: true},
			{Content: damage}, {Content: damage}, {Content: damage}, {Content: damage}, {Content: damage},
		},
	})

	if _, _, err := f.storage.Restore(t.Context(), mia); err != nil {
		t.Fatalf("Restore() error = %v, want nil", err)
	}
	if got := f.fake.Files(miaToken)[0].Content; !bytes.Equal(got, kept) {
		t.Errorf("file %s holds\n%s\nwant the state kept for its day:\n%s", id, got, kept)
	}
	if lines := f.linesOf("drive_recovered"); len(lines) != 1 || lines[0]["tried"] != int64(5) {
		t.Errorf("drive_recovered lines = %v, want one, after five tries", lines)
	}
}

// A file no state of which reads is reported as one nothing can read, and
// left as it is: the service does not write over a file it cannot read.
func TestAFileNoStateOfWhichReadsIsLeftAsItIs(t *testing.T) {
	t.Parallel()

	f := newFixture(t, 5*time.Second)
	f.fake.Put(miaToken, &drivetest.File{
		Name: fileName, MimeType: fileType, AppProperties: maps.Clone(profileMarker),
		Revisions: []drivetest.Revision{{Content: []byte("{}")}, {Content: damage}},
	})

	_, _, err := f.storage.Restore(t.Context(), mia)
	if !errors.Is(err, store.ErrCorrupted) || errors.Is(err, store.ErrDamaged) {
		t.Fatalf("Restore() error = %v, want %v and not %v: there is nothing left to put back", err, store.ErrCorrupted, store.ErrDamaged)
	}
	if calls := f.fake.Calls(); calls["update"] != 0 {
		t.Errorf("a file nothing reads was written %d times, want never", calls["update"])
	}
	if got := f.profileFile(t, miaToken).Content; !bytes.Equal(got, damage) {
		t.Errorf("the file holds %s, want the damage left as it was", got)
	}
	if lines := f.linesOf("drive_recovered"); len(lines) != 1 || lines[0]["outcome"] != "nothing_readable" {
		t.Errorf("drive_recovered lines = %v, want one, nothing_readable", lines)
	}
}

// A recovery that cannot reach Drive says that, and not that nothing reads:
// nobody should be offered a new start because Drive did not answer.
func TestARecoveryThatCannotReachDriveSaysSo(t *testing.T) {
	t.Parallel()

	f := newFixture(t, 5*time.Second)
	f.create(t, mia, child("Mia"))
	plant(t, f.fake, miaToken, damage)
	failEveryTry(f.fake, drivetest.ListRevisions, http.StatusServiceUnavailable, "backendError")

	_, _, err := f.storage.Restore(t.Context(), mia)
	if !errors.Is(err, store.ErrUnavailable) || errors.Is(err, store.ErrCorrupted) {
		t.Errorf("Restore() error = %v, want %v and not %v", err, store.ErrUnavailable, store.ErrCorrupted)
	}
}

// A file mended while its history was read is not written over with an older
// state: what it holds now counts, and reading again finds it.
func TestAFileMendedDuringItsRecoveryIsNotWrittenOver(t *testing.T) {
	t.Parallel()

	f := newFixture(t, 5*time.Second)
	f.create(t, mia, child("Mia"))
	plant(t, f.fake, miaToken, damage)
	mended := fileOf(t, child("Mia mended by hand"))
	letGo := f.fake.Hold(drivetest.KeepRevision)
	f.fake.ResetCalls()

	done := make(chan error, 1)
	go func() {
		_, _, err := f.storage.Restore(t.Context(), mia)
		done <- err
	}()
	for f.fake.Calls()["keep"] == 0 {
		time.Sleep(time.Millisecond)
	}
	plant(t, f.fake, miaToken, mended)
	letGo()
	if err := <-done; !errors.Is(err, store.ErrConflict) {
		t.Errorf("Restore() of a file mended during its recovery: error = %v, want %v", err, store.ErrConflict)
	}
	if got := f.profileFile(t, miaToken).Content; !bytes.Equal(got, mended) {
		t.Errorf("the file holds\n%s\nwant it as it was mended:\n%s", got, mended)
	}
}

// The first write of a day is kept forever, and no other write of that day
// is; a file that does not say when it was written is kept as one nothing is
// known of.
func TestTheFirstWriteOfADayIsKeptForever(t *testing.T) {
	t.Parallel()

	f := newFixture(t, 5*time.Second)
	f.create(t, mia, child("Mia"))
	p, read := f.loaded(t, f.storage, mia)
	read = f.saved(t, f.storage, mia, moved(p, "Mia on the day"), read)
	if kept := keptForever(f.profileFile(t, miaToken)); len(kept) != 0 {
		t.Fatalf("after a second write of the day %d revisions are kept forever, want none", len(kept))
	}

	nextDay := moment.Add(24 * time.Hour)
	f.fake.ResetCalls()
	p, _ = f.loaded(t, f.storage, mia)
	f.saved(t, f.storage, mia, movedAt(p, "Mia the next day", nextDay), read)
	file := f.profileFile(t, miaToken)
	kept := keptForever(file)
	if len(kept) != 1 || kept[0].ID != file.Revisions[len(file.Revisions)-1].ID {
		t.Errorf("after the first write of the next day the revisions kept forever are %d, want the new one alone", len(kept))
	}
	if got := f.fake.Calls(); got["revisions"] != 0 || got["delete"] != 0 {
		t.Errorf("the first write of a day read the history %d times and deleted %d revisions, want neither: a write takes nothing from the history",
			got["revisions"], got["delete"])
	}
}

// No write deletes a revision kept forever: the day's first write keeps its
// own while Drive has room, and once Drive keeps the most it will, the write
// lands without — every earlier day's state stays, for a recovery to put the
// file back from.
func TestNoWriteDeletesARevisionKeptForever(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name     string
		kept     int
		wantKept int
		keptNow  bool
	}{
		{"with room left", 150, 151, true},
		{"at the most Drive keeps", 200, 200, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			f := newFixture(t, 5*time.Second)
			history := make([]drivetest.Revision, 0, tc.kept+1)
			for range tc.kept {
				history = append(history, drivetest.Revision{Content: []byte("{}"), KeepForever: true})
			}
			history = append(history, drivetest.Revision{Content: fileOf(t, child("Mia"))})
			id := f.fake.Put(miaToken, &drivetest.File{
				Name: fileName, MimeType: fileType, AppProperties: maps.Clone(profileMarker), Revisions: history,
			})
			p, read := f.loaded(t, f.storage, mia)
			written := movedAt(p, "Mia the next day", moment.Add(24*time.Hour))
			want := fileOf(t, written)
			f.fake.ResetCalls()

			f.saved(t, f.storage, mia, written, read)
			after := f.fake.Files(miaToken)[0]
			if !bytes.Equal(after.Content, want) {
				t.Errorf("file %s holds\n%s\nwant the write landed:\n%s", id, after.Content, want)
			}
			kept := keptForever(after)
			if len(kept) != tc.wantKept || after.Revisions[len(after.Revisions)-1].KeepForever != tc.keptNow {
				t.Errorf("file %s keeps %d revisions forever, the one just written among them %t; want %d, %t",
					id, len(kept), after.Revisions[len(after.Revisions)-1].KeepForever, tc.wantKept, tc.keptNow)
			}
			if got := f.fake.Calls(); got["delete"] != 0 || got["revisions"] != 0 {
				t.Errorf("the write deleted %d revisions and read the history %d times, want neither", got["delete"], got["revisions"])
			}
		})
	}
}

// A revision Drive will not keep, or will not give, is passed over for the
// next one tried: what cannot be read is no state to put the file back to.
func TestARevisionDriveWillNotGiveIsPassedOver(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name  string
		spoil func(*drivetest.Drive)
	}{
		{"one Drive will not keep", func(fake *drivetest.Drive) {
			fake.Fail(drivetest.KeepRevision, http.StatusBadRequest, "badRequest")
		}},
		{"one Drive will not give", func(fake *drivetest.Drive) {
			fake.Fail(drivetest.DownloadRevision, http.StatusForbidden, "download_restricted_for_revision")
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			f := newFixture(t, 5*time.Second)
			older := fileOf(t, child("Mia two states ago"))
			f.fake.Put(miaToken, &drivetest.File{
				Name: fileName, MimeType: fileType, AppProperties: maps.Clone(profileMarker),
				Revisions: []drivetest.Revision{{Content: older}, {Content: fileOf(t, child("Mia a state ago"))}, {Content: damage}},
			})
			tc.spoil(f.fake)

			if _, _, err := f.storage.Restore(t.Context(), mia); err != nil {
				t.Fatalf("Restore() error = %v, want nil", err)
			}
			if got := f.profileFile(t, miaToken).Content; !bytes.Equal(got, older) {
				t.Errorf("the file holds\n%s\nwant the state before the one passed over:\n%s", got, older)
			}
		})
	}
}

// oversized is a file too large for any profile: nothing of it is read.
var oversized = bytes.Repeat([]byte(" "), 1<<20+1)

// A file too large for any profile is damage like any other: told as that,
// and put back from its history when asked.
func TestAFileTooLargeIsPutBackFromItsHistory(t *testing.T) {
	t.Parallel()

	f := newFixture(t, 5*time.Second)
	good := child("Mia")
	f.create(t, mia, good)
	plant(t, f.fake, miaToken, oversized)

	if _, _, err := f.storage.Load(t.Context(), mia); !errors.Is(err, store.ErrDamaged) {
		t.Fatalf("Load() of a file too large error = %v, want %v", err, store.ErrDamaged)
	}
	if _, _, err := f.storage.Restore(t.Context(), mia); err != nil {
		t.Fatalf("Restore() of a file too large error = %v, want nil", err)
	}
	if got := f.profileFile(t, miaToken).Content; !bytes.Equal(got, fileOf(t, good)) {
		t.Errorf("the file holds %d bytes, want the profile it held before", len(got))
	}
}

// A recovery that could not look through the history says why, and not that
// nothing reads: a history Drive would not list, or a Drive too full to keep a
// revision in, is no reason to offer a new start.
func TestARecoveryThatCouldNotLookSaysWhy(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name  string
		spoil func(*drivetest.Drive)
		want  error
	}{
		{"a history Drive will not list", func(fake *drivetest.Drive) {
			fake.Fail(drivetest.ListRevisions, http.StatusBadRequest, "badRequest")
		}, nil},
		{"a Drive too full to keep a revision", func(fake *drivetest.Drive) {
			fake.Fail(drivetest.KeepRevision, http.StatusForbidden, "storageQuotaExceeded")
		}, store.ErrStorageFull},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			f := newFixture(t, 5*time.Second)
			f.create(t, mia, child("Mia"))
			plant(t, f.fake, miaToken, damage)
			tc.spoil(f.fake)

			_, _, err := f.storage.Restore(t.Context(), mia)
			if err == nil || errors.Is(err, store.ErrCorrupted) || (tc.want != nil && !errors.Is(err, tc.want)) {
				t.Errorf("Restore() error = %v, want %v and never %v", err, tc.want, store.ErrCorrupted)
			}
		})
	}
}

// An upload asked to keep its revision that Drive refuses is made again
// without, and lands unkept, whatever the refusal — Drive does not say how it
// refuses a revision past the most it keeps — but for one an upload without
// keeping would meet as well, such as a full Drive.
func TestAKeptUploadDriveRefusesIsMadeWithoutKeeping(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name   string
		status int
		reason string
		lands  bool
	}{
		{"a refusal the store has no word for", http.StatusBadRequest, "badRequest", true},
		{"a refusal for a reason of its own", http.StatusForbidden, "keepForeverLimitReached", true},
		{"a refusal the store names for something else", http.StatusForbidden, "downloadRestrictedForRevision", true},
		{"a full Drive", http.StatusForbidden, "storageQuotaExceeded", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			f := newFixture(t, 5*time.Second)
			f.create(t, mia, child("Mia"))
			p, read := f.loaded(t, f.storage, mia)
			written := movedAt(p, "Mia the next day", moment.Add(24*time.Hour))
			want := fileOf(t, written)
			f.fake.Fail(drivetest.Update, tc.status, tc.reason)
			f.fake.ResetCalls()

			_, err := f.storage.Save(t.Context(), mia, written, read)
			file := f.profileFile(t, miaToken)
			if landed := bytes.Equal(file.Content, want); landed != tc.lands || (err == nil) != tc.lands {
				t.Errorf("Save() error = %v, and the write landed: %v; want it landed: %v", err, landed, tc.lands)
			}
			if kept := len(keptForever(file)); kept != 0 {
				t.Errorf("%d revisions are kept forever, want none: Drive refused the one asked for", kept)
			}
			if got, want := f.fake.Calls()["update"], map[bool]int{true: 2, false: 1}[tc.lands]; got != want {
				t.Errorf("the upload was sent %d times, want %d", got, want)
			}
		})
	}
}

// A profile that breaks a rule of this build is told as damage and left as it
// is, however readable its history: a newer build may have written it, or the
// parent edited it by hand, and putting an earlier state over either would
// lose it.
func TestAProfileThatBreaksARuleIsNotRolledBack(t *testing.T) {
	t.Parallel()

	f := newFixture(t, 5*time.Second)
	f.create(t, mia, child("Mia"))
	broken := fmt.Appendf(nil, `{"schema_version": %d}`, profile.Version)
	plant(t, f.fake, miaToken, broken)
	f.fake.ResetCalls()

	if _, _, err := f.storage.Load(t.Context(), mia); !errors.Is(err, store.ErrCorrupted) || errors.Is(err, store.ErrDamaged) {
		t.Fatalf("Load() of a profile that breaks a rule: error = %v, want %v and not %v", err, store.ErrCorrupted, store.ErrDamaged)
	}
	if _, _, err := f.storage.Restore(t.Context(), mia); !errors.Is(err, store.ErrCorrupted) || errors.Is(err, store.ErrDamaged) {
		t.Fatalf("Restore() of a profile that breaks a rule: error = %v, want %v and not %v", err, store.ErrCorrupted, store.ErrDamaged)
	}
	if calls := f.fake.Calls(); calls["revisions"] != 0 || calls["update"] != 0 {
		t.Errorf("Load() and Restore() of a profile that breaks a rule cost %v, want its history left alone", calls)
	}
	if got := f.profileFile(t, miaToken).Content; !bytes.Equal(got, broken) {
		t.Errorf("the file holds %s, want it left as it was", got)
	}
}

// At the most revisions Drive keeps forever, a recovery makes room before it
// keeps the ones it reads, and puts the file back.
func TestARecoveryAtTheMostRevisionsKeptMakesRoom(t *testing.T) {
	t.Parallel()

	f := newFixture(t, 5*time.Second)
	good := fileOf(t, child("Mia"))
	history := make([]drivetest.Revision, 0, 202)
	for range 200 {
		history = append(history, drivetest.Revision{Content: []byte("{}"), KeepForever: true})
	}
	history = append(history, drivetest.Revision{Content: good}, drivetest.Revision{Content: damage})
	f.fake.Put(miaToken, &drivetest.File{Name: fileName, MimeType: fileType, AppProperties: maps.Clone(profileMarker), Revisions: history})

	if _, _, err := f.storage.Restore(t.Context(), mia); err != nil {
		t.Fatalf("Restore() error = %v, want nil", err)
	}
	file := f.fake.Files(miaToken)[0]
	if !bytes.Equal(file.Content, good) {
		t.Errorf("the file holds %s, want the state before the damage", file.Content)
	}
	if kept := keptForever(file); len(kept) > 200 {
		t.Errorf("%d revisions are kept forever, want at most two hundred", len(kept))
	}
	// Of the revisions it reads, only the state before the damage was not
	// kept: one deleted makes its room, and none more is taken.
	if got := f.fake.Calls()["delete"]; got != 1 {
		t.Errorf("revisions deleted to make room = %d, want the one the revision kept to read needs", got)
	}
}

// A revision Drive refuses to keep, for a reason the store has no word for,
// leaves a recovery unsure: it is not told as a file nothing can read.
func TestARecoveryThatWasRefusedIsNotToldAsNothingReadable(t *testing.T) {
	t.Parallel()

	f := newFixture(t, 5*time.Second)
	f.create(t, mia, child("Mia"))
	plant(t, f.fake, miaToken, damage)
	f.fake.Fail(drivetest.KeepRevision, http.StatusBadRequest, "badRequest")

	_, revision, err := f.storage.Restore(t.Context(), mia)
	if err == nil || errors.Is(err, store.ErrCorrupted) || revision != "" {
		t.Errorf("Restore() = %q, %v, want the refusal and not %v", revision, err, store.ErrCorrupted)
	}
}

// A revision kept forever that is already gone when a recovery deletes it to
// make room is as good as deleted: making room goes on to the next one, and
// the file is put back.
func TestARevisionAlreadyGoneIsAsGoodAsDeleted(t *testing.T) {
	t.Parallel()

	f := newFixture(t, 5*time.Second)
	good := fileOf(t, child("Mia"))
	history := make([]drivetest.Revision, 0, 203)
	for range 200 {
		history = append(history, drivetest.Revision{Content: []byte("{}"), KeepForever: true})
	}
	// Two states not kept yet, which a recovery reads: room for both.
	history = append(history, drivetest.Revision{Content: good}, drivetest.Revision{Content: good},
		drivetest.Revision{Content: damage})
	f.fake.Put(miaToken, &drivetest.File{Name: fileName, MimeType: fileType, AppProperties: maps.Clone(profileMarker), Revisions: history})
	f.fake.Fail(drivetest.DeleteRevision, http.StatusNotFound, "notFound")
	f.fake.ResetCalls()

	if _, _, err := f.storage.Restore(t.Context(), mia); err != nil {
		t.Fatalf("Restore() error = %v, want nil", err)
	}
	if got := f.fake.Calls()["delete"]; got != 2 {
		t.Errorf("revisions deleted to make room = %d, want both, the one already gone among them", got)
	}
	if got := f.fake.Files(miaToken)[0].Content; !bytes.Equal(got, good) {
		t.Errorf("the file holds %s, want the state before the damage", got)
	}
}

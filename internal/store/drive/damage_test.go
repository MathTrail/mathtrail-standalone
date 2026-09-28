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

// A damaged file is put back to its latest earlier state that reads: kept
// forever first, since Drive gives no other, and written forward as a new
// revision. The read that did it answers that it did, and nothing else; the
// next read finds the profile as it was.
func TestADamagedFileIsPutBack(t *testing.T) {
	t.Parallel()

	f := newFixture(t, 5*time.Second)
	f.create(t, mia, child("Mia"))
	p, read := f.loaded(t, f.storage, mia)
	lastGood := fileOf(t, moved(p, "Mia the brave"))
	f.saved(t, f.storage, mia, p, read)
	plant(t, f.fake, miaToken, damage)
	f.fake.ResetCalls()

	if _, _, err := f.storage.Load(t.Context(), mia); !errors.Is(err, store.ErrRestored) {
		t.Fatalf("Load() of a damaged file error = %v, want %v", err, store.ErrRestored)
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

	got, _ := f.loaded(t, f.storage, mia)
	if file := fileOf(t, got); !bytes.Equal(file, lastGood) {
		t.Errorf("Load() after the file was put back =\n%s\nwant\n%s", file, lastGood)
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

	if _, _, err := f.storage.Load(t.Context(), mia); !errors.Is(err, store.ErrRestored) {
		t.Fatalf("Load() error = %v, want %v", err, store.ErrRestored)
	}
	if got := f.fake.Files(miaToken)[0].Content; !bytes.Equal(got, kept) {
		t.Errorf("file %s holds\n%s\nwant the state kept for its day:\n%s", id, got, kept)
	}
	if lines := f.linesOf("drive_recovered"); len(lines) != 1 || lines[0]["tried"] != int64(5) {
		t.Errorf("drive_recovered lines = %v, want one, after five tries", lines)
	}
}

// A file no state of which reads is reported as damage, and left as it is: the
// service does not write over a file it cannot read.
func TestAFileNoStateOfWhichReadsIsLeftAsItIs(t *testing.T) {
	t.Parallel()

	f := newFixture(t, 5*time.Second)
	f.fake.Put(miaToken, &drivetest.File{
		Name: fileName, MimeType: fileType, AppProperties: maps.Clone(profileMarker),
		Revisions: []drivetest.Revision{{Content: []byte("{}")}, {Content: damage}},
	})

	if _, _, err := f.storage.Load(t.Context(), mia); !errors.Is(err, store.ErrCorrupted) {
		t.Fatalf("Load() error = %v, want %v", err, store.ErrCorrupted)
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

	_, _, err := f.storage.Load(t.Context(), mia)
	if !errors.Is(err, store.ErrUnavailable) || errors.Is(err, store.ErrCorrupted) {
		t.Errorf("Load() error = %v, want %v and not %v", err, store.ErrUnavailable, store.ErrCorrupted)
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
		_, _, err := f.storage.Load(t.Context(), mia)
		done <- err
	}()
	for f.fake.Calls()["keep"] == 0 {
		time.Sleep(time.Millisecond)
	}
	plant(t, f.fake, miaToken, mended)
	letGo()
	if err := <-done; !errors.Is(err, store.ErrConflict) {
		t.Errorf("Load() of a file mended during its recovery: error = %v, want %v", err, store.ErrConflict)
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
	if got := f.fake.Calls(); got["revisions"] != 1 {
		t.Errorf("the first write of a day read the history %d times, want once, to count what is kept", got["revisions"])
	}
}

// Past a hundred revisions kept forever, the earliest are deleted: Drive keeps
// no more than two hundred so, and lets go of one only when it is deleted.
// Deleting what it cannot does not undo the write it follows.
func TestRevisionsKeptPastAHundredAreDeletedEarliestFirst(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name      string
		spoil     func(*drivetest.Drive)
		wantKept  int
		wantFirst int
	}{
		{"deletions Drive takes", func(*drivetest.Drive) {}, 100, 2},
		{"deletions Drive fails", func(fake *drivetest.Drive) {
			failEveryTry(fake, drivetest.DeleteRevision, http.StatusInternalServerError, "backendError")
		}, 102, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			f := newFixture(t, 5*time.Second)
			history := make([]drivetest.Revision, 0, 102)
			for range 101 {
				history = append(history, drivetest.Revision{Content: []byte("{}"), KeepForever: true})
			}
			history = append(history, drivetest.Revision{Content: fileOf(t, child("Mia"))})
			id := f.fake.Put(miaToken, &drivetest.File{
				Name: fileName, MimeType: fileType, AppProperties: maps.Clone(profileMarker), Revisions: history,
			})
			before := f.fake.Files(miaToken)[0].Revisions
			p, read := f.loaded(t, f.storage, mia)
			tc.spoil(f.fake)

			f.saved(t, f.storage, mia, movedAt(p, "Mia the next day", moment.Add(24*time.Hour)), read)
			after := f.fake.Files(miaToken)[0]
			kept := keptForever(after)
			if len(kept) != tc.wantKept {
				t.Fatalf("file %s keeps %d revisions forever, want %d", id, len(kept), tc.wantKept)
			}
			if kept[0].ID != before[tc.wantFirst].ID {
				t.Errorf("the earliest revision kept is not the one after the %d deleted", tc.wantFirst)
			}
			if kept[len(kept)-1].ID != after.Revisions[len(after.Revisions)-1].ID {
				t.Error("the revision just written is not among those kept")
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

			if _, _, err := f.storage.Load(t.Context(), mia); !errors.Is(err, store.ErrRestored) {
				t.Fatalf("Load() error = %v, want %v", err, store.ErrRestored)
			}
			if got := f.profileFile(t, miaToken).Content; !bytes.Equal(got, older) {
				t.Errorf("the file holds\n%s\nwant the state before the one passed over:\n%s", got, older)
			}
		})
	}
}

// oversized is a file too large for any profile: nothing of it is read.
var oversized = bytes.Repeat([]byte(" "), 1<<20+1)

// A file too large for any profile is damage like any other, and is put back
// from its history.
func TestAFileTooLargeIsPutBackFromItsHistory(t *testing.T) {
	t.Parallel()

	f := newFixture(t, 5*time.Second)
	good := child("Mia")
	f.create(t, mia, good)
	plant(t, f.fake, miaToken, oversized)

	if _, _, err := f.storage.Load(t.Context(), mia); !errors.Is(err, store.ErrRestored) {
		t.Fatalf("Load() of a file too large error = %v, want %v", err, store.ErrRestored)
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

			_, _, err := f.storage.Load(t.Context(), mia)
			if err == nil || errors.Is(err, store.ErrCorrupted) || (tc.want != nil && !errors.Is(err, tc.want)) {
				t.Errorf("Load() error = %v, want %v and never %v", err, tc.want, store.ErrCorrupted)
			}
		})
	}
}

// At the most revisions Drive keeps forever, the day's first write makes room
// before it keeps its own: the upload is not refused for want of it.
func TestTheDaysFirstWriteMakesRoomBeforeItKeepsItsRevision(t *testing.T) {
	t.Parallel()

	f := newFixture(t, 5*time.Second)
	history := make([]drivetest.Revision, 0, 201)
	for range 200 {
		history = append(history, drivetest.Revision{Content: []byte("{}"), KeepForever: true})
	}
	history = append(history, drivetest.Revision{Content: fileOf(t, child("Mia"))})
	f.fake.Put(miaToken, &drivetest.File{Name: fileName, MimeType: fileType, AppProperties: maps.Clone(profileMarker), Revisions: history})
	p, read := f.loaded(t, f.storage, mia)

	f.saved(t, f.storage, mia, movedAt(p, "Mia the next day", moment.Add(24*time.Hour)), read)
	file := f.fake.Files(miaToken)[0]
	if kept := keptForever(file); len(kept) != 100 || kept[len(kept)-1].ID != file.Revisions[len(file.Revisions)-1].ID {
		t.Errorf("%d revisions are kept forever, want a hundred, the one just written among them", len(kept))
	}
}

// An upload asked to keep its revision that Drive refuses for a reason the
// store has no word for is made again without: the write lands, unkept.
func TestAKeptUploadDriveRefusesIsMadeWithoutKeeping(t *testing.T) {
	t.Parallel()

	f := newFixture(t, 5*time.Second)
	f.create(t, mia, child("Mia"))
	p, read := f.loaded(t, f.storage, mia)
	written := movedAt(p, "Mia the next day", moment.Add(24*time.Hour))
	want := fileOf(t, written)
	f.fake.Fail(drivetest.Update, http.StatusBadRequest, "badRequest")
	f.fake.ResetCalls()

	f.saved(t, f.storage, mia, written, read)
	file := f.profileFile(t, miaToken)
	if !bytes.Equal(file.Content, want) || len(keptForever(file)) != 0 {
		t.Errorf("the file holds %s, with %d revisions kept forever, want the write landed and nothing kept", file.Content, len(keptForever(file)))
	}
	if got := f.fake.Calls()["update"]; got != 2 {
		t.Errorf("the upload was sent %d times, want twice: once kept, and once without", got)
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

	if _, _, err := f.storage.Load(t.Context(), mia); !errors.Is(err, store.ErrCorrupted) {
		t.Fatalf("Load() of a profile that breaks a rule: error = %v, want %v", err, store.ErrCorrupted)
	}
	if calls := f.fake.Calls(); calls["revisions"] != 0 || calls["update"] != 0 {
		t.Errorf("Load() of a profile that breaks a rule cost %v, want its history left alone", calls)
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

	if _, _, err := f.storage.Load(t.Context(), mia); !errors.Is(err, store.ErrRestored) {
		t.Fatalf("Load() error = %v, want %v", err, store.ErrRestored)
	}
	if got := f.fake.Files(miaToken)[0].Content; !bytes.Equal(got, good) {
		t.Errorf("the file holds %s, want the state before the damage", got)
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

	_, _, err := f.storage.Load(t.Context(), mia)
	if err == nil || errors.Is(err, store.ErrCorrupted) || errors.Is(err, store.ErrRestored) {
		t.Errorf("Load() error = %v, want the refusal, and neither %v nor %v", err, store.ErrCorrupted, store.ErrRestored)
	}
}

// A revision kept forever that is already gone when it is deleted is as good
// as deleted: making room goes on to the next one.
func TestARevisionAlreadyGoneIsAsGoodAsDeleted(t *testing.T) {
	t.Parallel()

	f := newFixture(t, 5*time.Second)
	history := make([]drivetest.Revision, 0, 102)
	for range 101 {
		history = append(history, drivetest.Revision{Content: []byte("{}"), KeepForever: true})
	}
	history = append(history, drivetest.Revision{Content: fileOf(t, child("Mia"))})
	f.fake.Put(miaToken, &drivetest.File{Name: fileName, MimeType: fileType, AppProperties: maps.Clone(profileMarker), Revisions: history})
	p, read := f.loaded(t, f.storage, mia)
	f.fake.Fail(drivetest.DeleteRevision, http.StatusNotFound, "notFound")
	f.fake.ResetCalls()

	f.saved(t, f.storage, mia, movedAt(p, "Mia the next day", moment.Add(24*time.Hour)), read)
	if got := f.fake.Calls()["delete"]; got != 2 {
		t.Errorf("revisions deleted to make room = %d, want both, the one already gone among them", got)
	}
}

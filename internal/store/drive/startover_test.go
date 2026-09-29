package drivestore_test

import (
	"bytes"
	"errors"
	"maps"
	"net/http"
	"reflect"
	"slices"
	"testing"
	"time"

	"github.com/MathTrail/mathtrail-standalone/internal/infra/drive/drivetest"
	"github.com/MathTrail/mathtrail-standalone/internal/store"
	"github.com/MathTrail/mathtrail-standalone/internal/store/storetest"
)

// setAsideFile is the one file of the account's Drive that was set aside, or
// the end of the case.
func (f *fixture) setAsideFile(t *testing.T, token string) *drivetest.File {
	t.Helper()

	var found []*drivetest.File
	for _, file := range f.fake.Files(token) {
		if file.AppProperties["mathtrail"] == "set-aside" {
			found = append(found, file)
		}
	}
	if len(found) != 1 {
		t.Fatalf("the Drive holds %d files set aside, want one", len(found))
	}
	return found[0]
}

// A profile nothing reads is started over: the damaged file is renamed by the
// day and marked as set aside, holding what it held, and a new profile is made
// beside it in the folder.
func TestAStartOverSetsTheDamagedFileAside(t *testing.T) {
	t.Parallel()

	f := newFixture(t, 5*time.Second)
	plant(t, f.fake, miaToken, damage)
	fresh := child("Mia again")
	want := fileOf(t, fresh)

	if _, err := f.storage.StartOver(t.Context(), mia, fresh); err != nil {
		t.Fatalf("StartOver() error = %v, want nil", err)
	}
	aside := f.setAsideFile(t, miaToken)
	switch {
	case aside.Name != "mathtrail-profile set aside 2026-09-28.json":
		t.Errorf("the file set aside is called %q, want it named by the day", aside.Name)
	case !bytes.Equal(aside.Content, damage):
		t.Errorf("the file set aside holds %s, want what it held", aside.Content)
	case aside.Trashed:
		t.Error("the file set aside is in the bin, want it where it was")
	}
	if got := f.profileFile(t, miaToken).Content; !bytes.Equal(got, want) {
		t.Errorf("the profile holds\n%s\nwant the new one:\n%s", got, want)
	}
	if lines := f.linesOf("drive_started_over"); len(lines) != 1 || lines[0]["set_aside"] != int64(1) {
		t.Errorf("drive_started_over lines = %v, want one, one file set aside", lines)
	}
}

// A profile in the bin is set aside where it lies, so that restoring it later
// makes no second profile beside the new one.
func TestAStartOverSetsAsideTheProfileInTheBin(t *testing.T) {
	t.Parallel()

	f := newFixture(t, 5*time.Second)
	old := child("Mia in the bin")
	f.create(t, mia, old)
	binned := f.profileFile(t, miaToken).ID
	f.fake.Edit(miaToken, binned, func(file *drivetest.File) { file.Trashed = true })
	if _, _, err := f.instance(t, 5*time.Second).Load(t.Context(), mia); !errors.Is(err, store.ErrInBin) {
		t.Fatalf("Load() of a profile in the bin error = %v, want %v", err, store.ErrInBin)
	}

	fresh := child("Mia anew")
	if _, err := f.storage.StartOver(t.Context(), mia, fresh); err != nil {
		t.Fatalf("StartOver() error = %v, want nil", err)
	}
	if aside := f.setAsideFile(t, miaToken); aside.ID != binned || !aside.Trashed {
		t.Errorf("the file set aside is %s, in the bin: %t; want the one in the bin, left there", aside.ID, aside.Trashed)
	}
	f.fake.Edit(miaToken, binned, func(file *drivetest.File) { file.Trashed = false })
	got, _ := f.loaded(t, f.instance(t, 5*time.Second), mia)
	if file := fileOf(t, got); !bytes.Equal(file, fileOf(t, fresh)) {
		t.Errorf("Load() once the old file is restored =\n%s\nwant the new profile", file)
	}
}

// An instance that still remembers the file set aside reads it, finds it
// damaged, and searches before it mends anything: it reads the new profile,
// and never the old file as the profile.
func TestAnInstanceThatRemembersTheFileSetAsideReadsTheNewProfile(t *testing.T) {
	t.Parallel()

	f := newFixture(t, 5*time.Second)
	f.create(t, mia, child("Mia"))
	other := f.instance(t, 5*time.Second)
	f.loaded(t, other, mia)
	plant(t, f.fake, miaToken, damage)
	f.fake.Purge(miaToken, f.profileFile(t, miaToken).ID)
	if _, _, err := f.storage.Load(t.Context(), mia); !errors.Is(err, store.ErrCorrupted) {
		t.Fatalf("Load() of a damaged file with nothing to put it back from: error = %v, want %v", err, store.ErrCorrupted)
	}
	fresh := child("Mia anew")
	if _, err := f.storage.StartOver(t.Context(), mia, fresh); err != nil {
		t.Fatalf("StartOver() error = %v, want nil", err)
	}

	got, _ := f.loaded(t, other, mia)
	if file := fileOf(t, got); !bytes.Equal(file, fileOf(t, fresh)) {
		t.Errorf("Load() by the instance that remembered the old file =\n%s\nwant the new profile", file)
	}
	if aside := f.setAsideFile(t, miaToken); !bytes.Equal(aside.Content, damage) {
		t.Errorf("the file set aside holds %s, want the damage it held", aside.Content)
	}
}

// An instance that remembers a newer build's file another has set aside
// searches again, as it does for a damaged one, and reads the new profile
// rather than telling the set-aside file to wait.
func TestAnInstanceThatRemembersANewerFileSetAsideReadsTheNewProfile(t *testing.T) {
	t.Parallel()

	f := newFixture(t, 5*time.Second)
	f.create(t, mia, child("Mia"))
	other := f.instance(t, 5*time.Second)
	f.loaded(t, other, mia)
	plant(t, f.fake, miaToken, storetest.NewerFile(moment))
	fresh := child("Mia anew")
	if _, err := f.storage.StartOver(t.Context(), mia, fresh); err != nil {
		t.Fatalf("StartOver() over a newer build's file error = %v, want nil", err)
	}

	got, _ := f.loaded(t, other, mia)
	if file := fileOf(t, got); !bytes.Equal(file, fileOf(t, fresh)) {
		t.Errorf("Load() by the instance that remembered the newer file =\n%s\nwant the new profile", file)
	}
}

// A profile that reads is never started over, and nothing is set aside: what
// the parent was told no longer holds.
func TestAReadableProfileIsNeverStartedOver(t *testing.T) {
	t.Parallel()

	f := newFixture(t, 5*time.Second)
	p := child("Mia")
	want := fileOf(t, p)
	f.create(t, mia, p)

	if _, err := f.storage.StartOver(t.Context(), mia, child("Mia anew")); !errors.Is(err, store.ErrConflict) {
		t.Errorf("StartOver() over a readable profile error = %v, want %v", err, store.ErrConflict)
	}
	if got := f.profileFile(t, miaToken).Content; !bytes.Equal(got, want) {
		t.Errorf("the profile holds\n%s\nwant it as it was", got)
	}
	if aside := setAside(f.fake, miaToken); len(aside) != 0 {
		t.Errorf("%d files were set aside, want none", len(aside))
	}
}

// A search that finds no profile looks in the bin, where a profile the parent
// put there is theirs to restore — and not a profile missing. An instance that
// remembers the file goes on with it for as long as it trusts what it
// remembers, and the first write that lands in the bin tells it otherwise.
func TestAProfileInTheBinIsNotAProfileMissing(t *testing.T) {
	t.Parallel()

	f := newFixture(t, 5*time.Second)
	f.create(t, mia, child("Mia"))
	p, read := f.loaded(t, f.storage, mia)
	f.fake.Edit(miaToken, f.profileFile(t, miaToken).ID, func(file *drivetest.File) { file.Trashed = true })

	cold := f.instance(t, 5*time.Second)
	f.fake.ResetCalls()
	if _, _, err := cold.Load(t.Context(), mia); !errors.Is(err, store.ErrInBin) {
		t.Errorf("Load() by an instance that searches: error = %v, want %v", err, store.ErrInBin)
	}
	if calls := f.fake.Calls(); !maps.Equal(calls, drivetest.Calls{"list": 2}) {
		t.Errorf("Load() of a profile in the bin cost %v, want the search and the search of the bin", calls)
	}
	if _, err := cold.Export(t.Context(), mia); !errors.Is(err, store.ErrInBin) {
		t.Errorf("Export() of a profile in the bin error = %v, want %v", err, store.ErrInBin)
	}

	if _, err := f.storage.Save(t.Context(), mia, moved(p, "Mia in the bin"), read); !errors.Is(err, store.ErrConflict) {
		t.Errorf("Save() into a file in the bin: error = %v, want %v", err, store.ErrConflict)
	}
	if lines := f.linesOf("drive_conflict"); len(lines) != 1 || lines[0]["reason"] != "in_bin" {
		t.Errorf("drive_conflict lines = %v, want one, in_bin", lines)
	}
	if _, _, err := f.storage.Load(t.Context(), mia); !errors.Is(err, store.ErrInBin) {
		t.Errorf("Load() after the write into the bin: error = %v, want %v", err, store.ErrInBin)
	}
}

// Two files that carry a profile are both named where the profile is
// exported: the one read and written, and the other for the parent to look at
// and delete.
func TestAnExportNamesTheOtherProfileFiles(t *testing.T) {
	t.Parallel()

	f := newFixture(t, 5*time.Second)
	f.create(t, mia, child("Mia"))
	older := f.fake.Put(miaToken, &drivetest.File{
		Name: "mathtrail-profile (1).json", MimeType: fileType, AppProperties: maps.Clone(profileMarker),
		Content: fileOf(t, child("Mia before")), ModifiedTime: moment.Add(-24 * time.Hour),
	})

	location, err := f.storage.Export(t.Context(), mia)
	if err != nil {
		t.Fatalf("Export() error = %v, want nil", err)
	}
	want := []store.Elsewhere{{File: "mathtrail-profile (1).json", Link: drivetest.WebViewLink(older)}}
	if location.File != fileName || !reflect.DeepEqual(location.Others, want) {
		t.Errorf("Export() = %+v, want %s and the other file %+v", location, fileName, want)
	}
	if slices.ContainsFunc(location.Others, func(other store.Elsewhere) bool { return other.File == fileName }) {
		t.Error("Export() names the profile among the other files")
	}
}

// A new start that fails before its new profile is made leaves the account as
// it found it: the damaged file still the profile, told as damage, and never
// as no profile at all.
func TestANewStartThatFailsFirstLeavesTheAccountAsItWas(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name  string
		spoil func(*drivetest.Drive)
		want  error
	}{
		{"the profile gone meanwhile", func(fake *drivetest.Drive) {
			fake.Fail(drivetest.Download, http.StatusNotFound, "notFound")
		}, store.ErrConflict},
		{"Drive not answering the read", func(fake *drivetest.Drive) {
			failEveryTry(fake, drivetest.Download, http.StatusServiceUnavailable, "backendError")
		}, store.ErrUnavailable},
		{"Drive failing the new file", func(fake *drivetest.Drive) {
			fake.Fail(drivetest.Create, http.StatusServiceUnavailable, "backendError")
		}, store.ErrUnavailable},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			f := newFixture(t, 5*time.Second)
			plant(t, f.fake, miaToken, damage)
			tc.spoil(f.fake)

			if _, err := f.storage.StartOver(t.Context(), mia, child("Mia anew")); !errors.Is(err, tc.want) {
				t.Errorf("StartOver() error = %v, want %v", err, tc.want)
			}
			if got := f.profileFile(t, miaToken).Content; !bytes.Equal(got, damage) {
				t.Errorf("the profile holds\n%s\nwant the damaged file, left as it was", got)
			}
			if aside := setAside(f.fake, miaToken); len(aside) != 0 {
				t.Errorf("%d files were set aside, want none", len(aside))
			}
		})
	}
}

// A new start that fails after its new profile is made leaves the new profile
// beside the old file, which is read no more: the account is never left with
// no profile at all.
func TestANewStartThatFailsLastLeavesTheNewProfile(t *testing.T) {
	t.Parallel()

	f := newFixture(t, 5*time.Second)
	plant(t, f.fake, miaToken, damage)
	f.fake.Fail(drivetest.Update, http.StatusServiceUnavailable, "backendError")
	fresh := child("Mia anew")

	if _, err := f.storage.StartOver(t.Context(), mia, fresh); !errors.Is(err, store.ErrUnavailable) {
		t.Errorf("StartOver() error = %v, want %v", err, store.ErrUnavailable)
	}
	got, _ := f.loaded(t, f.instance(t, 5*time.Second), mia)
	if file := fileOf(t, got); !bytes.Equal(file, fileOf(t, fresh)) {
		t.Errorf("Load() after a new start that failed last =\n%s\nwant the new profile", file)
	}
}

// A file too large for any profile is one nothing can read: it is set aside
// like any other, and the new profile starts.
func TestAFileTooLargeIsStartedOver(t *testing.T) {
	t.Parallel()

	f := newFixture(t, 5*time.Second)
	plant(t, f.fake, miaToken, oversized)
	if _, _, err := f.storage.Load(t.Context(), mia); !errors.Is(err, store.ErrCorrupted) {
		t.Fatalf("Load() of a file too large with no history: error = %v, want %v", err, store.ErrCorrupted)
	}

	if _, err := f.storage.StartOver(t.Context(), mia, child("Mia anew")); err != nil {
		t.Fatalf("StartOver() in place of a file too large: error = %v, want nil", err)
	}
	if aside := f.setAsideFile(t, miaToken); len(aside.Content) != len(oversized) {
		t.Errorf("the file set aside holds %d bytes, want the %d it held", len(aside.Content), len(oversized))
	}
}

// A file beside the damaged profile that some build reads — this one, or a
// newer — is a profile too: a new start sets the damage aside and leaves it as
// it is, for the parent to find.
func TestANewStartLeavesAReadableFileBesideTheDamage(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name  string
		older []byte
	}{
		{"this build's", fileOf(t, child("Mia restored from the bin"))},
		{"a newer build's", storetest.NewerFile(moment)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			newStartBeside(t, tc.older)
		})
	}
}

// newStartBeside starts over from a damaged profile with an older file that
// reads beside it, and holds the older file to being left a profile.
func newStartBeside(t *testing.T, content []byte) {
	t.Helper()

	f := newFixture(t, 5*time.Second)
	older := f.fake.Put(miaToken, &drivetest.File{
		Name: "mathtrail-profile (1).json", MimeType: fileType, AppProperties: maps.Clone(profileMarker),
		Content: content, ModifiedTime: moment.Add(-24 * time.Hour),
	})
	f.fake.Put(miaToken, &drivetest.File{
		Name: fileName, MimeType: fileType, AppProperties: maps.Clone(profileMarker), Content: damage,
	})

	fresh := child("Mia anew")
	if _, err := f.storage.StartOver(t.Context(), mia, fresh); err != nil {
		t.Fatalf("StartOver() error = %v, want nil", err)
	}
	if aside := setAside(f.fake, miaToken); len(aside) != 1 || !bytes.Equal(aside[0], damage) {
		t.Errorf("the files set aside hold %q, want the damaged one alone", aside)
	}
	for _, file := range f.fake.Files(miaToken) {
		if file.ID == older && file.AppProperties["mathtrail"] != "profile" {
			t.Errorf("the older file that reads is marked %q, want it left a profile", file.AppProperties["mathtrail"])
		}
	}
	got, _ := f.loaded(t, f.instance(t, 5*time.Second), mia)
	if file := fileOf(t, got); !bytes.Equal(file, fileOf(t, fresh)) {
		t.Errorf("Load() after the new start =\n%s\nwant the new profile", file)
	}
}

// A first profile is not made beside one in the bin: that one is the parent's
// to restore, and a new one would be a second profile once they did.
func TestAFirstProfileIsNotMadeBesideOneInTheBin(t *testing.T) {
	t.Parallel()

	f := newFixture(t, 5*time.Second)
	f.create(t, mia, child("Mia"))
	f.fake.Edit(miaToken, f.profileFile(t, miaToken).ID, func(file *drivetest.File) { file.Trashed = true })

	if _, err := f.instance(t, 5*time.Second).Create(t.Context(), mia, child("A second Mia")); !errors.Is(err, store.ErrInBin) {
		t.Errorf("Create() beside a profile in the bin: error = %v, want %v", err, store.ErrInBin)
	}
	if files := f.fake.Files(miaToken); len(files) != 2 {
		t.Errorf("the Drive holds %d files, want the folder and the profile in the bin alone", len(files))
	}
}

package drivestore_test

import (
	"bytes"
	"cmp"
	"context"
	"errors"
	"maps"
	"net/http"
	"reflect"
	"slices"
	"strconv"
	"sync"
	"testing"
	"time"

	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/profile"
	"github.com/MathTrail/mathtrail-standalone/internal/infra/drive"
	"github.com/MathTrail/mathtrail-standalone/internal/infra/drive/drivetest"
	"github.com/MathTrail/mathtrail-standalone/internal/store"
	drivestore "github.com/MathTrail/mathtrail-standalone/internal/store/drive"
	"github.com/MathTrail/mathtrail-standalone/internal/store/storetest"
)

// The layout the store keeps in a parent's Drive. It is written out here
// rather than taken from the package: the folders and files already in
// parents' Drives carry it, so a change to it has to fail a test.
const (
	folderName = "MathTrail"
	fileName   = "mathtrail-profile.json"
	fileType   = "application/json"
)

var (
	profileMarker = map[string]string{"mathtrail": "profile"}
	folderMarker  = map[string]string{"mathtrail": "folder"}
	// schema is the copy of the file's version a write leaves beside the
	// marker.
	schema = strconv.Itoa(profile.Version)
)

// The parent whose Drive the cases reach, and their account.
const miaToken = "ya29.the-token-of-mias-parent"

var mia = store.NewAccount("account-mia", miaToken, time.Time{})

// moment is when every profile of these cases is made and moved on.
var moment = time.Date(2026, time.September, 28, 9, 30, 0, 0, time.UTC)

// child is a profile as the first sign-in makes it.
func child(pseudonym string) *profile.Profile {
	return profile.New(profile.Student{
		ExcludedSkills: []string{},
		Grade:          3,
		Interests:      []string{"space"},
		Pseudonym:      pseudonym,
	}, "drivestore_test", moment)
}

// moved is a profile moved on the way a tool moves it before a write.
func moved(p *profile.Profile, pseudonym string) *profile.Profile {
	p.Student.Pseudonym = pseudonym
	p.Touch("drivestore_test", moment)
	return p
}

// fileOf is the profile as the store writes it.
func fileOf(t *testing.T, p *profile.Profile) []byte {
	t.Helper()

	raw, err := profile.Marshal(p)
	if err != nil {
		t.Fatalf("Marshal() error = %v, want nil", err)
	}
	return raw
}

// fixture is a stand-in Drive, a store over it, and everything the store
// records, kept in memory. The store's clock moves only when a case moves it,
// and its waits are written down rather than waited.
type fixture struct {
	fake    *drivetest.Drive
	storage store.Storage
	spans   *tracetest.SpanRecorder
	traces  *sdktrace.TracerProvider
	logs    *observer.ObservedLogs
	log     *zap.Logger
	clock   *clock
	waits   *waits
}

// newFixture is a store over an empty Drive, whose calls are given the time
// given.
func newFixture(t *testing.T, timeout time.Duration) *fixture {
	t.Helper()

	f := &fixture{fake: drivetest.New(t), spans: tracetest.NewSpanRecorder(), clock: &clock{at: moment}, waits: &waits{}}
	f.traces = sdktrace.NewTracerProvider(sdktrace.WithSampler(sdktrace.AlwaysSample()), sdktrace.WithSpanProcessor(f.spans))
	t.Cleanup(func() { _ = f.traces.Shutdown(context.Background()) })
	core, logs := observer.New(zapcore.DebugLevel)
	f.logs = logs
	f.log = zap.New(core)
	f.storage = f.instance(t, timeout)
	return f
}

// instance is another store over the same Drive: another instance of the
// service, which remembers nothing yet.
func (f *fixture) instance(t *testing.T, timeout time.Duration) store.Storage {
	t.Helper()

	settings := &drive.Settings{Root: f.fake.Root(), Timeout: timeout}
	files, err := drive.NewFiles(settings)
	if err != nil {
		t.Fatalf("NewFiles() error = %v, want nil", err)
	}
	revisions, err := drive.NewRevisions(settings)
	if err != nil {
		t.Fatalf("NewRevisions() error = %v, want nil", err)
	}
	s, err := drivestore.New(&drivestore.Settings{
		Files: files, Revisions: revisions, Timeout: timeout, Logger: f.log, Traces: f.traces, ProjectID: "a-project",
		Now: f.clock.now, Sleep: f.waits.wait,
	})
	if err != nil {
		t.Fatalf("New() error = %v, want nil", err)
	}
	return s
}

// clock is a clock a case moves on by hand, from any goroutine.
type clock struct {
	mu sync.Mutex
	at time.Time
}

func (c *clock) now() time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.at
}

func (c *clock) advance(by time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.at = c.at.Add(by)
}

// waits writes down every wait the store asks for, and waits for none of
// them, unless the context has ended.
type waits struct {
	mu    sync.Mutex
	asked []time.Duration
}

func (w *waits) wait(ctx context.Context, d time.Duration) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.asked = append(w.asked, d)
	return ctx.Err()
}

// taken is every wait the store asked for since the last time it was read.
func (w *waits) taken() []time.Duration {
	w.mu.Lock()
	defer w.mu.Unlock()
	asked := w.asked
	w.asked = nil
	return asked
}

// tries is how many times a call Drive keeps refusing is made: once, and once
// more after each pause.
const tries = 3

// failEveryTry has Drive refuse every try of the next call of a kind.
func failEveryTry(fake *drivetest.Drive, kind string, status int, reason string) {
	for range tries {
		fake.Fail(kind, status, reason)
	}
}

// setAside is the content of every file of the Drive a token reaches that was
// set aside, in the order the files were made.
func setAside(fake *drivetest.Drive, token string) [][]byte {
	var aside [][]byte
	for _, file := range fake.Files(token) {
		if file.AppProperties["mathtrail"] == "set-aside" {
			aside = append(aside, file.Content)
		}
	}
	return aside
}

// create keeps a first profile for the account, and ends the case if it
// cannot.
func (f *fixture) create(t *testing.T, account store.Account, p *profile.Profile) store.Revision {
	t.Helper()

	revision, err := f.storage.Create(t.Context(), account, p)
	if err != nil {
		t.Fatalf("Create() error = %v, want nil", err)
	}
	return revision
}

// profileFile is the one file of the account's Drive the store keeps a profile
// in, or the end of the case.
func (f *fixture) profileFile(t *testing.T, token string) *drivetest.File {
	t.Helper()

	var found []*drivetest.File
	for _, file := range f.fake.Files(token) {
		if file.AppProperties["mathtrail"] == "profile" {
			found = append(found, file)
		}
	}
	if len(found) != 1 {
		t.Fatalf("the Drive holds %d profile files, want one", len(found))
	}
	return found[0]
}

// plant writes bytes where the store keeps an account's profile, the way
// something other than the service would: into the file that carries the
// profile's marker, or into a new one when there is none.
func plant(t *testing.T, fake *drivetest.Drive, token string, raw []byte) {
	t.Helper()

	for _, file := range fake.Files(token) {
		if file.AppProperties["mathtrail"] == "profile" && !file.Trashed && !file.Own {
			fake.Edit(token, file.ID, func(planted *drivetest.File) { planted.Content = bytes.Clone(raw) })
			return
		}
	}
	fake.Put(token, &drivetest.File{
		Name: fileName, MimeType: fileType, AppProperties: maps.Clone(profileMarker), Content: bytes.Clone(raw),
	})
}

func TestTheStoreInDriveKeepsTheContract(t *testing.T) {
	t.Parallel()

	storetest.Run(t, func(t *testing.T) storetest.Harness {
		f := newFixture(t, 5*time.Second)
		return storetest.Harness{
			Storage: f.storage,
			Plant: func(t *testing.T, account store.Account, raw []byte) {
				t.Helper()
				plant(t, f.fake, account.Token(), raw)
			},
			SetAside: func(t *testing.T, account store.Account) [][]byte {
				t.Helper()
				return setAside(f.fake, account.Token())
			},
			Revoke: func(t *testing.T, account store.Account) {
				t.Helper()
				f.fake.Revoke(account.Token())
			},
			Now: f.clock.now,
		}
	})
}

// The first profile of an account is a file in a folder of its own, both
// marked, in the parent's Drive — and a new instance reads it back as it was
// written.
func TestAFirstSignInMakesTheFolderAndTheFileInIt(t *testing.T) {
	t.Parallel()

	f := newFixture(t, 5*time.Second)
	p := child("Mia")
	want := fileOf(t, p)
	created := f.create(t, mia, p)

	files := f.fake.Files(miaToken)
	if len(files) != 2 {
		t.Fatalf("the Drive holds %d files, want a folder and a file", len(files))
	}
	folder, file := files[0], files[1]
	switch {
	case folder.Name != folderName || folder.MimeType != drive.FolderType || len(folder.Parents) != 0:
		t.Errorf("folder = %q of type %q in %v, want %q of type %q in My Drive",
			folder.Name, folder.MimeType, folder.Parents, folderName, drive.FolderType)
	case !maps.Equal(folder.AppProperties, folderMarker):
		t.Errorf("folder properties = %v, want %v", folder.AppProperties, folderMarker)
	case file.Name != fileName || file.MimeType != fileType:
		t.Errorf("file = %q of type %q, want %q of type %q", file.Name, file.MimeType, fileName, fileType)
	case !slices.Equal(file.Parents, []string{folder.ID}):
		t.Errorf("file parents = %v, want the folder", file.Parents)
	case !maps.Equal(file.AppProperties, map[string]string{"mathtrail": "profile", "schema": schema}):
		t.Errorf("file properties = %v, want the profile's marker and schema %s", file.AppProperties, schema)
	case !bytes.Equal(file.Content, want):
		t.Errorf("file content =\n%s\nwant the profile as it was created:\n%s", file.Content, want)
	}
	if calls := f.fake.Calls(); !maps.Equal(calls, drivetest.Calls{"list": 3, "create": 2}) {
		t.Errorf("a first profile cost %v, want a search for it, one of the bin, a search for the folder and two creations", calls)
	}

	loaded, revision, err := f.instance(t, 5*time.Second).Load(t.Context(), mia)
	if err != nil {
		t.Fatalf("Load() by another instance error = %v, want nil", err)
	}
	if got := fileOf(t, loaded); !bytes.Equal(got, want) {
		t.Errorf("Load() by another instance =\n%s\nwant\n%s", got, want)
	}
	if revision != created {
		t.Errorf("Load() by another instance revision = %q, want %q, the one Create answered", revision, created)
	}
}

// A folder the store made once is found again by its marker, however the
// parent has renamed it, and a new profile goes into it.
func TestAFolderFoundByItsMarkerIsUsedAgain(t *testing.T) {
	t.Parallel()

	f := newFixture(t, 5*time.Second)
	folder := f.fake.Put(miaToken, &drivetest.File{
		Name: "Maths of the kids", MimeType: drive.FolderType, AppProperties: maps.Clone(folderMarker),
	})
	f.create(t, mia, child("Mia"))

	if got := f.profileFile(t, miaToken).Parents; !slices.Equal(got, []string{folder}) {
		t.Errorf("the profile is in %v, want the folder found by its marker", got)
	}
	if calls := f.fake.Calls(); !maps.Equal(calls, drivetest.Calls{"list": 3, "create": 1}) {
		t.Errorf("a first profile in a found folder cost %v, want three searches and one creation", calls)
	}
}

// The profile is found by its marker and never by its name or its place: a
// file the parent renamed and moved is still found by an instance that never
// saw it.
func TestAMovedAndRenamedProfileIsFoundByItsMarker(t *testing.T) {
	t.Parallel()

	f := newFixture(t, 5*time.Second)
	p := child("Mia")
	want := fileOf(t, p)
	f.create(t, mia, p)
	file := f.profileFile(t, miaToken)
	f.fake.Edit(miaToken, file.ID, func(edited *drivetest.File) {
		edited.Name = "notes about Mia.json"
		edited.Parents = nil
	})

	loaded, _, err := f.instance(t, 5*time.Second).Load(t.Context(), mia)
	if err != nil {
		t.Fatalf("Load() of a moved profile error = %v, want nil", err)
	}
	if got := fileOf(t, loaded); !bytes.Equal(got, want) {
		t.Errorf("Load() of a moved profile =\n%s\nwant\n%s", got, want)
	}
}

// A file the parent made themselves is never taken for the profile, whatever
// it is called and wherever it is: the service is not shown it, and does not
// look for its name.
func TestAFileOfTheParentsOwnIsNeverTakenForTheProfile(t *testing.T) {
	t.Parallel()

	f := newFixture(t, 5*time.Second)
	theirs := fileOf(t, child("Somebody else"))
	f.fake.Put(miaToken, &drivetest.File{Name: folderName, MimeType: drive.FolderType, Own: true})
	own := f.fake.Put(miaToken, &drivetest.File{Name: fileName, MimeType: fileType, Content: theirs, Own: true})
	// A file the service made itself, named like the profile but not marked
	// as one, is not the profile either.
	unmarked := f.fake.Put(miaToken, &drivetest.File{Name: fileName, MimeType: fileType, Content: theirs})

	if _, _, err := f.storage.Load(t.Context(), mia); !errors.Is(err, store.ErrNotFound) {
		t.Fatalf("Load() error = %v, want %v", err, store.ErrNotFound)
	}
	f.create(t, mia, child("Mia"))
	if made := f.profileFile(t, miaToken).ID; made == own || made == unmarked {
		t.Error("the profile is kept in a file found by its name, want a file of its own")
	}
	for _, file := range f.fake.Files(miaToken) {
		if (file.ID == own || file.ID == unmarked) && !bytes.Equal(file.Content, theirs) {
			t.Errorf("file %q was changed to\n%s\nwant it left as it was", file.Name, file.Content)
		}
	}
}

// Of two profiles — the old one restored from the bin beside a new one — the
// one changed last is read, and written, and the other is left as it is.
func TestOfTwoProfilesTheNewestIsReadAndWritten(t *testing.T) {
	t.Parallel()

	f := newFixture(t, 5*time.Second)
	older := fileOf(t, child("Mia before"))
	newer := fileOf(t, child("Mia after"))
	f.fake.Put(miaToken, &drivetest.File{Name: fileName, MimeType: fileType, AppProperties: maps.Clone(profileMarker), Content: older})
	newest := f.fake.Put(miaToken, &drivetest.File{Name: fileName, MimeType: fileType, AppProperties: maps.Clone(profileMarker), Content: newer})

	p, read, err := f.storage.Load(t.Context(), mia)
	if err != nil {
		t.Fatalf("Load() error = %v, want nil", err)
	}
	if got := fileOf(t, p); !bytes.Equal(got, newer) {
		t.Fatalf("Load() =\n%s\nwant the profile changed last:\n%s", got, newer)
	}
	written := moved(p, "Mia again")
	if _, err := f.storage.Save(t.Context(), mia, written, read); err != nil {
		t.Fatalf("Save() error = %v, want nil", err)
	}
	for _, file := range f.fake.Files(miaToken) {
		want := older
		if file.ID == newest {
			want = fileOf(t, written)
		}
		if !bytes.Equal(file.Content, want) {
			t.Errorf("file %q holds\n%s\nwant\n%s", file.Name, file.Content, want)
		}
	}
}

// An instance that has read a profile remembers its file, and reads it again
// with one call; an instance that has not searches first.
func TestAWarmInstanceReadsWithOneCall(t *testing.T) {
	t.Parallel()

	f := newFixture(t, 5*time.Second)
	f.create(t, mia, child("Mia"))
	cold := f.instance(t, 5*time.Second)

	for _, tc := range []struct {
		name    string
		storage store.Storage
		want    drivetest.Calls
	}{
		{"an instance that made the profile", f.storage, drivetest.Calls{"download": 1}},
		{"an instance that never saw it", cold, drivetest.Calls{"list": 1, "download": 1}},
		{"the same instance once it has", cold, drivetest.Calls{"download": 1}},
	} {
		f.fake.ResetCalls()
		if _, _, err := tc.storage.Load(t.Context(), mia); err != nil {
			t.Fatalf("%s: Load() error = %v, want nil", tc.name, err)
		}
		if calls := f.fake.Calls(); !maps.Equal(calls, tc.want) {
			t.Errorf("%s: Load() cost %v, want %v", tc.name, calls, tc.want)
		}
	}
}

// A file that went from under what an instance remembers is searched for
// again: the profile is wherever the marker is now, and nowhere when no file
// carries it.
func TestAFileGoneFromUnderTheMemoryIsSearchedFor(t *testing.T) {
	t.Parallel()

	f := newFixture(t, 5*time.Second)
	f.create(t, mia, child("Mia"))
	f.fake.Delete(miaToken, f.profileFile(t, miaToken).ID)
	restored := fileOf(t, child("Mia restored"))
	f.fake.Put(miaToken, &drivetest.File{Name: fileName, MimeType: fileType, AppProperties: maps.Clone(profileMarker), Content: restored})
	f.fake.ResetCalls()

	p, _, err := f.storage.Load(t.Context(), mia)
	if err != nil {
		t.Fatalf("Load() error = %v, want nil", err)
	}
	if got := fileOf(t, p); !bytes.Equal(got, restored) {
		t.Errorf("Load() =\n%s\nwant the profile the search finds:\n%s", got, restored)
	}
	if calls := f.fake.Calls(); !maps.Equal(calls, drivetest.Calls{"download": 2, "list": 1}) {
		t.Errorf("Load() cost %v, want the remembered file, a search and the found one", calls)
	}

	f.fake.Delete(miaToken, f.profileFile(t, miaToken).ID)
	if _, _, err := f.storage.Load(t.Context(), mia); !errors.Is(err, store.ErrNotFound) {
		t.Errorf("Load() with no profile file left: error = %v, want %v", err, store.ErrNotFound)
	}
}

// An account that carries no token reaches no Drive, and nothing is asked of
// any.
func TestAnAccountWithNoTokenReachesNothing(t *testing.T) {
	t.Parallel()

	f := newFixture(t, 5*time.Second)
	f.create(t, mia, child("Mia"))
	_, read, err := f.storage.Load(t.Context(), mia)
	if err != nil {
		t.Fatalf("Load() error = %v, want nil", err)
	}
	f.fake.ResetCalls()
	tokenless := store.NewAccount(mia.ID, "", time.Time{})

	if _, err := f.storage.Create(t.Context(), tokenless, child("Mia")); err == nil {
		t.Error("Create() for an account with no token: error = nil, want a refusal")
	}
	if _, _, err := f.storage.Load(t.Context(), tokenless); err == nil || errors.Is(err, store.ErrNotFound) {
		t.Errorf("Load() for an account with no token: error = %v, want a refusal that is not %v", err, store.ErrNotFound)
	}
	if _, err := f.storage.Save(t.Context(), tokenless, moved(child("Mia"), "Mia"), read); err == nil {
		t.Error("Save() for an account with no token: error = nil, want a refusal")
	}
	if _, err := f.storage.Export(t.Context(), tokenless); err == nil {
		t.Error("Export() for an account with no token: error = nil, want a refusal")
	}
	if _, err := f.storage.StartOver(t.Context(), tokenless, child("Mia")); err == nil {
		t.Error("StartOver() for an account with no token: error = nil, want a refusal")
	}
	if calls := f.fake.Calls(); len(calls) != 0 {
		t.Errorf("an account with no token cost %v, want no call at all", calls)
	}
}

// Drive's refusals reach the caller as the store's own where the store has a
// word for them, and as Drive's otherwise — never as a refusal a caller would
// act on wrongly.
func TestDrivesRefusalsAreToldInTheStoresWords(t *testing.T) {
	t.Parallel()

	refusals := []error{
		store.ErrNotFound, store.ErrConflict, store.ErrCorrupted, store.ErrDamaged, store.ErrAccessRevoked,
		store.ErrAccessExpired, store.ErrInBin, store.ErrBehind, store.ErrStorageFull, store.ErrUnavailable,
	}
	for _, tc := range []struct {
		name string
		// spoil makes the next read of the profile go wrong.
		spoil func(t *testing.T, f *fixture)
		want  error
		// timeout is the time a call to Drive is given, when a case needs one
		// shorter than the time any honest call takes.
		timeout time.Duration
	}{
		{
			name:  "access taken back",
			spoil: func(_ *testing.T, f *fixture) { f.fake.Revoke(miaToken) },
			want:  store.ErrAccessRevoked,
		},
		{
			name:  "a file too large for any profile",
			spoil: func(t *testing.T, f *fixture) { plant(t, f.fake, miaToken, oversized) },
			want:  store.ErrDamaged,
		},
		{
			name: "a pause Drive asks for, and goes on asking for",
			spoil: func(_ *testing.T, f *fixture) {
				failEveryTry(f.fake, drivetest.Download, http.StatusTooManyRequests, "rateLimitExceeded")
			},
			want: store.ErrUnavailable,
		},
		{
			name: "a failure of Drive's that goes on",
			spoil: func(_ *testing.T, f *fixture) {
				failEveryTry(f.fake, drivetest.Download, http.StatusServiceUnavailable, "backendError")
			},
			want: store.ErrUnavailable,
		},
		{
			name: "a token that grants nothing of Drive",
			spoil: func(_ *testing.T, f *fixture) {
				f.fake.Fail(drivetest.Download, http.StatusForbidden, "insufficientPermissions")
			},
			want: store.ErrAccessRevoked,
		},
		{
			name:    "a call that takes too long",
			spoil:   func(_ *testing.T, f *fixture) { f.fake.Stall(drivetest.Download) },
			want:    context.DeadlineExceeded,
			timeout: 200 * time.Millisecond,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			f := newFixture(t, cmp.Or(tc.timeout, 5*time.Second))
			f.create(t, mia, child("Mia"))
			tc.spoil(t, f)

			p, _, err := f.storage.Load(t.Context(), mia)
			if !errors.Is(err, tc.want) {
				t.Errorf("Load() error = %v, want %v", err, tc.want)
			}
			if p != nil {
				t.Error("Load() returned a profile alongside the refusal")
			}
			for _, refusal := range refusals {
				if errors.Is(err, refusal) && !errors.Is(tc.want, refusal) {
					t.Errorf("Load() error = %v, want it not to be %v", err, refusal)
				}
			}
		})
	}
}

// A search Drive failed to answer is not an account with no profile: taken
// for one, it would have a second profile made beside the first. So a failed
// search reaches the caller as the failure it is, and nothing is created.
func TestASearchThatFailedIsNotAProfileMissing(t *testing.T) {
	t.Parallel()

	f := newFixture(t, 5*time.Second)
	f.create(t, mia, child("Mia"))
	cold := f.instance(t, 5*time.Second)
	for name, try := range map[string]func() error{
		"Load": func() error {
			_, _, err := cold.Load(t.Context(), mia)
			return err
		},
		"Create": func() error {
			_, err := cold.Create(t.Context(), mia, child("A second Mia"))
			return err
		},
		"Export": func() error {
			_, err := cold.Export(t.Context(), mia)
			return err
		},
	} {
		failEveryTry(f.fake, drivetest.List, http.StatusServiceUnavailable, "backendError")
		err := try()
		if !errors.Is(err, store.ErrUnavailable) || errors.Is(err, store.ErrNotFound) {
			t.Errorf("%s() while the search fails: error = %v, want %v and not %v", name, err, store.ErrUnavailable, store.ErrNotFound)
		}
	}
	f.profileFile(t, miaToken)
}

// A profile deleted between its read and its write is not written, and not
// made again in its place: the account has no profile, and starting over is
// the parent's to ask for.
func TestAProfileDeletedBeforeItsWriteIsNotMadeAgain(t *testing.T) {
	t.Parallel()

	f := newFixture(t, 5*time.Second)
	f.create(t, mia, child("Mia"))
	p, read, err := f.storage.Load(t.Context(), mia)
	if err != nil {
		t.Fatalf("Load() error = %v, want nil", err)
	}
	f.fake.Delete(miaToken, f.profileFile(t, miaToken).ID)

	if _, err := f.storage.Save(t.Context(), mia, moved(p, "Mia the brave"), read); !errors.Is(err, store.ErrNotFound) {
		t.Errorf("Save() of a deleted profile error = %v, want %v", err, store.ErrNotFound)
	}
	for _, file := range f.fake.Files(miaToken) {
		if file.AppProperties["mathtrail"] == "profile" {
			t.Errorf("a profile file %q is in the Drive, want none", file.Name)
		}
	}
	if _, _, err := f.storage.Load(t.Context(), mia); !errors.Is(err, store.ErrNotFound) {
		t.Errorf("Load() after the deleted profile's write error = %v, want %v", err, store.ErrNotFound)
	}
}

// A full Drive refuses a write, and is not taken for any refusal of the
// store's: nothing was saved, and reading again would not help.
func TestAFullDriveRefusesTheWrite(t *testing.T) {
	t.Parallel()

	f := newFixture(t, 5*time.Second)
	f.create(t, mia, child("Mia"))
	p, read, err := f.storage.Load(t.Context(), mia)
	if err != nil {
		t.Fatalf("Load() error = %v, want nil", err)
	}
	f.fake.Fail(drivetest.Update, http.StatusForbidden, "storageQuotaExceeded")

	_, err = f.storage.Save(t.Context(), mia, moved(p, "Mia the brave"), read)
	if !errors.Is(err, store.ErrStorageFull) {
		t.Errorf("Save() error = %v, want %v", err, store.ErrStorageFull)
	}
	for _, refusal := range []error{store.ErrNotFound, store.ErrConflict, store.ErrCorrupted, store.ErrAccessRevoked} {
		if errors.Is(err, refusal) {
			t.Errorf("Save() error = %v, want it not to be %v", err, refusal)
		}
	}

	// A first profile in a full Drive has no room either: the folder it would
	// go into is refused, and the profile with it.
	leo := store.NewAccount("account-leo", "ya29.the-token-of-leos-parent", time.Time{})
	f.fake.Fail(drivetest.Create, http.StatusForbidden, "storageQuotaExceeded")
	if _, err := f.storage.Create(t.Context(), leo, child("Leo")); !errors.Is(err, store.ErrStorageFull) {
		t.Errorf("Create() in a full Drive error = %v, want %v", err, store.ErrStorageFull)
	}
	if files := f.fake.Files(leo.Token()); len(files) != 0 {
		t.Errorf("a full Drive holds %d files after a refused first profile, want none", len(files))
	}
}

// Every write carries a copy of the version of the file's shape beside the
// marker, whatever the copy said before: the two never drift apart.
func TestEveryWriteCarriesTheVersionOfTheShape(t *testing.T) {
	t.Parallel()

	f := newFixture(t, 5*time.Second)
	f.create(t, mia, child("Mia"))
	file := f.profileFile(t, miaToken)
	f.fake.Edit(miaToken, file.ID, func(edited *drivetest.File) { edited.AppProperties["schema"] = "0" })

	p, read, err := f.storage.Load(t.Context(), mia)
	if err != nil {
		t.Fatalf("Load() error = %v, want nil", err)
	}
	if _, err := f.storage.Save(t.Context(), mia, moved(p, "Mia the brave"), read); err != nil {
		t.Fatalf("Save() error = %v, want nil", err)
	}
	want := map[string]string{"mathtrail": "profile", "schema": schema}
	if got := f.profileFile(t, miaToken).AppProperties; !maps.Equal(got, want) {
		t.Errorf("after Save() the properties are %v, want %v", got, want)
	}
}

// The profile is exported as the file it is, where the parent can open it:
// the folder it is in, the file, and a link — and a folder of the parent's
// own, which the service may not see, leaves only the folder's name unsaid.
func TestTheProfileIsExportedAsWhereItIs(t *testing.T) {
	t.Parallel()

	f := newFixture(t, 5*time.Second)
	f.create(t, mia, child("Mia"))
	file := f.profileFile(t, miaToken)

	location, err := f.storage.Export(t.Context(), mia)
	if err != nil {
		t.Fatalf("Export() error = %v, want nil", err)
	}
	want := store.Location{Folder: folderName, File: fileName, Link: drivetest.WebViewLink(file.ID)}
	if !reflect.DeepEqual(location, want) {
		t.Errorf("Export() = %+v, want %+v", location, want)
	}

	theirs := f.fake.Put(miaToken, &drivetest.File{Name: "Kids", MimeType: drive.FolderType, Own: true})
	f.fake.Edit(miaToken, file.ID, func(edited *drivetest.File) { edited.Parents = []string{theirs} })
	location, err = f.storage.Export(t.Context(), mia)
	if err != nil {
		t.Fatalf("Export() from a folder of the parent's own error = %v, want nil", err)
	}
	want.Folder = ""
	if !reflect.DeepEqual(location, want) {
		t.Errorf("Export() from a folder of the parent's own = %+v, want %+v", location, want)
	}
}

func TestAStoreMissingAPartIsNotBuilt(t *testing.T) {
	t.Parallel()

	files, err := drive.NewFiles(&drive.Settings{Root: drive.Google, Timeout: time.Second})
	if err != nil {
		t.Fatalf("NewFiles() error = %v, want nil", err)
	}
	revisions, err := drive.NewRevisions(&drive.Settings{Root: drive.Google, Timeout: time.Second})
	if err != nil {
		t.Fatalf("NewRevisions() error = %v, want nil", err)
	}
	whole := drivestore.Settings{
		Files: files, Revisions: revisions, Timeout: time.Second, Logger: zap.NewNop(), Traces: sdktrace.NewTracerProvider(),
	}
	if _, err := drivestore.New(&whole); err != nil {
		t.Fatalf("New() with every part error = %v, want nil", err)
	}
	if _, err := drivestore.New(nil); !errors.Is(err, drivestore.ErrSettings) {
		t.Errorf("New(nil) error = %v, want %v", err, drivestore.ErrSettings)
	}
	for name, remove := range map[string]func(*drivestore.Settings){
		"the calls":         func(s *drivestore.Settings) { s.Files = nil },
		"the history":       func(s *drivestore.Settings) { s.Revisions = nil },
		"a time for a call": func(s *drivestore.Settings) { s.Timeout = 0 },
		"a logger":          func(s *drivestore.Settings) { s.Logger = nil },
		"traces":            func(s *drivestore.Settings) { s.Traces = nil },
	} {
		settings := whole
		remove(&settings)
		if _, err := drivestore.New(&settings); !errors.Is(err, drivestore.ErrSettings) {
			t.Errorf("New() without %s error = %v, want %v", name, err, drivestore.ErrSettings)
		}
	}
}

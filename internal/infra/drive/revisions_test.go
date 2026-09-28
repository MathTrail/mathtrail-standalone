package drive_test

import (
	"bytes"
	"errors"
	"maps"
	"net/http"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/MathTrail/mathtrail-standalone/internal/infra/drive"
	"github.com/MathTrail/mathtrail-standalone/internal/infra/drive/drivetest"
)

// revisionsAt is the calls to the history of files, made to the API at root
// and given the time given.
func revisionsAt(t *testing.T, root string, timeout time.Duration) drive.Revisions {
	t.Helper()

	revisions, err := drive.NewRevisions(&drive.Settings{Root: root, Timeout: timeout})
	if err != nil {
		t.Fatalf("NewRevisions() error = %v, want nil", err)
	}
	return revisions
}

// withHistory is a file of the stand-in's with three states of its content,
// the earliest kept forever, and the calls to its history.
func withHistory(t *testing.T) (fake *drivetest.Drive, revisions drive.Revisions, id string, states []drivetest.Revision) {
	t.Helper()

	fake = drivetest.New(t)
	id = fake.Put(token, &drivetest.File{MimeType: "application/json", Revisions: []drivetest.Revision{
		{KeepForever: true, Content: []byte(`{"state": 1}`)},
		{Content: []byte(`{"state": 2}`)},
		{Content: []byte(`{"state": 3}`)},
	}})
	return fake, revisionsAt(t, fake.Root(), 5*time.Second), id, fileIn(t, fake, token, id).Revisions
}

func TestTheCallsToHistoryAreNotBuiltOnSettingsNoCallCouldBeMadeWith(t *testing.T) {
	t.Parallel()

	if _, err := drive.NewRevisions(nil); !errors.Is(err, drive.ErrSettings) {
		t.Errorf("NewRevisions(nil) error = %v, want %v", err, drive.ErrSettings)
	}
	if _, err := drive.NewRevisions(&drive.Settings{Root: drive.Google}); !errors.Is(err, drive.ErrSettings) {
		t.Errorf("NewRevisions() with no time for a call: error = %v, want %v", err, drive.ErrSettings)
	}
}

// Every revision of a file is listed, with when it was made and whether it is
// kept forever, by one request.
func TestEveryRevisionIsListed(t *testing.T) {
	t.Parallel()

	fake, revisions, id, states := withHistory(t)
	listed, err := revisions.List(t.Context(), token, id)
	if err != nil {
		t.Fatalf("List() error = %v, want nil", err)
	}
	want := make(map[string]drive.Revision, len(states))
	for _, state := range states {
		want[state.ID] = drive.Revision{ID: state.ID, ModifiedTime: state.ModifiedTime, KeepForever: state.KeepForever}
	}
	got := make(map[string]drive.Revision, len(listed))
	for _, revision := range listed {
		got[revision.ID] = revision
	}
	if !maps.EqualFunc(got, want, func(a, b drive.Revision) bool {
		return a.ID == b.ID && a.KeepForever == b.KeepForever && a.ModifiedTime.Equal(b.ModifiedTime)
	}) {
		t.Errorf("List() = %v, want %v", got, want)
	}
	if calls := fake.Calls(); !maps.Equal(calls, drivetest.Calls{"revisions": 1}) {
		t.Errorf("List() cost %v, want one request", calls)
	}
}

// What a revision held is given only once it is kept forever: keeping one is
// what makes it readable, and it can then be deleted, which is how a revision
// kept forever is let go.
func TestARevisionIsReadOnceKeptAndLetGoByDeleting(t *testing.T) {
	t.Parallel()

	fake, revisions, id, states := withHistory(t)
	ctx := t.Context()
	earliest, middle := states[0], states[1]

	if got := held(t, revisions, id, earliest.ID); !bytes.Equal(got, earliest.Content) {
		t.Errorf("Download() of a revision kept forever = %s, want %s", got, earliest.Content)
	}
	if err := second(revisions.Download(ctx, token, id, middle.ID, 1<<10)); !errors.Is(err, drive.ErrNotKept) {
		t.Errorf("Download() of a revision not kept forever error = %v, want %v", err, drive.ErrNotKept)
	}
	if err := second(revisions.Download(ctx, token, id, earliest.ID, 4)); !errors.Is(err, drive.ErrTooLarge) {
		t.Errorf("Download() of a revision past its limit error = %v, want %v", err, drive.ErrTooLarge)
	}

	if err := revisions.Keep(ctx, token, id, middle.ID); err != nil {
		t.Fatalf("Keep() error = %v, want nil", err)
	}
	if got := held(t, revisions, id, middle.ID); !bytes.Equal(got, middle.Content) {
		t.Errorf("Download() of a revision once kept = %s, want %s", got, middle.Content)
	}

	if err := revisions.Delete(ctx, token, id, earliest.ID); err != nil {
		t.Fatalf("Delete() error = %v, want nil", err)
	}
	left := fileIn(t, fake, token, id).Revisions
	if slices.ContainsFunc(left, func(revision drivetest.Revision) bool { return revision.ID == earliest.ID }) {
		t.Error("the deleted revision is still in the file's history")
	}
	if err := second(revisions.Download(ctx, token, id, earliest.ID, 1<<10)); !errors.Is(err, drive.ErrNotFound) {
		t.Errorf("Download() of a deleted revision error = %v, want %v", err, drive.ErrNotFound)
	}
}

// held is what a revision held, or the end of the case.
func held(t *testing.T, revisions drive.Revisions, file, revision string) []byte {
	t.Helper()

	content, err := revisions.Download(t.Context(), token, file, revision, 1<<10)
	if err != nil {
		t.Fatalf("Download() error = %v, want nil", err)
	}
	return content
}

// A call about a revision names the file and the revision, or it is not sent;
// and no error of the history names either, repeats Drive's words or carries
// the token.
func TestARevisionIsAlwaysNamedAndNeverTold(t *testing.T) {
	t.Parallel()

	fake, revisions, id, states := withHistory(t)
	ctx := t.Context()
	for call, err := range map[string]error{
		"List":     second(revisions.List(ctx, token, "")),
		"Download": second(revisions.Download(ctx, token, id, "", 1<<10)),
		"Keep":     revisions.Keep(ctx, "", "", states[1].ID),
		"Delete":   revisions.Delete(ctx, token, "", states[0].ID),
	} {
		if err == nil {
			t.Errorf("%s() of no file or no revision: error = nil, want a refusal", call)
		}
	}
	if calls := fake.Calls(); len(calls) != 0 {
		t.Errorf("calls of no file cost %v, want none sent", calls)
	}

	fake.Fail(drivetest.DownloadRevision, http.StatusForbidden, "download_restricted_for_revision")
	_, restricted := revisions.Download(ctx, token, id, states[2].ID, 1<<10)
	_, missing := revisions.Download(ctx, token, id, "a-revision-there-never-was", 1<<10)
	for _, err := range []error{restricted, missing} {
		for _, secret := range []string{id, states[2].ID, "a-revision-there-never-was", token, "cannot be downloaded", "Revision not found"} {
			if err != nil && strings.Contains(err.Error(), secret) {
				t.Errorf("error = %q, which carries %q", err, secret)
			}
		}
	}
}

// Access taken back reaches no file's history either.
func TestARevokedTokenReachesNoHistory(t *testing.T) {
	t.Parallel()

	fake, revisions, id, states := withHistory(t)
	fake.Revoke(token)
	ctx := t.Context()
	for call, err := range map[string]error{
		"List":     second(revisions.List(ctx, token, id)),
		"Download": second(revisions.Download(ctx, token, id, states[0].ID, 1<<10)),
		"Keep":     revisions.Keep(ctx, token, id, states[1].ID),
		"Delete":   revisions.Delete(ctx, token, id, states[0].ID),
	} {
		if !errors.Is(err, drive.ErrUnauthorized) {
			t.Errorf("%s() with a revoked token error = %v, want %v", call, err, drive.ErrUnauthorized)
		}
	}
}

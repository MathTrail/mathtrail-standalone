package mcpserver_test

import (
	"context"
	"maps"
	"net/http"
	"slices"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"go.uber.org/zap/zapcore"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/profile"
	"github.com/MathTrail/mathtrail-standalone/internal/infra/drive"
	"github.com/MathTrail/mathtrail-standalone/internal/infra/drive/drivetest"
	"github.com/MathTrail/mathtrail-standalone/internal/store"
	drivestore "github.com/MathTrail/mathtrail-standalone/internal/store/drive"
	mcpserver "github.com/MathTrail/mathtrail-standalone/internal/transport/mcp"
)

// The parent the tools act for in these cases, as the sign-in of bearer_test
// vouches for them: their account, reached with the Google token of their
// Drive.
var parent = store.NewAccount(vouchedUser, googleToken, time.Time{})

// overDrive is an instance of the service over a store of its own in the
// stand-in Drive: a client connected to its tools, the store, and the clock
// its tools write the profile by.
type overDrive struct {
	session *mcp.ClientSession
	kept    store.Storage
	moving  *clock
	h       *harness
}

// instanceOverDrive serves the tools of the lesson over a store of its own in
// the stand-in Drive — another instance of the service, which remembers
// nothing yet — to the parent signed in with the token the reader vouches for.
func instanceOverDrive(t *testing.T, fake *drivetest.Drive) *overDrive {
	t.Helper()
	return instanceOverDriveReading(t, fake, readerEndingIn(time.Hour))
}

// instanceOverDriveReading is instanceOverDrive, signing the parent in with
// the reader given.
func instanceOverDriveReading(t *testing.T, fake *drivetest.Drive, reader mcpserver.TokenReader) *overDrive {
	t.Helper()

	settings := &drive.Settings{Root: fake.Root(), Timeout: 5 * time.Second}
	files, err := drive.NewFiles(settings)
	if err != nil {
		t.Fatalf("NewFiles() error = %v, want nil", err)
	}
	revisions, err := drive.NewRevisions(settings)
	if err != nil {
		t.Fatalf("NewRevisions() error = %v, want nil", err)
	}
	h := newHarness(t)
	kept, err := drivestore.New(&drivestore.Settings{
		Files: files, Revisions: revisions, Timeout: 5 * time.Second, Logger: h.log, Traces: h.traces, ProjectID: projectID,
	})
	if err != nil {
		t.Fatalf("drivestore.New() error = %v, want nil", err)
	}
	moving := &clock{at: lessonDay}
	service := lessonService(t, h, kept, moving, nil)
	h.start(t, mcpserver.BearerSignIn(reader, metadata),
		slices.Concat(service.ProfileTools(), service.TaskTools())...)
	session, err := h.connectWith(t, vouchedToken)
	if err != nil {
		t.Fatalf("connectWith() error = %v, want nil", err)
	}
	return &overDrive{session: session, kept: kept, moving: moving, h: h}
}

// costOf calls a tool, holds it to having answered, and is what the call cost
// in calls to Drive.
func costOf(t *testing.T, fake *drivetest.Drive, session *mcp.ClientSession, tool string, args any) drivetest.Calls {
	t.Helper()

	fake.ResetCalls()
	if result := call(t, session, tool, args); result.IsError {
		t.Fatalf("%s failed: %s", tool, textOf(t, result))
	}
	return fake.Calls()
}

// What each tool costs in calls to Drive, on an instance that has not yet
// found the parent's file and on one that remembers it: a read is one call
// once the file is known, and a write reads the file, reads it once more to be
// sure nothing changed it in between, and writes it. A whole task — asked for,
// its package fetched, handed in, answered — is ten calls on a warm instance,
// and each look the card waiting for it takes is one more. A task written
// ahead costs the same writes, and next_task hands it out for one write.
// The profile's own tool and the progress also say where the file is, which
// takes a search and the folder's name.
func TestEveryToolStaysWithinItsDriveBudget(t *testing.T) {
	t.Parallel()

	fake := drivetest.New(t)
	warm := instanceOverDrive(t, fake)
	session := warm.session
	read := drivetest.Calls{"download": 1}
	write := drivetest.Calls{"download": 2, "update": 1}
	located := drivetest.Calls{"download": 1, "list": 1, "get": 1}

	// No profile yet: the search that finds none, and the search of the bin.
	if got, want := costOf(t, fake, session, "get_profile", map[string]any{}), (drivetest.Calls{"list": 2}); !maps.Equal(got, want) {
		t.Errorf("get_profile with no profile cost %v, want %v", got, want)
	}
	// The first profile: the two searches of the read, the two the creation
	// makes sure with, the search for the folder, the folder and the file.
	if got, want := costOf(t, fake, session, "save_profile", map[string]any{
		"pseudonym": "Otter", "grade": 2, "interests": []string{"sport"},
	}), (drivetest.Calls{"list": 5, "create": 2}); !maps.Equal(got, want) {
		t.Errorf("save_profile making the first profile cost %v, want %v", got, want)
	}

	for _, step := range []struct {
		tool string
		args any
		want drivetest.Calls
	}{
		{"get_profile", map[string]any{}, located},
		{"get_progress", map[string]any{}, located},
		{"read_progress", map[string]any{}, located},
		{"save_profile", map[string]any{"interests": []string{"sport", "space"}}, write},
		{"save_profile", map[string]any{"interests": []string{"sport", "space"}}, read},
		{"edit_profile", map[string]any{"grade": 3}, write},
		{"edit_profile", map[string]any{"grade": 3}, read},
		{"next_task", raceChoice, write},
		// Asked again while the request is open, it hands the same one back.
		{"next_task", raceChoice, read},
	} {
		if got := costOf(t, fake, session, step.tool, step.args); !maps.Equal(got, step.want) {
			t.Errorf("%s on a warm instance cost %v, want %v", step.tool, got, step.want)
		}
	}

	p, _, err := warm.kept.Load(t.Context(), parent)
	if err != nil || p.OpenRequest == nil {
		t.Fatalf("Load() = %v, %v, want the request next_task opened", p, err)
	}
	for _, tool := range []string{"get_package", "read_task"} {
		if got := costOf(t, fake, session, tool, map[string]any{"request_id": p.OpenRequest.ID}); !maps.Equal(got, read) {
			t.Errorf("%s on a warm instance cost %v, want %v", tool, got, read)
		}
	}
	if got := costOf(t, fake, session, "submit_task", raceOn(p.OpenRequest)); !maps.Equal(got, write) {
		t.Errorf("submit_task on a warm instance cost %v, want %v", got, write)
	}
	if p, _, err = warm.kept.Load(t.Context(), parent); err != nil || p.CurrentTask == nil {
		t.Fatalf("Load() = %v, %v, want the task submit_task accepted", p, err)
	}
	answer := map[string]any{"task_id": p.CurrentTask.ID, "answer": "C"}
	if got := costOf(t, fake, session, "submit_answer", answer); !maps.Equal(got, write) {
		t.Errorf("submit_answer on a warm instance cost %v, want %v", got, write)
	}
	// The same answer again is told again, and written nowhere.
	if got := costOf(t, fake, session, "submit_answer", answer); !maps.Equal(got, read) {
		t.Errorf("submit_answer given again cost %v, want %v", got, read)
	}

	aheadStaysWithinItsBudget(t, fake, warm)
}

// aheadStaysWithinItsBudget holds the next task, written ahead after the task
// on the card, to the Drive budget: its request opened, its package handed
// back while it is written, it kept, nothing more to write, and next_task
// handing it out on the card it draws.
func aheadStaysWithinItsBudget(t *testing.T, fake *drivetest.Drive, warm *overDrive) {
	t.Helper()

	session := warm.session
	read := drivetest.Calls{"download": 1}
	write := drivetest.Calls{"download": 2, "update": 1}
	for _, step := range []struct {
		tool string
		want drivetest.Calls
	}{
		{"prepare_task", write},
		{"prepare_task", read},
	} {
		if got := costOf(t, fake, session, step.tool, aheadChoice); !maps.Equal(got, step.want) {
			t.Errorf("%s on a warm instance cost %v, want %v", step.tool, got, step.want)
		}
	}
	p, _, err := warm.kept.Load(t.Context(), parent)
	if err != nil || p.OpenRequest == nil {
		t.Fatalf("Load() = %v, %v, want the request prepare_task opened", p, err)
	}
	if got := costOf(t, fake, session, "submit_task", relayOn(p.OpenRequest)); !maps.Equal(got, write) {
		t.Errorf("submit_task kept ahead cost %v, want %v", got, write)
	}
	if got := costOf(t, fake, session, "prepare_task", aheadChoice); !maps.Equal(got, read) {
		t.Errorf("prepare_task with a task kept cost %v, want %v", got, read)
	}
	if got := costOf(t, fake, session, "next_task", map[string]any{"language": "en"}); !maps.Equal(got, write) {
		t.Errorf("next_task handing out the task kept cost %v, want %v", got, write)
	}
}

// When Drive cannot say where the profile's file is, the progress and the
// profile are shown all the same, with nothing of where the file is kept: a
// card draws no parent's data, and the model is told to say so if the adult
// asks. The failure is the line of the call that failed, one for each.
func TestAProfileDriveCannotPlaceJustNowIsShownAllTheSame(t *testing.T) {
	t.Parallel()

	fake := drivetest.New(t)
	warm := instanceOverDrive(t, fake)
	if result := call(t, warm.session, "save_profile", map[string]any{"pseudonym": "Otter", "grade": 2}); result.IsError {
		t.Fatalf("save_profile failed: %s", textOf(t, result))
	}

	tools := []string{"get_progress", "read_progress", "get_profile"}
	for _, tool := range tools {
		// The name of the folder is the last thing a location asks Drive for.
		fake.Fail(drivetest.Get, http.StatusBadRequest, "badRequest")
		result := call(t, warm.session, tool, map[string]any{})
		if raw := string(rawPayload(t, result)); strings.Contains(raw, `"location"`) {
			t.Errorf("%s hands back %s, want no location in it", tool, raw)
		}
		if words := textOf(t, result); !strings.Contains(words, "could not be found just now") {
			t.Errorf("%s says %q, want it to say where the file is could not be found just now", tool, words)
		}
	}

	warm.h.settle()
	failed := 0
	for _, line := range linesOf(warm.h, "drive_call") {
		if line.Level == zapcore.WarnLevel && line.ContextMap()["op"] == "get" {
			failed++
		}
	}
	if failed != len(tools) {
		t.Errorf("failed drive_call lines of the folder = %d, want one for each of the %d calls", failed, len(tools))
	}
}

// An instance that has not yet found the parent's file searches for it once,
// and then knows it; and the first write of a day keeps its revision forever
// in the same upload, reading nothing of the history: a write takes nothing
// from it.
func TestAColdInstanceSearchesOnceAndADayKeepsItsFirstWrite(t *testing.T) {
	t.Parallel()

	fake := drivetest.New(t)
	first := instanceOverDrive(t, fake)
	if result := call(t, first.session, "save_profile", map[string]any{"pseudonym": "Otter", "grade": 2}); result.IsError {
		t.Fatalf("save_profile failed: %s", textOf(t, result))
	}

	for _, step := range []struct {
		tool string
		args any
		want drivetest.Calls
	}{
		{"get_progress", map[string]any{}, drivetest.Calls{"list": 2, "download": 1, "get": 1}},
		{"get_profile", map[string]any{}, drivetest.Calls{"list": 2, "download": 1, "get": 1}},
		{"save_profile", map[string]any{"interests": []string{"space"}}, drivetest.Calls{"list": 1, "download": 2, "update": 1}},
	} {
		cold := instanceOverDrive(t, fake)
		if got := costOf(t, fake, cold.session, step.tool, step.args); !maps.Equal(got, step.want) {
			t.Errorf("%s on a cold instance cost %v, want %v", step.tool, got, step.want)
		}
	}

	first.moving.advance(24 * time.Hour)
	if got, want := costOf(t, fake, first.session, "save_profile", map[string]any{"interests": []string{"sport"}}),
		(drivetest.Calls{"download": 2, "update": 1}); !maps.Equal(got, want) {
		t.Errorf("the first save_profile of a day cost %v, want %v", got, want)
	}
}

// goneWhenPlaced is a store whose profile is deleted as the progress asks
// where its file is: the read before found it, every read after finds none.
type goneWhenPlaced struct {
	store.Storage
	gone atomic.Bool
}

func (s *goneWhenPlaced) Load(ctx context.Context, account store.Account) (*profile.Profile, store.Revision, error) {
	if s.gone.Load() {
		return nil, "", store.ErrNotFound
	}
	return s.Storage.Load(ctx, account)
}

func (s *goneWhenPlaced) Export(context.Context, store.Account) (store.Location, error) {
	s.gone.Store(true)
	return store.Location{}, store.ErrNotFound
}

// A profile deleted between its read and the search for its file is not
// shown as if it were there: it is read again, and found gone.
func TestAProfileGoneWhileItIsPlacedIsToldAsGone(t *testing.T) {
	t.Parallel()

	for _, tool := range []string{"get_progress", "get_profile"} {
		_, session := lesson(t, &goneWhenPlaced{Storage: keptWith(t, "masha")})
		if got := payloadOf[progressPayload](t, call(t, session, tool, nil)); got.Screen != "first_run" {
			t.Errorf("%s shows %q, want the first sign-in of a profile gone", tool, got.Screen)
		}
	}
}

// revokedWhenPlaced is a store whose access is taken back as the progress asks
// where its file is.
type revokedWhenPlaced struct{ store.Storage }

func (revokedWhenPlaced) Export(context.Context, store.Account) (store.Location, error) {
	return store.Location{}, store.ErrAccessRevoked
}

// Access taken back as the file is placed is no failure of Drive's to show the
// progress past: it is told as it is told everywhere, with the sign-in asked
// for again.
func TestAccessTakenBackWhileTheFileIsPlacedIsToldSo(t *testing.T) {
	t.Parallel()

	for _, tool := range []string{"get_progress", "get_profile"} {
		_, session := lesson(t, revokedWhenPlaced{keptWith(t, "masha")})
		wantOurSentence(t, call(t, session, tool, nil), "MathTrail can no longer reach the child's profile")
	}
}

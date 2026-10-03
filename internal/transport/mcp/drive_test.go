package mcpserver_test

import (
	"maps"
	"slices"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"

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
// handed in, answered — is nine calls on a warm instance. The profile's own
// tool and the progress also say where the file is, which takes a search and
// the folder's name.
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
}

// An instance that has not yet found the parent's file searches for it once,
// and then knows it; and the first write of a day keeps its revision forever
// in the same upload, reading the history to count what is kept and the file
// once more after it.
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
		(drivetest.Calls{"download": 3, "update": 1, "revisions": 1}); !maps.Equal(got, want) {
		t.Errorf("the first save_profile of a day cost %v, want %v", got, want)
	}
}

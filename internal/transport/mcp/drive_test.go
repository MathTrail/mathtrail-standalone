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
var parent = store.NewAccount(vouchedUser, "a-google-access-token")

// instanceOverDrive serves the tools of the lesson over a store of its own in
// the stand-in Drive — another instance of the service, which remembers
// nothing yet — to the parent signed in with the token the reader vouches for.
func instanceOverDrive(t *testing.T, fake *drivetest.Drive) (*mcp.ClientSession, store.Storage) {
	t.Helper()

	files, err := drive.NewFiles(&drive.Settings{Root: fake.Root(), Timeout: 5 * time.Second})
	if err != nil {
		t.Fatalf("NewFiles() error = %v, want nil", err)
	}
	h := newHarness(t)
	kept, err := drivestore.New(&drivestore.Settings{Files: files, Logger: h.log, Traces: h.traces, ProjectID: projectID})
	if err != nil {
		t.Fatalf("drivestore.New() error = %v, want nil", err)
	}
	service := lessonService(t, h, kept, &clock{at: lessonDay}, nil)
	h.start(t, mcpserver.BearerSignIn(readerEndingIn(time.Hour), metadata),
		slices.Concat(service.ProfileTools(), service.TaskTools())...)
	session, err := h.connectWith(t, vouchedToken)
	if err != nil {
		t.Fatalf("connectWith() error = %v, want nil", err)
	}
	return session, kept
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
// handed in, answered — is nine calls on a warm instance.
func TestEveryToolStaysWithinItsDriveBudget(t *testing.T) {
	t.Parallel()

	fake := drivetest.New(t)
	session, kept := instanceOverDrive(t, fake)
	read := drivetest.Calls{"download": 1}
	write := drivetest.Calls{"download": 2, "update": 1}

	// The first profile: the search that finds none, the search the creation
	// makes sure with, the search for the folder, the folder and the file.
	if got, want := costOf(t, fake, session, "save_profile", map[string]any{
		"pseudonym": "Otter", "grade": 2, "interests": []string{"sport"},
	}), (drivetest.Calls{"list": 3, "create": 2}); !maps.Equal(got, want) {
		t.Errorf("save_profile making the first profile cost %v, want %v", got, want)
	}

	for _, step := range []struct {
		tool string
		args any
		want drivetest.Calls
	}{
		{"get_profile", map[string]any{}, read},
		{"get_progress", map[string]any{}, read},
		{"read_progress", map[string]any{}, read},
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

	p, _, err := kept.Load(t.Context(), parent)
	if err != nil || p.OpenRequest == nil {
		t.Fatalf("Load() = %v, %v, want the request next_task opened", p, err)
	}
	if got := costOf(t, fake, session, "submit_task", raceOn(p.OpenRequest)); !maps.Equal(got, write) {
		t.Errorf("submit_task on a warm instance cost %v, want %v", got, write)
	}
	if p, _, err = kept.Load(t.Context(), parent); err != nil || p.CurrentTask == nil {
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

	// Another instance searches for the file once, and then knows it.
	cold, _ := instanceOverDrive(t, fake)
	if got, want := costOf(t, fake, cold, "get_profile", map[string]any{}), (drivetest.Calls{"list": 1, "download": 1}); !maps.Equal(got, want) {
		t.Errorf("get_profile on a cold instance cost %v, want %v", got, want)
	}
	colder, _ := instanceOverDrive(t, fake)
	if got, want := costOf(t, fake, colder, "save_profile", map[string]any{"interests": []string{"sport"}}),
		(drivetest.Calls{"list": 1, "download": 2, "update": 1}); !maps.Equal(got, want) {
		t.Errorf("save_profile on a cold instance cost %v, want %v", got, want)
	}
}

package mcpserver_test

import (
	"maps"
	"reflect"
	"testing"
	"time"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/profile"
	"github.com/MathTrail/mathtrail-standalone/internal/infra/drive/drivetest"
)

// A host asks a person before a call, or lets it through, by what the tool
// says it does. These cases hold what the tools say to what they do in the
// parent's Drive.

// changes are the kinds of call that change what Drive holds: a file made or
// rewritten, and a state of one kept forever or deleted.
var changes = []string{drivetest.Create, drivetest.Update, drivetest.KeepRevision, drivetest.DeleteRevision}

// keptAtTheMost is the most states of one file Drive keeps forever.
const keptAtTheMost = 200

// settled is a profile an adult set up the day before the lesson, with
// settings of their own in it.
func settled() *profile.Profile {
	return profile.New(profile.Student{
		Pseudonym: "Otter", Grade: 2, Interests: []string{"space"}, ExcludedSkills: []string{}, Notes: "Likes races.",
	}, "effect_test", lessonDay.Add(-24*time.Hour))
}

// withFullHistory puts a profile file in the Drive that holds the profile
// given, with as many earlier states kept forever as Drive keeps.
func withFullHistory(t *testing.T, fake *drivetest.Drive, p *profile.Profile) {
	t.Helper()

	raw, err := profile.Marshal(p)
	if err != nil {
		t.Fatalf("Marshal() error = %v, want nil", err)
	}
	history := make([]drivetest.Revision, 0, keptAtTheMost+1)
	for range keptAtTheMost {
		history = append(history, drivetest.Revision{Content: raw, KeepForever: true})
	}
	fake.Put(googleToken, &drivetest.File{
		Name: "mathtrail-profile.json", MimeType: "application/json", AppProperties: maps.Clone(profileMarker),
		Revisions: append(history, drivetest.Revision{Content: raw}),
	})
}

// profileFileIn is the profile's file as the Drive holds it.
func profileFileIn(t *testing.T, fake *drivetest.Drive) *drivetest.File {
	t.Helper()

	for _, file := range fake.Files(googleToken) {
		if file.AppProperties["mathtrail"] == "profile" {
			return file
		}
	}
	t.Fatal("the Drive holds no profile file")
	return nil
}

// keptForeverIn is how many states of a file Drive keeps forever.
func keptForeverIn(file *drivetest.File) int {
	kept := 0
	for _, revision := range file.Revisions {
		if revision.KeepForever {
			kept++
		}
	}
	return kept
}

// setUp makes the profile as an adult first sets it up.
func setUp(t *testing.T, d *overDrive) {
	t.Helper()

	made := call(t, d.session, "save_profile", map[string]any{"pseudonym": "Otter", "grade": 2, "interests": []string{"sport"}})
	if made.IsError {
		t.Fatalf("save_profile failed: %s", textOf(t, made))
	}
}

// raceAskedFor asks for the race, and is the request it opened.
func raceAskedFor(t *testing.T, d *overDrive) *profile.OpenRequest {
	t.Helper()

	if asked := call(t, d.session, "next_task", raceChoice); asked.IsError {
		t.Fatalf("next_task failed: %s", textOf(t, asked))
	}
	p, _, err := d.kept.Load(t.Context(), parent)
	if err != nil || p.OpenRequest == nil {
		t.Fatalf("Load() = %v, %v, want the request next_task opened", p, err)
	}
	return p.OpenRequest
}

// raceOnCard asks for the race and hands it in, and is the id of the task on
// the card.
func raceOnCard(t *testing.T, d *overDrive) string {
	t.Helper()

	if handed := call(t, d.session, "submit_task", raceOn(raceAskedFor(t, d))); handed.IsError {
		t.Fatalf("submit_task failed: %s", textOf(t, handed))
	}
	p, _, err := d.kept.Load(t.Context(), parent)
	if err != nil || p.CurrentTask == nil {
		t.Fatalf("Load() = %v, %v, want the task submit_task accepted", p, err)
	}
	return p.CurrentTask.ID
}

// readsChangeNothing calls every tool that says it only reads, about the
// request given and the task on the card, and holds each to having asked
// Drive for no change.
func readsChangeNothing(t *testing.T, fake *drivetest.Drive, d *overDrive, request string) {
	t.Helper()

	for _, read := range []struct {
		tool string
		args map[string]any
	}{
		{"get_profile", map[string]any{}},
		{"get_progress", map[string]any{}},
		{"read_progress", map[string]any{}},
		{"get_package", map[string]any{"request_id": request}},
		{"read_task", map[string]any{"request_id": request}},
		{"show_result", map[string]any{"task_id": taskOnTheCard(t, d)}},
	} {
		fake.ResetCalls()
		call(t, d.session, read.tool, read.args)
		for _, kind := range changes {
			if got := fake.Calls()[kind]; got != 0 {
				t.Errorf("%s made %d calls of kind %q, want none: it only reads", read.tool, got, kind)
			}
		}
	}
}

// taskOnTheCard is the id of the task on the child's card as the store holds
// it, or one no card holds when there is none, or no profile.
func taskOnTheCard(t *testing.T, d *overDrive) string {
	t.Helper()

	p, _, err := d.kept.Load(t.Context(), parent)
	if err != nil || p.CurrentTask == nil {
		return "tsk_none"
	}
	return p.CurrentTask.ID
}

// A tool that says it only reads changes nothing in the parent's Drive,
// whatever it finds there — on the instance that wrote the file and on one
// that has not seen it yet. A damaged file is told, never put back, and a
// history Drive keeps no more of is left as it is.
func TestAToolThatReadsChangesNothing(t *testing.T) {
	t.Parallel()

	for _, state := range []struct {
		name string
		// make leaves the Drive in the state, and is the request open in it.
		make func(t *testing.T, fake *drivetest.Drive, d *overDrive) (request string)
	}{
		{"no profile yet", func(*testing.T, *drivetest.Drive, *overDrive) string { return "req_none" }},
		{"a profile and no lesson yet", func(t *testing.T, _ *drivetest.Drive, d *overDrive) string {
			setUp(t, d)
			return "req_none"
		}},
		{"a task asked for", func(t *testing.T, _ *drivetest.Drive, d *overDrive) string {
			setUp(t, d)
			return raceAskedFor(t, d).ID
		}},
		{"a task on the card", func(t *testing.T, _ *drivetest.Drive, d *overDrive) string {
			setUp(t, d)
			raceOnCard(t, d)
			return "req_none"
		}},
		{"a task answered on the card", func(t *testing.T, _ *drivetest.Drive, d *overDrive) string {
			setUp(t, d)
			answerIt(t, d.session, raceOnCard(t, d), "A", false)
			return "req_none"
		}},
		{"a profile in the bin", func(t *testing.T, fake *drivetest.Drive, d *overDrive) string {
			setUp(t, d)
			fake.Edit(googleToken, profileFileIn(t, fake).ID, func(binned *drivetest.File) { binned.Trashed = true })
			return "req_none"
		}},
		{"a damaged file with a state before the damage", func(t *testing.T, fake *drivetest.Drive, _ *overDrive) string {
			fileWithHistory(t, fake, settled())
			return "req_none"
		}},
		{"a history Drive keeps no more of", func(t *testing.T, fake *drivetest.Drive, _ *overDrive) string {
			withFullHistory(t, fake, settled())
			return "req_none"
		}},
	} {
		t.Run(state.name, func(t *testing.T) {
			t.Parallel()

			fake := drivetest.New(t)
			warm := instanceOverDrive(t, fake)
			request := state.make(t, fake, warm)
			before := fake.Files(googleToken)

			readsChangeNothing(t, fake, warm, request)
			readsChangeNothing(t, fake, instanceOverDrive(t, fake), request)
			if after := fake.Files(googleToken); !reflect.DeepEqual(after, before) {
				t.Errorf("the Drive holds %d files after the reads, changed from the %d before them, want it as it was",
					len(after), len(before))
			}
		})
	}
}

// A tool that only adds to the record of the lessons takes nothing from it:
// a task answered, or left on the card and skipped by the next one asked for,
// the next one got ready ahead, or taken by the card, leaves the adult's
// settings as they were set, and deletes no state Drive
// keeps of the file — not even once Drive keeps the most it will, when the
// day's first write lands without keeping its own.
func TestAToolThatAddsTakesNothingAway(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name string
		// last is the call that ends the lesson, given the task on the card.
		last func(taskID string) (tool string, args map[string]any)
		// added says whether the profile records what the last call added.
		added func(p *profile.Profile) bool
	}{
		{
			name: "a task answered",
			last: func(taskID string) (string, map[string]any) {
				return "submit_answer", map[string]any{"task_id": taskID, "answer": "C"}
			},
			added: func(p *profile.Profile) bool { return p.Ratings.Answers == 1 },
		},
		{
			name: "a task skipped",
			last: func(string) (string, map[string]any) { return "next_task", raceChoice },
			added: func(p *profile.Profile) bool {
				return p.Topics["logic.ordering"].Skipped == 1 && p.CurrentTask == nil
			},
		},
		{
			name:  "the next task got ready ahead",
			last:  func(string) (string, map[string]any) { return "prepare_task", aheadChoice },
			added: func(p *profile.Profile) bool { return p.OpenRequest != nil && p.OpenRequest.Ahead },
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			fake := drivetest.New(t)
			set := settled()
			withFullHistory(t, fake, set)
			d := instanceOverDrive(t, fake)

			tool, args := tc.last(raceOnCard(t, d))
			if ended := call(t, d.session, tool, args); ended.IsError {
				t.Fatalf("%s failed: %s", tool, textOf(t, ended))
			}

			p := historyKept(t, fake, d)
			if !tc.added(p) || !reflect.DeepEqual(p.Student, set.Student) {
				t.Errorf("after %s: what it adds recorded: %v, the settings %+v; want it recorded and the settings %+v",
					tool, tc.added(p), p.Student, set.Student)
			}
		})
	}
}

// historyKept holds the lesson to having deleted no state of the file, and to
// Drive keeping forever as many as it kept before, and is the profile the
// file holds now.
func historyKept(t *testing.T, fake *drivetest.Drive, d *overDrive) *profile.Profile {
	t.Helper()

	if got := fake.Calls()[drivetest.DeleteRevision]; got != 0 {
		t.Errorf("the lesson deleted %d states of the file, want none", got)
	}
	if kept := keptForeverIn(profileFileIn(t, fake)); kept != keptAtTheMost {
		t.Errorf("Drive keeps %d states of the file forever, want the %d kept before the lesson", kept, keptAtTheMost)
	}
	p, _, err := d.kept.Load(t.Context(), parent)
	if err != nil {
		t.Fatalf("Load() error = %v, want nil", err)
	}
	return p
}

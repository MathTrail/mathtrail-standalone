package profile_test

import (
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/profile"
)

// Every limit of the file, one case each, and every refusal has to name the
// field it is about: a profile is written by a parent through a screen, and
// "invalid profile" leaves them nothing to fix.
func TestEveryLimitIsRefusedByName(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		breakIt func(p *profile.Profile)
		wantSay string
	}{
		{
			name:    "no pseudonym",
			breakIt: func(p *profile.Profile) { p.Student.Pseudonym = "" },
			wantSay: "student.pseudonym",
		},
		{
			name:    "a pseudonym past the limit",
			breakIt: func(p *profile.Profile) { p.Student.Pseudonym = strings.Repeat("o", profile.MaxPseudonym+1) },
			wantSay: "student.pseudonym",
		},
		{
			name:    "a pseudonym with a control character",
			breakIt: func(p *profile.Profile) { p.Student.Pseudonym = "Ot\x07ter" },
			wantSay: "student.pseudonym",
		},
		{
			name:    "a grade nobody is in",
			breakIt: func(p *profile.Profile) { p.Student.Grade = profile.MaxGrade + 1 },
			wantSay: "student.grade",
		},
		{
			name: "too many interests",
			breakIt: func(p *profile.Profile) {
				p.Student.Interests = make([]string, profile.MaxInterests+1)
			},
			wantSay: "student.interests",
		},
		{
			name: "an interest past the limit",
			breakIt: func(p *profile.Profile) {
				p.Student.Interests = []string{strings.Repeat("a", profile.MaxInterest+1)}
			},
			wantSay: "student.interests",
		},
		{
			name: "too many excluded skills",
			breakIt: func(p *profile.Profile) {
				p.Student.ExcludedSkills = make([]string, profile.MaxExcludedSkills+1)
			},
			wantSay: "student.excluded_skills",
		},
		{
			name:    "notes past the limit",
			breakIt: func(p *profile.Profile) { p.Student.Notes = strings.Repeat("n", profile.MaxNotes+1) },
			wantSay: "student.notes",
		},
		{
			name: "a language that is an empty string rather than none",
			breakIt: func(p *profile.Profile) {
				empty := ""
				p.Student.UILanguage = &empty
			},
			wantSay: "student.ui_language",
		},
		{
			name:    "a level from another shape of the file",
			breakIt: func(p *profile.Profile) { p.SchemaVersion = profile.Version + 1 },
			wantSay: "schema_version",
		},
		{
			name:    "no identifier",
			breakIt: func(p *profile.Profile) { p.StudentID = "" },
			wantSay: "student_id",
		},
		{
			name:    "a revision before the first",
			breakIt: func(p *profile.Profile) { p.Revision = 0 },
			wantSay: "revision",
		},
		{
			name:    "a negative count of answers",
			breakIt: func(p *profile.Profile) { p.Ratings.Answers = -1 },
			wantSay: "ratings",
		},
		{
			name: "a topic with more correct answers than answers",
			breakIt: func(p *profile.Profile) {
				topic := p.Topics["counting.gaps"]
				topic.Correct = topic.Answers + 1
				p.Topics["counting.gaps"] = topic
			},
			wantSay: "topics[counting.gaps]",
		},
		{
			// Every answer in the window is a good one, so nothing but the
			// size of the window itself can be what is refused.
			name: "a window past its size",
			breakIt: func(p *profile.Profile) {
				p.Recent = make([]profile.Answer, 0, profile.MaxRecent+1)
				for i := range profile.MaxRecent + 1 {
					answer := parseFixture(t, "dima").Recent[0]
					answer.TaskID = fmt.Sprintf("tsk_%02d", i)
					p.Recent = append(p.Recent, answer)
				}
			},
			wantSay: "the window is",
		},
		{
			name:    "an answer with a pace nobody measures",
			breakIt: func(p *profile.Profile) { p.Recent[0].Pace = "quick" },
			wantSay: "recent[0].pace",
		},
		{
			name:    "an answer with a level no task has",
			breakIt: func(p *profile.Profile) { p.Recent[0].Difficulty = 9 },
			wantSay: "recent[0].difficulty",
		},
		{
			name:    "an answer that picked an option nobody offered",
			breakIt: func(p *profile.Profile) { p.Recent[1].Chosen = "F" },
			wantSay: "recent[1].chosen",
		},
		{
			// A right answer has no mistake to name, and naming one would put
			// a trap of the lesson in flight where it does not belong.
			name:    "a correct answer that still names a trap",
			breakIt: func(p *profile.Profile) { p.Recent[0].Trap = "off_by_one" },
			wantSay: "recent[0]",
		},
		{
			name: "more fingerprints than are kept",
			breakIt: func(p *profile.Profile) {
				p.TaskFingerprints = make([]string, profile.MaxFingerprints+1)
			},
			wantSay: "task_fingerprints",
		},
		{
			name:    "a day the counters do not belong to",
			breakIt: func(p *profile.Profile) { p.Daily.Date = profile.Date{} },
			wantSay: "daily",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			p := parseFixture(t, "dima")
			tc.breakIt(p)

			err := p.Validate()
			if !errors.Is(err, profile.ErrInvalid) {
				t.Fatalf("Validate() error = %v, want %v", err, profile.ErrInvalid)
			}
			if !strings.Contains(err.Error(), tc.wantSay) {
				t.Errorf("Validate() said %q, want it to name %s", err, tc.wantSay)
			}
			// And nothing this wrong can be written into the parent's Drive.
			if _, err := profile.Marshal(p); !errors.Is(err, profile.ErrInvalid) {
				t.Errorf("Marshal() error = %v, want it to refuse what Validate refuses", err)
			}
		})
	}
}

// A task in flight has to be whole: five options, a wording, and something
// sealed. A task with nothing sealed is a task with its answer in the open.
func TestATaskInFlightIsRefusedWhenItIsNotWhole(t *testing.T) {
	t.Parallel()

	cases := map[string]func(t *profile.CurrentTask){
		"current_task.sealed": func(t *profile.CurrentTask) { t.Sealed = "" },
		"current_task needs":  func(t *profile.CurrentTask) { t.Wording = "" },
		"current_task offers": func(t *profile.CurrentTask) { delete(t.Options, "E") },
		"issued_at":           func(t *profile.CurrentTask) { t.IssuedAt = profile.Time{} },
		"difficulty":          func(t *profile.CurrentTask) { t.Difficulty = 0 },
	}

	for wantSay, breakIt := range cases {
		t.Run(wantSay, func(t *testing.T) {
			t.Parallel()

			p := parseFixture(t, "masha")
			if p.CurrentTask == nil {
				t.Fatal("the fixture carries no task in flight")
			}
			breakIt(p.CurrentTask)

			err := p.Validate()
			if !errors.Is(err, profile.ErrInvalid) {
				t.Fatalf("Validate() error = %v, want %v", err, profile.ErrInvalid)
			}
			if !strings.Contains(err.Error(), wantSay) {
				t.Errorf("Validate() said %q, want it to name %s", err, wantSay)
			}
		})
	}
}

// A request being written is refused on the same terms.
func TestAnOpenRequestIsRefusedWhenItIsNotWhole(t *testing.T) {
	t.Parallel()

	whole := profile.OpenRequest{
		Attempts:  1,
		ID:        "req_1",
		Language:  "en-US",
		OpenedAt:  profile.At(time.Date(2026, 9, 4, 10, 0, 0, 0, time.UTC)),
		TutorMode: profile.TutorRule,
		Brief: profile.Brief{
			Difficulty:      3,
			PedagogicalGoal: profile.GoalNewTopic,
			TargetConcept:   "counting.gaps",
		},
	}

	p := parseFixture(t, "dima")
	p.OpenRequest = &whole
	if err := p.Validate(); err != nil {
		t.Fatalf("Validate() error = %v, want a whole request to pass", err)
	}

	cases := map[string]func(r *profile.OpenRequest){
		"open_request has no id":              func(r *profile.OpenRequest) { r.ID = "" },
		"open_request.attempts":               func(r *profile.OpenRequest) { r.Attempts = profile.MaxAttempts + 1 },
		"open_request.tutor_mode":             func(r *profile.OpenRequest) { r.TutorMode = "guesswork" },
		"open_request.brief names no topic":   func(r *profile.OpenRequest) { r.Brief.TargetConcept = "" },
		"open_request.brief.pedagogical_goal": func(r *profile.OpenRequest) { r.Brief.PedagogicalGoal = "whatever" },
		"open_request has no opened_at":       func(r *profile.OpenRequest) { r.OpenedAt = profile.Time{} },
	}

	for wantSay, breakIt := range cases {
		t.Run(wantSay, func(t *testing.T) {
			t.Parallel()

			p := parseFixture(t, "dima")
			request := whole
			breakIt(&request)
			p.OpenRequest = &request

			err := p.Validate()
			if !errors.Is(err, profile.ErrInvalid) {
				t.Fatalf("Validate() error = %v, want %v", err, profile.ErrInvalid)
			}
			if !strings.Contains(err.Error(), wantSay) {
				t.Errorf("Validate() said %q, want it to name %s", err, wantSay)
			}
		})
	}
}

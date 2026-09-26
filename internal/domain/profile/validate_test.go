package profile_test

import (
	"errors"
	"fmt"
	"math"
	"strings"
	"testing"
	"time"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/profile"
	"github.com/MathTrail/mathtrail-standalone/internal/domain/rating"
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
			name:    "an answer that picked two options at once",
			breakIt: func(p *profile.Profile) { p.Recent[1].Chosen = "BC" },
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
		{
			name:    "no moment the profile was made at",
			breakIt: func(p *profile.Profile) { p.CreatedAt = profile.Time{} },
			wantSay: "created_at",
		},
		{
			name:    "a level that is not a number",
			breakIt: func(p *profile.Profile) { p.Ratings.Theta = math.NaN() },
			wantSay: "ratings.theta",
		},
		{
			name: "a topic level that is not a number",
			breakIt: func(p *profile.Profile) {
				topic := p.Topics["counting.gaps"]
				topic.Delta = math.Inf(1)
				p.Topics["counting.gaps"] = topic
			},
			wantSay: "topics[counting.gaps].delta",
		},
		{
			name: "a topic with no id",
			breakIt: func(p *profile.Profile) {
				p.Topics[""] = profile.Topic{Traps: map[string]int{}}
			},
			wantSay: "topics has an entry with no id",
		},
		{
			name: "a topic with a negative count",
			breakIt: func(p *profile.Profile) {
				topic := p.Topics["counting.gaps"]
				topic.WrongStreak = -1
				p.Topics["counting.gaps"] = topic
			},
			wantSay: "topics[counting.gaps] counts answers",
		},
		{
			name: "a trap counted no times at all",
			breakIt: func(p *profile.Profile) {
				topic := p.Topics["counting.gaps"]
				topic.Traps["never_happened"] = 0
				p.Topics["counting.gaps"] = topic
			},
			wantSay: "topics[counting.gaps].traps",
		},
		{
			name: "a topic mastered on a day and at no level",
			breakIt: func(p *profile.Profile) {
				topic := p.Topics["counting.gaps"]
				since := profile.DateOf(time.Date(2026, 9, 2, 0, 0, 0, 0, time.UTC))
				topic.MasteredSince = &since
				p.Topics["counting.gaps"] = topic
			},
			wantSay: "topics[counting.gaps] has mastered_since and mastered_level apart",
		},
		{
			name: "a topic mastered at a level and on no day",
			breakIt: func(p *profile.Profile) {
				topic := p.Topics["counting.gaps"]
				level := rating.Grades12
				topic.MasteredLevel = &level
				p.Topics["counting.gaps"] = topic
			},
			wantSay: "topics[counting.gaps] has mastered_since and mastered_level apart",
		},
		{
			name: "a topic mastered at a level there is not",
			breakIt: func(p *profile.Profile) {
				topic := p.Topics["counting.gaps"]
				since, level := profile.DateOf(time.Date(2026, 9, 2, 0, 0, 0, 0, time.UTC)), rating.GradeLevel("7-8")
				topic.MasteredSince, topic.MasteredLevel = &since, &level
				p.Topics["counting.gaps"] = topic
			},
			wantSay: "topics[counting.gaps].mastered_level",
		},
		{
			name:    "a start that is no number",
			breakIt: func(p *profile.Profile) { p.Ratings.Start = math.Inf(1) },
			wantSay: "ratings.start",
		},
		{
			name:    "an answer that names no task",
			breakIt: func(p *profile.Profile) { p.Recent[0].TaskID = "" },
			wantSay: "recent[0] names no task",
		},
		{
			name:    "an answer at no level",
			breakIt: func(p *profile.Profile) { p.Recent[0].GradeLevel = "" },
			wantSay: "recent[0].grade_level",
		},
		{
			name:    "an answer with no moment",
			breakIt: func(p *profile.Profile) { p.Recent[0].AnsweredAt = profile.Time{} },
			wantSay: "recent[0] has no answered_at",
		},
		{
			name:    "a daily counter below zero",
			breakIt: func(p *profile.Profile) { p.Daily.Accepted = -1 },
			wantSay: "daily counts tasks",
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
		"grade_level":         func(t *profile.CurrentTask) { t.GradeLevel = "5-7" },
		// Five options, and one of them says nothing: a child cannot pick it,
		// and the count alone would not notice.
		"options shows nothing": func(t *profile.CurrentTask) { t.Options["C"] = "" },
		// Nor can the child pick one that shows nothing on the card, however
		// many characters it is made of.
		"options shows nothing under \"D\"": func(t *profile.CurrentTask) { t.Options["D"] = " \u200b " },
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
			GradeLevel:      rating.Grades12,
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
		"open_request.brief.difficulty":       func(r *profile.OpenRequest) { r.Brief.Difficulty = 0 },
		"open_request.brief.grade_level":      func(r *profile.OpenRequest) { r.Brief.GradeLevel = "" },
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

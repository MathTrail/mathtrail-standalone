package profile_test

import (
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/profile"
)

// skills stands in for the catalog of skills: a few ids it has, and none else.
func skills(id string) bool {
	return slices.Contains([]string{"fractions", "division_with_remainder", "negative_numbers"}, id)
}

func text(s string) *string { return &s }
func number(n int) *int     { return &n }

var editedAt = time.Date(2026, 9, 27, 10, 0, 0, 0, time.UTC)

// fields names the fields of a set of problems, in the order they came.
func fields(problems []profile.Problem) []string {
	names := make([]string, 0, len(problems))
	for _, problem := range problems {
		names = append(names, problem.Field)
	}
	return names
}

// Every field is taken as the parent meant it: what nobody can see is
// dropped, a line break reads as a space, and a list given twice the same
// entry keeps it once.
func TestAnEditIsTakenAsTheParentMeantIt(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name  string
		edit  profile.Edit
		check func(t *testing.T, s *profile.Student)
	}{
		{"a pseudonym with spaces and a bell", profile.Edit{Pseudonym: text("  Com\aet ")},
			func(t *testing.T, s *profile.Student) { wantText(t, "pseudonym", s.Pseudonym, "Comet") }},
		{"notes over several lines", profile.Edit{Notes: text("Reads slowly.\nLikes cats.\t")},
			func(t *testing.T, s *profile.Student) { wantText(t, "notes", s.Notes, "Reads slowly. Likes cats.") }},
		{"notes cleared", profile.Edit{Notes: text("")},
			func(t *testing.T, s *profile.Student) { wantText(t, "notes", s.Notes, "") }},
		{"a pseudonym shown backwards", profile.Edit{Pseudonym: text("\u202eComet\u202c")},
			func(t *testing.T, s *profile.Student) { wantText(t, "pseudonym", s.Pseudonym, "Comet") }},
		{"a pseudonym with a zero-width space", profile.Edit{Pseudonym: text("Com\u200bet")},
			func(t *testing.T, s *profile.Student) { wantText(t, "pseudonym", s.Pseudonym, "Comet") }},
		{"notes with a sentence spelt in tags", profile.Edit{Notes: text("hi\U000E0049\U000E0047\U000E004E there")},
			func(t *testing.T, s *profile.Student) { wantText(t, "notes", s.Notes, "hi there") }},
		{"an interest with a variant of its glyph", profile.Edit{Interests: []string{"\u2764\ufe0f"}},
			func(t *testing.T, s *profile.Student) { wantList(t, "interests", s.Interests, "\u2764") }},
		{"notes with an emoji of joined parts", profile.Edit{Notes: text("\U0001F469\u200d\U0001F4BB")},
			func(t *testing.T, s *profile.Student) { wantText(t, "notes", s.Notes, "\U0001F469\u200d\U0001F4BB") }},
		{"an interest behind a byte-order mark", profile.Edit{Interests: []string{"\ufeffspace"}},
			func(t *testing.T, s *profile.Student) { wantList(t, "interests", s.Interests, "space") }},
		{"notes with a mark of direction", profile.Edit{Notes: text("Calm\u200e.")},
			func(t *testing.T, s *profile.Student) { wantText(t, "notes", s.Notes, "Calm.") }},
		{"a skill with spaces around it", profile.Edit{ExcludedSkills: []string{" fractions\n"}},
			func(t *testing.T, s *profile.Student) { wantList(t, "excluded_skills", s.ExcludedSkills, "fractions") }},
		{"notes in a script written with joiners", profile.Edit{Notes: text("می\u200cخواهد")},
			func(t *testing.T, s *profile.Student) { wantText(t, "notes", s.Notes, "می\u200cخواهد") }},
		{"interests given twice", profile.Edit{Interests: []string{" space", "cats", "space "}},
			func(t *testing.T, s *profile.Student) { wantList(t, "interests", s.Interests, "space", "cats") }},
		{"interests cleared", profile.Edit{Interests: []string{}},
			func(t *testing.T, s *profile.Student) { wantList(t, "interests", s.Interests) }},
		{"a skill given twice", profile.Edit{ExcludedSkills: []string{"fractions", "fractions"}},
			func(t *testing.T, s *profile.Student) { wantList(t, "excluded_skills", s.ExcludedSkills, "fractions") }},
		{"a language in lower case", profile.Edit{UILanguage: text("pt-br")},
			func(t *testing.T, s *profile.Student) { wantLanguage(t, s.UILanguage, "pt-BR") }},
		{"a language cleared", profile.Edit{UILanguage: text("")},
			func(t *testing.T, s *profile.Student) { wantNoLanguage(t, s.UILanguage) }},
		{"a language written with an underscore", profile.Edit{UILanguage: text("en_US")},
			func(t *testing.T, s *profile.Student) { wantLanguage(t, s.UILanguage, "en-US") }},
		{"notes in paragraphs", profile.Edit{Notes: text("Calm.\u2029Kind.\u00a0Slow.")},
			func(t *testing.T, s *profile.Student) { wantText(t, "notes", s.Notes, "Calm. Kind. Slow.") }},
		{"a new grade", profile.Edit{Grade: number(4)},
			func(t *testing.T, s *profile.Student) {
				if s.Grade != 4 {
					t.Errorf("grade = %d, want 4", s.Grade)
				}
			}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			p := profile.New(newStudent(), "1.2.3", editedAt)
			p.Student.UILanguage = text("ru")
			if _, problems := p.Change(&tc.edit, skills, "1.2.4", editedAt); len(problems) > 0 {
				t.Fatalf("Change() problems = %v, want none", problems)
			}
			tc.check(t, &p.Student)
		})
	}
}

// A field that breaks a rule is refused by name, with the rule, and the rule
// never repeats what the field held: those are the parent's words, and a
// refusal travels to the model, to its card and to wherever it is written.
func TestARefusalNamesTheFieldAndNeverRepeatsIt(t *testing.T) {
	t.Parallel()

	const said = "Mariya-Ivanova-from-School-Number-Five"
	cases := []struct {
		name  string
		edit  profile.Edit
		field string
	}{
		{"a pseudonym past the limit", profile.Edit{Pseudonym: text(said)}, "pseudonym"},
		{"a pseudonym of nothing but spaces", profile.Edit{Pseudonym: text("   ")}, "pseudonym"},
		{"a pseudonym nobody can see", profile.Edit{Pseudonym: text("\u200b\u2060")}, "pseudonym"},
		{"a pseudonym of a filler", profile.Edit{Pseudonym: text("\u3164")}, "pseudonym"},
		{"a grade nobody is in", profile.Edit{Grade: number(7)}, "grade"},
		{"an interest past the limit", profile.Edit{Interests: []string{said + said}}, "interests"},
		{"an interest that is empty", profile.Edit{Interests: []string{"\t"}}, "interests"},
		{"too many interests", profile.Edit{Interests: count(profile.MaxInterests + 1)}, "interests"},
		{"a skill the catalog does not have", profile.Edit{ExcludedSkills: []string{said}}, "excluded_skills"},
		{"too many skills, not all of the catalog", profile.Edit{ExcludedSkills: append(count(profile.MaxExcludedSkills), said)},
			"excluded_skills"},
		{"notes past the limit", profile.Edit{Notes: text(strings.Repeat(said, 20))}, "notes"},
		{"a language that is none", profile.Edit{UILanguage: text(said)}, "ui_language"},
		{"a language nobody speaks", profile.Edit{UILanguage: text("und")}, "ui_language"},
		{"a language only guessed from a country", profile.Edit{UILanguage: text("und-US")}, "ui_language"},
		{"a language of nobody's", profile.Edit{UILanguage: text("x-klingon")}, "ui_language"},
		{"no language at all", profile.Edit{UILanguage: text("zxx")}, "ui_language"},
		{"several languages", profile.Edit{UILanguage: text("mul")}, "ui_language"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			p := profile.New(newStudent(), "1.2.3", editedAt)
			before := p.Student
			changed, problems := p.Change(&tc.edit, skills, "1.2.4", editedAt)
			if changed || !slices.Equal(fields(problems), []string{tc.field}) {
				t.Fatalf("Change() = %v, %v; want nothing changed and one problem with %s", changed, problems, tc.field)
			}
			if strings.Contains(problems[0].Rule, said) || strings.Contains(problems[0].String(), said) {
				t.Errorf("the problem %q repeats what the field held", problems[0])
			}
			if p.Revision != 1 || p.Student.Pseudonym != before.Pseudonym || p.Student.Grade != before.Grade {
				t.Errorf("a refused edit changed the profile: revision %d, student %+v", p.Revision, p.Student)
			}
		})
	}
}

// A rule says how far off the field is, as a count: a count is not what the
// parent wrote, and it tells the model how much to take away.
func TestARuleSaysHowFarOffTheFieldIs(t *testing.T) {
	t.Parallel()

	anySkill := func(string) bool { return true }
	for _, tc := range []struct {
		edit  profile.Edit
		known func(string) bool
		want  string
	}{
		{profile.Edit{Pseudonym: text(strings.Repeat("o", profile.MaxPseudonym+6))}, skills, "at most 32 characters, not 38"},
		{profile.Edit{Interests: count(profile.MaxInterests + 1)}, skills, "at most 10 interests, not 11"},
		{profile.Edit{ExcludedSkills: count(profile.MaxExcludedSkills + 2)}, anySkill, "at most 25 skills, not 27"},
		{profile.Edit{Notes: text(strings.Repeat("n", profile.MaxNotes+3))}, skills, "at most 500 characters, not 503"},
		{profile.Edit{Interests: []string{strings.Repeat("i", profile.MaxInterest+2)}}, skills, "1 to 40 characters; one is 42"},
		{profile.Edit{UILanguage: text("en-x-" + strings.Repeat("abcdefgh-", 4) + "abc")}, skills, "at most 35 characters, not 44"},
	} {
		p := profile.New(newStudent(), "1.2.3", editedAt)
		if _, problems := p.Change(&tc.edit, tc.known, "1.2.4", editedAt); len(problems) != 1 ||
			!strings.Contains(problems[0].Rule, tc.want) {
			t.Errorf("problems = %v, want one saying %q", problems, tc.want)
		}
	}
}

// A skill the catalog does not have is named by its place in the list: a model
// resending a list with a skill the catalog has since dropped learns which one
// to leave out, and nothing it sent is repeated back.
func TestASkillOutsideTheCatalogIsNamedByItsPlace(t *testing.T) {
	t.Parallel()

	p := profile.New(newStudent(), "1.2.3", editedAt)
	edit := profile.Edit{ExcludedSkills: []string{"fractions", "fractions", "long_division_by_hand"}}
	_, problems := p.Change(&edit, skills, "1.2.4", editedAt)
	if len(problems) != 1 || !strings.Contains(problems[0].Rule, "entry 3") ||
		strings.Contains(problems[0].Rule, "long_division") {
		t.Errorf("problems = %v, want the third entry, as it was sent, named by its place alone", problems)
	}
}

// A list is counted once its repeats are merged: every skill of the catalog
// with one of them given twice still fits, and so does one interest given more
// times than the cap.
func TestAListThatFitsOnceItsRepeatsAreMergedIsKept(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name  string
		edit  profile.Edit
		field string
	}{
		{"one skill given more times than the cap", profile.Edit{
			ExcludedSkills: slices.Repeat([]string{"fractions"}, profile.MaxExcludedSkills+1),
		}, "excluded_skills"},
		{"one interest given more times than the cap", profile.Edit{
			Interests: slices.Repeat([]string{"space"}, profile.MaxInterests+1),
		}, "interests"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			p := profile.New(newStudent(), "1.2.3", editedAt)
			if _, problems := p.Change(&tc.edit, skills, "1.2.4", editedAt); len(problems) > 0 {
				t.Errorf("problems = %v, want %s kept once its repeats were merged", problems, tc.field)
			}
		})
	}
}

// The cap on the notes is on what is kept, not on what was sent: a note of the
// full length with a line break in it fits, and one character more does not.
func TestTheNotesAreMeasuredAfterTheyAreCleaned(t *testing.T) {
	t.Parallel()

	full := strings.Repeat("n", profile.MaxNotes-1)
	for _, tc := range []struct {
		name  string
		notes string
		fits  bool
	}{
		{"the full length with a bell in it", "\a" + full + "n", true},
		{"the full length with a line break in it", full[:10] + "\n" + full[10:], true},
		{"one past the full length", full + "nn", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			p := profile.New(newStudent(), "1.2.3", editedAt)
			_, problems := p.Change(&profile.Edit{Notes: text(tc.notes)}, skills, "1.2.4", editedAt)
			if fits := len(problems) == 0; fits != tc.fits {
				t.Errorf("fits = %v (problems %v), want %v", fits, problems, tc.fits)
			}
		})
	}
}

// A change is a write, and only a change is: an edit that asks for what is
// already there leaves the profile and its revision alone.
func TestOnlyAChangeTouchesTheProfile(t *testing.T) {
	t.Parallel()

	p := profile.New(newStudent(), "1.2.3", editedAt)
	later := editedAt.Add(time.Hour)

	same := profile.Edit{Pseudonym: text(" Otter "), Interests: []string{"space", "cats"}}
	if changed, problems := p.Change(&same, skills, "1.2.4", later); changed || len(problems) > 0 {
		t.Fatalf("Change(the same) = %v, %v; want nothing changed", changed, problems)
	}
	if p.Revision != 1 || p.AppVersion != "1.2.3" || !p.UpdatedAt.Equal(editedAt) {
		t.Errorf("an edit that changed nothing touched the profile: revision %d, %s at %v",
			p.Revision, p.AppVersion, p.UpdatedAt)
	}

	if changed, _ := p.Change(&profile.Edit{Notes: text("Counts on fingers.")}, skills, "1.2.4", later); !changed {
		t.Fatal("Change(new notes) changed nothing, want the notes changed")
	}
	if p.Revision != 2 || p.AppVersion != "1.2.4" || !p.UpdatedAt.Equal(later) {
		t.Errorf("a change was not a touch: revision %d, %s at %v", p.Revision, p.AppVersion, p.UpdatedAt)
	}
}

// The grade decides where a child starts, once. Changed later it is a label:
// the start, the level and the trial series stay where the answers put them.
func TestALaterGradeIsALabel(t *testing.T) {
	t.Parallel()

	p := parseFixture(t, "masha")
	ratings := p.Ratings
	if changed, problems := p.Change(&profile.Edit{Grade: number(6)}, skills, "1.2.4", editedAt); !changed ||
		len(problems) > 0 {
		t.Fatalf("Change(grade 6) = %v, %v; want the grade changed", changed, problems)
	}
	if p.Student.Grade != 6 || p.Ratings != ratings {
		t.Errorf("grade %d with ratings %+v, want grade 6 with the ratings as they were, %+v",
			p.Student.Grade, p.Ratings, ratings)
	}
}

// A profile needs a pseudonym and a grade to be made at all, and says which is
// missing; given both, it starts where the grade puts it.
func TestANewStudentNeedsAPseudonymAndAGrade(t *testing.T) {
	t.Parallel()

	if _, problems := profile.NewStudent(&profile.Edit{Interests: []string{"space"}}, skills); !slices.Equal(
		fields(problems), []string{"grade", "pseudonym"}) {
		t.Errorf("NewStudent(interests only) problems = %v, want grade and pseudonym", problems)
	}

	student, problems := profile.NewStudent(&profile.Edit{Pseudonym: text("Comet"), Grade: number(3)}, skills)
	if len(problems) > 0 {
		t.Fatalf("NewStudent(pseudonym and grade) problems = %v, want none", problems)
	}
	p := profile.New(student, "1.2.3", editedAt)
	if err := p.Validate(); err != nil {
		t.Fatalf("Validate() error = %v, want the new profile usable", err)
	}
	if p.Ratings.Start != 2.5 || p.Ratings.Theta != 2.5 {
		t.Errorf("a child of grade 3 starts at %v with θ %v, want 2.5 for both", p.Ratings.Start, p.Ratings.Theta)
	}
}

// A profile says "none" with an empty list, never with null: the rule copies
// the skills into every brief as they stand, and a model handed null gives
// null back where a list is required.
func TestAProfileWritesEmptyListsRatherThanNull(t *testing.T) {
	t.Parallel()

	written, err := profile.Marshal(profile.New(profile.Student{Pseudonym: "Comet", Grade: 3}, "1.2.3", editedAt))
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}
	for _, want := range []string{`"excluded_skills": []`, `"interests": []`} {
		if !strings.Contains(string(written), want) {
			t.Errorf("the file does not say %s:\n%s", want, written)
		}
	}
}

// The last answer is the end of the window, and a child who has answered
// nothing has none.
func TestTheLastAnswerIsTheEndOfTheWindow(t *testing.T) {
	t.Parallel()

	p := parseFixture(t, "masha")
	last, answered := p.LastAnswer()
	if !answered || last.TaskID != p.Recent[len(p.Recent)-1].TaskID {
		t.Errorf("LastAnswer() = %s, %v; want %s", last.TaskID, answered, p.Recent[len(p.Recent)-1].TaskID)
	}
	if _, answered := parseFixture(t, "sasha").LastAnswer(); answered {
		t.Error("LastAnswer() found an answer for a child who has given none")
	}
}

// What a model hands save_profile is anything at all. An edit made of it never
// takes the call down, and an edit it lets through is always a profile the
// file can be written as.
func FuzzEdit(f *testing.F) {
	f.Add("Comet", 3, "space", "fractions", "Reads slowly.\nLikes cats.", "pt-br")
	f.Add("", 0, "", "", "", "")
	f.Add("\a\u2028Otter\u0085", 7, "\t", "calculus", strings.Repeat("ж", profile.MaxNotes+1), "und")
	f.Add(strings.Repeat("o", profile.MaxPseudonym+1), -1, strings.Repeat("a", 41), "negative_numbers", "\x00", "en_US")

	f.Fuzz(func(t *testing.T, pseudonym string, grade int, interest, skill, notes, language string) {
		edit := profile.Edit{
			Pseudonym:      &pseudonym,
			Grade:          &grade,
			Interests:      []string{interest, interest},
			ExcludedSkills: []string{skill},
			Notes:          &notes,
			UILanguage:     &language,
		}
		student, problems := profile.NewStudent(&edit, skills)
		if len(problems) > 0 {
			return
		}
		p := profile.New(student, "fuzz", editedAt)
		if err := p.Validate(); err != nil {
			t.Fatalf("an edit let through makes a profile Validate refuses: %v", err)
		}
		if _, err := profile.Marshal(p); err != nil {
			t.Fatalf("an edit let through makes a profile that cannot be written: %v", err)
		}
	})
}

func count(n int) []string {
	items := make([]string, n)
	for i := range items {
		items[i] = strings.Repeat("x", i+1)
	}
	return items
}

func wantText(t *testing.T, field, got, want string) {
	t.Helper()
	if got != want {
		t.Errorf("%s = %q, want %q", field, got, want)
	}
}

func wantList(t *testing.T, field string, got []string, want ...string) {
	t.Helper()
	if got == nil || !slices.Equal(got, want) {
		t.Errorf("%s = %#v, want %q", field, got, want)
	}
}

func wantLanguage(t *testing.T, got *string, want string) {
	t.Helper()
	if got == nil || *got != want {
		t.Errorf("ui_language = %v, want %q", got, want)
	}
}

func wantNoLanguage(t *testing.T, got *string) {
	t.Helper()
	if got != nil {
		t.Errorf("ui_language = %q, want null", *got)
	}
}

// The skills left out are a set: the same skills in another order change
// nothing and write nothing. The interests are not: tasks are dressed in them
// in turn, so their order is a change.
func TestTheSameSkillsInAnotherOrderAreNoChange(t *testing.T) {
	t.Parallel()

	p := profile.New(newStudent(), "1.2.3", editedAt)
	p.Student.ExcludedSkills = []string{"fractions", "negative_numbers"}
	if changed, problems := p.Change(&profile.Edit{ExcludedSkills: []string{"negative_numbers", "fractions"}},
		skills, "1.2.4", editedAt); changed || len(problems) > 0 {
		t.Errorf("Change(the skills reordered) = %v, %v; want nothing changed", changed, problems)
	}
	if changed, _ := p.Change(&profile.Edit{Interests: []string{"cats", "space"}}, skills, "1.2.4", editedAt); !changed {
		t.Error("Change(the interests reordered) changed nothing, want the order kept as a change")
	}
}

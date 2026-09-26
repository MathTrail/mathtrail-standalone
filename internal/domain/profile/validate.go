package profile

import (
	"fmt"
	"math"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/rating"
	"github.com/MathTrail/mathtrail-standalone/internal/domain/solver"
)

// The caps of the file, counted in characters rather than bytes: a pseudonym
// of thirty-two letters is thirty-two letters in every script.
//
// They are exported because they are also what a screen tells the parent
// before they type past one, and a limit enforced in one place and announced
// in another drifts apart.
const (
	// MaxPseudonym is how long a pseudonym may be.
	MaxPseudonym = 32
	// MinGrade is the first school year this service is for.
	MinGrade = 1
	// MaxGrade is the last.
	MaxGrade = 6
	// MaxInterests is how many settings a task may be dressed in.
	MaxInterests = 10
	// MaxInterest is how long one of them may be.
	MaxInterest = 40
	// MaxExcludedSkills is the size of the skill catalog: excluding more than
	// all of them is not a profile anybody could be taught from.
	MaxExcludedSkills = 15
	// MaxNotes is how much free-form context the parent may write. It is a
	// size budget as much as a privacy one: every generation carries it.
	MaxNotes = 500
	// MaxRecent is how many answers the history window holds.
	MaxRecent = 20
	// MinRecent is how far that window may ever be pruned: the rule reads the
	// end of it after a failure, and an empty one would leave it nothing to
	// work on.
	MinRecent = 5
	// MaxFingerprints is how many past tasks are remembered as sketches.
	MaxFingerprints = 200
	// MaxAttempts is how many times the model may hand back a task that does
	// not pass the checks.
	MaxAttempts = 3
)

// MinDifficulty and MaxDifficulty bound a task's difficulty inside its level:
// from the first of the difficulties the rating puts on its ladder to the
// last, and no others.
const (
	MinDifficulty = 1
	MaxDifficulty = rating.Difficulties
)

// isNumber reports whether a level is one. A level that is not — the result
// of a division nobody meant — cannot be written to the file at all, and the
// encoder that refuses it names JSON rather than the field it came from.
func isNumber(level float64) bool { return !math.IsNaN(level) && !math.IsInf(level, 0) }

// Validate reports the first thing about a profile this service could not
// work with, naming the field and the limit it broke.
func (p *Profile) Validate() error {
	if p.SchemaVersion != Version {
		return fmt.Errorf("%w: schema_version is %d, this service reads %d", ErrInvalid, p.SchemaVersion, Version)
	}
	if p.StudentID == "" {
		return fmt.Errorf("%w: student_id is empty", ErrInvalid)
	}
	if p.CreatedAt.IsZero() || p.UpdatedAt.IsZero() {
		return fmt.Errorf("%w: created_at and updated_at are both required", ErrInvalid)
	}
	if p.Revision < 1 {
		return fmt.Errorf("%w: revision is %d, want 1 or more", ErrInvalid, p.Revision)
	}

	if err := p.Student.validate(); err != nil {
		return err
	}
	if p.Ratings.Answers < 0 || p.Ratings.ConsecutiveFailures < 0 {
		return fmt.Errorf("%w: ratings count answers, and a count is never negative", ErrInvalid)
	}
	if !isNumber(p.Ratings.Theta) {
		return fmt.Errorf("%w: ratings.theta is %v, and a level is a number", ErrInvalid, p.Ratings.Theta)
	}
	if !isNumber(p.Ratings.Start) {
		return fmt.Errorf("%w: ratings.start is %v, and a level is a number", ErrInvalid, p.Ratings.Start)
	}
	if err := validateTopics(p.Topics); err != nil {
		return err
	}
	if err := validateRecent(p.Recent); err != nil {
		return err
	}
	if len(p.TaskFingerprints) > MaxFingerprints {
		return fmt.Errorf("%w: task_fingerprints holds %d, the limit is %d",
			ErrInvalid, len(p.TaskFingerprints), MaxFingerprints)
	}
	if err := p.CurrentTask.validate(); err != nil {
		return err
	}
	if err := p.OpenRequest.validate(); err != nil {
		return err
	}
	return p.Daily.validate()
}

func (s *Student) validate() error {
	switch count := utf8.RuneCountInString(s.Pseudonym); {
	case count == 0:
		return fmt.Errorf("%w: student.pseudonym is empty", ErrInvalid)
	case count > MaxPseudonym:
		return fmt.Errorf("%w: student.pseudonym is %d characters, the limit is %d",
			ErrInvalid, count, MaxPseudonym)
	case strings.ContainsFunc(s.Pseudonym, unicode.IsControl):
		return fmt.Errorf("%w: student.pseudonym carries a control character", ErrInvalid)
	}

	if s.Grade < MinGrade || s.Grade > MaxGrade {
		return fmt.Errorf("%w: student.grade is %d, the grades are %d to %d",
			ErrInvalid, s.Grade, MinGrade, MaxGrade)
	}
	if len(s.Interests) > MaxInterests {
		return fmt.Errorf("%w: student.interests holds %d, the limit is %d",
			ErrInvalid, len(s.Interests), MaxInterests)
	}
	for _, interest := range s.Interests {
		if count := utf8.RuneCountInString(interest); count == 0 || count > MaxInterest {
			return fmt.Errorf("%w: student.interests has one of %d characters, the limit is 1 to %d",
				ErrInvalid, count, MaxInterest)
		}
	}
	if len(s.ExcludedSkills) > MaxExcludedSkills {
		return fmt.Errorf("%w: student.excluded_skills holds %d, the limit is %d",
			ErrInvalid, len(s.ExcludedSkills), MaxExcludedSkills)
	}
	if count := utf8.RuneCountInString(s.Notes); count > MaxNotes {
		return fmt.Errorf("%w: student.notes is %d characters, the limit is %d", ErrInvalid, count, MaxNotes)
	}
	if s.UILanguage != nil && *s.UILanguage == "" {
		return fmt.Errorf("%w: student.ui_language is an empty string; a profile that follows the chat says null", ErrInvalid)
	}
	return nil
}

func validateTopics(topics map[string]Topic) error {
	for id, topic := range topics {
		switch {
		case id == "":
			return fmt.Errorf("%w: topics has an entry with no id", ErrInvalid)
		case topic.Answers < 0 || topic.Correct < 0 || topic.TopStreak < 0 || topic.WrongStreak < 0:
			return fmt.Errorf("%w: topics[%s] counts answers, and a count is never negative", ErrInvalid, id)
		case !isNumber(topic.Delta):
			return fmt.Errorf("%w: topics[%s].delta is %v, and a level is a number", ErrInvalid, id, topic.Delta)
		case topic.Correct > topic.Answers:
			return fmt.Errorf("%w: topics[%s] has %d correct out of %d answers",
				ErrInvalid, id, topic.Correct, topic.Answers)
		case (topic.MasteredSince == nil) != (topic.MasteredLevel == nil):
			return fmt.Errorf("%w: topics[%s] has mastered_since and mastered_level apart; a topic is mastered on a day and at a level, or not at all",
				ErrInvalid, id)
		case topic.MasteredLevel != nil && !topic.MasteredLevel.Known():
			return fmt.Errorf("%w: topics[%s].mastered_level is %q, want one of %v",
				ErrInvalid, id, *topic.MasteredLevel, rating.GradeLevels())
		}
		for trap, times := range topic.Traps {
			if trap == "" || times < 1 {
				return fmt.Errorf("%w: topics[%s].traps counts a trap %d times", ErrInvalid, id, times)
			}
		}
	}
	return nil
}

func validateRecent(recent []Answer) error {
	if len(recent) > MaxRecent {
		return fmt.Errorf("%w: recent holds %d answers, the window is %d", ErrInvalid, len(recent), MaxRecent)
	}
	for i := range recent {
		answer := &recent[i]
		switch {
		case answer.TaskID == "" || answer.Topic == "":
			return fmt.Errorf("%w: recent[%d] names no task or no topic", ErrInvalid, i)
		case answer.AnsweredAt.IsZero():
			return fmt.Errorf("%w: recent[%d] has no answered_at", ErrInvalid, i)
		case answer.Difficulty < MinDifficulty || answer.Difficulty > MaxDifficulty:
			return fmt.Errorf("%w: recent[%d].difficulty is %d, the difficulties are %d to %d",
				ErrInvalid, i, answer.Difficulty, MinDifficulty, MaxDifficulty)
		case !answer.GradeLevel.Known():
			return fmt.Errorf("%w: recent[%d].grade_level is %q, want one of %v",
				ErrInvalid, i, answer.GradeLevel, rating.GradeLevels())
		case answer.Pace != PaceFast && answer.Pace != PaceNormal && answer.Pace != PaceSlow:
			return fmt.Errorf("%w: recent[%d].pace is %q, want fast, normal or slow", ErrInvalid, i, answer.Pace)
		case answer.Chosen != "" && solver.Place(answer.Chosen) < 0:
			return fmt.Errorf("%w: recent[%d].chosen is %q, want one of %s",
				ErrInvalid, i, answer.Chosen, strings.Join(solver.Letters(), ", "))
		case answer.Correct && (answer.Chosen != "" || answer.Trap != ""):
			// A correct answer has no mistake to name, and naming one would
			// put a trap of the current lesson where it does not belong.
			return fmt.Errorf("%w: recent[%d] is correct and still names a chosen option or a trap", ErrInvalid, i)
		}
	}
	return nil
}

func (t *CurrentTask) validate() error {
	if t == nil {
		return nil
	}
	switch {
	case t.ID == "" || t.Topic == "" || t.Wording == "":
		return fmt.Errorf("%w: current_task needs an id, a topic and a wording", ErrInvalid)
	case t.Sealed == "":
		return fmt.Errorf("%w: current_task.sealed is empty; a task with nothing sealed has its answer in the open", ErrInvalid)
	case t.IssuedAt.IsZero():
		return fmt.Errorf("%w: current_task has no issued_at, and the pace is measured from it", ErrInvalid)
	case t.Difficulty < MinDifficulty || t.Difficulty > MaxDifficulty:
		return fmt.Errorf("%w: current_task.difficulty is %d, the difficulties are %d to %d",
			ErrInvalid, t.Difficulty, MinDifficulty, MaxDifficulty)
	case !t.GradeLevel.Known():
		return fmt.Errorf("%w: current_task.grade_level is %q, want one of %v",
			ErrInvalid, t.GradeLevel, rating.GradeLevels())
	case len(t.Options) != solver.Count:
		return fmt.Errorf("%w: current_task offers %d options, want %d", ErrInvalid, len(t.Options), solver.Count)
	}
	for place := range solver.Count {
		if letter := solver.Letter(place); solver.Blank(t.Options[letter]) {
			return fmt.Errorf("%w: current_task.options shows nothing under %q", ErrInvalid, letter)
		}
	}
	return nil
}

func (r *OpenRequest) validate() error {
	if r == nil {
		return nil
	}
	switch {
	case r.ID == "":
		return fmt.Errorf("%w: open_request has no id", ErrInvalid)
	case r.OpenedAt.IsZero():
		return fmt.Errorf("%w: open_request has no opened_at, and the window is measured from it", ErrInvalid)
	case r.Attempts < 0 || r.Attempts > MaxAttempts:
		return fmt.Errorf("%w: open_request.attempts is %d, the limit is %d", ErrInvalid, r.Attempts, MaxAttempts)
	case r.TutorMode != TutorRule && r.TutorMode != TutorLLM:
		return fmt.Errorf("%w: open_request.tutor_mode is %q, want rule or llm", ErrInvalid, r.TutorMode)
	case r.Brief.TargetConcept == "":
		return fmt.Errorf("%w: open_request.brief names no topic", ErrInvalid)
	case r.Brief.Difficulty < MinDifficulty || r.Brief.Difficulty > MaxDifficulty:
		return fmt.Errorf("%w: open_request.brief.difficulty is %d, the difficulties are %d to %d",
			ErrInvalid, r.Brief.Difficulty, MinDifficulty, MaxDifficulty)
	case !r.Brief.GradeLevel.Known():
		return fmt.Errorf("%w: open_request.brief.grade_level is %q, want one of %v",
			ErrInvalid, r.Brief.GradeLevel, rating.GradeLevels())
	case r.Brief.PedagogicalGoal != GoalReinforce && r.Brief.PedagogicalGoal != GoalNewTopic:
		return fmt.Errorf("%w: open_request.brief.pedagogical_goal is %q, want reinforce or new_topic",
			ErrInvalid, r.Brief.PedagogicalGoal)
	}
	return nil
}

func (d *Daily) validate() error {
	if d.Accepted < 0 || d.Failed < 0 {
		return fmt.Errorf("%w: daily counts tasks, and a count is never negative", ErrInvalid)
	}
	if d.Date.IsZero() {
		return fmt.Errorf("%w: daily has no date, and the counters belong to a day", ErrInvalid)
	}
	return nil
}

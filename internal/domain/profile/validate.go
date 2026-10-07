package profile

import (
	"fmt"
	"math"
	"strings"

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
	// all of them is not a profile anybody could be taught from. A test of the
	// content holds the two to the same number.
	MaxExcludedSkills = 25
	// MaxNotes is how much free-form context the parent may write. It is a
	// size budget as much as a privacy one: every generation carries it.
	MaxNotes = 500
	// MaxLanguageTag is how long a language tag may be, the lessons' and a
	// task's alike: the length every reader of a BCP 47 tag is asked to hold,
	// and longer than any language a child is taught in needs.
	MaxLanguageTag = 35
	// MaxCountry is how long a country may be: its code in ISO 3166-1, two
	// letters.
	MaxCountry = 2
	// MaxRegion is how long a region may be: its code in ISO 3166-2, the
	// country's two letters, a hyphen and at most three more.
	MaxRegion = 6
	// MaxLessonTopic is how long the topic chosen for the lessons may be: a
	// topic's id in the catalog, a family and a name joined by a dot, which
	// none comes near.
	MaxLessonTopic = 64
	// MaxRecent is how many entries the history window holds, answers and
	// skipped tasks together.
	MaxRecent = 20
	// MinRecent is how many of the latest answers the window keeps, whatever
	// else it has to let go of: the rule reads the last answer after a failure
	// and the trial series reads its answers back, and a run of skipped tasks
	// must not push them out.
	MinRecent = 5
	// MaxFingerprints is how many past tasks are remembered as sketches.
	MaxFingerprints = 200
	// MaxAttempts is how many times the model may hand back a task that does
	// not pass the checks.
	MaxAttempts = 3
)

// mostCount is the most any count in the file may be — the revision, the
// answers, the streaks, the day's tasks: far past what the lessons of a
// childhood add up to, and far enough below where a number stops counting
// that one more always fits.
const mostCount = 1 << 30

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
func (p *Profile) Validate() error { return p.validate(mostCount) }

// validate is Validate with the most a count may be handed in: a profile is
// written with its counts up to mostCount, and read with room in each for one
// more.
func (p *Profile) validate(most int) error {
	if err := p.validateHeader(); err != nil {
		return err
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
	if p.Ratings.MasteryRule != "" && p.Ratings.MasteryRule != MasteryRuleCautious {
		return fmt.Errorf("%w: ratings.mastery_rule is %q, want %s or none", ErrInvalid, p.Ratings.MasteryRule, MasteryRuleCautious)
	}
	if err := validateTopics(p.Topics); err != nil {
		return err
	}
	if err := validateRecent(p.Recent); err != nil {
		return err
	}
	if err := validateRatingDays(p.RatingDays); err != nil {
		return err
	}
	if len(p.TaskFingerprints) > MaxFingerprints {
		return fmt.Errorf("%w: task_fingerprints holds %d, the limit is %d",
			ErrInvalid, len(p.TaskFingerprints), MaxFingerprints)
	}
	if err := p.validateTasks(); err != nil {
		return err
	}
	if err := p.Daily.validate(); err != nil {
		return err
	}
	return p.countsUpTo(most)
}

// validateTasks checks the tasks the file holds: the one on the card, the one
// kept ready and the one being written.
func (p *Profile) validateTasks() error {
	if err := p.CurrentTask.validate(); err != nil {
		return err
	}
	if err := p.ReadyTask.validate(); err != nil {
		return err
	}
	return p.OpenRequest.validate()
}

// validateHeader checks what the file says of itself: its shape, whose it is,
// when it was made and last changed, and that it has been written.
func (p *Profile) validateHeader() error {
	switch {
	case p.SchemaVersion != Version:
		return fmt.Errorf("%w: schema_version is %d, this service reads %d", ErrInvalid, p.SchemaVersion, Version)
	case p.StudentID == "":
		return fmt.Errorf("%w: student_id is empty", ErrInvalid)
	case p.CreatedAt.IsZero() || p.UpdatedAt.IsZero():
		return fmt.Errorf("%w: created_at and updated_at are both required", ErrInvalid)
	case p.Revision < 1:
		return fmt.Errorf("%w: revision is %d, want 1 or more", ErrInvalid, p.Revision)
	}
	return nil
}

// countsUpTo reports a count of the profile past the most given. A step of a
// lesson moves each count by one at most, so a profile read with every count
// below the most can always be written after it.
func (p *Profile) countsUpTo(most int) error {
	past := func(counts ...int) bool {
		for _, count := range counts {
			if count > most {
				return true
			}
		}
		return false
	}
	switch {
	case past(p.Revision):
		return fmt.Errorf("%w: revision is past %d", ErrInvalid, most)
	case past(p.Ratings.Answers, p.Ratings.ConsecutiveFailures):
		return fmt.Errorf("%w: ratings count past %d", ErrInvalid, most)
	case past(p.Daily.Accepted, p.Daily.Failed):
		return fmt.Errorf("%w: daily counts past %d", ErrInvalid, most)
	}
	for _, topic := range p.Topics {
		if past(topic.Answers, topic.Correct, topic.Skipped, topic.TopStreak, topic.WrongStreak) {
			return fmt.Errorf("%w: a topic counts past %d", ErrInvalid, most)
		}
		for _, times := range topic.Traps {
			if past(times) {
				return fmt.Errorf("%w: a topic counts a trap past %d", ErrInvalid, most)
			}
		}
	}
	return nil
}

func (s *Student) validate() error {
	if found := s.problems(); len(found) > 0 {
		return fmt.Errorf("%w: student.%s", ErrInvalid, found[0])
	}
	return nil
}

func validateTopics(topics map[string]Topic) error {
	for id, topic := range topics {
		switch {
		case id == "":
			return fmt.Errorf("%w: topics has an entry with no id", ErrInvalid)
		case topic.Answers < 0 || topic.Correct < 0 || topic.Skipped < 0 || topic.TopStreak < 0 || topic.WrongStreak < 0:
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
		return fmt.Errorf("%w: recent holds %d entries, the window is %d", ErrInvalid, len(recent), MaxRecent)
	}
	for i := range recent {
		if err := recent[i].validate(i); err != nil {
			return err
		}
	}
	return nil
}

// validateRatingDays checks the days the history of the ratings keeps: no
// more than the week reaches, each on a date of its own and in the order of
// their dates, every level a number and every topic named.
func validateRatingDays(days []RatingDay) error {
	if len(days) > RatingDaysKept {
		return fmt.Errorf("%w: rating_days holds %d days, the week is %d", ErrInvalid, len(days), RatingDaysKept)
	}
	for i := range days {
		day := &days[i]
		switch {
		case day.Date.IsZero():
			return fmt.Errorf("%w: rating_days[%d] has no date", ErrInvalid, i)
		case i > 0 && !days[i-1].Date.Before(day.Date.Time):
			return fmt.Errorf("%w: rating_days[%d] is not after the day before it; the days are kept oldest first, one to a date",
				ErrInvalid, i)
		case !isNumber(day.Theta):
			return fmt.Errorf("%w: rating_days[%d].theta is %v, and a level is a number", ErrInvalid, i, day.Theta)
		case day.Deltas == nil:
			return fmt.Errorf("%w: rating_days[%d] has no deltas; a day with no topic answered before it has {}", ErrInvalid, i)
		}
		for id, delta := range day.Deltas {
			if id == "" || !isNumber(delta) {
				return fmt.Errorf("%w: rating_days[%d].deltas holds %v for %q; a correction is a number, of a named topic",
					ErrInvalid, i, delta, id)
			}
		}
	}
	return nil
}

// validate checks one entry of the window, the i-th: where the task stood and
// when, and then the outcome — which an answer has and a skipped task does not.
func (a *Answer) validate(i int) error {
	switch {
	case a.Before != nil && (!isNumber(a.Before.Theta) || !isNumber(a.Before.Delta)):
		return fmt.Errorf("%w: recent[%d].before holds a level that is not a number", ErrInvalid, i)
	case a.TaskID == "" || a.Topic == "":
		return fmt.Errorf("%w: recent[%d] names no task or no topic", ErrInvalid, i)
	case a.AnsweredAt.IsZero():
		return fmt.Errorf("%w: recent[%d] has no answered_at", ErrInvalid, i)
	case a.Difficulty < MinDifficulty || a.Difficulty > MaxDifficulty:
		return fmt.Errorf("%w: recent[%d].difficulty is %d, the difficulties are %d to %d",
			ErrInvalid, i, a.Difficulty, MinDifficulty, MaxDifficulty)
	case !a.GradeLevel.Known():
		return fmt.Errorf("%w: recent[%d].grade_level is %q, want one of %v",
			ErrInvalid, i, a.GradeLevel, rating.GradeLevels())
	case a.Skipped:
		return a.validateSkipped(i)
	case a.Pace != PaceFast && a.Pace != PaceNormal && a.Pace != PaceSlow:
		return fmt.Errorf("%w: recent[%d].pace is %q, want fast, normal or slow", ErrInvalid, i, a.Pace)
	case a.Chosen != "" && solver.Place(a.Chosen) < 0:
		return fmt.Errorf("%w: recent[%d].chosen is %q, want one of %s",
			ErrInvalid, i, a.Chosen, strings.Join(solver.Letters(), ", "))
	case a.Correct && (a.Chosen != "" || a.Trap != ""):
		// A correct answer has no mistake to name, and naming one would put a
		// trap of the current lesson where it does not belong.
		return fmt.Errorf("%w: recent[%d] is correct and still names a chosen option or a trap", ErrInvalid, i)
	case a.Confused && (a.Correct || a.Chosen != "" || a.Trap != ""):
		// "I don't know" showed the solution instead of taking a choice: it is
		// never right, it chose no option, and the map of the child's mistakes
		// must not count a trap it never fell for.
		return fmt.Errorf("%w: recent[%d] is \"I don't know\" and still says it was right, or names a chosen option or a trap",
			ErrInvalid, i)
	}
	return nil
}

// validateSkipped checks that a skipped task carries no outcome: there was no
// answer, so there is nothing right or wrong about it, no option chosen, no
// hint opened, no time taken and no level it moved from. A file saying
// otherwise would be read by the progress screen as an answer the child never
// gave.
func (a *Answer) validateSkipped(i int) error {
	if a.Correct || a.Chosen != "" || a.Trap != "" || a.HintUsed || a.Confused || a.Pace != "" || a.Before != nil {
		return fmt.Errorf("%w: recent[%d] is a skipped task and still carries an outcome, which a task left without an answer does not have",
			ErrInvalid, i)
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
	case t.Language == "":
		return fmt.Errorf("%w: current_task has no language, and it is written in one", ErrInvalid)
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
	case t.TutorMode != "" && !t.TutorMode.Known():
		return fmt.Errorf("%w: current_task.tutor_mode is %q, want rule, llm, person or none", ErrInvalid, t.TutorMode)
	}
	if err := validateOptions("current_task", t.Options); err != nil {
		return err
	}
	return t.Answered.validate()
}

// validateOptions checks that a task offers its five options, each showing
// something: a blank one could be chosen by nobody, and its letter would be
// the answer by elimination or a choice the card cannot show.
func validateOptions(field string, options map[string]string) error {
	if len(options) != solver.Count {
		return fmt.Errorf("%w: %s offers %d options, want %d", ErrInvalid, field, len(options), solver.Count)
	}
	for place := range solver.Count {
		if letter := solver.Letter(place); solver.Blank(options[letter]) {
			return fmt.Errorf("%w: %s.options shows nothing under %q", ErrInvalid, field, letter)
		}
	}
	return nil
}

// validate checks a task kept ready as the task on the card is checked, but
// for what only a task handed out has — the moment it was handed out and an
// answer — and with how its writing went: when its request was opened, when
// it was accepted, and an attempt the checks could have accepted it at.
func (r *ReadyTask) validate() error {
	if r == nil {
		return nil
	}
	switch {
	case r.ID == "" || r.Topic == "" || r.Wording == "":
		return fmt.Errorf("%w: ready_task needs an id, a topic and a wording", ErrInvalid)
	case r.Language == "":
		return fmt.Errorf("%w: ready_task has no language, and it is written in one", ErrInvalid)
	case r.Sealed == "":
		return fmt.Errorf("%w: ready_task.sealed is empty; a task with nothing sealed has its answer in the open", ErrInvalid)
	case r.OpenedAt.IsZero() || r.WrittenAt.IsZero():
		return fmt.Errorf("%w: ready_task needs an opened_at and a written_at, which its writing is timed by", ErrInvalid)
	case r.Attempts < 1 || r.Attempts > MaxAttempts:
		return fmt.Errorf("%w: ready_task.attempts is %d, and a task is accepted at an attempt from 1 to %d",
			ErrInvalid, r.Attempts, MaxAttempts)
	case r.Difficulty < MinDifficulty || r.Difficulty > MaxDifficulty:
		return fmt.Errorf("%w: ready_task.difficulty is %d, the difficulties are %d to %d",
			ErrInvalid, r.Difficulty, MinDifficulty, MaxDifficulty)
	case !r.GradeLevel.Known():
		return fmt.Errorf("%w: ready_task.grade_level is %q, want one of %v",
			ErrInvalid, r.GradeLevel, rating.GradeLevels())
	case !r.TutorMode.Known():
		return fmt.Errorf("%w: ready_task.tutor_mode is %q, want rule, llm or person", ErrInvalid, r.TutorMode)
	}
	return validateOptions("ready_task", r.Options)
}

// validate checks the answer a task keeps: what it tells again has to be an
// answer the child could give, and where the child stood has to be a place on
// the ladder.
func (g *Given) validate() error {
	switch {
	case g == nil:
		return nil
	case !choosable(g.Choice):
		return fmt.Errorf("%w: current_task.answered.choice is %q, want one of %s or %s",
			ErrInvalid, g.Choice, strings.Join(solver.Letters(), ", "), DontKnow)
	case !isNumber(g.LevelBefore) || !isNumber(g.LevelAfter):
		return fmt.Errorf("%w: current_task.answered holds a level that is not a number", ErrInvalid)
	case g.Trial < 0 || g.Trial > rating.TrialAnswers:
		return fmt.Errorf("%w: current_task.answered.trial is %d, and the trial series has %d answers",
			ErrInvalid, g.Trial, rating.TrialAnswers)
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
	case r.Language == "":
		return fmt.Errorf("%w: open_request has no language, and the task is to be written in it", ErrInvalid)
	case r.OpenedAt.IsZero():
		return fmt.Errorf("%w: open_request has no opened_at, and the window is measured from it", ErrInvalid)
	case r.Attempts < 0 || r.Attempts > MaxAttempts:
		return fmt.Errorf("%w: open_request.attempts is %d, the limit is %d", ErrInvalid, r.Attempts, MaxAttempts)
	case !r.TutorMode.Known():
		return fmt.Errorf("%w: open_request.tutor_mode is %q, want rule, llm or person", ErrInvalid, r.TutorMode)
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

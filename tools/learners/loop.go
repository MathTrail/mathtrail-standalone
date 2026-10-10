package main

import (
	"fmt"
	"time"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/profile"
	"github.com/MathTrail/mathtrail-standalone/internal/domain/rating"
	"github.com/MathTrail/mathtrail-standalone/internal/domain/solver"
	"github.com/MathTrail/mathtrail-standalone/internal/domain/tutor"
)

// How a session of the simulation runs: answers two minutes after the task is
// issued, the next task at once, and a new day after every tenth answer.
const (
	answerTime     = 2 * time.Minute
	answersPerDay  = 10
	simulationTask = "A simulated task."
)

// firstDay is when every simulated child starts.
var firstDay = time.Date(2026, time.October, 1, 9, 0, 0, 0, time.UTC)

// placeholderSketch is the sketch every simulated task is given: the
// near-duplicate check is not run here, and the profile only keeps it.
var placeholderSketch = "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"

// rule is a way of estimating where a child stands, from which the service's
// rule chooses each task.
type rule struct {
	name  string
	shape structure
	// service marks the product's own path, which no estimate of the
	// harness's replaces.
	service bool
	// trial marks a rule that starts after the service's trial series, whose
	// estimator follows the series until then.
	trial bool
	// ahead marks the service's path with the next task written ahead, as the
	// service writes it once a task is on the card: chosen while the child
	// still works on the task before it, and handed out after that answer,
	// whatever it was.
	ahead bool
	// ceiling marks a rule that knows what no rule of answers can: where the
	// child truly stands. It shows how far a rule could go, and is no rule to
	// choose or to compare with as one.
	ceiling bool
	// candidacy is what a choice reads of a rule that is a candidate in it;
	// nil for every other rule.
	candidacy *candidacy
	make      func(c *child, start float64) estimator
	// mastery makes the test of mastery a rule puts in the service's place
	// for a child; nil for a rule that keeps the service's own.
	mastery func() masteryTest
}

// world is what every run shares: the catalog, its topics, the levels each
// topic is taught at, the lowest first, and the points of its ladder, the
// sealer and how many answers a child gives.
type world struct {
	catalog tutor.Catalog
	topics  []string
	levels  map[string][]rating.GradeLevel
	ladders map[string][]rating.Point
	sealer  profile.Sealer
	answers int
}

// session is one child under one rule, the rule's mastery in the service's
// place, if it has one of its own, and the next task, once a rule that writes
// it ahead has chosen it.
type session struct {
	w         *world
	r         *rule
	c         *child
	p         *profile.Profile
	est       estimator
	mastering *mastering
	ahead     *chosen
	now       time.Time
	result    *childResult
}

// chosen is a brief, and who chose it.
type chosen struct {
	brief profile.Brief
	mode  profile.TutorMode
}

// run gives a child every answer of a session under a rule and says what the
// measures came to.
func run(w *world, r *rule, c *child) (*childResult, error) {
	s := newSession(w, r, c)
	for k := range w.answers {
		if err := s.step(k); err != nil {
			return nil, fmt.Errorf("%s under %s, answer %d: %w", c.id, r.name, k+1, err)
		}
	}
	return s.result, nil
}

// newSession starts a child under a rule, with a new profile at the start the
// service gives the child's grade.
func newSession(w *world, r *rule, c *child) *session {
	p := profile.New(profile.Student{Grade: c.grade, Pseudonym: "sim"}, "simulation", firstDay)
	s := &session{w: w, r: r, c: c, p: p, now: firstDay, result: newChildResult(c)}
	s.est = service{p: p}
	if !r.service {
		s.est = r.make(c, p.Ratings.Start)
	}
	if r.mastery != nil {
		s.mastering = newMastering(r.mastery())
	}
	return s
}

// writes says whether the harness puts the rule's levels into the profile
// before the service chooses: every rule but the service's own, and a step rule
// only once the trial series is over.
func (s *session) writes(k int) bool {
	return !s.r.service && (!s.r.trial || k >= rating.TrialAnswers)
}

// step is one task and its answer.
func (s *session) step(k int) error {
	if s.writes(k) {
		s.writeLevels()
	}
	brief, mode, err := s.next()
	if err != nil {
		return err
	}
	point := rating.Point{GradeLevel: brief.GradeLevel, Difficulty: brief.Difficulty}
	topic, beta := brief.TargetConcept, point.Beta()
	draws := s.c.next()
	truth := s.c.chance(topic, s.c.writtenDifficulty(beta, k, draws.writing))
	correct := draws.correct < truth
	before := s.observeBefore(&brief, beta, truth, correct, k)
	recorded, err := s.take(&brief, mode, &draws, correct, draws.hint < s.c.hint, k)
	if err != nil {
		return err
	}
	s.result.after(s, &brief, &recorded, &before, k)
	s.c.after(topic, k)
	s.result.checkpoint(s, k)
	s.advance(k)
	return nil
}

// next is the brief of the task the child gets now: the one chosen while the
// child worked on the task before it, under a rule that writes the next task
// ahead, or else the one the service's rule chooses now.
func (s *session) next() (profile.Brief, profile.TutorMode, error) {
	if ahead := s.ahead; ahead != nil {
		s.ahead = nil
		return ahead.brief, ahead.mode, nil
	}
	return tutor.Next(s.p, s.w.catalog, tutor.Choice{})
}

// writeLevels puts the rule's estimate into the profile, where the service's
// rule and its mastery read a child's levels.
func (s *session) writeLevels() {
	overall := s.est.overall()
	s.p.Ratings.Theta = overall
	for _, topic := range s.w.topics {
		summary := s.p.Topics[topic]
		summary.Delta = s.est.level(topic) - overall
		s.p.Topics[topic] = summary
	}
}

// take is what a lesson does with an answer once the task is set and
// answered: the service records it, the rule's estimate takes it in, and the
// rule's own mastery, if it has one, judges it.
func (s *session) take(brief *profile.Brief, mode profile.TutorMode, draws *answerDraws, correct, hint bool, k int) (profile.Recorded, error) {
	recorded, err := s.answer(brief, mode, draws, correct, hint)
	if err != nil {
		return profile.Recorded{}, err
	}
	point := rating.Point{GradeLevel: brief.GradeLevel, Difficulty: brief.Difficulty}
	s.learn(brief.TargetConcept, point.Beta(), correct, k)
	s.master(brief, point.Beta(), &recorded, correct, hint)
	return recorded, nil
}

// answer issues the task and records the child's answer as the service does.
// A rule that writes the next task ahead chooses it in between, while the
// child works on this one, as the service does once a task is on the card.
func (s *session) answer(brief *profile.Brief, mode profile.TutorMode, draws *answerDraws, correct, hint bool) (profile.Recorded, error) {
	task, choice, err := s.issue(brief, mode, draws, correct)
	if err != nil {
		return profile.Recorded{}, err
	}
	if s.r.ahead {
		next, nextMode, err := tutor.Ahead(s.p, s.w.catalog, tutor.Choice{})
		if err != nil {
			return profile.Recorded{}, err
		}
		s.ahead = &chosen{brief: next, mode: nextMode}
	}
	s.now = s.now.Add(answerTime)
	return s.p.Record(profile.Answered{TaskID: task.ID, Choice: choice, HintUsed: hint, At: s.now}, s.w.sealer, s.w.levels[brief.TargetConcept])
}

// issue puts the task on the child's card as the service does, and is the
// task and the option the child chooses: the key when the answer is to be
// right, and else the wrong option the draws name.
func (s *session) issue(brief *profile.Brief, mode profile.TutorMode, draws *answerDraws, correct bool) (*profile.CurrentTask, string, error) {
	s.p.Ask(brief, mode, "en", s.now)
	letters := solver.Letters()
	key := letters[draws.key]
	written := &profile.Written{Wording: simulationTask, Options: map[string]string{}, Fingerprint: placeholderSketch}
	secret := profile.TaskSecret{Answer: key, Distractors: map[string]profile.Distractor{}, Solution: simulationTask, Solver: simulationTask}
	var wrong []string
	for i, letter := range letters {
		written.Options[letter] = fmt.Sprint(i + 1)
		if letter != key {
			secret.Distractors[letter] = profile.Distractor{Trap: "off_by_one", Text: simulationTask}
			wrong = append(wrong, letter)
		}
	}
	task, err := s.p.Issue(written, &secret, s.w.sealer, s.now)
	if err != nil {
		return nil, "", err
	}
	choice := key
	if !correct {
		choice = wrong[draws.wrong]
	}
	return task, choice, nil
}

// learn takes the answer into the rule's estimate. A rule that starts after
// the trial series is handed the series' answers and the series' estimate as
// it stands after each, so that until its own first step it chooses and
// predicts from the estimate the service has.
func (s *session) learn(topic string, beta float64, correct bool, k int) {
	if s.r.service {
		return
	}
	if f, following := s.est.(follower); following && s.r.trial && k < rating.TrialAnswers {
		f.followed(topic, beta, correct, s.p.Ratings.Theta)
		return
	}
	s.est.answered(topic, beta, correct)
}

// master puts the rule's own mastery in the service's place, if it has one:
// the rule judges the answer once its estimate has taken it in, its verdict
// replaces the service's in what is recorded, and what it holds replaces what
// the service's rule wrote into the profile.
func (s *session) master(brief *profile.Brief, beta float64, recorded *profile.Recorded, correct, hint bool) {
	if s.mastering == nil {
		return
	}
	topic := brief.TargetConcept
	summary := s.p.Topics[topic]
	a := masteryAnswer{
		topic: topic, level: brief.GradeLevel, beta: beta, correct: correct, hint: hint, chance: recorded.Probability,
		answers: s.p.Ratings.Answers, inTopic: summary.Answers, wrongRun: summary.WrongStreak,
		estimate: s.est.level(topic), levels: s.w.levels[topic],
	}
	recorded.Mastered, recorded.Unmastered = s.mastering.judged(s.p, &a, profile.DateOf(s.now))
}

// advance moves the clock to the next task: at once, or to the next morning
// after every tenth answer.
func (s *session) advance(k int) {
	if (k+1)%answersPerDay == 0 {
		day := s.now.Truncate(24 * time.Hour).Add(24*time.Hour + 9*time.Hour)
		s.now = day
	}
}

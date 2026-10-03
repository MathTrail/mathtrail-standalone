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
	// trial marks a step rule that starts after the service's trial series.
	trial bool
	// ceiling marks a rule that knows what no rule of answers can: where the
	// child truly stands. It shows how far a rule could go, and is no rule to
	// choose or to compare with as one.
	ceiling bool
	make    func(c *child, start float64) estimator
}

// world is what every run shares: the catalog, its topics and the points of
// each topic's ladder, the sealer and how many answers a child gives.
type world struct {
	catalog tutor.Catalog
	topics  []string
	ladders map[string][]rating.Point
	sealer  profile.Sealer
	answers int
}

// session is one child under one rule.
type session struct {
	w      *world
	r      *rule
	c      *child
	p      *profile.Profile
	est    estimator
	now    time.Time
	result *childResult
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
	brief, mode, err := tutor.Next(s.p, s.w.catalog, tutor.Choice{})
	if err != nil {
		return err
	}
	point := rating.Point{GradeLevel: brief.GradeLevel, Difficulty: brief.Difficulty}
	topic, beta := brief.TargetConcept, point.Beta()
	draws := s.c.next()
	truth := s.c.chance(topic, s.c.writtenDifficulty(beta, k, draws.writing))
	correct := draws.correct < truth
	before := s.observeBefore(&brief, beta, truth, correct, k)
	recorded, err := s.answer(&brief, mode, &draws, correct)
	if err != nil {
		return err
	}
	s.learn(topic, beta, correct, k)
	s.result.after(s, &brief, &recorded, &before, k)
	s.c.after(topic, k)
	s.result.checkpoint(s, k)
	s.advance(k)
	return nil
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

// answer issues the task and records the child's answer as the service does.
func (s *session) answer(brief *profile.Brief, mode profile.TutorMode, draws *answerDraws, correct bool) (profile.Recorded, error) {
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
	task, err := s.p.Issue(written, secret, s.w.sealer, s.now)
	if err != nil {
		return profile.Recorded{}, err
	}
	choice := key
	if !correct {
		choice = wrong[draws.wrong]
	}
	s.now = s.now.Add(answerTime)
	return s.p.Record(profile.Answered{TaskID: task.ID, Choice: choice, HintUsed: draws.hint < s.c.hint, At: s.now}, s.w.sealer)
}

// learn takes the answer into the rule's estimate. A step rule that starts
// after the trial series only counts the series' answers and holds the
// series' estimate as it stands after each, so that until its own first step
// it chooses and predicts from the estimate the service has.
func (s *session) learn(topic string, beta float64, correct bool, k int) {
	if s.r.service {
		return
	}
	if steps, following := s.est.(*stepRule); following && s.r.trial && k < rating.TrialAnswers {
		steps.counted(topic)
		steps.takeOver(s.p.Ratings.Theta)
		return
	}
	s.est.answered(topic, beta, correct)
}

// advance moves the clock to the next task: at once, or to the next morning
// after every tenth answer.
func (s *session) advance(k int) {
	if (k+1)%answersPerDay == 0 {
		day := s.now.Truncate(24 * time.Hour).Add(24*time.Hour + 9*time.Hour)
		s.now = day
	}
}

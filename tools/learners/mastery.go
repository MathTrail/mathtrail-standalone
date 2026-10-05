package main

import (
	"math"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/profile"
	"github.com/MathTrail/mathtrail-standalone/internal/domain/rating"
)

// masteryAnswer is what a rule of mastery is told of an answer once the
// service has recorded it and the rule's estimate has taken it in: the topic,
// the task's level and difficulty on the scale, whether it was right and
// whether the hint was opened, the chance the task was handed out at, the
// answers given in all and in the topic, the topic's wrong answers in a row,
// the estimate's level in the topic, and the levels the topic is taught at,
// the lowest first. The counts and the run of wrong answers are the
// profile's, which the service keeps the same whatever the rule of mastery.
type masteryAnswer struct {
	topic            string
	level            rating.GradeLevel
	beta             float64
	correct, hint    bool
	chance           float64
	answers, inTopic int
	wrongRun         int
	estimate         float64
	levels           []rating.GradeLevel
}

// masteryTest is what tells, answer by answer, that a topic is mastered: it
// takes every answer in and names the highest level the answer shows the
// topic mastered at, if any.
type masteryTest interface {
	passed(a *masteryAnswer) (rating.GradeLevel, bool)
}

// mastering is a rule of mastery put in the service's place: a test, and the
// level each topic is held mastered at and since when. A topic is declared
// mastered at a level the test names above the one it is held at, and loses
// mastery after as many wrong answers in a row as the service's rule allows;
// what is held is written into the profile after every answer, over whatever
// the service's own rule wrote there, since the profile is where the
// service's choice of tasks reads mastery.
type mastering struct {
	test masteryTest
	held map[string]heldMastery
}

// heldMastery is the level a topic is held mastered at and the day it was.
type heldMastery struct {
	level rating.GradeLevel
	since profile.Date
}

func newMastering(test masteryTest) *mastering {
	return &mastering{test: test, held: map[string]heldMastery{}}
}

// judged takes an answer in, writes what the topic is held at into the
// profile, and says whether the answer earned mastery or lost it.
func (m *mastering) judged(p *profile.Profile, a *masteryAnswer, day profile.Date) (mastered, unmastered bool) {
	level, passed := m.test.passed(a)
	held, holds := m.held[a.topic]
	switch {
	case holds && a.wrongRun >= profile.MasteryLostAfter:
		delete(m.held, a.topic)
		unmastered = true
	case passed && (!holds || held.level.Shift() < level.Shift()):
		m.held[a.topic] = heldMastery{level: level, since: day}
		mastered = true
	}
	summary := p.Topics[a.topic]
	summary.MasteredLevel, summary.MasteredSince = nil, nil
	if now, holdsNow := m.held[a.topic]; holdsNow {
		summary.MasteredLevel, summary.MasteredSince = &now.level, &now.since
	}
	p.Topics[a.topic] = summary
	return mastered, unmastered
}

// runOf is the test the service had before the cautious estimate, with a run
// of its own length: a run of right answers at the middle of the corridor or
// harder and unaided, which a wrong answer ends and a right one to an easy or
// aided task leaves as it is, at as many answers in the topic as the service
// asks for. A run that completes is spent, whatever it earns, and names the
// level of the task that completed it.
type runOf struct {
	length int
	runs   map[string]int
}

func newRunOf(length int) *runOf { return &runOf{length: length, runs: map[string]int{}} }

func (r *runOf) passed(a *masteryAnswer) (rating.GradeLevel, bool) {
	run := r.runs[a.topic]
	switch {
	case !a.correct:
		run = 0
	case a.chance <= rating.CorridorMiddle && !a.hint:
		run++
	}
	if a.inTopic >= profile.MasteryAnswers && run >= r.length {
		r.runs[a.topic] = 0
		return a.level, true
	}
	r.runs[a.topic] = run
	return "", false
}

// cautious is the test of a cautious estimate: after a right and unaided
// answer, with enough answers in the topic, the topic is mastered at the
// highest of its levels, at or below the task's, whose middle task a child at
// the estimate less z times its uncertainty answers at the corridor's middle
// or better. An infinite z masters nothing.
type cautious struct {
	z float64
}

func (c cautious) passed(a *masteryAnswer) (rating.GradeLevel, bool) {
	if !a.correct || a.hint || a.inTopic < profile.MasteryAnswers {
		return "", false
	}
	low := a.estimate - c.z*math.Sqrt(uncertaintyOf(a.answers, a.inTopic))
	for i := len(a.levels) - 1; i >= 0; i-- {
		level := a.levels[i]
		if level.Shift() > a.level.Shift() {
			continue
		}
		if rating.Probability(low, middleTaskOf(level)) >= rating.CorridorMiddle {
			return level, true
		}
	}
	return "", false
}

// uncertaintyOf is how uncertain the level in a topic is after so many
// answers in all and in the topic, from the counts alone: the overall level's
// uncertainty, the spread the trial series starts from narrowed by what every
// answer tells, plus the topic's, the uncertainty the service's step reads its
// first step from narrowed by the topic's answers. It overstates the
// uncertainty a little, since a topic's answers tell of the overall level too,
// which suits a cautious test.
func uncertaintyOf(answers, inTopic int) float64 {
	overall := 1 / (1/(trialSpread*trialSpread) + answerInformation*float64(answers))
	topic := 1 / (answerInformation/serviceDecay + answerInformation*float64(inTopic))
	return overall + topic
}

// middleTaskOf is where the middle task of a level stands on the scale.
func middleTaskOf(level rating.GradeLevel) float64 {
	return rating.Point{GradeLevel: level, Difficulty: middleDifficulty}.Beta()
}

// middleDifficulty is the difficulty of a level's middle task.
const middleDifficulty = 3

// Wald's sequential test of whether a child is mastered at a level: the
// chance on the level's middle task at most the corridor's bottom, against at
// least its top. A false "mastered" is let through at most alpha of the time,
// a true one missed at most beta of it.
const (
	waldAlpha = 0.05
	waldBeta  = 0.2
)

// The bounds a test's sum is held between: past the upper the child is
// mastered, under the lower the test starts again.
var (
	waldUpper = math.Log((1 - waldBeta) / waldAlpha)
	waldLower = math.Log(waldBeta / (1 - waldAlpha))
)

// wald is Wald's sequential test run at every level of a topic at once: each
// level keeps a sum of how much likelier every answer was for a child at the
// top of the corridor on its middle task than at its bottom, taking the
// task's own difficulty in. Every wrong answer and every right and unaided
// one counts; a sum that falls under the lower bound starts again. After a
// right and unaided answer, with enough answers in the topic, the topic is
// mastered at the highest level at or below the task's whose sum is past the
// upper bound, and that level's sum and those below it are spent. The sums
// are kept in the levels' order, so that every run comes to the same bits.
type wald struct {
	sums map[string][]float64
}

func newWald() *wald { return &wald{sums: map[string][]float64{}} }

func (w *wald) passed(a *masteryAnswer) (rating.GradeLevel, bool) {
	if a.correct && a.hint {
		return "", false
	}
	sums, kept := w.sums[a.topic]
	if !kept {
		sums = make([]float64, len(a.levels))
		w.sums[a.topic] = sums
	}
	for i, level := range a.levels {
		sums[i] = waldStep(sums[i], level, a.beta, a.correct)
		if sums[i] <= waldLower {
			sums[i] = 0
		}
	}
	if !a.correct || a.inTopic < profile.MasteryAnswers {
		return "", false
	}
	for i := len(a.levels) - 1; i >= 0; i-- {
		if a.levels[i].Shift() <= a.level.Shift() && sums[i] >= waldUpper {
			clear(sums[:i+1])
			return a.levels[i], true
		}
	}
	return "", false
}

// waldStep is a test's sum at a level after an answer to a task of
// difficulty beta: what the answer adds is the logarithm of how much likelier
// it was for a child whose chance on the level's middle task is the
// corridor's top than for one whose chance there is its bottom.
func waldStep(sum float64, level rating.GradeLevel, beta float64, correct bool) float64 {
	bottom := rating.Probability(middleTaskOf(level)+levelAt(rating.CorridorLow), beta)
	top := rating.Probability(middleTaskOf(level)+levelAt(rating.CorridorHigh), beta)
	if correct {
		return sum + math.Log(top/bottom)
	}
	return sum + math.Log((1-top)/(1-bottom))
}

// levelAt is how far above a task a child stands whose chance on it is p.
func levelAt(p float64) float64 {
	share := (p - rating.Guess) / (1 - rating.Guess)
	return math.Log(share / (1 - share))
}

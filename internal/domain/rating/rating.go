// Package rating puts a child's level and a task's difficulty on one scale and
// moves the level after every answer.
//
// The scale is the distance between the two: a task at the child's own level
// is an even chance, a level above is a stretch, a level below is a warm-up.
// One number is the child's level overall, a second corrects it for a topic —
// a child may count well and reason about parity badly — and a task's
// difficulty is the third point on the same line. Nothing here is stored or
// loaded: a caller hands in the numbers it holds and is handed back the
// numbers that replace them.
//
// A task is a grade level and a difficulty inside it, and the fifteen of them
// stand on that one line as a ladder. The levels overlap, so a child who finds
// the tasks of their grade too hard has easier ones to go down to, and one who
// finds them easy has harder ones above; the grade decides only where a child
// starts on it.
//
// Every rating moves the same way: by the distance between what was expected
// and what happened, narrowed as answers accumulate so that a settled rating
// is not thrown by one bad day. The first answers are the exception. The start
// is only a guess made from the grade, so until there are enough answers the
// level is estimated from all of them at once, and a child placed too high
// meets easier tasks straight away rather than after a run of failures. Only
// correctness enters either. A hint, a slow answer, a run of failures — all of
// it is recorded elsewhere and changes what the child is asked next, never
// what the rating says the child knows.
//
// A task's difficulty is never updated. A task is written for one child and
// handed out once, so there is no second child whose answer could sharpen the
// estimate, and the difficulty is taken from the task and discarded.
package rating

import "math"

// Guess is the floor under the chance of a correct answer: there are five
// options, and nothing is lost by picking one of them at random.
const Guess = 0.2

const (
	// k0Theta and k0Delta are how far the very first answer may move the
	// overall level and the correction for one topic. The topic moves faster
	// because it starts at the overall level and has to find its own place.
	// They differ for a measured reason: one value for both moved the level in
	// a topic by nearly a whole difficulty after two failures in a row.
	k0Theta = 0.2
	k0Delta = 0.4

	// decay narrows every step as answers accumulate: k0 / (1 + decay·n).
	decay = 0.05

	// middleDifficulty is the difficulty a level is centred on: at the
	// youngest level it sits at zero on the scale, so that a child starting
	// there meets it as an even chance.
	middleDifficulty = 3
)

// Probability is the chance that a child at this level solves a task of this
// difficulty. Both arrive on the same scale, so what the formula reads is the
// distance between them, and no child ever falls below the guessing floor.
func Probability(level, beta float64) float64 {
	return Guess + (1-Guess)*sigmoid(level-beta)
}

// State is what a child's ratings say before an answer: the level overall, the
// correction for the topic this task belongs to, and how many answers each of
// the two already rests on.
type State struct {
	// Theta is the child's level across every topic.
	Theta float64
	// Delta corrects that level for one topic, and is zero for a topic the
	// child has never met — a new topic therefore starts from what the child
	// can do generally.
	Delta float64
	// Answers is how many answers the child has given in total.
	Answers int
	// TopicAnswers is how many of them were in this topic.
	TopicAnswers int
}

// Level is where the child stands in this topic: the overall level with the
// topic's correction applied. It is what a probability is computed against.
func (s State) Level() float64 { return s.Theta + s.Delta }

// Result is what one answer leaves behind: the state that replaces the one it
// was given, and the numbers that produced it, so that a caller can record or
// show them without computing anything a second time.
type Result struct {
	// State is what to store. Both levels have moved, and this answer is
	// counted in both totals — the counts are what narrow the next step, so
	// they are carried here rather than left to whoever calls.
	State
	// Probability is the chance of a correct answer as it stood before the
	// answer arrived — not after, because the answer is what it was measured
	// against.
	Probability float64
	// KTheta and KDelta are how far this answer was allowed to move each level.
	KTheta float64
	KDelta float64
}

// Update moves both levels by one answer to a task of difficulty beta.
func Update(s State, beta float64, correct bool) Result {
	probability := Probability(s.Level(), beta)

	// How far the answer was from what was expected, and in which direction.
	// An easy task answered correctly says almost nothing and moves almost
	// nothing; the same task answered wrongly moves a great deal.
	score := 0.0
	if correct {
		score = 1
	}
	surprise := score - probability

	kTheta, kDelta := step(k0Theta, s.Answers), step(k0Delta, s.TopicAnswers)
	return Result{
		State: State{
			Theta:        s.Theta + kTheta*surprise,
			Delta:        s.Delta + kDelta*surprise,
			Answers:      s.Answers + 1,
			TopicAnswers: s.TopicAnswers + 1,
		},
		Probability: probability,
		KTheta:      kTheta,
		KDelta:      kDelta,
	}
}

// step is how far one answer may move a level: wide at first, so that a new
// child's rating finds its place within a handful of answers, and narrower
// afterwards, so that a rating resting on fifty answers stays where it is. A
// count below zero is no count at all, rather than a step without end.
func step(k0 float64, answers int) float64 { return k0 / (1 + decay*float64(max(answers, 0))) }

func sigmoid(x float64) float64 { return 1 / (1 + math.Exp(-x)) }

// logit reads that curve backwards: how far above a difficulty a child has to
// stand for the chance of a correct answer to be exactly p. It is defined
// above the guessing floor, which is where the corridor lies.
func logit(p float64) float64 {
	above := (p - Guess) / (1 - Guess)
	return math.Log(above / (1 - above))
}

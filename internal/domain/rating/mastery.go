package rating

import "math"

// A topic counts as mastered when the child is sure to stand where its tasks
// are: their level in it less how far that level may be off — the cautious
// level — has to answer the middle task of a level at the corridor's middle or
// better. A run of right answers says less than it seems to, since a child
// whose true chance is 60 % gives three in a row in one try of five; a level
// read with its margin does not.
//
// How far a level may be off is counted from the answers alone and never
// stored. The overall level starts as uncertain as the trial series lets a
// child stand from the start, and every answer narrows it by what an answer at
// the corridor's middle tells; a topic's correction starts as uncertain as
// twenty such answers would leave it — the step's decay read as an
// uncertainty — and the topic's own answers narrow it. The two are added,
// which overstates a little, since a topic's answers tell of the overall level
// too: a cautious reading can afford it.

// answerInformation is what one answer at the corridor's middle tells of a
// level, where the rule keeps a child's tasks.
var answerInformation = informationAt(CorridorMiddle)

// informationAt is what an answer given at chance p tells of a level: how far
// the chance rises with the level there, squared, over the variance of the
// answer — worked out as how far a surprise of one moves the level for every
// unit of uncertainty, times that rise, so that nothing is divided by the
// variance itself.
func informationAt(p float64) float64 {
	gain := (p - Guess) / (p * (1 - Guess))
	slope := (p - Guess) * (1 - p) / (1 - Guess)
	return gain * slope
}

// Uncertainty is how far a child's level in a topic may be off — a variance,
// in the scale's own units — after so many answers in all and in the topic.
// A count below zero is no count at all.
func Uncertainty(answers, topicAnswers int) float64 {
	overall := 1 / (1/(startSpread*startSpread) + answerInformation*float64(max(answers, 0)))
	topic := 1 / (answerInformation/decay + answerInformation*float64(max(topicAnswers, 0)))
	return overall + topic
}

// MasteredAt is the level a topic counts as mastered at after a right and
// unaided answer: the highest of the levels it is taught at, at or below the
// level of the task answered, whose middle task the child at the cautious
// level would answer at the corridor's middle or better. level is where the
// child stands in the topic once the answer has moved it, and answers and
// topicAnswers count that answer too. It is false when no such level clears
// the bar. The order the levels are given in decides nothing.
func MasteredAt(level float64, answers, topicAnswers int, task GradeLevel, taught []GradeLevel) (GradeLevel, bool) {
	cautious := level - math.Sqrt(Uncertainty(answers, topicAnswers))
	var held GradeLevel
	clears := false
	for _, candidate := range taught {
		if candidate.Shift() > task.Shift() || (clears && candidate.Shift() <= held.Shift()) {
			continue
		}
		middle := Point{GradeLevel: candidate, Difficulty: middleDifficulty}
		if Probability(cautious, middle.Beta()) >= CorridorMiddle {
			held, clears = candidate, true
		}
	}
	return held, clears
}

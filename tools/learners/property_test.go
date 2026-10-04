package main

import (
	"fmt"
	"math"
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/profile"
	"github.com/MathTrail/mathtrail-standalone/internal/domain/rating"
)

// The chosen model is what the service runs, not a copy that happens to agree
// on the children the bench draws: on any sequence of answers — any topics in
// any order, tasks at any point of their ladders, right or wrong, with the
// hint or without — it stands where the service stands after every answer,
// overall and in every topic, and holds the masteries the service holds,
// declared and lost on the same answers.

// sequenceLength is how many answers a sequence gives: past the 61st, where
// the floor under the overall step binds, and enough in each of its topics to
// master them and lose them again.
const sequenceLength = 150

// scriptedAnswer is one answer of a sequence: which of its topics, which point
// of the topic's ladder, the draw that says whether it was right, and whether
// it took the hint.
type scriptedAnswer struct {
	topic, point int
	draw         float64
	hint         bool
}

// sequence is a child's grade, the topics it is set, how often it is right,
// and its answers.
type sequence struct {
	grade   int
	topics  []int
	skill   float64
	answers []scriptedAnswer
}

func TestTheChosenModelIsTheProductOnAnySequenceOfAnswers(t *testing.T) {
	t.Parallel()
	w, err := newWorld(sequenceLength)
	if err != nil {
		t.Fatal(err)
	}
	model, err := chosenModel()
	if err != nil {
		t.Fatal(err)
	}
	declared, lost := 0, 0
	properties := gopter.NewProperties(nil)
	properties.Property("after every answer the model stands where the service stands and holds what it holds", prop.ForAll(
		func(seq sequence) string {
			why, verdicts := sameAnswerByAnswer(w, model, &seq)
			declared += verdicts.declared
			lost += verdicts.lost
			return why
		},
		genSequence(len(w.topics)),
	))
	properties.TestingRun(t)
	if declared == 0 || lost == 0 {
		t.Errorf("the sequences declared %d masteries and lost %d, want some of both, or the property shows nothing of mastery", declared, lost)
	}
}

// verdicts counts the masteries the service declared and lost over a sequence.
type verdicts struct {
	declared, lost int
}

// sameAnswerByAnswer runs a sequence under the service and under the model,
// side by side, and says how the model first parted from the service, or
// nothing when it never did.
func sameAnswerByAnswer(w *world, model *rule, seq *sequence) (string, verdicts) {
	var counted verdicts
	theService := newSession(w, serviceRule(), &child{grade: seq.grade})
	theModel := newSession(w, model, &child{grade: seq.grade})
	for k, a := range seq.answers {
		topic := w.topics[seq.topics[a.topic]]
		ladder := w.ladders[topic]
		point := ladder[a.point%len(ladder)]
		correct := a.draw < seq.skill
		theirs, err := answerScripted(theService, k, topic, point, correct, a.hint)
		if err != nil {
			return err.Error(), counted
		}
		ours, err := answerScripted(theModel, k, topic, point, correct, a.hint)
		if err != nil {
			return err.Error(), counted
		}
		if theirs.Mastered {
			counted.declared++
		}
		if theirs.Unmastered {
			counted.lost++
		}
		if why := partedAt(w, theService, theModel, &theirs, &ours); why != "" {
			return fmt.Sprintf("answer %d, %s at %s difficulty %d, right %v, hint %v: %s",
				k+1, topic, point.GradeLevel, point.Difficulty, correct, a.hint, why), counted
		}
	}
	return "", counted
}

// answerScripted gives a session one answer of a sequence as a lesson takes
// it once the task is set: the rule's levels written where the service reads
// them, the task issued at the point given rather than the one the service
// would choose, the answer recorded and taken in, and the clock moved on.
func answerScripted(s *session, k int, topic string, point rating.Point, correct, hint bool) (profile.Recorded, error) {
	if s.writes(k) {
		s.writeLevels()
	}
	brief := profile.Brief{TargetConcept: topic, GradeLevel: point.GradeLevel, Difficulty: point.Difficulty}
	recorded, err := s.take(&brief, profile.TutorRule, &answerDraws{}, correct, hint, k)
	if err != nil {
		return profile.Recorded{}, err
	}
	s.advance(k)
	return recorded, nil
}

// partedAt says where the model stands apart from the service after an
// answer: overall, in a topic, in a mastery held, or in what the answer
// declared or lost; nothing when it stands where the service does.
func partedAt(w *world, theService, theModel *session, theirs, ours *profile.Recorded) string {
	if got, want := theModel.est.overall(), theService.p.Ratings.Theta; math.Abs(got-want) > 1e-12 {
		return fmt.Sprintf("the overall level is %v, the service's %v", got, want)
	}
	for _, topic := range w.topics {
		if got, want := theModel.est.level(topic), theService.p.LevelIn(topic); math.Abs(got-want) > 1e-12 {
			return fmt.Sprintf("the level in %s is %v, the service's %v", topic, got, want)
		}
		if got, want := theModel.p.Topics[topic], theService.p.Topics[topic]; !sameMasteryHeld(&got, &want) {
			return fmt.Sprintf("%s is held mastered at %v since %v, the service's at %v since %v",
				topic, got.MasteredLevel, got.MasteredSince, want.MasteredLevel, want.MasteredSince)
		}
	}
	if ours.Mastered != theirs.Mastered || ours.Unmastered != theirs.Unmastered {
		return fmt.Sprintf("the answer declared %v and lost %v, the service's %v and %v",
			ours.Mastered, ours.Unmastered, theirs.Mastered, theirs.Unmastered)
	}
	return ""
}

// sameMasteryHeld says whether two topics are held mastered at the same level
// since the same day, or neither is.
func sameMasteryHeld(a, b *profile.Topic) bool {
	if (a.MasteredLevel == nil) != (b.MasteredLevel == nil) || (a.MasteredSince == nil) != (b.MasteredSince == nil) {
		return false
	}
	return (a.MasteredLevel == nil || *a.MasteredLevel == *b.MasteredLevel) &&
		(a.MasteredSince == nil || a.MasteredSince.Equal(b.MasteredSince.Time))
}

// genSequence is a child of any grade set three of the catalog's topics, a
// topic perhaps twice, right as often as anything from three times in ten to
// almost always.
func genSequence(topics int) gopter.Gen {
	grades := gen.IntRange(profile.MinGrade, profile.MaxGrade)
	pools := gen.SliceOfN(3, gen.IntRange(0, topics-1))
	skills := gen.Float64Range(0.3, 0.95)
	answers := gen.SliceOfN(sequenceLength, genScriptedAnswer())
	return gopter.CombineGens(grades, pools, skills, answers).Map(func(values []any) sequence {
		grade, _ := values[0].(int)
		pool, _ := values[1].([]int)
		skill, _ := values[2].(float64)
		given, _ := values[3].([]scriptedAnswer)
		return sequence{grade: grade, topics: pool, skill: skill, answers: given}
	})
}

// genScriptedAnswer is an answer to a task of one of the three topics, at any
// point of its ladder, with the hint taken one time in five.
func genScriptedAnswer() gopter.Gen {
	places, points, draws, hints := gen.IntRange(0, 2), gen.IntRange(0, 14), gen.Float64Range(0, 1), gen.IntRange(0, 4)
	return gopter.CombineGens(places, points, draws, hints).Map(func(values []any) scriptedAnswer {
		place, _ := values[0].(int)
		point, _ := values[1].(int)
		draw, _ := values[2].(float64)
		hint, _ := values[3].(int)
		return scriptedAnswer{topic: place, point: point, draw: draw, hint: hint == 0}
	})
}

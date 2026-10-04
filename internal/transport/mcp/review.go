package mcpserver

import (
	"fmt"
	"strings"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/progress"
)

// reviewOut is the review of where the child stands, for the adult, as a card
// and a model read it: topics and traps by their catalog ids, and why each is
// named, and what each step advises, by codes. There are no words in it: the
// card says them in its own language, and the model has the review in words.
// During the trial series there is none.
type reviewOut struct {
	Strong  []judgedOut `json:"strong"`
	Develop []judgedOut `json:"develop"`
	Early   []string    `json:"early"`
	Steps   []stepOut   `json:"steps"`
}

// judgedOut is a topic the review names: its catalog id, why it is named, the
// trap its answers keep falling for where that is one of the reasons, and, for
// one to develop, whether its last answer was right.
type judgedOut struct {
	Topic   string   `json:"topic"`
	Reasons []string `json:"reasons"`
	Trap    string   `json:"trap,omitempty"`
	Moving  bool     `json:"moving,omitempty"`
}

// stepOut is one step of the review: what it advises, the topic it is for — none
// for the step that holds for every topic — and the trap whose advice it is.
type stepOut struct {
	Kind  string `json:"kind"`
	Topic string `json:"topic,omitempty"`
	Trap  string `json:"trap,omitempty"`
}

func reviewOf(review *progress.Review) *reviewOut {
	if review == nil {
		return nil
	}
	steps := make([]stepOut, 0, len(review.Steps))
	for _, step := range review.Steps {
		steps = append(steps, stepOut{Kind: string(step.Kind), Topic: step.Topic, Trap: step.Trap})
	}
	return &reviewOut{
		Strong:  judgedOf(review.Strong),
		Develop: judgedOf(review.Develop),
		Early:   append([]string{}, review.Early...),
		Steps:   steps,
	}
}

func judgedOf(topics []progress.Judged) []judgedOut {
	out := make([]judgedOut, 0, len(topics))
	for _, topic := range topics {
		reasons := make([]string, 0, len(topic.Reasons))
		for _, reason := range topic.Reasons {
			reasons = append(reasons, string(reason))
		}
		out = append(out, judgedOut{Topic: topic.Topic, Reasons: reasons, Trap: topic.Trap, Moving: topic.Moving})
	}
	return out
}

// reviewText is the review in words, for the adult: the strong topics and
// why, the ones to develop and what keeps them back, the ones too early to
// judge — which is no verdict on them — and the steps to take, each with its
// advice. Without a card these words are the review; during the trial series
// there is none to tell.
func (s *Service) reviewText(review *progress.Review) string {
	if review == nil {
		return ""
	}
	var parts []string
	if len(review.Strong) > 0 {
		parts = append(parts, "Strong: "+s.judgedText(review.Strong)+".")
	}
	if len(review.Develop) > 0 {
		parts = append(parts, "To develop: "+s.judgedText(review.Develop)+".")
	}
	if len(review.Early) > 0 {
		names := make([]string, 0, len(review.Early))
		for _, id := range review.Early {
			names = append(names, s.topicName(id))
		}
		parts = append(parts, "Too early to judge, which is no verdict yet: "+strings.Join(names, ", ")+".")
	}
	if len(review.Steps) > 0 {
		parts = append(parts, "What to do next: "+s.stepsText(review.Steps))
	}
	if len(parts) == 0 {
		return "Review for the adult: no topic stands apart from the overall level yet."
	}
	return "Review for the adult. " + strings.Join(parts, " ")
}

// reasonWords are the words of each reason a topic is named for. A trap's are
// written with the trap's own description, by judgedText.
var reasonWords = map[progress.Reason]string{
	progress.ReasonMastered: "mastered",
	progress.ReasonHigh:     "well above the overall level",
	progress.ReasonRose:     "rose this week",
	progress.ReasonLow:      "well below the overall level",
	progress.ReasonFailures: "answered wrongly several times in a row",
	progress.ReasonHints:    "the hint used in half its latest answers",
	progress.ReasonFell:     "fell this week",
}

// judgedText names each topic with why it is named, and, for one whose last
// answer was right, that a move has begun.
func (s *Service) judgedText(topics []progress.Judged) string {
	named := make([]string, 0, len(topics))
	for _, topic := range topics {
		said := make([]string, 0, len(topic.Reasons)+1)
		for _, reason := range topic.Reasons {
			if reason == progress.ReasonTrap {
				said = append(said, "keeps falling for this: "+s.trapDescribed(topic.Trap))
				continue
			}
			said = append(said, reasonWords[reason])
		}
		if topic.Moving {
			said = append(said, "its last answer was right, so a move has begun")
		}
		named = append(named, fmt.Sprintf("%s (%s)", s.topicName(topic.Topic), strings.Join(said, "; ")))
	}
	return strings.Join(named, ", ")
}

// stepsText is the steps of the review, numbered, each with its advice.
func (s *Service) stepsText(steps []progress.Step) string {
	said := make([]string, 0, len(steps))
	for place, step := range steps {
		advice := s.stepAdvice(step)
		if step.Topic != "" {
			advice = s.topicName(step.Topic) + ": " + advice
		}
		said = append(said, fmt.Sprintf("%d. %s", place+1, advice))
	}
	return strings.Join(said, " ")
}

// stepAdvice is what a step advises, in a sentence: a trap's advice as the
// catalog writes it, or the advice of the step's kind.
func (s *Service) stepAdvice(step progress.Step) string {
	switch step.Kind {
	case progress.StepTrap:
		if advice, ok := s.content.TrapAdvice(step.Trap); ok {
			return advice
		}
		return "watch for this mistake: " + s.trapDescribed(step.Trap) + "."
	case progress.StepRhythm:
		return "two or three short tasks a day, until it comes back."
	case progress.StepPractice:
		return "a few more tasks in it; MathTrail sets them where the child stands."
	default:
		return "try each task first without the hint."
	}
}

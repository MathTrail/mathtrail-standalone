package mcpserver

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/profile"
	"github.com/MathTrail/mathtrail-standalone/internal/domain/progress"
)

// The screens a card can show, as a payload names them. The widget never
// guesses a screen from the shape of the data, because a wrong guess would
// show a child the wrong thing.
const (
	screenFirstRun = "first_run"
	screenProfile  = "profile"
	screenProgress = "progress"
	screenResult   = "result"
	screenTask     = "task"
	screenWaiting  = "waiting"
)

// The statuses a refusal gives itself at the top of its payload: what the model
// can act on, and what marks an answer as a refusal wherever it is read.
const (
	// statusRejected is a call refused for what it asked: arguments that break
	// a rule, a task that failed its checks.
	statusRejected = "rejected"
	// statusLimited is a call refused by a limit of the day or the minute.
	statusLimited = "limited"
	// statusStale is a call about something no longer there: a task handed in
	// for a request that is not open, or an answer to a task not on the card.
	statusStale = "stale"
)

// noArguments is what a tool that reads takes: nothing. The schema made of it
// refuses anything a caller adds, a student id above all — the profile is the
// one the token belongs to.
type noArguments struct{}

// answerLine is the last answer the child gave, in one line. Every result
// carries it: the card records an answer without the model, and a model that
// missed the card's message learns from its next call that the child has
// answered, and how.
type answerLine struct {
	TaskID     string `json:"task_id"`
	Topic      string `json:"topic"`
	Correct    bool   `json:"correct"`
	AnsweredAt string `json:"answered_at"`
}

// lastAnswerOf is the last answer of a profile, or nil when there has been
// none.
func lastAnswerOf(p *profile.Profile) *answerLine {
	answer, answered := p.LastAnswer()
	if !answered {
		return nil
	}
	return &answerLine{
		TaskID:     answer.TaskID,
		Topic:      answer.Topic,
		Correct:    answer.Correct,
		AnsweredAt: moment(answer.AnsweredAt),
	}
}

// childOut are the child's details a card shows, named as save_profile takes
// them: what the progress carries, and what the profile carries beside the
// parent's notes.
type childOut struct {
	Pseudonym      string   `json:"pseudonym"`
	Grade          int      `json:"grade"`
	Interests      []string `json:"interests"`
	ExcludedSkills []string `json:"excluded_skills"`
	UILanguage     *string  `json:"ui_language"`
}

func childOf(s *profile.Student) *childOut {
	return &childOut{
		Pseudonym:      s.Pseudonym,
		Grade:          s.Grade,
		Interests:      append([]string{}, s.Interests...),
		ExcludedSkills: append([]string{}, s.ExcludedSkills...),
		UILanguage:     s.UILanguage,
	}
}

// trialOut is how far the trial series has got: while it runs there is no
// rating to show, only this.
type trialOut struct {
	Answered int `json:"answered"`
	Of       int `json:"of"`
}

func trialOf(trial *progress.Trial) *trialOut {
	if trial == nil {
		return nil
	}
	return &trialOut{Answered: trial.Answered, Of: trial.Of}
}

// recommendationOut is what the rule would set next. The topic is its catalog
// id: a card names it in its own language, and the model has the name in the
// words of the result.
type recommendationOut struct {
	Topic      string `json:"topic"`
	GradeLevel string `json:"grade_level"`
	Difficulty int    `json:"difficulty"`
	Goal       string `json:"goal"`
}

func recommendationOf(next *progress.Recommendation) *recommendationOut {
	return &recommendationOut{
		Topic:      next.Topic,
		GradeLevel: string(next.GradeLevel),
		Difficulty: next.Difficulty,
		Goal:       string(next.Goal),
	}
}

// moment is a moment as a payload writes it: RFC 3339, in UTC.
func moment(t profile.Time) string { return t.UTC().Format(time.RFC3339) }

// The words below are for the model, in English: it relays them in the chat's
// own language, and without a card they are the whole of what the child and
// the adult are shown.

// quoted is a text the parent wrote, as the words for the model carry it: a
// string of JSON, so that it stands apart from every sentence of the
// service's and cannot pass for one. What the parent can see stays as they
// wrote it, the joiners of a script or an emoji among it; a quotation mark or a
// backslash inside is escaped.
func quoted(text string) string {
	var out strings.Builder
	encoder := json.NewEncoder(&out)
	encoder.SetEscapeHTML(false)
	_ = encoder.Encode(text) // a string always encodes
	return strings.TrimSuffix(out.String(), "\n")
}

// quotedEach is every text of a list the parent wrote, each quoted.
func quotedEach(texts []string) []string {
	quotes := make([]string, 0, len(texts))
	for _, text := range texts {
		quotes = append(quotes, quoted(text))
	}
	return quotes
}

// firstRunText is what the model is told when there is no profile. The
// service keeps nothing of its own, so it cannot tell a first sign-in from a
// profile deleted for good or kept in another Google account, and says all
// three.
const firstRunText = "There is no profile yet. If the adult made one before, it has been deleted for good, " +
	"or they signed in with another Google account. Ask the adult for a pseudonym — never the child's real name — " +
	"and the school grade from 1 to 6. Interests, skills the child has not met at school yet, " +
	"notes about the child and the language of the cards are optional. Then create the profile with save_profile."

// topicName is what the catalog calls a topic, or its id when the catalog no
// longer has it.
func (s *Service) topicName(id string) string {
	if topic, found := s.content.Topic(id); found {
		return topic.Name
	}
	return id
}

// topicDescribed is a topic's name with the catalog's own line about it. A
// name alone can be read two ways — "Ordering" relayed as putting numbers in
// order, where the topic is restoring an order from comparisons — so where the
// words tell what comes next, or which topics have been met, the model is
// given what the topic is as well as what it is called.
func (s *Service) topicDescribed(id string) string {
	topic, found := s.content.Topic(id)
	if !found {
		return id
	}
	return topic.Name + " (" + strings.TrimSuffix(topic.Description, ".") + ")"
}

// skillName is how the catalog describes a skill, as an item of a list, or its
// id when the catalog no longer has it: a parent may have typed one into the
// file by hand, and the words must not say "none" where the card lists one.
func (s *Service) skillName(id string) string {
	if skill, found := s.content.Skill(id); found {
		// A description is a sentence of its own; in a list it is an item.
		return strings.TrimSuffix(skill.Description, ".")
	}
	return id
}

// trapDescribed is what the catalog says of a trap, as an item of a list, or
// its id when the catalog has no such trap.
func (s *Service) trapDescribed(id string) string {
	if description, found := s.content.TrapDescription(id); found {
		return strings.TrimSuffix(description, ".")
	}
	return id
}

// trialLine says how far the trial series has got and what comes after it, or
// nothing once it is over.
func trialLine(trial *progress.Trial) string {
	if trial == nil {
		return ""
	}
	return fmt.Sprintf("The trial series is under way: %d of %d tasks done. "+
		"It finds where the child stands; the rating comes after it.", trial.Answered, trial.Of)
}

// howItWent is how an answer went, in one word.
func howItWent(correct bool) string {
	if correct {
		return "right"
	}
	return "wrong"
}

// lastAnswerText says what the child last answered, or nothing when they have
// not answered yet.
func (s *Service) lastAnswerText(p *profile.Profile) string {
	answer, answered := p.LastAnswer()
	if !answered {
		return ""
	}
	return fmt.Sprintf("The last recorded answer, to task %s in %s, was %s.",
		answer.TaskID, s.topicName(answer.Topic), howItWent(answer.Correct))
}

// nextText says what the rule would set next, and why now.
func (s *Service) nextText(next *progress.Recommendation) string {
	why := "as something new"
	if next.Goal == profile.GoalReinforce {
		why = "to go over it again after a mistake"
	}
	return fmt.Sprintf("Next, the rule suggests %s at level %s, difficulty %d, %s.",
		s.topicDescribed(next.Topic), next.GradeLevel, next.Difficulty, why)
}

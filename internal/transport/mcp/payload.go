package mcpserver

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/profile"
	"github.com/MathTrail/mathtrail-standalone/internal/domain/progress"
	"github.com/MathTrail/mathtrail-standalone/internal/store"
)

// The screens a card can show, as a payload names them. The widget never
// guesses a screen from the shape of the data, because a wrong guess would
// show a child the wrong thing.
const (
	// screenComing is a task on its way: the card waits for it, and turns into
	// it once it is handed out.
	screenComing   = "coming"
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
	TaskID  string `json:"task_id"`
	Topic   string `json:"topic"`
	Correct bool   `json:"correct"`
}

// lastAnswerOf is the last answer of a profile, or nil when there has been
// none.
func lastAnswerOf(p *profile.Profile) *answerLine {
	answer, answered := p.LastAnswer()
	if !answered {
		return nil
	}
	return &answerLine{TaskID: answer.TaskID, Topic: answer.Topic, Correct: answer.Correct}
}

// problemOut is one field that broke a rule: the field, the code of the rule,
// which a card says in words of its own language, and the rule in words for
// the model. It never repeats what the field held.
type problemOut struct {
	Field string `json:"field"`
	Code  string `json:"code"`
	Rule  string `json:"rule"`
}

// problemsOf are problems as a payload carries them, and as the words for the
// model carry them: one line, the problems set apart by semicolons.
func problemsOf(problems []profile.Problem) (out []problemOut, lines string) {
	out = make([]problemOut, 0, len(problems))
	broken := make([]string, 0, len(problems))
	for _, problem := range problems {
		out = append(out, problemOut{Field: problem.Field, Code: problem.Code, Rule: problem.Rule})
		broken = append(broken, problem.String())
	}
	return out, strings.Join(broken, "; ")
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
	// Country and Region are where the family lives, by their codes, when
	// the adult said; null when they did not.
	Country *string `json:"country"`
	Region  *string `json:"region"`
	// LessonTopic is the topic the child or the adult keeps the lessons to,
	// by its id, as the file names it; null while the rule chooses.
	LessonTopic *string `json:"lesson_topic"`
	// SignInCountryOff says the country the adult signs in from is left out
	// of what is counted.
	SignInCountryOff bool `json:"signin_country_off"`
}

func childOf(s *profile.Student) *childOut {
	return &childOut{
		Pseudonym:        s.Pseudonym,
		Grade:            s.Grade,
		Interests:        append([]string{}, s.Interests...),
		ExcludedSkills:   append([]string{}, s.ExcludedSkills...),
		UILanguage:       s.UILanguage,
		Country:          given(s.Country),
		Region:           given(s.Region),
		LessonTopic:      given(s.LessonTopic),
		SignInCountryOff: s.SignInCountryOff,
	}
}

// given is a text that was given, or nil for one that was not.
func given(text string) *string {
	if text == "" {
		return nil
	}
	return &text
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
	// Chosen says that the topic is the one the child or the adult keeps the
	// lessons to, rather than the rule's own; left out when it is the rule's.
	Chosen bool `json:"chosen,omitempty"`
}

func recommendationOf(next *progress.Recommendation) *recommendationOut {
	return &recommendationOut{
		Topic:      next.Topic,
		GradeLevel: string(next.GradeLevel),
		Difficulty: next.Difficulty,
		Goal:       string(next.Goal),
		Chosen:     next.Chosen,
	}
}

// locationOut is where the adult finds the child's profile for themselves:
// the file is the export, and there is no other. Files beside it that hold a
// profile too are named, for the adult to look at and delete.
type locationOut struct {
	Folder string         `json:"folder"`
	File   string         `json:"file"`
	Link   string         `json:"link"`
	Others []elsewhereOut `json:"others"`
}

// elsewhereOut is another file that holds a profile.
type elsewhereOut struct {
	File string `json:"file"`
	Link string `json:"link"`
}

// whereKept is where the profile's file is, as a payload and the words carry
// it. When Drive cannot say just now — it failed, or took too long — it is no
// location and a sentence saying so: what was read is read already, and the
// name of a folder is no reason to withhold it. Nothing is logged here, since
// the store's own line of the call that failed says so. What is no failure of
// Drive's is told as it is told everywhere: a profile gone since it was read,
// by reading it again; access taken back or ending, by asking for the sign-in
// again; and a call given up on, by its end.
func (s *Service) whereKept(ctx context.Context, account store.Account) (*locationOut, string, error) {
	location, err := s.store.Export(ctx, account)
	switch {
	case err == nil:
		return locationOf(&location), locationText(&location), nil
	case ctx.Err() != nil, errors.Is(err, store.ErrNotFound),
		errors.Is(err, store.ErrAccessRevoked), errors.Is(err, store.ErrAccessExpired):
		return nil, "", fmt.Errorf("mcp: find the profile: %w", err)
	}
	return nil, locationUnknown, nil
}

// locationUnknown is what the model is told when Drive could not say where the
// profile's file is.
const locationUnknown = "Where the profile's file is kept could not be found just now: if the adult asks, say so, " +
	"and that asking again later will tell."

// locationOf is where the profile is, as the payload carries it, or nothing
// when it is kept nowhere a person could open it.
func locationOf(location *store.Location) *locationOut {
	if location.File == "" {
		return nil
	}
	out := &locationOut{Folder: location.Folder, File: location.File, Link: location.Link, Others: []elsewhereOut{}}
	for _, other := range location.Others {
		out.Others = append(out.Others, elsewhereOut(other))
	}
	return out
}

// The words below are for the model, in English: it relays them in the chat's
// own language, and without a card they are the whole of what the child and
// the adult are shown.

// quoted is a text the parent wrote, as the words for the model carry it: a
// string of JSON, so that it stands apart from every sentence of the
// service's and cannot pass for one. What the parent can see stays as they
// wrote it, the joiners of a script or an emoji among it; a quotation mark or a
// backslash inside is escaped.
func quoted(text string) string {
	out, _ := encoded(text) // a string always encodes
	return string(out)
}

// encoded is a value as JSON with its characters as they are. The escapes
// that keep <, > and & out of a web page would show a code where the parent
// wrote a sign, and weigh six bytes for one in the measure of a hand-in;
// nothing here is put in a page.
func encoded(value any) (json.RawMessage, error) {
	var out bytes.Buffer
	encoder := json.NewEncoder(&out)
	encoder.SetEscapeHTML(false)
	if err := encoder.Encode(value); err != nil {
		return nil, err
	}
	return bytes.TrimSuffix(out.Bytes(), []byte("\n")), nil
}

// quotedEach is every text of a list the parent wrote, each quoted.
func quotedEach(texts []string) []string {
	quotes := make([]string, 0, len(texts))
	for _, text := range texts {
		quotes = append(quotes, quoted(text))
	}
	return quotes
}

// locationText is where the profile is, in words: the file is the export.
func locationText(location *store.Location) string {
	if location.File == "" {
		return ""
	}
	where := "as the file " + quoted(location.File)
	if location.Folder != "" {
		where += " in the folder " + quoted(location.Folder)
	}
	text := fmt.Sprintf("The profile is kept in the adult's Google Drive %s: %s. That file is the export: "+
		"the adult can open, download or copy it like any other file.", where, location.Link)
	if len(location.Others) == 0 {
		return text
	}
	others := make([]string, 0, len(location.Others))
	for _, other := range location.Others {
		others = append(others, quoted(other.File)+" ("+other.Link+")")
	}
	return text + " Other files in the adult's Drive hold a profile too: " + strings.Join(others, ", ") +
		". MathTrail reads and writes only the newest, the one above, and never merges them; " +
		"the adult can delete the others."
}

// firstRunText is what the model is told when there is no profile. The
// service keeps nothing of its own, so it cannot tell a first sign-in from a
// profile deleted for good or kept in another Google account, and says all
// three. A profile is made in the chat alone — a card only tells the adult to
// ask for one there — so these words are where the adult is asked to say they
// are the child's parent or tutor and told where the profile lives, before any
// detail of the child is asked for.
const firstRunText = "There is no profile yet. If the adult made one before, it has been deleted for good, " +
	"or they signed in with another Google account. First ask the adult to say they are the child's parent or " +
	"tutor, unless they have said so already, and tell them the profile is one file in their own Google Drive, " +
	"where the child is known by a pseudonym alone. Then ask for a pseudonym — never the child's real name, birth date, age " +
	"or school — and the school grade from 1 to 6. Interests, skills the child has not met at school yet, notes " +
	"about the child and the language of the lessons are optional. Then create the profile with save_profile."

// onTheCardText is what the model is told of a task on the child's card,
// accepted or handed in again, followed by the task in words. Where a card
// shows the task, the child answers there and the card records the answer: a
// model told only to record it asks for the letter in the chat, and every such
// message spends one of the adult's chat messages on an answer the card
// already has. It adds nothing of its own until the child answers or the adult
// asks, since a word about why the task came, or about the way to solve it,
// gives the task away. Without a card the task is read out for the adult to
// read to the child, and the answer the adult gives in the chat is recorded. A
// host may keep only the start of the instructions, so the words that come
// with the task are where these rules are sure to be read.
const onTheCardText = "Where the card next_task drew shows it, the child answers there, and the card records the " +
	"answer itself: do not ask for the answer in the chat, which would cost the adult a message, and add nothing " +
	"of your own about the task — not its topic, not why it came, not how to solve it — until the child has " +
	"answered or the adult asks about it. Without a card, or if the adult says the card shows no task, read out " +
	"the question and the options A to E below and nothing else, for the adult to read to the child; " +
	"give the hint only when it is asked for, and record with submit_answer the child's answer the adult gives " +
	"in the chat."

// shownByTheCard is what the model is told of a task a card shows by itself.
// A model that says the task is on the card speaks before a card still asking
// how its task stands has it, and the child looks for a task that is not there
// yet; the card needs no word to show it. It holds where a card shows the task
// alone: without one, the task is read out, as the words after it say.
const shownByTheCard = "Where a card shows the task, it shows it by itself: do not announce it."

// forYouAlone is what the model may tell the adult while it writes a task: that
// one is on its way, and nothing of what it was handed to write it from.
const forYouAlone = "Until the task is accepted, tell the adult only that one is on its way: the package, and " +
	"what these words say of past answers, are for you alone."

// noPackageTool is what a model is told to do when it has no tool to fetch the
// package with: a chat keeps the list of tools it began with, and one begun
// before the package had a tool of its own has none to write the task from.
const noPackageTool = "If get_package is not among your tools, this chat has an earlier list of MathTrail's " +
	"tools: ask the adult to start a new chat."

// aboutTheStep is how an explanation is worded. The model talks with the adult,
// and the child hears the explanation as the adult reads it out, so it is
// about the step and addresses nobody. It does not show whether the child is a
// boy or a girl either, which nothing tells the service: he or she shows it,
// and in a language with grammatical gender so does a past-tense sentence about
// what the child did, while the pseudonym and praise of the step in the present
// tense do not.
const aboutTheStep = "Word it about the step, addressing nobody, in short sentences that fit the child's grade, so " +
	"the adult can read it out as it is, and so it does not show whether the child is a boy or a girl: speak of " +
	"the child by the pseudonym, never as he or she, praise the step, not the child, and keep to the present tense."

// lessonLanguageText tells the model the language the parent chose for the
// lessons, when they chose one: a lesson starts with next_task, so the model
// may not have read the profile, and it talks to the adult in that language as
// well as writing the tasks in it.
func lessonLanguageText(student *profile.Student) string {
	chosen, ok := student.ChosenLanguage()
	if !ok {
		return ""
	}
	return fmt.Sprintf("The parent chose %s for the lessons: talk in it, and every task is written in it.", chosen)
}

// stillInText says that the task already asked for keeps the language it was
// asked in, when the parent has chosen another since: its card speaks that
// one, and the language chosen starts with the next task.
func stillInText(p *profile.Profile) string {
	chosen, ok := p.Student.ChosenLanguage()
	if !ok {
		return ""
	}
	var asked string
	switch task := p.InFlight(); {
	case task != nil:
		asked = task.Language
	case p.OpenRequest != nil:
		asked = p.OpenRequest.Language
	}
	if asked == "" || asked == chosen {
		return ""
	}
	return fmt.Sprintf("The task already asked for stays in %s, and its card speaks it; the next one comes in %s.", asked, chosen)
}

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

// nextText says what the rule would set next, and why now — or, when the
// lessons are kept to a topic someone chose, that the next task is on it.
func (s *Service) nextText(next *progress.Recommendation) string {
	if next.Chosen {
		return fmt.Sprintf("Next, the topic chosen for the lessons, %s, at level %s, difficulty %d.",
			s.topicDescribed(next.Topic), next.GradeLevel, next.Difficulty)
	}
	why := "as something new"
	if next.Goal == profile.GoalReinforce {
		why = "to go over it again after a mistake"
	}
	return fmt.Sprintf("Next, the rule suggests %s at level %s, difficulty %d, %s.",
		s.topicDescribed(next.Topic), next.GradeLevel, next.Difficulty, why)
}

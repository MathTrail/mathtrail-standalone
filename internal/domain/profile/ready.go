package profile

import (
	"errors"
	"maps"
	"slices"
	"time"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/rating"
)

var (
	// ErrNoReadyTask means no task is kept ready to be handed out.
	ErrNoReadyTask = errors.New("profile: no task kept ready")
	// ErrReadyTaskKept means a task is kept ready already, so none is written
	// ahead: a second would only wait behind the first.
	ErrReadyTaskKept = errors.New("profile: a task is kept ready already")
)

// ReadyTask is a task written ahead: the model wrote it while the child worked
// on the one before, the checks accepted it, and it waits for the child to ask
// for the next task. It is sealed under the id it will have on the card, so
// handing it out moves it there as it is, and its answer is as far out of
// reach as the answer of the task on the card.
//
// Beside the task it keeps what it was written for — the lesson's language,
// the topic the lessons were kept to and the skills kept out of the tasks — so
// that a task the lesson has moved away from since is let go rather than
// handed out; and how its writing went, for the line about it.
type ReadyTask struct {
	// Attempts is how many hand-ins its writing took, the accepted one among
	// them.
	Attempts int `json:"attempts"`
	// Difficulty is the difficulty of the task inside its level, from
	// MinDifficulty to MaxDifficulty.
	Difficulty int `json:"difficulty"`
	// Drawing is the picture in text, when the task has one.
	Drawing string `json:"drawing,omitempty"`
	// ExcludedSkills are the skills the task was asked to keep out.
	ExcludedSkills []string `json:"excluded_skills"`
	// Fingerprint is the sketch that joins the list of past tasks once the
	// task is handed out.
	Fingerprint string `json:"fingerprint"`
	// GradeLevel is the level the task was written for.
	GradeLevel rating.GradeLevel `json:"grade_level"`
	// Hint is the nudge the child may ask for.
	Hint string `json:"hint"`
	// ID is the id the task will have on the card, which its seal is tied to.
	ID string `json:"id"`
	// InstructionsVersion is the version of the instructions the task was
	// written to.
	InstructionsVersion string `json:"instructions_version"`
	// Language is what the task is written in.
	Language string `json:"language"`
	// LessonTopic is the topic the lessons were kept to when the task was
	// asked for, or none.
	LessonTopic string `json:"lesson_topic,omitempty"`
	// OpenedAt is when its request was opened, which the time its writing took
	// is counted from.
	OpenedAt Time `json:"opened_at"`
	// Options are the ones the child will choose between, keyed by letter.
	Options map[string]string `json:"options"`
	// Sealed is everything that would give the answer away, sealed as the
	// answer of a task on the card is.
	Sealed string `json:"sealed"`
	// Topic is the catalog id of what is being asked.
	Topic string `json:"topic"`
	// TutorMode is who chose the task's topic and difficulty, as its request
	// had it.
	TutorMode TutorMode `json:"tutor_mode"`
	// Wording is the question as the child will read it.
	Wording string `json:"wording"`
	// WrittenAt is when the checks accepted the task.
	WrittenAt Time `json:"written_at"`
}

// AskAhead records a request for the task after the one on the card, which the
// model writes while the child works on that one: a request as Ask records it,
// written ahead, with the topic the lessons are kept to now. The task it
// brings is kept rather than handed out. It refuses while a task is kept
// already.
func (p *Profile) AskAhead(brief *Brief, mode TutorMode, language, lesson string, now time.Time) (*OpenRequest, error) {
	if p.ReadyTask != nil {
		return nil, ErrReadyTaskKept
	}
	request := p.Ask(brief, mode, language, now)
	request.Ahead, request.LessonTopic = true, lesson
	return request, nil
}

// Await makes the task being written ahead one the child waits for: the child
// asked for the next task before it was handed in, so it goes to the card once
// it is. Its window starts again, as a task's asked for now does: the model is
// sent to write it now, and the card waits for it from now on, however long it
// lay open ahead. takenAfter is the task a card asked after, when a card asked.
func (r *OpenRequest) Await(takenAfter string, now time.Time) {
	r.Ahead, r.TakenAfter, r.OpenedAt = false, takenAfter, At(now)
}

// Keep puts the task written for the open request aside, for the child to be
// handed when they ask for the next one: sealed as the task on the card is,
// under the id it will have there, and standing where the request asked. The
// request is closed. Nothing changes if the task cannot be sealed.
func (p *Profile) Keep(written *Written, secret TaskSecret, sealer Sealer, now time.Time) (*ReadyTask, error) {
	request := p.OpenRequest
	if request == nil {
		return nil, ErrNoRequest
	}
	id := TaskIDFor(request.ID)
	sealed, err := sealSecret(sealer, secret, unansweredBinding(p.StudentID, id))
	if err != nil {
		return nil, err
	}
	p.ReadyTask = &ReadyTask{
		Attempts:            request.Attempts + 1,
		Difficulty:          request.Brief.Difficulty,
		Drawing:             written.Drawing,
		ExcludedSkills:      append([]string{}, request.Brief.ExcludedSkills...),
		Fingerprint:         written.Fingerprint,
		GradeLevel:          request.Brief.GradeLevel,
		Hint:                written.Hint,
		ID:                  id,
		InstructionsVersion: written.InstructionsVersion,
		Language:            request.Language,
		LessonTopic:         request.LessonTopic,
		OpenedAt:            request.OpenedAt,
		Options:             maps.Clone(written.Options),
		Sealed:              sealed,
		Topic:               request.Brief.TargetConcept,
		TutorMode:           request.TutorMode,
		Wording:             written.Wording,
		WrittenAt:           At(now),
	}
	p.OpenRequest = nil
	return p.ReadyTask, nil
}

// HandOutReady hands the kept task to the child, who asked for the next one.
// It becomes the task on the card as it was written, under its own id and with
// its own seal, and it is handed out as any task is: the one left on the card
// without an answer is recorded as skipped, its fingerprint joins the tasks
// given, its topic records the day, and the day's count of tasks goes up. It
// is marked kept, and takenAfter is the task a card asked after, when a card
// asked.
func (p *Profile) HandOutReady(takenAfter string, now time.Time) (*CurrentTask, error) {
	ready := p.ReadyTask
	if ready == nil {
		return nil, ErrNoReadyTask
	}
	p.handOut(&CurrentTask{
		Difficulty:          ready.Difficulty,
		Drawing:             ready.Drawing,
		Fingerprint:         ready.Fingerprint,
		GradeLevel:          ready.GradeLevel,
		Hint:                ready.Hint,
		ID:                  ready.ID,
		InstructionsVersion: ready.InstructionsVersion,
		IssuedAt:            At(now),
		Kept:                true,
		Language:            ready.Language,
		Options:             maps.Clone(ready.Options),
		Sealed:              ready.Sealed,
		TakenAfter:          takenAfter,
		Topic:               ready.Topic,
		TutorMode:           ready.TutorMode,
		Wording:             ready.Wording,
	}, now)
	p.ReadyTask = nil
	return p.CurrentTask, nil
}

// DropReady lets go of the kept task, which no longer fits the lesson. Nothing
// else moves: it never reached the child, so it is neither among the tasks
// given nor counted against the day.
func (p *Profile) DropReady() { p.ReadyTask = nil }

// CloseRequest lets go of the open request, whose task nobody will take: a
// task written ahead that no longer fits the lesson. A task handed in for it
// afterwards is refused as one for a request that is over.
func (p *Profile) CloseRequest() { p.OpenRequest = nil }

// Misfit is why a task written ahead no longer fits the lesson, in the words
// the line about letting it go uses, or MisfitNone.
type Misfit string

const (
	// MisfitNone is a task that still fits.
	MisfitNone Misfit = ""
	// MisfitLanguage is a task in another language than the lesson's now.
	MisfitLanguage Misfit = "language"
	// MisfitLessonTopic is a task written while the lessons were kept to
	// another topic, or to none.
	MisfitLessonTopic Misfit = "lesson_topic"
	// MisfitExcludedSkills is a task that was not asked to keep out a skill
	// the tasks keep out now.
	MisfitExcludedSkills Misfit = "excluded_skills"
	// MisfitAsked is a task that stands elsewhere than a person asked for.
	MisfitAsked Misfit = "asked"
	// MisfitInstructions is a task written to another version of the
	// instructions than tasks are written to now.
	MisfitInstructions Misfit = "instructions"
)

// Lesson is the lesson as it stands when a task written ahead is about to be
// handed out or written on: what that task has to agree with.
type Lesson struct {
	// Language is the lesson's language now.
	Language string
	// Topic is the topic the lessons are kept to now, or none.
	Topic string
	// ExcludedSkills are the skills kept out of the tasks now.
	ExcludedSkills []string
	// InstructionsVersion is the version of the instructions tasks are written
	// to now.
	InstructionsVersion string
	// Asked is where a person asked the task to stand this time; its zero
	// value asks for nothing.
	Asked Place
}

// Place is where on the ladder a task stands, or is asked to stand: its
// topic, its level and its difficulty. A place asked for may leave any of them
// out.
type Place struct {
	Topic      string
	GradeLevel rating.GradeLevel
	Difficulty int
}

// holds reports whether a task standing at place stands where p asks: on each
// part p names, and anywhere on the parts it leaves out.
func (p Place) holds(place Place) bool {
	return (p.Topic == "" || p.Topic == place.Topic) &&
		(p.GradeLevel == "" || p.GradeLevel == place.GradeLevel) &&
		(p.Difficulty == 0 || p.Difficulty == place.Difficulty)
}

// Misfit says why the kept task no longer fits the lesson, or MisfitNone. A
// task written to another version of the instructions is let go too: it was
// written to a guide the model no longer reads, and checked by the checks
// that went with it.
func (r *ReadyTask) Misfit(now *Lesson) Misfit {
	if found := misfit(r.Language, r.LessonTopic, r.ExcludedSkills, Place{r.Topic, r.GradeLevel, r.Difficulty}, now); found != MisfitNone {
		return found
	}
	if r.InstructionsVersion != now.InstructionsVersion {
		return MisfitInstructions
	}
	return MisfitNone
}

// Misfit says why the task being written ahead for this request no longer
// fits the lesson, or MisfitNone. Its version of the instructions is not
// asked: nothing has been written to it yet that a newer one would not.
func (r *OpenRequest) Misfit(now *Lesson) Misfit {
	return misfit(r.Language, r.LessonTopic, r.Brief.ExcludedSkills,
		Place{r.Brief.TargetConcept, r.Brief.GradeLevel, r.Brief.Difficulty}, now)
}

// misfit is why a task written ahead, in language, while the lessons were
// kept to lesson, keeping out excluded and standing at place, no longer fits
// the lesson now. A skill kept out since is a misfit, and one let back in is
// not: the task keeps out more than it now has to.
func misfit(language, lesson string, excluded []string, place Place, now *Lesson) Misfit {
	switch {
	case language != now.Language:
		return MisfitLanguage
	case lesson != now.Topic:
		return MisfitLessonTopic
	case slices.ContainsFunc(now.ExcludedSkills, func(skill string) bool { return !slices.Contains(excluded, skill) }):
		return MisfitExcludedSkills
	case !now.Asked.holds(place):
		return MisfitAsked
	}
	return MisfitNone
}

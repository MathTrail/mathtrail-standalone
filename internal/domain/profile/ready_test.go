package profile_test

import (
	"errors"
	"reflect"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/profile"
	"github.com/MathTrail/mathtrail-standalone/internal/domain/rating"
)

// keptAt is when the tests of this file have a task written ahead accepted.
var keptAt = asked.Add(40 * time.Second)

// keepOne has the model write a task ahead on p, on topic, and the checks
// accept it at the second attempt: the request it was written for, and the
// task kept.
func keepOne(t *testing.T, p *profile.Profile, sealer profile.Sealer, topic string) (*profile.OpenRequest, *profile.ReadyTask) {
	t.Helper()

	brief := briefOn(topic)
	request, err := p.AskAhead(&brief, profile.TutorRule, "en", "", asked)
	if err != nil {
		t.Fatalf("AskAhead() error = %v, want nil", err)
	}
	if _, refused := p.Refuse(asked); refused != nil {
		t.Fatalf("Refuse() error = %v, want nil", refused)
	}
	kept, err := p.Keep(written(), secret(), sealer, keptAt)
	if err != nil {
		t.Fatalf("Keep() error = %v, want nil", err)
	}
	return request, kept
}

// A task written ahead is kept, not handed out: it stands where its request
// asked, sealed under the id it will have on the card, and its request is
// closed. Nothing about the child moves — the task has not reached them, so it
// is neither among the tasks given nor counted against the day, and the task
// on the card stays where it is.
func TestATaskWrittenAheadIsKeptUntilTheChildAsks(t *testing.T) {
	t.Parallel()

	p := parseFixture(t, "masha")
	onTheCard, daily := *p.CurrentTask, p.Daily
	fingerprints, topics := slices.Clone(p.TaskFingerprints), p.Topics["counting.gaps"]

	request, kept := keepOne(t, p, newSealer(t), "counting.gaps")

	want := profile.ReadyTask{
		Attempts:            2,
		Difficulty:          3,
		ExcludedSkills:      []string{},
		Fingerprint:         "sketch-of-the-gaps",
		GradeLevel:          rating.Grades12,
		Hint:                written().Hint,
		ID:                  profile.TaskIDFor(request.ID),
		InstructionsVersion: written().InstructionsVersion,
		Language:            "en",
		OpenedAt:            profile.At(asked),
		Options:             written().Options,
		Sealed:              kept.Sealed,
		Topic:               "counting.gaps",
		TutorMode:           profile.TutorRule,
		Wording:             written().Wording,
		WrittenAt:           profile.At(keptAt),
	}
	if !reflect.DeepEqual(*kept, want) {
		t.Errorf("the task kept is %+v, want %+v", *kept, want)
	}
	if p.ReadyTask != kept || p.OpenRequest != nil {
		t.Errorf("the file keeps %v with the request %v open, want the task kept and the request closed", p.ReadyTask, p.OpenRequest)
	}
	if strings.Contains(kept.Sealed, secret().Solution) {
		t.Error("the sealed value carries the solution in the open")
	}
	if !reflect.DeepEqual(*p.CurrentTask, onTheCard) || p.Daily != daily ||
		!slices.Equal(p.TaskFingerprints, fingerprints) || !reflect.DeepEqual(p.Topics["counting.gaps"], topics) {
		t.Error("keeping a task moved the task on the card, the day's count, the tasks given or the topic, want none of them moved")
	}
	if err := p.Validate(); err != nil {
		t.Errorf("Validate() error = %v, want a profile with a task kept to be one", err)
	}
}

// A task kept is handed out as it was written, under its own id and with its
// own seal, so its answer opens on the card as any task's does, and marked as
// kept rather than written while the child waited. It is handed
// out as any task is: the task left on the card without an answer is recorded
// as skipped, the new one's fingerprint joins the tasks given, its topic
// records the day, and the day's count goes up.
func TestAKeptTaskIsHandedOutAsItWasWritten(t *testing.T) {
	t.Parallel()

	sealer := newSealer(t)
	p := parseFixture(t, "masha")
	left := p.CurrentTask.ID
	answers := len(p.Recent)
	_, kept := keepOne(t, p, sealer, "counting.gaps")
	ready := *kept

	handed := keptAt.Add(time.Minute)
	task, err := p.HandOutReady("tsk_on_the_card_before", handed)
	if err != nil {
		t.Fatalf("HandOutReady() error = %v, want nil", err)
	}

	want := profile.CurrentTask{
		Difficulty:          ready.Difficulty,
		Fingerprint:         ready.Fingerprint,
		GradeLevel:          ready.GradeLevel,
		Hint:                ready.Hint,
		ID:                  ready.ID,
		InstructionsVersion: ready.InstructionsVersion,
		IssuedAt:            profile.At(handed),
		Kept:                true,
		Language:            ready.Language,
		Options:             ready.Options,
		Sealed:              ready.Sealed,
		TakenAfter:          "tsk_on_the_card_before",
		Topic:               ready.Topic,
		TutorMode:           ready.TutorMode,
		Wording:             ready.Wording,
	}
	if !reflect.DeepEqual(*task, want) {
		t.Errorf("the task handed out is %+v, want %+v", *task, want)
	}
	if p.CurrentTask != task || p.ReadyTask != nil {
		t.Error("the task handed out is not the one on the card, or a task is still kept")
	}
	opened, err := p.OpenTask(sealer)
	if err != nil {
		t.Fatalf("OpenTask() error = %v, want the kept seal to open on the card", err)
	}
	if opened.Answer != secret().Answer || opened.Solution != secret().Solution {
		t.Errorf("the seal opened to %+v, want the secret it was kept with", opened)
	}

	if len(p.Recent) != answers+1 || !p.Recent[len(p.Recent)-1].Skipped || p.Recent[len(p.Recent)-1].TaskID != left {
		t.Errorf("the window ends with %+v, want %s recorded as skipped", p.Recent[len(p.Recent)-1], left)
	}
	if last := p.TaskFingerprints[len(p.TaskFingerprints)-1]; last != ready.Fingerprint {
		t.Errorf("the last fingerprint is %q, want the kept task's, %q", last, ready.Fingerprint)
	}
	if day := p.Topics["counting.gaps"].LastIssued; !day.Equal(profile.DateOf(handed).Time) {
		t.Errorf("the topic was last given on %v, want %v", day, profile.DateOf(handed))
	}
	if today := p.Daily.Today(handed); today.Accepted != 1 {
		t.Errorf("the day counts %d tasks, want the one handed out", today.Accepted)
	}
	if err := p.Validate(); err != nil {
		t.Errorf("Validate() error = %v, want nil", err)
	}
}

// A task is written ahead only while none is kept: a second would only wait
// behind the first. Asked for anyway, nothing changes.
func TestNoTaskIsWrittenAheadWhileOneIsKept(t *testing.T) {
	t.Parallel()

	p := parseFixture(t, "dima")
	_, kept := keepOne(t, p, newSealer(t), "counting.gaps")

	brief := briefOn("logic.weighing")
	if _, err := p.AskAhead(&brief, profile.TutorRule, "en", "", keptAt); !errors.Is(err, profile.ErrReadyTaskKept) {
		t.Errorf("AskAhead() error = %v, want %v", err, profile.ErrReadyTaskKept)
	}
	if p.ReadyTask != kept || p.OpenRequest != nil {
		t.Error("a refused request ahead changed the task kept or opened a request")
	}
}

// Keeping takes a request to keep the task of, and handing out a task kept
// takes one kept: without them, nothing changes.
func TestNothingIsKeptOrHandedOutFromNothing(t *testing.T) {
	t.Parallel()

	p := parseFixture(t, "dima")
	if _, err := p.Keep(written(), secret(), newSealer(t), keptAt); !errors.Is(err, profile.ErrNoRequest) {
		t.Errorf("Keep() error = %v, want %v", err, profile.ErrNoRequest)
	}
	if _, err := p.HandOutReady("", keptAt); !errors.Is(err, profile.ErrNoReadyTask) {
		t.Errorf("HandOutReady() error = %v, want %v", err, profile.ErrNoReadyTask)
	}
	if p.ReadyTask != nil || p.CurrentTask != nil {
		t.Error("keeping or handing out nothing put a task in the file")
	}
}

// A task written ahead that cannot be sealed is not kept: a task kept with its
// answer in the open is worse than none. Its request stays open, its attempt
// unspent.
func TestATaskThatCannotBeSealedIsNotKept(t *testing.T) {
	t.Parallel()

	p := parseFixture(t, "dima")
	brief := briefOn("counting.gaps")
	request, err := p.AskAhead(&brief, profile.TutorRule, "en", "", asked)
	if err != nil {
		t.Fatalf("AskAhead() error = %v, want nil", err)
	}

	if _, err := p.Keep(written(), secret(), failingSealer{}, keptAt); err == nil {
		t.Fatal("Keep() error = nil, want the sealer's refusal")
	}
	if p.ReadyTask != nil || p.OpenRequest != request || request.Attempts != 0 {
		t.Error("a task that could not be sealed was kept, or its request moved")
	}
}

// Asking for a task to be written now lets the task kept go: the caller has
// decided it is not the one to hand out.
func TestAskingForATaskNowLetsTheKeptOneGo(t *testing.T) {
	t.Parallel()

	p := parseFixture(t, "dima")
	keepOne(t, p, newSealer(t), "counting.gaps")

	brief := briefOn("logic.weighing")
	request := p.Ask(&brief, profile.TutorLLM, "en", keptAt)
	if p.ReadyTask != nil || p.OpenRequest != request {
		t.Error("a request for a task now left the kept task in the file")
	}
}

// A child who asks for the next task while it is still being written ahead
// waits for it: once handed in it goes to the card, not aside, and it says
// which card asked, as does the task it brings.
func TestARequestAheadIsWaitedForOnceTheChildAsks(t *testing.T) {
	t.Parallel()

	sealer := newSealer(t)
	p := parseFixture(t, "dima")
	brief := briefOn("counting.gaps")
	request, err := p.AskAhead(&brief, profile.TutorRule, "en", "", asked)
	if err != nil {
		t.Fatalf("AskAhead() error = %v, want nil", err)
	}
	if !request.Ahead {
		t.Fatal("a request ahead is not marked as one")
	}

	request.Await("tsk_the_card_asked_after", asked)
	if request.Ahead || request.TakenAfter != "tsk_the_card_asked_after" {
		t.Errorf("the request awaited is %+v, want it waited for, after the card's task", *request)
	}
	task, err := p.Issue(written(), secret(), sealer, keptAt)
	if err != nil {
		t.Fatalf("Issue() error = %v, want nil", err)
	}
	if task.TakenAfter != "tsk_the_card_asked_after" || task.ID != profile.TaskIDFor(request.ID) {
		t.Errorf("the task handed out is %s after %q, want %s after the card's task",
			task.ID, task.TakenAfter, profile.TaskIDFor(request.ID))
	}
}

// A request written ahead that the child comes to wait for has its window
// start again, however long it lay open ahead: the model is sent to write its
// task now, and the card waits for it from now on.
func TestARequestAheadWaitedForStartsItsWindowAgain(t *testing.T) {
	t.Parallel()

	const window = 15 * time.Minute
	p := parseFixture(t, "dima")
	brief := briefOn("counting.gaps")
	request, err := p.AskAhead(&brief, profile.TutorRule, "en", "", asked)
	if err != nil {
		t.Fatalf("AskAhead() error = %v, want nil", err)
	}

	waited := asked.Add(window - time.Minute)
	request.Await("", waited)
	if handedIn := waited.Add(2 * time.Minute); !request.Awaited(window, handedIn) {
		t.Errorf("the request waited for since %v is over at %v, want its window started again", waited, handedIn)
	}
}

// A task a person asked for says so on the card, so that the task written
// ahead after it can be set the same.
func TestAPersonsAskIsCarriedToTheTaskOnTheCard(t *testing.T) {
	t.Parallel()

	p := parseFixture(t, "dima")
	brief := briefOn("combinatorics.enumeration")
	p.Ask(&brief, profile.TutorLLM, "en", asked).Asked = true
	task, err := p.Issue(written(), secret(), newSealer(t), keptAt)
	if err != nil {
		t.Fatalf("Issue() error = %v, want nil", err)
	}
	if !task.Asked {
		t.Error("a task a person asked for does not say so")
	}
}

// The task of a request kept ready stands as kept, and moves on to the card
// once it is handed out.
func TestAKeptTaskStandsAsKept(t *testing.T) {
	t.Parallel()

	p := parseFixture(t, "dima")
	request, _ := keepOne(t, p, newSealer(t), "counting.gaps")

	if state, _ := p.TaskFor(request.ID, requestWindow, keptAt); state != profile.TaskKept {
		t.Errorf("TaskFor() = %s, want %s", state, profile.TaskKept)
	}
	if _, err := p.HandOutReady("", keptAt); err != nil {
		t.Fatalf("HandOutReady() error = %v, want nil", err)
	}
	if state, _ := p.TaskFor(request.ID, requestWindow, keptAt); state != profile.TaskOnTheCard {
		t.Errorf("TaskFor() = %s, want %s", state, profile.TaskOnTheCard)
	}
}

// A task written ahead is handed out only while the lesson is as it was
// written for, and stands where a person asks this time. A skill let back in
// since is no misfit: the task keeps out more than it now has to. The version
// of the instructions is asked of a task kept, not of one still being written.
func TestATaskWrittenAheadFitsOnlyTheLessonItWasWrittenFor(t *testing.T) {
	t.Parallel()

	fits := profile.Lesson{
		Language:            "en",
		ExcludedSkills:      []string{"fractions"},
		InstructionsVersion: "357968db0310",
	}
	for _, tc := range []struct {
		name    string
		change  func(*profile.Lesson)
		kept    profile.Misfit
		writing profile.Misfit
	}{
		{"as it was written for", func(*profile.Lesson) {}, profile.MisfitNone, profile.MisfitNone},
		{"in another language", func(l *profile.Lesson) { l.Language = "ru" }, profile.MisfitLanguage, profile.MisfitLanguage},
		{"kept to a topic since", func(l *profile.Lesson) { l.Topic = "counting.gaps" }, profile.MisfitLessonTopic, profile.MisfitLessonTopic},
		{"keeping out a skill more", func(l *profile.Lesson) { l.ExcludedSkills = []string{"fractions", "division"} },
			profile.MisfitExcludedSkills, profile.MisfitExcludedSkills},
		{"keeping out a skill less", func(l *profile.Lesson) { l.ExcludedSkills = []string{} }, profile.MisfitNone, profile.MisfitNone},
		{"asked for its own topic", func(l *profile.Lesson) { l.Asked = profile.Place{Topic: "counting.gaps"} },
			profile.MisfitNone, profile.MisfitNone},
		{"asked for its own topic and difficulty", func(l *profile.Lesson) {
			l.Asked = profile.Place{Topic: "counting.gaps", Difficulty: 3}
		}, profile.MisfitNone, profile.MisfitNone},
		{"asked for another topic", func(l *profile.Lesson) { l.Asked = profile.Place{Topic: "logic.weighing"} },
			profile.MisfitAsked, profile.MisfitAsked},
		{"asked for another level", func(l *profile.Lesson) { l.Asked = profile.Place{GradeLevel: rating.Grades34} },
			profile.MisfitAsked, profile.MisfitAsked},
		{"asked for another difficulty", func(l *profile.Lesson) { l.Asked = profile.Place{Difficulty: 4} },
			profile.MisfitAsked, profile.MisfitAsked},
		{"written to other instructions", func(l *profile.Lesson) { l.InstructionsVersion = "a1b2c3d4e5f6" },
			profile.MisfitInstructions, profile.MisfitNone},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			p := parseFixture(t, "dima")
			brief := briefOn("counting.gaps")
			brief.ExcludedSkills = []string{"fractions"}
			request, err := p.AskAhead(&brief, profile.TutorRule, "en", "", asked)
			if err != nil {
				t.Fatalf("AskAhead() error = %v, want nil", err)
			}
			now := fits
			tc.change(&now)
			if got := request.Misfit(&now); got != tc.writing {
				t.Errorf("the request's Misfit() = %q, want %q", got, tc.writing)
			}

			kept, err := p.Keep(written(), secret(), newSealer(t), keptAt)
			if err != nil {
				t.Fatalf("Keep() error = %v, want nil", err)
			}
			if got := kept.Misfit(&now); got != tc.kept {
				t.Errorf("the kept task's Misfit() = %q, want %q", got, tc.kept)
			}
		})
	}
}

// A task written while the lessons were kept to a topic fits only while they
// still are: given back to the rule, the lessons have moved away from it.
func TestATaskWrittenOnTheLessonsTopicFitsOnlyWhileTheyAreKeptToIt(t *testing.T) {
	t.Parallel()

	p := parseFixture(t, "dima")
	brief := briefOn("counting.gaps")
	if _, err := p.AskAhead(&brief, profile.TutorPerson, "en", "counting.gaps", asked); err != nil {
		t.Fatalf("AskAhead() error = %v, want nil", err)
	}
	kept, err := p.Keep(written(), secret(), newSealer(t), keptAt)
	if err != nil {
		t.Fatalf("Keep() error = %v, want nil", err)
	}

	lesson := profile.Lesson{Language: "en", Topic: "counting.gaps", InstructionsVersion: kept.InstructionsVersion}
	if got := kept.Misfit(&lesson); got != profile.MisfitNone {
		t.Errorf("on the same topic Misfit() = %q, want none", got)
	}
	lesson.Topic = ""
	if got := kept.Misfit(&lesson); got != profile.MisfitLessonTopic {
		t.Errorf("given back to the rule Misfit() = %q, want %q", got, profile.MisfitLessonTopic)
	}
}

// A task kept is held to the rules of the file as the task on the card is,
// with how its writing went: a file edited to break one is refused, naming it.
func TestAKeptTaskIsHeldToTheRulesOfTheFile(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name   string
		breaks func(*profile.ReadyTask)
		names  string
	}{
		{"no id", func(r *profile.ReadyTask) { r.ID = "" }, "ready_task needs an id"},
		{"no language", func(r *profile.ReadyTask) { r.Language = "" }, "ready_task has no language"},
		{"nothing sealed", func(r *profile.ReadyTask) { r.Sealed = "" }, "ready_task.sealed is empty"},
		{"no moment it was accepted", func(r *profile.ReadyTask) { r.WrittenAt = profile.Time{} }, "written_at"},
		{"no attempt", func(r *profile.ReadyTask) { r.Attempts = 0 }, "ready_task.attempts is 0"},
		{"an attempt past the last", func(r *profile.ReadyTask) { r.Attempts = profile.MaxAttempts + 1 }, "ready_task.attempts is 4"},
		{"a difficulty off the ladder", func(r *profile.ReadyTask) { r.Difficulty = 9 }, "ready_task.difficulty is 9"},
		{"a level nobody teaches", func(r *profile.ReadyTask) { r.GradeLevel = "7-8" }, "ready_task.grade_level"},
		{"nobody who chose it", func(r *profile.ReadyTask) { r.TutorMode = "" }, "ready_task.tutor_mode"},
		{"four options", func(r *profile.ReadyTask) { delete(r.Options, "E") }, "ready_task offers 4 options"},
		{"a blank option", func(r *profile.ReadyTask) { r.Options["B"] = " " }, `ready_task.options shows nothing under "B"`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			p := parseFixture(t, "dima")
			_, kept := keepOne(t, p, newSealer(t), "counting.gaps")
			tc.breaks(kept)

			err := p.Validate()
			if !errors.Is(err, profile.ErrInvalid) || !strings.Contains(err.Error(), tc.names) {
				t.Errorf("Validate() error = %v, want %v naming %q", err, profile.ErrInvalid, tc.names)
			}
		})
	}
}

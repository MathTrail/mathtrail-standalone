package mcpserver_test

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/profile"
	"github.com/MathTrail/mathtrail-standalone/internal/domain/rating"
	"github.com/MathTrail/mathtrail-standalone/internal/store"
)

// The next task is written ahead: once a task is on the card, the model writes
// the one after it while the child works, and the service keeps it, sealed,
// until another is asked for. Every task written ahead here is a relay
// of three swimmers — the race again in other words, so that it is no near
// copy of the race on the card.

// relayQuestion is the relay's wording, which the child reads.
const relayQuestion = "Ivy, Joe and Lea swam in the pool. Joe swam before Lea. Ivy swam after Lea. Who swam first?"

// relaySolution would give the relay away before the child has answered.
const relaySolution = "Joe swam before Lea, and Lea swam before Ivy, so Joe swam first."

// relayOn is the relay handed in for a request.
func relayOn(request *profile.OpenRequest) map[string]any {
	relay := raceOn(request)
	relay["task"] = map[string]any{
		"core_idea":      "Order three swimmers from two comparisons.",
		"question":       relayQuestion,
		"options":        map[string]string{"A": "Ivy", "B": "Lea", "C": "Joe", "D": "Nobody", "E": "All at once"},
		"correct_answer": "C",
		"hint":           "Who swam before Lea?",
		"solution":       relaySolution,
		"distractors": map[string]map[string]string{
			"A": {"trap": raceTraps[0], "text": "Ivy swam after Lea, so she was the last to swim."},
			"B": {"trap": raceTraps[1], "text": "Lea is in the middle: Joe swam before her."},
			"D": {"trap": raceTraps[2], "text": "Somebody always swims first when three take turns."},
			"E": {"trap": raceTraps[3], "text": "The three swam one after another, not together."},
		},
	}
	relay["solver"] = `def solve(options):
    firsts = []
    for order in permutations(["Ivy", "Joe", "Lea"]):
        place = {name: i for i, name in enumerate(order)}
        if place["Joe"] < place["Lea"] and place["Lea"] < place["Ivy"]:
            firsts.append(order[0])
    return match(options, firsts[0])
`
	relay["self_check"] = map[string]any{
		"issues": []any{},
		"option_check": map[string]string{
			"A": "Ivy is last.", "B": "Lea is second.", "C": "Joe is first.", "D": "Someone was first.", "E": "Nobody tied.",
		},
		"final_answer": "C",
	}
	return relay
}

// aheadChoice asks for the task written ahead on the relay's topic, level and
// difficulty, as the model's own choice.
var aheadChoice = map[string]any{
	"language": "en", "topic": "logic.ordering", "grade_level": "1-2", "difficulty": 2,
	"reason": "Another ordering task, a little different.",
}

// prepared asks for the next task to be got ready and is its words, which
// carry no payload: the package is for the model alone.
func prepared(t *testing.T, session *mcp.ClientSession, args map[string]any) string {
	t.Helper()

	result := call(t, session, "prepare_task", args)
	if result.IsError || result.StructuredContent != nil {
		t.Fatalf("prepare_task = %+v, %s, want words alone", result.StructuredContent, textOf(t, result))
	}
	return textOf(t, result)
}

// aheadOpen is the request open for the task written ahead, as the profile
// keeps it.
func aheadOpen(t *testing.T, kept store.Storage) *profile.OpenRequest {
	t.Helper()

	p, _ := loadKept(t, kept)
	if p.OpenRequest == nil || !p.OpenRequest.Ahead {
		t.Fatalf("the profile holds the request %+v, want one for the task written ahead", p.OpenRequest)
	}
	return p.OpenRequest
}

// keepTheRelay has the race on the card and the relay written ahead and kept,
// as a first task and the one after it go, and is the race's request and the
// id the relay will have on the card.
func keepTheRelay(t *testing.T, session *mcp.ClientSession, kept store.Storage) (race *profile.OpenRequest, relay string) {
	t.Helper()

	race = raceHandedOut(t, session, kept)
	prepared(t, session, aheadChoice)
	ahead := aheadOpen(t, kept)
	if text := textOf(t, call(t, session, "submit_task", relayOn(ahead))); !strings.HasPrefix(text, "Accepted at attempt 1 and kept") {
		t.Fatalf("submit_task for the relay said %q, want it accepted and kept", text)
	}
	return race, profile.TaskIDFor(ahead.ID)
}

// The first task of a lesson is written when it is asked for, as ever, and its
// acceptance sends the model on to get the next one ready. That one is
// written ahead and kept, sealed, while the child works on the first: it is
// not handed out, not counted against the day, and nothing more is written
// while it waits.
func TestTheNextTaskIsWrittenAheadOnceATaskIsOnTheCard(t *testing.T) {
	t.Parallel()

	kept := racer(t)
	h, session := lesson(t, kept)

	race := askForTheRace(t, session, kept)
	if text := textOf(t, call(t, session, "submit_task", raceOn(race))); !strings.Contains(text, "Now call prepare_task") {
		t.Errorf("the race accepted says %q, want the model sent on to get the next task ready", text)
	}
	words := prepared(t, session, aheadChoice)
	ahead := aheadOpen(t, kept)
	if !strings.Contains(words, "Request "+ahead.ID+" is open for the next task, written ahead") ||
		!strings.Contains(words, "Say nothing about it, now or once it is kept") || !strings.Contains(words, "\n\nPackage:\n") {
		t.Errorf("prepare_task says %q, want the request written ahead, nothing said of it, and the package", leadOf(words))
	}
	if ahead.TutorMode != profile.TutorLLM || ahead.Brief.TargetConcept != "logic.ordering" || ahead.Brief.Difficulty != 2 {
		t.Errorf("the request ahead is %+v, want it where the model asked", ahead)
	}

	keptAs := payloadOf[handedInPayload](t, call(t, session, "submit_task", relayOn(ahead)))
	if keptAs.Screen != "task" || keptAs.Task == nil || keptAs.Task.ID != profile.TaskIDFor(race.ID) {
		t.Errorf("the relay kept hands a card %+v, want the race the child is working on", keptAs)
	}
	p, _ := loadKept(t, kept)
	if p.ReadyTask == nil || p.ReadyTask.ID != profile.TaskIDFor(ahead.ID) || p.OpenRequest != nil ||
		p.CurrentTask.ID != profile.TaskIDFor(race.ID) || p.Daily.Accepted != 1 || len(p.TaskFingerprints) != 1 {
		t.Fatalf("the profile keeps %+v beside the request %+v, %d tasks counted, want the relay kept and the race on the card",
			p.ReadyTask, p.OpenRequest, p.Daily.Accepted)
	}

	if again := prepared(t, session, map[string]any{"language": "en"}); !strings.HasPrefix(again, "The next task is written already") {
		t.Errorf("prepare_task with a task kept says %q, want nothing to write", again)
	}
	h.settle()
	wantLessonLines(t, h, map[string]int{"task_requested": 2, "task_accepted": 1, "task_kept": 1})
	if line := linesOf(h, "task_kept")[0].ContextMap(); line["topic"] != "logic.ordering" || line["attempts"] != int64(1) {
		t.Errorf("task_kept = %v, want the relay kept at the first attempt", line)
	}
}

// A child who asks for another task in the chat gets the task kept at once:
// next_task hands it out on the card it draws, reads it out to the model, and
// sends the model on to write the one after it. The task left on the card
// without an answer is skipped, and the task kept opens on the card as any
// task does.
func TestTheTaskKeptComesAtOnceWhenTheChildAsksInTheChat(t *testing.T) {
	t.Parallel()

	kept := racer(t)
	h, session := lesson(t, kept)
	race, relay := keepTheRelay(t, session, kept)

	asked := call(t, session, "next_task", map[string]any{"language": "en"})
	card := payloadOf[requestPayload](t, asked)
	if card.Screen != "task" || card.Task == nil || card.Task.ID != relay || card.Task.Question != relayQuestion ||
		card.Language != "en" {
		t.Fatalf("next_task = %+v, want the relay on the card at once", card)
	}
	wantSaid(t, textOf(t, asked), "Task "+relay+", written ahead, is on the child's card now.",
		"Task "+profile.TaskIDFor(race.ID)+", left on the child's card without an answer, is recorded as skipped.",
		"Now call prepare_task", relayQuestion)
	p, _ := loadKept(t, kept)
	if p.CurrentTask.ID != relay || p.ReadyTask != nil || p.OpenRequest != nil || p.Daily.Accepted != 2 {
		t.Errorf("the profile holds %+v, %+v, %+v, %d counted, want the relay handed out and nothing else",
			p.CurrentTask, p.ReadyTask, p.OpenRequest, p.Daily.Accepted)
	}

	answered := payloadOf[answerPayload](t, call(t, session, "submit_answer", map[string]any{"task_id": relay, "answer": "C"}))
	if answered.Result == nil || !answered.Result.Correct {
		t.Errorf("the answer to the relay = %+v, want it recorded as right", answered)
	}
	h.settle()
	accepted := linesOf(h, "task_accepted")
	if len(accepted) != 2 || accepted[1].ContextMap()["ready"] != true {
		t.Errorf("the relay's task_accepted = %v, want it ready", accepted)
	}
	if _, byCard := accepted[len(accepted)-1].ContextMap()["by_card"]; byCard {
		t.Errorf("the relay's task_accepted = %v, want no word of a card taking it", accepted)
	}
}

// A child who asks for another in the chat while the next task is still being
// written ahead waits for that one: next_task draws a card waiting for it,
// the task the card before showed without an answer is skipped, and the model
// finishes the task rather than starting another. Once handed in, it goes to
// that card, written while the child waited.
func TestAskingInTheChatWaitsForTheTaskBeingWrittenAhead(t *testing.T) {
	t.Parallel()

	kept := racer(t)
	h, session := lesson(t, kept)
	raceHandedOut(t, session, kept)
	prepared(t, session, aheadChoice)
	ahead := aheadOpen(t, kept)

	asked := call(t, session, "next_task", map[string]any{"language": "en"})
	coming := wantComing(t, asked)
	if coming.RequestID != ahead.ID || !coming.AlreadyOpen {
		t.Errorf("next_task = %+v, want the card waiting for the request ahead", coming)
	}
	wantSaid(t, textOf(t, asked), "its task was being written ahead, and now the child waits for it")
	p, _ := loadKept(t, kept)
	if p.OpenRequest == nil || p.OpenRequest.Ahead || p.CurrentTask != nil {
		t.Fatalf("the profile holds %+v and %+v, want the request waited for and the race skipped", p.OpenRequest, p.CurrentTask)
	}

	handed := call(t, session, "submit_task", relayOn(p.OpenRequest))
	wantSaid(t, textOf(t, handed), "Accepted at attempt 1: task "+profile.TaskIDFor(ahead.ID)+" goes on the child's card now",
		"Now call prepare_task")
	if shown := awaited(t, session, ahead.ID); shown.Screen != "task" || shown.Task == nil || shown.Task.Question != relayQuestion {
		t.Errorf("read_task once the relay is accepted = %+v, want it on the card next_task drew", shown)
	}
	h.settle()
	wantLessonLines(t, h, map[string]int{"task_skipped": 1, "task_accepted": 2, "task_kept": 0})
	accepted := linesOf(h, "task_accepted")[1].ContextMap()
	if _, byCard := accepted["by_card"]; accepted["ready"] != false || byCard {
		t.Errorf("the relay's task_accepted = %v, want it written while the child waited, with no word of a card", accepted)
	}
}

// A task handed in twice, the answer to the first lost, is kept once: the
// second hand-in is told it was kept already, and spends nothing.
func TestATaskWrittenAheadIsKeptOnce(t *testing.T) {
	t.Parallel()

	kept := racer(t)
	_, session := lesson(t, kept)
	raceHandedOut(t, session, kept)
	prepared(t, session, aheadChoice)
	relay := relayOn(aheadOpen(t, kept))
	call(t, session, "submit_task", relay)
	before, revision := loadKept(t, kept)

	again := call(t, session, "submit_task", relay)
	wantSaid(t, textOf(t, again), "was accepted already and is kept")
	if after, now := loadKept(t, kept); now != revision || after.ReadyTask.ID != before.ReadyTask.ID {
		t.Error("a task handed in again changed the task kept")
	}
}

// A request written ahead that runs out of attempts asks for no new task: the
// child is working on the task on the card and waits for none.
func TestATaskWrittenAheadThatNeverPassesAsksForNothing(t *testing.T) {
	t.Parallel()

	kept := racer(t)
	_, session := lesson(t, kept)
	raceHandedOut(t, session, kept)
	prepared(t, session, aheadChoice)
	ahead := aheadOpen(t, kept)

	var last string
	for range profile.MaxAttempts {
		last = textOf(t, call(t, session, "submit_task", broken(ahead)))
	}
	wantSaid(t, last, "request "+ahead.ID+" is closed and nothing was kept", "ask for no new task")
	if strings.Contains(last, "next_task") {
		t.Errorf("the last refusal says %q, want no new task asked for", last)
	}
}

// Once the day's tasks are over, nothing is written ahead: the task would wait
// for a day the lesson may never come back to.
func TestNothingIsWrittenAheadOnceTheDaysTasksAreOver(t *testing.T) {
	t.Parallel()

	kept := keptAs(t, "olya", func(p *profile.Profile) {
		p.Daily = profile.Daily{Date: profile.DateOf(lessonDay), Accepted: 20}
	})
	_, session := lesson(t, kept)

	if words := prepared(t, session, map[string]any{"language": "en"}); !strings.HasPrefix(words, "The child's tasks for today are over") {
		t.Errorf("prepare_task says %q, want nothing to write", words)
	}
	if p, _ := loadKept(t, kept); p.OpenRequest != nil {
		t.Errorf("the profile holds the request %+v, want none", p.OpenRequest)
	}
}

// A choice of the model's own no task can be set at is refused, field by
// field, and nothing is written.
func TestAChoiceNoTaskAheadCanBeSetAtIsRefused(t *testing.T) {
	t.Parallel()

	kept := racer(t)
	_, session := lesson(t, kept)
	raceHandedOut(t, session, kept)

	refused := payloadOf[requestPayload](t, call(t, session, "prepare_task", map[string]any{
		"language": "en", "topic": "no.such_topic", "reason": "Something else.",
	}))
	if refused.Status != "rejected" || refused.Code != "invalid_arguments" || len(refused.Problems) != 1 ||
		refused.Problems[0].Field != "topic" {
		t.Errorf("prepare_task = %+v, want the topic refused", refused)
	}
	if p, _ := loadKept(t, kept); p.OpenRequest != nil {
		t.Errorf("the profile holds the request %+v, want none", p.OpenRequest)
	}
}

// A task a person asked for is followed, written ahead, by one more like it,
// in case the child wants another: the same topic, level and difficulty. The
// one after that is the rule's again.
func TestATaskAPersonAskedForIsFollowedByOneLikeIt(t *testing.T) {
	t.Parallel()

	kept := racer(t)
	_, session := lesson(t, kept)
	raceHandedOut(t, session, kept)
	if p, _ := loadKept(t, kept); !p.CurrentTask.Asked {
		t.Fatal("the race a person asked for does not say so")
	}

	prepared(t, session, map[string]any{"language": "en"})
	ahead := aheadOpen(t, kept)
	if ahead.Brief.TargetConcept != "logic.ordering" || ahead.Brief.GradeLevel != "1-2" || ahead.Brief.Difficulty != 2 ||
		ahead.Asked {
		t.Errorf("the request ahead is %+v, want one more like the race, asked for by nobody", ahead)
	}
}

// A person's ask that the task written ahead fits is served with it — kept, or
// still being written — and the task counts as asked for: the task written
// ahead after it is one more like it, as after any task a person asked for.
func TestATaskWrittenAheadForAPersonsAskCountsAsAskedFor(t *testing.T) {
	t.Parallel()

	ask := map[string]any{"language": "en", "topic": "logic.ordering", "reason": "The child wants another ordering task."}
	for _, tc := range []struct {
		name  string
		serve func(t *testing.T, session *mcp.ClientSession, kept store.Storage) string
	}{
		{"kept", func(t *testing.T, session *mcp.ClientSession, kept store.Storage) string {
			_, relay := keepTheRelay(t, session, kept)
			call(t, session, "next_task", ask)
			return relay
		}},
		{"still being written", func(t *testing.T, session *mcp.ClientSession, kept store.Storage) string {
			raceHandedOut(t, session, kept)
			prepared(t, session, aheadChoice)
			call(t, session, "next_task", ask)
			p, _ := loadKept(t, kept)
			if p.OpenRequest == nil || !p.OpenRequest.Asked {
				t.Fatalf("the request waited for is %+v, want it asked for", p.OpenRequest)
			}
			call(t, session, "submit_task", relayOn(p.OpenRequest))
			return profile.TaskIDFor(p.OpenRequest.ID)
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			kept := racer(t)
			_, session := lesson(t, kept)
			relay := tc.serve(t, session, kept)
			if p, _ := loadKept(t, kept); p.CurrentTask == nil || p.CurrentTask.ID != relay || !p.CurrentTask.Asked {
				t.Fatalf("the task on the card is %+v, want the relay, asked for", p.CurrentTask)
			}

			prepared(t, session, map[string]any{"language": "en"})
			if ahead := aheadOpen(t, kept); ahead.Brief.TargetConcept != "logic.ordering" ||
				!strings.Contains(ahead.Brief.Rationale, "One more like the task the adult asked for") {
				t.Errorf("the request ahead is %+v, want one more like the relay", ahead.Brief)
			}
		})
	}
}

// One more like a task a person asked for that no task can be set at any more
// — its level no longer taught in its topic, say — is the rule's: the model is
// not told to fix arguments it never passed.
func TestOneMoreLikeATaskNoLongerSetIsTheRules(t *testing.T) {
	t.Parallel()

	kept := racer(t)
	_, session := lesson(t, kept)
	raceHandedOut(t, session, kept)
	changeKept(t, kept, func(p *profile.Profile) {
		p.CurrentTask.Topic, p.CurrentTask.GradeLevel = "logic.knights_liars", rating.Grades12
	})

	words := prepared(t, session, map[string]any{"language": "en"})
	if !strings.Contains(words, "is open for the next task, written ahead") {
		t.Fatalf("prepare_task says %q, want the next task written ahead", leadOf(words))
	}
	if ahead := aheadOpen(t, kept); ahead.TutorMode != profile.TutorRule {
		t.Errorf("the request ahead is %+v, want it the rule's", ahead)
	}
}

// A topic chosen on the card while the next task is being written ahead on
// another tells the model nothing of that task keeping its topic: it is let go
// once the next task is asked for.
func TestATopicChosenWhileATaskIsWrittenAheadKeepsNothingOfIt(t *testing.T) {
	t.Parallel()

	kept := keptWith(t, "olya")
	_, session := lesson(t, kept)
	raceHandedOut(t, session, kept)
	prepared(t, session, aheadChoice)
	aheadOpen(t, kept)

	words := textOf(t, call(t, session, "edit_profile", map[string]any{"lesson_topic": "time.clocks"}))
	if strings.Contains(words, "stays on its own topic") {
		t.Errorf("edit_profile says %q, want nothing of the task written ahead keeping its topic", words)
	}
}

// An answer to no task on the card is not told that a task is on its way
// while the only request open is for a task written ahead: nobody waits for
// that one.
func TestAnAnswerToNoTaskIsNotToldOfATaskWrittenAhead(t *testing.T) {
	t.Parallel()

	kept := racer(t)
	_, session := lesson(t, kept)
	prepared(t, session, map[string]any{"language": "en"})
	aheadOpen(t, kept)

	words := textOf(t, call(t, session, "submit_answer", map[string]any{"task_id": "tsk_never_given", "answer": "C"}))
	wantSaid(t, words, "There is no task on the card: ask for a new one with next_task")
}

// changeKept changes the development account's profile in the store, as a
// person changing the details or a newer release would leave it.
func changeKept(t *testing.T, kept store.Storage, change func(*profile.Profile)) {
	t.Helper()

	p, revision := loadKept(t, kept)
	change(p)
	p.Touch("test", lessonDay)
	if _, err := kept.Save(t.Context(), devAccount, p, revision); err != nil {
		t.Fatalf("Save() error = %v, want the profile changed", err)
	}
}

// A task kept is handed out only while the lesson is as it was written for,
// and a task a person asks for is set where they asked: one the lesson moved
// away from since — another language of the lessons, a skill kept out since,
// another version of the instructions, another place asked for — is let go,
// the line saying why, and the task asked for is written now. One that still
// fits is handed out.
func TestATaskKeptIsLetGoWhenTheLessonMovesAway(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name   string
		change func(*profile.Profile)
		ask    map[string]any
		why    string
	}{
		{"the lessons are held in another language", func(p *profile.Profile) {
			russian := "ru"
			p.Student.UILanguage = &russian
		}, map[string]any{"language": "en"}, "language"},
		{"a skill is kept out of the tasks since", func(p *profile.Profile) {
			p.Student.ExcludedSkills = append(p.Student.ExcludedSkills, "fractions")
		}, map[string]any{"language": "en"}, "excluded_skills"},
		{"it was written to other instructions", func(p *profile.Profile) {
			p.ReadyTask.InstructionsVersion = "a1b2c3d4e5f6"
		}, map[string]any{"language": "en"}, "instructions"},
		{"a person asks for another difficulty", func(*profile.Profile) {}, map[string]any{
			"language": "en", "difficulty": 4, "reason": "The child asked for a harder one.",
		}, "asked"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			kept := racer(t)
			h, session := lesson(t, kept)
			_, relay := keepTheRelay(t, session, kept)
			changeKept(t, kept, tc.change)

			coming := wantComing(t, call(t, session, "next_task", tc.ask))
			p, _ := loadKept(t, kept)
			if p.ReadyTask != nil || p.OpenRequest == nil || coming.RequestID != p.OpenRequest.ID ||
				p.CurrentTask != nil && p.CurrentTask.ID == relay {
				t.Fatalf("the profile keeps %+v with the request %+v, want the relay let go and a task asked for now",
					p.ReadyTask, p.OpenRequest)
			}
			h.settle()
			dropped := linesOf(h, "task_dropped")
			if len(dropped) != 1 || dropped[0].ContextMap()["reason"] != tc.why || dropped[0].ContextMap()["written"] != true {
				t.Errorf("task_dropped lines = %v, want one, the relay let go for %s", dropped, tc.why)
			}
		})
	}
}

// A task still being written ahead is let go as a task kept is once the
// lesson moves away from it — here to lessons held in another language: the
// child is not made to wait for a task in the old one. The task asked for is
// written now, in the lesson's language, and the line says the task let go
// had not been written yet.
func TestATaskBeingWrittenAheadIsLetGoWhenTheLessonMovesAway(t *testing.T) {
	t.Parallel()

	kept := racer(t)
	h, session := lesson(t, kept)
	raceHandedOut(t, session, kept)
	prepared(t, session, aheadChoice)
	ahead := aheadOpen(t, kept)
	changeKept(t, kept, func(p *profile.Profile) {
		russian := "ru"
		p.Student.UILanguage = &russian
	})

	coming := wantComing(t, call(t, session, "next_task", map[string]any{"language": "en"}))
	p, _ := loadKept(t, kept)
	if p.OpenRequest == nil || p.OpenRequest.ID == ahead.ID || p.OpenRequest.Ahead || coming.RequestID != p.OpenRequest.ID ||
		coming.Language != "ru" {
		t.Fatalf("next_task = %+v with the request %+v, want the request ahead let go and a task asked for now, in ru",
			coming, p.OpenRequest)
	}
	h.settle()
	dropped := linesOf(h, "task_dropped")
	if len(dropped) != 1 || dropped[0].ContextMap()["reason"] != "language" || dropped[0].ContextMap()["written"] != false {
		t.Errorf("task_dropped lines = %v, want one, the task being written let go for language", dropped)
	}
}

// A task kept is handed out when a person asks for where it stands: asking
// for its own topic asks for nothing it is not.
func TestATaskKeptIsHandedOutForAnAskItFits(t *testing.T) {
	t.Parallel()

	kept := racer(t)
	_, session := lesson(t, kept)
	_, relay := keepTheRelay(t, session, kept)

	card := payloadOf[requestPayload](t, call(t, session, "next_task", map[string]any{
		"language": "en", "topic": "logic.ordering", "reason": "The child wants another ordering task.",
	}))
	if card.Screen != "task" || card.Task == nil || card.Task.ID != relay {
		t.Errorf("next_task = %+v, want the relay handed out", card)
	}
}

// A topic chosen on the card lets go of a task kept on another: the next task
// asked for, on the card next_task draws, is one on the topic chosen.
func TestATopicChosenOnTheCardLetsGoOfATaskKeptOnAnother(t *testing.T) {
	t.Parallel()

	kept := keptWith(t, "olya")
	h, session := lesson(t, kept)
	keepTheRelay(t, session, kept)
	call(t, session, "edit_profile", map[string]any{"lesson_topic": "time.clocks"})

	coming := wantComing(t, call(t, session, "next_task", map[string]any{"language": "en"}))
	p, _ := loadKept(t, kept)
	if p.OpenRequest == nil || coming.RequestID != p.OpenRequest.ID || p.OpenRequest.Brief.TargetConcept != "time.clocks" ||
		p.OpenRequest.TutorMode != profile.TutorPerson || p.ReadyTask != nil {
		t.Fatalf("next_task = %+v with the request %+v, want a new card waiting for a task on the topic chosen",
			coming, p.OpenRequest)
	}
	h.settle()
	if dropped := linesOf(h, "task_dropped"); len(dropped) != 1 || dropped[0].ContextMap()["reason"] != "lesson_topic" {
		t.Errorf("task_dropped lines = %v, want the relay let go for the topic of the lessons", dropped)
	}
}

// However a lesson goes — a task asked for in the chat, the next one got
// ready, a task handed in for whatever request is open, the card's ask for the
// next task, an answer, time passing — the profile never keeps a task ready
// beside a request open, and the day counts each task handed out once.
func TestATaskIsNeverKeptBesideARequestOpen(t *testing.T) {
	t.Parallel()

	// A lesson runs every call through the endpoint, so forty of them, from eight
	// to twenty-four steps long, are what this property can afford: long enough
	// for a task to be kept and handed out.
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests, parameters.MinSize, parameters.MaxSize = 40, 8, 24
	properties := gopter.NewProperties(parameters)
	properties.Property("one task at a time is kept or being written, and every task handed out counts once", prop.ForAll(
		func(steps []int) bool {
			kept, moving := racer(t), &clock{at: lessonDay}
			h, session := lessonWith(t, kept, moving, nil)
			for _, step := range steps {
				lessonStep(t, session, kept, moving, step)
				p, _ := loadKept(t, kept)
				if p.ReadyTask != nil && p.OpenRequest != nil {
					t.Logf("after step %d the profile keeps %s beside the request %s", step, p.ReadyTask.ID, p.OpenRequest.ID)
					return false
				}
			}
			p, _ := loadKept(t, kept)
			h.settle()
			if handed := len(linesOf(h, "task_accepted")); p.Daily.Accepted != handed {
				t.Logf("the day counts %d tasks, and %d were handed out", p.Daily.Accepted, handed)
				return false
			}
			return true
		},
		gen.SliceOf(genLessonStep()),
	))

	properties.TestingRun(t)
}

// genLessonStep picks the steps of a lesson as often as a lesson takes them:
// a task handed in most often, and time passing least.
func genLessonStep() gopter.Gen {
	return gen.Weighted([]gen.WeightedGen{
		{Weight: 2, Gen: gen.Const(0)},
		{Weight: 2, Gen: gen.Const(1)},
		{Weight: 4, Gen: gen.Const(2)},
		{Weight: 2, Gen: gen.Const(3)},
		{Weight: 1, Gen: gen.Const(4)},
		{Weight: 1, Gen: gen.Const(5)},
	})
}

// lessonStep does one thing a lesson does, as the model and the card do it:
// ask for a task in the chat, get the next one ready, hand a task in for the
// request open — the tasks given forgotten, so that the same one passes again —
// ask for the next task as the card's button does, in words alone, answer the
// task on the card, or let six minutes pass.
func lessonStep(t *testing.T, session *mcp.ClientSession, kept store.Storage, moving *clock, step int) {
	t.Helper()

	p, _ := loadKept(t, kept)
	switch step {
	case 0:
		call(t, session, "next_task", raceChoice)
	case 1:
		call(t, session, "prepare_task", aheadChoice)
	case 2:
		if p.OpenRequest != nil {
			// The same relay every time: the tasks given are forgotten first, so
			// that it is no near copy of the last and the lesson moves on.
			changeKept(t, kept, func(p *profile.Profile) { p.TaskFingerprints = []string{} })
			call(t, session, "submit_task", relayOn(p.OpenRequest))
		}
	case 3:
		call(t, session, "next_task", map[string]any{"language": "en"})
	case 4:
		if p.CurrentTask != nil {
			call(t, session, "submit_answer", map[string]any{"task_id": p.CurrentTask.ID, "answer": "C"})
		}
	case 5:
		moving.advance(6 * time.Minute)
	}
}

// A day with no room for another task hands nothing out, kept or not: the
// model asking in the chat is told the day is over, and the task kept waits for
// the next day, as the task on the card stays where it is.
func TestADayWithNoRoomHandsOutNoTaskKept(t *testing.T) {
	t.Parallel()

	kept := racer(t)
	h, session := lesson(t, kept)
	race, relay := keepTheRelay(t, session, kept)
	changeKept(t, kept, func(p *profile.Profile) {
		p.Daily = profile.Daily{Date: profile.DateOf(lessonDay), Accepted: 20}
	})

	if asked := payloadOf[requestPayload](t, call(t, session, "next_task", map[string]any{"language": "en"})); asked.Status != "limited" {
		t.Errorf("next_task on a full day = %+v, want the day's end", asked)
	}
	p, _ := loadKept(t, kept)
	if p.ReadyTask == nil || p.ReadyTask.ID != relay || p.CurrentTask == nil || p.CurrentTask.ID != profile.TaskIDFor(race.ID) {
		t.Errorf("the profile keeps %+v with %+v on the card, want the relay kept and the race on the card", p.ReadyTask, p.CurrentTask)
	}
	h.settle()
	if hit := linesOf(h, "limit_hit"); len(hit) != 1 {
		t.Errorf("limit_hit lines = %d, want one for the refusal", len(hit))
	}
}

// Asked again while the task ahead is being written, prepare_task hands back
// the same request and its package, writing nothing, and says that a choice
// passed again was not applied.
func TestTheRequestAheadIsHandedBackAsItWasOpened(t *testing.T) {
	t.Parallel()

	kept := racer(t)
	_, session := lesson(t, kept)
	raceHandedOut(t, session, kept)
	prepared(t, session, aheadChoice)
	ahead := aheadOpen(t, kept)
	_, revision := loadKept(t, kept)

	words := prepared(t, session, map[string]any{
		"language": "en", "difficulty": 4, "reason": "A harder one, after all.",
	})
	wantSaid(t, leadOf(words), "Request "+ahead.ID+" is open for the next task, written ahead",
		"The arguments of this call were not applied")
	if p, now := loadKept(t, kept); now != revision || p.OpenRequest.Brief.Difficulty != 2 {
		t.Errorf("the request is %+v, want it as it was opened, and nothing written", p.OpenRequest)
	}
}

// A task written ahead while the child works on the task on the card comes
// after that task, so it is built on the idea after the one that task was
// built on, the two being of one topic, and is shown the reference tasks the
// next task is shown once the child has answered: the package of the same
// request, fetched again after the answer, names the same idea and shows the
// same reference tasks.
func TestATaskWrittenAheadIsBuiltOnTheIdeaAfterTheTaskOnTheCard(t *testing.T) {
	t.Parallel()

	kept := racer(t)
	_, session := lesson(t, kept)
	race := askForTheRace(t, session, kept)
	ofTheRace := packageIn(t, fetchPackage(t, session, race.ID))
	if text := textOf(t, call(t, session, "submit_task", raceOn(race))); !strings.HasPrefix(text, "Accepted") {
		t.Fatalf("submit_task for the race said %q, want it accepted", text)
	}

	ahead := packageIn(t, prepared(t, session, aheadChoice))
	request := aheadOpen(t, kept)
	answered(t, answerIt(t, session, profile.TaskIDFor(race.ID), "C", false))
	afterTheAnswer := packageIn(t, fetchPackage(t, session, request.ID))

	type idea struct {
		Number, Round int
		Text          string
	}
	var raced, next idea
	if err := json.Unmarshal(ofTheRace["idea"], &raced); err != nil {
		t.Fatalf("read the race's idea: %v", err)
	}
	if err := json.Unmarshal(ahead["idea"], &next); err != nil {
		t.Fatalf("read the idea of the task written ahead: %v", err)
	}
	if next.Number != raced.Number+1 || next.Round != raced.Round || next.Text == raced.Text {
		t.Errorf("the task written ahead is built on %+v, want the idea after the race's, %+v", next, raced)
	}
	for _, part := range []string{"idea", "examples"} {
		if string(afterTheAnswer[part]) != string(ahead[part]) {
			t.Errorf("after the answer the package's %s is %s, want the one written ahead from, %s",
				part, afterTheAnswer[part], ahead[part])
		}
	}
}

// A child who asks for another task in the chat while the next one is still
// being written ahead, on a Drive with no room left, has nothing waited for:
// the model is told the Drive is full, the request stays written ahead, and
// the task on the card stays where it is.
func TestAskingForTheTaskAheadOnAFullDriveChangesNothing(t *testing.T) {
	t.Parallel()

	kept := &failingLater{Storage: racer(t)}
	_, session := lesson(t, kept)
	race := raceHandedOut(t, session, kept)
	prepared(t, session, aheadChoice)
	ahead := aheadOpen(t, kept)
	kept.armed.Store(true)

	wantOurSentence(t, call(t, session, "next_task", map[string]any{"language": "en"}), "The adult's Google Drive is full")
	if p, _ := loadKept(t, kept); p.OpenRequest == nil || p.OpenRequest.ID != ahead.ID || !p.OpenRequest.Ahead ||
		p.CurrentTask == nil || p.CurrentTask.ID != profile.TaskIDFor(race.ID) {
		t.Errorf("the profile holds the request %+v and the task %+v, want the request ahead and the race as they were",
			p.OpenRequest, p.CurrentTask)
	}
}

// A request for the task written ahead that cannot be saved — the Drive full —
// is told in the words of that failure, and leaves no request open.
func TestARequestAheadThatCannotBeSavedLeavesNoneOpen(t *testing.T) {
	t.Parallel()

	kept := &failingLater{Storage: racer(t)}
	_, session := lesson(t, kept)
	raceHandedOut(t, session, kept)
	kept.armed.Store(true)

	wantOurSentence(t, call(t, session, "prepare_task", aheadChoice), "The adult's Google Drive is full")
	if p, _ := loadKept(t, kept); p.OpenRequest != nil {
		t.Errorf("the profile holds the request %+v, want none", p.OpenRequest)
	}
}

// A task written ahead that passes its checks on a Drive with no room left is
// not kept: the model is told the Drive is full, and the request stays open
// with no attempt spent, for the same task to be handed in again.
func TestATaskWrittenAheadThatCannotBeSavedIsNotKept(t *testing.T) {
	t.Parallel()

	kept := &failingLater{Storage: racer(t)}
	h, session := lesson(t, kept)
	raceHandedOut(t, session, kept)
	prepared(t, session, aheadChoice)
	ahead := aheadOpen(t, kept)
	kept.armed.Store(true)

	wantOurSentence(t, call(t, session, "submit_task", relayOn(ahead)), "The adult's Google Drive is full")
	if p, _ := loadKept(t, kept); p.ReadyTask != nil || p.OpenRequest == nil || p.OpenRequest.ID != ahead.ID ||
		p.OpenRequest.Attempts != 0 {
		t.Errorf("the profile keeps %+v beside the request %+v, want nothing kept and no attempt spent", p.ReadyTask, p.OpenRequest)
	}
	h.settle()
	if lines := linesOf(h, "task_kept"); len(lines) != 0 {
		t.Errorf("task_kept lines = %d, want none for a task the file does not keep", len(lines))
	}
}

// A chat language no task can be written in is refused before anything is
// asked for, by the rule it broke: none given, or one that names no language.
// Nothing is written.
func TestAChatLanguageNoTaskAheadCanBeWrittenInIsRefused(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name, language, code string
	}{
		{"none given", "", "required"},
		{"no language", "not a language", "not_a_language"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			kept := racer(t)
			_, session := lesson(t, kept)
			raceHandedOut(t, session, kept)
			_, revision := loadKept(t, kept)

			refused := payloadOf[requestPayload](t, call(t, session, "prepare_task", map[string]any{"language": tc.language}))
			if refused.Status != "rejected" || refused.Code != "invalid_arguments" || len(refused.Problems) != 1 ||
				refused.Problems[0].Field != "language" || refused.Problems[0].Code != tc.code {
				t.Errorf("prepare_task = %+v, want the language refused as %s", refused, tc.code)
			}
			if p, now := loadKept(t, kept); now != revision || p.OpenRequest != nil {
				t.Errorf("the profile holds the request %+v at revision %s, want nothing written", p.OpenRequest, now)
			}
		})
	}
}

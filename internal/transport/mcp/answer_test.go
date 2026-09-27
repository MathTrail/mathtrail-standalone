package mcpserver_test

import (
	"bytes"
	"encoding/json"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/profile"
	"github.com/MathTrail/mathtrail-standalone/internal/domain/rating"
	"github.com/MathTrail/mathtrail-standalone/internal/domain/solver"
	"github.com/MathTrail/mathtrail-standalone/internal/store"
	"github.com/MathTrail/mathtrail-standalone/internal/store/memory"
)

// The answer: a letter from the card or from the model, or "I don't know",
// recorded once against the task on the card and told back — to the card, which
// turns to its result, and to the model, which explains it. Most cases start
// from the race of the task tests already on the card, handed out the way an
// accepted task is; one goes the whole way from the rule.

// resultPayload is how an answer went, as submit_answer tells it.
type resultPayload struct {
	TaskID        string `json:"task_id"`
	Topic         string `json:"topic"`
	Choice        string `json:"choice"`
	Correct       bool   `json:"correct"`
	CorrectAnswer string `json:"correct_answer"`
	Trap          *struct {
		ID   string `json:"id"`
		Text string `json:"text"`
	} `json:"trap"`
	Solution string `json:"solution"`
	HintUsed bool   `json:"hint_used"`
	Rating   *struct {
		Before int `json:"before"`
		After  int `json:"after"`
	} `json:"rating"`
	Trial           *trialPayload `json:"trial"`
	AlreadyAnswered bool          `json:"already_answered"`
}

// answerPayload is what submit_answer hands back.
type answerPayload struct {
	Screen   string `json:"screen"`
	Status   string `json:"status"`
	Code     string `json:"code"`
	Problems []struct {
		Field string `json:"field"`
		Rule  string `json:"rule"`
	} `json:"problems"`
	LastAnswer *struct {
		TaskID  string `json:"task_id"`
		Correct bool   `json:"correct"`
	} `json:"last_answer"`
	Result *resultPayload `json:"result"`
}

// raceInstructions is the version of the instructions the race on the card was
// written to.
const raceInstructions = "0a1b2c3d4e5f"

// raceOnTheCard is a child of grade 2 with the race on the card, handed out as
// an accepted task is: its answer, the traps behind its wrong options with
// what the child is told about each, and its solution sealed with the seal of
// these cases. answered is how many answers the child's level rests on: from
// the length of the trial series on, the series is behind them.
func raceOnTheCard(t testing.TB, answered int) *profile.Profile {
	t.Helper()

	p := fuzzProfile(t)
	p.Ratings.Answers = answered
	options, _ := raceTask(nil)["options"].(map[string]string)
	distractors := map[string]profile.Distractor{}
	for letter, wrong := range raceDistractors() {
		distractors[letter] = profile.Distractor{Trap: wrong["trap"], Text: wrong["text"]}
	}
	if _, err := p.Issue(&profile.Written{
		Wording: raceQuestion, Options: options, Hint: "Who finished before Kim?",
		Fingerprint: "the-race", InstructionsVersion: raceInstructions,
	}, profile.TaskSecret{Answer: "C", Distractors: distractors, Solution: raceSolution, Solver: raceSolver},
		sealer(t), lessonDay); err != nil {
		t.Fatalf("Issue() error = %v, want the race on the card", err)
	}
	return p
}

// answerIt answers the task on the card with this id, as the card or the model
// sends it.
func answerIt(t *testing.T, session *mcp.ClientSession, id, answer string, hint bool) *mcp.CallToolResult {
	t.Helper()

	arguments := map[string]any{"task_id": id, "answer": answer}
	if hint {
		arguments["hint_used"] = true
	}
	return call(t, session, "submit_answer", arguments)
}

// answered is the payload of an answer that was told how it went.
func answered(t *testing.T, result *mcp.CallToolResult) *resultPayload {
	t.Helper()

	payload := payloadOf[answerPayload](t, result)
	if payload.Screen != "result" || payload.Status != "" || payload.Result == nil {
		t.Fatalf("submit_answer = %+v, want the result of an answer", payload)
	}
	return payload.Result
}

// The whole way, as a model and a card go it: a task asked for and handed in,
// on the child's card, and the child's answer to it recorded and told back —
// every step a line. The child is new, so the answer is the first of the trial
// series and there is no rating to show yet.
func TestAnAnswerGoesFromTheCardToItsResult(t *testing.T) {
	t.Parallel()

	kept := racer(t)
	h, session := lesson(t, kept)
	card := wantOnTheCard(t, call(t, session, "submit_task", raceOn(askForTheRace(t, session, kept))))

	answer := answerIt(t, session, card.Task.ID, "C", false)
	result := answered(t, answer)
	if result.TaskID != card.Task.ID || result.Choice != "C" || !result.Correct || result.CorrectAnswer != "C" ||
		result.Solution != raceSolution || result.Trap != nil {
		t.Errorf("result = %+v, want C, right, with the solution and no trap", result)
	}
	if result.Rating != nil || result.Trial == nil || *result.Trial != (trialPayload{Answered: 1, Of: rating.TrialAnswers}) {
		t.Errorf("rating %+v and trial %+v, want the first of the trial series and no rating yet", result.Rating, result.Trial)
	}
	if text := textOf(t, answer); !strings.Contains(text, "and it is right") || !strings.Contains(text, raceSolution) ||
		!strings.Contains(text, "1 of 5 tasks done") {
		t.Errorf("the words are %q, want the answer right, the solution and the trial series", text)
	}

	p, _ := loadKept(t, kept)
	if last, _ := p.LastAnswer(); last.TaskID != card.Task.ID || !last.Correct || p.CurrentTask.Answered == nil {
		t.Errorf("the window ends with %+v and the card holds %+v, want the answer recorded and kept", last, p.CurrentTask)
	}

	h.settle()
	wantLessonLines(t, h, map[string]int{
		"task_requested": 1, "solver_run": 2, "task_submitted": 1, "task_accepted": 1, "answer_recorded": 1,
	})
}

// Right, wrong and "I don't know", each told with what the result side of the
// card shows: the child's choice and the right option, the trap behind a wrong
// letter with what the child is told about it, the solution, and the rating in
// the topic before and after. The child is past the trial series.
func TestAnAnswerIsToldHowItWent(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name, answer string
		correct      bool
		trap         string
		says         string
	}{
		{name: "right", answer: "C", correct: true, says: "and it is right"},
		{name: "wrong", answer: "A", trap: "reversed_relation", says: raceExplained[0]},
		{name: "I don't know", answer: "?", says: "The child did not know, which counts as a wrong answer"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			p := raceOnTheCard(t, rating.TrialAnswers)
			_, session := lesson(t, keptAsIs(t, p))
			answer := answerIt(t, session, p.CurrentTask.ID, tc.answer, false)

			wantToldAsItWent(t, answered(t, answer), tc.answer, tc.correct, tc.trap)
			if text := textOf(t, answer); !strings.Contains(text, tc.says) || !strings.Contains(text, "The rating in Ordering went from") {
				t.Errorf("the words are %q, want them to say %q and how the rating moved", text, tc.says)
			}
		})
	}
}

// wantToldAsItWent holds the result of an answer to how it went: the child's
// choice, right or wrong, C as the right option, the trap behind a wrong
// letter, the solution, and the rating in the topic moved up by a right answer
// and down by any other.
func wantToldAsItWent(t *testing.T, result *resultPayload, choice string, correct bool, trap string) {
	t.Helper()

	if result.Choice != choice || result.Correct != correct || result.CorrectAnswer != "C" ||
		result.Solution != raceSolution || result.AlreadyAnswered {
		t.Errorf("result = %+v, want %s told as it went, C as the right option and the solution", result, choice)
	}
	if got := trapOf(result); got != trap {
		t.Errorf("the trap is %q, want %q", got, trap)
	}
	if result.Trial != nil || result.Rating == nil || (result.Rating.After > result.Rating.Before) != correct {
		t.Errorf("rating %+v and trial %+v, want the rating in the topic moved by the answer", result.Rating, result.Trial)
	}
}

// trapOf is the trap a result names, or nothing.
func trapOf(result *resultPayload) string {
	if result.Trap == nil {
		return ""
	}
	if result.Trap.Text == "" {
		return "a trap with nothing to tell the child"
	}
	return result.Trap.ID
}

// "I don't know" is a wrong answer that took no trap: it moves the child
// exactly as a wrong letter does, and the map of their mistakes and the line
// about it name no trap.
func TestIDontKnowIsRecordedAsAWrongAnswerThatTookNoTrap(t *testing.T) {
	t.Parallel()

	unsure, mistaken := keptAsIs(t, raceOnTheCard(t, rating.TrialAnswers)), keptAsIs(t, raceOnTheCard(t, rating.TrialAnswers))
	h, unsureSession := lesson(t, unsure)
	_, mistakenSession := lesson(t, mistaken)
	answerIt(t, unsureSession, loadedOf(t, unsure).CurrentTask.ID, "?", true)
	answerIt(t, mistakenSession, loadedOf(t, mistaken).CurrentTask.ID, "A", true)

	doubtful, wrong := loadedOf(t, unsure), loadedOf(t, mistaken)
	if doubtful.Ratings != wrong.Ratings || doubtful.Topics["logic.ordering"].Delta != wrong.Topics["logic.ordering"].Delta {
		t.Errorf("I don't know moved the child to %+v, want %+v as a wrong letter does", doubtful.Ratings, wrong.Ratings)
	}
	if traps := doubtful.Topics["logic.ordering"].Traps; len(traps) != 0 {
		t.Errorf("the topic counts the traps %v after I don't know, want none", traps)
	}
	last := doubtful.Recent[len(doubtful.Recent)-1]
	if !last.Confused || last.Correct || last.Chosen != "" || last.Trap != "" || !last.HintUsed {
		t.Errorf("the window ends with %+v, want a wrong answer marked confused, with the hint and no option or trap", last)
	}

	h.settle()
	lines := linesOf(h, "answer_recorded")
	if len(lines) != 1 {
		t.Fatalf("answer_recorded lines = %d, want 1", len(lines))
	}
	if fields := lines[0].ContextMap(); fields["confused"] != true || fields["correct"] != false || fields["trap"] != "" ||
		fields["hint_used"] != true {
		t.Errorf("answer_recorded = %v, want a wrong answer marked confused, with the hint and no trap", fields)
	}
}

// loadedOf is the development account's profile as a store holds it now.
func loadedOf(t *testing.T, kept store.Storage) *profile.Profile {
	t.Helper()

	p, _ := loadKept(t, kept)
	return p
}

// How long the child took is measured by the service, from the moment it
// handed the task out, and written beside the answer with the hint — never in
// the result a card is drawn from.
func TestThePaceIsMeasuredFromTheMomentTheTaskWasHandedOut(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		took time.Duration
		want profile.Pace
	}{
		{10 * time.Second, profile.PaceFast},
		{100 * time.Second, profile.PaceNormal},
		{200 * time.Second, profile.PaceSlow},
	} {
		t.Run(string(tc.want), func(t *testing.T) {
			t.Parallel()

			p := raceOnTheCard(t, rating.TrialAnswers)
			kept, moving := keptAsIs(t, p), &clock{at: lessonDay}
			h, session := lessonWith(t, kept, moving, nil)
			moving.advance(tc.took)

			if result := answered(t, answerIt(t, session, p.CurrentTask.ID, "C", true)); !result.HintUsed {
				t.Errorf("result = %+v, want the hint recorded as used", result)
			}
			window := loadedOf(t, kept).Recent
			if last := window[len(window)-1]; last.Pace != tc.want || !last.HintUsed {
				t.Errorf("the window ends with %+v, want pace %s and the hint", last, tc.want)
			}
			h.settle()
			if line := linesOf(h, "answer_recorded"); len(line) != 1 || line[0].ContextMap()["pace"] != string(tc.want) {
				t.Errorf("answer_recorded lines = %v, want one with pace %s", line, tc.want)
			}
		})
	}
}

// The letter is read as a child or a model types it: in either case, with
// spaces around it.
func TestTheLetterIsReadAsItWasTyped(t *testing.T) {
	t.Parallel()

	p := raceOnTheCard(t, rating.TrialAnswers)
	_, session := lesson(t, keptAsIs(t, p))
	if result := answered(t, answerIt(t, session, p.CurrentTask.ID, " c ", false)); result.Choice != "C" || !result.Correct {
		t.Errorf("result = %+v, want the answer read as C, the right one", result)
	}
}

// The card records the answer and the model records it again — the child
// pressed a button and also typed the letter. The first is recorded, and the
// second, whatever it says, is told the same result as an answer recorded
// before: nothing is written, the rating moves once, and one line is left.
func TestAnAnswerFromTheCardAndTheModelIsRecordedOnce(t *testing.T) {
	t.Parallel()

	p := raceOnTheCard(t, rating.TrialAnswers)
	kept := keptAsIs(t, p)
	h, session := lesson(t, kept)

	first := answered(t, answerIt(t, session, p.CurrentTask.ID, "C", true))
	after, revision := loadKept(t, kept)
	for _, again := range []string{"A", " ? ", "C"} {
		told := answered(t, answerIt(t, session, p.CurrentTask.ID, again, false))
		if !told.AlreadyAnswered {
			t.Errorf("sent again with %q, it was recorded anew: %+v", again, told)
		}
		told.AlreadyAnswered = false
		if !sameResult(told, first) {
			t.Errorf("sent again with %q, it was told %+v, want what was recorded: %+v", again, told, first)
		}
	}
	if now, at := loadKept(t, kept); at != revision || now.Ratings != after.Ratings {
		t.Errorf("an answer sent again wrote the profile over revision %v", revision)
	}

	h.settle()
	if lines := linesOf(h, "answer_recorded"); len(lines) != 1 {
		t.Errorf("answer_recorded lines = %d, want the one answer recorded", len(lines))
	}
}

// sameResult reports whether two results tell the same thing.
func sameResult(a, b *resultPayload) bool {
	encodedA, errA := json.Marshal(a)
	encodedB, errB := json.Marshal(b)
	return errA == nil && errB == nil && bytes.Equal(encodedA, encodedB)
}

// An answer to a task that is not the one on the card — never given, skipped,
// left behind by the next, or answered and gone — records nothing, and is told
// what is on the card now. The card that sent it stays on its task.
func TestAnAnswerToATaskNotOnTheCardRecordsNothing(t *testing.T) {
	t.Parallel()

	const neverGiven = "tsk_never_given"
	for _, tc := range []struct {
		name   string
		before func(t *testing.T, session *mcp.ClientSession, kept store.Storage) string
		says   string
	}{
		{"a task never given", func(*testing.T, *mcp.ClientSession, store.Storage) string {
			return neverGiven
		}, "is on the card: record the child's answer to it"},
		{"a task never given, beside one answered", func(t *testing.T, session *mcp.ClientSession, kept store.Storage) string {
			answerIt(t, session, loadedOf(t, kept).CurrentTask.ID, "C", false)
			return neverGiven
		}, "has been answered already"},
		{"a task skipped for another", func(t *testing.T, session *mcp.ClientSession, kept store.Storage) string {
			left := loadedOf(t, kept).CurrentTask.ID
			call(t, session, "next_task", map[string]any{"language": "en"})
			return left
		}, "is open, and its task is still to be handed in with submit_task"},
		{"a task never given, with nothing on the card", func(t *testing.T, _ *mcp.ClientSession, kept store.Storage) string {
			takeOffTheCard(t, kept)
			return neverGiven
		}, "There is no task on the card: ask for a new one with next_task"},
		{"a task answered and taken off the card", func(t *testing.T, session *mcp.ClientSession, kept store.Storage) string {
			left := loadedOf(t, kept).CurrentTask.ID
			answerIt(t, session, left, "B", false)
			call(t, session, "next_task", map[string]any{"language": "en"})
			return left
		}, "was answered already, and that answer, wrong, is recorded"},
		{"a task left behind by the next one", func(t *testing.T, session *mcp.ClientSession, kept store.Storage) string {
			left := loadedOf(t, kept).CurrentTask.ID
			answerIt(t, session, left, "C", false)
			handOutAnother(t, kept)
			return left
		}, "is on the card: record the child's answer to it"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			kept := keptAsIs(t, raceOnTheCard(t, rating.TrialAnswers))
			_, session := lesson(t, kept)
			stale := tc.before(t, session, kept)
			_, revision := loadKept(t, kept)

			result := answerIt(t, session, stale, "C", false)
			payload := payloadOf[answerPayload](t, result)
			if payload.Status != "stale" || payload.Code != "stale_task" || payload.Screen != "task" || payload.Result != nil {
				t.Errorf("submit_answer = %+v, want stale_task on the card's own task, with no result", payload)
			}
			if text := textOf(t, result); !strings.Contains(text, tc.says) || strings.Contains(text, neverGiven) {
				t.Errorf("the words are %q, want them to say %q and not to repeat an id that was sent", text, tc.says)
			}
			if _, at := loadKept(t, kept); at != revision {
				t.Error("an answer to a task not on the card wrote the profile")
			}
		})
	}
}

// handOutAnother puts another race on the card, over the task there: the way
// a file edited by hand can hold an open request beside an answered task.
func handOutAnother(t *testing.T, kept store.Storage) {
	t.Helper()

	p, revision := loadKept(t, kept)
	if p.OpenRequest == nil {
		asked := fuzzProfile(t).OpenRequest.Brief
		p.Ask(&asked, profile.TutorLLM, "en", lessonDay)
	}
	if _, err := p.Issue(&profile.Written{
		Wording: "Who finished last?", Options: map[string]string{"A": "Ann", "B": "Kim", "C": "Ben", "D": "Nobody", "E": "All"},
		Hint: "Who finished after Kim?", Fingerprint: "another-race", InstructionsVersion: raceInstructions,
	}, profile.TaskSecret{Answer: "A", Solution: "Ann finished after Kim."}, sealer(t), lessonDay); err != nil {
		t.Fatalf("Issue() error = %v, want another task on the card", err)
	}
	p.Touch("test", lessonDay)
	if _, err := kept.Save(t.Context(), devAccount, p, revision); err != nil {
		t.Fatalf("Save() error = %v", err)
	}
}

// An answer that is no letter of the options and no "I don't know" is refused
// by the rule it breaks, never by what arrived, and nothing is recorded.
func TestAnAnswerThatIsNoLetterIsRefusedByItsRule(t *testing.T) {
	t.Parallel()

	for _, typed := range []string{"F", "zebra", "", "AB", "C)"} {
		t.Run(typed, func(t *testing.T) {
			t.Parallel()

			p := raceOnTheCard(t, rating.TrialAnswers)
			kept := keptAsIs(t, p)
			_, session := lesson(t, kept)
			_, revision := loadKept(t, kept)

			result := answerIt(t, session, p.CurrentTask.ID, typed, false)
			payload := payloadOf[answerPayload](t, result)
			if payload.Status != "rejected" || payload.Code != "invalid_arguments" || len(payload.Problems) != 1 ||
				payload.Problems[0].Field != "answer" || payload.Result != nil {
				t.Errorf("submit_answer = %+v, want invalid_arguments naming the answer", payload)
			}
			if typed == "zebra" && strings.Contains(textOf(t, result), typed) {
				t.Errorf("the words are %q, want the rule and not what was sent", textOf(t, result))
			}
			if _, at := loadKept(t, kept); at != revision {
				t.Error("a refused answer wrote the profile")
			}
		})
	}
}

// After a wrong answer the next task goes over the same topic again, with the
// trap the child just fell for among the traps to use, first.
func TestAWrongAnswerIsGoneOverAgainWithItsTrapFirst(t *testing.T) {
	t.Parallel()

	p := raceOnTheCard(t, rating.TrialAnswers)
	_, session := lesson(t, keptAsIs(t, p))
	answerIt(t, session, p.CurrentTask.ID, "A", false)

	var brief profile.Brief
	asked := call(t, session, "next_task", map[string]any{"language": "en"})
	if err := json.Unmarshal(packageIn(t, textOf(t, asked))["brief"], &brief); err != nil {
		t.Fatalf("the package carries no brief: %v", err)
	}
	if brief.PedagogicalGoal != profile.GoalReinforce || brief.TargetConcept != "logic.ordering" ||
		len(brief.TrapsToUse) == 0 || brief.TrapsToUse[0] != "reversed_relation" {
		t.Errorf("the brief is %+v, want the race's topic gone over again with reversed_relation first", brief)
	}
}

// While the trial series runs an answer shows how far it has got instead of a
// rating, and the answer that ends it shows it ended; the next answer shows
// the rating in the topic, before and after.
func TestTheTrialSeriesEndsOnItsFifthAnswer(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name     string
		answered int
		trial    *trialPayload
		says     string
	}{
		{"the fifth answer", rating.TrialAnswers - 1, &trialPayload{Answered: 5, Of: 5}, "That was the last of the 5 tasks"},
		{"the sixth answer", rating.TrialAnswers, nil, "The rating in Ordering went from"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			p := raceOnTheCard(t, tc.answered)
			_, session := lesson(t, keptAsIs(t, p))
			answer := answerIt(t, session, p.CurrentTask.ID, "C", false)

			result := answered(t, answer)
			if (tc.trial == nil) != (result.Trial == nil) || (tc.trial != nil && *result.Trial != *tc.trial) ||
				(tc.trial == nil) != (result.Rating != nil) {
				t.Errorf("trial %+v and rating %+v, want the trial %+v and a rating only once it is over",
					result.Trial, result.Rating, tc.trial)
			}
			if text := textOf(t, answer); !strings.Contains(text, tc.says) {
				t.Errorf("the words are %q, want them to say %q", text, tc.says)
			}
		})
	}
}

// A task whose seal will not open — the key that sealed it has been retired —
// cannot be checked. The answer is refused as a failure of ours, in a sentence
// of its own, and the task leaves the card with nothing recorded: the next ask
// for a task does not count it as one the child leafed past.
func TestATaskThatCannotBeCheckedLeavesTheCard(t *testing.T) {
	t.Parallel()

	kept := keptWith(t, "masha")
	before := loadedOf(t, kept)
	h, session := lesson(t, kept)

	wantOurSentence(t, answerIt(t, session, before.CurrentTask.ID, "C", false), "This task can no longer be checked")
	after := loadedOf(t, kept)
	if after.CurrentTask != nil || len(after.Recent) != len(before.Recent) || after.Ratings != before.Ratings {
		t.Errorf("the card holds %+v and the window %d entries, want the task gone and nothing recorded",
			after.CurrentTask, len(after.Recent))
	}
	if text := textOf(t, call(t, session, "next_task", map[string]any{"language": "en"})); strings.Contains(text, "recorded as skipped") {
		t.Errorf("next_task says %q, want no skip of a task the service lost", text)
	}

	h.settle()
	wantFailed(t, h, "submit_answer", "sealed")
}

// An answer whose write lost to another made in between is told so, and is
// not counted: the line about it is written once the file holds it.
func TestAnAnswerWhoseWriteLostIsNotCounted(t *testing.T) {
	t.Parallel()

	p := raceOnTheCard(t, rating.TrialAnswers)
	h, session := lesson(t, conflicted{keptAsIs(t, p)})
	wantOurSentence(t, answerIt(t, session, p.CurrentTask.ID, "C", false),
		"The child's profile was changed somewhere else at the same moment")

	h.settle()
	wantFailed(t, h, "submit_answer", "conflict")
	if lines := linesOf(h, "answer_recorded"); len(lines) != 0 {
		t.Errorf("answer_recorded lines = %d for an answer that was not kept, want none", len(lines))
	}
}

// With no profile there is no task to answer: the first sign-in, and nothing
// recorded.
func TestWithNoProfileThereIsNothingToAnswer(t *testing.T) {
	t.Parallel()

	_, session := lesson(t, memory.New())
	payload := payloadOf[answerPayload](t, answerIt(t, session, "tsk_nothing", "C", false))
	if payload.Screen != "first_run" || payload.Status != "stale" || payload.Code != "stale_task" || payload.Result != nil {
		t.Errorf("submit_answer = %+v, want the first sign-in and stale_task", payload)
	}
}

// The line about an answer carries the version of the instructions its task
// was written to — a task handed out before the service took newer ones is
// counted against the ones that produced it — and names its topic. Both are
// read from a file a person can edit, so the line keeps a version only while
// it has the shape of one, and a topic only while the catalog has it.
func TestTheLineOfAnAnswerNamesOnlyWhatTheServiceWrites(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name           string
		edit           func(task *profile.CurrentTask)
		version, topic string
	}{
		{"as the service wrote it", func(*profile.CurrentTask) {}, raceInstructions, "logic.ordering"},
		{"a version typed over", func(task *profile.CurrentTask) { task.InstructionsVersion = pseudonym }, other, "logic.ordering"},
		{"a topic typed over", func(task *profile.CurrentTask) { task.Topic = pseudonym }, raceInstructions, other},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			p := raceOnTheCard(t, rating.TrialAnswers)
			tc.edit(p.CurrentTask)
			h, session := lesson(t, keptAsIs(t, p))
			answered(t, answerIt(t, session, p.CurrentTask.ID, "C", false))

			h.settle()
			lines := linesOf(h, "answer_recorded")
			if len(lines) != 1 {
				t.Fatalf("answer_recorded lines = %d, want 1", len(lines))
			}
			if fields := lines[0].ContextMap(); fields["instructions_version"] != tc.version || fields["topic"] != tc.topic {
				t.Errorf("answer_recorded = %v, want version %s and topic %s", fields, tc.version, tc.topic)
			}
		})
	}
}

// A task answered and then left for the next is no skipped task: asking for
// the next one takes it off the card and records nothing about it.
func TestAnAnsweredTaskIsNotSkippedByTheNextAsk(t *testing.T) {
	t.Parallel()

	p := raceOnTheCard(t, rating.TrialAnswers)
	kept := keptAsIs(t, p)
	h, session := lesson(t, kept)
	answerIt(t, session, p.CurrentTask.ID, "B", false)
	window := len(loadedOf(t, kept).Recent)

	asked := call(t, session, "next_task", map[string]any{"language": "en"})
	if text := textOf(t, asked); strings.Contains(text, "recorded as skipped") {
		t.Errorf("next_task says %q, want no skip of a task that was answered", text)
	}
	if after := loadedOf(t, kept); after.CurrentTask != nil || len(after.Recent) != window {
		t.Errorf("the card holds %+v and the window %d entries, want the task gone and nothing added",
			after.CurrentTask, len(after.Recent))
	}
	h.settle()
	if lines := linesOf(h, "task_skipped"); len(lines) != 0 {
		t.Errorf("task_skipped lines = %d, want none", len(lines))
	}
}

// The answer is told to the card and the model, and to nothing else: not a
// letter, not the pace, not the rating reaches a span, a line or the label of
// a measurement, and neither does a word of the task. The span of the answer
// says whether it was right, and the trap by its catalog id, and nothing more.
func TestNothingOfTheAnswerReachesASpanALineOrALabel(t *testing.T) {
	t.Parallel()

	for _, choice := range []string{"C", "A", "?"} {
		t.Run(choice, func(t *testing.T) {
			t.Parallel()

			p := raceOnTheCard(t, rating.TrialAnswers)
			h, session := lesson(t, keptAsIs(t, p))
			answered(t, answerIt(t, session, p.CurrentTask.ID, choice, true))
			h.settle()

			wantTheAnswerSpan(t, h)
			wantNoLetters(t, h)
			// The words of the task, and the pace, which the line about the
			// answer keeps and nothing traced may.
			words := slices.Concat([]string{raceSolution, raceQuestion, "Ben", "Kim", "Ann"}, raceExplained)
			for _, span := range h.spans.Ended() {
				wantNoneOf(t, "span "+span.Name(), spanTexts(span), slices.Concat(words, []string{"fast", "normal", "slow"}))
			}
			lines := h.logs.All()
			for i := range lines {
				wantNoneOf(t, "line "+lines[i].Message, lineTexts(t, &lines[i]), words)
			}
		})
	}
}

// wantTheAnswerSpan holds the span of an answer to its place and its words:
// under the call that recorded it, saying whether it was right and the trap by
// its catalog id, and nothing else.
func wantTheAnswerSpan(t *testing.T, h *harness) {
	t.Helper()

	recorded := h.spanNamed(t, "record_answer")
	if recorded.Parent().SpanID() != h.spanNamed(t, "tools/call submit_answer").SpanContext().SpanID() {
		t.Error("the answer is not recorded under the call")
	}
	for _, attr := range recorded.Attributes() {
		if key := string(attr.Key); key != "mathtrail.answer.correct" && key != "mathtrail.answer.trap" {
			t.Errorf("the span of the answer carries %s, want whether it was right and the trap alone", key)
		}
	}
}

// wantNoLetters holds every span, every line and every label of a measurement
// to carrying no letter of an option and no "I don't know" as a value, and to
// naming neither a pace nor a rating: the answer, the child's or the right
// one, is told to nobody but the card and the model.
func wantNoLetters(t *testing.T, h *harness) {
	t.Helper()

	for _, label := range everyLabel(t, h) {
		if value := strings.TrimSpace(label.Value.String()); slices.Contains(solver.Letters(), value) || value == profile.DontKnow {
			t.Errorf("%s carries %q, a letter of the answer", label.Key, value)
		}
		if key := string(label.Key); key != "pace" && (strings.Contains(key, "pace") || strings.Contains(key, "rating")) {
			t.Errorf("%s names a pace or a rating", key)
		}
	}
}

// everyLabel is every value a case recorded beside its calls, with its name:
// the attributes of every span, the text fields of every line, and the labels
// of every measurement.
func everyLabel(t *testing.T, h *harness) []attribute.KeyValue {
	t.Helper()

	labels := labelsOf(t, h)
	for _, span := range h.spans.Ended() {
		labels = append(labels, span.Attributes()...)
	}
	lines := h.logs.All()
	for i := range lines {
		for key, value := range lines[i].ContextMap() {
			if text, isText := value.(string); isText {
				labels = append(labels, attribute.String(key, text))
			}
		}
	}
	return labels
}

// labelsOf are the labels every measurement was recorded under.
func labelsOf(t *testing.T, h *harness) []attribute.KeyValue {
	t.Helper()

	var collected metricdata.ResourceMetrics
	if err := h.metrics.Collect(t.Context(), &collected); err != nil {
		t.Fatalf("Collect() error = %v, want nil", err)
	}
	var labels []attribute.KeyValue
	for _, scope := range collected.ScopeMetrics {
		for _, recorded := range scope.Metrics {
			switch data := recorded.Data.(type) {
			case metricdata.Sum[int64]:
				labels = append(labels, labelsOfPoints(data.DataPoints, pointLabels)...)
			case metricdata.Histogram[float64]:
				labels = append(labels, labelsOfPoints(data.DataPoints, histogramLabels)...)
			case metricdata.Histogram[int64]:
				labels = append(labels, labelsOfPoints(data.DataPoints, histogramLabels)...)
			default:
				t.Fatalf("%s is a %T, want a sum or a histogram", recorded.Name, recorded.Data)
			}
		}
	}
	if len(labels) == 0 {
		t.Fatal("nothing was measured under a label, and a guard over no labels guards nothing")
	}
	return labels
}

// labelsOfPoints are the labels of every point of one measurement.
func labelsOfPoints[P any](points []P, labelsOf func(P) attribute.Set) []attribute.KeyValue {
	var labels []attribute.KeyValue
	for _, point := range points {
		set := labelsOf(point)
		labels = append(labels, set.ToSlice()...)
	}
	return labels
}

func pointLabels[N int64 | float64](point metricdata.DataPoint[N]) attribute.Set {
	return point.Attributes
}

func histogramLabels[N int64 | float64](point metricdata.HistogramDataPoint[N]) attribute.Set {
	return point.Attributes
}

// A wrong letter is explained by what the task says went wrong on the way to
// it, and the trap behind it is named on a span and in a line only when the
// catalog has it; a letter the task says nothing about is explained by the
// solution alone.
func TestAWrongLetterIsExplainedByWhatTheTaskSays(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name, answer, trap, says string
	}{
		{"a trap the catalog has", "A", "reversed_relation", "Start from what went wrong on the way to it"},
		{"a trap the catalog does not have", "B", other, "Start from what went wrong on the way to it"},
		{"no trap at all", "D", "", "and it is wrong; the right option is C) \"Ben\". Go through the solution"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			p := fuzzProfile(t)
			p.Ratings.Answers = rating.TrialAnswers
			raceWithOtherTraps(t, p)
			h, session := lesson(t, keptAsIs(t, p))
			if text := textOf(t, answerIt(t, session, p.CurrentTask.ID, tc.answer, false)); !strings.Contains(text, tc.says) {
				t.Errorf("the words are %q, want them to say %q", text, tc.says)
			}

			h.settle()
			lines := linesOf(h, "answer_recorded")
			if len(lines) != 1 || lines[0].ContextMap()["trap"] != tc.trap {
				t.Errorf("answer_recorded lines = %v, want one naming the trap %q", lines, tc.trap)
			}
			if got := attributeOf(h.spanNamed(t, "record_answer"), "mathtrail.answer.trap"); got != tc.trap {
				t.Errorf("the span of the answer names the trap %q, want %q", got, tc.trap)
			}
		})
	}
}

// other is what a span and a line name a trap by when the catalog does not
// have it.
const other = "other"

// raceWithOtherTraps puts the race on the card with its wrong options changed:
// A leads to a trap of the catalog, B to one the catalog does not have, and D
// to nothing the task explains. It is handed out for the request the profile
// holds open.
func raceWithOtherTraps(t *testing.T, p *profile.Profile) {
	t.Helper()

	options, _ := raceTask(nil)["options"].(map[string]string)
	if _, err := p.Issue(&profile.Written{
		Wording: raceQuestion, Options: options, Hint: "Who finished before Kim?",
		Fingerprint: "the-race-again", InstructionsVersion: raceInstructions,
	}, profile.TaskSecret{Answer: "C", Solution: raceSolution, Distractors: map[string]profile.Distractor{
		"A": {Trap: "reversed_relation", Text: raceExplained[0]},
		"B": {Trap: "a_trap_nobody_named", Text: raceExplained[1]},
	}}, sealer(t), lessonDay); err != nil {
		t.Fatalf("Issue() error = %v, want the race on the card", err)
	}
}

// A profile that cannot be read is a failure of ours, told in our one
// sentence, and nothing is recorded.
func TestAnAnswerToAProfileThatCannotBeReadIsOurFailure(t *testing.T) {
	t.Parallel()

	h, session := lesson(t, unreadable{memory.New()})
	wantOurSentence(t, answerIt(t, session, "tsk_any", "C", false), "Something went wrong inside MathTrail.")
	h.settle()
	wantFailed(t, h, "submit_answer", "internal")
}

// A task that cannot be checked is taken off the card by a write like any
// other, and a write that lost to another is told as one: the same answer sent
// again finds the task where it was and takes it off then.
func TestATaskThatCannotBeCheckedIsTakenOffByAWriteThatMayLose(t *testing.T) {
	t.Parallel()

	kept := keptWith(t, "masha")
	h, session := lesson(t, conflicted{kept})
	wantOurSentence(t, answerIt(t, session, loadedOf(t, kept).CurrentTask.ID, "C", false),
		"The child's profile was changed somewhere else at the same moment")
	if loadedOf(t, kept).CurrentTask == nil {
		t.Error("the task left the card by a write that was lost")
	}
	h.settle()
	wantFailed(t, h, "submit_answer", "conflict")
}

// A task answered and then left with a seal that no longer opens — the key
// retired between the answer and the same answer sent again — is told that its
// answer counts: only telling it again is gone, and the task leaves the card.
func TestAnAnswerRecordedBeforeTheSealWasLostStillCounts(t *testing.T) {
	t.Parallel()

	p := raceOnTheCard(t, rating.TrialAnswers)
	if _, err := p.Record(profile.Answered{TaskID: p.CurrentTask.ID, Choice: "C", At: lessonDay}, sealer(t)); err != nil {
		t.Fatalf("Record() error = %v, want the race answered", err)
	}
	p.CurrentTask.Sealed = "mt1.t.gone.sealed-by-a-key-nobody-has"
	kept := keptAsIs(t, p)
	h, session := lesson(t, kept)

	wantOurSentence(t, answerIt(t, session, p.CurrentTask.ID, "C", false),
		"This task was answered already and the answer is recorded")
	if after := loadedOf(t, kept); after.CurrentTask != nil || len(after.Recent) != len(p.Recent) {
		t.Errorf("the card holds %+v and the window %d entries, want the task gone and its answer kept",
			after.CurrentTask, len(after.Recent))
	}
	h.settle()
	wantFailed(t, h, "submit_answer", "sealed")
}

// takeOffTheCard leaves a profile with no task on the card and no request open.
func takeOffTheCard(t *testing.T, kept store.Storage) {
	t.Helper()

	p, revision := loadKept(t, kept)
	p.DiscardTask()
	p.OpenRequest = nil
	p.Touch("test", lessonDay)
	if _, err := kept.Save(t.Context(), devAccount, p, revision); err != nil {
		t.Fatalf("Save() error = %v", err)
	}
}

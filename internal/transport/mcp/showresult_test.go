package mcpserver_test

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"reflect"
	"regexp"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/profile"
	"github.com/MathTrail/mathtrail-standalone/internal/domain/rating"
	"github.com/MathTrail/mathtrail-standalone/internal/infra/seal"
	"github.com/MathTrail/mathtrail-standalone/internal/store"
	"github.com/MathTrail/mathtrail-standalone/internal/store/memory"
	mcpserver "github.com/MathTrail/mathtrail-standalone/internal/transport/mcp"
)

// The card of how an answer went: drawn by the model's call once the answer
// is recorded, on the card or in the chat, it shows the whole of it — and
// nothing at all of a task with no answer yet. Writing nothing, it reads an
// answer recorded a moment ago, which may not have reached the file yet.

// shownPayload is what show_result hands its card.
type shownPayload struct {
	Screen     string `json:"screen"`
	Status     string `json:"status"`
	Code       string `json:"code"`
	LastAnswer *struct {
		TaskID  string `json:"task_id"`
		Correct bool   `json:"correct"`
	} `json:"last_answer"`
	Child *struct {
		Pseudonym string `json:"pseudonym"`
		Grade     int    `json:"grade"`
	} `json:"child"`
	Language string `json:"language"`
	Task     *struct {
		ID       string            `json:"id"`
		Topic    string            `json:"topic"`
		Language string            `json:"language"`
		Options  map[string]string `json:"options"`
		Slug     string            `json:"slug"`
		SitePage bool              `json:"site_page"`
	} `json:"task"`
	Result      *resultPayload `json:"result"`
	TopicChoice *struct {
		Chosen      *string  `json:"chosen"`
		Recommended []string `json:"recommended"`
	} `json:"topic_choice"`
	Site *struct {
		URL       string   `json:"url"`
		Languages []string `json:"languages"`
	} `json:"site"`
}

// showIt asks for the card of how the answer to the task with this id went.
func showIt(t *testing.T, session *mcp.ClientSession, id string) *mcp.CallToolResult {
	t.Helper()
	return call(t, session, "show_result", map[string]any{"task_id": id})
}

// lessonShowing serves the lesson over kept, as lesson does, with a pause for
// an answer still on its way short enough not to slow a case, and with
// whatever else the case changes in its parts.
func lessonShowing(t *testing.T, kept store.Storage, changes ...func(*mcpserver.Parts)) (*harness, *mcp.ClientSession) {
	t.Helper()

	h := newHarness(t)
	service := lessonService(t, h, kept, &clock{at: lessonDay}, nil, changes...)
	mcpserver.PauseForAnswer(service, time.Millisecond)
	h.start(t, mcpserver.DevSignIn, slices.Concat(service.ProfileTools(), service.TaskTools())...)
	session := h.connect(t, "")
	return h, session
}

// The card shows how the answer went as the answer was told, whoever recorded
// it: the result, the task's options for the verdict to name, its topic's page,
// whose card it is and in which language, and, once the trial series is over,
// the choice of the topic. Showing it writes nothing.
func TestTheCardOfHowAnAnswerWentShowsIt(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name     string
		answers  int
		answer   string
		says     string
		inSeries bool
	}{
		{name: "right", answers: rating.TrialAnswers, answer: "C", says: "and it is right. Praise briefly"},
		{name: "wrong", answers: rating.TrialAnswers, answer: "A", says: raceExplained[0]},
		{name: "I don't know", answers: rating.TrialAnswers, answer: "?", says: "Without a card, go through the solution"},
		{name: "in the trial series", answer: "B", says: "1 of 5 tasks done", inSeries: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			kept := keptAsIs(t, raceOnTheCard(t, tc.answers))
			_, session := lessonShowing(t, kept)
			id := loadedOf(t, kept).CurrentTask.ID
			told := answered(t, answerIt(t, session, id, tc.answer, false))
			before, _ := profile.Marshal(loadedOf(t, kept))

			shown := showIt(t, session, id)
			payload := payloadOf[shownPayload](t, shown)
			if payload.Screen != "result" || payload.Status != "" || payload.Result == nil || !reflect.DeepEqual(payload.Result, told) {
				t.Fatalf("show_result = %+v, want the result as the answer was told: %+v", payload, told)
			}
			wantTheTaskNamed(t, &payload, id)
			if (payload.TopicChoice == nil) != tc.inSeries {
				t.Errorf("topic_choice = %+v, want one only once the trial series is over", payload.TopicChoice)
			}
			if text := textOf(t, shown); !strings.Contains(text, "the card below shows how the answer to task "+id) ||
				!strings.Contains(text, tc.says) || !strings.Contains(text, "Solution: ") || !strings.Contains(text, notRetold) {
				t.Errorf("the words are %q, want the card named and retold by nobody, %q and the solution", text, tc.says)
			}
			if after, _ := profile.Marshal(loadedOf(t, kept)); !bytes.Equal(after, before) {
				t.Error("show_result changed the profile, want it as the answer left it")
			}
		})
	}
}

// wantTheTaskNamed holds a card of how an answer went to naming its task: the
// race's id, topic and options in its language, the page of its topic on the
// site, and whose card it is.
func wantTheTaskNamed(t *testing.T, payload *shownPayload, id string) {
	t.Helper()

	options, _ := raceTask(nil)["options"].(map[string]string)
	task := payload.Task
	if task == nil || task.ID != id || task.Topic != "logic.ordering" || task.Language != payload.Language ||
		task.Language == "" || len(task.Options) != len(options) || task.Options["C"] != options["C"] {
		t.Errorf("task = %+v in %q, want the race in its language with its options", task, payload.Language)
	}
	if task != nil && (task.Slug == "" || !task.SitePage) {
		t.Errorf("the topic's page = %q, published %v, want the page of Ordering", task.Slug, task.SitePage)
	}
	if payload.Child == nil || payload.Child.Pseudonym == "" || payload.Site == nil || payload.Site.URL == "" {
		t.Errorf("child %+v and site %+v, want whose card it is and the site", payload.Child, payload.Site)
	}
}

// A task on the card with no answer yet is shown not at all: neither the
// payload nor the words hold its right option, its traps or its solution, and
// it is read again before it is told it has none.
func TestNothingIsShownOfATaskBeforeItsAnswer(t *testing.T) {
	t.Parallel()

	counted := &countingLoads{Storage: keptAsIs(t, raceOnTheCard(t, rating.TrialAnswers))}
	_, session := lessonShowing(t, counted)
	shown := showIt(t, session, loadedOf(t, counted).CurrentTask.ID)

	payload := payloadOf[shownPayload](t, shown)
	if payload.Screen != "result" || payload.Status != "rejected" || payload.Code != "not_answered" ||
		payload.Result != nil || payload.Task != nil || payload.Child == nil {
		t.Errorf("show_result = %+v, want not_answered, whose card it is, and nothing of the task", payload)
	}
	options, _ := raceTask(nil)["options"].(map[string]string)
	for _, secret := range append([]string{options["C"], raceSolution}, raceExplained...) {
		if bytes.Contains(rawPayload(t, shown), []byte(secret)) || strings.Contains(textOf(t, shown), secret) {
			t.Errorf("show_result before the answer gave away %q", secret)
		}
	}
	if reads := counted.reads(); reads != 4 {
		t.Errorf("the profile was read %d times, want the three reads of the answer after the session's own", reads)
	}
	if text := textOf(t, shown); !strings.Contains(text, "only just recorded with submit_answer may still be on its way") ||
		!strings.Contains(text, "then, and only then, call show_result again") {
		t.Errorf("show_result says %q, want the model to ask again only after an answer it has just recorded", text)
	}
}

// countingLoads is a store that counts how many times the profile was read.
type countingLoads struct {
	store.Storage
	mu    sync.Mutex
	loads int
}

func (c *countingLoads) Load(ctx context.Context, account store.Account) (*profile.Profile, store.Revision, error) {
	c.mu.Lock()
	c.loads++
	c.mu.Unlock()
	return c.Storage.Load(ctx, account)
}

func (c *countingLoads) reads() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.loads
}

// An answer recorded a moment ago may not be in the file the next read finds:
// the card is told before the answer is written, and another instance may be
// the one writing it. An answer that lands while it is read for is shown; one
// that has not landed by the last read is told as no answer.
func TestAnAnswerStillOnItsWayIsWaitedFor(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name  string
		stale int
		code  string
	}{
		{"it lands before the last read", 2, ""},
		{"it has not landed by the last read", 3, "not_answered"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			p := raceOnTheCard(t, rating.TrialAnswers)
			before, err := profile.Marshal(p)
			if err != nil {
				t.Fatalf("Marshal() error = %v", err)
			}
			if _, err := p.Record(profile.Answered{TaskID: p.CurrentTask.ID, Choice: "A", At: lessonDay},
				sealer(t), rating.GradeLevels()); err != nil {
				t.Fatalf("Record() error = %v", err)
			}
			lagged := &answerLags{Storage: keptAsIs(t, p), before: before, stale: tc.stale}
			_, session := lessonShowing(t, lagged)

			payload := payloadOf[shownPayload](t, showIt(t, session, p.CurrentTask.ID))
			if payload.Code != tc.code || (payload.Result != nil) != (tc.code == "") {
				t.Errorf("show_result = %+v, want the code %q", payload, tc.code)
			}
		})
	}
}

// answerLags is a store whose first reads find the profile as it was before
// the answer: the file another instance has not written yet.
type answerLags struct {
	store.Storage
	before []byte
	mu     sync.Mutex
	stale  int
}

func (l *answerLags) Load(ctx context.Context, account store.Account) (*profile.Profile, store.Revision, error) {
	l.mu.Lock()
	behind := l.stale > 0
	l.stale--
	l.mu.Unlock()
	p, revision, err := l.Storage.Load(ctx, account)
	if err != nil || !behind {
		return p, revision, err
	}
	was, err := profile.Parse(l.before)
	return was, revision, err
}

// A task that is not the one on the card shows nothing; one that left it with
// its answer is told how that answer went, so the model does not say it was
// lost, and the words never repeat an id the profile does not know.
func TestATaskOffTheCardShowsNoMoreThanHowItWent(t *testing.T) {
	t.Parallel()

	p := raceOnTheCard(t, rating.TrialAnswers)
	id := p.CurrentTask.ID
	if _, err := p.Record(profile.Answered{TaskID: id, Choice: "A", At: lessonDay}, sealer(t), rating.GradeLevels()); err != nil {
		t.Fatalf("Record() error = %v", err)
	}
	p.DiscardTask()
	_, session := lessonShowing(t, keptAsIs(t, p))

	for _, tc := range []struct{ name, id, says string }{
		{"left the card with its answer", id, "has left the card since its answer, wrong, was recorded"},
		{"never given", "tsk_never_given", "The task_id is not the task on the child's card"},
	} {
		shown := showIt(t, session, tc.id)
		payload := payloadOf[shownPayload](t, shown)
		if payload.Status != "stale" || payload.Code != "stale_task" || payload.Result != nil || payload.Task != nil {
			t.Errorf("%s: show_result = %+v, want stale_task and nothing of the task", tc.name, payload)
		}
		if text := textOf(t, shown); !strings.Contains(text, tc.says) || strings.Contains(text, "tsk_never_given") {
			t.Errorf("%s: the words are %q, want %q and no id the profile does not know", tc.name, text, tc.says)
		}
	}
}

// An answer whose seal no longer opens counts, and only how it went can no
// longer be told: nothing is shown, and nothing is written.
func TestAnAnswerWhoseSealIsLostIsToldNoMore(t *testing.T) {
	t.Parallel()

	p := raceOnTheCard(t, rating.TrialAnswers)
	if _, err := p.Record(profile.Answered{TaskID: p.CurrentTask.ID, Choice: "C", At: lessonDay},
		sealer(t), rating.GradeLevels()); err != nil {
		t.Fatalf("Record() error = %v", err)
	}
	kept := keptAsIs(t, p)
	before, _ := profile.Marshal(loadedOf(t, kept))
	_, session := lessonShowing(t, kept, func(parts *mcpserver.Parts) { parts.Sealer = retiredSeal(t) })

	payload := payloadOf[shownPayload](t, showIt(t, session, p.CurrentTask.ID))
	if payload.Status != "stale" || payload.Code != "told_no_more" || payload.Result != nil {
		t.Errorf("show_result = %+v, want told_no_more and nothing shown", payload)
	}
	if after, _ := profile.Marshal(loadedOf(t, kept)); !bytes.Equal(after, before) {
		t.Error("show_result wrote the profile, want nothing written")
	}
}

// retiredSeal is a ring that never sealed anything of these cases: the shape a
// retired key leaves behind.
func retiredSeal(t *testing.T) profile.Sealer {
	t.Helper()

	key := sha256.Sum256([]byte("a key these cases never sealed with"))
	ring, err := seal.NewKeyRing(base64.StdEncoding.EncodeToString(key[:]), "")
	if err != nil {
		t.Fatalf("NewKeyRing() error = %v", err)
	}
	return ring.For(seal.PurposeTaskAnswer)
}

// With no profile there is no answer to show, and the card says how one is
// made.
func TestThereIsNoAnswerToShowWithoutAProfile(t *testing.T) {
	t.Parallel()

	_, session := lessonShowing(t, memory.New())
	payload := payloadOf[shownPayload](t, showIt(t, session, "tsk_anything"))
	if payload.Screen != "first_run" || payload.Result != nil {
		t.Errorf("show_result = %+v, want the first run", payload)
	}
}

// The picture of the solution and the total under it are sealed with the
// answer: before the answer they reach no payload, no word for the model and
// no open field of the file; no line ever holds them, and the hand-in's line
// names no more of the picture than its kind and its size. Once the answer is
// in, the reply to it carries them, for the card that took it, and so does the
// card of how it went.
func TestThePictureOfTheSolutionIsSealedUntilTheAnswer(t *testing.T) {
	t.Parallel()

	const marker = "1979"
	kept := racer(t)
	h, session := lessonShowing(t, kept)
	request := askForTheRace(t, session, kept)
	race := raceOn(request)
	task, _ := race["task"].(map[string]any)
	task["solution_picture"] = map[string]any{
		"kind": "row", "items": []map[string]string{{"label": marker}, {"label": "2026"}},
	}
	task["solution_total"] = "1970 + 9 = " + marker

	handed := call(t, session, "submit_task", race)
	card := wantOnTheCard(t, handed)
	asked := call(t, session, "read_task", map[string]any{"request_id": request.ID})
	written, err := profile.Marshal(loadedOf(t, kept))
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}
	before := []string{
		string(rawPayload(t, handed)), textOf(t, handed), string(rawPayload(t, asked)), textOf(t, asked),
		string(writtenWithoutSeal(t, written)),
	}
	for _, seen := range before {
		if strings.Contains(seen, marker) {
			t.Fatalf("before the answer, the picture of the solution shows in %s", seen)
		}
	}

	type pictured struct {
		Result struct {
			SolutionPicture map[string]any `json:"solution_picture"`
			SolutionTotal   string         `json:"solution_total"`
		} `json:"result"`
	}
	answered := answerIt(t, session, card.Task.ID, "A", false)
	shown := showIt(t, session, card.Task.ID)
	for tool, after := range map[string]*mcp.CallToolResult{"submit_answer": answered, "show_result": shown} {
		if shown := payloadOf[pictured](t, after); shown.Result.SolutionPicture["kind"] != "row" ||
			shown.Result.SolutionTotal != "1970 + 9 = "+marker {
			t.Errorf("%s = %+v, want the picture of the solution and its total", tool, shown.Result)
		}
	}

	h.settle()
	for _, line := range h.logs.All() {
		if strings.Contains(fmt.Sprint(line.Message, line.ContextMap()), marker) {
			t.Errorf("the line %s holds the picture of the solution", line.Message)
		}
	}
	submitted := linesOf(h, "task_submitted")
	if fields := submitted[len(submitted)-1].ContextMap(); fields["solution_picture"] != "row" ||
		fields["solution_picture_bytes"] == int64(0) {
		t.Errorf("task_submitted = %v, want the picture's kind and its size", fields)
	}
}

// writtenWithoutSeal is the profile's file as written, with every sealed value
// taken out: what a person who opens the file reads.
func writtenWithoutSeal(t *testing.T, written []byte) []byte {
	t.Helper()
	return regexp.MustCompile(`"mt1\.[^"]*"`).ReplaceAll(written, []byte(`"sealed"`))
}

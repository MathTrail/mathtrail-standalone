package mcpserver_test

import (
	"slices"
	"strings"
	"testing"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/profile"
	"github.com/MathTrail/mathtrail-standalone/internal/domain/tutor"
	"github.com/MathTrail/mathtrail-standalone/internal/store"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// pausedChild is a child of grade 3 past the trial series, shaped like the
// child the probes were made for and made up as every profile of these cases
// is: counting and the calendar within reach, clocks within reach but behind
// the overall rank, ordering on pause, its easiest task at a chance of 0.65,
// and eleven tasks behind the child, so that the next takes a probe's turn.
func pausedChild(t *testing.T) store.Storage {
	t.Helper()

	p := profile.New(profile.Student{Grade: 3, Pseudonym: "Otter"}, "test", lessonDay)
	p.Ratings.Answers = 11
	p.Ratings.Theta = -0.5
	for topic, delta := range map[string]float64{
		"counting.gaps": 0.5, "time.calendar": 0.4, "time.clocks": -0.75, "logic.ordering": -1.25,
	} {
		p.Topics[topic] = profile.Topic{
			Answers: 3, Correct: 2, Delta: delta, LastIssued: profile.DateOf(lessonDay), Traps: map[string]int{},
		}
	}
	return keptAsIs(t, p)
}

// probeCard is what a card of a task is handed that tells a probe from any
// other task.
type probeCard struct {
	Screen string `json:"screen"`
	Task   *struct {
		ID    string `json:"id"`
		Probe bool   `json:"probe"`
	} `json:"task"`
}

// pausedTopic is a topic as get_progress hands its card, and pausedTopics all
// of them: whether each is on pause.
type pausedTopic struct {
	Topic  string `json:"topic"`
	Paused bool   `json:"paused"`
}

type pausedTopics struct {
	Topics []pausedTopic `json:"topics"`
}

// probeOnTheCard asks for the next task of the paused child, which the rule
// makes a probe of ordering, hands the race in for it, and is the request and
// the card the task came to.
func probeOnTheCard(t *testing.T, kept store.Storage) (*harness, *mcp.ClientSession, *profile.OpenRequest, probeCard) {
	t.Helper()

	h, session := lesson(t, kept)
	coming := wantComing(t, call(t, session, "next_task", map[string]any{"language": "en"}))
	fetchPackage(t, session, coming.RequestID)
	p, _ := loadKept(t, kept)
	open := p.OpenRequest
	if open == nil || open.ID != coming.RequestID || open.Brief.TargetConcept != "logic.ordering" ||
		open.Brief.Difficulty != 1 || open.TutorMode != profile.TutorRule {
		t.Fatalf("the request is %+v, want the rule's probe of logic.ordering at difficulty 1", open)
	}
	card := payloadOf[probeCard](t, call(t, session, "submit_task", raceOn(open)))
	if card.Screen != "task" || card.Task == nil {
		t.Fatalf("submit_task = %+v, want the race on the card", card)
	}
	return h, session, open, card
}

func TestAProbeIsMarkedOnEveryCardThatDrawsIt(t *testing.T) {
	t.Parallel()
	kept := pausedChild(t)
	h, session, open, card := probeOnTheCard(t, kept)

	if !card.Task.Probe {
		t.Errorf("submit_task's card = %+v, want the task marked a probe", card.Task)
	}
	if read := payloadOf[probeCard](t, call(t, session, "read_task", map[string]any{"request_id": open.ID})); read.Task == nil ||
		!read.Task.Probe {
		t.Errorf("read_task's card = %+v, want the task marked a probe", read.Task)
	}
	h.settle()
	if requested := linesOf(h, "task_requested"); len(requested) != 1 || requested[0].ContextMap()["probe"] != true {
		t.Errorf("task_requested = %v, want one line, a probe", requested)
	}
}

func TestARightAnswerToAProbeTakesItsTopicOffThePause(t *testing.T) {
	t.Parallel()
	kept := pausedChild(t)
	h, session, _, card := probeOnTheCard(t, kept)

	answered := payloadOf[answerPayload](t, call(t, session, "submit_answer",
		map[string]any{"task_id": card.Task.ID, "answer": "C"}))
	if answered.Result == nil || !answered.Result.Correct {
		t.Fatalf("the answer = %+v, want it recorded as right", answered)
	}
	p, _ := loadKept(t, kept)
	loaded, err := shipped()
	if err != nil {
		t.Fatalf("load the content: %v", err)
	}
	if !slices.Contains(tutor.WithinReach(p, loaded), "logic.ordering") {
		t.Errorf("after a right answer to its probe, logic.ordering stands at %v, out of reach", p.LevelIn("logic.ordering"))
	}
	shown := payloadOf[pausedTopics](t, call(t, session, "get_progress", map[string]any{}))
	for _, topic := range shown.Topics {
		if topic.Paused {
			t.Errorf("get_progress shows %s on pause after the right answer to its probe", topic.Topic)
		}
	}
	h.settle()
	if recorded := linesOf(h, "answer_recorded"); len(recorded) != 1 || recorded[0].ContextMap()["probe"] != true {
		t.Errorf("answer_recorded = %v, want one line, an answer to a probe", recorded)
	}
}

func TestAWrongAnswerToAProbeLeavesItsTopicOnPause(t *testing.T) {
	t.Parallel()
	kept := pausedChild(t)
	_, session, _, card := probeOnTheCard(t, kept)

	call(t, session, "submit_answer", map[string]any{"task_id": card.Task.ID, "answer": "B"})
	result := call(t, session, "get_progress", map[string]any{})
	shown := payloadOf[pausedTopics](t, result)
	paused := slices.IndexFunc(shown.Topics, func(topic pausedTopic) bool {
		return topic.Topic == "logic.ordering" && topic.Paused
	})
	if paused < 0 {
		t.Errorf("get_progress = %+v, want logic.ordering still on pause after a wrong answer to its probe", shown.Topics)
	}
	if text := textOf(t, result); !strings.Contains(text, ", paused") ||
		!strings.Contains(text, "comes back as a probe") {
		t.Errorf("the words of get_progress = %q, want the topic said to be paused and what that means", text)
	}
}

// A task the model set on a topic on pause, with a reason of its own, is the
// model's choice rather than a probe, and its card says nothing of one.
func TestATaskTheModelChoseOnATopicOnPauseIsNoProbe(t *testing.T) {
	t.Parallel()
	kept := pausedChild(t)
	h, session := lesson(t, kept)

	open := wantRequestOpened(t, session, call(t, session, "next_task", raceChoice), kept)
	card := payloadOf[probeCard](t, call(t, session, "submit_task", raceOn(open)))
	if card.Task == nil || card.Task.Probe {
		t.Errorf("the card of the model's task = %+v, want it on the card and not marked a probe", card.Task)
	}
	h.settle()
	if requested := linesOf(h, "task_requested"); len(requested) != 1 || requested[0].ContextMap()["probe"] != false {
		t.Errorf("task_requested = %v, want one line, no probe", requested)
	}
}

func TestTheProgressSaysWhatPauseAndBehindMean(t *testing.T) {
	t.Parallel()
	_, session := lesson(t, pausedChild(t))

	result := call(t, session, "get_progress", map[string]any{})
	shown := payloadOf[pausedTopics](t, result)
	for _, topic := range shown.Topics {
		if want := topic.Topic == "logic.ordering"; topic.Paused != want {
			t.Errorf("get_progress shows %s paused %t, want %t", topic.Topic, topic.Paused, want)
		}
	}
	text := textOf(t, result)
	for _, said := range []string{
		"a topic behind the overall rank waits half as long for its turn",
		"once in 6 tasks a paused topic comes back as a probe",
		"as a probe of a topic on pause",
	} {
		if !strings.Contains(text, said) {
			t.Errorf("the words of get_progress = %q, want them to say %q", text, said)
		}
	}
}

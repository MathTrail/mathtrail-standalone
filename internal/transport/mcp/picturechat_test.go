package mcpserver_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/profile"
	"github.com/MathTrail/mathtrail-standalone/internal/store"
)

// A task solved from the picture holds its facts on the card alone, so it
// reaches only a chat known to draw the card: ChatGPT, or a client that
// declared MCP Apps with its call. Every other chat — one that declared none,
// and one on an earlier protocol, whose calls declare nothing — is a chat of
// words for it.

// threeChats is one lesson over kept, as a chat that draws the cards, as one
// that shows the words alone, and as a client on an earlier protocol, whose
// calls carry nothing the server can read, reach it.
func threeChats(t *testing.T, kept store.Storage) (h *harness, cards, words, unread *mcp.ClientSession) {
	t.Helper()

	h = serveLesson(t, kept, &clock{at: lessonDay}, nil)
	cards, words = h.connectChat(t, "", drawingCards), h.connectChat(t, "", wordsAlone)
	unread = h.connectChat(t, "2025-11-25", drawingCards)
	return h, cards, words, unread
}

// fromThePicture makes the brief of the request open one that asks for a task
// solved from the picture, as the rule writes it for such an idea.
func fromThePicture(p *profile.Profile) {
	p.OpenRequest.Brief.Constraints = []string{profile.FactsOnThePicture}
}

// droppedFor is the reasons and whether each had been written, of the lines
// of the tasks let go.
func droppedFor(h *harness) []string {
	lines := linesOf(h, "task_dropped")
	told := make([]string, 0, len(lines))
	for i := range lines {
		fields := lines[i].ContextMap()
		written := map[bool]string{true: "written", false: "being written"}[fields["written"] == true]
		told = append(told, fmt.Sprint(fields["reason"], " ", written))
	}
	return told
}

// A request open for a task solved from the picture is handed its package, and
// handed back as it was, only in a chat that draws the card. Anywhere else it
// is gone: its package is refused as a request past its window is, a task
// handed in for it is told the same, and nothing is written; asked for a task
// there, the service lets it go before anything else and opens one in words.
func TestARequestForATaskSolvedFromThePictureReachesOnlyAChatThatDrawsTheCard(t *testing.T) {
	t.Parallel()

	kept := racer(t)
	h, cards, words, unread := threeChats(t, kept)
	request := askForTheRace(t, cards, kept)
	changeKept(t, kept, fromThePicture)
	_, revision := loadKept(t, kept)

	for name, session := range map[string]*mcp.ClientSession{"words alone": words, "an earlier protocol": unread} {
		gone := payloadOf[requestPayload](t, call(t, session, "get_package", map[string]any{"request_id": request.ID}))
		if gone.Status != "stale" || gone.Code != "stale_request" {
			t.Errorf("%s: get_package = %+v, want the request gone, stale_request", name, gone)
		}
		handed := payloadOf[handedInPayload](t, call(t, session, "submit_task", raceOn(request)))
		if handed.Status != "stale" || handed.Code != "stale_request" {
			t.Errorf("%s: submit_task = %+v, want the request gone, stale_request", name, handed)
		}
	}
	if _, now := loadKept(t, kept); now != revision {
		t.Fatal("a chat of words wrote the profile, want nothing written")
	}
	if text := fetchPackage(t, cards, request.ID); !strings.Contains(text, "Package:") {
		t.Errorf("get_package in a chat with cards said %q, want the package", text)
	}
	again := wantComing(t, call(t, cards, "next_task", map[string]any{"language": "en"}))
	if !again.AlreadyOpen || again.RequestID != request.ID {
		t.Errorf("next_task in a chat with cards = %+v, want the request handed back", again)
	}

	asked := wantComing(t, call(t, words, "next_task", map[string]any{"language": "en"}))
	p, _ := loadKept(t, kept)
	if asked.AlreadyOpen || asked.RequestID == request.ID || p.OpenRequest == nil || p.OpenRequest.ID != asked.RequestID ||
		p.OpenRequest.Brief.FromPicture() {
		t.Fatalf("next_task in a chat of words = %+v with %+v open, want the request let go and one in words opened",
			asked, p.OpenRequest)
	}
	h.settle()
	if got := droppedFor(h); len(got) != 1 || got[0] != "chat being written" {
		t.Errorf("task_dropped lines = %v, want the request let go for the chat", got)
	}
}

// A task solved from the picture written ahead and kept is let go when the
// next task is asked for in a chat that does not draw the card, and a task in
// words is asked for in its place. Where the card is drawn it is handed out,
// its words saying it is not read out, and its line saying it was solved from
// the picture.
func TestATaskSolvedFromThePictureKeptIsHandedOutOnlyWhereTheCardIsDrawn(t *testing.T) {
	t.Parallel()

	t.Run("in a chat of words", func(t *testing.T) {
		t.Parallel()

		kept := racer(t)
		h, cards, words, _ := threeChats(t, kept)
		keepTheRelay(t, cards, kept)
		changeKept(t, kept, func(p *profile.Profile) { p.ReadyTask.FromPicture = true })

		coming := wantComing(t, call(t, words, "next_task", map[string]any{"language": "en"}))
		p, _ := loadKept(t, kept)
		if p.ReadyTask != nil || p.OpenRequest == nil || coming.RequestID != p.OpenRequest.ID {
			t.Fatalf("next_task = %+v, the profile keeping %+v, want the task kept let go and one asked for",
				coming, p.ReadyTask)
		}
		h.settle()
		if got := droppedFor(h); len(got) != 1 || got[0] != "chat written" {
			t.Errorf("task_dropped lines = %v, want the task kept let go for the chat", got)
		}
	})
	t.Run("in a chat with cards", func(t *testing.T) {
		t.Parallel()

		kept := racer(t)
		h, cards, _, _ := threeChats(t, kept)
		_, relay := keepTheRelay(t, cards, kept)
		changeKept(t, kept, func(p *profile.Profile) { p.ReadyTask.FromPicture = true })

		result := call(t, cards, "next_task", map[string]any{"language": "en"})
		card, text := payloadOf[requestPayload](t, result), textOf(t, result)
		if card.Screen != "task" || card.Task == nil || card.Task.ID != relay {
			t.Fatalf("next_task = %+v, want the task kept handed out", card)
		}
		if !strings.Contains(text, "This task is solved from its picture, which no words carry") ||
			strings.Contains(text, "read out the question and the options") {
			t.Errorf("next_task said %q, want the task not read out, and another asked for should no card show it", text)
		}
		h.settle()
		accepted := linesOf(h, "task_accepted")
		if last := accepted[len(accepted)-1].ContextMap(); last["from_picture"] != true || last["ready"] != true {
			t.Errorf("the last task_accepted line = %v, want the task kept, solved from the picture", last)
		}
	})
}

// A task written ahead is still being written when a chat that does not draw
// the card asks for the next one ahead: its request, for a task solved from
// the picture, is let go, and one in words is opened in its place.
func TestATaskSolvedFromThePictureBeingWrittenAheadIsLetGoInAChatOfWords(t *testing.T) {
	t.Parallel()

	kept := racer(t)
	h, cards, words, _ := threeChats(t, kept)
	raceHandedOut(t, cards, kept)
	prepared(t, cards, aheadChoice)
	ahead := aheadOpen(t, kept)
	changeKept(t, kept, fromThePicture)

	prepared(t, words, map[string]any{"language": "en"})
	p, _ := loadKept(t, kept)
	if p.OpenRequest == nil || p.OpenRequest.ID == ahead.ID || !p.OpenRequest.Ahead || p.OpenRequest.Brief.FromPicture() {
		t.Fatalf("the request open is %+v, want the one ahead let go and one in words opened ahead", p.OpenRequest)
	}
	h.settle()
	if got := droppedFor(h); len(got) != 1 || got[0] != "chat being written" {
		t.Errorf("task_dropped lines = %v, want the request ahead let go for the chat", got)
	}
}

// A hand-in for a request for a task solved from the picture says so on its
// line, as the task it puts on the card does, so that such tasks, their
// refusals and their seconds are counted apart; one for a task in words says
// it is not.
func TestTheLinesOfATaskSayWhetherItIsSolvedFromThePicture(t *testing.T) {
	t.Parallel()

	for _, fromPicture := range []bool{false, true} {
		t.Run(map[bool]string{false: "in words", true: "from the picture"}[fromPicture], func(t *testing.T) {
			t.Parallel()

			kept := racer(t)
			h, cards, _, _ := threeChats(t, kept)
			request := askForTheRace(t, cards, kept)
			if fromPicture {
				changeKept(t, kept, fromThePicture)
			}
			call(t, cards, "submit_task", raceOn(request))
			if got := submittedLine(t, h)["from_picture"]; got != fromPicture {
				t.Errorf("task_submitted from_picture = %v, want %v", got, fromPicture)
			}
		})
	}
}

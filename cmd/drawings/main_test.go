package main

import (
	"context"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"

	"github.com/MathTrail/mathtrail-standalone/content"
	"github.com/MathTrail/mathtrail-standalone/internal/domain/checks"
)

// theSet is the calibration set, with the frames the content ships.
func theSet(t *testing.T) []sample {
	t.Helper()

	loaded, err := content.Load()
	if err != nil {
		t.Fatalf("content.Load() error = %v", err)
	}
	return calibrationSet(loaded.Frames())
}

// Every drawing of the set is exactly as large as it says, measured as the
// checks measure a drawing, and uses no character a drawing may not — but the
// one that shows those characters for comparison, which is refused for them
// alone: a rating is only worth what the size it is written down against is.
func TestEveryDrawingIsAsLargeAsItSays(t *testing.T) {
	t.Parallel()

	for _, drawing := range theSet(t) {
		t.Run(drawing.name, func(t *testing.T) {
			t.Parallel()

			fits := checks.DrawingLimits{
				Width: drawing.width, Height: drawing.height, SpaceRun: checks.DefaultDrawingLimits().SpaceRun,
			}
			problems := checks.DrawingFormat(drawing.drawing, fits)
			switch {
			case drawing.outOfSet && (len(problems) != 1 || !strings.Contains(problems[0].Message, "may not use")):
				t.Errorf("DrawingFormat() at %dx%d = %v, want it refused for its characters alone",
					drawing.width, drawing.height, problems)
			case !drawing.outOfSet && len(problems) != 0:
				t.Errorf("DrawingFormat() at %dx%d = %v, want none", drawing.width, drawing.height, problems)
			}
			narrower, shorter := fits, fits
			narrower.Width--
			shorter.Height--
			if len(checks.DrawingFormat(drawing.drawing, narrower)) == len(problems) {
				t.Errorf("the drawing fits %d cells, want it %d wide", narrower.Width, drawing.width)
			}
			if len(checks.DrawingFormat(drawing.drawing, shorter)) == len(problems) {
				t.Errorf("the drawing fits %d lines, want it %d tall", shorter.Height, drawing.height)
			}
		})
	}
}

// The set holds every width, every height, the four classes of characters, the
// characters left out beside them, and both frames: a frame renamed in the
// content would otherwise drop out of it unseen.
func TestTheSetHoldsEveryCase(t *testing.T) {
	t.Parallel()

	widths, heights := ladders()
	if got, want := len(theSet(t)), len(widths)+len(heights)+5+2; got != want {
		t.Errorf("the set holds %d drawings, want %d", got, want)
	}
}

// A drawing is shown on a task card as the widget reads one, and the rating
// the card sends back is written down against the drawing and the surface its
// card was shown on.
func TestACardIsShownAndItsRatingTaken(t *testing.T) {
	t.Parallel()

	set := theSet(t)
	endpoint := httptest.NewServer(mcp.NewStreamableHTTPHandler(
		func(*http.Request) *mcp.Server { return newServer(set, zap.NewNop()) },
		&mcp.StreamableHTTPOptions{Stateless: true},
	))
	t.Cleanup(endpoint.Close)
	client := mcp.NewClient(&mcp.Implementation{Name: "test", Version: "1"}, nil)
	session, err := client.Connect(t.Context(), &mcp.StreamableClientTransport{Endpoint: endpoint.URL}, nil)
	if err != nil {
		t.Fatalf("Connect() error = %v", err)
	}
	t.Cleanup(func() { _ = session.Close() })

	shown, err := session.CallTool(t.Context(), &mcp.CallToolParams{
		Name: "show_drawing", Arguments: map[string]any{"number": 3, "surface": "claude-phone"},
	})
	if err != nil || shown.IsError {
		t.Fatalf("show_drawing = %+v, %v, want a card", shown, err)
	}
	var card cardOut
	decode(t, shown.StructuredContent, &card)
	if card.Screen != "task" || card.Task.Drawing != set[2].drawing || len(card.Task.Options) != 5 ||
		card.Task.ID != "drawing-3-claude-phone" {
		t.Errorf("the card = %+v, want drawing 3 on a task card for claude-phone", card)
	}

	rated, err := session.CallTool(t.Context(), &mcp.CallToolParams{
		Name: "submit_answer", Arguments: map[string]any{"task_id": card.Task.ID, "answer": "c"},
	})
	if err != nil || rated.IsError {
		t.Fatalf("submit_answer = %+v, %v, want the rating taken", rated, err)
	}
	var result ratedOut
	decode(t, rated.StructuredContent, &result)
	if result.Screen != "result" || result.Result.CorrectAnswer != "C" || result.Result.TaskID != card.Task.ID ||
		result.Result.Solution != "Written down: Scrolls sideways." {
		t.Errorf("the rating = %+v, want C written down for the card", result)
	}
}

// "I don't know", or anything but one of the five options, is no rating, and
// neither is a rating of a card the set did not draw: each is refused, so that
// the card lets the adult rate again, and nothing is written down.
func TestOnlyAnOptionOfACardOfTheSetIsARating(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name, taskID, answer string
	}{
		{"I don't know", taskID("claude-web", 1), "?"},
		{"a letter past the options", taskID("claude-web", 1), "F"},
		{"no answer at all", taskID("claude-web", 1), ""},
		{"a card of no drawing of the set", taskID("claude-web", 99), "A"},
		{"a card of a surface not named so", "drawing-1-Claude web", "A"},
		{"a card the set did not draw", "tsk_other", "A"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			core, logs := observer.New(zapcore.InfoLevel)
			shows := &shower{set: theSet(t), log: zap.New(core)}
			if _, _, err := shows.rate(t.Context(), &mcp.CallToolRequest{}, rateIn{TaskID: tc.taskID, Answer: tc.answer}); err == nil {
				t.Errorf("rate(%q, %q) error = nil, want it refused as no rating", tc.taskID, tc.answer)
			}
			if rated := logs.FilterMessage("rated").Len(); rated != 0 {
				t.Errorf("rated lines = %d, want none written down", rated)
			}
		})
	}
}

// A drawing is shown only for a surface named as the tool names it: a rating
// is grouped by its surface, and a surface named another way would make a
// group of its own.
func TestADrawingIsShownOnlyForANamedSurface(t *testing.T) {
	t.Parallel()

	for _, surface := range []string{" ", "Claude web", "claude_web"} {
		t.Run(surface, func(t *testing.T) {
			t.Parallel()

			shows := &shower{set: theSet(t), log: zap.NewNop()}
			if _, _, err := shows.show(t.Context(), &mcp.CallToolRequest{}, showIn{Number: 1, Surface: surface}); err == nil {
				t.Errorf("show() for surface %q error = nil, want a refusal", surface)
			}
		})
	}
}

// A drawing the set does not have is refused, and no card is drawn.
func TestADrawingOutsideTheSetIsRefused(t *testing.T) {
	t.Parallel()

	for _, number := range []int{0, len(theSet(t)) + 1} {
		if _, _, err := (&shower{set: theSet(t), log: zap.NewNop()}).show(
			t.Context(), &mcp.CallToolRequest{}, showIn{Number: number, Surface: "claude-web"},
		); err == nil {
			t.Errorf("show(%d) error = nil, want a refusal", number)
		}
	}
}

// A card's id carries its drawing and its surface there and back, a surface
// with a dash in it included; an id that names no surface names an unknown
// one.
func TestACardsIDCarriesItsDrawingAndSurface(t *testing.T) {
	t.Parallel()

	surface, number := fromTaskID(taskID("chatgpt-web", 12))
	if surface != "chatgpt-web" || number != 12 {
		t.Errorf("fromTaskID(taskID(chatgpt-web, 12)) = %q, %d, want chatgpt-web, 12", surface, number)
	}
	if surface, number := fromTaskID("tsk_other"); surface != "unknown" || number != 0 {
		t.Errorf("fromTaskID(tsk_other) = %q, %d, want unknown, 0", surface, number)
	}
	if surface, number := fromTaskID("drawing-5"); surface != "unknown" || number != 5 {
		t.Errorf("fromTaskID(drawing-5) = %q, %d, want unknown, 5", surface, number)
	}
}

// A call that carries neither the host's name nor a user agent — a host that
// said its name only when it connected — is still written down, its host as
// unknown.
func TestACallThatNamesNoHostIsWrittenDownAsUnknown(t *testing.T) {
	t.Parallel()

	core, logs := observer.New(zapcore.InfoLevel)
	shows := &shower{set: theSet(t), log: zap.New(core)}
	if _, _, err := shows.show(t.Context(), &mcp.CallToolRequest{}, showIn{Number: 1, Surface: "claude-web"}); err != nil {
		t.Fatalf("show() error = %v, want a card", err)
	}
	shown := logs.FilterMessage("shown").All()
	if len(shown) != 1 {
		t.Fatalf("shown lines = %d, want 1", len(shown))
	}
	if fields := shown[0].ContextMap(); fields["client"] != "unknown" || fields["user_agent"] != "" {
		t.Errorf("the line names client %v and user agent %q, want unknown and none",
			fields["client"], fields["user_agent"])
	}
}

// The server serves the set — its cards on the protocol's endpoint, and its
// health — until it is told to stop, and then stops without an error.
func TestTheSetIsServedUntilTheServerIsStopped(t *testing.T) {
	t.Parallel()

	addr := freeAddress(t)
	ctx, cancel := context.WithCancel(t.Context())
	stopped := make(chan error, 1)
	go func() { stopped <- run(ctx, addr, zap.NewNop()) }()
	waitUntilHealthy(t, addr)

	client := mcp.NewClient(&mcp.Implementation{Name: "test", Version: "1"}, nil)
	session, err := client.Connect(t.Context(), &mcp.StreamableClientTransport{Endpoint: "http://" + addr + "/mcp"}, nil)
	if err != nil {
		t.Fatalf("Connect() error = %v", err)
	}
	shown, err := session.CallTool(t.Context(), &mcp.CallToolParams{
		Name: "show_drawing", Arguments: map[string]any{"number": 1, "surface": "chatgpt-phone"},
	})
	_ = session.Close()
	if err != nil || shown.IsError {
		t.Fatalf("show_drawing = %+v, %v, want a card", shown, err)
	}
	var card cardOut
	decode(t, shown.StructuredContent, &card)
	if card.Task.ID != taskID("chatgpt-phone", 1) {
		t.Errorf("the card's id = %q, want %q", card.Task.ID, taskID("chatgpt-phone", 1))
	}

	cancel()
	select {
	case err := <-stopped:
		if err != nil {
			t.Errorf("run() error = %v, want nil once told to stop", err)
		}
	case <-time.After(10 * time.Second):
		t.Fatal("run() did not return within 10s of being told to stop")
	}
}

// An address the server cannot listen on is reported, and nothing is served.
func TestAnAddressItCannotListenOnIsReported(t *testing.T) {
	t.Parallel()

	var config net.ListenConfig
	taken, err := config.Listen(t.Context(), "tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen for an address to take: %v", err)
	}
	t.Cleanup(func() { _ = taken.Close() })

	if err := run(t.Context(), taken.Addr().String(), zap.NewNop()); err == nil ||
		!strings.Contains(err.Error(), "drawings: listen") {
		t.Errorf("run() on a taken address error = %v, want it refused to listen", err)
	}
}

// freeAddress is a loopback address nothing listens on yet.
func freeAddress(t *testing.T) string {
	t.Helper()

	var config net.ListenConfig
	probe, err := config.Listen(t.Context(), "tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen for a free address: %v", err)
	}
	addr := probe.Addr().String()
	if err := probe.Close(); err != nil {
		t.Fatalf("close the probe listener: %v", err)
	}
	return addr
}

// waitUntilHealthy waits until the server at addr answers its health check.
func waitUntilHealthy(t *testing.T, addr string) {
	t.Helper()

	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		if healthy(t, addr) {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("nothing answers the health check at %s after 10s", addr)
}

// healthy says whether the server at addr answers its health check with ok.
func healthy(t *testing.T, addr string) bool {
	t.Helper()

	request, err := http.NewRequestWithContext(t.Context(), http.MethodGet, "http://"+addr+"/healthz", http.NoBody)
	if err != nil {
		t.Fatalf("build the health check: %v", err)
	}
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return false
	}
	defer func() { _ = response.Body.Close() }()
	body, err := io.ReadAll(response.Body)
	return err == nil && response.StatusCode == http.StatusOK && string(body) == "ok"
}

// decode reads a result's structured content into out.
func decode(t *testing.T, structured, out any) {
	t.Helper()

	raw, err := json.Marshal(structured)
	if err != nil {
		t.Fatalf("encode the structured content: %v", err)
	}
	if err := json.Unmarshal(raw, out); err != nil {
		t.Fatalf("decode the structured content %s: %v", raw, err)
	}
}

// Command drawings shows a chat host the calibration set of text drawings,
// each on the widget's own task card, so that the limits a drawing is held to
// can be chosen by how the cards and the chat show them on every surface. The
// adult rates each card with its five options, and every rating is written to
// the log with the surface it was given on.
//
// It has no sign-in and keeps nothing: it is run behind a tunnel for a look,
// and taken down.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"slices"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"go.uber.org/zap"

	"github.com/MathTrail/mathtrail-standalone/content"
	"github.com/MathTrail/mathtrail-standalone/internal/logger"
	mcpserver "github.com/MathTrail/mathtrail-standalone/internal/transport/mcp"
	"github.com/MathTrail/mathtrail-standalone/internal/widget"
)

// widgetURI names the page the cards are drawn by. It is not the service's
// own address, so that a host keeping pages by their address never takes one
// for the other.
const widgetURI = "ui://mathtrail-drawings/app.html"

// instructions tell the model what the server is for and what to do.
const instructions = "This server shows a calibration set of text drawings for MathTrail, one drawing per card. " +
	"When the adult asks to see drawings, call show_drawing once for each number they name, in order, " +
	"passing the surface they say they are looking at. After each call, reproduce the drawing from its result " +
	"exactly as it is, in a code block, and add nothing else. The adult rates each card with its options; " +
	"do not rate the drawings yourself."

// surfaces are what the adult may be looking at: every rating is grouped by
// one of them, so a surface named any other way is refused rather than made a
// group of its own.
var surfaces = []string{"claude-web", "claude-phone", "chatgpt-web", "chatgpt-phone"}

// surfacesNamed is the surfaces as a sentence names them.
func surfacesNamed() string {
	return strings.Join(surfaces[:len(surfaces)-1], ", ") + " or " + surfaces[len(surfaces)-1]
}

// ratings are the five options of every card: how the drawing looks on it.
var ratings = map[string]string{
	"A": "Fits, edges straight",
	"B": "Fits, edges off",
	"C": "Scrolls sideways",
	"D": "Too small to read",
	"E": "Something else",
}

// shutdownGrace is how long a request under way is given to finish when the
// server is told to stop: a rating on its way is the run's result.
const shutdownGrace = 5 * time.Second

// main owns what a process owns and a function should not: its flags, the
// signals that stop it and the code it exits with.
func main() {
	addr := flag.String("addr", "127.0.0.1:8090", "the address to listen on")
	flag.Parse()

	log, err := logger.New("info", "json")
	if err != nil {
		fmt.Fprintln(os.Stderr, "drawings: logger:", err)
		os.Exit(1)
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	err = run(ctx, *addr, log)
	stop()
	if err != nil {
		log.Error("drawings stopped", zap.Error(err))
		os.Exit(1)
	}
}

// run serves the set until ctx ends, and lets a request under way finish
// before it returns.
func run(ctx context.Context, addr string, log *zap.Logger) error {
	loaded, err := content.Load()
	if err != nil {
		return fmt.Errorf("drawings: load the content: %w", err)
	}
	set := calibrationSet(loaded.Frames())

	protocol := newServer(set, log)
	mux := http.NewServeMux()
	mux.Handle("/mcp", mcp.NewStreamableHTTPHandler(
		func(*http.Request) *mcp.Server { return protocol },
		&mcp.StreamableHTTPOptions{Stateless: true},
	))
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("ok"))
	})
	server := &http.Server{Addr: addr, Handler: mux, ReadHeaderTimeout: 10 * time.Second}

	listener, err := new(net.ListenConfig).Listen(ctx, "tcp", addr)
	if err != nil {
		return fmt.Errorf("drawings: listen: %w", err)
	}
	stopped := make(chan struct{})
	go func() {
		defer close(stopped)
		<-ctx.Done()
		shutdown, cancel := context.WithTimeout(context.WithoutCancel(ctx), shutdownGrace)
		defer cancel()
		_ = server.Shutdown(shutdown)
	}()
	log.Info("listening", zap.String("addr", addr), zap.Int("drawings", len(set)))
	if err := server.Serve(listener); !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("drawings: serve: %w", err)
	}
	<-stopped
	return nil
}

// newServer is the protocol's server: the widget's page, as the service
// describes it, and the two tools.
func newServer(set []sample, log *zap.Logger) *mcp.Server {
	server := mcp.NewServer(
		&mcp.Implementation{Name: "mathtrail-drawings", Title: "MathTrail drawings", Version: "calibration"},
		&mcp.ServerOptions{
			Instructions: instructions,
			Capabilities: &mcp.ServerCapabilities{Tools: &mcp.ToolCapabilities{}, Resources: &mcp.ResourceCapabilities{}},
		},
	)
	// Run behind a tunnel for a look, it has no origin of its own to name, and
	// its cards link nowhere.
	mcpserver.AddWidget(server, widgetURI, widget.Page(), mcpserver.Sandbox{})
	shows := &shower{set: set, log: log}
	no := false
	mcp.AddTool(server, &mcp.Tool{
		Name:  "show_drawing",
		Title: "Show a calibration drawing",
		Description: fmt.Sprintf("Shows drawing number `number` of the calibration set, from 1 to %d, on a card. "+
			"`surface` is where the adult is looking: %s.", len(set), surfacesNamed()),
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true, IdempotentHint: true, DestructiveHint: &no, OpenWorldHint: &no},
		Meta:        mcp.Meta{"ui": map[string]any{"resourceUri": widgetURI}, "ui/resourceUri": widgetURI},
	}, shows.show)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "submit_answer",
		Title:       "Rate a calibration drawing",
		Description: "The card's rating of a drawing. Only the card calls it.",
		Annotations: &mcp.ToolAnnotations{DestructiveHint: &no, OpenWorldHint: &no},
		Meta:        mcp.Meta{"ui": map[string]any{"visibility": []string{"app"}}},
	}, shows.rate)
	return server
}

// shower shows the drawings of the set and writes down their ratings.
type shower struct {
	set []sample
	log *zap.Logger
}

// showIn is what show_drawing takes. The surface is required: a rating is
// worth only what it can be told apart by.
type showIn struct {
	Number  int    `json:"number" jsonschema:"which drawing of the set, from 1"`
	Surface string `json:"surface" jsonschema:"where the adult is looking: claude-web, claude-phone, chatgpt-web or chatgpt-phone"`
}

// cardOut is a task card, as the widget reads one.
type cardOut struct {
	Screen string   `json:"screen"`
	Child  childOut `json:"child"`
	Task   taskOut  `json:"task"`
}

type childOut struct {
	Pseudonym  string `json:"pseudonym"`
	Grade      int    `json:"grade"`
	UILanguage string `json:"ui_language"`
}

type taskOut struct {
	ID       string            `json:"id"`
	Topic    string            `json:"topic"`
	Language string            `json:"language"`
	Question string            `json:"question"`
	Drawing  string            `json:"drawing"`
	Options  map[string]string `json:"options"`
	Hint     string            `json:"hint"`
}

func (s *shower) show(_ context.Context, req *mcp.CallToolRequest, in showIn) (*mcp.CallToolResult, cardOut, error) {
	if in.Number < 1 || in.Number > len(s.set) {
		return nil, cardOut{}, fmt.Errorf("there are %d drawings, numbered from 1", len(s.set))
	}
	surface := strings.TrimSpace(in.Surface)
	if !slices.Contains(surfaces, surface) {
		return nil, cardOut{}, fmt.Errorf("name the surface the adult is looking at as one of %s", surfacesNamed())
	}
	drawing := s.set[in.Number-1]
	s.log.Info("shown", zap.String("surface", surface), zap.Int("drawing", in.Number),
		zap.String("name", drawing.name), zap.String("client", clientOf(req)), zap.String("user_agent", userAgentOf(req)))
	text := fmt.Sprintf("Drawing %d of %d: %s, %d cells wide and %d lines tall. "+
		"Reproduce it exactly as it is, in a code block, and add nothing else:\n```\n%s\n```",
		in.Number, len(s.set), drawing.name, drawing.width, drawing.height, drawing.drawing)
	card := cardOut{
		Screen: "task",
		Child:  childOut{Pseudonym: "Calibration", Grade: 3, UILanguage: "en"},
		Task: taskOut{
			ID:       taskID(surface, in.Number),
			Topic:    "calibration",
			Language: "en",
			Question: fmt.Sprintf("Drawing %d of %d: %s. %s", in.Number, len(s.set), drawing.name, drawing.look),
			Drawing:  drawing.drawing,
			Options:  ratings,
			Hint:     "Rate the drawing with one of the options: the rating is written down for this surface.",
		},
	}
	return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: text}}}, card, nil
}

// rateIn is what the card sends when an option is pressed.
type rateIn struct {
	TaskID   string `json:"task_id"`
	Answer   string `json:"answer"`
	HintUsed bool   `json:"hint_used,omitempty"`
}

// ratedOut is the card's answer, as the widget reads one: the rating written
// down, shown as the right option with the rating in words as its solution.
type ratedOut struct {
	Screen string    `json:"screen"`
	Result resultOut `json:"result"`
}

type resultOut struct {
	TaskID        string `json:"task_id"`
	Topic         string `json:"topic"`
	Choice        string `json:"choice"`
	Correct       bool   `json:"correct"`
	CorrectAnswer string `json:"correct_answer"`
	// A rating has no trap, no rating in a topic and no trial series; the card
	// reads each as an object or null, and here it is always null.
	Trap            any    `json:"trap"`
	Solution        string `json:"solution"`
	HintUsed        bool   `json:"hint_used"`
	Rating          any    `json:"rating"`
	Trial           any    `json:"trial"`
	AlreadyAnswered bool   `json:"already_answered"`
}

// rate writes a card's rating down. Anything but one of the five options is no
// rating: it is refused, and the card says so and lets the adult press an
// option instead.
func (s *shower) rate(_ context.Context, req *mcp.CallToolRequest, in rateIn) (*mcp.CallToolResult, ratedOut, error) {
	surface, number := fromTaskID(in.TaskID)
	if number < 1 || number > len(s.set) || !slices.Contains(surfaces, surface) {
		return nil, ratedOut{}, errors.New("not rated: the card is not one of this set's")
	}
	rating := strings.ToUpper(strings.TrimSpace(in.Answer))
	meaning, known := ratings[rating]
	if !known {
		return nil, ratedOut{}, errors.New("not rated: rate the drawing with one of its five options")
	}
	name := s.set[number-1].name
	s.log.Info("rated", zap.String("surface", surface), zap.Int("drawing", number), zap.String("name", name),
		zap.String("rating", rating), zap.String("meaning", meaning),
		zap.String("client", clientOf(req)), zap.String("user_agent", userAgentOf(req)))

	result := resultOut{
		TaskID: in.TaskID, Topic: "calibration", Choice: rating, Correct: true, CorrectAnswer: rating,
		Solution: "Written down: " + meaning + ".", HintUsed: in.HintUsed,
	}
	text := fmt.Sprintf("The adult rated drawing %d on %s: %s.", number, surface, meaning)
	return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: text}}},
		ratedOut{Screen: "result", Result: result}, nil
}

// taskID ties a card to the surface it was shown on and the drawing it shows,
// so that its rating, which carries nothing else, is written down against both.
func taskID(surface string, number int) string {
	return fmt.Sprintf("drawing-%d-%s", number, surface)
}

// fromTaskID is the surface and the drawing a card's id names, unknown and 0
// for an id it did not make.
func fromTaskID(id string) (surface string, number int) {
	rest, found := strings.CutPrefix(id, "drawing-")
	if !found {
		return "unknown", 0
	}
	digits, surface, _ := strings.Cut(rest, "-")
	number, _ = strconv.Atoi(digits)
	if surface == "" {
		surface = "unknown"
	}
	return surface, number
}

// clientOf is the name the host gave itself, or unknown: a host that speaks an
// older protocol says it only when it connects, and not with every call.
func clientOf(req *mcp.CallToolRequest) string {
	if info := req.ClientInfo(); info != nil && info.Name != "" {
		return info.Name
	}
	return "unknown"
}

// userAgentOf is the agent the call's request named, which tells hosts apart
// where the name they gave themselves is not known.
func userAgentOf(req *mcp.CallToolRequest) string {
	if req.Extra == nil || req.Extra.Header == nil {
		return ""
	}
	return req.Extra.Header.Get("User-Agent")
}

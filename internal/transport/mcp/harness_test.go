package mcpserver_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"

	"github.com/MathTrail/mathtrail-standalone/internal/ratelimit/ratelimittest"
	"github.com/MathTrail/mathtrail-standalone/internal/store"
	"github.com/MathTrail/mathtrail-standalone/internal/telemetry"
	httpserver "github.com/MathTrail/mathtrail-standalone/internal/transport/http"
	mcpserver "github.com/MathTrail/mathtrail-standalone/internal/transport/mcp"
)

// The framework keeps its mode in a package variable. A test binary sets it
// once, here, rather than letting a constructor decide it for every other test
// running beside it.
func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)
	os.Exit(m.Run())
}

// What a child could be recognised by, and what must stay hidden until the
// child has answered. The tools of these cases are handed both, and neither
// may reach a span or a line.
const (
	pseudonym = "masha-ivanova"
	answer    = "the answer is C"
)

// said is what the cases ask the tools to repeat.
const said = pseudonym + " says " + answer

const (
	projectID           = "a-project"
	instructionsVersion = "0123456789ab"
	instructions        = "You coach one child through short maths tasks."
	// page stands in for the widget every card is drawn by.
	page = "<!doctype html><title>the widget</title>"
)

// sayIn is what the tools of these cases take.
type sayIn struct {
	Say string `json:"say"`
}

// sayOut is what most of them give back: what they were told, whom for, and a
// status when they refuse.
type sayOut struct {
	Status string `json:"status,omitempty"`
	Said   string `json:"said"`
	For    string `json:"for"`
}

// nestedOut carries a status deeper down, which is not the payload's own.
type nestedOut struct {
	Inner sayOut `json:"inner"`
}

// brokenOut declares a number in its schema and writes a text: a payload that
// does not fit the tool's own output schema.
type brokenOut struct {
	Count int `json:"count"`
}

func (brokenOut) MarshalJSON() ([]byte, error) {
	return []byte(`{"count":"` + pseudonym + `"}`), nil
}

// tools are the tools every case is served, one for each way a call can end.
func tools() []mcpserver.Tool {
	return []mcpserver.Tool{
		mcpserver.Define(mcpserver.Spec{
			Name: "echo", Title: "Echo", Description: "Says back what it was given.",
			ReadOnly: true, Idempotent: true, DrawsCard: true,
		}, func(ctx context.Context, account store.Account, in sayIn) (mcpserver.Reply[sayOut], error) {
			// Whatever a tool does is recorded inside the tool's own span.
			_, work := trace.SpanFromContext(ctx).TracerProvider().Tracer("test").Start(ctx, "work")
			work.End()
			return mcpserver.Reply[sayOut]{Text: "said: " + in.Say, Payload: sayOut{Said: in.Say, For: account.ID}}, nil
		}),
		mcpserver.Define(mcpserver.Spec{Name: "refuse", Title: "Refuse"},
			func(_ context.Context, account store.Account, in sayIn) (mcpserver.Reply[sayOut], error) {
				return mcpserver.Reply[sayOut]{
					Text:    "refused",
					Payload: sayOut{Status: "rejected", Said: in.Say, For: account.ID},
				}, nil
			}),
		mcpserver.Define(mcpserver.Spec{Name: "nest", Title: "Nest"},
			func(_ context.Context, account store.Account, in sayIn) (mcpserver.Reply[nestedOut], error) {
				return mcpserver.Reply[nestedOut]{
					Text:    "nested",
					Payload: nestedOut{Inner: sayOut{Status: "rejected", Said: in.Say, For: account.ID}},
				}, nil
			}),
		mcpserver.Define(mcpserver.Spec{Name: "fail", Title: "Fail"},
			func(_ context.Context, _ store.Account, in sayIn) (mcpserver.Reply[sayOut], error) {
				return mcpserver.Reply[sayOut]{}, fmt.Errorf("drive: the file of %s is gone", in.Say)
			}),
		mcpserver.Define(mcpserver.Spec{Name: "slow", Title: "Slow"},
			func(_ context.Context, _ store.Account, in sayIn) (mcpserver.Reply[sayOut], error) {
				return mcpserver.Reply[sayOut]{}, fmt.Errorf("drive: reading %s: %w", in.Say, context.DeadlineExceeded)
			}),
		mcpserver.Define(mcpserver.Spec{Name: "explode", Title: "Explode"},
			func(_ context.Context, _ store.Account, in sayIn) (mcpserver.Reply[sayOut], error) {
				panic("the tool knew " + in.Say)
			}),
		mcpserver.Define(mcpserver.Spec{Name: "break", Title: "Break"},
			func(context.Context, store.Account, sayIn) (mcpserver.Reply[brokenOut], error) {
				return mcpserver.Reply[brokenOut]{Text: "broken"}, nil
			}),
	}
}

// harness is the endpoint behind the real router on a real listener, with
// everything it records kept in memory.
type harness struct {
	server  *httptest.Server
	spans   *tracetest.SpanRecorder
	logs    *observer.ObservedLogs
	metrics *sdkmetric.ManualReader
	// traces, meters and log are what the endpoint records with, for the
	// tools of a case that record inside the calls they serve.
	traces trace.TracerProvider
	meters *sdkmetric.MeterProvider
	log    *zap.Logger
	// limits are the paces the endpoint holds every message to: none a case
	// comes near, unless the case is about them.
	limits mcpserver.Limits
}

// serve starts the endpoint with the sign-in a case chooses, serving the tools
// made for these cases.
func serve(t *testing.T, signIn mcpserver.SignIn) *harness {
	t.Helper()
	return serveTools(t, signIn, tools()...)
}

// serveTools starts the endpoint with the sign-in and the tools a case
// chooses.
func serveTools(t *testing.T, signIn mcpserver.SignIn, served ...mcpserver.Tool) *harness {
	t.Helper()

	h := newHarness(t)
	h.start(t, signIn, served...)
	return h
}

// newHarness is a harness that records and serves nothing yet, so that the
// tools of a case can be built to record into it. Every span is kept, through
// the same filter the service installs, every line is kept as the logger
// received it, and every measurement waits until a case reads it.
func newHarness(t *testing.T) *harness {
	t.Helper()

	h := &harness{spans: tracetest.NewSpanRecorder(), metrics: sdkmetric.NewManualReader(), limits: roomyLimits(t)}
	traces := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithSpanProcessor(telemetry.Redactor()),
		sdktrace.WithSpanProcessor(h.spans),
	)
	t.Cleanup(func() { _ = traces.Shutdown(context.Background()) })
	h.traces = traces
	h.meters = sdkmetric.NewMeterProvider(sdkmetric.WithReader(h.metrics))
	t.Cleanup(func() { _ = h.meters.Shutdown(context.Background()) })

	core, logs := observer.New(zapcore.DebugLevel)
	h.logs = logs
	h.log = zap.New(core)
	return h
}

// start serves the endpoint with the sign-in and the tools a case chooses.
func (h *harness) start(t *testing.T, signIn mcpserver.SignIn, served ...mcpserver.Tool) {
	t.Helper()

	// The address is known once the listener is, and the router is built for
	// the address it is served at.
	h.server = httptest.NewUnstartedServer(nil)
	endpoint, err := mcpserver.NewHandler(&mcpserver.Settings{
		Instructions:        instructions,
		InstructionsVersion: instructionsVersion,
		Version:             "test",
		SignIn:              signIn,
		Limits:              h.limits,
		Traces:              h.traces,
		Logger:              h.log,
		ProjectID:           projectID,
		Widget:              page,
	}, served...)
	if err != nil {
		t.Fatalf("NewHandler() error = %v, want nil", err)
	}
	// The sign-in has no part in these cases: its paths answer nothing.
	router, err := httpserver.NewRouter("http://"+h.server.Listener.Addr().String(), &httpserver.Endpoints{
		Health:           httpserver.NewHealthHandler(),
		MCP:              endpoint,
		ResourceMetadata: http.NotFoundHandler(),
		ServerMetadata:   http.NotFoundHandler(),
		Register:         http.NotFoundHandler(),
		Authorize:        http.NotFoundHandler(),
		Consent:          http.NotFoundHandler(),
		Callback:         http.NotFoundHandler(),
		Token:            http.NotFoundHandler(),
		Revoke:           http.NotFoundHandler(),
		Busy:             http.NotFoundHandler(),
	}, roomyAddresses(t), h.log, httpserver.Observability{
		Traces:    h.traces,
		Meters:    h.meters,
		Flush:     func(context.Context, bool) error { return nil },
		ProjectID: projectID,
	})
	if err != nil {
		t.Fatalf("NewRouter() error = %v, want nil", err)
	}
	h.server.Config.Handler = router
	h.server.Start()
	t.Cleanup(h.server.Close)
}

// connect is a client of the protocol, speaking the version a case chooses —
// the newest one when it chooses none — and named as a chat host would be.
func (h *harness) connect(t *testing.T, version string) *mcp.ClientSession {
	t.Helper()

	client := mcp.NewClient(&mcp.Implementation{Name: "claude-code", Version: "1.0.0"}, nil)
	session, err := client.Connect(t.Context(), &mcp.StreamableClientTransport{
		Endpoint:             h.server.URL + "/mcp",
		DisableStandaloneSSE: true,
		MaxRetries:           -1,
	}, &mcp.ClientSessionOptions{ProtocolVersion: version})
	if err != nil {
		t.Fatalf("Connect() error = %v, want nil", err)
	}
	t.Cleanup(func() { _ = session.Close() })
	return session
}

// settle waits for every request to be over. An answer can reach the client
// before the request that carried it is done: the span of the request closes
// and its line is written after the answer has left. What was recorded is
// read only once the server has finished with every request.
func (h *harness) settle() { h.server.Close() }

// toolLines are the lines tool calls left.
func (h *harness) toolLines() []observer.LoggedEntry {
	return h.logs.FilterMessage("tool_call").All()
}

// lineOf is the one line a call of this tool left.
func (h *harness) lineOf(t *testing.T, tool string) *observer.LoggedEntry {
	t.Helper()

	lines := h.toolLines()
	var found []*observer.LoggedEntry
	for i := range lines {
		if lines[i].ContextMap()["tool"] == tool {
			found = append(found, &lines[i])
		}
	}
	if len(found) != 1 {
		t.Fatalf("tool_call lines for %s = %d, want exactly 1", tool, len(found))
	}
	return found[0]
}

// spanNamed is the one span of that name.
func (h *harness) spanNamed(t *testing.T, name string) sdktrace.ReadOnlySpan {
	t.Helper()

	var found []sdktrace.ReadOnlySpan
	for _, span := range h.spans.Ended() {
		if span.Name() == name {
			found = append(found, span)
		}
	}
	if len(found) != 1 {
		t.Fatalf("spans named %q = %d, want exactly 1", name, len(found))
	}
	return found[0]
}

// received is an answer of the endpoint, read whole.
type received struct {
	status int
	header http.Header
	body   []byte
}

// post sends one request to the endpoint as a client of an older protocol
// would, with no version of its own, and reads the whole answer, so that a case
// can see it as the client did.
func (h *harness) post(t *testing.T, body string, header http.Header) received {
	t.Helper()

	req, err := http.NewRequestWithContext(t.Context(), http.MethodPost, h.server.URL+"/mcp", strings.NewReader(body))
	if err != nil {
		t.Fatalf("NewRequest() error = %v, want nil", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json, text/event-stream")
	for key, values := range header {
		req.Header[key] = values
	}
	resp, err := h.server.Client().Do(req)
	if err != nil {
		t.Fatalf("POST /mcp error = %v, want nil", err)
	}
	defer func() { _ = resp.Body.Close() }()
	read, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("reading the answer: %v", err)
	}
	return received{status: resp.StatusCode, header: resp.Header, body: read}
}

// message is the JSON-RPC message an answer carries, streamed or not.
func message(t *testing.T, got received) map[string]any {
	t.Helper()

	payload := got.body
	if strings.HasPrefix(got.header.Get("Content-Type"), "text/event-stream") {
		payload = nil
		for line := range strings.SplitSeq(string(got.body), "\n") {
			if data, isData := strings.CutPrefix(line, "data: "); isData {
				payload = []byte(data)
				break
			}
		}
	}
	var decoded map[string]any
	if err := json.Unmarshal(payload, &decoded); err != nil {
		t.Fatalf("the answer carries no JSON-RPC message: %v; body = %s", err, got.body)
	}
	return decoded
}

// call calls a tool and expects the protocol to answer.
func call(t *testing.T, session *mcp.ClientSession, tool string, args any) *mcp.CallToolResult {
	t.Helper()

	result, err := session.CallTool(t.Context(), &mcp.CallToolParams{Name: tool, Arguments: args})
	if err != nil {
		t.Fatalf("CallTool(%s) error = %v, want an answer", tool, err)
	}
	return result
}

// textOf is the words a result gives the model.
func textOf(t *testing.T, result *mcp.CallToolResult) string {
	t.Helper()

	if len(result.Content) != 1 {
		t.Fatalf("content blocks = %d, want 1", len(result.Content))
	}
	text, isText := result.Content[0].(*mcp.TextContent)
	if !isText {
		t.Fatalf("content = %T, want text", result.Content[0])
	}
	return text.Text
}

// field is one field of a line, as a string.
func field(t *testing.T, line *observer.LoggedEntry, key string) string {
	t.Helper()

	value, found := line.ContextMap()[key]
	if !found {
		t.Fatalf("the line has no %q field: %v", key, line.ContextMap())
	}
	text, isText := value.(string)
	if !isText {
		t.Fatalf("%q = %T, want a string", key, value)
	}
	return text
}

// legacyCall is a tool call as a client of an older protocol sends it.
func legacyCall(tool, args string) string {
	return `{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"` + tool + `","arguments":` + args + `}}`
}

// pacesOf are the paces of an account and of the instance, so many messages a
// minute each, on a clock that stands still.
func pacesOf(t testing.TB, perAccount, instance int) mcpserver.Limits {
	t.Helper()
	return mcpserver.Limits{PerAccount: ratelimittest.Keyed(t, perAccount), Instance: ratelimittest.Shared(t, instance)}
}

// roomyLimits are paces no case comes near.
func roomyLimits(t testing.TB) mcpserver.Limits {
	t.Helper()
	return mcpserver.Limits{PerAccount: ratelimittest.Roomy(t), Instance: ratelimittest.Roomy(t)}
}

// roomyAddresses are the router's paces, which no case here comes near: the
// sign-in has no part in these cases.
func roomyAddresses(t testing.TB) httpserver.Limits {
	t.Helper()
	return httpserver.Limits{PerAddress: ratelimittest.Roomy(t), Instance: ratelimittest.Roomy(t)}
}

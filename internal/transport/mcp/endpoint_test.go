package mcpserver_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"go.opentelemetry.io/otel/codes"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
	tracenoop "go.opentelemetry.io/otel/trace/noop"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"

	"github.com/MathTrail/mathtrail-standalone/internal/store"
	mcpserver "github.com/MathTrail/mathtrail-standalone/internal/transport/mcp"
)

// A tool is listed as it was defined, with what every tool of the service
// says about itself: it changes nothing beyond the parent's own file, and it
// needs the one scope a token grants.
func TestAToolIsListedAsItWasDefined(t *testing.T) {
	t.Parallel()

	h := serve(t, mcpserver.DevSignIn)
	listed, err := h.connect(t, "").ListTools(t.Context(), nil)
	if err != nil {
		t.Fatalf("ListTools() error = %v, want nil", err)
	}

	var echo *mcp.Tool
	for _, tool := range listed.Tools {
		if tool.Name == "echo" {
			echo = tool
		}
	}
	if echo == nil {
		t.Fatalf("the tools listed are %v, want echo among them", listed.Tools)
	}
	if echo.Title != "Echo" {
		t.Errorf("title = %q, want %q", echo.Title, "Echo")
	}
	hints := echo.Annotations
	if hints == nil || !hints.ReadOnlyHint || !hints.IdempotentHint {
		t.Fatalf("annotations = %+v, want read-only and idempotent", hints)
	}
	if hints.DestructiveHint == nil || *hints.DestructiveHint {
		t.Error("destructiveHint is not false, want it said rather than left to its default of true")
	}
	if hints.OpenWorldHint == nil || *hints.OpenWorldHint {
		t.Error("openWorldHint is not false, want it said rather than left to its default of true")
	}
	if echo.OutputSchema == nil {
		t.Error("the tool declares no output schema, want one a host can check a result against")
	}
	schemes, err := json.Marshal(echo.Meta["securitySchemes"])
	if err != nil {
		t.Fatalf("securitySchemes do not marshal: %v", err)
	}
	if got, want := string(schemes), `[{"scopes":["mcp"],"type":"oauth2"}]`; got != want {
		t.Errorf("_meta.securitySchemes = %s, want %s", got, want)
	}
}

// A call is answered with words for the model and a payload for a widget, for
// the account that signed in, and it leaves one line saying what it was.
func TestACallIsAnsweredAndLeavesOneLine(t *testing.T) {
	t.Parallel()

	h := serve(t, mcpserver.DevSignIn)
	result := call(t, h.connect(t, ""), "echo", map[string]any{"say": "hello"})

	if result.IsError {
		t.Fatalf("the result is an error: %s", textOf(t, result))
	}
	if got, want := textOf(t, result), "said: hello"; got != want {
		t.Errorf("text = %q, want %q", got, want)
	}
	payload, isObject := result.StructuredContent.(map[string]any)
	if !isObject || payload["said"] != "hello" || payload["for"] != mcpserver.DevAccount {
		t.Errorf("structuredContent = %v, want what was said, for %s", result.StructuredContent, mcpserver.DevAccount)
	}

	h.settle()
	line := h.lineOf(t, "echo")
	if line.Level != zapcore.InfoLevel {
		t.Errorf("level = %s, want info", line.Level)
	}
	for key, want := range map[string]string{
		"outcome":              "ok",
		"protocol_version":     "2026-07-28",
		"client":               "claude",
		"user":                 mcpserver.DevAccount,
		"instructions_version": instructionsVersion,
	} {
		if got := field(t, line, key); got != want {
			t.Errorf("%s = %q, want %q", key, got, want)
		}
	}
	if _, timed := line.ContextMap()["duration_ms"]; !timed {
		t.Error("the line has no duration_ms, want the call timed")
	}

	// The request's own line and the call's line are one request's lines.
	id := field(t, line, "request_id")
	var requestLines int
	for _, request := range h.logs.FilterMessage("http_request").All() {
		if request.ContextMap()["request_id"] == id {
			requestLines++
		}
	}
	if id == "" || requestLines != 1 {
		t.Errorf("request_id = %q names %d request lines, want exactly the call's own", id, requestLines)
	}
}

// The service speaks the newest version of the protocol, and it answers a
// client of an older one rather than refusing it: a chat host probes with an
// older handshake before it settles on the newest, and a refusal there breaks
// adding the service at all.
func TestTheNewestProtocolIsSpokenAndOlderOnesAreAnswered(t *testing.T) {
	t.Parallel()

	t.Run("a client of the newest", func(t *testing.T) {
		t.Parallel()

		h := serve(t, mcpserver.DevSignIn)
		call(t, h.connect(t, ""), "echo", map[string]any{"say": "hello"})
		h.settle()

		if got := field(t, h.lineOf(t, "echo"), "protocol_version"); got != "2026-07-28" {
			t.Errorf("protocol_version = %q, want 2026-07-28", got)
		}
	})

	t.Run("a client of an older one", func(t *testing.T) {
		t.Parallel()

		h := serve(t, mcpserver.DevSignIn)
		session := h.connect(t, "2025-11-25")
		if _, err := session.ListTools(t.Context(), nil); err != nil {
			t.Fatalf("ListTools() error = %v, want the older client answered", err)
		}
		call(t, session, "echo", map[string]any{"say": "hello"})
		h.settle()

		line := h.lineOf(t, "echo")
		if got := field(t, line, "protocol_version"); got != "2025-11-25" {
			t.Errorf("protocol_version = %q, want 2025-11-25", got)
		}
		// A client of an older protocol introduces itself once, in a
		// handshake no instance keeps, so its calls come from nobody known.
		if got := field(t, line, "client"); got != "unknown" {
			t.Errorf("client = %q, want unknown", got)
		}
	})

	t.Run("a handshake with no version", func(t *testing.T) {
		t.Parallel()

		h := serve(t, mcpserver.DevSignIn)
		resp := h.post(t, `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"",`+
			`"capabilities":{},"clientInfo":{"name":"probe","version":"0"}}}`, nil)
		if resp.status != http.StatusOK {
			t.Fatalf("status = %d, want %d", resp.status, http.StatusOK)
		}
		result, isObject := message(t, resp)["result"].(map[string]any)
		if !isObject || result["protocolVersion"] != "2025-11-25" {
			t.Errorf("result = %v, want a handshake on 2025-11-25", result)
		}
	})
}

// Without the development sign-in nobody is let in, and the refusal is the
// one a client expects of a service that wants a token: 401, and a challenge
// naming the scope. No tool runs.
func TestNobodyIsLetInWithoutTheDevelopmentSignIn(t *testing.T) {
	t.Parallel()

	h := serve(t, mcpserver.NobodySignsIn)
	for _, credential := range []string{"", "Bearer a-token-nothing-issued"} {
		header := http.Header{}
		if credential != "" {
			header.Set("Authorization", credential)
		}
		resp := h.post(t, legacyCall("echo", `{"say":"hello"}`), header)

		if resp.status != http.StatusUnauthorized {
			t.Errorf("with credential %q: status = %d, want %d", credential, resp.status, http.StatusUnauthorized)
		}
		if got, want := resp.header.Get("WWW-Authenticate"), `Bearer scope="mcp"`; got != want {
			t.Errorf("with credential %q: WWW-Authenticate = %q, want %q", credential, got, want)
		}
		if got, want := resp.header.Get("Cache-Control"), "no-store, no-transform"; got != want {
			t.Errorf("with credential %q: Cache-Control = %q, want %q", credential, got, want)
		}
	}

	client := mcp.NewClient(&mcp.Implementation{Name: "claude-code", Version: "1.0.0"}, nil)
	if _, err := client.Connect(t.Context(), &mcp.StreamableClientTransport{
		Endpoint: h.server.URL + "/mcp", DisableStandaloneSSE: true, MaxRetries: -1,
	}, nil); err == nil {
		t.Error("Connect() error = nil, want the client refused")
	}

	h.settle()
	if lines := h.toolLines(); len(lines) != 0 {
		t.Errorf("tool_call lines = %d, want none: no tool may run for nobody", len(lines))
	}
}

// A tool that could not do its work is told to the model in one sentence of
// ours, never in the words of what went wrong, which may name a child or a
// file. The line and the span say it failed and why, by a word from a list.
func TestAFailureIsOneSentenceOfOurs(t *testing.T) {
	t.Parallel()

	cases := []struct {
		tool     string
		sentence string
		kind     string
	}{
		{tool: "fail", sentence: "Something went wrong inside MathTrail.", kind: "internal"},
		{tool: "slow", sentence: "MathTrail took too long to answer.", kind: "timeout"},
		// A payload that does not fit the tool's own schema is our fault, and
		// the protocol's description of it is in the words of the payload.
		{tool: "break", sentence: "Something went wrong inside MathTrail.", kind: "internal"},
	}
	for _, tc := range cases {
		t.Run(tc.tool, func(t *testing.T) {
			t.Parallel()

			h := serve(t, mcpserver.DevSignIn)
			result := call(t, h.connect(t, ""), tc.tool, map[string]any{"say": said})
			wantOurSentence(t, result, tc.sentence)

			h.settle()
			wantFailed(t, h, tc.tool, tc.kind)
		})
	}
}

// wantOurSentence holds a result to what a failure tells the model: marked as
// an error, in the sentence written for it, and in none of the words of what
// went wrong.
func wantOurSentence(t *testing.T, result *mcp.CallToolResult, sentence string) {
	t.Helper()

	text := textOf(t, result)
	if !result.IsError {
		t.Errorf("isError = false, want the failure marked for the model")
	}
	if !strings.HasPrefix(text, sentence) {
		t.Errorf("text = %q, want it to begin %q", text, sentence)
	}
	for _, secret := range []string{pseudonym, "drive", "validating"} {
		if strings.Contains(text, secret) {
			t.Errorf("text = %q, want it to carry no %q", text, secret)
		}
	}
}

// wantFailed holds a call's line and span to what a failure leaves: an error
// line naming the kind, and a span marked failed with nothing of the error
// recorded on it.
func wantFailed(t *testing.T, h *harness, tool, kind string) {
	t.Helper()

	line := h.lineOf(t, tool)
	if line.Level != zapcore.ErrorLevel {
		t.Errorf("level = %s, want error", line.Level)
	}
	if got := field(t, line, "outcome"); got != "failed" {
		t.Errorf("outcome = %q, want failed", got)
	}
	if got := field(t, line, "error"); got != kind {
		t.Errorf("error = %q, want %q", got, kind)
	}
	span := h.spanNamed(t, "tools/call "+tool)
	if span.Status().Code != codes.Error {
		t.Errorf("span status = %v, want error", span.Status().Code)
	}
	if len(span.Events()) != 0 {
		t.Errorf("span events = %v, want none: an error's text is never recorded", span.Events())
	}
}

// A tool that panics is a failure like any other: the model is told in the
// general sentence, the line says what it panicked with by its kind alone and
// where, and the endpoint goes on answering the next call.
func TestAPanicIsAFailureAndTheEndpointGoesOn(t *testing.T) {
	t.Parallel()

	h := serve(t, mcpserver.DevSignIn)
	session := h.connect(t, "")

	result := call(t, session, "explode", map[string]any{"say": said})
	if !result.IsError || !strings.HasPrefix(textOf(t, result), "Something went wrong inside MathTrail.") {
		t.Errorf("result = %q (isError %t), want the general failure", textOf(t, result), result.IsError)
	}
	if next := call(t, session, "echo", map[string]any{"say": "hello"}); next.IsError {
		t.Errorf("the call after the panic failed: %s", textOf(t, next))
	}

	h.settle()
	line := h.lineOf(t, "explode")
	if got := field(t, line, "error"); got != "panic" {
		t.Errorf("error = %q, want panic", got)
	}
	if got := field(t, line, "panic"); got != "a value of type string" {
		t.Errorf("panic = %q, want it named by its kind alone", got)
	}
	if stack := field(t, line, "stack"); !strings.Contains(strings.SplitN(stack, "\n", 2)[0], "/mcp_test.tools") {
		t.Errorf("stack begins %q, want it to begin at the code that panicked", strings.SplitN(stack, "\n", 2)[0])
	}
}

// A refusal is an answer the model is meant to act on, not a failure: it is
// not marked as an error, and its line says which refusal it was. Only the
// status at the top of the payload makes one.
func TestARefusalIsAnAnswer(t *testing.T) {
	t.Parallel()

	h := serve(t, mcpserver.DevSignIn)
	session := h.connect(t, "")
	if result := call(t, session, "refuse", map[string]any{"say": "hello"}); result.IsError {
		t.Errorf("isError = true, want a refusal answered as a result")
	}
	call(t, session, "nest", map[string]any{"say": "hello"})

	h.settle()
	refused := h.lineOf(t, "refuse")
	if got := field(t, refused, "outcome"); got != "refused" {
		t.Errorf("outcome = %q, want refused", got)
	}
	if got := field(t, refused, "status"); got != "rejected" {
		t.Errorf("status = %q, want rejected", got)
	}
	if span := h.spanNamed(t, "tools/call refuse"); span.Status().Code == codes.Error {
		t.Error("the refusal's span is marked failed, want a refusal left unmarked")
	}
	if got := field(t, h.lineOf(t, "nest"), "outcome"); got != "ok" {
		t.Errorf("a status deeper in the payload: outcome = %q, want ok", got)
	}
}

// A call the protocol refuses before any tool runs is the caller's mistake:
// arguments that do not fit the tool, or a tool nobody defined. It is still a
// call, so it leaves its line, and the name a caller made up is not written.
func TestACallTheProtocolRefusesIsInvalid(t *testing.T) {
	t.Parallel()

	h := serve(t, mcpserver.DevSignIn)
	session := h.connect(t, "")

	if result := call(t, session, "echo", map[string]any{"say": 42}); !result.IsError {
		t.Error("arguments that do not fit were accepted, want them refused")
	}
	if _, err := session.CallTool(t.Context(), &mcp.CallToolParams{Name: pseudonym}); err == nil {
		t.Error("a tool nobody defined was called, want the call refused")
	}

	h.settle()
	arguments := h.lineOf(t, "echo")
	if arguments.Level != zapcore.WarnLevel {
		t.Errorf("level = %s, want warn", arguments.Level)
	}
	if got := field(t, arguments, "outcome"); got != "invalid" {
		t.Errorf("outcome = %q, want invalid", got)
	}
	if got := field(t, arguments, "error"); got != "arguments" {
		t.Errorf("error = %q, want arguments", got)
	}
	nobody := h.lineOf(t, "other")
	if got := field(t, nobody, "error"); got != "protocol" {
		t.Errorf("error = %q, want protocol", got)
	}
}

// A tool call is one tree in the trace: the request, the call inside it, and
// what the tool did inside the call. The call's line names the same trace and
// the call's own span, so that a reader can go from one to the other.
func TestAToolCallIsOneTree(t *testing.T) {
	t.Parallel()

	h := serve(t, mcpserver.DevSignIn)
	call(t, h.connect(t, ""), "echo", map[string]any{"say": "hello"})
	h.settle()

	tool := h.spanNamed(t, "tools/call echo")
	work := h.spanNamed(t, "work")
	var request sdktrace.ReadOnlySpan
	for _, span := range h.spans.Ended() {
		if span.SpanContext().SpanID() == tool.Parent().SpanID() {
			request = span
		}
	}
	if request == nil {
		t.Fatal("the call's span has no parent among the recorded spans, want the request's")
	}
	if request.Name() != "POST /mcp" {
		t.Errorf("the call's parent is %q, want the request, POST /mcp", request.Name())
	}
	if work.Parent().SpanID() != tool.SpanContext().SpanID() {
		t.Error("what the tool did is not inside the call's span")
	}
	trees := map[trace.TraceID]bool{}
	for _, span := range []sdktrace.ReadOnlySpan{request, tool, work} {
		trees[span.SpanContext().TraceID()] = true
	}
	if len(trees) != 1 {
		t.Errorf("the three spans are in %d traces, want one", len(trees))
	}
	if tool.SpanKind() != trace.SpanKindServer {
		t.Errorf("the call's span kind = %v, want server", tool.SpanKind())
	}
	for key, want := range map[string]string{
		"mcp.method.name":                "tools/call",
		"gen_ai.operation.name":          "execute_tool",
		"gen_ai.tool.name":               "echo",
		"mcp.protocol.version":           "2026-07-28",
		"mathtrail.instructions_version": instructionsVersion,
		"mathtrail.tool.outcome":         "ok",
	} {
		if got := attributeOf(tool, key); got != want {
			t.Errorf("attribute %s = %q, want %q", key, got, want)
		}
	}

	line := h.lineOf(t, "echo")
	if got, want := field(t, line, "logging.googleapis.com/trace"),
		"projects/"+projectID+"/traces/"+tool.SpanContext().TraceID().String(); got != want {
		t.Errorf("trace = %q, want %q", got, want)
	}
	if got, want := field(t, line, "logging.googleapis.com/spanId"), tool.SpanContext().SpanID().String(); got != want {
		t.Errorf("spanId = %q, want the call's own span %q", got, want)
	}
}

// The guard. Nothing a child could be recognised by and nothing of the answer
// reaches a span or a line, whatever way the call ended: in the arguments, in
// the payload, in an error's text or in what a tool panicked with.
func TestNothingAboutTheChildReachesASpanOrALine(t *testing.T) {
	t.Parallel()

	h := serve(t, mcpserver.DevSignIn)
	session := h.connect(t, "")
	for _, tool := range []string{"echo", "refuse", "nest", "fail", "slow", "explode", "break"} {
		call(t, session, tool, map[string]any{"say": said})
	}
	call(t, session, "echo", map[string]any{"say": said, "and": said})
	h.settle()

	secrets := []string{pseudonym, answer, "drive"}
	for _, span := range h.spans.Ended() {
		wantNoneOf(t, "span "+span.Name(), spanTexts(span), secrets)
	}
	lines := h.logs.All()
	for i := range lines {
		wantNoneOf(t, "line "+lines[i].Message, lineTexts(t, &lines[i]), secrets)
	}
}

// spanTexts is every text a span carries: its name, its status, its
// attributes and its events.
func spanTexts(span sdktrace.ReadOnlySpan) []string {
	texts := []string{span.Name(), span.Status().Description}
	for _, attr := range span.Attributes() {
		texts = append(texts, attr.Value.String())
	}
	for _, event := range span.Events() {
		texts = append(texts, event.Name)
		for _, attr := range event.Attributes {
			texts = append(texts, attr.Value.String())
		}
	}
	return texts
}

// lineTexts is every text a line carries: its message and its fields, as the
// encoder would write them.
func lineTexts(t *testing.T, line *observer.LoggedEntry) []string {
	t.Helper()

	encoded, err := json.Marshal(line.ContextMap())
	if err != nil {
		t.Fatalf("a line does not encode: %v", err)
	}
	return []string{line.Message, string(encoded)}
}

// wantNoneOf holds texts to carrying none of the secrets.
func wantNoneOf(t *testing.T, where string, texts, secrets []string) {
	t.Helper()

	for _, text := range texts {
		for _, secret := range secrets {
			if strings.Contains(text, secret) {
				t.Errorf("%s carries %q in %q", where, secret, text)
			}
		}
	}
}

// A request larger than any honest one is refused before anything reads it,
// and no tool runs.
func TestARequestOverTheLimitIsRefused(t *testing.T) {
	t.Parallel()

	h := serve(t, mcpserver.DevSignIn)
	resp := h.post(t, legacyCall("echo", `{"say":"`+strings.Repeat("a", 1<<20)+`"}`), nil)

	if resp.status != http.StatusRequestEntityTooLarge {
		t.Errorf("status = %d, want %d", resp.status, http.StatusRequestEntityTooLarge)
	}
	if got, want := resp.header.Get("Cache-Control"), "no-store, no-transform"; got != want {
		t.Errorf("Cache-Control = %q, want %q", got, want)
	}
	h.settle()
	if lines := h.toolLines(); len(lines) != 0 {
		t.Errorf("tool_call lines = %d, want none", len(lines))
	}
}

// An answer of the endpoint carries a child's profile, tasks and progress, and
// nothing on its way may keep a copy — although the protocol's library would
// allow one.
func TestAnAnswerIsNeverStored(t *testing.T) {
	t.Parallel()

	h := serve(t, mcpserver.DevSignIn)
	resp := h.post(t, legacyCall("echo", `{"say":"hello"}`), nil)

	if resp.status != http.StatusOK {
		t.Fatalf("status = %d, want %d", resp.status, http.StatusOK)
	}
	if got, want := resp.header.Get("Cache-Control"), "no-store, no-transform"; got != want {
		t.Errorf("Cache-Control = %q, want %q", got, want)
	}
	if _, answered := message(t, resp)["result"]; !answered {
		t.Error("the answer carries no result, want the call answered")
	}
}

// An endpoint that could not be served as it should — with nobody deciding
// who may sign in, or two tools under one name — is not built.
func TestAnEndpointThatCouldNotBeServedIsNotBuilt(t *testing.T) {
	t.Parallel()

	settings := func() *mcpserver.Settings {
		return &mcpserver.Settings{
			Instructions:        instructions,
			InstructionsVersion: instructionsVersion,
			SignIn:              mcpserver.DevSignIn,
			Traces:              tracenoop.NewTracerProvider(),
			Logger:              zap.NewNop(),
		}
	}
	withoutSignIn := settings()
	withoutSignIn.SignIn = nil
	withoutTraces := settings()
	withoutTraces.Traces = nil
	withoutLogger := settings()
	withoutLogger.Logger = nil
	withoutInstructions := settings()
	withoutInstructions.Instructions = ""
	withoutVersion := settings()
	withoutVersion.InstructionsVersion = ""
	echo := tools()[0]
	loose := mcpserver.Define(mcpserver.Spec{Name: "loose", Title: "Loose"},
		func(context.Context, store.Account, string) (mcpserver.Reply[sayOut], error) {
			return mcpserver.Reply[sayOut]{}, nil
		})
	misnamed := mcpserver.Define(mcpserver.Spec{Name: "next task", Title: "Next task"},
		func(context.Context, store.Account, sayIn) (mcpserver.Reply[sayOut], error) {
			return mcpserver.Reply[sayOut]{}, nil
		})

	cases := []struct {
		name     string
		settings *mcpserver.Settings
		tools    []mcpserver.Tool
		want     string
	}{
		{name: "no settings", settings: nil, want: "settings"},
		{name: "no sign-in", settings: withoutSignIn, want: "SignIn"},
		{name: "no tracer", settings: withoutTraces, want: "Traces"},
		{name: "no logger", settings: withoutLogger, want: "Logger"},
		{name: "no instructions", settings: withoutInstructions, want: "Instructions"},
		{name: "no version of the instructions", settings: withoutVersion, want: "InstructionsVersion"},
		{name: "two tools of one name", settings: settings(), tools: []mcpserver.Tool{echo, echo}, want: `"echo"`},
		{name: "a name no tool can have", settings: settings(), tools: []mcpserver.Tool{misnamed}, want: `"next task"`},
		{name: "arguments that are not an object", settings: settings(), tools: []mcpserver.Tool{loose}, want: `"loose"`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			_, err := mcpserver.NewHandler(tc.settings, tc.tools...)
			if !errors.Is(err, mcpserver.ErrSettings) {
				t.Fatalf("NewHandler() error = %v, want it refused", err)
			}
			if !strings.Contains(err.Error(), tc.want) {
				t.Errorf("error = %q, want it to name %s", err, tc.want)
			}
		})
	}
}

// attributeOf is the value of one attribute of a span, as text.
func attributeOf(span sdktrace.ReadOnlySpan, key string) string {
	for _, attr := range span.Attributes() {
		if string(attr.Key) == key {
			return attr.Value.String()
		}
	}
	return ""
}

// A sign-in that lets a request through without naming anybody leaves the tool
// nobody to act for, and the call is refused as such rather than run for
// nobody.
func TestACallThatReachesAToolWithNobodySignedInIsRefused(t *testing.T) {
	t.Parallel()

	h := serve(t, func(next http.Handler) http.Handler { return next })
	result := call(t, h.connect(t, ""), "echo", map[string]any{"say": "hello"})
	wantOurSentence(t, result, "Nobody is signed in to MathTrail.")

	h.settle()
	wantFailed(t, h, "echo", "not_signed_in")
	if _, named := h.lineOf(t, "echo").ContextMap()["user"]; named {
		t.Error("the line names a user, want none for a call nobody signed in to")
	}
}

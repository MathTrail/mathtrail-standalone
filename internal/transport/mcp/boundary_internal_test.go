package mcpserver

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/jsonrpc"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"go.opentelemetry.io/otel/trace"
	tracenoop "go.opentelemetry.io/otel/trace/noop"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"

	"github.com/MathTrail/mathtrail-standalone/internal/store"
)

// observedBoundary is a boundary that writes its lines where a case can read
// them, and records no spans.
func observedBoundary(traces trace.TracerProvider) (*boundary, *observer.ObservedLogs) {
	core, logs := observer.New(zapcore.DebugLevel)
	return newBoundary(&Settings{
		Traces: traces,
		Logger: zap.New(core),
	}, map[string]struct{}{}), logs
}

// A request that is not a tool call and panics is answered as an internal
// error, and its line says which method it was, for whom, and what it panicked
// with by its kind alone.
func TestAPanicOfAnotherMethodIsAnInternalError(t *testing.T) {
	t.Parallel()

	b, logs := observedBoundary(tracenoop.NewTracerProvider())
	handler := b.wrap(func(context.Context, string, mcp.Request) (mcp.Result, error) {
		panic("the resource of masha-ivanova")
	})

	ctx := withAccount(t.Context(), store.NewAccount(DevAccount, ""))
	result, err := handler(ctx, "resources/read", &mcp.ReadResourceRequest{})

	wantInternalError(t, result, err)
	fields := onlyPanicLine(t, logs)
	if fields["method"] != "resources/read" || fields["panic"] != "a value of type string" || fields["user"] != DevAccount {
		t.Errorf("line = %v, want the method, the panic's kind and the user", fields)
	}
}

// The boundary's own work around a tool call — opening its span, reading who
// called — runs on the library's goroutine too, and a panic there is answered
// like any other rather than ending the process.
func TestAPanicAroundAToolCallIsAnInternalError(t *testing.T) {
	t.Parallel()

	b, logs := observedBoundary(burningTracers{})
	handler := b.wrap(func(context.Context, string, mcp.Request) (mcp.Result, error) {
		t.Error("the tool ran, want the call stopped where the boundary failed")
		return nil, nil
	})

	result, err := handler(t.Context(), "tools/call", &mcp.CallToolRequest{Params: &mcp.CallToolParamsRaw{Name: "echo"}})

	wantInternalError(t, result, err)
	if fields := onlyPanicLine(t, logs); fields["method"] != "tools/call" {
		t.Errorf("line = %v, want it to name the method", fields)
	}
}

// burningTracers hand out a tracer that panics as a span is opened.
type burningTracers struct{ tracenoop.TracerProvider }

func (burningTracers) Tracer(string, ...trace.TracerOption) trace.Tracer { return burningTracer{} }

type burningTracer struct{ tracenoop.Tracer }

func (burningTracer) Start(context.Context, string, ...trace.SpanStartOption) (context.Context, trace.Span) {
	panic("the tracer is on fire")
}

// onlyPanicLine is the one line a panic outside a tool left, as its fields.
func onlyPanicLine(t *testing.T, logs *observer.ObservedLogs) map[string]any {
	t.Helper()

	lines := logs.FilterMessage("mcp_panic").All()
	if len(lines) != 1 {
		t.Fatalf("mcp_panic lines = %d, want 1", len(lines))
	}
	return lines[0].ContextMap()
}

// What the library answers a request that is not a tool call is passed on as
// it is, errors included: they describe the request in the protocol's terms,
// and the client's own mistake is not an internal error of ours.
func TestAnotherMethodsAnswerIsPassedOnAsItIs(t *testing.T) {
	t.Parallel()

	protocol := &jsonrpc.Error{Code: jsonrpc.CodeInvalidParams, Message: "unknown resource"}
	for _, sent := range []error{
		protocol,
		fmt.Errorf("reading: %w", protocol),
		errors.New("failed to decode cursor: illegal base64 data at input byte 4"),
	} {
		b, logs := observedBoundary(tracenoop.NewTracerProvider())
		handler := b.wrap(func(context.Context, string, mcp.Request) (mcp.Result, error) {
			return nil, sent
		})

		if _, err := handler(t.Context(), "tools/list", &mcp.ListToolsRequest{}); !errors.Is(err, sent) {
			t.Errorf("error = %v, want %v passed on", err, sent)
		}
		if logs.Len() != 0 {
			t.Errorf("lines = %v, want none for an answer passed on", logs.All())
		}
	}
}

// wantInternalError holds an answer to the one a client gets when something of
// ours failed under a request that is not a tool call.
func wantInternalError(t *testing.T, result mcp.Result, err error) {
	t.Helper()

	var wire *jsonrpc.Error
	if result != nil || !errors.As(err, &wire) {
		t.Fatalf("answer = %v, %v, want an internal error of the protocol", result, err)
	}
	if wire.Code != jsonrpc.CodeInternalError || wire.Message != "internal error" {
		t.Errorf("error = %d %q, want %d %q", wire.Code, wire.Message, jsonrpc.CodeInternalError, "internal error")
	}
}

// How a call ended is read from what is about to be sent, and nothing else.
func TestHowACallEndedIsReadFromWhatIsSent(t *testing.T) {
	t.Parallel()

	answered := func(payload string) *mcp.CallToolResult {
		return &mcp.CallToolResult{StructuredContent: json.RawMessage(payload)}
	}
	var argumentsRefused mcp.CallToolResult
	argumentsRefused.SetError(errors.New("validating \"arguments\": say: want string"))

	cases := []struct {
		name   string
		result mcp.Result
		err    error
		want   verdict
	}{
		{name: "the protocol refused it", err: &jsonrpc.Error{Code: jsonrpc.CodeInvalidParams},
			want: verdict{outcome: outcomeInvalid, kind: kindProtocol}},
		{name: "nothing was sent", want: verdict{outcome: outcomeOK}},
		{name: "not a call's result", result: &mcp.ListToolsResult{}, want: verdict{outcome: outcomeOK}},
		{name: "a failure of ours", result: failed(explain(context.DeadlineExceeded)),
			want: verdict{outcome: outcomeFailed, kind: kindTimeout}},
		{name: "arguments the protocol refused", result: &argumentsRefused,
			want: verdict{outcome: outcomeInvalid, kind: kindArguments}},
		{name: "the protocol failed on this side", err: &jsonrpc.Error{Code: jsonrpc.CodeInternalError},
			want: verdict{outcome: outcomeFailed, kind: kindInternal}},
		{name: "a refusal", result: answered(`{"status":"limited","retry":"tomorrow"}`),
			want: verdict{outcome: outcomeRefused, status: "limited"}},
		{name: "a status that is no refusal", result: answered(`{"status":"accepted"}`),
			want: verdict{outcome: outcomeOK}},
		{name: "a payload that is not an object", result: answered(`[{"status":"rejected"}]`),
			want: verdict{outcome: outcomeOK}},
		{name: "a payload not yet written out", result: &mcp.CallToolResult{
			StructuredContent: map[string]any{"status": "rejected"},
		}, want: verdict{outcome: outcomeOK}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if got := judge(tc.result, tc.err); got != tc.want {
				t.Errorf("judge() = %+v, want %+v", got, tc.want)
			}
		})
	}
}

// A call that names no tool at all is written as a tool nobody defined.
func TestACallWithoutParametersNamesNoTool(t *testing.T) {
	t.Parallel()

	b, _ := observedBoundary(tracenoop.NewTracerProvider())
	if got := b.toolLabel(toolName(&mcp.CallToolRequest{})); got != other {
		t.Errorf("tool = %q, want %q", got, other)
	}
}

// The rule against storing is in place however an answer starts to leave —
// by its status, by its first bytes, or by a flush of nothing — and the writer
// underneath stays reachable to whatever looks past this one.
func TestAnAnswerForbidsStoringHoweverItStarts(t *testing.T) {
	t.Parallel()

	starts := map[string]func(w *noStore){
		"a status": func(w *noStore) { w.WriteHeader(http.StatusOK) },
		"a body":   func(w *noStore) { _, _ = w.Write([]byte("{}")) },
		"a flush":  func(w *noStore) { w.Flush() },
	}
	for name, start := range starts {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			rec := httptest.NewRecorder()
			rec.Header().Set("Cache-Control", "no-cache, no-transform")
			writer := &noStore{ResponseWriter: rec}
			start(writer)

			if got := rec.Result().Header.Get("Cache-Control"); !strings.HasPrefix(got, "no-store") {
				t.Errorf("Cache-Control = %q, want no-store", got)
			}
			if writer.Unwrap() != rec {
				t.Error("Unwrap() is not the writer underneath")
			}
		})
	}
}

// A protocol error a tool call ends with is sent in its own words: the library
// would send the text of whatever wrapped it, and that text is not ours.
func TestAProtocolErrorIsSentInItsOwnWords(t *testing.T) {
	t.Parallel()

	wire := &jsonrpc.Error{Code: jsonrpc.CodeInvalidParams, Message: "unknown tool"}
	result, err := settle(nil, fmt.Errorf("drive: the profile of masha-ivanova: %w", wire))

	if result != nil || !errors.Is(err, wire) || err.Error() != wire.Error() {
		t.Errorf("settle() = %v, %v, want the protocol's own error and nothing else", result, err)
	}
}

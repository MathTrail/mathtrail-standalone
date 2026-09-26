package mcpserver_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	tracenoop "go.opentelemetry.io/otel/trace/noop"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"

	mcpserver "github.com/MathTrail/mathtrail-standalone/internal/transport/mcp"
)

// The arguments of a call are written by the chat's model. Nothing it writes
// may take the endpoint down or become a failure of ours: whatever arrives is
// answered, as a call the tool did or as arguments that do not fit it.
func FuzzToolArguments(f *testing.F) {
	for _, seed := range []string{
		`{"say":"hello"}`,
		`{"say":42}`,
		`{"say":"hello","and":{"deeper":[1,2,3]}}`,
		`{}`,
		`[]`,
		`null`,
		`"hello"`,
		`not json at all`,
	} {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, arguments string) {
		// Arguments that are not JSON would be refused with the whole request,
		// before any tool is looked for; as a string they reach the tool's
		// schema, which is what is under test.
		if !json.Valid([]byte(arguments)) {
			quoted, err := json.Marshal(arguments)
			if err != nil {
				t.Fatalf("quoting the arguments: %v", err)
			}
			arguments = string(quoted)
		}

		core, logs := observer.New(zapcore.DebugLevel)
		endpoint, err := mcpserver.NewHandler(&mcpserver.Settings{
			Instructions:        instructions,
			InstructionsVersion: instructionsVersion,
			SignIn:              mcpserver.DevSignIn,
			Traces:              tracenoop.NewTracerProvider(),
			Logger:              zap.New(core),
		}, tools()[0])
		if err != nil {
			t.Fatalf("NewHandler() error = %v, want nil", err)
		}

		req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/mcp",
			strings.NewReader(legacyCall("echo", arguments)))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json, text/event-stream")
		rec := httptest.NewRecorder()
		endpoint.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d for arguments %q, want the call answered", rec.Code, arguments)
		}
		lines := logs.FilterMessage("tool_call").All()
		if len(lines) != 1 {
			t.Fatalf("tool_call lines = %d for arguments %q, want 1", len(lines), arguments)
		}
		if outcome := lines[0].ContextMap()["outcome"]; outcome != "ok" && outcome != "invalid" {
			t.Errorf("outcome = %v for arguments %q, want ok or invalid", outcome, arguments)
		}
	})
}

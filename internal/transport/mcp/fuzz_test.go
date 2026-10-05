package mcpserver_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"testing"

	tracenoop "go.opentelemetry.io/otel/trace/noop"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/profile"
	"github.com/MathTrail/mathtrail-standalone/internal/domain/rating"
	"github.com/MathTrail/mathtrail-standalone/internal/domain/tutor"
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
			Limits:              roomyLimits(t),
			Traces:              tracenoop.NewTracerProvider(),
			Logger:              zap.New(core),
			Widget:              page,
			Origin:              serviceOrigin,
			Site:                siteOrigin,
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

// fuzzRequest is the request every case of the fuzzing below hands its task in
// against, so that a task written well gets all the way through.
const fuzzRequest = "req_fuzz"

// The arguments of the tools of a task are written by the chat's model as much
// as any, and the whole way runs behind them: the rule, the package, the
// checks, the sandbox, the seal and the store. Whatever arrives is answered —
// a task asked for or handed in, a refusal, or arguments that do not fit — and
// never as a failure of ours.
func FuzzTaskArguments(f *testing.F) {
	request := openRace(f)
	race, err := json.Marshal(raceOn(request))
	if err != nil {
		f.Fatalf("the race does not encode: %v", err)
	}
	raceQuoted, err := json.Marshal(partsAsStrings(f, raceOn(request)))
	if err != nil {
		f.Fatalf("the race with its parts as strings does not encode: %v", err)
	}
	for _, seed := range []struct {
		tool      uint8
		arguments string
	}{
		{0, `{"language":"en"}`},
		{0, `{"language":"en","topic":"logic.ordering","grade_level":"1-2","difficulty":2,"reason":"a race"}`},
		{0, `{"language":"","topic":"nowhere","difficulty":-7}`},
		{0, `{"language":"zh-Hant-TW","grade_level":"5-6","reason":"` + strings.Repeat("\u202e", 400) + `"}`},
		{1, string(race)},
		{1, string(raceQuoted)},
		{1, strings.Replace(string(race), `"correct_answer":"C"`, `"correct_answer":"A"`, 1)},
		{1, `{"request_id":"req_fuzz","brief":"a string","task":[1,2],"solver":"","self_check":null}`},
		{1, `{"request_id":"req_fuzz","brief":{},"task":{"options":{"A":1}},"solver":"def solve(options): return","self_check":{}}`},
		{1, `{}`},
		{2, `{"request_id":"` + request.ID + `"}`},
		{2, `{"request_id":"req_\u0000","extra":true}`},
		{3, `{"request_id":"` + request.ID + `"}`},
		{3, `{"request_id":["req_fuzz"]}`},
	} {
		f.Add(seed.tool, seed.arguments)
	}

	f.Fuzz(func(t *testing.T, pick uint8, arguments string) {
		if !json.Valid([]byte(arguments)) {
			quoted, err := json.Marshal(arguments)
			if err != nil {
				t.Fatalf("quoting the arguments: %v", err)
			}
			arguments = string(quoted)
		}
		tools := []string{"next_task", "submit_task", "get_package", "read_task"}
		tool := tools[int(pick)%len(tools)]

		if line := callOn(t, fuzzProfile(t), tool, arguments); line["outcome"] == "failed" {
			t.Errorf("%s(%q) failed as ours: %v", tool, arguments, line)
		}
	})
}

// callOn makes one call of a tool of a task, for the child of this profile,
// through an endpoint of its own over a store of its own, and is the line the
// call left.
func callOn(t *testing.T, p *profile.Profile, tool, arguments string) map[string]any {
	t.Helper()

	core, logs := observer.New(zapcore.DebugLevel)
	parts := allParts(t)
	parts.Store, parts.Logger = keptAsIs(t, p), zap.New(core)
	service, err := mcpserver.NewService(parts)
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}
	endpoint, err := mcpserver.NewHandler(&mcpserver.Settings{
		Instructions:        instructions,
		InstructionsVersion: instructionsVersion,
		SignIn:              mcpserver.DevSignIn,
		Limits:              roomyLimits(t),
		Traces:              tracenoop.NewTracerProvider(),
		Logger:              zap.New(core),
		Widget:              page,
		Origin:              serviceOrigin,
		Site:                siteOrigin,
	}, slices.Concat(service.ProfileTools(), service.TaskTools())...)
	if err != nil {
		t.Fatalf("NewHandler() error = %v, want nil", err)
	}

	req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/mcp",
		strings.NewReader(legacyCall(tool, arguments)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json, text/event-stream")
	rec := httptest.NewRecorder()
	endpoint.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d for %s(%q), want the call answered", rec.Code, tool, arguments)
	}
	lines := logs.FilterMessage("tool_call").All()
	if len(lines) != 1 {
		t.Fatalf("tool_call lines = %d for %s(%q), want 1", len(lines), tool, arguments)
	}
	return lines[0].ContextMap()
}

// fuzzProfile is a child of grade 2 with the race asked for, under the request
// the fuzzing hands its tasks in against.
func fuzzProfile(t testing.TB) *profile.Profile {
	t.Helper()

	p := profile.New(profile.Student{Grade: 2, Pseudonym: "Otter", Interests: []string{"sport"}}, "test", lessonDay)
	loaded, err := shipped()
	if err != nil {
		t.Fatalf("load the content: %v", err)
	}
	brief, mode, err := tutor.Next(p, loaded, tutor.Choice{
		Topic: "logic.ordering", GradeLevel: rating.Grades12, Difficulty: 2, Reason: "The child asked for a race.",
	})
	if err != nil {
		t.Fatalf("Next() error = %v", err)
	}
	p.Ask(&brief, mode, "en", lessonDay).ID = fuzzRequest
	return p
}

// openRace is the request the fuzzing hands its tasks in against.
func openRace(t testing.TB) *profile.OpenRequest {
	t.Helper()
	return fuzzProfile(t).OpenRequest
}

// fuzzTask stands in the arguments of the fuzzing below for the id of the task
// on the card, which is new every time a task is handed out.
const fuzzTask = "tsk_fuzz"

// The arguments of an answer are written by the chat's model, and by a card a
// host stands between: whatever arrives is answered — recorded, told again,
// refused as a task not on the card or as no answer at all — and never as a
// failure of ours.
func FuzzAnswerArguments(f *testing.F) {
	for _, seed := range []string{
		`{"task_id":"` + fuzzTask + `","answer":"C"}`,
		`{"task_id":"` + fuzzTask + `","answer":" a ","hint_used":true}`,
		`{"task_id":"` + fuzzTask + `","answer":"?"}`,
		`{"task_id":"` + fuzzTask + `","answer":"` + "\u202e" + `C"}`,
		`{"task_id":"` + fuzzTask + `","answer":"CC","hint_used":"yes"}`,
		`{"task_id":"tsk_another","answer":"B"}`,
		`{"task_id":"","answer":""}`,
		`{"answer":"C"}`,
		`{"task_id":7,"answer":["C"]}`,
		`{}`,
	} {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, arguments string) {
		if !json.Valid([]byte(arguments)) {
			quoted, err := json.Marshal(arguments)
			if err != nil {
				t.Fatalf("quoting the arguments: %v", err)
			}
			arguments = string(quoted)
		}
		p := raceOnTheCard(t, rating.TrialAnswers)
		arguments = strings.ReplaceAll(arguments, fuzzTask, p.CurrentTask.ID)

		if line := callOn(t, p, "submit_answer", arguments); line["outcome"] == "failed" {
			t.Errorf("submit_answer(%q) failed as ours: %v", arguments, line)
		}
	})
}

// The form on a card sends what the adult typed, with a host between: whatever
// arrives is answered — the change written, refused field by field, or
// arguments that do not fit — and never as a failure of ours, which is how a
// change let through that the file could not be written as would show.
func FuzzEditArguments(f *testing.F) {
	for _, seed := range []string{
		`{"pseudonym":"Comet"}`,
		`{"grade":4,"interests":["space","robots"]}`,
		`{"excluded_skills":["fractions","calculus"]}`,
		`{"ui_language":"pt-br"}`,
		`{"ui_language":""}`,
		`{"pseudonym":"\tOtter\n","interests":["",""]}`,
		`{"grade":"three"}`,
		`{"notes":"Loves comets."}`,
		`{"start_over":true}`,
		`{"lesson_topic":" time.clocks "}`,
		`{"lesson_topic":""}`,
		`{"lesson_topic":"astronomy.stars","grade":2}`,
		`{"lesson_topic":7}`,
		`{}`,
	} {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, arguments string) {
		if !json.Valid([]byte(arguments)) {
			quoted, err := json.Marshal(arguments)
			if err != nil {
				t.Fatalf("quoting the arguments: %v", err)
			}
			arguments = string(quoted)
		}
		if line := callOn(t, fuzzProfile(t), "edit_profile", arguments); line["outcome"] == "failed" {
			t.Errorf("edit_profile(%q) failed as ours: %v", arguments, line)
		}
	})
}

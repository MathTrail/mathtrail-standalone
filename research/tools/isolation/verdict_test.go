package main

import (
	"slices"
	"strings"
	"testing"
)

func TestOfferOfReadsEachShapeOfRequest(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		target    string
		body      string
		wantModel string
		wantTools []string
		wantReply bool
	}{
		{
			name:      "messages with a client's tool and a service's tool",
			target:    "/v1/messages?beta=true",
			body:      `{"model":"claude-x","messages":[],"tools":[{"name":"Bash","input_schema":{}},{"type":"web_search_20250305","name":"web_search"}]}`,
			wantModel: "claude-x",
			wantTools: []string{"Bash", "web_search"},
			wantReply: true,
		},
		{
			name:      "messages with no tools",
			target:    "/v1/messages",
			body:      `{"model":"claude-x","messages":[{"role":"user","content":"hi"}]}`,
			wantModel: "claude-x",
			wantReply: true,
		},
		{
			name:      "an empty list of tools",
			target:    "/v1/messages",
			body:      `{"model":"claude-x","messages":[],"tools":[]}`,
			wantModel: "claude-x",
			wantReply: true,
		},
		{
			name:      "responses with a function and a service's tool named by its type",
			target:    "/v1/responses",
			body:      `{"model":"gpt-x","input":[],"tools":[{"type":"function","name":"shell"},{"type":"web_search"}]}`,
			wantModel: "gpt-x",
			wantTools: []string{"shell", "web_search"},
			wantReply: true,
		},
		{
			name:      "gemini with functions and a built-in tool, the model in the path",
			target:    "/v1beta/models/gemini-x:streamGenerateContent?alt=sse",
			body:      `{"contents":[],"tools":[{"functionDeclarations":[{"name":"read_file"},{"name":"run_shell_command"}]},{"googleSearch":{}}]}`,
			wantModel: "gemini-x",
			wantTools: []string{"read_file", "run_shell_command", "googleSearch"},
			wantReply: true,
		},
		{
			name:      "gemini with its functions in snake case",
			target:    "/v1beta/models/gemini-x:generateContent",
			body:      `{"contents":[],"tools":[{"function_declarations":[{"name":"glob"}]}]}`,
			wantModel: "gemini-x",
			wantTools: []string{"glob"},
			wantReply: true,
		},
		{
			name:      "gemini as code assist wraps it",
			target:    "/v1internal:streamGenerateContent?alt=sse",
			body:      `{"model":"gemini-x","project":"p","request":{"contents":[],"tools":[{"functionDeclarations":[{"name":"glob"}]}]}}`,
			wantModel: "gemini-x",
			wantTools: []string{"glob"},
			wantReply: true,
		},
		{
			name:      "responses with its tools in namespaces inside an item of input",
			target:    "/v1/responses",
			body:      `{"model":"gpt-y","input":[{"type":"additional_tools","role":"developer","tools":[{"type":"namespace","name":"functions","tools":[{"type":"function","name":"exec"}]},{"type":"namespace","name":"clock","tools":[{"type":"function","name":"sleep"}]}]},{"type":"message","role":"user","content":[]}]}`,
			wantModel: "gpt-y",
			wantTools: []string{"functions.exec", "clock.sleep"},
			wantReply: true,
		},
		{
			name:      "a namespace that holds no tools",
			target:    "/v1/responses",
			body:      `{"model":"gpt-y","input":[],"tools":[{"type":"namespace","name":"collaboration","tools":[]}]}`,
			wantModel: "gpt-y",
			wantReply: true,
		},
		{
			name:      "a gemini entry that declares no functions",
			target:    "/v1beta/models/gemini-x:streamGenerateContent?alt=sse",
			body:      `{"contents":[],"tools":[{"functionDeclarations":[]}]}`,
			wantModel: "gemini-x",
			wantReply: true,
		},
		{
			name:      "tools listed somewhere no API puts them",
			target:    "/v1/messages",
			body:      `{"messages":[],"extra":{"deep":[{"tools":[{"name":"hidden"}]}]}}`,
			wantTools: []string{"hidden"},
			wantReply: true,
		},
		{
			name:      "an MCP server handed to the model",
			target:    "/v1/messages",
			body:      `{"model":"claude-x","messages":[],"mcp_servers":[{"type":"url","name":"drive","url":"https://example.com/mcp"}]}`,
			wantModel: "claude-x",
			wantTools: []string{"drive"},
			wantReply: true,
		},
		{
			name:      "tools given as bare names",
			target:    "/v1/messages",
			body:      `{"messages":[],"tools":["shell",7]}`,
			wantTools: []string{"shell", "(unnamed)"},
			wantReply: true,
		},
		{
			name:   "a request with no body",
			target: "/v1/models",
		},
		{
			name:      "a request for no reply that still offers a tool",
			target:    "/v1/tools",
			body:      `{"tools":[{"name":"Read"}]}`,
			wantTools: []string{"Read"},
		},
		{
			name:      "a gemini tool that declares built-in tools beside its functions",
			target:    "/v1beta/models/gemini-x:generateContent",
			body:      `{"contents":[],"tools":[{"functionDeclarations":[],"googleSearch":{},"codeExecution":{}}]}`,
			wantModel: "gemini-x",
			wantTools: []string{"codeExecution", "googleSearch"},
			wantReply: true,
		},
		{
			name:      "a namespace that only names and describes itself",
			target:    "/v1/responses",
			body:      `{"model":"gpt-y","input":[],"tools":[{"type":"namespace","name":"ns","description":"d","tools":[]}]}`,
			wantModel: "gpt-y",
			wantReply: true,
		},
		{
			name:      "tools keyed by name in an object",
			target:    "/v1/messages",
			body:      `{"messages":[],"tools":{"shell":{"description":"runs commands"}},"mcp_servers":{"drive":{}}}`,
			wantTools: []string{"drive", "shell"},
			wantReply: true,
		},
		{
			name:      "tools given as a single value",
			target:    "/v1/messages",
			body:      `{"messages":[],"tools":"all"}`,
			wantTools: []string{"(unnamed)"},
			wantReply: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			request := Request{Method: "POST", Target: tt.target}
			if tt.body != "" {
				request.Body = []byte(tt.body)
			}

			offer, reply, err := OfferOf(request)

			if err != nil {
				t.Fatalf("OfferOf: %v", err)
			}
			wantOffer(t, offer, reply, Offer{Model: tt.wantModel, Tools: tt.wantTools}, tt.wantReply)
		})
	}
}

func TestJudge(t *testing.T) {
	t.Parallel()
	reply := Request{Method: "POST", Target: "/v1/messages", Body: []byte(`{"model":"m","messages":[]}`)}
	replyWithTool := Request{Method: "POST", Target: "/v1/messages", Body: []byte(`{"model":"m","messages":[],"tools":[{"name":"Bash"}]}`)}
	noReply := Request{Method: "GET", Target: "/v1/models"}
	noReplyWithTool := Request{Method: "POST", Target: "/v1/tools", Body: []byte(`{"tools":[{"name":"Read"}]}`)}
	tests := []struct {
		name         string
		requests     []Request
		wantIsolated bool
		wantSummary  string
	}{
		{"no request at all", nil, false, "not shown: none of 0 requests"},
		{"no request for a reply", []Request{noReply}, false, "not shown: none of 1 request"},
		{"replies with no tools", []Request{noReply, reply, reply}, true, "isolated: 2 requests for a reply"},
		{"one reply with a tool", []Request{reply, replyWithTool}, false, "not isolated: 1 of 2 requests"},
		{"a tool offered outside a reply", []Request{noReplyWithTool, reply}, false, "not isolated: 1 of 2 requests"},
		{"tools and no reply", []Request{noReplyWithTool}, false, "not isolated: 1 of 1 request"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			verdict := Judge(tt.requests)

			if verdict.Isolated != tt.wantIsolated {
				t.Errorf("isolated = %v, want %v", verdict.Isolated, tt.wantIsolated)
			}
			if !strings.HasPrefix(verdict.Summary, tt.wantSummary) {
				t.Errorf("summary = %q, want it to start with %q", verdict.Summary, tt.wantSummary)
			}
			if len(verdict.Lines) != len(tt.requests) {
				t.Errorf("%d lines, want one per request: %d", len(verdict.Lines), len(tt.requests))
			}
		})
	}
}

func TestJudgeNamesWhatEachRequestOffered(t *testing.T) {
	t.Parallel()
	requests := []Request{
		{Method: "GET", Target: "/v1/models"},
		{Method: "POST", Target: "/v1/responses", Body: []byte(`{"model":"gpt-x","input":[],"tools":[{"type":"function","name":"shell"},{"type":"web_search"}]}`)},
	}

	verdict := Judge(requests)

	want := []string{
		"GET /v1/models: asks for no reply, no tools",
		"POST /v1/responses: model gpt-x, 2 tools: shell, web_search",
	}
	if !slices.Equal(verdict.Lines, want) {
		t.Errorf("lines = %q, want %q", verdict.Lines, want)
	}
}

func TestOfferOfRefusesABodyItCannotRead(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		request Request
	}{
		{"a JSON string", Request{Body: []byte(`"plain text"`)}},
		{"a JSON array", Request{Body: []byte(`[{"tools":[{"name":"Bash"}]}]`)}},
		{"JSON null", Request{Body: []byte(`null`)}},
		{"an encoding the endpoint could not undo", Request{Encoding: "zstd", Body: []byte(`"(µ/ý"`)}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			tt.request.Method, tt.request.Target = "POST", "/v1/responses"

			if _, _, err := OfferOf(tt.request); err == nil {
				t.Error("OfferOf read the body, want an error")
			}
		})
	}
}

func TestJudgeDoesNotPassARequestItCannotRead(t *testing.T) {
	t.Parallel()
	requests := []Request{
		{Method: "POST", Target: "/v1/messages", Body: []byte(`{"model":"m","messages":[]}`)},
		{Method: "POST", Target: "/v1/responses", Encoding: "zstd", Body: []byte(`"compressed"`)},
	}

	verdict := Judge(requests)

	if verdict.Isolated {
		t.Errorf("isolated with an unreadable request; summary %q", verdict.Summary)
	}
	if want := "not shown: 1 of 2 requests could not be read"; !strings.HasPrefix(verdict.Summary, want) {
		t.Errorf("summary = %q, want it to start with %q", verdict.Summary, want)
	}
	if want := "POST /v1/responses: cannot be read: the body is encoded as zstd"; verdict.Lines[1] != want {
		t.Errorf("line = %q, want %q", verdict.Lines[1], want)
	}
}

func wantOffer(t *testing.T, got Offer, gotReply bool, want Offer, wantReply bool) {
	t.Helper()
	if gotReply != wantReply {
		t.Errorf("reply = %v, want %v", gotReply, wantReply)
	}
	if got.Model != want.Model {
		t.Errorf("model = %q, want %q", got.Model, want.Model)
	}
	if !slices.Equal(got.Tools, want.Tools) {
		t.Errorf("tools = %q, want %q", got.Tools, want.Tools)
	}
}

package mcpserver

import (
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// A client is written down as the chat host it belongs to, a word from a
// closed list, and never by the name it gave itself.
func TestAClientIsNamedByItsHost(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		info *mcp.Implementation
		want string
	}{
		{name: "no information", info: nil, want: "unknown"},
		{name: "no name", info: &mcp.Implementation{}, want: "unknown"},
		{name: "Claude on the web", info: &mcp.Implementation{Name: "Anthropic/ClaudeAI"}, want: "claude"},
		{name: "Claude's widget calls", info: &mcp.Implementation{Name: "claude-ai"}, want: "claude"},
		{name: "Claude Code", info: &mcp.Implementation{Name: "claude-code"}, want: "claude"},
		{name: "ChatGPT", info: &mcp.Implementation{Name: "openai-mcp"}, want: "chatgpt"},
		{name: "MCP Inspector", info: &mcp.Implementation{Name: "mcp-inspector"}, want: "inspector"},
		{name: "anyone else", info: &mcp.Implementation{Name: "masha@school.example"}, want: other},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if got := clientFamily(tc.info); got != tc.want {
				t.Errorf("clientFamily() = %q, want %q", got, tc.want)
			}
		})
	}
}

// A protocol version, a method and a tool are written as they are only when
// they are the protocol's or ours; anything a caller made up is other.
func TestWhatACallerNamesIsHeldToAList(t *testing.T) {
	t.Parallel()

	b := &boundary{tools: map[string]struct{}{"get_profile": {}}}
	cases := []struct {
		name string
		got  string
		want string
	}{
		{name: "the newest protocol", got: protocolLabel("2026-07-28"), want: "2026-07-28"},
		{name: "an older protocol", got: protocolLabel("2025-11-25"), want: "2025-11-25"},
		{name: "a made-up protocol", got: protocolLabel("masha"), want: other},
		{name: "a method of the protocol", got: methodLabel("tools/list"), want: "tools/list"},
		{name: "a made-up method", got: methodLabel("masha/list"), want: other},
		{name: "a tool of ours", got: b.toolLabel("get_profile"), want: "get_profile"},
		{name: "a made-up tool", got: b.toolLabel("masha"), want: other},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if tc.got != tc.want {
				t.Errorf("label = %q, want %q", tc.got, tc.want)
			}
		})
	}
}

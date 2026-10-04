package mcpserver

import (
	"regexp"
	"strings"
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
		{name: "the load tool", info: &mcp.Implementation{Name: "mathtrail-load"}, want: "load"},
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

// The language of a task is written by its primary subtag, and only when the
// profile holds a tag the service would keep itself: the file can be edited by
// hand, and anything else typed there is other.
func TestALanguageIsNamedByItsPrimarySubtag(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		tag  string
		want string
	}{
		{name: "a language", tag: "en", want: "en"},
		{name: "a language of a country", tag: "pt-BR", want: "pt"},
		{name: "a language in a script", tag: "zh-Hant-TW", want: "zh"},
		{name: "a tag set about with spaces", tag: " ru ", want: "ru"},
		{name: "no tag", tag: "", want: other},
		{name: "a pseudonym typed in by hand", tag: "Otter", want: other},
		{name: "a country that only suggests a language", tag: "und-US", want: other},
		{name: "a private tag", tag: "x-mathtrail", want: other},
		{name: "no language at all", tag: "zxx", want: other},
		{name: "a tag longer than any the service keeps", tag: "en-" + strings.Repeat("abcdefgh-", 5), want: other},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if got := languageLabel(tc.tag); got != tc.want {
				t.Errorf("languageLabel(%q) = %q, want %q", tc.tag, got, tc.want)
			}
		})
	}
}

// Whatever the profile says a task's language is, the line names a primary
// language subtag or other, and never the text that was there.
func FuzzLanguageLabel(f *testing.F) {
	for _, seed := range []string{"en", "pt-BR", "zh-Hant-TW", "", "Otter", "und-US", "x-mathtrail", "zxx", "en-GB-oxendict", "sr-Latn-RS"} {
		f.Add(seed)
	}
	subtag := regexp.MustCompile(`^[a-z]{2,3}$`)
	f.Fuzz(func(t *testing.T, tag string) {
		if got := languageLabel(tag); got != other && !subtag.MatchString(got) {
			t.Fatalf("languageLabel(%q) = %q, want a primary language subtag or %q", tag, got, other)
		}
	})
}

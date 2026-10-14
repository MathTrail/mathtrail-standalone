package mcpserver

import (
	"encoding/json"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// capabilitiesOf is a client's capabilities as JSON brings them, the way a
// client sends them with a call.
func capabilitiesOf(t *testing.T, declared string) *mcp.ClientCapabilities {
	t.Helper()

	var client mcp.ClientCapabilities
	if err := json.Unmarshal([]byte(declared), &client); err != nil {
		t.Fatalf("json.Unmarshal(%s) error = %v, want nil", declared, err)
	}
	return &client
}

// What a client declared of MCP Apps is MCP Apps under its extension's name,
// whatever kinds of page it names for them, or none; and unknown where its call
// came with no capabilities the server can read.
func TestWhatAClientDeclaredOfMCPAppsIsRead(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name, declared, want string
	}{
		{"MCP Apps for the widget's page", `{"extensions":{"io.modelcontextprotocol/ui":{"mimeTypes":["text/html;profile=mcp-app"]}}}`, appsDeclared},
		{"MCP Apps for the page written with a space", `{"extensions":{"io.modelcontextprotocol/ui":{"mimeTypes":["text/html; profile=mcp-app"]}}}`, appsDeclared},
		{"MCP Apps naming no page", `{"extensions":{"io.modelcontextprotocol/ui":{}}}`, appsDeclared},
		{"another extension alone", `{"extensions":{"openai/elicitation":{"form":{}}}}`, appsNone},
		{"no extension", `{"elicitation":{}}`, appsNone},
		{"nothing", `{}`, appsNone},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := appsOf(capabilitiesOf(t, tc.declared)); got != tc.want {
				t.Errorf("appsOf(%s) = %q, want %q", tc.declared, got, tc.want)
			}
		})
	}

	t.Run("capabilities the server cannot read", func(t *testing.T) {
		t.Parallel()
		if got := appsOf(nil); got != appsUnknown {
			t.Errorf("appsOf(nil) = %q, want %q", got, appsUnknown)
		}
	})
}

// A chat draws the cards when its client declares MCP Apps, when the server
// cannot read what it declares, and whenever it is ChatGPT, which draws them on
// every surface it has; any other client that declares no MCP Apps shows the
// words alone.
func TestAChatIsWhatItsClientDeclaresOrChatGPT(t *testing.T) {
	t.Parallel()

	declared := `{"extensions":{"io.modelcontextprotocol/ui":{"mimeTypes":["text/html;profile=mcp-app"]}}}`
	for _, tc := range []struct {
		name, client, declared string
		unread                 bool
		want                   chatKind
	}{
		{name: "Claude, declaring MCP Apps", client: "claude-ai", declared: declared, want: cardsChat},
		{name: "Claude Code, declaring none", client: "claude-code", declared: `{}`, want: wordsChat},
		{name: "ChatGPT, declaring none", client: "openai-mcp", declared: `{}`, want: cardsChat},
		{name: "ChatGPT by its other name, declaring none", client: "ChatGPT", declared: `{}`, want: cardsChat},
		{name: "a client of no known host, declaring none", client: "some-client", declared: `{}`, want: wordsChat},
		{name: "a client whose capabilities are unread", client: "some-client", unread: true, want: cardsChat},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var client *mcp.ClientCapabilities
			if !tc.unread {
				client = capabilitiesOf(t, tc.declared)
			}
			if got := chatOf(client, &mcp.Implementation{Name: tc.client}); got != tc.want {
				t.Errorf("chatOf(%s, %q) = %s, want %s", tc.declared, tc.client, chatNames[got], chatNames[tc.want])
			}
		})
	}
}

// The frame tells a tool its chat in the call's context, and a call that came
// with none is taken to come from a chat that draws the cards.
func TestAToolHearsTheChatItsCallCameFrom(t *testing.T) {
	t.Parallel()

	if got := chatFrom(t.Context()); got != cardsChat {
		t.Errorf("chatFrom of a context told nothing = %s, want cardsChat", chatNames[got])
	}
	for _, chat := range []chatKind{cardsChat, wordsChat} {
		if got := chatFrom(withChat(t.Context(), chat)); got != chat {
			t.Errorf("chatFrom(withChat(%s)) = %s, want %s", chatNames[chat], chatNames[got], chatNames[chat])
		}
	}
}

// chatNames are the chats by the names a failure tells them by.
var chatNames = map[chatKind]string{cardsChat: "cardsChat", wordsChat: "wordsChat"}

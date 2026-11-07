package mcpserver

import (
	"encoding/json"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/profile"
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
// words alone. A task solved from the picture goes only where the card is
// known to be drawn: to ChatGPT and to a client that declares MCP Apps, never
// to one whose capabilities the server cannot read.
func TestAChatIsWhatItsClientDeclaresOrChatGPT(t *testing.T) {
	t.Parallel()

	for _, tc := range clients {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			client := tc.capabilities(t)
			info := &mcp.Implementation{Name: tc.client}
			if got := chatOf(client, info); got != tc.words {
				t.Errorf("chatOf(%s, %q) = %s, want %s", tc.declared, tc.client, chatNames[got], chatNames[tc.words])
			}
			if got := pictureChatOf(client, info); got != tc.pictures {
				t.Errorf("pictureChatOf(%s, %q) = %s, want %s", tc.declared, tc.client, pictureChatNames[got],
					pictureChatNames[tc.pictures])
			}
		})
	}
}

// The frame tells a tool its chat in the call's context, both ways it is read,
// and a call that came with none is taken to come from a chat that draws the
// cards, but not from one a task solved from the picture may go to.
func TestAToolHearsTheChatItsCallCameFrom(t *testing.T) {
	t.Parallel()

	if got := chatFrom(t.Context()); got != cardsChat {
		t.Errorf("chatFrom of a context told nothing = %s, want cardsChat", chatNames[got])
	}
	if got := pictureChatFrom(t.Context()); got != profile.ChatOfWords {
		t.Errorf("pictureChatFrom of a context told nothing = %s, want ChatOfWords", pictureChatNames[got])
	}
	for _, tc := range clients {
		ctx := withChat(t.Context(), tc.capabilities(t), &mcp.Implementation{Name: tc.client})
		if got := chatFrom(ctx); got != tc.words {
			t.Errorf("%s: chatFrom(withChat) = %s, want %s", tc.name, chatNames[got], chatNames[tc.words])
		}
		if got := pictureChatFrom(ctx); got != tc.pictures {
			t.Errorf("%s: pictureChatFrom(withChat) = %s, want %s", tc.name, pictureChatNames[got],
				pictureChatNames[tc.pictures])
		}
	}
}

// aClient is a client as a call names it and what it declares with the call,
// and the chat it is taken for, as its words are told it and as a task solved
// from the picture is given to it.
type aClient struct {
	name, client, declared string
	unread                 bool
	words                  chatKind
	pictures               profile.Chat
}

// capabilities are what the client's call carries: none the server can read
// for a client on an earlier version of the protocol.
func (c *aClient) capabilities(t *testing.T) *mcp.ClientCapabilities {
	t.Helper()

	if c.unread {
		return nil
	}
	return capabilitiesOf(t, c.declared)
}

// declaringApps is a client's declaration of MCP Apps for the widget's page.
const declaringApps = `{"extensions":{"io.modelcontextprotocol/ui":{"mimeTypes":["text/html;profile=mcp-app"]}}}`

// clients are the clients a call may come from.
var clients = []aClient{
	{name: "Claude, declaring MCP Apps", client: "claude-ai", declared: declaringApps, words: cardsChat,
		pictures: profile.ChatWithCards},
	{name: "Claude Code, declaring none", client: "claude-code", declared: `{}`, words: wordsChat,
		pictures: profile.ChatOfWords},
	{name: "ChatGPT, declaring none", client: "openai-mcp", declared: `{}`, words: cardsChat,
		pictures: profile.ChatWithCards},
	{name: "ChatGPT by its other name, declaring none", client: "ChatGPT", declared: `{}`, words: cardsChat,
		pictures: profile.ChatWithCards},
	{name: "ChatGPT, whose capabilities are unread", client: "chatgpt", unread: true, words: cardsChat,
		pictures: profile.ChatWithCards},
	{name: "a client of no known host, declaring none", client: "some-client", declared: `{}`, words: wordsChat,
		pictures: profile.ChatOfWords},
	{name: "a client whose capabilities are unread", client: "some-client", unread: true, words: cardsChat,
		pictures: profile.ChatOfWords},
}

// chatNames are the chats by the names a failure tells them by, and
// pictureChatNames the same chats as a task solved from the picture is given
// to them.
var (
	chatNames        = map[chatKind]string{cardsChat: "cardsChat", wordsChat: "wordsChat"}
	pictureChatNames = map[profile.Chat]string{profile.ChatOfWords: "ChatOfWords", profile.ChatWithCards: "ChatWithCards"}
)

package mcpserver

import (
	"encoding/json"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// A chat draws the cards when its client declares MCP Apps, whatever kinds of
// page it names for them, and shows the words alone when it declares none. A
// client whose capabilities the server cannot read is taken to draw them. The
// capabilities are read as JSON brings them, the way a client sends them with a
// call.
func TestAChatIsWhatItsClientDeclares(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name, declared string
		want           chatKind
	}{
		{"MCP Apps for the widget's page", `{"extensions":{"io.modelcontextprotocol/ui":{"mimeTypes":["text/html;profile=mcp-app"]}}}`, cardsChat},
		{"MCP Apps for the page written with a space", `{"extensions":{"io.modelcontextprotocol/ui":{"mimeTypes":["text/html; profile=mcp-app"]}}}`, cardsChat},
		{"MCP Apps naming no page", `{"extensions":{"io.modelcontextprotocol/ui":{}}}`, cardsChat},
		{"another extension alone", `{"extensions":{"openai/elicitation":{"form":{}}}}`, wordsChat},
		{"no extension", `{"elicitation":{}}`, wordsChat},
		{"nothing", `{}`, wordsChat},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var client mcp.ClientCapabilities
			if err := json.Unmarshal([]byte(tc.declared), &client); err != nil {
				t.Fatalf("json.Unmarshal(%s) error = %v, want nil", tc.declared, err)
			}
			if got := chatOf(&client); got != tc.want {
				t.Errorf("chatOf(%s) = %s, want %s", tc.declared, chatNames[got], chatNames[tc.want])
			}
		})
	}

	t.Run("capabilities the server cannot read", func(t *testing.T) {
		t.Parallel()
		if got := chatOf(nil); got != cardsChat {
			t.Errorf("chatOf(nil) = %s, want cardsChat", chatNames[got])
		}
	})
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

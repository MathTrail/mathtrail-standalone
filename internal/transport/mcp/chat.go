package mcpserver

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// chatKind is how the chat a call came from shows a tool's result, as the
// client that made the call says: drawn as a card, or as the words alone. The
// words a tool hands back tell each chat its own case alone. Told both, a model
// that never sees the card it drew may take the case without one for its own,
// as ChatGPT's did: it read the task out under the card that showed it, and
// drew a quiz of its own beside it.
type chatKind int

const (
	// cardsChat is a chat that draws a tool's result as a card. A client that
	// says nothing the server can read is taken for one, as every chat
	// MathTrail is listed in draws cards.
	cardsChat chatKind = iota
	// wordsChat is a chat that shows the words alone: its client declares no
	// MCP Apps.
	wordsChat
)

// appsExtension names MCP Apps among the extensions a client declares: the one
// that draws a tool's result as a card.
const appsExtension = "io.modelcontextprotocol/ui"

// chatOf is the chat a client's capabilities describe. A client that declares
// MCP Apps draws the cards, whatever kinds of page it names for them: a chat
// taken wrongly for one that draws cards costs the adult a message, to say no
// card shows the task, while one taken wrongly for words alone has the task
// written out under the card that shows it. A client on the newest version of
// the protocol declares its capabilities with every request. One on an earlier
// version declared them once, in a handshake a server that keeps no session
// never sees, and its calls come with none, which is no word that it draws no
// cards.
func chatOf(client *mcp.ClientCapabilities) chatKind {
	if client == nil {
		return cardsChat
	}
	if _, declared := client.Extensions[appsExtension]; declared {
		return cardsChat
	}
	return wordsChat
}

// chatKey holds the chat a call came from. It is a type of its own, so that
// nothing else a context carries can be taken for it.
type chatKey struct{}

// withChat is how the frame of a call tells the tool which chat the call came
// from. The context carries it to the function that words the tool's reply,
// and from there it travels as an argument: it changes the words a tool hands
// back, never the work the tool does, so it is no argument of every tool's.
func withChat(ctx context.Context, chat chatKind) context.Context {
	return context.WithValue(ctx, chatKey{}, chat)
}

// chatFrom is the chat the frame put in ctx. A call that came with none is
// taken to come from a chat that draws the cards, as a client that says nothing
// is: cardsChat is the zero chatKind.
func chatFrom(ctx context.Context) chatKind {
	chat, _ := ctx.Value(chatKey{}).(chatKind)
	return chat
}

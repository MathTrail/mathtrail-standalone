package mcpserver

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/MathTrail/mathtrail-standalone/internal/store"
)

// Tool is one tool of the endpoint, with everything the frame needs to serve
// it. Define makes one, and nothing else can: a tool added any other way would
// skip the frame.
type Tool interface {
	name() string
	add(server *mcp.Server) error
}

// Spec is how a tool presents itself to the host and to the model that calls
// it.
type Spec struct {
	// Name is what a call asks for. A host may show it with a prefix of its
	// own, so nothing written for the model depends on its being exact.
	Name string
	// Title is the name a person reads.
	Title string
	// Description says what the tool does, what it takes, returns and changes,
	// and when to call it. How to behave in a lesson is the instructions' to
	// say: a directory turns down a description that tells the model how to
	// behave.
	Description string
	// Effect is what the tool does to the parent's file. Every tool says it:
	// the hints a host asks a person's permission by are made from it.
	Effect Effect
	// Idempotent says a repeated call with the same arguments changes nothing
	// the first one did not.
	Idempotent bool
	// DrawsCard says a host draws the tool's result as a card, with the
	// widget. Which screen the card shows is for the payload to say.
	DrawsCard bool
	// WidgetOnly hides the tool from the model: only a card calls it, for what
	// that card shows, and the model has another tool for the same thing or no
	// use for it. Such a tool draws no card, since the card that called it is
	// the one showing what it answers.
	WidgetOnly bool
}

// Effect is what a tool does to the parent's file, said once, so that the two
// hints a host reads of it — that the tool only reads, and that it may destroy
// something — can never disagree. No tool reaches anything beyond that file.
type Effect int

const (
	// effectUnsaid is a tool that has not said what it does, which no tool is
	// let be.
	effectUnsaid Effect = iota
	// Reads changes nothing, whatever it finds: a damaged file is told, not
	// mended.
	Reads
	// Adds writes only what adds to the record of the lessons — a request, an
	// attempt, a task handed out or skipped, an answer — with the ratings and
	// counts that follow from it. It never replaces what the adult set, and
	// never deletes a state of the file.
	Adds
	// Overwrites replaces what the adult set, or puts an earlier state of the
	// file back over the one it holds: what it replaces is gone from the
	// profile.
	Overwrites
)

// Handler does one tool's work for one call, for the account the call acts
// for.
type Handler[In, Out any] func(ctx context.Context, account store.Account, in In) (Reply[Out], error)

// Reply is what a tool hands back when it did its work: the words a model
// relays — the whole lesson when no widget is drawn — and the payload a widget
// draws from. A refusal the model can act on is a reply too, with a status
// saying so in its payload; an error is for a tool that could not do its work.
type Reply[Out any] struct {
	Text    string
	Payload Out
}

// definition is a tool as Define made it.
type definition[In, Out any] struct {
	spec   Spec
	handle Handler[In, Out]
}

// Define makes a tool of a description and the handler that does its work.
func Define[In, Out any](spec Spec, handle Handler[In, Out]) Tool {
	return definition[In, Out]{spec: spec, handle: handle}
}

func (d definition[In, Out]) name() string { return d.spec.Name }

// add registers the tool with the protocol, which derives its input and output
// schemas from In and Out and checks every call's arguments against the first
// before the handler sees them. The library panics on a type it cannot derive
// a schema from; that is a mistake in the tool's definition, and it is told as
// a refusal to build the endpoint rather than as a crash.
func (d definition[In, Out]) add(server *mcp.Server) (err error) {
	switch {
	case d.spec.WidgetOnly && d.spec.DrawsCard:
		// A host would draw a second card under the one that asked.
		return fmt.Errorf("%w: tool %q is for a card alone and cannot draw one", ErrSettings, d.spec.Name)
	case d.spec.Title == "":
		// A host shows the title when it asks a person about a call.
		return fmt.Errorf("%w: tool %q has no title a person can read", ErrSettings, d.spec.Name)
	case d.spec.Effect == effectUnsaid:
		return fmt.Errorf("%w: tool %q does not say what it does to the parent's file", ErrSettings, d.spec.Name)
	}
	defer func() {
		if recovered := recover(); recovered != nil {
			err = fmt.Errorf("%w: tool %q cannot be defined: %v", ErrSettings, d.spec.Name, recovered)
		}
	}()
	mcp.AddTool(server, d.tool(), d.call)
	return nil
}

// tool is the definition a host lists.
//
// Every hint is said, none left to the default a host would read into its
// absence. A tool only reads when its effect says so, and may destroy
// something only when it overwrites; and since no tool reaches anything beyond
// the parent's own file, none is open to the world. Every tool also says it
// wants the one scope a token grants, where a host that reads such a
// declaration looks for it.
//
// A tool that draws a card names the widget's page twice: under the key hosts
// read now, and under the flat key earlier hosts read, as the library's own
// helper for such tools writes both. A tool for a card alone says who may see
// it, which a host reads as keeping it from the model.
func (d definition[In, Out]) tool() *mcp.Tool {
	no, destructive := false, d.spec.Effect == Overwrites
	meta := mcp.Meta{
		"securitySchemes": []map[string]any{{"type": "oauth2", "scopes": []string{Scope}}},
	}
	ui := map[string]any{}
	if d.spec.DrawsCard {
		ui["resourceUri"] = WidgetURI
		meta["ui/resourceUri"] = WidgetURI
	}
	if d.spec.WidgetOnly {
		ui["visibility"] = []string{"app"}
	}
	if len(ui) > 0 {
		meta["ui"] = ui
	}
	return &mcp.Tool{
		Name:        d.spec.Name,
		Title:       d.spec.Title,
		Description: d.spec.Description,
		Annotations: &mcp.ToolAnnotations{
			Title:           d.spec.Title,
			ReadOnlyHint:    d.spec.Effect == Reads,
			IdempotentHint:  d.spec.Idempotent,
			DestructiveHint: &destructive,
			OpenWorldHint:   &no,
		},
		Meta: meta,
	}
}

// call runs the handler for the account the call acts for, and turns whatever
// it hands back into what the protocol sends.
//
// An error never reaches the protocol as it is: the protocol would put its
// text in front of the model, and the text of an error is whatever the code
// that made it had in hand. It is replaced by a failure, whose text is one
// sentence of ours.
func (d definition[In, Out]) call(ctx context.Context, _ *mcp.CallToolRequest, in In) (*mcp.CallToolResult, Out, error) {
	var none Out
	account, signedIn := accountFrom(ctx)
	if !signedIn {
		return nil, none, explain(errNotSignedIn)
	}
	reply, err := d.handle(ctx, account, in)
	if err != nil {
		return nil, none, explain(err)
	}
	return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: reply.Text}}}, reply.Payload, nil
}

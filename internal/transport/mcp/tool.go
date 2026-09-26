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
	// Description tells the model when and how to call the tool.
	Description string
	// ReadOnly says the tool changes nothing.
	ReadOnly bool
	// Idempotent says a repeated call with the same arguments changes nothing
	// the first one did not.
	Idempotent bool
}

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
// Every tool changes nothing but the parent's own file and reaches nothing
// beyond it, so neither hint is left to its default of true. Every tool also
// says it wants the one scope a token grants, where a host that reads such a
// declaration looks for it.
func (d definition[In, Out]) tool() *mcp.Tool {
	no := false
	return &mcp.Tool{
		Name:        d.spec.Name,
		Title:       d.spec.Title,
		Description: d.spec.Description,
		Annotations: &mcp.ToolAnnotations{
			Title:           d.spec.Title,
			ReadOnlyHint:    d.spec.ReadOnly,
			IdempotentHint:  d.spec.Idempotent,
			DestructiveHint: &no,
			OpenWorldHint:   &no,
		},
		Meta: mcp.Meta{
			"securitySchemes": []map[string]any{{"type": "oauth2", "scopes": []string{Scope}}},
		},
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

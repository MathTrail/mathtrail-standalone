package mcpserver

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// WidgetURI names the page every card is drawn by. A tool whose result a host
// should draw as a card points the host here.
const WidgetURI = "ui://mathtrail/app.html"

// widgetMIME tells a host the page is an app to run in its sandbox rather than
// a document to show.
const widgetMIME = "text/html;profile=mcp-app"

// addWidget serves the widget's page as a resource.
//
// What a host needs to know about the page travels twice: with the resource
// where it is listed, and with the page where it is read. A host takes what
// came with the page and falls back to the listing, and the library copies
// neither onto the other.
func addWidget(server *mcp.Server, page string) {
	server.AddResource(&mcp.Resource{
		URI:      WidgetURI,
		Name:     "widget",
		Title:    "MathTrail",
		MIMEType: widgetMIME,
		Meta:     widgetMeta(),
	}, func(context.Context, *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
		return widgetRead(page), nil
	})
}

// widgetRead is one read of the widget: the page, and what a host needs to
// know about it. Every read gets a result of its own, down to its maps. The
// library writes into whatever a handler returns — the address, the type, the
// cache hint, the server's own name — so a result shared between reads would
// be written by several of them at once.
func widgetRead(page string) *mcp.ReadResourceResult {
	return &mcp.ReadResourceResult{
		Contents: []*mcp.ResourceContents{{
			URI:      WidgetURI,
			MIMEType: widgetMIME,
			Text:     page,
			Meta:     widgetMeta(),
		}},
	}
}

// widgetMeta is what a host reads about the page before it runs it.
//
// All four lists of the page's content security policy are empty, and sent as
// empty lists rather than left out: the page may reach no network, load
// nothing from anywhere, frame nothing and set its base nowhere else. A widget
// that cannot reach the network cannot leak what it holds, and what it holds
// is a child's task. It asks for no permission and for no origin of its own,
// and for a visible border: a card is a task, and a boundary is what makes it
// read as one.
func widgetMeta() mcp.Meta {
	return mcp.Meta{"ui": map[string]any{
		"csp": map[string]any{
			"connectDomains":  []string{},
			"resourceDomains": []string{},
			"frameDomains":    []string{},
			"baseUriDomains":  []string{},
		},
		"prefersBorder": true,
	}}
}

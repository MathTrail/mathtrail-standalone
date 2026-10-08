package mcpserver_test

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/jsonrpc"
	"github.com/modelcontextprotocol/go-sdk/mcp"

	mcpserver "github.com/MathTrail/mathtrail-standalone/internal/transport/mcp"
)

// widgetPolicy is what a host must be told about the page before it runs it:
// no network, nothing loaded from anywhere, no frames and no base elsewhere —
// four lists that are empty and said to be — and no border of the host's,
// since the card draws its own. Under ChatGPT's own keys, the same lists, the
// site as the one origin a card's links lead to, and the service's origin to
// make the page's sandbox from.
const widgetPolicy = `{"openai/widgetCSP":{"connect_domains":[],"frame_domains":[],"redirect_domains":["` + siteOrigin +
	`"],"resource_domains":[]},"openai/widgetDomain":"` + serviceOrigin + `",` +
	`"ui":{"csp":{"baseUriDomains":[],"connectDomains":[],"frameDomains":[],"resourceDomains":[]},"prefersBorder":false}}`

// metaJSON is a description's metadata as it goes out.
func metaJSON(t *testing.T, meta mcp.Meta) string {
	t.Helper()

	encoded, err := json.Marshal(meta)
	if err != nil {
		t.Fatalf("the metadata does not marshal: %v", err)
	}
	return string(encoded)
}

// wantDescribedAsTheWidget fails unless what a host was handed — the listing
// or the page — names the widget as an app and carries its policy.
func wantDescribedAsTheWidget(t *testing.T, uri, mimeType string, meta mcp.Meta) {
	t.Helper()

	if uri != mcpserver.WidgetURI {
		t.Errorf("uri = %q, want %q", uri, mcpserver.WidgetURI)
	}
	if mimeType != "text/html;profile=mcp-app" {
		t.Errorf("mimeType = %q, want text/html;profile=mcp-app", mimeType)
	}
	if got := metaJSON(t, meta); got != widgetPolicy {
		t.Errorf("_meta = %s, want %s", got, widgetPolicy)
	}
}

// The widget is the endpoint's one resource, listed with the type that makes a
// host run it as an app and with the policy it is run under, whichever version
// of the protocol the host speaks.
func TestTheWidgetIsListedAsAnApp(t *testing.T) {
	t.Parallel()

	for _, version := range []string{"", "2025-11-25"} {
		t.Run("protocol "+version, func(t *testing.T) {
			t.Parallel()

			h := serve(t, mcpserver.DevSignIn)
			listed, err := h.connect(t, version).ListResources(t.Context(), nil)
			if err != nil {
				t.Fatalf("ListResources() error = %v, want nil", err)
			}
			if len(listed.Resources) != 1 {
				t.Fatalf("resources = %d, want the widget alone", len(listed.Resources))
			}
			widget := listed.Resources[0]
			wantDescribedAsTheWidget(t, widget.URI, widget.MIMEType, widget.Meta)
		})
	}
}

// Reading the widget gives the page the endpoint was built with, and the page
// carries its policy with it: a host takes the policy that came with the page
// before the one in the listing.
func TestTheWidgetIsReadWithItsPolicy(t *testing.T) {
	t.Parallel()

	for _, version := range []string{"", "2025-11-25"} {
		t.Run("protocol "+version, func(t *testing.T) {
			t.Parallel()

			h := serve(t, mcpserver.DevSignIn)
			read, err := h.connect(t, version).ReadResource(t.Context(),
				&mcp.ReadResourceParams{URI: mcpserver.WidgetURI})
			if err != nil {
				t.Fatalf("ReadResource() error = %v, want nil", err)
			}
			if len(read.Contents) != 1 {
				t.Fatalf("contents = %d, want the page alone", len(read.Contents))
			}
			content := read.Contents[0]
			if content.Text != page {
				t.Errorf("text = %q, want the page the endpoint was built with", content.Text)
			}
			wantDescribedAsTheWidget(t, content.URI, content.MIMEType, content.Meta)
		})
	}
}

// Any other address is answered in the protocol's own words: the resource is
// not there.
func TestAnotherResourceIsNotFound(t *testing.T) {
	t.Parallel()

	h := serve(t, mcpserver.DevSignIn)
	_, err := h.connect(t, "").ReadResource(t.Context(),
		&mcp.ReadResourceParams{URI: "ui://mathtrail/other.html"})

	var wire *jsonrpc.Error
	if !errors.As(err, &wire) {
		t.Fatalf("ReadResource() error = %v, want the protocol's refusal", err)
	}
	if wire.Code != jsonrpc.CodeInvalidParams || wire.Message != "Resource not found" {
		t.Errorf("error = %d %q, want %d %q", wire.Code, wire.Message, jsonrpc.CodeInvalidParams, "Resource not found")
	}
}

// The handshake offers resources, and does not promise that their list will
// change: nothing would ever announce it.
func TestResourcesAreOfferedAsAListThatDoesNotChange(t *testing.T) {
	t.Parallel()

	h := serve(t, mcpserver.DevSignIn)
	resp := h.post(t, `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-11-25",`+
		`"capabilities":{},"clientInfo":{"name":"probe","version":"0"}}}`, nil)

	result, isObject := message(t, resp)["result"].(map[string]any)
	if !isObject {
		t.Fatalf("the handshake has no result: %s", resp.body)
	}
	capabilities, _ := result["capabilities"].(map[string]any)
	resources, offered := capabilities["resources"].(map[string]any)
	if !offered {
		t.Fatalf("capabilities = %v, want resources among them", capabilities)
	}
	if len(resources) != 0 {
		t.Errorf("resources = %v, want them offered with no promise of changes", resources)
	}
}

// A tool that draws a card points the host at the widget, under the key hosts
// read now and the one earlier hosts read; a tool that draws none says nothing
// about a widget.
func TestAToolThatDrawsACardPointsAtTheWidget(t *testing.T) {
	t.Parallel()

	h := serve(t, mcpserver.DevSignIn)
	listed, err := h.connect(t, "").ListTools(t.Context(), nil)
	if err != nil {
		t.Fatalf("ListTools() error = %v, want nil", err)
	}
	byName := map[string]*mcp.Tool{}
	for _, tool := range listed.Tools {
		byName[tool.Name] = tool
	}

	echo, refuse := byName["echo"], byName["refuse"]
	if echo == nil || refuse == nil {
		t.Fatalf("the tools listed are %v, want echo and refuse among them", listed.Tools)
	}
	if got, want := metaJSON(t, mcp.Meta{"ui": echo.Meta["ui"]}), `{"ui":{"resourceUri":"`+mcpserver.WidgetURI+`"}}`; got != want {
		t.Errorf("echo _meta.ui = %s, want %s", got, want)
	}
	if got := echo.Meta["ui/resourceUri"]; got != mcpserver.WidgetURI {
		t.Errorf(`echo _meta["ui/resourceUri"] = %v, want %q`, got, mcpserver.WidgetURI)
	}
	for _, key := range []string{"ui", "ui/resourceUri"} {
		if value, named := refuse.Meta[key]; named {
			t.Errorf("refuse _meta[%q] = %v, want nothing about a widget from a tool that draws no card", key, value)
		}
	}
}

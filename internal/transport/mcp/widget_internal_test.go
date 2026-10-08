package mcpserver

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Every read of the widget gets a result that shares nothing with another: what
// the library writes into one — and anything else that changes it — leaves the
// next read as it was.
func TestEveryReadOfTheWidgetGetsAResultOfItsOwn(t *testing.T) {
	t.Parallel()

	sandbox := Sandbox{Origin: "https://mcp.example", Site: "https://site.example"}
	first, second := widgetRead(WidgetURI, "<p>the page</p>", sandbox), widgetRead(WidgetURI, "<p>the page</p>", sandbox)
	want, err := json.Marshal(second)
	if err != nil {
		t.Fatalf("the result does not marshal: %v", err)
	}

	// What the library writes into a result a handler returns, and then some.
	first.CacheScope = "private"
	first.Meta = mcp.Meta{"io.modelcontextprotocol/serverInfo": "the server"}
	first.Contents[0].Meta["written"] = true
	ui, _ := first.Contents[0].Meta["ui"].(map[string]any)
	border, _ := ui["prefersBorder"].(bool)
	ui["prefersBorder"] = !border
	csp, _ := ui["csp"].(map[string]any)
	csp["connectDomains"] = []string{"https://elsewhere.example"}
	widgetCSP, _ := first.Contents[0].Meta["openai/widgetCSP"].(map[string]any)
	redirects, _ := widgetCSP["redirect_domains"].([]string)
	redirects[0] = "https://elsewhere.example"

	got, err := json.Marshal(second)
	if err != nil {
		t.Fatalf("the result does not marshal: %v", err)
	}
	if !bytes.Equal(got, want) {
		t.Errorf("the second read became %s after the first was written, want it %s", got, want)
	}
}

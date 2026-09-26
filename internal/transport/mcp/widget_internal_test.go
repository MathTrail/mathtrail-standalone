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

	first, second := widgetRead("<p>the page</p>"), widgetRead("<p>the page</p>")
	want, err := json.Marshal(second)
	if err != nil {
		t.Fatalf("the result does not marshal: %v", err)
	}

	// What the library writes into a result a handler returns, and then some.
	first.CacheScope = "private"
	first.Meta = mcp.Meta{"io.modelcontextprotocol/serverInfo": "the server"}
	first.Contents[0].Meta["written"] = true
	ui, _ := first.Contents[0].Meta["ui"].(map[string]any)
	ui["prefersBorder"] = false
	csp, _ := ui["csp"].(map[string]any)
	csp["connectDomains"] = []string{"https://elsewhere.example"}

	got, err := json.Marshal(second)
	if err != nil {
		t.Fatalf("the result does not marshal: %v", err)
	}
	if !bytes.Equal(got, want) {
		t.Errorf("the second read became %s after the first was written, want it %s", got, want)
	}
}

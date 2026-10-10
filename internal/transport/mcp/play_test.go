package mcpserver_test

import (
	"encoding/json"
	"os"
	"regexp"
	"slices"
	"strings"
	"testing"

	"github.com/MathTrail/mathtrail-standalone/internal/store/memory"
)

// playTools is the line of the justfile that names the tools a live run of
// the lesson lets the chat's model call, and playServer the name the run gives
// the server, which the chat puts before the name of each of its tools.
var (
	playTools  = regexp.MustCompile(`(?m)^PLAY_TOOLS := "([^"]*)"$`)
	playServer = regexp.MustCompile(`"mcpServers": \{"([^"]+)":`)
)

// A live run of the lesson lets the chat's model call the tools it names, and
// refuses it a call of any other. It names every tool a host shows the model,
// so that the run meets the lesson a host holds, and none a card alone calls,
// since a host keeps those from the model.
func TestALiveRunLetsTheModelCallEveryToolItIsShown(t *testing.T) {
	t.Parallel()

	written, err := os.ReadFile("../../../justfile")
	if err != nil {
		t.Fatalf("read the justfile: %v", err)
	}
	line := playTools.FindSubmatch(written)
	if line == nil {
		t.Fatalf("the justfile names no PLAY_TOOLS")
	}
	server := playServer.FindSubmatch(written)
	if server == nil {
		t.Fatalf("the justfile names no server in the .mcp.json of the run's chat")
	}
	named := strings.Fields(string(line[1]))
	slices.Sort(named)
	prefix := "mcp__" + string(server[1]) + "__"

	_, session := lesson(t, memory.New())
	listed, err := session.ListTools(t.Context(), nil)
	if err != nil {
		t.Fatalf("ListTools() error = %v", err)
	}
	var shown []string
	for _, tool := range listed.Tools {
		ui, _ := tool.Meta["ui"].(map[string]any)
		visibility, _ := json.Marshal(ui["visibility"])
		if string(visibility) != `["app"]` {
			shown = append(shown, prefix+tool.Name)
		}
	}
	slices.Sort(shown)

	if !slices.Equal(named, shown) {
		t.Errorf("PLAY_TOOLS = %v, want the tools the model is shown: %v", named, shown)
	}
}

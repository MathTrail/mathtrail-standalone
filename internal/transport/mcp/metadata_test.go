package mcpserver_test

import (
	"encoding/json"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// aMoment is a date with its time of day, as RFC 3339 writes one.
var aMoment = regexp.MustCompile(`\d{4}-\d{2}-\d{2}T\d{2}:\d{2}`)

// Nothing a tool hands back carries a moment, an age or the version of the
// instructions, in its words or in its payload: the model has no use for any
// of them, and a host reviews them as logging metadata. The ids a later call
// passes back — of a request, of a task — stay.
func TestNoResultCarriesAMomentAnAgeOrAVersion(t *testing.T) {
	t.Parallel()

	loaded, err := shipped()
	if err != nil {
		t.Fatalf("load the content: %v", err)
	}
	kept, moving := racer(t), &clock{at: lessonDay}
	_, session := lessonWith(t, kept, moving, nil)

	results := map[string]*mcp.CallToolResult{}
	request := askForTheRace(t, session, kept)
	moving.advance(30 * time.Second)
	results["next_task asked again"] = call(t, session, "next_task", raceChoice)
	results["get_package"] = call(t, session, "get_package", map[string]any{"request_id": request.ID})
	results["submit_task"] = call(t, session, "submit_task", raceOn(request))
	results["read_task"] = call(t, session, "read_task", map[string]any{"request_id": request.ID})
	p, _ := loadKept(t, kept)
	if p.CurrentTask == nil {
		t.Fatal("no task is on the card, want the race submit_task accepted")
	}
	results["submit_answer"] = call(t, session, "submit_answer", map[string]any{"task_id": p.CurrentTask.ID, "answer": "C"})
	for _, tool := range []string{"get_progress", "read_progress", "get_profile"} {
		results[tool] = call(t, session, tool, map[string]any{})
	}

	for name, result := range results {
		wantNoMetadata(t, name, result, loaded.InstructionsVersion())
	}
}

// wantNoMetadata holds a result, its words and its payload, to no field of a
// moment, an age or a version, no date with its time, and not the version of
// the instructions itself.
func wantNoMetadata(t *testing.T, name string, result *mcp.CallToolResult, version string) {
	t.Helper()

	payload, err := json.Marshal(result.StructuredContent)
	if err != nil {
		t.Fatalf("%s: the payload is no JSON: %v", name, err)
	}
	for _, said := range []string{textOf(t, result), string(payload)} {
		for _, metadata := range []string{"answered_at", "age_seconds", "seconds ago", "instructions_version", version} {
			if strings.Contains(said, metadata) {
				t.Errorf("%s says %q, want nothing of it", name, metadata)
			}
		}
		if moment := aMoment.FindString(said); moment != "" {
			t.Errorf("%s says the moment %s, want none", name, moment)
		}
	}
}

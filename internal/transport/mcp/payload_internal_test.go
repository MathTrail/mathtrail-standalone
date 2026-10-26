package mcpserver

import (
	"os"
	"regexp"
	"strconv"
	"testing"

	"github.com/MathTrail/mathtrail-standalone/content"
	"github.com/MathTrail/mathtrail-standalone/internal/domain/profile"
)

// A trap is named in the words for the model as the catalog describes it, an
// item of a list; one the catalog does not have keeps its id rather than
// leaving a blank where its name would be.
func TestATrapIsDescribedAsTheCatalogHasIt(t *testing.T) {
	t.Parallel()

	loaded, err := content.Load()
	if err != nil {
		t.Fatalf("content.Load() error = %v", err)
	}
	s := &Service{content: loaded}

	for _, tc := range []struct{ id, want string }{
		{"double_count", "Counted the same option twice"},
		{"counted_the_cat", "counted_the_cat"},
	} {
		if got := s.trapDescribed(tc.id); got != tc.want {
			t.Errorf("trapDescribed(%q) = %q, want %q", tc.id, got, tc.want)
		}
	}
}

// cardsStep is how the card's line for the model words an explanation, as the
// widget writes it: one double-quoted string after the name, on the line below.
var cardsStep = regexp.MustCompile(`const aboutTheStep =\s*"((?:[^"\\]|\\.)*)";`)

// The card tells the model how to word an explanation in the same words the
// service does: the card's line comes after an answer given on it, the
// service's after one recorded through the chat, and the model is not to hear
// two rules for one thing.
func TestTheCardWordsAnExplanationAsTheServiceDoes(t *testing.T) {
	t.Parallel()

	written, err := os.ReadFile("../../../web/src/widget/lesson.ts")
	if err != nil {
		t.Fatalf("read the card's lesson: %v", err)
	}
	found := cardsStep.FindSubmatch(written)
	if found == nil {
		t.Fatalf("the card's lesson has no aboutTheStep written as one string")
	}
	card, err := strconv.Unquote(`"` + string(found[1]) + `"`)
	if err != nil {
		t.Fatalf("read the card's aboutTheStep: %v", err)
	}
	if card != aboutTheStep {
		t.Errorf("the card words an explanation:\n%s\nwant the service's words:\n%s", card, aboutTheStep)
	}
}

// cardsMarks are how the card's check begins the ids of a task and of the
// request it was written for.
var cardsMarks = regexp.MustCompile(`const taskMark = "([^"]*)";\s*const requestMark = "([^"]*)";`)

// The card names the request of its task by the rule the service names a task
// after its request: a card drawn again asks how that request stands before it
// asks for another task, and a request the service does not know would leave
// it asking after nothing.
func TestTheCardNamesATasksRequestAsTheServiceDoes(t *testing.T) {
	t.Parallel()

	written, err := os.ReadFile("../../../web/src/widget/stillOn.ts")
	if err != nil {
		t.Fatalf("read the card's check: %v", err)
	}
	found := cardsMarks.FindSubmatch(written)
	if found == nil {
		t.Fatalf("the card's check names no taskMark and requestMark")
	}
	taskMark, requestMark := string(found[1]), string(found[2])
	if got := profile.TaskIDFor(requestMark + "0b5f"); got != taskMark+"0b5f" {
		t.Errorf("TaskIDFor(%q) = %q, want the card's %q", requestMark+"0b5f", got, taskMark+"0b5f")
	}
	if got := profile.RequestIDFor(taskMark + "0b5f"); got != requestMark+"0b5f" {
		t.Errorf("RequestIDFor(%q) = %q, want the card's %q", taskMark+"0b5f", got, requestMark+"0b5f")
	}
}

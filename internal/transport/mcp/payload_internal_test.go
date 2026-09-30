package mcpserver

import (
	"testing"

	"github.com/MathTrail/mathtrail-standalone/content"
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

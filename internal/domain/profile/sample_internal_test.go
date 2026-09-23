package profile

import (
	"os"
	"path/filepath"
	"testing"
)

// sampleProfile is a profile the internal tests can take apart: the fixture of
// a child with a short history, read through the reader it is about to be
// read by again.
func sampleProfile(t *testing.T) *Profile {
	t.Helper()

	raw, err := os.ReadFile(filepath.Join("..", "..", "..", "testdata", "profiles", "dima.json"))
	if err != nil {
		t.Fatalf("read the fixture: %v", err)
	}
	p, err := Parse(raw)
	if err != nil {
		t.Fatalf("Parse() error = %v, want nil", err)
	}
	p.Student.Pseudonym = "Otter"
	return p
}

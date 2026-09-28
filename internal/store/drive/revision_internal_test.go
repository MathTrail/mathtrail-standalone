package drivestore

import (
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"

	"github.com/MathTrail/mathtrail-standalone/internal/store"
)

// A revision is read back as it was made, whatever the file's ID holds —
// separators included — whatever the profile's number, and whatever the bytes;
// and bytes that differ make revisions that differ.
func TestARevisionHoldsItsProperties(t *testing.T) {
	t.Parallel()

	properties := gopter.NewProperties(nil)
	file := gen.AnyString().SuchThat(func(id string) bool { return id != "" })
	counter := gen.IntRange(0, 1<<30)
	content := gen.SliceOf(gen.UInt8())

	properties.Property("a revision reads back as it was made", prop.ForAll(
		func(id string, number int, raw []byte) bool {
			read, readable := parseRevision(revisionOf(id, number, raw))
			return readable && read == revision{file: id, counter: number, digest: digestOf(raw)}
		},
		file, counter, content,
	))
	properties.Property("bytes that differ make revisions that differ", prop.ForAll(
		func(id string, number int, raw []byte, extra uint8) bool {
			return revisionOf(id, number, raw) != revisionOf(id, number, append(raw, extra))
		},
		file, counter, content, gen.UInt8(),
	))
	properties.TestingRun(t)
}

// What this store did not hand out is not a revision of its.
func TestWhatIsNoRevisionIsNotReadAsOne(t *testing.T) {
	t.Parallel()

	for _, text := range []store.Revision{
		"",
		"just-a-file",
		"a-file/7",
		"/7/digest",
		"a-file/seven/digest",
		"a-file/-1/digest",
		"a-file/7/",
		"42", // what the store in memory hands out
	} {
		if read, readable := parseRevision(text); readable {
			t.Errorf("parseRevision(%q) = %+v, want no revision", text, read)
		}
	}
}

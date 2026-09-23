package profile_test

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
	"unicode"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/profile"
)

// The properties here name what has to hold for every child rather than for
// the five in the fixtures. Two things in this package have that shape: the
// file, which must come back as it went in whatever the parent typed into it,
// and the sealed block, which must leave no trace of itself anywhere else —
// the whole point of sealing is that the rest of the file says nothing about
// what is inside it.

// genText produces what a parent might type: any script, any length, with the
// control characters and the length the file refuses taken out, because a
// profile past a cap is refused rather than written.
func genText(limit int) gopter.Gen {
	return gen.AnyString().Map(func(text string) string {
		text = strings.Map(func(r rune) rune {
			if unicode.IsControl(r) {
				return -1
			}
			return r
		}, text)
		if runes := []rune(text); len(runes) > limit {
			text = string(runes[:limit])
		}
		return text
	})
}

func TestTheFileHoldsItsProperties(t *testing.T) {
	t.Parallel()

	properties := gopter.NewProperties(nil)

	properties.Property("what a parent typed comes back as they typed it", prop.ForAll(
		func(pseudonym, notes, interest string) bool {
			if pseudonym == "" || interest == "" {
				return true // a profile with neither is refused, and says so
			}
			p := parseFixture(t, "dima")
			p.Student.Pseudonym, p.Student.Notes = pseudonym, notes
			p.Student.Interests = []string{interest}

			written, err := profile.Marshal(p)
			if err != nil {
				return false
			}
			read, err := profile.Parse(written)
			if err != nil {
				return false
			}
			return read.Student.Pseudonym == pseudonym &&
				read.Student.Notes == notes &&
				len(read.Student.Interests) == 1 && read.Student.Interests[0] == interest
		},
		genText(profile.MaxPseudonym), genText(profile.MaxNotes), genText(profile.MaxInterest),
	))

	properties.Property("every object is written with its keys in order", prop.ForAll(
		func(pseudonym string, trap string, times int) bool {
			if pseudonym == "" || trap == "" {
				return true
			}
			p := parseFixture(t, "dima")
			p.Student.Pseudonym = pseudonym
			topic := p.Topics["counting.gaps"]
			topic.Traps[trap] = times
			topic.Answers += times
			p.Topics["counting.gaps"] = topic

			written, err := profile.Marshal(p)
			if err != nil {
				return false
			}
			checkSorted(t, "", written)
			return !t.Failed()
		},
		genText(profile.MaxPseudonym), gen.Identifier(), gen.IntRange(1, 9),
	))

	properties.TestingRun(t)
}

// Whatever is sealed, nothing else in the file moves with it. A file whose
// open part changed shape with the answer — a longer wording for a longer
// explanation, a field that appears only sometimes — would give away by its
// outline what the sealing is there to hide.
func TestNothingOutsideTheSealedBlockFollowsIt(t *testing.T) {
	t.Parallel()

	properties := gopter.NewProperties(nil)

	// Two generators rather than one used twice, so that each argument reads
	// as the sealed value it stands for.
	before, after := genText(200), genText(200)

	properties.Property("the file differs only where the sealed value does", prop.ForAll(
		func(first, second string) bool {
			// The prefix is what a sealed value always carries, and it keeps a
			// generated one from colliding with a word elsewhere in the file.
			first, second = "mt1.t.kQ7fWx."+first, "mt1.t.kQ7fWx."+second
			if first == second {
				return true
			}

			written := make([][]byte, 2)
			for i, sealed := range []string{first, second} {
				p := parseFixture(t, "masha")
				p.CurrentTask.Sealed = sealed
				out, err := profile.Marshal(p)
				if err != nil {
					return false
				}
				written[i] = out
			}

			// Put the second sealed value where the first one was: if nothing
			// else followed it, the two files are now the same file.
			was, now := asWritten(first), asWritten(second)
			return bytes.Equal(bytes.Replace(written[0], was, now, 1), written[1])
		},
		before, after,
	))

	properties.TestingRun(t)
}

// asWritten is a string as the profile writes it. The profile leaves the
// characters of HTML alone and encoding/json does not, so a value read back
// out of a written file has to be looked for in the spelling the file uses.
func asWritten(text string) []byte {
	var out bytes.Buffer
	encoder := json.NewEncoder(&out)
	encoder.SetEscapeHTML(false)
	if err := encoder.Encode(text); err != nil {
		return nil
	}
	return bytes.TrimRight(out.Bytes(), "\n")
}

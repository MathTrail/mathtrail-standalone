package checks_test

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/checks"
)

// genLines are the lines of a lattice, each from one corner of a row of cells
// to the corner beside it, at the places given.
func genLines() gopter.Gen {
	return gen.SliceOfN(12, gen.IntRange(0, 7)).Map(func(places []int) []string {
		lines := make([]string, len(places))
		for i, at := range places {
			lines[i] = fmt.Sprintf("[%d,%d,%d,%d]", at, i%8, at+1, i%8)
		}
		return lines
	})
}

// latticeOf is a lattice drawing lines, its members and lines written in the
// order given.
func latticeOf(lines []string, membersFirst bool) json.RawMessage {
	drawn := `"lines":[` + strings.Join(lines, ",") + `]`
	if membersFirst {
		return json.RawMessage(`{"kind":"lattice","width":8,"height":8,` + drawn + `}`)
	}
	return json.RawMessage(`{` + drawn + `,"height":8,"width":8,"kind":"lattice"}`)
}

// reversed are some lines written the other way round in the list.
func reversed(lines []string) []string {
	turned := make([]string, len(lines))
	for i, line := range lines {
		turned[len(lines)-1-i] = line
	}
	return turned
}

// The sketch of a picture names what it draws, not how its description is
// written: the order of its members and of the items of its lists change
// nothing, so that a picture always repeats itself.
func TestThePictureOfATaskRepeatsItselfHoweverItIsWritten(t *testing.T) {
	t.Parallel()

	properties := gopter.NewProperties(nil)
	properties.Property("a picture written in another order has the same sketch", prop.ForAll(
		func(lines []string) bool {
			return checks.PictureFingerprint(latticeOf(lines, true)) ==
				checks.PictureFingerprint(latticeOf(reversed(lines), false))
		},
		genLines(),
	))
	properties.Property("a picture repeats a past one it is", prop.ForAll(
		func(lines []string) bool {
			past := []string{checks.PictureFingerprint(latticeOf(lines, true))}
			found := checks.RepeatsPicture(latticeOf(reversed(lines), false), past)
			return len(found) == 1 && found[0].NearDuplicateOf == checks.EarlierPastPicture
		},
		genLines(),
	))
	properties.TestingRun(t)
}

// A picture comes from the chat's model and is untrusted: whatever it holds,
// its sketch is none or one the profile can keep, and a picture with one
// repeats itself.
func FuzzPictureFingerprint(f *testing.F) {
	for _, seed := range []string{
		`{"kind":"lattice","width":3,"height":3,"lines":[[1,0,1,2]],"marks":[{"at":[0,0],"label":"S"}]}`,
		`{"kind":"tiles","sides":5,"layout":"ring","tiles":[["1","","","","?"]]}`,
		`{"kind":"stairs","steps":8,"marks":[{"step":0,"label":"A"}],"jumps":[{"from":0,"to":2}]}`,
		`{"a":{"b":{"c":[1e400,null,{"d":[]}]}}}`, `{"kind":"\u0000"}`, `[]`, `null`, `{`, ``,
	} {
		f.Add([]byte(seed))
	}
	f.Fuzz(func(t *testing.T, raw []byte) {
		sketch := checks.PictureFingerprint(raw)
		if sketch == "" {
			return
		}
		if len(sketch) != 128 {
			t.Fatalf("PictureFingerprint(%q) is %d characters, want 128", raw, len(sketch))
		}
		if found := checks.RepeatsPicture(raw, []string{sketch}); len(found) != 1 {
			t.Fatalf("RepeatsPicture(%q, its own sketch) = %v, want one repeat", raw, found)
		}
	})
}

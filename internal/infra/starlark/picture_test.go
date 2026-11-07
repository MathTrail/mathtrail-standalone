package starlark_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"maps"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/solver"
	"github.com/MathTrail/mathtrail-standalone/internal/infra/starlark"
)

// fence is a picture as a model writes one: a list of numbers, a list of
// cells, a list of objects, a label and a flag.
const fence = `{"kind":"row","lines":[1,2,3],"cells":[[2,3],[0,1]],"items":[{"label":"A"},{}],"gaps":"?","line":true}`

// runWith runs one program with a picture against the five options of these
// tests.
func runWith(t *testing.T, source, picture string) solver.Result {
	t.Helper()
	var description json.RawMessage
	if picture != "" {
		description = json.RawMessage(picture)
	}
	result, err := sandbox(t, limits()).Run(context.Background(), source, options, description)
	if err != nil {
		t.Fatalf("Run: got error %v, want none", err)
	}
	return result
}

// The picture reaches a solve of two arguments as the model wrote it, frozen:
// an object as a dictionary, a list of numbers or cells as a tuple, a list of
// objects as a list of dictionaries, a whole number as an integer, and no
// picture, or one that is no object, as None. Each program answers 5, option
// B, only when the picture reads as it should.
func TestTheSolverIsHandedThePicture(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		name, picture, check string
	}{
		{"no picture is None", "", `picture == None`},
		{"a description that is no object is None", `[1, 2]`, `picture == None`},
		{"the description reads as a dictionary of its own type", fence, `type(picture) == "picture" and picture["kind"] == "row"`},
		{"a list of numbers is a tuple", fence, `picture["lines"] == (1, 2, 3)`},
		{"a cell can stand in a set", fence, `(2, 3) in set(picture["cells"])`},
		{"a list of objects is a list of dictionaries", fence,
			`type(picture["items"]) == "list" and picture["items"][0]["label"] == "A"`},
		{"a label is a text", fence, `picture["gaps"] == "?"`},
		{"a flag is a flag", fence, `picture["line"] == True`},
		{"a whole number written with a point is whole", `{"steps":3.0}`, `picture["steps"] == 3 and type(picture["steps"]) == "int"`},
		{"a number with a fraction is a float", `{"at":2.5}`, `picture["at"] == 2.5`},
		{"a number too large for a float is infinite", `{"copies":1e400}`, `picture["copies"] > 1e300`},
		{"an integer past 64 bits is a float", `{"n":123456789012345678901234567890}`, `picture["n"] > 1e29`},
		{"the capped built-ins can measure it", fence, `len(picture) == 6 and sorted(picture)[0] == "cells" and len(list(picture["lines"])) == 3`},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			source := fmt.Sprintf("def solve(options, picture):\n    return match(options, 5 if %s else 4)\n", test.check)
			result := runWith(t, source, test.picture)
			if !result.One() || result.Letters[0] != "B" {
				t.Errorf("Run = %s %v %q, want B: %s", result.Status, result.Letters, result.Message, test.check)
			}
		})
	}
}

// A program reads the picture and does not get to edit it, at its top or
// anywhere inside.
func TestThePictureCannotBeEdited(t *testing.T) {
	t.Parallel()

	for _, edit := range []string{
		`picture["gaps"] = "1"`,
		`picture["items"].append({})`,
		`picture["items"][0]["label"] = "B"`,
	} {
		t.Run(edit, func(t *testing.T) {
			t.Parallel()
			source := "def solve(options, picture):\n    " + edit + "\n    return match(options, 5)\n"
			if result := runWith(t, source, fence); result.Status != solver.StatusError {
				t.Errorf("Run = %s %v, want %s: the picture is frozen", result.Status, result.Letters, solver.StatusError)
			}
		})
	}
}

// A program that holds the picture up against a dictionary of its own is
// answered as any program is: the two are never equal, and an order between
// them is the program's error, never one inside the interpreter.
func TestThePictureIsNoDictionaryOfTheProgramsOwn(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		name, check string
		status      solver.Status
	}{
		{"compared with a dictionary", `not ({} == picture) and {} != picture`, solver.StatusOK},
		{"looked for among dictionaries", `not (picture in [{}])`, solver.StatusOK},
		{"compared inside a list", `[{}] != [picture]`, solver.StatusOK},
		{"sorted among dictionaries", `len(sorted([{}, picture])) == 2`, solver.StatusError},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			source := fmt.Sprintf("def solve(options, picture):\n    return match(options, 5 if %s else 4)\n", test.check)
			result := runWith(t, source, fence)
			if result.Status != test.status || (test.status == solver.StatusOK && !slices.Equal(result.Letters, []string{"B"})) {
				t.Errorf("Run = %s %v %q, want %s: %s", result.Status, result.Letters, result.Message, test.status, test.check)
			}
		})
	}
}

// What a run read is the members it took out of the description's own object,
// whichever way it took them, and none it only asked about.
func TestWhatARunReadsOfThePicture(t *testing.T) {
	t.Parallel()

	every := []string{"cells", "gaps", "items", "kind", "line", "lines"}
	for _, test := range []struct {
		name, body string
		want       []string
	}{
		{"one member by its name", `picture["lines"]`, []string{"lines"}},
		{"one member with get", `picture.get("lines")`, []string{"lines"}},
		{"a member inside a member", `picture["items"][0]["label"]`, []string{"items"}},
		{"a member the picture does not have", `picture.get("walls", ())`, nil},
		{"whether a member is there", `"lines" in picture`, nil},
		{"the names of the members", `picture.keys()`, nil},
		{"a walk over the names", `[name for name in picture]`, nil},
		{"how many members there are", `len(picture)`, nil},
		{"every member with items", `picture.items()`, every},
		{"every member with values", `picture.values()`, every},
		{"every member with dict", `dict(picture)`, every},
		{"every member by its name", `[picture[name] for name in picture.keys()]`, every},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			source := "def solve(options, picture):\n    seen = " + test.body + "\n    return match(options, 5)\n"
			result := runWith(t, source, fence)
			if result.Status != solver.StatusOK || !slices.Equal(result.Read, test.want) {
				t.Errorf("Run = %s, read %q, want %s and %q", result.Status, result.Read, solver.StatusOK, test.want)
			}
		})
	}
}

// Only a run that ran to the end says what it read: one that failed may have
// stopped before it read what it would have, and one handed nothing read
// nothing.
func TestOnlyARunThatEndedSaysWhatItRead(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		name, source, picture string
	}{
		{"a run that failed after reading", "def solve(options, picture):\n    seen = picture[\"lines\"]\n    fail(\"no\")\n", fence},
		{"a run of a solve that takes no picture", "def solve(options):\n    return match(options, 5)\n", fence},
		{"a run handed no picture", "def solve(options, picture):\n    return match(options, 5)\n", ""},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			if result := runWith(t, test.source, test.picture); result.Read != nil {
				t.Errorf("Run = %s, read %q, want nothing read", result.Status, result.Read)
			}
		})
	}
}

// Two runs on one sandbox each say what they themselves read, which is what
// lets a verdict hold either of them to the picture.
func TestEachRunSaysWhatItRead(t *testing.T) {
	t.Parallel()

	runner := sandbox(t, limits())
	source := "def solve(options, picture):\n" +
		"    seen = picture[\"lines\"] if options[\"A\"] == \"4\" else picture[\"cells\"]\n" +
		"    return match(options, 5)\n"
	rotated := options.Rotated()
	for _, test := range []struct {
		options solver.Options
		want    []string
	}{
		{options, []string{"lines"}},
		{rotated, []string{"cells"}},
	} {
		result, err := runner.Run(context.Background(), source, test.options, json.RawMessage(fence))
		if err != nil || !slices.Equal(result.Read, test.want) {
			t.Errorf("Run(%v) = %v, read %q, want %q", test.options, err, result.Read, test.want)
		}
	}
}

// Handing a program the picture costs it nothing: the picture is built before
// the run starts, and reading it is no step of the interpreter's.
func TestThePictureCostsNoSteps(t *testing.T) {
	t.Parallel()

	members := make([]string, 0, 200)
	for i := range 200 {
		members = append(members, fmt.Sprintf(`"m%d":[[%d,%d],[%d,%d]]`, i, i, i, i, i))
	}
	large := "{" + strings.Join(members, ",") + "}"

	without := runWith(t, "def solve(options):\n    return match(options, 5)\n", large)
	with := runWith(t, "def solve(options, picture):\n    return match(options, 5)\n", large)
	if without.Status != solver.StatusOK || with.Steps != without.Steps {
		t.Errorf("steps: got %d with the picture and %d without, want the same", with.Steps, without.Steps)
	}
}

// Whatever the model wrote as a picture, a run is answered, not refused: a
// description out of the format is the format's to refuse, and a failure here
// would be the service's. What a run says it read are members the description
// has.
func FuzzPicture(f *testing.F) {
	for _, seed := range []string{
		fence,
		`{"kind":"row","items":[{"skip":true,"label":"\u202e"}],"line":null,"copies":1e400}`,
		`{"kind":"bars","bars":[{"parts":1e9,"segments":[{"size":-1}],"shaded":"3"}]}`,
		`{"kind":"grid","rows":["A","B"],"cols":["1","2"],"filled":["A1"],"marks":{"B2":"X"}}`,
		`[{"kind":"clock"}]`, `{`, `null`, `"?"`, `{"a":{"b":{"c":[[[]]]}}}`, `{"":0,"\u0000":-0.0}`,
	} {
		f.Add([]byte(seed))
	}

	runner, err := starlark.New(starlark.Limits{Steps: 100_000, Timeout: time.Second, Concurrency: 2, Wait: time.Minute})
	if err != nil {
		f.Fatalf("New: %v", err)
	}
	const reader = "def solve(options, picture):\n" +
		"    if type(picture) == \"dict\":\n" +
		"        seen = [str(value) for value in picture.values()]\n" +
		"    return match(options, 5)\n"

	f.Fuzz(func(t *testing.T, raw []byte) {
		result, err := runner.Run(context.Background(), reader, options, raw)
		if err != nil {
			t.Fatalf("Run: got error %v, want none", err)
		}
		var members map[string]any
		decoder := json.NewDecoder(bytes.NewReader(raw))
		decoder.UseNumber()
		if decoder.Decode(&members) != nil {
			members = nil
		}
		for _, name := range result.Read {
			if _, held := members[name]; !held {
				t.Fatalf("read %q, which %q has no member called", name, raw)
			}
		}
	})
}

// Whatever members a picture has, a program that reads each of them says it
// read each, and one that reads none says it read none: what a run says is
// what it did, the property a solver of a task solved from the picture is
// judged by.
func TestWhatARunSaysItReadIsWhatItRead(t *testing.T) {
	t.Parallel()

	runner := sandbox(t, limits())
	runOn := func(source string, members map[string]int) []string {
		raw, err := json.Marshal(members)
		if err != nil {
			t.Fatalf("marshal %v: %v", members, err)
		}
		result, err := runner.Run(context.Background(), source, options, raw)
		if err != nil || result.Status != solver.StatusOK {
			t.Fatalf("Run on %s = %v %s %q", raw, err, result.Status, result.Message)
		}
		return result.Read
	}
	const readsEach = "def solve(options, picture):\n" +
		"    for name in picture.keys():\n" +
		"        seen = picture[name]\n" +
		"    return match(options, 5)\n"
	const readsNone = "def solve(options, picture):\n    return match(options, len(picture) * 0 + 5)\n"

	properties := gopter.NewProperties(nil)
	properties.Property("a program that reads every member says it read every one", prop.ForAll(
		func(members map[string]int) bool {
			return slices.Equal(runOn(readsEach, members), slices.Sorted(maps.Keys(members)))
		},
		gen.MapOf(gen.Identifier(), gen.Int()),
	))
	properties.Property("a program that reads no member says it read none", prop.ForAll(
		func(members map[string]int) bool { return runOn(readsNone, members) == nil },
		gen.MapOf(gen.Identifier(), gen.Int()),
	))
	properties.TestingRun(t)
}

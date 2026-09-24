package starlark_test

import (
	"strings"
	"testing"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/solver"
	"github.com/MathTrail/mathtrail-standalone/internal/infra/starlark"
)

// A run that stops at a string Starlark cannot fill with % is told apart from
// any other failure, so that the model hears which rule of the dialect it
// broke; a stray % the run never reaches costs nothing, since the program still
// proves its answer; and a run that stops for another reason is not blamed on
// a format elsewhere in it.
func TestARunThatStopsAtAFormatStarlarkCannotFillSaysSo(t *testing.T) {
	t.Parallel()
	for _, test := range []runCase{
		{
			name:     "a stray percent sign",
			source:   "def solve(options):\n    note = \"10% of %d\" % 6\n    return match(options, 6)\n",
			status:   solver.StatusBadFormat,
			contains: "line 2: a string formatted with %",
		},
		{
			name:     "a precision, which Python has and Starlark lacks",
			source:   "def solve(options):\n    return match(options, \"%.1f\" % 6.0)\n",
			status:   solver.StatusBadFormat,
			contains: "which Starlark cannot fill",
		},
		{
			name:     "a width",
			source:   "def solve(options):\n    return match(options, \"%3d\" % 6)\n",
			status:   solver.StatusBadFormat,
			contains: "which Starlark cannot fill",
		},
		{
			name:     "fewer values than the string asks for",
			source:   "def solve(options):\n    note = \"%d of %d\" % (6,)\n    return match(options, 6)\n",
			status:   solver.StatusBadFormat,
			contains: "asks for 2 values and is given 1",
		},
		{
			name:     "more values than the string asks for",
			source:   "def solve(options):\n    note = \"%d\" % (6, 7)\n    return match(options, 6)\n",
			status:   solver.StatusBadFormat,
			contains: "asks for 1 values and is given 2",
		},
		{
			name:     "a string kept in a constant",
			source:   "RISE = \"a rise of 10% on %d\"\n\ndef solve(options):\n    return match(options, len(RISE % 6))\n",
			status:   solver.StatusBadFormat,
			contains: "line 4: a string formatted with %",
		},
		{
			name: "a string kept in a constant among other names",
			source: "limit = 6\nSIZES = [1, 2]\nRISE = \"a rise of 10% on %d\"\n\n" +
				"def solve(options):\n    return match(options, len(RISE % limit))\n",
			status:   solver.StatusBadFormat,
			contains: "line 6: a string formatted with %",
		},
		{
			name:     "a string kept in a name in lower case",
			source:   "note = \"10% of %d\"\n\ndef solve(options):\n    return match(options, len(note % 6))\n",
			status:   solver.StatusBadFormat,
			contains: "line 4: a string formatted with %",
		},
		{
			name:     "a stray percent sign at the top of the program",
			source:   "NOTE = \"10% of %d\" % 6\n\ndef solve(options):\n    return match(options, 6)\n",
			status:   solver.StatusBadFormat,
			contains: "line 1: a string formatted with %",
		},
		{
			name:     "fewer values than the string asks for, from a built-in",
			source:   "def solve(options):\n    note = \"%d of %d\" % len(options)\n    return match(options, 6)\n",
			status:   solver.StatusBadFormat,
			contains: "asks for 2 values and is given 1",
		},
		{
			name:    "a stray percent sign the run never reaches",
			source:  "def solve(options):\n    if len(options) != 5:\n        fail(\"a rise of 10% on %d\" % 6)\n    return match(options, 6)\n",
			status:  solver.StatusOK,
			letters: []string{"C"},
		},
		{
			name:     "another failure, with a stray percent sign elsewhere",
			source:   "def solve(options):\n    if len(options) != 5:\n        fail(\"a rise of 10% on %d\" % 6)\n    return [1, 2][5]\n",
			status:   solver.StatusError,
			contains: "out of range",
		},
		{
			name:     "a string put together while the program runs",
			source:   "def solve(options):\n    text = \"10\" + \"% of %d\"\n    return match(options, len(text % 6))\n",
			status:   solver.StatusError,
			contains: "unknown conversion",
		},
		{
			name:     "a local that shadows a constant of the same name",
			source:   "MSG = \"10% of %d\"\n\ndef solve(options):\n    MSG = \"%d\"\n    note = MSG % \"six\"\n    return match(options, 6)\n",
			status:   solver.StatusError,
			contains: "format requires",
		},
		{
			name:     "a function of the program's own named after a built-in",
			source:   "def len(x):\n    return \"six\"\n\ndef solve(options):\n    note = \"%d of %d\" % len(options)\n    return match(options, 6)\n",
			status:   solver.StatusError,
			contains: "format requires",
		},
		{
			name:     "a name bound again by a function of the same name",
			source:   "MSG = \"10% of %d\"\n\ndef MSG():\n    return 1\n\ndef solve(options):\n    note = MSG % 6\n    return match(options, 6)\n",
			status:   solver.StatusError,
			contains: "unknown binary op",
		},
		{
			name:    "a percent sign written twice",
			source:  "def solve(options):\n    note = \"%d%%\" % 6\n    return match(options, 6)\n",
			status:  solver.StatusOK,
			letters: []string{"C"},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			test.check(t, run(t, test.source))
		})
	}
}

// A program others start from is read throughout: a stray % is found whether a
// run would reach it or not, where a run is told only about the one it stopped
// at. Its names are read as a run reads them, so a local that shadows a
// constant, or a function of the program's own named after a built-in, is not
// taken for the other; and a name bound more than once holds no one string, so
// it is no constant.
func TestUnfillableFormatReadsTheWholeProgram(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		name, source, want string
	}{
		{
			name:   "a stray percent sign a run never reaches",
			source: "def solve(options):\n    if len(options) != 5:\n        fail(\"a rise of 10% on %d\" % 6)\n    return match(options, 6)\n",
			want:   "line 3: a string formatted with %",
		},
		{
			name:   "every format one Starlark fills",
			source: "def solve(options):\n    return match(options, len(\"%d%%\" % 6))\n",
		},
		{
			name:   "a string kept in a name in lower case",
			source: "note = \"a rise of 10% on %d\"\n\ndef solve(options):\n    if len(options) != 5:\n        fail(note % 6)\n    return match(options, 6)\n",
			want:   "line 5: a string formatted with %",
		},
		{
			name:   "a local that shadows a constant of the same name",
			source: "MSG = \"10% of %d\"\n\ndef solve(options):\n    MSG = \"%d\"\n    note = MSG % 6\n    return match(options, 6)\n",
		},
		{
			name: "a constant a function's local of the same name leaves alone",
			source: "MSG = \"10% of %d\"\n\ndef other():\n    MSG = \"%d\"\n    return MSG\n\n" +
				"def solve(options):\n    if len(options) != 5:\n        fail(MSG % 6)\n    return match(options, 6)\n",
			want: "line 9: a string formatted with %",
		},
		{
			name:   "a constant assigned in brackets",
			source: "(MSG) = \"10% of %d\"\n\ndef solve(options):\n    if len(options) != 5:\n        fail(MSG % 6)\n    return match(options, 6)\n",
			want:   "line 5: a string formatted with %",
		},
		{
			name:   "a function of the program's own named after a built-in",
			source: "def len(x):\n    return (1, 2)\n\ndef solve(options):\n    note = \"%d of %d\" % len(options)\n    return match(options, 6)\n",
		},
		{
			name:   "a string made longer with +=",
			source: "MSG = \"%d\"\nMSG += \" of %d\"\nNOTE = MSG % (1, 2)\n\ndef solve(options):\n    return match(options, 6)\n",
		},
		{
			name:   "a name bound again after the string it held was formatted",
			source: "MSG = \"%d\"\nNOTE = MSG % 6\nMSG = \"10% of %d\"\n\ndef solve(options):\n    return match(options, 6)\n",
		},
		{
			name:   "a name bound again by unpacking a tuple",
			source: "MSG = \"10% of %d\"\n(MSG, N) = (\"%d\", 2)\nNOTE = MSG % 6\n\ndef solve(options):\n    return match(options, 6)\n",
		},
		{
			name:   "a name bound again by unpacking a list",
			source: "MSG = \"10% of %d\"\n[MSG, N] = [\"%d\", 2]\nNOTE = MSG % 6\n\ndef solve(options):\n    return match(options, 6)\n",
		},
		{
			name:   "a name bound again as a loop's variable",
			source: "MSG = \"10% of %d\"\nfor MSG in [\"%d\"]:\n    NOTE = MSG % 6\n\ndef solve(options):\n    return match(options, 6)\n",
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			got, err := starlark.UnfillableFormat("solver.star", test.source)
			if err != nil {
				t.Fatalf("UnfillableFormat: got error %v, want none", err)
			}
			if (test.want == "") != (got == "") || !strings.Contains(got, test.want) {
				t.Errorf("UnfillableFormat() = %q, want %q", got, test.want)
			}
		})
	}
}

// A program a run could not read is refused here too, rather than checked with
// names nobody could tell apart, and the refusal names the program once.
func TestUnfillableFormatRefusesAProgramARunCouldNotRead(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		name, source string
	}{
		{"a source longer than a solver may be", "# " + strings.Repeat("x", starlark.MaxSource) + "\n"},
		{"a source that does not parse", "def solve(options)\n    return []\n"},
		{"a name the program never defines", "def solve(options):\n    return match(options, answer)\n"},
		{"a load", "load('helpers.star', 'helper')\n\ndef solve(options):\n    return []\n"},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			const program = "examples/solvers/sample.star"
			_, err := starlark.UnfillableFormat(program, test.source)
			if err == nil {
				t.Fatal("UnfillableFormat: got no error, want one")
			}
			if named := strings.Count(err.Error(), program); named != 1 {
				t.Errorf("UnfillableFormat: error %q names the program %d times, want once", err, named)
			}
		})
	}
}

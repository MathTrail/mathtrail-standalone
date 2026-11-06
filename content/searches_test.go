package content_test

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/MathTrail/mathtrail-standalone/internal/config"
	"github.com/MathTrail/mathtrail-standalone/internal/domain/solver"
)

// Two searches of tasks solved from the picture outrun the limits when written
// the naive way: four shapes each tried from 0 to 20 spend more than the steps
// a run has, and ten sticks each placed on one of four sides ask product for
// more than a million placements. These are the searches that do not: each,
// at the largest picture its genre draws, proves its answer on both runs
// within a quarter of the steps, reading the members of the picture that
// carry its facts.
func TestASearchOfAPictureProvesItsAnswerWithinTheLimits(t *testing.T) {
	t.Parallel()

	sandbox := serviceSandbox(t)
	for _, test := range []search{
		{
			name:    "five rows of six shapes, all four of them",
			program: "shapes.star",
			picture: `{"kind":"shapes","rows":[` +
				`{"shapes":["triangle","triangle","triangle","square","square","square"],"total":"33"},` +
				`{"shapes":["square","square","circle","circle","circle","circle"],"total":"44"},` +
				`{"shapes":["circle","diamond","diamond","diamond","diamond","diamond"],"total":"19"},` +
				`{"shapes":["triangle","triangle","square","square","diamond","diamond"],"total":"26"},` +
				`{"shapes":["triangle","square","circle","diamond","diamond","diamond"],"total":"?"}]}`,
			options: solver.Options{"24", "26", "28", "30", "It cannot be told"},
			answer:  "B", reads: []string{"rows"},
		},
		{
			name:    "rows that each hold two shapes, which give nothing away at first",
			program: "shapes.star",
			picture: `{"kind":"shapes","rows":[` +
				`{"shapes":["triangle","square","square","square","square","square"],"total":"27"},` +
				`{"shapes":["square","circle","circle","circle","circle","circle"],"total":"49"},` +
				`{"shapes":["circle","diamond","diamond","diamond","diamond","diamond"],"total":"19"},` +
				`{"shapes":["diamond","triangle","triangle","triangle","triangle","triangle"],"total":"37"},` +
				`{"shapes":["triangle","square","circle","diamond","diamond","diamond"],"total":"?"}]}`,
			options: solver.Options{"24", "26", "28", "30", "It cannot be told"},
			answer:  "B", reads: []string{"rows"},
		},
		{
			name:    "rows that allow two answers",
			program: "shapes.star",
			picture: `{"kind":"shapes","rows":[` +
				`{"shapes":["triangle","square"],"total":"9"},` +
				`{"shapes":["triangle"],"total":"?"}]}`,
			options: solver.Options{"4", "5", "6", "9", "It cannot be told"},
			answer:  "E", reads: []string{"rows"},
		},
		{
			name:    "ten sticks of two lengths",
			program: "sticks.star",
			picture: `{"kind":"lattice","width":8,"height":7,"lines":[[0,0,3,0],[4,0,7,0],[0,2,3,2],[4,2,7,2],` +
				`[0,4,1,4],[2,4,3,4],[4,4,5,4],[6,4,7,4],[0,6,2,6],[3,6,5,6]]}`,
			options: solver.Options{"5", "20", "3", "10", "It is impossible"},
			answer:  "A", reads: []string{"lines"},
		},
		{
			name:    "as many sticks as squared paper draws lines",
			program: "sticks.star",
			picture: sticks(24, 24),
			options: solver.Options{"12", "18", "24", "36", "It is impossible"},
			answer:  "B", reads: []string{"lines"},
		},
		{
			name:    "sticks that make no square",
			program: "sticks.star",
			picture: sticks(2, 1),
			options: solver.Options{"1", "2", "3", "4", "It is impossible"},
			answer:  "E", reads: []string{"lines"},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			test.proves(t, sandbox)
		})
	}
}

// search is one search of a picture, the picture it is run on, and what it has
// to prove of it.
type search struct {
	name, program, picture string
	options                solver.Options
	answer                 string
	reads                  []string
}

// proves runs the search as a submitted solver is run and holds both runs to
// its answer, to a quarter of the steps and to the members it has to read.
func (s *search) proves(t *testing.T, sandbox solver.Runner) {
	t.Helper()

	program, err := os.ReadFile(filepath.Join("testdata", "searches", s.program))
	if err != nil {
		t.Fatalf("read %s: %v", s.program, err)
	}
	agreement, err := solver.Verdict(t.Context(), sandbox, string(program), s.options, json.RawMessage(s.picture))
	if err != nil {
		t.Fatalf("Verdict() error = %v, want the runs to happen", err)
	}
	if agreement.Letter != s.answer {
		t.Fatalf("the search proves %q, want %s: %s", agreement.Letter, s.answer, agreement.Explain())
	}
	for i, run := range agreement.Runs {
		if run.Steps > config.DefaultSolverSteps/4 {
			t.Errorf("run %d spent %d steps, want at most a quarter of %d", i+1, run.Steps, config.DefaultSolverSteps)
		}
		if !slices.Equal(run.Read, s.reads) {
			t.Errorf("run %d read %q, want %q", i+1, run.Read, s.reads)
		}
	}
}

// A stick that does not lie along a line of the paper, or that has no length,
// is no stick the search can lay: it says so rather than count it wrong or
// divide by it.
func TestTheSticksSearchRefusesWhatIsNoStick(t *testing.T) {
	t.Parallel()

	sandbox := serviceSandbox(t)
	program, err := os.ReadFile(filepath.Join("testdata", "searches", "sticks.star"))
	if err != nil {
		t.Fatalf("read sticks.star: %v", err)
	}
	for _, line := range []string{"[0,0,1,1]", "[2,3,2,3]"} {
		picture := json.RawMessage(`{"kind":"lattice","width":4,"height":4,"lines":[[0,0,1,0],` + line + `]}`)
		result, err := sandbox.Run(t.Context(), string(program), solver.Options{"1", "2", "3", "4", "It is impossible"}, picture)
		if err != nil || result.Status != solver.StatusError || !strings.Contains(result.Message, "at least a cell long") {
			t.Errorf("Run with the line %s = %v %s %q, want the search to say it is no stick", line, err, result.Status, result.Message)
		}
	}
}

// sticks is a picture of short sticks one cell long and long ones two, laid
// along the lines of squared paper of eight by eight, up to the forty-eight
// lines such a picture draws.
func sticks(short, long int) string {
	lines := make([]string, 0, short+long)
	for i := range short {
		lines = append(lines, fmt.Sprintf("[%d,%d,%d,%d]", i%8, i/8, i%8+1, i/8))
	}
	for i := range long {
		lines = append(lines, fmt.Sprintf("[%d,%d,%d,%d]", i/4, (i%4)*2, i/4, (i%4)*2+2))
	}
	return `{"kind":"lattice","width":8,"height":8,"lines":[` + strings.Join(lines, ",") + "]}"
}

package rating_test

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/rating"
)

// The vectors in testdata/golden are the prototype's numbers, produced by its
// own code. They are the reference this package is held against, so nothing
// here ever writes to them: a test that can rewrite its own answer key proves
// nothing.
//
// One part of the file is deliberately not replayed. Its doc_example runs
// three answers with one step size for both levels — what the prototype used
// before the two were separated — and the file carries config_example for the
// same three answers under the sizes this package keeps. The corridors of both
// are checked either way, because a corridor depends on the level alone.

const (
	// tolerance is what the reference promises: four decimals, so agreement
	// past that is not claimed.
	tolerance = 1e-4

	// eloTolerance is looser for one reason: a rating is a level multiplied by
	// about 174, and the level it was computed from is printed to four
	// decimals. A hundredth of a rating point is still far below the whole
	// number anybody is shown.
	eloTolerance = 0.01
)

type goldenFile struct {
	DocExample    goldenSequence   `json:"doc_example"`
	ConfigExample goldenSequence   `json:"config_example"`
	CorridorGrid  []goldenCorridor `json:"corridor_grid"`
	SeedReplays   []goldenReplay   `json:"seed_replays"`
}

type goldenSequence struct {
	Note  string       `json:"note"`
	Steps []goldenStep `json:"steps"`
}

type goldenStep struct {
	Step          int            `json:"step"`
	Difficulty    int            `json:"difficulty"`
	Correct       *bool          `json:"correct"`
	Probability   float64        `json:"probability"`
	BetaBefore    float64        `json:"beta_before"`
	ThetaBefore   float64        `json:"theta_before"`
	ThetaAfter    float64        `json:"theta_after"`
	DeltaBefore   float64        `json:"delta_before"`
	DeltaAfter    float64        `json:"delta_after"`
	KStudent      float64        `json:"k_student"`
	KTopic        float64        `json:"k_topic"`
	CorridorAfter goldenCorridor `json:"corridor_after"`
}

type goldenCorridor struct {
	Level         float64            `json:"level"`
	Elo           float64            `json:"elo"`
	BetaMin       float64            `json:"beta_min"`
	BetaMax       float64            `json:"beta_max"`
	Probabilities map[string]float64 `json:"probabilities"`
	Inside        []int              `json:"inside"`
	Recommended   int                `json:"recommended"`
	Fit           string             `json:"fit"`
}

type goldenReplay struct {
	Student string                 `json:"student"`
	Theta   float64                `json:"theta"`
	Answers int                    `json:"answers"`
	Elo     float64                `json:"elo"`
	Topics  map[string]goldenTopic `json:"topics"`
}

type goldenTopic struct {
	Delta   float64 `json:"delta"`
	Answers int     `json:"answers"`
	Level   float64 `json:"level"`
	Elo     float64 `json:"elo"`
}

func loadGolden(t *testing.T) goldenFile {
	t.Helper()

	path := filepath.Join("..", "..", "..", "testdata", "golden", "ratings.json")
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	var golden goldenFile
	if err := json.Unmarshal(raw, &golden); err != nil {
		t.Fatalf("parse %s: %v", path, err)
	}
	return golden
}

// nearly fails when two numbers differ by more than the reference claims.
func nearly(t *testing.T, got, want, allowed float64, what string) {
	t.Helper()

	if math.Abs(got-want) > allowed {
		t.Errorf("%s = %.6f, want %.6f (off by %.2g)", what, got, want, math.Abs(got-want))
	}
}

// The three answers of the reference, replayed one at a time: a child who has
// never answered anything, one topic, and every number the reference printed
// along the way.
func TestTheReferenceSequenceReplays(t *testing.T) {
	t.Parallel()

	steps := loadGolden(t).ConfigExample.Steps
	if len(steps) == 0 {
		t.Fatal("the reference sequence has no steps, want the example the prototype printed")
	}
	state := rating.State{}
	for _, step := range steps {
		t.Run(fmt.Sprintf("step %d", step.Step), func(t *testing.T) {
			nearly(t, state.Theta, step.ThetaBefore, tolerance, "the overall level before")
			nearly(t, state.Delta, step.DeltaBefore, tolerance, "the topic before")
			nearly(t, rating.Beta(step.Difficulty), step.BetaBefore, tolerance, "the difficulty")

			// The reference has a third outcome, "I don't understand", which
			// this service does not: that is a button asking for a simpler
			// explanation, recorded beside the answer rather than instead of
			// it. It scored zero there, exactly as a wrong answer does here.
			result := rating.Update(state, rating.Beta(step.Difficulty), step.Correct != nil && *step.Correct)

			nearly(t, result.Probability, step.Probability, tolerance, "the chance of a correct answer")
			nearly(t, result.KTheta, step.KStudent, tolerance, "how far the overall level could move")
			nearly(t, result.KDelta, step.KTopic, tolerance, "how far the topic could move")
			nearly(t, result.Theta, step.ThetaAfter, tolerance, "the overall level after")
			nearly(t, result.Delta, step.DeltaAfter, tolerance, "the topic after")

			state = result.State
		})
	}
}

// Every corridor the reference printed: after each step of both sequences, and
// on a grid from well below the starting level to well above it.
func TestTheReferenceCorridorsAgree(t *testing.T) {
	t.Parallel()

	golden := loadGolden(t)
	rows := append([]goldenCorridor{}, golden.CorridorGrid...)
	for _, sequence := range []goldenSequence{golden.DocExample, golden.ConfigExample} {
		for _, step := range sequence.Steps {
			rows = append(rows, step.CorridorAfter)
		}
	}

	for _, want := range rows {
		t.Run(fmt.Sprintf("level %+.4f", want.Level), func(t *testing.T) {
			checkCorridor(t, &want)
		})
	}
}

// checkCorridor holds the package to one row of the reference.
func checkCorridor(t *testing.T, want *goldenCorridor) {
	t.Helper()

	got := rating.NewCorridor(want.Level)

	nearly(t, got.BetaMin, want.BetaMin, tolerance, "the hard end of the corridor")
	nearly(t, got.BetaMax, want.BetaMax, tolerance, "the easy end of the corridor")
	nearly(t, rating.Elo(want.Level), want.Elo, eloTolerance, "the rating")

	for difficulty := 1; difficulty <= rating.Difficulties; difficulty++ {
		key := strconv.Itoa(difficulty)
		nearly(t, got.Probability(difficulty), want.Probabilities[key], tolerance,
			"the chance at difficulty "+key)
	}

	sameDifficulties(t, got.Inside, want.Inside)
	if got.Recommended != want.Recommended {
		t.Errorf("recommended difficulty = %d, want %d", got.Recommended, want.Recommended)
	}
	if string(got.Fit) != want.Fit {
		t.Errorf("fit = %q, want %q", got.Fit, want.Fit)
	}
}

// The five seed profiles of the prototype, replayed by its own code. The
// answers behind them are not in this repository — they arrive with the
// profile fixtures — so what is checked here is everything those final
// numbers say about each other: a rating is its level on the chess scale, and
// a topic's level is the overall level with the topic's correction applied.
func TestTheSeedProfilesSitOnTheScale(t *testing.T) {
	t.Parallel()

	replays := loadGolden(t).SeedReplays
	if len(replays) == 0 {
		t.Fatal("the reference carries no replayed profiles")
	}

	for _, replay := range replays {
		t.Run(replay.Student, func(t *testing.T) {
			nearly(t, rating.Elo(replay.Theta), replay.Elo, eloTolerance, "the overall rating")

			for topic, want := range replay.Topics {
				nearly(t, replay.Theta+want.Delta, want.Level, tolerance, topic+": the level")
				nearly(t, rating.Elo(want.Level), want.Elo, eloTolerance, topic+": the rating")
			}
		})
	}
}

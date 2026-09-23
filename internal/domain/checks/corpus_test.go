package checks

import (
	"cmp"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"testing"
)

// The reference questions are read straight from their files rather than
// through the content package: here they are data the measure is held to, as
// the golden vectors are, and the domain does not import the content.

// referenceQuestion is what the measure needs of a reference task.
type referenceQuestion struct {
	ID         string `json:"id"`
	GradeLevel string `json:"grade_level"`
	Question   string `json:"question"`
}

// corpus is every reference question in the order of their ids, each cut into
// shingles once, and every pair of them measured.
type corpus struct {
	questions []referenceQuestion
	sets      []shingles
	pairs     []measuredPair
}

// measuredPair is the similarity of two reference questions, by their places.
type measuredPair struct {
	a, b       int
	similarity float64
}

func readCorpus(t *testing.T) corpus {
	t.Helper()

	questions := readReferenceQuestions(t)
	sets := make([]shingles, len(questions))
	for i, question := range questions {
		sets[i] = shinglesOf(question.Question)
	}
	pairs := make([]measuredPair, 0, len(questions)*(len(questions)-1)/2)
	for a := range questions {
		for b := a + 1; b < len(questions); b++ {
			pairs = append(pairs, measuredPair{a: a, b: b, similarity: sets[a].jaccard(sets[b])})
		}
	}
	return corpus{questions: questions, sets: sets, pairs: pairs}
}

// readReferenceQuestions reads every reference question from its file, in the
// order of their ids.
func readReferenceQuestions(t *testing.T) []referenceQuestion {
	t.Helper()

	files, err := filepath.Glob(filepath.Join("..", "..", "..", "content", "examples", "*.json"))
	if err != nil || len(files) == 0 {
		t.Fatalf("find the reference tasks: %v (%d files)", err, len(files))
	}
	var questions []referenceQuestion
	for _, file := range files {
		raw, err := os.ReadFile(file)
		if err != nil {
			t.Fatalf("read %s: %v", file, err)
		}
		var tasks []referenceQuestion
		if err := json.Unmarshal(raw, &tasks); err != nil {
			t.Fatalf("parse %s: %v", file, err)
		}
		questions = append(questions, tasks...)
	}
	slices.SortFunc(questions, func(x, y referenceQuestion) int { return cmp.Compare(x.ID, y.ID) })
	return questions
}

// goldenCorpus is what Postgres's similarity() said about the reference
// questions themselves, exported by the prototype's own code: the fifty
// most alike pairs, and how close each question's nearest neighbour is.
type goldenCorpus struct {
	TopPairs []struct {
		A          string  `json:"a"`
		B          string  `json:"b"`
		Similarity float64 `json:"similarity"`
	} `json:"examples_top_pairs"`
	Nearest struct {
		Count     int            `json:"count"`
		Max       float64        `json:"max"`
		Median    float64        `json:"median"`
		Histogram map[string]int `json:"histogram"`
	} `json:"examples_nearest_neighbour"`
}

func readGoldenCorpus(t *testing.T) goldenCorpus {
	t.Helper()

	raw, err := os.ReadFile(filepath.Join("..", "..", "..", "testdata", "golden", "trgm_similarity.json"))
	if err != nil {
		t.Fatalf("read the golden vectors: %v", err)
	}
	var golden goldenCorpus
	if err := json.Unmarshal(raw, &golden); err != nil {
		t.Fatalf("parse the golden vectors: %v", err)
	}
	return golden
}

// The measure is the one the prototype's threshold was chosen with, so it has
// to say what Postgres said about all 450 reference questions, not only about
// the fifteen pairs picked for the unit test: the fifty most alike pairs, and
// the whole distribution of nearest neighbours.
func TestTheReferenceCorpusMeasuresAsPostgresMeasuredIt(t *testing.T) {
	t.Parallel()

	golden := readGoldenCorpus(t)
	references := readCorpus(t)

	t.Run("the fifty most alike pairs", func(t *testing.T) {
		t.Parallel()
		checkTopPairs(t, golden, references)
	})
	t.Run("the nearest neighbour of every question", func(t *testing.T) {
		t.Parallel()
		checkNearest(t, golden, references)
	})
}

// checkTopPairs holds the fifty most alike pairs to Postgres's. Pairs that tie
// are ordered by how Postgres sorts ids, so the fifty are compared by the
// values they hold rather than by position.
func checkTopPairs(t *testing.T, golden goldenCorpus, references corpus) {
	t.Helper()

	ours := make(map[[2]string]float64, len(references.pairs))
	for _, pair := range references.pairs {
		ours[pairKey(references.questions[pair.a].ID, references.questions[pair.b].ID)] = pair.similarity
	}
	for _, want := range golden.TopPairs {
		got, found := ours[pairKey(want.A, want.B)]
		if !found || float32(got) != float32(want.Similarity) {
			t.Errorf("similarity of %s and %s = %.7f, want %.7f", want.A, want.B, got, want.Similarity)
		}
	}

	// A copy is sorted: the other subtest reads the same pairs at the same time.
	sorted := slices.Clone(references.pairs)
	slices.SortFunc(sorted, func(x, y measuredPair) int { return cmp.Compare(y.similarity, x.similarity) })
	for i, want := range golden.TopPairs {
		if float32(sorted[i].similarity) != float32(want.Similarity) {
			t.Errorf("pair %d of the most alike has similarity %.7f, want %.7f", i+1, sorted[i].similarity, want.Similarity)
		}
	}
}

// checkNearest holds the distribution of every question's nearest neighbour to
// Postgres's: the count, the closest, the median and every bucket.
func checkNearest(t *testing.T, golden goldenCorpus, references corpus) {
	t.Helper()

	nearest := nearestOfEach(t, references)
	if len(nearest) != golden.Nearest.Count {
		t.Fatalf("%d questions measured, want %d", len(nearest), golden.Nearest.Count)
	}
	if got := slices.Max(nearest); float32(got) != float32(golden.Nearest.Max) {
		t.Errorf("the closest neighbour is at %.7f, want %.7f", got, golden.Nearest.Max)
	}
	if got := median(nearest); math.Abs(got-golden.Nearest.Median) > 1e-6 {
		t.Errorf("the median neighbour is at %.7f, want %.7f", got, golden.Nearest.Median)
	}
	got := histogram(nearest)
	for bucket, want := range golden.Nearest.Histogram {
		if got[bucket] != want {
			t.Errorf("bucket %s holds %d questions, want %d", bucket, got[bucket], want)
		}
	}
	for bucket, count := range got {
		if _, known := golden.Nearest.Histogram[bucket]; !known {
			t.Errorf("bucket %s holds %d questions, and Postgres put none there", bucket, count)
		}
	}
}

// The profile keeps a sketch of each past task instead of its text, so a
// repeat is judged on an estimate, and its 64 positions sample the measure
// with an error of about 0.06. Measured over every pair of reference questions
// of one level, the estimate has to stay inside that, and where it decides
// differently from the exact measure it has to be a pair sitting on the
// threshold anyway.
func TestTheSketchEstimatesTheMeasureWithinItsError(t *testing.T) {
	t.Parallel()

	references := readCorpus(t)
	sketches := make([][sketchSize]byte, len(references.sets))
	for i, set := range references.sets {
		sketches[i] = sketchOf(set)
	}

	var compared int
	var squares float64
	for _, pair := range references.pairs {
		first, second := references.questions[pair.a], references.questions[pair.b]
		if first.GradeLevel != second.GradeLevel {
			continue
		}
		estimate := resemblance(sketches[pair.a], sketches[pair.b])
		miss := estimate - pair.similarity
		compared++
		squares += miss * miss

		if math.Abs(miss) > 0.3 {
			t.Errorf("%s and %s are %.3f alike, and their sketches say %.3f",
				first.ID, second.ID, pair.similarity, estimate)
		}
		if (estimate >= Threshold) != (pair.similarity >= Threshold) &&
			math.Abs(pair.similarity-Threshold) > 0.15 {
			t.Errorf("%s and %s are %.3f alike, and their sketches decide otherwise at %.3f",
				first.ID, second.ID, pair.similarity, estimate)
		}
	}
	if spread := math.Sqrt(squares / float64(compared)); spread > 0.06 {
		t.Errorf("over %d pairs the sketches miss by %.4f on average, want at most 0.06", compared, spread)
	}
}

// pairKey is a pair of ids in the order the corpus measures them in.
func pairKey(a, b string) [2]string {
	if b < a {
		a, b = b, a
	}
	return [2]string{a, b}
}

// nearestOfEach is, for every question, how alike the closest other one is,
// as the export received that number.
func nearestOfEach(t *testing.T, references corpus) []float64 {
	t.Helper()

	closest := make([]float64, len(references.questions))
	for _, pair := range references.pairs {
		closest[pair.a] = max(closest[pair.a], pair.similarity)
		closest[pair.b] = max(closest[pair.b], pair.similarity)
	}
	nearest := make([]float64, len(closest))
	for i, similarity := range closest {
		nearest[i] = asPostgres(t, similarity)
	}
	return nearest
}

// asPostgres is a similarity as the export received it: Postgres computes it in
// single precision and sends it as the shortest decimal that reads back to
// that float, and Python reads the decimal as a double. The buckets below are
// cut on that number, and a value exactly on an edge — 0.65 is thirteen
// twentieths — falls to one side or the other depending on it.
func asPostgres(t *testing.T, similarity float64) float64 {
	t.Helper()

	sent := strconv.FormatFloat(float64(float32(similarity)), 'g', -1, 32)
	received, err := strconv.ParseFloat(sent, 64)
	if err != nil {
		t.Fatalf("read back %q: %v", sent, err)
	}
	return received
}

// median is the middle of the values, or the mean of the two middle ones.
func median(values []float64) float64 {
	sorted := slices.Sorted(slices.Values(values))
	middle := len(sorted) / 2
	if len(sorted)%2 == 1 {
		return sorted[middle]
	}
	return (sorted[middle-1] + sorted[middle]) / 2
}

// histogram counts the values in buckets a twentieth wide, the way the export
// did: the last bucket runs to 1 and holds it.
func histogram(values []float64) map[string]int {
	counts := map[string]int{}
	for _, value := range values {
		low := min(math.Trunc(value*20)/20, 0.95)
		counts[fmt.Sprintf("%.2f-%.2f", low, low+0.05)]++
	}
	return counts
}

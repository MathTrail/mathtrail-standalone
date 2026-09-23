package checks_test

import (
	"encoding/base64"
	"testing"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/checks"
)

// A fingerprint is stored in every child's profile and compared against for as
// long as the profile lives. These are the sketches this service has always
// made of two questions, written out by hand: if either changes, every
// fingerprint already stored stops matching anything, and a child can be
// handed the same task twice with nothing to notice it. They are not to be
// regenerated; a change of algorithm needs a way to carry the old sketches.
func TestAFingerprintIsTheSameAsEverStored(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		question, want string
	}{
		{
			"Four spaceships dock in pairs. How many different pairs can they make?",
			"cEpEOM5/PodUbTQwjQI2RwjWZTJ4vjbl6lIXlbXUdU3LHILXhD51wXaU/HJgkIACR5XiCBcbdwJCu48XWTRJxg==",
		},
		{
			"小明有三个苹果。",
			"T/rZi7S23yOMz0yzAeAYxg1ngcnagmnzl8qJBzkKMfa8Bz2KN2vFa3VBUqwOCgtKL6/TTQeLy4T358gF0o/bNA==",
		},
	} {
		if got := checks.Fingerprint(test.question); got != test.want {
			t.Errorf("Fingerprint(%q) = %s, want %s", test.question, got, test.want)
		}
	}
}

// Sixty-four bytes, written as 88 characters of base64: the size the profile
// budget counts on, two hundred of them to a file.
func TestAFingerprintIsSixtyFourBytes(t *testing.T) {
	t.Parallel()

	fingerprint := checks.Fingerprint("Four spaceships dock in pairs. How many different pairs can they make?")
	if len(fingerprint) != 88 {
		t.Errorf("len = %d, want 88", len(fingerprint))
	}
	raw, err := base64.StdEncoding.DecodeString(fingerprint)
	if err != nil {
		t.Fatalf("decode %q: %v", fingerprint, err)
	}
	if len(raw) != 64 {
		t.Errorf("%d bytes, want 64", len(raw))
	}
}

// The share of positions at which two sketches agree estimates the similarity
// of the two questions — that is the whole contract of the sketch. On the
// prototype's measured pairs, sixty-four positions land within 0.2 of it.
func TestAgreeingPositionsEstimateTheSimilarity(t *testing.T) {
	t.Parallel()

	golden := readGoldenPairs(t)
	for _, pair := range golden.Pairs {
		if pair.A == "" || pair.B == "" {
			continue // nothing to sketch
		}
		t.Run(pair.Name, func(t *testing.T) {
			t.Parallel()

			estimate := agreement(t, checks.Fingerprint(pair.A), checks.Fingerprint(pair.B))
			if diff := estimate - pair.Similarity; diff > 0.2 || diff < -0.2 {
				t.Errorf("the sketches agree at %.3f of positions, and the questions are %.3f alike", estimate, pair.Similarity)
			}
		})
	}
}

// agreement is the share of positions at which two fingerprints agree.
func agreement(t *testing.T, a, b string) float64 {
	t.Helper()

	first, err := base64.StdEncoding.DecodeString(a)
	if err != nil {
		t.Fatalf("decode %q: %v", a, err)
	}
	second, err := base64.StdEncoding.DecodeString(b)
	if err != nil {
		t.Fatalf("decode %q: %v", b, err)
	}
	same := 0
	for i := range first {
		if first[i] == second[i] {
			same++
		}
	}
	return float64(same) / float64(len(first))
}

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
			"Ckjv501A0meGUo5lonVUXbwnTlFkwgACdSh7civ3lJZ8RtVctc3wgARnMBYRI7PFnNrFsMk8pUi+ob5MMoKsNL6d6QpIEzL3ZwVXxko3zOcq9+vjcv2a9leTAFNPEpWh",
		},
		{
			"小明有三个苹果。",
			"+ptG88/DEIbXGaKTepeaFsfae1tRLOq6/z17tHeFL7RYnxn9rZUApPFj972FcIsDuB+i7S3eEFYQna5E0BzAsjR5tpqoyWX30wT9uJuV/OWBXldQ9y78XA+PydOJAfAo",
		},
	} {
		if got := checks.Fingerprint(test.question, ""); got != test.want {
			t.Errorf("Fingerprint(%q) = %s, want %s", test.question, got, test.want)
		}
	}
}

// Ninety-six bytes, written as 128 characters of base64: the size the profile
// budget counts on, two hundred of them to a file.
func TestAFingerprintIsNinetySixBytes(t *testing.T) {
	t.Parallel()

	fingerprint := checks.Fingerprint("Four spaceships dock in pairs. How many different pairs can they make?", "")
	if len(fingerprint) != 128 {
		t.Errorf("len = %d, want 128", len(fingerprint))
	}
	raw, err := base64.StdEncoding.DecodeString(fingerprint)
	if err != nil {
		t.Fatalf("decode %q: %v", fingerprint, err)
	}
	if len(raw) != 96 {
		t.Errorf("%d bytes, want 96", len(raw))
	}
}

// The positions at which two sketches agree estimate the similarity of the two
// questions — that is the whole contract of the sketch. On the prototype's
// measured pairs, a hundred and ninety-two positions land within 0.12 of it.
func TestAgreeingPositionsEstimateTheSimilarity(t *testing.T) {
	t.Parallel()

	golden := readGoldenPairs(t)
	for _, pair := range golden.Pairs {
		if pair.A == "" || pair.B == "" {
			continue // nothing to sketch
		}
		t.Run(pair.Name, func(t *testing.T) {
			t.Parallel()

			estimate := agreement(t, checks.Fingerprint(pair.A, ""), checks.Fingerprint(pair.B, ""))
			if diff := estimate - pair.Similarity; diff > 0.12 || diff < -0.12 {
				t.Errorf("the sketches estimate %.3f, and the questions are %.3f alike", estimate, pair.Similarity)
			}
		})
	}
}

// agreement is the similarity two fingerprints estimate, read the way the
// profile stores them: four bits a position, two positions to a byte, the
// first in the high half. The share of positions that agree is taken back by
// the one in sixteen that agree by chance.
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
		if first[i]>>4 == second[i]>>4 {
			same++
		}
		if first[i]&0x0f == second[i]&0x0f {
			same++
		}
	}
	const chance = 1.0 / 16
	share := float64(same) / float64(2*len(first))
	return max(0, (share-chance)/(1-chance))
}

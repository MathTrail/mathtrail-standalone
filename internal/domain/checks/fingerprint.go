package checks

import (
	"encoding/base64"
	"hash/fnv"
	"math"
)

// sketchSize is how many positions a fingerprint has, of which four bits each
// are kept, two positions to a byte.
//
// The number and the width are chosen together, against the room a profile has
// for two hundred sketches. A hundred and ninety-two positions estimate the
// similarity with an error of about 0.03, where sixty-four erred by about 0.06
// — and at 0.06 the same task given new numbers in Russian, 0.74 alike against
// a threshold of 0.7, slipped under it about one time in five. Four bits a
// position keep the sketch at 96 bytes, 128 characters of base64, which leaves
// the profile inside its budget with all two hundred in it; a whole byte a
// position would buy half the positions for the same room.
const sketchSize = 192

// positionValues is how many values four bits of a position can take, and so
// how often two positions agree by chance when the elements behind them differ.
const positionValues = 16

// sketchBytes is the size of a sketch as the profile keeps it: the first of each
// two positions in the high half of a byte, the second in the low one.
const sketchBytes = sketchSize / 2

// sketchSeeds turn one hash of a shingle into as many independent ones as a
// sketch has positions.
//
// They are part of every fingerprint ever stored: a profile keeps the sketches
// of the tasks a child was given, and a sketch made with other seeds agrees
// with them nowhere. Changing any of this — the seeds, the hash, the mixing,
// which bits are kept — silently stops the service from noticing a repeat of
// anything given before the change.
var sketchSeeds = func() [sketchSize]uint64 {
	var seeds [sketchSize]uint64
	for i := range seeds {
		seeds[i] = uint64(i+1) * 0x9e3779b97f4a7c15 // the golden ratio, as a 64-bit increment
	}
	return seeds
}()

// Fingerprint is what the profile keeps of a question instead of its text: a
// MinHash sketch of the same shingles the duplicate check compares, enough to
// notice a repeat and not enough to read back what the child was asked. It is
// cut as the duplicate check cuts the question, by the language of the task.
func Fingerprint(question, language string) string {
	sketch := sketchOf(shinglesOf(question, unitFor(question, language)))
	var packed [sketchBytes]byte
	for i := range packed {
		packed[i] = sketch[2*i]<<4 | sketch[2*i+1]
	}
	return base64.StdEncoding.EncodeToString(packed[:])
}

// sketchOf keeps, for each hash function, the lowest four bits of the smallest
// hash any shingle has under it. Two sets agree at a position as often as the
// element holding the minimum is one they share — which is their Jaccard
// index — and, where it is not, one time in sixteen by chance; resemblance
// takes that chance back out.
func sketchOf(set shingles) [sketchSize]byte {
	var smallest [sketchSize]uint64
	for i := range smallest {
		smallest[i] = math.MaxUint64
	}
	for _, shingle := range set {
		hash := fnv.New64a()
		_, _ = hash.Write([]byte(shingle)) // writing to a hash never fails
		base := hash.Sum64()
		for i, seed := range &sketchSeeds {
			if value := mix(base ^ seed); value < smallest[i] {
				smallest[i] = value
			}
		}
	}

	var sketch [sketchSize]byte
	for i, value := range &smallest {
		sketch[i] = byte(value & (positionValues - 1)) // the lowest bits: the whole of them agrees or does not
	}
	return sketch
}

// mix is the finaliser of SplitMix64: it spreads every bit of its input over
// every bit of its output, so that one hash xor-ed with many seeds reads as
// many unrelated hashes.
func mix(value uint64) uint64 {
	value ^= value >> 30
	value *= 0xbf58476d1ce4e5b9
	value ^= value >> 27
	value *= 0x94d049bb133111eb
	value ^= value >> 31
	return value
}

// resemblance estimates the similarity of the two questions two sketches were
// made from. The share of positions that agree overstates it, because a
// position whose elements differ still agrees one time in sixteen; taking that
// out is what makes the estimate the similarity itself rather than a number a
// little above it, most above it for the questions least alike.
func resemblance(a, b *[sketchSize]byte) float64 {
	agree := 0
	for i := range a {
		if a[i] == b[i] {
			agree++
		}
	}
	const chance = 1.0 / positionValues
	share := float64(agree) / sketchSize
	return max(0, (share-chance)/(1-chance))
}

// readFingerprint reads a fingerprint back from the profile, and says whether
// it is one: a sketch that was damaged in the file is a repeat nobody can see,
// not a reason to refuse a task.
func readFingerprint(stored string) ([sketchSize]byte, bool) {
	var sketch [sketchSize]byte
	raw, err := base64.StdEncoding.DecodeString(stored)
	if err != nil || len(raw) != sketchBytes {
		return sketch, false
	}
	for i, packed := range raw {
		sketch[2*i], sketch[2*i+1] = packed>>4, packed&0x0f
	}
	return sketch, true
}

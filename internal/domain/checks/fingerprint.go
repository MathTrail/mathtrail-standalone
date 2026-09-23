package checks

import (
	"encoding/base64"
	"hash/fnv"
	"math"
)

// sketchSize is how many positions a fingerprint has. Sixty-four is a sampling
// error of about 0.06 on the estimated similarity, which is why the thresholds
// are not placed near a cliff, and 88 characters of base64 in the profile.
const sketchSize = 64

// sketchSeeds turn one hash of a shingle into sixty-four independent ones.
//
// They are part of every fingerprint ever stored: a profile keeps the sketches
// of the tasks a child was given, and a sketch made with other seeds agrees
// with them nowhere. Changing any of this — the seeds, the hash, the mixing,
// which byte is kept — silently stops the service from noticing a repeat of
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
// notice a repeat and not enough to read back what the child was asked.
func Fingerprint(question string) string {
	sketch := sketchOf(shinglesOf(question))
	return base64.StdEncoding.EncodeToString(sketch[:])
}

// sketchOf keeps, for each of sixty-four hash functions, the lowest byte of the
// smallest hash any shingle has under it. Two sets agree at a position as
// often as the element holding the minimum is one they share, which is their
// Jaccard index; one byte instead of eight adds agreements by chance, about
// one in 256 of the positions that differ.
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
		sketch[i] = byte(value & 0xff) // the lowest byte: the whole of it is what agrees or does not
	}
	return sketch
}

// mix is the finaliser of SplitMix64: it spreads every bit of its input over
// every bit of its output, so that one hash xor-ed with sixty-four seeds reads
// as sixty-four unrelated hashes.
func mix(value uint64) uint64 {
	value ^= value >> 30
	value *= 0xbf58476d1ce4e5b9
	value ^= value >> 27
	value *= 0x94d049bb133111eb
	value ^= value >> 31
	return value
}

// resemblance is the share of positions at which two sketches agree, which
// estimates the similarity of the two questions they were made from.
func resemblance(a, b [sketchSize]byte) float64 {
	agree := 0
	for i := range a {
		if a[i] == b[i] {
			agree++
		}
	}
	return float64(agree) / sketchSize
}

// readFingerprint reads a fingerprint back from the profile, and says whether
// it is one: a sketch that was damaged in the file is a repeat nobody can see,
// not a reason to refuse a task.
func readFingerprint(stored string) ([sketchSize]byte, bool) {
	var sketch [sketchSize]byte
	raw, err := base64.StdEncoding.DecodeString(stored)
	if err != nil || len(raw) != sketchSize {
		return sketch, false
	}
	copy(sketch[:], raw)
	return sketch, true
}

package main

import (
	"encoding/base64"
	"math"
	"math/rand/v2"
	"slices"
)

// The interval every share is reported with: 95 %, two-sided.
const (
	tail       = 0.025
	resamples  = 2000
	bisections = 100
)

// clopperPearson is the exact 95 % interval of a share of k in n: the
// smallest and the largest chance under which so few, or so many, would come
// up with probability at least 2.5 %.
func clopperPearson(k, n int) (low, high float64) {
	if n == 0 {
		return 0, 1
	}
	low, high = 0, 1
	if k > 0 {
		low = bisect(func(p float64) bool { return 1-binomialAtMost(k-1, n, p) < tail })
	}
	if k < n {
		high = bisect(func(p float64) bool { return binomialAtMost(k, n, p) > tail })
	}
	return low, high
}

// bisect is the point of [0, 1] where a condition that holds below it stops
// holding.
func bisect(below func(p float64) bool) float64 {
	low, high := 0.0, 1.0
	for range bisections {
		middle := (low + high) / 2
		if below(middle) {
			low = middle
		} else {
			high = middle
		}
	}
	return (low + high) / 2
}

// binomialAtMost is the chance of at most k successes in n tries of chance p.
func binomialAtMost(k, n int, p float64) float64 {
	switch {
	case k >= n:
		return 1
	case k < 0:
		return 0
	case p <= 0:
		return 1
	case p >= 1:
		return 0
	}
	sum := 0.0
	for i := 0; i <= k; i++ {
		sum += math.Exp(logChoose(n, i) + float64(i)*math.Log(p) + float64(n-i)*math.Log1p(-p))
	}
	return min(sum, 1)
}

func logChoose(n, k int) float64 {
	whole, _ := math.Lgamma(float64(n + 1))
	first, _ := math.Lgamma(float64(k + 1))
	rest, _ := math.Lgamma(float64(n - k + 1))
	return whole - first - rest
}

// bootstrapByHost is the 95 % percentile interval of a pooled share when the
// cases of one host belong together: hosts are drawn with replacement, each
// bringing all its cases.
func bootstrapByHost(perHost []share, rng *rand.Rand) (low, high float64) {
	if len(perHost) == 0 {
		return 0, 1
	}
	shares := make([]float64, 0, resamples)
	for range resamples {
		hits, total := 0, 0
		for range perHost {
			drawn := perHost[rng.IntN(len(perHost))]
			hits += drawn.hits
			total += drawn.total
		}
		if total > 0 {
			shares = append(shares, float64(hits)/float64(total))
		}
	}
	slices.Sort(shares)
	return percentile(shares, tail), percentile(shares, 1-tail)
}

// percentile is the value at this share of a sorted list.
func percentile(sorted []float64, at float64) float64 {
	if len(sorted) == 0 {
		return math.NaN()
	}
	return sorted[min(len(sorted)-1, int(at*float64(len(sorted))))]
}

// sketchSimilarity estimates how alike two questions are from their sketches,
// as the service does when it compares a task with the child's history: the
// share of the 192 four-bit positions that agree, less the one time in sixteen
// two positions agree by chance.
func sketchSimilarity(a, b string) (float64, bool) {
	first, err := base64.StdEncoding.DecodeString(a)
	if err != nil {
		return 0, false
	}
	second, err := base64.StdEncoding.DecodeString(b)
	if err != nil || len(first) != len(second) || len(first) == 0 {
		return 0, false
	}
	agree := 0
	for i := range first {
		if first[i]>>4 == second[i]>>4 {
			agree++
		}
		if first[i]&0x0f == second[i]&0x0f {
			agree++
		}
	}
	share := float64(agree) / float64(2*len(first))
	return max(0, (share-1.0/16)/(15.0/16)), true
}

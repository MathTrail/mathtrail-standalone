package main

import (
	"fmt"
	"hash/fnv"
	"math"
	"math/rand/v2"
	"slices"
	"time"
)

// The interval of a median: a 95 % percentile bootstrap over tasks, each drawn
// with all of its reviews.
const (
	resamples = 2000
	tail      = 0.025
)

// masterSeed is where every random draw of the experiment comes from, and
// experiment names it in every seed.
const (
	masterSeed = 20261001
	experiment = "E-A4"
)

// quantile is the q-th quantile of sorted values: the value at index ⌊q·n⌋,
// the last for q = 1, which is how the other experiments of the paper read
// their percentiles. Every quantile reported is thus a value measured.
func quantile(sorted []float64, q float64) float64 {
	if len(sorted) == 0 {
		return math.NaN()
	}
	return sorted[min(len(sorted)-1, int(q*float64(len(sorted))))]
}

// spread is a set of values read as its median, its 95th and 99th percentiles
// and its largest.
type spread struct {
	median, p95, p99, max float64
}

func spreadOf(values []float64) spread {
	sorted := slices.Sorted(slices.Values(values))
	return spread{
		median: quantile(sorted, 0.5),
		p95:    quantile(sorted, 0.95),
		p99:    quantile(sorted, 0.99),
		max:    quantile(sorted, 1),
	}
}

// medianInterval is the 95 % bootstrap interval of the median of values that
// come in groups, one group a task: the groups are drawn again with their
// values, so that a task's ten reviews are not taken for ten tasks.
func medianInterval(groups [][]float64, unit string) (low, high float64) {
	rng := seeded(unit, "bootstrap")
	medians := make([]float64, 0, resamples)
	var pooled []float64
	for range resamples {
		pooled = pooled[:0]
		for range groups {
			pooled = append(pooled, groups[rng.IntN(len(groups))]...)
		}
		slices.Sort(pooled)
		medians = append(medians, quantile(pooled, 0.5))
	}
	slices.Sort(medians)
	return quantile(medians, tail), quantile(medians, 1-tail)
}

// seeded is the random stream of one unit of the experiment for one purpose:
// the draws of a bootstrap are repeatable, though the times they draw from
// are not.
func seeded(unit, purpose string) *rand.Rand {
	hash := fnv.New64a()
	_, _ = fmt.Fprintf(hash, "%d|%s|%s|%s", masterSeed, experiment, unit, purpose)
	return rand.New(rand.NewPCG(hash.Sum64(), masterSeed)) //nolint:gosec // a repeatable resampling, not a secret
}

// milliseconds are durations as milliseconds.
func milliseconds(durations []time.Duration) []float64 {
	out := make([]float64, len(durations))
	for i, d := range durations {
		out[i] = float64(d) / float64(time.Millisecond)
	}
	return out
}

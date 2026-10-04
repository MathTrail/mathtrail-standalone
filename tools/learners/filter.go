package main

import (
	"math"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/rating"
)

// fullFilter is a Kalman filter over the overall level and every topic's
// offset at once, linearised at each answer: the whole matrix of how
// uncertain the levels are and how their errors go together, where the step
// from counted answers keeps one number for each and takes them to be apart.
// An answer narrows the matrix by what it tells at the chance the child had,
// and adds q to the uncertainty of the overall level and of the topic
// answered, as the step from counted answers does. It starts at the end of
// the trial series, from the uncertainty after the series that step is given,
// with every topic at the spread of topics and no error shared. What it gains
// over that step is what the step's simplicity costs.
//
// The topics take their places in the order they are first answered, so that
// every sum over them comes to the same bits every time.
type fullFilter struct {
	levels                   []float64   // the overall level, then each topic's offset
	uncertainty              [][]float64 // how the errors of the levels go together
	places                   map[string]int
	topicVariance            float64
	addedOverall, addedTopic float64
	limit                    float64
}

func newFullFilter(start, afterSeries, topicSpread, addedOverall, addedTopic, limit float64) *fullFilter {
	return &fullFilter{
		levels: []float64{start}, uncertainty: [][]float64{{afterSeries}}, places: map[string]int{},
		topicVariance: topicSpread * topicSpread, addedOverall: addedOverall, addedTopic: addedTopic, limit: limit,
	}
}

func (f *fullFilter) overall() float64 { return f.levels[0] }

func (f *fullFilter) level(topic string) float64 {
	if at, seen := f.places[topic]; seen {
		return f.levels[0] + f.levels[at]
	}
	return f.levels[0]
}

func (f *fullFilter) chance(topic string, beta float64) float64 {
	return rating.Guess + (1-rating.Guess)*logistic(f.level(topic)-beta)
}

// followed stands where the trial series' estimate stands.
func (f *fullFilter) followed(_ string, _ float64, _ bool, estimate float64) {
	f.levels[0] = estimate
}

// answered moves every level by its share of the surprise, which is how far
// its error goes with the error of the level in the topic answered, and
// narrows the matrix by what the answer tells.
func (f *fullFilter) answered(topic string, beta float64, correct bool) {
	at := f.placeOf(topic)
	p := f.chance(topic, beta)
	surprise := surpriseOf(correct, p)
	// along is how the error of every level goes with the error of the level
	// in the topic, the overall level plus the topic's offset.
	along := make([]float64, len(f.levels))
	for i, row := range f.uncertainty {
		along[i] = row[0] + row[at]
	}
	gain, information := gainAt(p), informationAt(p)
	together := 1 + information*(along[0]+along[at])
	step := gain / together * heldTo(f.limit, (along[0]+along[at])*gain/together*math.Abs(surprise))
	for i := range f.levels {
		f.levels[i] += along[i] * step * surprise
	}
	for i, row := range f.uncertainty {
		for j := range row {
			row[j] -= along[i] * along[j] * information / together
		}
	}
	f.uncertainty[0][0] += f.addedOverall
	f.uncertainty[at][at] += f.addedTopic
}

// placeOf is a topic's place among the levels, given it at its first answer
// with the spread of topics as its uncertainty and no error shared.
func (f *fullFilter) placeOf(topic string) int {
	if at, seen := f.places[topic]; seen {
		return at
	}
	at := len(f.levels)
	f.places[topic] = at
	f.levels = append(f.levels, 0)
	for i := range f.uncertainty {
		f.uncertainty[i] = append(f.uncertainty[i], 0)
	}
	row := make([]float64, at+1)
	row[at] = f.topicVariance
	f.uncertainty = append(f.uncertainty, row)
	return at
}

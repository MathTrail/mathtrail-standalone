package main

import (
	"fmt"
	"math"
	"slices"
	"strconv"
	"testing"
)

// manner is how a child answers and changes, apart from where they stand.
type manner struct {
	slope, floor, hint, writing float64
	jumps, learns, twoHosts     bool
}

// mannerOf is the manner of a child as drawn.
func mannerOf(c *child) manner {
	return manner{
		slope: c.slope, floor: c.floor, hint: c.hint, writing: c.writing,
		jumps: c.jumpAt >= 0, learns: c.learns, twoHosts: c.twoHosts,
	}
}

// Each generator departs from the base population in its one respect, half
// its children each way where it has two, so that a difference between the
// generators' results comes from that respect alone.
func TestEachGeneratorDepartsInItsOneRespect(t *testing.T) {
	t.Parallel()
	base := manner{slope: baseSlope, floor: baseFloor, writing: writingError}
	with := func(change func(m *manner)) manner {
		m := base
		change(&m)
		return m
	}
	// The manner of a child of the lower half, then of the upper.
	want := map[generator][2]manner{
		staticChildren:   {base, base},
		misplaced:        {base, base},
		learning:         {with(func(m *manner) { m.learns = true }), with(func(m *manner) { m.learns = true })},
		jumping:          {with(func(m *manner) { m.jumps = true }), with(func(m *manner) { m.jumps = true })},
		harderHost:       {with(func(m *manner) { m.twoHosts = true }), with(func(m *manner) { m.twoHosts = true })},
		linkedTopics:     {base, base},
		otherSlope:       {with(func(m *manner) { m.slope = 0.5 }), with(func(m *manner) { m.slope = 2 })},
		otherFloor:       {with(func(m *manner) { m.floor = 0 }), with(func(m *manner) { m.floor = 0.3 })},
		hintsUsed:        {with(func(m *manner) { m.hint = hintShare }), with(func(m *manner) { m.hint = hintShare })},
		exactlyAsWritten: {with(func(m *manner) { m.writing = 0 }), with(func(m *manner) { m.writing = 0 })},
	}
	for _, gen := range allGenerators {
		if _, listed := want[gen]; !listed {
			t.Errorf("generator %s has no departure listed here", gen)
		}
	}
	topics := []string{"a", "b", "c"}
	for gen, halves := range want {
		for i, wanted := range halves {
			if got := mannerOf(newChild(gen, i, topics)); got != wanted {
				t.Errorf("child %d of %s: %+v, want %+v", i, gen, got, wanted)
			}
		}
	}
}

// Linked topics fall into two groups — the first half of the catalog's order,
// rounded up, and the rest — and a child of linked topics stands to one side
// in the first and to the other in the second.
func TestLinkedTopicsSplitTheCatalogInTwo(t *testing.T) {
	t.Parallel()
	for topics, want := range map[int][]float64{
		1: {1}, 2: {1, -1}, 4: {1, 1, -1, -1}, 5: {1, 1, 1, -1, -1},
	} {
		t.Run(fmt.Sprintf("%d topics", topics), func(t *testing.T) {
			t.Parallel()
			sides := make([]float64, topics)
			for place := range topics {
				sides[place] = groupSide(place, topics)
			}
			if !slices.Equal(sides, want) {
				t.Errorf("sides %v, want %v", sides, want)
			}
		})
	}
	topics := make([]string, 20)
	for i := range topics {
		topics[i] = "t" + strconv.Itoa(i)
	}
	for i := range 10 {
		c := newChild(linkedTopics, i, topics)
		var groups [2]float64
		for place, topic := range topics {
			groups[2*place/len(topics)] += c.delta[topic]
		}
		if groups[0]*groups[1] >= 0 {
			t.Errorf("child %d: the two groups' offsets add up to %v and %v, want them on opposite sides", i, groups[0], groups[1])
		}
	}
}

// A child who learns grows overall and in the topic answered, and in every
// other topic drifts part of the way back to where they first stood; a child
// who jumps grows once, after the answer the jump is at; and any other child
// stays as drawn.
func TestAChildChangesOnlyAsItsGeneratorSays(t *testing.T) {
	t.Parallel()
	topics := []string{"a", "b"}
	c := newChild(learning, 0, topics)
	theta, a, b := c.theta, c.delta["a"], c.delta["b"]
	c.after("a", 0)
	if c.theta != theta+learningPerAnswer || c.delta["a"] != a+topicLearning || c.delta["b"] != b {
		t.Fatalf("after an answer in a: θ %v, a %v, b %v; want %v, %v and %v", c.theta, c.delta["a"], c.delta["b"], theta+learningPerAnswer, a+topicLearning, b)
	}
	c.after("b", 1)
	if got, want := c.delta["a"]-a, (1-forgetting)*topicLearning; math.Abs(got-want) > 1e-15 {
		t.Errorf("after an answer in b, a stands %v from where it first stood, want %v", got, want)
	}
	jumper := newChild(jumping, 0, topics)
	start := jumper.theta
	for k := range jumpLatest + 1 {
		jumper.after("a", k)
		want := start
		if k+1 >= jumper.jumpAt {
			want = start + jumpBy
		}
		if jumper.theta != want {
			t.Fatalf("after answer %d of a jump at %d: θ %v, want %v", k+1, jumper.jumpAt, jumper.theta, want)
		}
	}
	still := newChild(staticChildren, 0, topics)
	drawn := still.level("a")
	for k := range 20 {
		still.after("a", k)
	}
	if still.level("a") != drawn {
		t.Errorf("a child who neither learns nor jumps moved from %v to %v", drawn, still.level("a"))
	}
}

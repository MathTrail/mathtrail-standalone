package profile

import (
	"maps"
	"slices"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/rating"
)

// MaxIdeasPerList is the most ideas a topic's list may hold at a level the
// topic is taught at. An idea is named by its place on its list, from 1, so it
// is also the furthest place the file names: each list has a length of its
// own, which the content knows, and the file holds the places of the longest.
const MaxIdeasPerList = 40

// Idea is the idea of its topic's list a task is built on, at the level the
// task is written for: its place on the list, from 1, and the round of the
// list it was named in, the first being 1. The zero Idea names none: a request
// opened, or a task kept, by a build that named no idea.
type Idea struct {
	Number int `json:"number"`
	Round  int `json:"round"`
}

// IdeaRound is how far a child has come through one list of ideas: the round
// they are on, the first being 1, and the ideas of the list handed to them in
// it, the earliest first. An idea comes back only in a later round, once every
// idea of the list has been given in this one.
type IdeaRound struct {
	Given []int `json:"given"`
	Round int   `json:"round"`
}

// gave records that a task built on an idea of the topic's list at a level
// reached the child. An idea named in a round after the one the list is on
// opens that round with itself; one named in the round the list is on, or in
// an earlier one, joins the ideas given in the round the list is on, once. A
// task built on no idea records nothing.
//
// The record is replaced rather than written into, so that a copy of the
// profile taken before shares nothing that changes.
func (t *Topic) gave(level rating.GradeLevel, idea Idea) {
	if idea.Number < 1 {
		return
	}
	kept := t.Ideas[level]
	switch {
	case idea.Round > kept.Round:
		kept = IdeaRound{Given: []int{idea.Number}, Round: idea.Round}
	case slices.Contains(kept.Given, idea.Number):
		return
	default:
		kept.Given = append(slices.Clip(kept.Given), idea.Number)
	}
	ideas := maps.Clone(t.Ideas)
	if ideas == nil {
		ideas = map[rating.GradeLevel]IdeaRound{}
	}
	ideas[level] = kept
	t.Ideas = ideas
}

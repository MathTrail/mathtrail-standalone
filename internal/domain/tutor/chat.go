package tutor

import (
	"slices"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/profile"
	"github.com/MathTrail/mathtrail-standalone/internal/domain/rating"
)

// Shown is the catalog as a chat is shown it: the whole of it in a chat known
// to draw the card, and without the topics all of whose tasks are solved from
// the picture in any other. There they count as topics the catalog does not
// have: none is within reach or on pause, and one the lessons are kept to
// counts as no choice, so that the rule chooses in its place.
func Shown(catalog Catalog, chat profile.Chat) Catalog {
	if chat.ShowsPictures() {
		return catalog
	}
	return wordTopics{catalog}
}

// wordTopics is a catalog without its topics solved from the picture.
type wordTopics struct{ Catalog }

// TopicIDs lists the topics of the catalog told in words, in catalog order,
// in a list of its own: the catalog's may be the one it keeps.
func (w wordTopics) TopicIDs() []string {
	return slices.DeleteFunc(slices.Clone(w.Catalog.TopicIDs()), w.SolvedFromPicture)
}

// LevelsOf lists the levels a topic is taught at, and none for a topic solved
// from the picture, which the chat is not shown.
func (w wordTopics) LevelsOf(topic string) []rating.GradeLevel {
	if w.SolvedFromPicture(topic) {
		return nil
	}
	return w.Catalog.LevelsOf(topic)
}

package tutor

import (
	"crypto/sha256"
	"encoding/binary"
	"slices"
	"strconv"
	"strings"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/profile"
	"github.com/MathTrail/mathtrail-standalone/internal/domain/rating"
)

// Ideas is what the draw of an idea has to know of the content: how many
// ideas a topic's list holds at a level, and which of them are solved from
// the picture.
type Ideas interface {
	// IdeaCount is how many ideas the list of a topic holds at a level.
	IdeaCount(topic string, level rating.GradeLevel) int
	// IdeaFromPicture reports whether the task built on the idea at a place,
	// from 1, of that list is solved from the picture.
	IdeaFromPicture(topic string, level rating.GradeLevel, number int) bool
}

// IdeaFor is the idea the task of a brief is to be built on, for the chat the
// task would go to, and it marks the brief facts_on_the_picture when the idea
// is one solved from the picture. A chat not known to draw the card, and the
// trial series in any chat, pass over the ideas solved from the picture, which
// stay in their round for a chat that does; a topic all of whose tasks are
// solved from it passes over none, since its brief is one only where such a
// task may go.
func IdeaFor(p *profile.Profile, ideas Ideas, brief *profile.Brief, chat profile.Chat) profile.Idea {
	topic, level := brief.TargetConcept, brief.GradeLevel
	pictured := func(number int) bool { return ideas.IdeaFromPicture(topic, level, number) }
	var passed func(int) bool
	if !brief.FromPicture() && (!chat.ShowsPictures() || p.Ratings.InTrial()) {
		passed = pictured
	}
	idea := drawIdea(p, topic, level, ideas.IdeaCount(topic, level), passed)
	if idea.Number > 0 && pictured(idea.Number) && !brief.FromPicture() {
		brief.Constraints = append(brief.Constraints, profile.FactsOnThePicture)
	}
	return idea
}

// RequestIdea is the idea a request's task is built on: the one the request
// named when it was opened, while its list still holds it as an idea of the
// kind the request's brief asked for — solved from the picture, or told in
// words —, and otherwise the one IdeaAgain names. A request opened by an
// earlier build named none, and one opened before a list was changed may name
// a place the list has no more, or one of the other kind now. Nothing given
// changes while a request is open, so its package and its task are built on
// one idea whenever either is asked for.
func RequestIdea(p *profile.Profile, ideas Ideas, request *profile.OpenRequest) profile.Idea {
	brief, named := &request.Brief, request.Idea
	topic, level := brief.TargetConcept, brief.GradeLevel
	if named != (profile.Idea{}) && named.Number <= ideas.IdeaCount(topic, level) &&
		ideas.IdeaFromPicture(topic, level, named.Number) == brief.FromPicture() {
		return named
	}
	return IdeaAgain(p, ideas, brief)
}

// IdeaAgain is the idea a request opened on brief is built on when the idea
// it named is no longer on its list as it was: one a request opened now would
// name, among the ideas of the kind its brief asked for, since the brief, its
// package and its checks are already that kind's — or among every idea of the
// list, should it hold none of that kind any more, since a request built on
// no idea could not be written at all.
func IdeaAgain(p *profile.Profile, ideas Ideas, brief *profile.Brief) profile.Idea {
	topic, level := brief.TargetConcept, brief.GradeLevel
	length := ideas.IdeaCount(topic, level)
	otherKind := func(number int) bool { return ideas.IdeaFromPicture(topic, level, number) != brief.FromPicture() }
	if idea := drawIdea(p, topic, level, length, otherKind); idea != (profile.Idea{}) {
		return idea
	}
	return drawIdea(p, topic, level, length, nil)
}

// NextIdea is the idea of a topic's list at a level that the next task of the
// topic at that level is to be built on, among every idea of the list, as
// drawIdea draws it.
func NextIdea(p *profile.Profile, topic string, level rating.GradeLevel, length int) profile.Idea {
	return drawIdea(p, topic, level, length, nil)
}

// drawIdea is the idea of a topic's list at a level that the next task of the
// topic at that level is to be built on: one not yet given in the round the
// list is on, or, once the round has given every idea, one of a new round
// other than the idea given last, so that no two tasks in a row are built on
// one idea. The list holds length ideas, a length of its own: one that has
// grown since its round opened gives its new ideas in that round, one of a
// single idea gives it in every round, and a list of none names no idea.
//
// The places passed over are drawn from no round: they stay in the round for
// a draw that takes them, and a round whose ideas left are all passed over
// opens the next round, as a round that has given every idea does. A list
// whose every idea is passed over names no idea.
//
// Which idea it is looks drawn at random — another order for every child and
// every round — and is not: it is read off a hash of the child's id, the
// topic, the level, the round and how many ideas the record of the list holds,
// so that every draw of a round is a draw of its own. The same profile always
// names the same idea, so a request opened again for a task that never reached
// the child names what it named before, and a test can say which. A record the
// file was edited into names an idea of the list all the same: numbers off the
// list are passed over.
func drawIdea(p *profile.Profile, topic string, level rating.GradeLevel, length int, passed func(int) bool) profile.Idea {
	kept := p.Topics[topic].Ideas[level]
	round, given := max(kept.Round, 1), kept.Given
	taken := func(places []int) []int {
		if passed == nil {
			return places
		}
		return slices.DeleteFunc(places, passed)
	}
	left := taken(ideasBut(length, given...))
	if len(left) == 0 && len(given) > 0 {
		round++
		if left = taken(ideasBut(length, given[len(given)-1])); len(left) == 0 {
			left = taken(ideasBut(length))
		}
	}
	if len(left) == 0 {
		return profile.Idea{}
	}
	key := strings.Join([]string{p.StudentID, topic, string(level), strconv.Itoa(round), strconv.Itoa(len(given))}, "\x00")
	sum := sha256.Sum256([]byte(key))
	drawn := binary.BigEndian.Uint64(sum[:8]) % uint64(len(left))
	return profile.Idea{Number: left[drawn], Round: round}
}

// ideasBut is the places on a list of ideas of that length, in order, but for
// those given.
func ideasBut(length int, given ...int) []int {
	left := make([]int, 0, length)
	for number := 1; number <= length; number++ {
		if !slices.Contains(given, number) {
			left = append(left, number)
		}
	}
	return left
}

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

// NextIdea is the idea of a topic's list at a level that the next task of the
// topic at that level is to be built on: one not yet given in the round the
// list is on, or, once the round has given every idea, one of a new round
// other than the idea given last, so that no two tasks in a row are built on
// one idea. The list holds length ideas, a length of its own: one that has
// grown since its round opened gives its new ideas in that round, one of a
// single idea gives it in every round, and a list of none names no idea.
//
// Which of those it is looks drawn at random — another order for every child
// and every round — and is not: it is read off a hash of the child's id, the
// topic, the level, the round and how many ideas the record of the list holds,
// so that every draw of a round is a draw of its own. The same profile always
// names the same idea, so a request opened again for a task that never reached
// the child names what it named before, and a test can say which. A record the
// file was edited into names an idea of the list all the same: numbers off the
// list are passed over.
func NextIdea(p *profile.Profile, topic string, level rating.GradeLevel, length int) profile.Idea {
	if length < 1 {
		return profile.Idea{}
	}
	kept := p.Topics[topic].Ideas[level]
	round, given := max(kept.Round, 1), kept.Given
	left := ideasBut(length, given...)
	if len(left) == 0 {
		round++
		if left = ideasBut(length, given[len(given)-1]); len(left) == 0 {
			left = ideasBut(length)
		}
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

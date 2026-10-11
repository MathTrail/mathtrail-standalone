package tutor_test

import (
	"fmt"
	"slices"
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/profile"
	"github.com/MathTrail/mathtrail-standalone/internal/domain/rating"
	"github.com/MathTrail/mathtrail-standalone/internal/domain/tutor"
)

// gapsBrief is a brief on counting.gaps at a level, as the rule could write
// one.
func gapsBrief(level rating.GradeLevel) profile.Brief {
	return profile.Brief{
		Constraints: []string{}, Difficulty: 2, ExcludedSkills: []string{}, GradeLevel: level,
		PedagogicalGoal: profile.GoalReinforce, TargetConcept: "counting.gaps",
		TrapsToUse: []string{"off_by_one", "missed_case"},
	}
}

// childCalled is a child whose id is the one given, so that the order the
// rule names ideas in is the same on every run.
func childCalled(t *testing.T, id string) *profile.Profile {
	t.Helper()

	p := child(t)
	p.StudentID = id
	return p
}

// ideasHandedOut hands out tasks of counting.gaps at 1-2 one after another,
// each on the idea the rule names for it, and is those ideas, in order.
func ideasHandedOut(t *testing.T, p *profile.Profile, tasks int) []profile.Idea {
	t.Helper()

	sealer, err := taskSeal()
	if err != nil {
		t.Fatalf("build the seal: %v", err)
	}
	brief := gapsBrief(rating.Grades12)
	named := make([]profile.Idea, 0, tasks)
	for range tasks {
		named = append(named, tutor.NextIdea(p, brief.TargetConcept, brief.GradeLevel))
		handedOut(t, p, &brief, profile.TutorRule, sealer, day)
	}
	return named
}

// The rule names an idea not yet given in its round: any idea of a list never
// met, the one idea a round has left, and, once a round has given every idea,
// one of the next round other than the idea given last. A level has a round of
// its own, and a record a file was edited into still names an idea of the
// list.
func TestTheRuleNamesAnIdeaNotYetGivenInItsRound(t *testing.T) {
	t.Parallel()

	whole := []int{4, 1, 7, 2, 9, 3, 10, 5, 8, 6}
	for _, tc := range []struct {
		name  string
		ideas map[rating.GradeLevel]profile.IdeaRound
		level rating.GradeLevel
		holds func(profile.Idea) bool
	}{
		{"a list never met", nil, rating.Grades12,
			func(idea profile.Idea) bool { return idea.Round == 1 }},
		{"a round with one idea left", map[rating.GradeLevel]profile.IdeaRound{
			rating.Grades12: {Given: []int{4, 1, 7, 2, 9, 10, 5, 8, 6}, Round: 3},
		}, rating.Grades12, func(idea profile.Idea) bool { return idea == profile.Idea{Number: 3, Round: 3} }},
		{"a round with every idea given", map[rating.GradeLevel]profile.IdeaRound{
			rating.Grades12: {Given: whole, Round: 3},
		}, rating.Grades12, func(idea profile.Idea) bool { return idea.Round == 4 && idea.Number != 6 }},
		{"a level never met beside one gone through", map[rating.GradeLevel]profile.IdeaRound{
			rating.Grades12: {Given: whole, Round: 3},
		}, rating.Grades34, func(idea profile.Idea) bool { return idea.Round == 1 }},
		{"a record with places the list does not have", map[rating.GradeLevel]profile.IdeaRound{
			rating.Grades12: {Given: []int{0, 11, 3, -4}, Round: 2},
		}, rating.Grades12, func(idea profile.Idea) bool { return idea.Round == 2 && idea.Number != 3 }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			p := childCalled(t, "stu_ideas")
			p.Topics["counting.gaps"] = profile.Topic{Ideas: tc.ideas, Traps: map[string]int{}}
			idea := tutor.NextIdea(p, "counting.gaps", tc.level)
			if idea.Number < 1 || idea.Number > profile.IdeasPerList || !tc.holds(idea) {
				t.Errorf("NextIdea() = %+v, which is not what the record %+v leaves", idea, tc.ideas)
			}
		})
	}
}

// The order a list is gone through is each child's own: over a thousand
// children, the first idea of a list is every idea on it, and none is first
// twice as often as its share. A rule that named the first idea not given,
// rather than drawing one, would start every child on the same idea.
func TestChildrenMeetAListInOrdersOfTheirOwn(t *testing.T) {
	t.Parallel()

	const children = 1000
	first := map[int]int{}
	for i := range children {
		p := childCalled(t, fmt.Sprintf("stu_%04d", i))
		first[tutor.NextIdea(p, "counting.gaps", rating.Grades12).Number]++
	}
	if len(first) != profile.IdeasPerList {
		t.Errorf("the first ideas of %d children are %v, want every idea of the list among them", children, first)
	}
	for number, times := range first {
		if share := children / profile.IdeasPerList; times > 2*share {
			t.Errorf("idea %d comes first for %d children of %d, want at most twice its share, %d",
				number, times, children, 2*share)
		}
	}
}

// Every draw of a round is a draw of its own: among the children whose first
// draw takes the first idea of the list, the sixth draw takes the first idea
// left about as often as chance has it, one child in five. Draws read off one
// number for the whole round would tie the two, and take the first idea left
// at the sixth draw for every one of those children.
func TestEveryDrawOfARoundIsADrawOfItsOwn(t *testing.T) {
	t.Parallel()

	const children = 1000
	openedFirst, sixthFirst := 0, 0
	for i := range children {
		p := childCalled(t, fmt.Sprintf("stu_%04d", i))
		given := []int{}
		var places []int
		for range 6 {
			p.Topics["counting.gaps"] = profile.Topic{Ideas: map[rating.GradeLevel]profile.IdeaRound{
				rating.Grades12: {Given: given, Round: 1},
			}, Traps: map[string]int{}}
			left := slices.DeleteFunc(listOfIdeas(), func(n int) bool { return slices.Contains(given, n) })
			number := tutor.NextIdea(p, "counting.gaps", rating.Grades12).Number
			places = append(places, slices.Index(left, number))
			given = append(given, number)
		}
		if places[0] == 0 {
			openedFirst++
			if places[5] == 0 {
				sixthFirst++
			}
		}
	}
	if openedFirst == 0 || 2*sixthFirst > openedFirst {
		t.Errorf("of %d children whose first draw took the first idea, %d took the first idea left at the sixth, "+
			"want about one in five", openedFirst, sixthFirst)
	}
}

// Every round of a list is drawn afresh: two rounds a child opens with nothing
// given in either open on the same idea only as often as chance has it, about
// one child in ten. A draw that left the round out would open every round of a
// child's list on one idea.
func TestEveryRoundOfAListIsDrawnAfresh(t *testing.T) {
	t.Parallel()

	const children = 200
	opening := func(p *profile.Profile, round int) int {
		p.Topics["counting.gaps"] = profile.Topic{Ideas: map[rating.GradeLevel]profile.IdeaRound{
			rating.Grades12: {Given: []int{}, Round: round},
		}, Traps: map[string]int{}}
		return tutor.NextIdea(p, "counting.gaps", rating.Grades12).Number
	}
	alike := 0
	for i := range children {
		p := childCalled(t, fmt.Sprintf("stu_%04d", i))
		if opening(p, 1) == opening(p, 2) {
			alike++
		}
	}
	if alike > children/4 {
		t.Errorf("%d children of %d open two rounds of a list on one idea, want about one in ten", alike, children)
	}
}

// What the rule names over a run of tasks of one list: a round of it gives
// every idea once before any comes back, a new round never opens with the idea
// the last one closed on, and the same profile, or the file it is read back
// from, always names the same idea. Whatever record a file holds, the idea
// named is one of the list, in a round from the first.
func TestTheIdeasNamedHoldTheirProperties(t *testing.T) {
	t.Parallel()

	properties := gopter.NewProperties(nil)
	properties.Property("a round gives every idea of the list once", prop.ForAll(
		func(id string, tasks int) bool {
			return roundsGiveEachIdeaOnce(ideasHandedOut(t, childCalled(t, id), tasks))
		},
		gen.Identifier(), gen.IntRange(1, 3*profile.IdeasPerList+3),
	))
	properties.Property("a new round never opens with the idea the last one closed on", prop.ForAll(
		func(id string) bool {
			return noRoundOpensOnTheLast(ideasHandedOut(t, childCalled(t, id), 4*profile.IdeasPerList))
		},
		gen.Identifier(),
	))
	properties.Property("the same profile, and the file read back, name the same idea", prop.ForAll(
		func(id string, tasks int) bool {
			p := childCalled(t, id)
			ideasHandedOut(t, p, tasks)
			return namesTheSameIdeaReadBack(p)
		},
		gen.Identifier(), gen.IntRange(0, 2*profile.IdeasPerList+1),
	))
	properties.Property("whatever record a file holds, the idea is of the list in a round from the first", prop.ForAll(
		func(given []int, round int) bool {
			p := childCalled(t, "stu_any_record")
			p.Topics["counting.gaps"] = profile.Topic{Ideas: map[rating.GradeLevel]profile.IdeaRound{
				rating.Grades12: {Given: given, Round: round},
			}, Traps: map[string]int{}}
			return namedFrom(tutor.NextIdea(p, "counting.gaps", rating.Grades12), given, round)
		},
		gen.SliceOf(gen.IntRange(-2, profile.IdeasPerList+3)), gen.IntRange(-3, 1<<30-1),
	))
	properties.TestingRun(t)
}

// roundsGiveEachIdeaOnce reports whether ideas named one after another fill
// their rounds in turn, a list's worth to a round, and give no idea twice in
// a round.
func roundsGiveEachIdeaOnce(named []profile.Idea) bool {
	for start := 0; start < len(named); start += profile.IdeasPerList {
		round := named[start:min(start+profile.IdeasPerList, len(named))]
		numbers := make([]int, 0, len(round))
		for _, idea := range round {
			if idea.Round != start/profile.IdeasPerList+1 {
				return false
			}
			numbers = append(numbers, idea.Number)
		}
		slices.Sort(numbers)
		if len(slices.Compact(numbers)) != len(round) || numbers[0] < 1 || numbers[len(numbers)-1] > profile.IdeasPerList {
			return false
		}
	}
	return true
}

// noRoundOpensOnTheLast reports whether no round of ideas named one after
// another opens with the idea the round before it closed on.
func noRoundOpensOnTheLast(named []profile.Idea) bool {
	for start := profile.IdeasPerList; start < len(named); start += profile.IdeasPerList {
		if named[start].Number == named[start-1].Number {
			return false
		}
	}
	return true
}

// namesTheSameIdeaReadBack reports whether the profile names the same idea of
// counting.gaps at 1-2 when asked again, and when read back from its file.
func namesTheSameIdeaReadBack(p *profile.Profile) bool {
	next := tutor.NextIdea(p, "counting.gaps", rating.Grades12)
	file, err := profile.Marshal(p)
	if err != nil {
		return false
	}
	read, err := profile.Parse(file)
	return err == nil && tutor.NextIdea(p, "counting.gaps", rating.Grades12) == next &&
		tutor.NextIdea(read, "counting.gaps", rating.Grades12) == next
}

// namedFrom reports whether an idea is one a record of a round and the ideas
// given in it leaves: an idea of the list not given yet, in the same round, or,
// once the round has given every idea, one of the next round other than the
// idea given last.
func namedFrom(idea profile.Idea, given []int, round int) bool {
	if idea.Number < 1 || idea.Number > profile.IdeasPerList {
		return false
	}
	current := max(round, 1)
	if slices.ContainsFunc(listOfIdeas(), func(n int) bool { return !slices.Contains(given, n) }) {
		return idea.Round == current && !slices.Contains(given, idea.Number)
	}
	return idea.Round == current+1 && idea.Number != given[len(given)-1]
}

// listOfIdeas is every place on a list of ideas.
func listOfIdeas() []int {
	places := make([]int, 0, profile.IdeasPerList)
	for number := 1; number <= profile.IdeasPerList; number++ {
		places = append(places, number)
	}
	return places
}

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
// each on the idea the rule names for it from a list of that many ideas, and
// is those ideas, in order.
func ideasHandedOut(t *testing.T, p *profile.Profile, ideas, tasks int) []profile.Idea {
	t.Helper()

	sealer, err := taskSeal()
	if err != nil {
		t.Fatalf("build the seal: %v", err)
	}
	brief := gapsBrief(rating.Grades12)
	named := make([]profile.Idea, 0, tasks)
	for range tasks {
		named = append(named, tutor.NextIdea(p, brief.TargetConcept, brief.GradeLevel, ideas))
		handedOut(t, p, ideas, &brief, profile.TutorRule, sealer, day)
	}
	return named
}

// The rule names an idea not yet given in its round: any idea of a list never
// met, the one idea a round has left, and, once a round has given every idea,
// one of the next round other than the idea given last. A list that has grown
// since its round opened gives its new ideas in that round, a list of one idea
// gives it again, and a list of none names no idea. A level has a round of its
// own, and a record a file was edited into still names an idea of the list.
func TestTheRuleNamesAnIdeaNotYetGivenInItsRound(t *testing.T) {
	t.Parallel()

	whole := []int{4, 1, 7, 2, 9, 3, 10, 5, 8, 6}
	for _, tc := range []struct {
		name  string
		ideas int
		kept  map[rating.GradeLevel]profile.IdeaRound
		level rating.GradeLevel
		holds func(profile.Idea) bool
	}{
		{"a list never met", tenIdeas, nil, rating.Grades12,
			func(idea profile.Idea) bool { return idea.Round == 1 && onTheList(idea, tenIdeas) }},
		{"a round with one idea left", tenIdeas, map[rating.GradeLevel]profile.IdeaRound{
			rating.Grades12: {Given: []int{4, 1, 7, 2, 9, 10, 5, 8, 6}, Round: 3},
		}, rating.Grades12, func(idea profile.Idea) bool { return idea == profile.Idea{Number: 3, Round: 3} }},
		{"a round with every idea given", tenIdeas, map[rating.GradeLevel]profile.IdeaRound{
			rating.Grades12: {Given: whole, Round: 3},
		}, rating.Grades12, func(idea profile.Idea) bool {
			return idea.Round == 4 && idea.Number != 6 && onTheList(idea, tenIdeas)
		}},
		{"a level never met beside one gone through", tenIdeas, map[rating.GradeLevel]profile.IdeaRound{
			rating.Grades12: {Given: whole, Round: 3},
		}, rating.Grades34, func(idea profile.Idea) bool { return idea.Round == 1 && onTheList(idea, tenIdeas) }},
		{"a record with places the list does not have", tenIdeas, map[rating.GradeLevel]profile.IdeaRound{
			rating.Grades12: {Given: []int{0, 11, 3, -4}, Round: 2},
		}, rating.Grades12, func(idea profile.Idea) bool {
			return idea.Round == 2 && idea.Number != 3 && onTheList(idea, tenIdeas)
		}},
		{"a list grown in a later round", 20, map[rating.GradeLevel]profile.IdeaRound{
			rating.Grades12: {Given: whole, Round: 2},
		}, rating.Grades12, func(idea profile.Idea) bool {
			return idea.Round == 2 && idea.Number > len(whole) && onTheList(idea, 20)
		}},
		{"a list of one idea, its round over", 1, map[rating.GradeLevel]profile.IdeaRound{
			rating.Grades12: {Given: []int{1}, Round: 1},
		}, rating.Grades12, func(idea profile.Idea) bool { return idea == profile.Idea{Number: 1, Round: 2} }},
		{"a list of no ideas", 0, nil, rating.Grades12,
			func(idea profile.Idea) bool { return idea == profile.Idea{} }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			p := childCalled(t, "stu_ideas")
			p.Topics["counting.gaps"] = profile.Topic{Ideas: tc.kept, Traps: map[string]int{}}
			if idea := tutor.NextIdea(p, "counting.gaps", tc.level, tc.ideas); !tc.holds(idea) {
				t.Errorf("NextIdea() = %+v from a list of %d, which is not what the record %+v leaves",
					idea, tc.ideas, tc.kept)
			}
		})
	}
}

// The order a list is gone through is each child's own: over a thousand
// children, the first idea of a list is every idea on it, and none is first
// twice as often as its share, however long the list. A rule that named the
// first idea not given, rather than drawing one, would start every child on
// the same idea.
func TestChildrenMeetAListInOrdersOfTheirOwn(t *testing.T) {
	t.Parallel()

	const children = 1000
	for _, ideas := range []int{tenIdeas, 20, profile.MaxIdeasPerList} {
		first := map[int]int{}
		for i := range children {
			p := childCalled(t, fmt.Sprintf("stu_%04d", i))
			first[tutor.NextIdea(p, "counting.gaps", rating.Grades12, ideas).Number]++
		}
		if len(first) != ideas || slices.ContainsFunc(listOf(ideas), func(n int) bool { return first[n] == 0 }) {
			t.Errorf("the first ideas of %d children on a list of %d are %v, want every idea of the list among them",
				children, ideas, first)
		}
		for number, times := range first {
			if share := children / ideas; times > 2*share {
				t.Errorf("idea %d of %d comes first for %d children of %d, want at most twice its share, %d",
					number, ideas, times, children, 2*share)
			}
		}
	}
}

// Every draw of a round is a draw of its own: among the children whose first
// draw takes the first idea of a list of ten, the sixth draw takes the first
// idea left about as often as chance has it, one child in five. Draws read off
// one number for the whole round would tie the two, and take the first idea
// left at the sixth draw for every one of those children.
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
			left := slices.DeleteFunc(listOf(tenIdeas), func(n int) bool { return slices.Contains(given, n) })
			number := tutor.NextIdea(p, "counting.gaps", rating.Grades12, tenIdeas).Number
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
// given in either open on the same idea of a list of ten only as often as
// chance has it, about one child in ten. A draw that left the round out would
// open every round of a child's list on one idea.
func TestEveryRoundOfAListIsDrawnAfresh(t *testing.T) {
	t.Parallel()

	const children = 200
	opening := func(p *profile.Profile, round int) int {
		p.Topics["counting.gaps"] = profile.Topic{Ideas: map[rating.GradeLevel]profile.IdeaRound{
			rating.Grades12: {Given: []int{}, Round: round},
		}, Traps: map[string]int{}}
		return tutor.NextIdea(p, "counting.gaps", rating.Grades12, tenIdeas).Number
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

// What the rule names over a run of tasks of one list, however long the list:
// a round of it gives every idea once before any comes back, a new round never
// opens with the idea the last one closed on while the list has another, and
// the same profile, or the file it is read back from, always names the same
// idea. Whatever record a file holds and however many ideas the list holds,
// the idea named is one of the list, in a round from the first, and a list of
// none names none.
func TestTheIdeasNamedHoldTheirProperties(t *testing.T) {
	t.Parallel()

	properties := gopter.NewProperties(nil)
	properties.Property("a round gives every idea of the list once", prop.ForAll(
		func(id string, ideas, rounds, rest int) bool {
			tasks := rounds*ideas + rest%ideas
			return roundsGiveEachIdeaOnce(ideasHandedOut(t, childCalled(t, id), ideas, tasks), ideas)
		},
		gen.Identifier(), listLengths(1), gen.IntRange(1, 2), gen.IntRange(0, profile.MaxIdeasPerList),
	))
	properties.Property("a new round never opens with the idea the last one closed on", prop.ForAll(
		func(id string, ideas int) bool {
			return noRoundOpensOnTheLast(ideasHandedOut(t, childCalled(t, id), ideas, 3*ideas+1), ideas)
		},
		gen.Identifier(), listLengths(2),
	))
	properties.Property("the same profile, and the file read back, name the same idea", prop.ForAll(
		func(id string, ideas, tasks int) bool {
			p := childCalled(t, id)
			ideasHandedOut(t, p, ideas, tasks)
			return namesTheSameIdeaReadBack(p, ideas)
		},
		gen.Identifier(), listLengths(1), gen.IntRange(0, 2*profile.MaxIdeasPerList+1),
	))
	properties.Property("whatever record a file holds, the idea is of the list in a round from the first", prop.ForAll(
		func(given []int, round, ideas int) bool {
			p := childCalled(t, "stu_any_record")
			p.Topics["counting.gaps"] = profile.Topic{Ideas: map[rating.GradeLevel]profile.IdeaRound{
				rating.Grades12: {Given: given, Round: round},
			}, Traps: map[string]int{}}
			return namedFrom(tutor.NextIdea(p, "counting.gaps", rating.Grades12, ideas), given, round, ideas)
		},
		gen.SliceOf(gen.IntRange(-2, profile.MaxIdeasPerList+3)), gen.IntRange(-3, 1<<30-1), listLengths(0),
	))
	properties.TestingRun(t)
}

// listLengths is how long the list of a property may be, from the shortest
// given to the longest a list may be, with ten, twenty and forty, the lengths
// the lists are written to, drawn as often as all the others together.
func listLengths(shortest int) gopter.Gen {
	return gen.OneGenOf(gen.IntRange(shortest, profile.MaxIdeasPerList),
		gen.OneConstOf(tenIdeas, 20, profile.MaxIdeasPerList))
}

// roundsGiveEachIdeaOnce reports whether ideas named one after another from a
// list of that many fill their rounds in turn, a list's worth to a round, and
// give no idea twice in a round.
func roundsGiveEachIdeaOnce(named []profile.Idea, ideas int) bool {
	for start := 0; start < len(named); start += ideas {
		round := named[start:min(start+ideas, len(named))]
		numbers := make([]int, 0, len(round))
		for _, idea := range round {
			if idea.Round != start/ideas+1 {
				return false
			}
			numbers = append(numbers, idea.Number)
		}
		slices.Sort(numbers)
		if len(slices.Compact(numbers)) != len(round) || numbers[0] < 1 || numbers[len(numbers)-1] > ideas {
			return false
		}
	}
	return true
}

// noRoundOpensOnTheLast reports whether no round of ideas named one after
// another from a list of that many opens with the idea the round before it
// closed on.
func noRoundOpensOnTheLast(named []profile.Idea, ideas int) bool {
	for start := ideas; start < len(named); start += ideas {
		if named[start].Number == named[start-1].Number {
			return false
		}
	}
	return true
}

// namesTheSameIdeaReadBack reports whether the profile names the same idea of
// counting.gaps at 1-2, from a list of that many ideas, when asked again, and
// when read back from its file.
func namesTheSameIdeaReadBack(p *profile.Profile, ideas int) bool {
	next := tutor.NextIdea(p, "counting.gaps", rating.Grades12, ideas)
	file, err := profile.Marshal(p)
	if err != nil {
		return false
	}
	read, err := profile.Parse(file)
	return err == nil && tutor.NextIdea(p, "counting.gaps", rating.Grades12, ideas) == next &&
		tutor.NextIdea(read, "counting.gaps", rating.Grades12, ideas) == next
}

// namedFrom reports whether an idea is one a record of a round and the ideas
// given in it leaves on a list of that many: an idea of the list not given
// yet, in the same round, or, once the round has given every idea, one of the
// next round other than the idea given last, unless the list has no other. A
// list of none leaves no idea.
func namedFrom(idea profile.Idea, given []int, round, ideas int) bool {
	if ideas < 1 {
		return idea == profile.Idea{}
	}
	if !onTheList(idea, ideas) {
		return false
	}
	current := max(round, 1)
	if slices.ContainsFunc(listOf(ideas), func(n int) bool { return !slices.Contains(given, n) }) {
		return idea.Round == current && !slices.Contains(given, idea.Number)
	}
	return idea.Round == current+1 && (ideas == 1 || idea.Number != given[len(given)-1])
}

// onTheList reports whether an idea is a place on a list of that many ideas,
// named in a round from the first.
func onTheList(idea profile.Idea, ideas int) bool {
	return idea.Number >= 1 && idea.Number <= ideas && idea.Round >= 1
}

// listOf is every place on a list of that many ideas.
func listOf(ideas int) []int {
	places := make([]int, 0, ideas)
	for number := 1; number <= ideas; number++ {
		places = append(places, number)
	}
	return places
}

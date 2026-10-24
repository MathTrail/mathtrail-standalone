package profile_test

import (
	"bytes"
	"reflect"
	"slices"
	"testing"
	"time"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/profile"
	"github.com/MathTrail/mathtrail-standalone/internal/domain/rating"
)

// handOutOn asks for a task of counting.gaps at a level, built on an idea, and
// hands it out as an accepted task is handed out.
func handOutOn(t *testing.T, p *profile.Profile, sealer profile.Sealer, level rating.GradeLevel, idea profile.Idea) {
	t.Helper()

	brief := briefOn("counting.gaps")
	brief.GradeLevel = level
	p.Ask(&brief, profile.TutorRule, "en", asked).Idea = idea
	if _, err := p.Issue(written(), secret(), sealer, asked); err != nil {
		t.Fatalf("Issue() error = %v, want nil", err)
	}
}

// A task handed out gives the idea it was built on: the first idea of a list
// opens its first round, the next joins the round, an idea named in a later
// round opens that round alone, an idea the round has given already adds
// nothing, and one named in an earlier round joins the round the list is on.
// Each level keeps a round of its own, and a task built on no idea gives none,
// so the file holds no record of ideas at all.
func TestATaskHandedOutGivesItsIdea(t *testing.T) {
	t.Parallel()

	type handed struct {
		level rating.GradeLevel
		idea  profile.Idea
	}
	at12 := func(number, round int) handed {
		return handed{rating.Grades12, profile.Idea{Number: number, Round: round}}
	}
	for _, tc := range []struct {
		name   string
		handed []handed
		want   map[rating.GradeLevel]profile.IdeaRound
	}{
		{"the first idea of a list", []handed{at12(4, 1)},
			map[rating.GradeLevel]profile.IdeaRound{rating.Grades12: {Given: []int{4}, Round: 1}}},
		{"the next idea joins the round", []handed{at12(4, 1), at12(9, 1)},
			map[rating.GradeLevel]profile.IdeaRound{rating.Grades12: {Given: []int{4, 9}, Round: 1}}},
		{"an idea of a later round opens it", []handed{at12(4, 1), at12(9, 1), at12(9, 2)},
			map[rating.GradeLevel]profile.IdeaRound{rating.Grades12: {Given: []int{9}, Round: 2}}},
		{"an idea the round has given", []handed{at12(4, 1), at12(9, 1), at12(4, 1)},
			map[rating.GradeLevel]profile.IdeaRound{rating.Grades12: {Given: []int{4, 9}, Round: 1}}},
		{"an idea of an earlier round", []handed{at12(4, 3), at12(9, 2)},
			map[rating.GradeLevel]profile.IdeaRound{rating.Grades12: {Given: []int{4, 9}, Round: 3}}},
		{"a level of its own", []handed{at12(4, 1), {rating.Grades34, profile.Idea{Number: 4, Round: 1}}},
			map[rating.GradeLevel]profile.IdeaRound{
				rating.Grades12: {Given: []int{4}, Round: 1}, rating.Grades34: {Given: []int{4}, Round: 1},
			}},
		{"a task built on no idea", []handed{{rating.Grades12, profile.Idea{}}}, nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			p := parseFixture(t, "dima")
			sealer := newSealer(t)
			for _, h := range tc.handed {
				handOutOn(t, p, sealer, h.level, h.idea)
			}
			if got := p.Topics["counting.gaps"].Ideas; !reflect.DeepEqual(got, tc.want) {
				t.Errorf("the ideas given are %+v, want %+v", got, tc.want)
			}
			wantIdeasReadBack(t, p, tc.want)
		})
	}
}

// wantIdeasReadBack fails the test unless the file of the profile keeps the
// ideas given to counting.gaps as want says, and keeps no record at all when
// want has none.
func wantIdeasReadBack(t *testing.T, p *profile.Profile, want map[rating.GradeLevel]profile.IdeaRound) {
	t.Helper()

	file, err := profile.Marshal(p)
	if err != nil {
		t.Fatalf("Marshal() error = %v, want nil", err)
	}
	read, err := profile.Parse(file)
	if err != nil {
		t.Fatalf("Parse() error = %v, want the file read back", err)
	}
	if got := read.Topics["counting.gaps"].Ideas; !reflect.DeepEqual(got, want) {
		t.Errorf("read back, the ideas given are %+v, want %+v", got, want)
	}
	if want == nil && bytes.Contains(file, []byte(`"ideas"`)) {
		t.Error("the file keeps ideas for a topic no task of which was built on one")
	}
}

// A hand-out writes a new record of the ideas given rather than into the one
// the profile held: a copy of the profile taken before it keeps the record as
// it was.
func TestAHandOutLeavesARecordTakenBeforeItAsItWas(t *testing.T) {
	t.Parallel()

	p := parseFixture(t, "dima")
	sealer := newSealer(t)
	handOutOn(t, p, sealer, rating.Grades12, profile.Idea{Number: 4, Round: 1})
	before := p.Topics["counting.gaps"].Ideas
	handOutOn(t, p, sealer, rating.Grades12, profile.Idea{Number: 9, Round: 1})

	want := map[rating.GradeLevel]profile.IdeaRound{rating.Grades12: {Given: []int{4}, Round: 1}}
	if !reflect.DeepEqual(before, want) {
		t.Errorf("the record taken before the hand-out is %+v, want it as it was, %+v", before, want)
	}
}

// An idea is given only by a task that reaches the child. A request the checks
// refused to its last attempt, one replaced by another or let go, a task kept
// and let go, or dropped for a task asked for now — none of them reached the
// child, and none gives its idea. A task skipped or lost after it reached the
// child has given its idea already, and leaving takes nothing back.
func TestATaskThatNeverReachesTheChildGivesNoIdea(t *testing.T) {
	t.Parallel()

	idea := profile.Idea{Number: 6, Round: 1}
	ask := func(p *profile.Profile) *profile.OpenRequest {
		brief := briefOn("counting.gaps")
		request := p.Ask(&brief, profile.TutorRule, "en", asked)
		request.Idea = idea
		return request
	}
	keep := func(t *testing.T, p *profile.Profile) {
		t.Helper()
		_, kept := keepOne(t, p, newSealer(t), "counting.gaps")
		if kept.Idea == (profile.Idea{}) {
			t.Fatal("the task kept carries no idea, want the one its request named")
		}
	}
	for _, tc := range []struct {
		name   string
		before func(t *testing.T, p *profile.Profile)
		leave  func(t *testing.T, p *profile.Profile)
	}{
		{"a request refused to its last attempt", func(_ *testing.T, p *profile.Profile) { ask(p) },
			func(t *testing.T, p *profile.Profile) {
				t.Helper()
				for range profile.MaxAttempts {
					if _, err := p.Refuse(asked); err != nil {
						t.Fatalf("Refuse() error = %v, want nil", err)
					}
				}
			}},
		{"a request replaced by another", func(_ *testing.T, p *profile.Profile) { ask(p) },
			func(_ *testing.T, p *profile.Profile) { ask(p).Idea = profile.Idea{Number: 7, Round: 1} }},
		{"a request let go", func(_ *testing.T, p *profile.Profile) { ask(p) },
			func(_ *testing.T, p *profile.Profile) { p.CloseRequest() }},
		{"a task kept and let go", keep, func(_ *testing.T, p *profile.Profile) { p.DropReady() }},
		{"a task kept and dropped for one asked for now", keep, func(_ *testing.T, p *profile.Profile) { ask(p) }},
		{"a task on the card skipped", func(t *testing.T, p *profile.Profile) {
			t.Helper()
			handOutOn(t, p, newSealer(t), rating.Grades12, idea)
		}, func(_ *testing.T, p *profile.Profile) { p.Skip(asked.Add(time.Minute)) }},
		{"a task on the card lost", func(t *testing.T, p *profile.Profile) {
			t.Helper()
			handOutOn(t, p, newSealer(t), rating.Grades12, idea)
		}, func(_ *testing.T, p *profile.Profile) { p.DiscardTask() }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			p := parseFixture(t, "masha")
			tc.before(t, p)
			ideas := p.Topics["counting.gaps"].Ideas
			tc.leave(t, p)
			if got := p.Topics["counting.gaps"].Ideas; !reflect.DeepEqual(got, ideas) {
				t.Errorf("the ideas given are %+v, want them as they were, %+v", got, ideas)
			}
		})
	}
}

// Whatever ideas the tasks handed out were built on — none, the same twice,
// one of a round long past — the record of a list stays one a file can hold:
// each idea of its round once, every one a place on the list, on the latest
// round any of them was named in. And the file reads back as it was written.
func TestTheIdeasGivenHoldTheirProperties(t *testing.T) {
	t.Parallel()

	sealer := newSealer(t)
	anIdea := gopter.CombineGens(gen.IntRange(0, profile.MaxIdeasPerList), gen.IntRange(1, 4)).Map(
		func(values []any) profile.Idea {
			number, _ := values[0].(int)
			round, _ := values[1].(int)
			if number == 0 {
				return profile.Idea{} // a task whose request named no idea
			}
			return profile.Idea{Number: number, Round: round}
		})
	properties := gopter.NewProperties(nil)
	properties.Property("a list's record holds each idea of its round once, on the latest round, and reads back", prop.ForAll(
		func(ideas []profile.Idea) bool {
			p := parseFixture(t, "dima")
			latest := 0
			for _, idea := range ideas {
				handOutOn(t, p, sealer, rating.Grades12, idea)
				if idea.Number > 0 {
					latest = max(latest, idea.Round)
				}
			}
			kept, held := p.Topics["counting.gaps"].Ideas[rating.Grades12]
			if latest == 0 {
				return !held
			}
			return kept.Round == latest && eachOnceOnTheList(kept.Given) && readsBackAlike(p)
		},
		gen.SliceOfN(30, anIdea),
	))
	properties.TestingRun(t)
}

// eachOnceOnTheList reports whether ideas given in a round are some ideas of a
// list, each a place on it and none of them twice.
func eachOnceOnTheList(given []int) bool {
	sorted := slices.Sorted(slices.Values(given))
	return len(sorted) > 0 && len(slices.Compact(slices.Clone(sorted))) == len(sorted) &&
		sorted[0] >= 1 && sorted[len(sorted)-1] <= profile.MaxIdeasPerList
}

// readsBackAlike reports whether the profile's file reads back with the topics'
// ideas as they were written.
func readsBackAlike(p *profile.Profile) bool {
	file, err := profile.Marshal(p)
	if err != nil {
		return false
	}
	read, err := profile.Parse(file)
	return err == nil && reflect.DeepEqual(read.Topics["counting.gaps"].Ideas, p.Topics["counting.gaps"].Ideas)
}

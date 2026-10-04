package progress_test

import (
	"os"
	"path/filepath"
	"slices"
	"testing"
	"time"

	"github.com/MathTrail/mathtrail-standalone/content"
	"github.com/MathTrail/mathtrail-standalone/internal/domain/profile"
	"github.com/MathTrail/mathtrail-standalone/internal/domain/progress"
	"github.com/MathTrail/mathtrail-standalone/internal/domain/rating"
	"github.com/MathTrail/mathtrail-standalone/internal/domain/tutor"
)

// The fixtures are the prototype's seed profiles in this shape: Masha, Dima
// and Sasha are still in the trial series, Olya and Petya are past it, and
// Petya has topics mastered at the level of his grade.

func fixture(t *testing.T, student string) *profile.Profile {
	t.Helper()

	raw, err := os.ReadFile(filepath.Join("..", "..", "..", "testdata", "profiles", student+".json"))
	if err != nil {
		t.Fatalf("read the fixture of %s: %v", student, err)
	}
	p, err := profile.Parse(raw)
	if err != nil {
		t.Fatalf("parse the fixture of %s: %v", student, err)
	}
	return p
}

// embedded is the catalog the binary carries: the progress lists topics in
// its order and asks the rule about them.
func embedded(t *testing.T) *content.Content {
	t.Helper()

	loaded, err := content.Load()
	if err != nil {
		t.Fatalf("load the content: %v", err)
	}
	return loaded
}

// today is the day the summaries of these tests are asked for, after the
// latest answer of any fixture.
var today = profile.DateOf(time.Date(2026, time.October, 4, 12, 0, 0, 0, time.UTC))

func summaryOf(t *testing.T, p *profile.Profile, catalog tutor.Catalog) progress.Summary {
	t.Helper()

	summary, err := progress.Of(p, catalog, today)
	if err != nil {
		t.Fatalf("Of() error = %v, want nil", err)
	}
	return summary
}

// While the trial series runs it is still finding where the child stands, so
// there is no rating to show anywhere: only how many of its tasks are done.
func TestDuringTheTrialSeriesThereIsNoRatingToShow(t *testing.T) {
	t.Parallel()

	catalog := embedded(t)
	for _, tc := range []struct {
		student  string
		answered int
	}{
		{"sasha", 0},
		{"masha", 3},
		{"dima", 3},
	} {
		t.Run(tc.student, func(t *testing.T) {
			t.Parallel()

			summary := summaryOf(t, fixture(t, tc.student), catalog)
			want := progress.Trial{Answered: tc.answered, Of: rating.TrialAnswers}
			if summary.Trial == nil || *summary.Trial != want {
				t.Fatalf("trial = %+v, want %+v", summary.Trial, want)
			}
			if summary.Overall != nil {
				t.Errorf("overall = %+v during the trial series, want none", *summary.Overall)
			}
			wantNoTopicStanding(t, &summary)
		})
	}
}

// wantNoTopicStanding holds every topic listed to standing nowhere, as during
// the trial series, which is still finding where the child stands.
func wantNoTopicStanding(t *testing.T, summary *progress.Summary) {
	t.Helper()

	for _, topic := range summary.Topics {
		if topic.Standing != nil || topic.Compared != "" {
			t.Errorf("%s stands at %+v, %q, during the trial series, want nowhere", topic.ID, topic.Standing, topic.Compared)
		}
	}
}

// Once the series is over the child has one rating on the ladder, with its rank
// and how far through the rank it has come, and every topic answered stands
// somewhere of its own beside it.
func TestAfterTheTrialSeriesTheRatingHasItsRank(t *testing.T) {
	t.Parallel()

	catalog := embedded(t)
	for _, tc := range []struct {
		student string
		rank    int
		share   int
	}{
		// Olya stands just below where the youngest start, near the top of her
		// rank, and Petya a rank past the start of grades 3 and 4.
		{"olya", 2, 91},
		{"petya", 5, 92},
	} {
		t.Run(tc.student, func(t *testing.T) {
			t.Parallel()

			p := fixture(t, tc.student)
			summary := summaryOf(t, p, catalog)
			if summary.Trial != nil {
				t.Fatalf("trial = %+v, want the series over", *summary.Trial)
			}
			elo := rating.Elo(p.Ratings.Theta)
			want := progress.Standing{Rating: rating.Shown(elo), Rank: tc.rank, Share: tc.share}
			if summary.Overall == nil || *summary.Overall != want {
				t.Fatalf("overall = %+v, want %+v", summary.Overall, want)
			}
			wantTopicStandings(t, p, &summary)
		})
	}
}

// wantTopicStandings holds every topic answered to the standing of its own
// level — the overall one with the topic's correction — and every topic with
// no answer yet to none.
func wantTopicStandings(t *testing.T, p *profile.Profile, summary *progress.Summary) {
	t.Helper()

	for _, topic := range summary.Topics {
		if p.Topics[topic.ID].Answers == 0 {
			if topic.Standing != nil {
				t.Errorf("%s, with no answer, stands at %+v, want nowhere", topic.ID, *topic.Standing)
			}
			continue
		}
		elo := rating.Elo(p.Ratings.Theta + p.Topics[topic.ID].Delta)
		want := progress.Standing{Rating: rating.Shown(elo), Rank: rating.Rank(elo), Share: rating.Share(elo)}
		if topic.Standing == nil || *topic.Standing != want {
			t.Errorf("%s stands at %v, want %+v", topic.ID, topic.Standing, want)
		}
	}
}

// A topic is listed once the child has met it, or the rule could set it now,
// in the order of the catalog, with what the child made of it: the list grows
// as the child does.
func TestTheTopicsMetAndWithinReachAreListedInCatalogOrder(t *testing.T) {
	t.Parallel()

	catalog := embedded(t)
	p := fixture(t, "olya")
	summary := summaryOf(t, p, catalog)

	reachable := tutor.WithinReach(p, catalog)
	var want, unmet []string
	for _, id := range catalog.TopicIDs() {
		met := p.Topics[id].Answers > 0 || p.Topics[id].Skipped > 0
		if met || slices.Contains(reachable, id) {
			want = append(want, id)
		}
		if !met && slices.Contains(reachable, id) {
			unmet = append(unmet, id)
		}
	}
	if got := idsOf(summary.Topics); !slices.Equal(got, want) || len(unmet) == 0 {
		t.Errorf("topics = %v, want %v, %v of them not met yet", got, want, unmet)
	}
	for _, topic := range summary.Topics {
		kept := p.Topics[topic.ID]
		if topic.Answers != kept.Answers || topic.Correct != kept.Correct {
			t.Errorf("%s shows %d of %d right, want %d of %d",
				topic.ID, topic.Correct, topic.Answers, kept.Correct, kept.Answers)
		}
	}
}

// A child who has answered nothing yet is shown the topics the rule could set,
// none of them met.
func TestAChildWhoHasAnsweredNothingIsShownTheTopicsWithinReach(t *testing.T) {
	t.Parallel()

	catalog := embedded(t)
	p := fixture(t, "sasha")
	within := tutor.WithinReach(p, catalog)
	if got := idsOf(summaryOf(t, p, catalog).Topics); !slices.Equal(got, within) || len(within) == 0 {
		t.Errorf("topics = %v, want the %v within reach", got, within)
	}
}

// idsOf are the ids of topics, in their order.
func idsOf(topics []progress.Topic) []string {
	ids := make([]string, 0, len(topics))
	for _, topic := range topics {
		ids = append(ids, topic.ID)
	}
	return ids
}

// A topic's rank is compared with the overall one rank for rank, so that the
// word beside it agrees with the names drawn: a topic a little above the
// overall rating, but in the same rank, is even with it.
func TestATopicIsComparedByItsRank(t *testing.T) {
	t.Parallel()

	catalog := embedded(t)
	p := fixture(t, "petya")
	for id, delta := range map[string]float64{
		"arithmetic.tricks":         1,
		"combinatorics.enumeration": -1,
		"counting.gaps":             0.05,
	} {
		kept := p.Topics[id]
		kept.Delta = delta
		p.Topics[id] = kept
	}

	summary := summaryOf(t, p, catalog)
	if summary.Overall == nil || summary.Overall.Rank != 5 {
		t.Fatalf("overall = %+v, want Petya in rank 5 for this case", summary.Overall)
	}
	for id, want := range map[string]progress.Comparison{
		"arithmetic.tricks":         progress.Ahead,
		"combinatorics.enumeration": progress.Behind,
		"counting.gaps":             progress.Even,
	} {
		at := slices.IndexFunc(summary.Topics, func(topic progress.Topic) bool { return topic.ID == id })
		if at < 0 {
			t.Fatalf("%s is not among the topics %+v", id, summary.Topics)
		}
		if topic := summary.Topics[at]; topic.Compared != want {
			t.Errorf("%s at %+v is %q of the overall rank, want %q", id, topic.Standing, topic.Compared, want)
		}
	}
}

// "Mastered" on the screen is what the rule believes: a topic mastered at a
// level below the one its tasks now come from is back in the rotation, and is
// not shown as mastered either.
func TestAMasteryTheRuleNoLongerHoldsIsNotShown(t *testing.T) {
	t.Parallel()

	catalog := embedded(t)
	p := fixture(t, "petya")
	before := summaryOf(t, p, catalog)
	if !mastered(t, &before, "counting.gaps") {
		t.Fatal("counting.gaps is not shown as mastered at the level of its tasks, want it shown")
	}

	// A child well past the tasks of grades 1 and 2, with the topic mastered
	// among them.
	p.Ratings.Theta = 5
	below := rating.Grades12
	gaps := p.Topics["counting.gaps"]
	gaps.MasteredLevel = &below
	p.Topics["counting.gaps"] = gaps
	points := rating.Points(catalog.LevelsOf("counting.gaps")...)
	if now := rating.NewCorridor(p.Ratings.Theta+gaps.Delta, points).Recommended.GradeLevel; now == below {
		t.Fatalf("the tasks of counting.gaps still come from %s, want a level above it for this case", now)
	}
	after := summaryOf(t, p, catalog)
	if mastered(t, &after, "counting.gaps") {
		t.Error("counting.gaps mastered at 1-2 is shown as mastered where its tasks come from a level above, want it not")
	}
}

func mastered(t *testing.T, summary *progress.Summary, topic string) bool {
	t.Helper()

	for _, listed := range summary.Topics {
		if listed.ID == topic {
			return listed.Mastered
		}
	}
	t.Fatalf("%s is not among the topics %+v", topic, summary.Topics)
	return false
}

// What the progress says comes next is what the rule would set, point for
// point, and never the model's own choice.
func TestTheRecommendationIsTheRulesOwn(t *testing.T) {
	t.Parallel()

	catalog := embedded(t)
	for _, student := range []string{"masha", "olya", "petya", "sasha"} {
		t.Run(student, func(t *testing.T) {
			t.Parallel()

			p := fixture(t, student)
			brief, _, err := tutor.Next(p, catalog, tutor.Choice{})
			if err != nil {
				t.Fatalf("Next() error = %v", err)
			}
			want := progress.Recommendation{
				Topic:      brief.TargetConcept,
				GradeLevel: brief.GradeLevel,
				Difficulty: brief.Difficulty,
				Goal:       brief.PedagogicalGoal,
			}
			if got := summaryOf(t, p, catalog).Next; got != want {
				t.Errorf("next = %+v, want %+v", got, want)
			}
		})
	}
}

// The latest answer comes first, and the list is the summary's own: turning it
// round never turns the profile's window round with it.
func TestTheLatestAnswerComesFirst(t *testing.T) {
	t.Parallel()

	p := fixture(t, "masha")
	summary := summaryOf(t, p, embedded(t))

	if len(summary.Recent) != len(p.Recent) || len(p.Recent) < 2 {
		t.Fatalf("recent answers = %d, want the %d of the window", len(summary.Recent), len(p.Recent))
	}
	for i := range summary.Recent {
		if want := p.Recent[len(p.Recent)-1-i].TaskID; summary.Recent[i].TaskID != want {
			t.Errorf("recent[%d] is %s, want %s", i, summary.Recent[i].TaskID, want)
		}
	}
	summary.Recent[0].TaskID = "changed"
	if p.Recent[len(p.Recent)-1].TaskID == "changed" {
		t.Error("changing the summary changed the profile's window")
	}
}

// A skipped task is shown to the parent where it was skipped, among the latest
// answers, counted in its topic and in the total. A topic whose only tasks
// were skipped is listed for the count, standing nowhere: without an answer
// its rating would only be the overall one again.
func TestTheSkippedTasksAreShownToTheParent(t *testing.T) {
	t.Parallel()

	catalog := embedded(t)
	p := fixture(t, "olya")
	skippedOn := "geometry.grid"
	if p.Topics[skippedOn].Answers != 0 || p.Topics[skippedOn].Skipped != 0 {
		t.Fatalf("Olya has met %s already, and this needs a topic she has not", skippedOn)
	}
	p.CurrentTask = &profile.CurrentTask{ID: "tsk_left", Topic: skippedOn, GradeLevel: rating.Grades34, Difficulty: 2}
	p.Skip(time.Date(2026, 9, 6, 9, 0, 0, 0, time.UTC))

	summary := summaryOf(t, p, catalog)
	if latest := summary.Recent[0]; !latest.Skipped || latest.TaskID != "tsk_left" {
		t.Errorf("the latest entry is %+v, want the skipped task", latest)
	}
	at := slices.IndexFunc(summary.Topics, func(topic progress.Topic) bool { return topic.ID == skippedOn })
	if at < 0 {
		t.Fatalf("topics = %+v, want %s listed for its skipped task", summary.Topics, skippedOn)
	}
	if topic := summary.Topics[at]; topic.Skipped != 1 || topic.Answers != 0 || topic.Standing != nil || topic.Mastered {
		t.Errorf("%s shows %+v, want one skipped task, no answer and no standing", skippedOn, topic)
	}
	if summary.Skipped != 1 {
		t.Errorf("tasks left without an answer in all = %d, want the 1 skipped", summary.Skipped)
	}
}

// A catalog with no topic taught at any level leaves the rule nothing to
// suggest, and the progress says so rather than suggesting nothing.
func TestAProgressWithNothingToSuggestFails(t *testing.T) {
	t.Parallel()

	if _, err := progress.Of(fixture(t, "olya"), emptyCatalog{}, today); err == nil {
		t.Error("Of() error = nil for a catalog with no topics, want an error")
	}
}

type emptyCatalog struct{}

func (emptyCatalog) TopicIDs() []string                              { return nil }
func (emptyCatalog) LevelsOf(string) []rating.GradeLevel             { return nil }
func (emptyCatalog) TrapIDs() []string                               { return nil }
func (emptyCatalog) ExampleTraps(string, rating.GradeLevel) []string { return nil }

// A topic the catalog no longer has is not listed, and its skips still count
// in all: the total is every task the profile recorded as left without an
// answer.
func TestTheSkipsOfATopicNoLongerInTheCatalogCountInAll(t *testing.T) {
	t.Parallel()

	p := fixture(t, "olya")
	p.Topics["clocks.sundials"] = profile.Topic{Skipped: 2}

	summary := summaryOf(t, p, embedded(t))
	if slices.ContainsFunc(summary.Topics, func(topic progress.Topic) bool { return topic.ID == "clocks.sundials" }) {
		t.Errorf("topics = %+v, want clocks.sundials, which the catalog no longer has, left out", summary.Topics)
	}
	if summary.Skipped != 2 {
		t.Errorf("tasks left without an answer in all = %d, want the 2 of clocks.sundials", summary.Skipped)
	}
}

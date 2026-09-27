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

func summaryOf(t *testing.T, p *profile.Profile, catalog tutor.Catalog) progress.Summary {
	t.Helper()

	summary, err := progress.Of(p, catalog)
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
			for _, topic := range summary.Topics {
				if topic.Rating != nil {
					t.Errorf("%s has the rating %d during the trial series, want none", topic.ID, *topic.Rating)
				}
			}
		})
	}
}

// Once the series is over the child has one rating on the ladder, with the one
// rank there is above it, and every topic met has a rating of its own and no
// rank.
func TestAfterTheTrialSeriesTheRatingHasItsRank(t *testing.T) {
	t.Parallel()

	catalog := embedded(t)
	for _, tc := range []struct {
		student string
		rank    int
	}{
		// Olya stands just below where the youngest start, and Petya a rank
		// past the start of grades 3 and 4.
		{"olya", 2},
		{"petya", 5},
	} {
		t.Run(tc.student, func(t *testing.T) {
			t.Parallel()

			p := fixture(t, tc.student)
			summary := summaryOf(t, p, catalog)
			if summary.Trial != nil {
				t.Fatalf("trial = %+v, want the series over", *summary.Trial)
			}
			elo := rating.Elo(p.Ratings.Theta)
			want := progress.Standing{Rating: rating.Shown(elo), Rank: tc.rank}
			if summary.Overall == nil || *summary.Overall != want {
				t.Fatalf("overall = %+v, want %+v", summary.Overall, want)
			}
			wantTopicRatings(t, p, &summary)
		})
	}
}

// wantTopicRatings holds every topic listed to the rating of its own level: the
// overall one with the topic's correction.
func wantTopicRatings(t *testing.T, p *profile.Profile, summary *progress.Summary) {
	t.Helper()

	for _, topic := range summary.Topics {
		want := rating.Shown(rating.Elo(p.Ratings.Theta + p.Topics[topic.ID].Delta))
		if topic.Rating == nil || *topic.Rating != want {
			t.Errorf("%s has the rating %v, want %d", topic.ID, topic.Rating, want)
		}
	}
}

// A topic is listed once it has an answer behind it, in the order of the
// catalog, with what the child made of it. A topic never met has nothing to
// show.
func TestTheTopicsAnsweredAreListedInCatalogOrder(t *testing.T) {
	t.Parallel()

	catalog := embedded(t)
	p := fixture(t, "olya")
	summary := summaryOf(t, p, catalog)

	var got []string
	for _, topic := range summary.Topics {
		got = append(got, topic.ID)
		kept := p.Topics[topic.ID]
		if topic.Answers != kept.Answers || topic.Correct != kept.Correct {
			t.Errorf("%s shows %d of %d right, want %d of %d",
				topic.ID, topic.Correct, topic.Answers, kept.Correct, kept.Answers)
		}
	}
	var want []string
	for _, id := range catalog.TopicIDs() {
		if p.Topics[id].Answers > 0 {
			want = append(want, id)
		}
	}
	if !slices.Equal(got, want) || len(want) == 0 {
		t.Errorf("topics = %v, want %v", got, want)
	}

	if listed := summaryOf(t, fixture(t, "sasha"), catalog).Topics; len(listed) != 0 {
		t.Errorf("a child who has answered nothing shows the topics %+v, want none", listed)
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
// answers, and counted in its topic. A topic whose only tasks were skipped is
// listed for the count, with no rating: without an answer its rating would
// only be the overall one again.
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
	if topic := summary.Topics[at]; topic.Skipped != 1 || topic.Answers != 0 || topic.Rating != nil || topic.Mastered {
		t.Errorf("%s shows %+v, want one skipped task, no answer and no rating", skippedOn, topic)
	}
}

// A catalog with no topic taught at any level leaves the rule nothing to
// suggest, and the progress says so rather than suggesting nothing.
func TestAProgressWithNothingToSuggestFails(t *testing.T) {
	t.Parallel()

	if _, err := progress.Of(fixture(t, "olya"), emptyCatalog{}); err == nil {
		t.Error("Of() error = nil for a catalog with no topics, want an error")
	}
}

type emptyCatalog struct{}

func (emptyCatalog) TopicIDs() []string                              { return nil }
func (emptyCatalog) LevelsOf(string) []rating.GradeLevel             { return nil }
func (emptyCatalog) TrapIDs() []string                               { return nil }
func (emptyCatalog) ExampleTraps(string, rating.GradeLevel) []string { return nil }

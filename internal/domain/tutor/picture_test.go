package tutor_test

import (
	"errors"
	"slices"
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/profile"
	"github.com/MathTrail/mathtrail-standalone/internal/domain/rating"
	"github.com/MathTrail/mathtrail-standalone/internal/domain/tutor"
)

// A task solved from the picture holds its facts on the card alone, so it goes
// only to a chat known to draw the card, and the rule's own choice gives none
// during the trial series.

// figures is a catalog whose topic counting.figures is one all of whose tasks
// are solved from the picture.
type figures struct{ catalog }

func (figures) SolvedFromPicture(topic string) bool { return topic == "counting.figures" }

// withFigures is the three topics behind counting.figures, first in catalog
// order, so that the rule reaches it before any other whenever it may.
func withFigures() figures {
	c := threeTopics()
	c.topics = append([]string{"counting.figures"}, c.topics...)
	c.examples["counting.figures"] = []string{"missed_case", "off_by_one"}
	return figures{c}
}

// briefIn is the brief the rule builds for a chat, the model asking for
// choice.
func briefIn(t *testing.T, p *profile.Profile, c tutor.Catalog, choice tutor.Choice, chat profile.Chat) (
	profile.Brief, profile.TutorMode,
) {
	t.Helper()

	built, mode, err := tutor.Next(p, c, choice, chat)
	if err != nil {
		t.Fatalf("Next() error = %v, want nil", err)
	}
	return built, mode
}

// chats are both chats by the names a test tells them by.
var chats = map[string]profile.Chat{"a chat with cards": profile.ChatWithCards, "a chat of words": profile.ChatOfWords}

// A topic solved from the picture is the rule's in a chat known to draw the
// card, and its brief asks for a task solved from the picture; in a chat of
// words it is no topic of the catalog, and during the trial series the rule
// gives it in no chat. Every brief that is not on it asks for nothing of the
// kind.
func TestATopicSolvedFromThePictureIsTheRulesOnlyWhereTheCardIsDrawn(t *testing.T) {
	t.Parallel()

	c := withFigures()
	for _, test := range []struct {
		name    string
		learner func(*testing.T) *profile.Profile
		chat    profile.Chat
		figures bool
	}{
		{"past the series, a chat with cards", settled, profile.ChatWithCards, true},
		{"past the series, a chat of words", settled, profile.ChatOfWords, false},
		{"in the series, a chat with cards", child, profile.ChatWithCards, false},
		{"in the series, a chat of words", child, profile.ChatOfWords, false},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			got, _ := briefIn(t, test.learner(t), c, tutor.Choice{}, test.chat)
			if figures := got.TargetConcept == "counting.figures"; figures != test.figures {
				t.Errorf("topic = %s, want counting.figures: %v", got.TargetConcept, test.figures)
			}
			if got.FromPicture() != test.figures {
				t.Errorf("constraints = %v, want facts_on_the_picture: %v", got.Constraints, test.figures)
			}
		})
	}
}

// The model may name a topic solved from the picture, with a reason, wherever
// the card is known to be drawn, the trial series too: what the series keeps
// away is the rule's own choice. From a chat of words it is refused under a
// rule of its own, in the series and after it.
func TestATopicSolvedFromThePictureTheModelNamesNeedsTheCard(t *testing.T) {
	t.Parallel()

	c := withFigures()
	named := tutor.Choice{Topic: "counting.figures", Reason: "The adult asked for triangles to count."}
	for learnerName, learner := range map[string]func(*testing.T) *profile.Profile{"in the series": child, "past it": settled} {
		t.Run(learnerName+", a chat with cards", func(t *testing.T) {
			t.Parallel()

			got, mode := briefIn(t, learner(t), c, named, profile.ChatWithCards)
			if got.TargetConcept != "counting.figures" || !got.FromPicture() || mode != profile.TutorLLM {
				t.Errorf("brief on %s with %v, mode %s; want counting.figures solved from the picture, the model's",
					got.TargetConcept, got.Constraints, mode)
			}
		})
		t.Run(learnerName+", a chat of words", func(t *testing.T) {
			t.Parallel()

			_, _, err := tutor.Next(learner(t), c, named, profile.ChatOfWords)
			var refused *tutor.ChoiceError
			if !errors.As(err, &refused) || len(refused.Problems) != 1 || refused.Problems[0].Field != "topic" ||
				refused.Problems[0].Code != profile.CodeNeedsCard {
				t.Fatalf("Next() error = %v, want the topic refused as needs_card alone", err)
			}
		})
	}
}

// Lessons kept to a topic solved from the picture keep to it where the card is
// known to be drawn. In a chat of words the choice counts as none, as a topic
// the catalog no longer has does: the rule chooses, and the model may name
// another topic, which a choice that stands would refuse.
func TestLessonsKeptToATopicSolvedFromThePictureKeepToItWhereTheCardIsDrawn(t *testing.T) {
	t.Parallel()

	c := withFigures()
	p := keptTo(t, "counting.figures")
	if got, mode := briefIn(t, p, c, tutor.Choice{}, profile.ChatWithCards); got.TargetConcept != "counting.figures" ||
		mode != profile.TutorPerson || !got.FromPicture() {
		t.Errorf("a chat with cards: brief on %s, mode %s, constraints %v; want the topic chosen, solved from the picture",
			got.TargetConcept, mode, got.Constraints)
	}
	if got, mode := briefIn(t, p, c, tutor.Choice{}, profile.ChatOfWords); got.TargetConcept == "counting.figures" ||
		mode != profile.TutorRule || got.FromPicture() {
		t.Errorf("a chat of words: brief on %s, mode %s, constraints %v; want the rule's topic, in words",
			got.TargetConcept, mode, got.Constraints)
	}
	other := tutor.Choice{Topic: "time.clocks", Reason: "The adult asked for clocks."}
	if got, _ := briefIn(t, p, c, other, profile.ChatOfWords); got.TargetConcept != "time.clocks" {
		t.Errorf("a chat of words: the model's topic gave a brief on %s, want time.clocks", got.TargetConcept)
	}
	if got := tutor.LessonTopic(p, tutor.Shown(c, profile.ChatOfWords)); got != "" {
		t.Errorf("LessonTopic in a chat of words = %q, want none", got)
	}
}

// A failure on a topic solved from the picture is not worked over in a chat of
// words, nor is the topic on pause or probed there: every step of the rule
// reads the catalog the chat is shown.
func TestNoStepOfTheRuleTakesATopicSolvedFromThePictureToAChatOfWords(t *testing.T) {
	t.Parallel()

	c := withFigures()
	failed := settled(t)
	issued(failed, "counting.figures", day, 3)
	failedAt(failed, "counting.figures", 1)
	if got, _ := briefIn(t, failed, c, tutor.Choice{}, profile.ChatWithCards); got.TargetConcept != "counting.figures" ||
		got.PedagogicalGoal != profile.GoalReinforce {
		t.Fatalf("a chat with cards: brief on %s for %s, want the failure worked over, or this case proves nothing",
			got.TargetConcept, got.PedagogicalGoal)
	}
	if got, _ := briefIn(t, failed, c, tutor.Choice{}, profile.ChatOfWords); got.TargetConcept == "counting.figures" {
		t.Errorf("a chat of words: the failure on counting.figures was worked over, want another topic")
	}

	shown := tutor.Shown(c, profile.ChatOfWords)
	if got := shown.TopicIDs(); slices.Contains(got, "counting.figures") {
		t.Errorf("the topics of a chat of words = %v, want counting.figures left out", got)
	}
	if got := c.TopicIDs(); !slices.Contains(got, "counting.figures") {
		t.Errorf("the catalog's own topics = %v after a chat of words read them, want them left as they were", got)
	}
	if got := shown.LevelsOf("counting.figures"); len(got) != 0 {
		t.Errorf("the levels of counting.figures in a chat of words = %v, want none", got)
	}
	paused := settled(t)
	paused.Ratings.Answers = 20
	issued(paused, "counting.figures", day, 3)
	figures := paused.Topics["counting.figures"]
	figures.Delta = -3
	paused.Topics["counting.figures"] = figures
	if got := tutor.PausedTopics(paused, c); !slices.Contains(got, "counting.figures") {
		t.Fatalf("PausedTopics() = %v, want counting.figures on pause, or this case proves nothing", got)
	}
	if got := tutor.PausedTopics(paused, shown); slices.Contains(got, "counting.figures") {
		t.Errorf("PausedTopics() in a chat of words = %v, want counting.figures left out", got)
	}
}

// ideaList is a list of ideas at every level of every topic: so many of them,
// and those at the places given solved from the picture.
type ideaList struct {
	count    int
	pictured []int
}

func (l ideaList) IdeaCount(string, rating.GradeLevel) int { return l.count }

func (l ideaList) IdeaFromPicture(_ string, _ rating.GradeLevel, number int) bool {
	return slices.Contains(l.pictured, number)
}

// ideasGiven hands out tasks one after another, each on the idea IdeaFor draws
// for it in the chat, and is those ideas, in order, with the briefs they left.
func ideasGiven(t *testing.T, p *profile.Profile, list ideaList, chat profile.Chat, tasks int) ([]profile.Idea, []profile.Brief) {
	t.Helper()

	sealer, err := taskSeal()
	if err != nil {
		t.Fatalf("build the seal: %v", err)
	}
	var (
		ideas   []profile.Idea
		briefed []profile.Brief
	)
	for range tasks {
		brief := gapsBrief(rating.Grades12)
		request := p.Ask(&brief, profile.TutorRule, "en", day)
		request.Idea = tutor.IdeaFor(p, list, &request.Brief, chat)
		ideas, briefed = append(ideas, request.Idea), append(briefed, request.Brief)
		if _, err := p.Issue(&profile.Written{
			Wording: "a task", Options: map[string]string{"A": "1", "B": "2", "C": "3", "D": "4", "E": "5"},
			Hint: "hint", Fingerprint: "sketch", InstructionsVersion: "v",
		}, &profile.TaskSecret{Answer: "C", Solution: "the solution"}, sealer, day); err != nil {
			t.Fatalf("Issue() error = %v, want nil", err)
		}
	}
	return ideas, briefed
}

// An idea solved from the picture is drawn only for a chat known to draw the
// card, past the trial series, and the brief built on it asks for a task
// solved from the picture; a chat of words, and the series, pass over it, and
// it stays in its round for a chat that takes it.
func TestAnIdeaSolvedFromThePictureIsDrawnOnlyWhereTheCardIsDrawn(t *testing.T) {
	t.Parallel()

	list := ideaList{count: 10, pictured: []int{3, 7}}
	for name, chat := range chats {
		for _, series := range []bool{true, false} {
			t.Run(name+map[bool]string{true: ", in the series", false: ", past it"}[series], func(t *testing.T) {
				t.Parallel()

				p := childCalled(t, "stu_picture")
				if !series {
					p.Ratings.Answers = rating.TrialAnswers
				}
				ideas, briefs := ideasGiven(t, p, list, chat, list.count)
				wantRoundOf(t, list, ideas, briefs, chat.ShowsPictures() && !series)
			})
		}
	}
}

// wantRoundOf holds the ideas a round of a list gave, and the briefs they
// left, to a chat that takes the ideas solved from the picture, or to one that
// passes over them: each brief asks for a task solved from the picture when
// its idea is one, and a chat that passes over them gives the ideas in words
// alone, and then opens the next round.
func wantRoundOf(t *testing.T, list ideaList, ideas []profile.Idea, briefs []profile.Brief, takes bool) {
	t.Helper()

	drewPicture := false
	for i, idea := range ideas {
		pictured := slices.Contains(list.pictured, idea.Number)
		drewPicture = drewPicture || pictured
		if briefs[i].FromPicture() != pictured {
			t.Errorf("idea %d: brief constraints %v, want facts_on_the_picture: %v", idea.Number, briefs[i].Constraints, pictured)
		}
	}
	if drewPicture != takes {
		t.Errorf("ideas %v drew one solved from the picture: %v, want %v", ideas, drewPicture, takes)
	}
	if words := list.count - len(list.pictured); !takes && ideas[words].Round != 2 {
		t.Errorf("ideas %v: want a round of the %d ideas in words, then the next round", ideas, words)
	}
}

// A topic all of whose tasks are solved from the picture passes over none of
// its ideas, wherever its brief was built: the model may name it during the
// series, where the rule's own ideas would pass them over.
func TestEveryIdeaOfATopicSolvedFromThePictureIsDrawn(t *testing.T) {
	t.Parallel()

	p := child(t)
	brief := gapsBrief(rating.Grades12)
	brief.Constraints = []string{profile.FactsOnThePicture}
	every := ideaList{count: 10, pictured: []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}}
	if idea := tutor.IdeaFor(p, every, &brief, profile.ChatWithCards); idea.Number == 0 {
		t.Errorf("IdeaFor() = %+v, want an idea of the list", idea)
	}
	if !slices.Equal(brief.Constraints, []string{profile.FactsOnThePicture}) {
		t.Errorf("constraints = %v, want facts_on_the_picture once", brief.Constraints)
	}
}

// A request whose idea its list no longer has is given another of the kind
// its brief asked for: its package and its checks are that kind's already.
func TestAnIdeaDrawnAgainIsOfTheKindItsBriefAskedFor(t *testing.T) {
	t.Parallel()

	list := ideaList{count: 10, pictured: []int{3, 7}}
	words, pictured := gapsBrief(rating.Grades12), gapsBrief(rating.Grades12)
	pictured.Constraints = []string{profile.FactsOnThePicture}
	for _, id := range []string{"stu_a", "stu_b", "stu_c", "stu_d", "stu_e", "stu_f"} {
		p := childCalled(t, id)
		if got := tutor.IdeaAgain(p, list, &words); slices.Contains(list.pictured, got.Number) || got.Number == 0 {
			t.Errorf("%s: IdeaAgain(a brief in words) = %+v, want an idea in words", id, got)
		}
		if got := tutor.IdeaAgain(p, list, &pictured); !slices.Contains(list.pictured, got.Number) {
			t.Errorf("%s: IdeaAgain(a brief solved from the picture) = %+v, want an idea solved from it", id, got)
		}
	}
}

// Whatever a list marks and whatever its round holds, an idea a chat passes
// over is never drawn for it, the draw always names an idea while the list has
// one it takes, and a round of a chat of words gives each of the ideas in
// words once before any comes back.
func TestTheDrawOfAnIdeaHoldsItsPropertiesForEveryChat(t *testing.T) {
	t.Parallel()

	properties := gopter.NewProperties(nil)
	properties.Property("no idea a chat of words passes over is drawn, and none is drawn twice in a round",
		prop.ForAll(func(count int, marks []bool, id string) bool {
			list := ideaList{count: count}
			for place, marked := range marks {
				if marked && place < count {
					list.pictured = append(list.pictured, place+1)
				}
			}
			words := count - len(list.pictured)
			if words == 0 {
				return true
			}
			p := child(t)
			p.StudentID = "stu_" + id
			ideas, _ := ideasGiven(t, p, list, profile.ChatOfWords, words)
			seen := map[int]bool{}
			for _, idea := range ideas {
				if idea.Number == 0 || slices.Contains(list.pictured, idea.Number) || seen[idea.Number] {
					return false
				}
				seen[idea.Number] = true
			}
			return true
		},
			gen.IntRange(1, profile.MaxIdeasPerList),
			gen.SliceOfN(profile.MaxIdeasPerList, gen.Bool()),
			gen.Identifier(),
		))
	properties.TestingRun(t)
}

// Lessons kept to a topic solved from the picture are kept to none in a chat
// of words, so the rule's turns run there, probes among them, and the probe it
// sets is read as one in that chat: its line, its card and its answer, which
// lifts its topic off the pause. Where the card is drawn the lessons keep to
// the topic, and nothing is a probe.
func TestAProbeSetInAChatOfWordsIsOneThere(t *testing.T) {
	t.Parallel()

	c := withFigures()
	p := pausing(t, 11) // eleven tasks behind: the next takes a probe's turn
	p.Student.LessonTopic = "counting.figures"
	brief, mode := briefIn(t, p, c, tutor.Choice{}, profile.ChatOfWords)
	if brief.TargetConcept != "logic.ordering" || mode != profile.TutorRule {
		t.Fatalf("a chat of words: brief on %s by %s, want a probe of logic.ordering", brief.TargetConcept, mode)
	}
	p.CurrentTask = &profile.CurrentTask{ID: "tsk_probe", Topic: "logic.ordering", TutorMode: mode}
	if !tutor.JudgingOf(p, c, profile.ChatOfWords).Probe {
		t.Error("JudgingOf() in a chat of words is no probe, want the probe the rule set there")
	}
	if tutor.JudgingOf(p, c, profile.ChatWithCards).Probe {
		t.Error("JudgingOf() in a chat with cards is a probe, want none: the lessons keep to their topic there")
	}
}

// A request keeps the idea it named while its list holds that idea as one of
// the kind its brief asked for. A list changed since — the idea marked solved
// from the picture, or its mark taken away — names another of the brief's
// kind, and a list that holds none of that kind any more names one of either,
// rather than none a package could be built on.
func TestARequestKeepsItsIdeaWhileItIsOfItsBriefsKind(t *testing.T) {
	t.Parallel()

	words, pictured := gapsBrief(rating.Grades12), gapsBrief(rating.Grades12)
	pictured.Constraints = []string{profile.FactsOnThePicture}
	named := profile.Idea{Number: 3, Round: 1}
	for _, test := range []struct {
		name  string
		brief profile.Brief
		list  ideaList
		keeps bool
	}{
		{"in words, still in words", words, ideaList{count: 10, pictured: []int{7}}, true},
		{"in words, marked since", words, ideaList{count: 10, pictured: []int{3, 7}}, false},
		{"from the picture, still marked", pictured, ideaList{count: 10, pictured: []int{3, 7}}, true},
		{"from the picture, its mark taken away", pictured, ideaList{count: 10, pictured: []int{7}}, false},
		{"from the picture, no mark left", pictured, ideaList{count: 10}, false},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			request := &profile.OpenRequest{Brief: test.brief, Idea: named}
			got := tutor.RequestIdea(childCalled(t, "stu_request"), test.list, request)
			if kept := got == named; kept != test.keeps {
				t.Errorf("RequestIdea() = %+v, want the idea named kept: %v", got, test.keeps)
			}
			if got.Number < 1 || got.Number > test.list.count {
				t.Errorf("RequestIdea() = %+v, want an idea of the list", got)
			}
			if kind := slices.Contains(test.list.pictured, got.Number); len(test.list.pictured) > 0 &&
				kind != test.brief.FromPicture() {
				t.Errorf("RequestIdea() = %+v, solved from the picture: %v, want the brief's kind", got, kind)
			}
		})
	}
}

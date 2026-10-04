package mcpserver_test

import (
	"slices"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/MathTrail/mathtrail-standalone/content"
	"github.com/MathTrail/mathtrail-standalone/internal/domain/profile"
	"github.com/MathTrail/mathtrail-standalone/internal/domain/progress"
	"github.com/MathTrail/mathtrail-standalone/internal/domain/rating"
	"github.com/MathTrail/mathtrail-standalone/internal/store"
	mcpserver "github.com/MathTrail/mathtrail-standalone/internal/transport/mcp"
)

// The child or the adult may keep the lessons to a topic: on the card of a
// task, which saves it with edit_profile, or by asking the model, which saves
// it with save_profile. Once the trial series is over every task is on it, and
// the card of a task is told what it needs to offer the choice.

// lessonTopicPayload is as much of a profile's answer as the topic of the
// lessons is told in: whether the change was written, why not, and the topic
// the details now name.
type lessonTopicPayload struct {
	Status   string           `json:"status"`
	Code     string           `json:"code"`
	Problems []problemPayload `json:"problems"`
	Changed  bool             `json:"changed"`
	Profile  *struct {
		LessonTopic *string `json:"lesson_topic"`
	} `json:"profile"`
}

// choicePayload is as much of what read_task tells a card as its choice of the
// topic: the one chosen, the ones the review suggests, and the site.
type choicePayload struct {
	Screen      string `json:"screen"`
	TopicChoice *struct {
		Chosen      *string  `json:"chosen"`
		Recommended []string `json:"recommended"`
		Site        *struct {
			URL       string   `json:"url"`
			Languages []string `json:"languages"`
		} `json:"site"`
	} `json:"topic_choice"`
}

// keptTo is the fixture of student, the lessons kept to topic.
func keptTo(t *testing.T, student, topic string) store.Storage {
	t.Helper()

	return keptAs(t, student, func(p *profile.Profile) { p.Student.LessonTopic = topic })
}

// wantSaid holds words to saying every one of these.
func wantSaid(t *testing.T, words string, say ...string) {
	t.Helper()

	for _, part := range say {
		if !strings.Contains(words, part) {
			t.Errorf("the words %q do not say %q", words, part)
		}
	}
}

func TestTheCardKeepsTheLessonsToATopicAndGivesItBack(t *testing.T) {
	t.Parallel()

	kept := keptWith(t, "olya")
	_, session := lesson(t, kept)

	chosen := call(t, session, "edit_profile", map[string]any{"lesson_topic": " time.clocks "})
	got := payloadOf[lessonTopicPayload](t, chosen)
	if !got.Changed || got.Profile == nil || got.Profile.LessonTopic == nil || *got.Profile.LessonTopic != "time.clocks" {
		t.Errorf("payload = %+v, want the lessons kept to time.clocks", got)
	}
	words := textOf(t, chosen)
	wantSaid(t, words, "On the card, the child or the adult chose the topic of the lessons.",
		"keeps the lessons to Clocks", "Do not explain the choice")
	if strings.Contains(words, "with the form") {
		t.Errorf("the words %q tell of the form, want the choice on the card of a task", words)
	}

	back := call(t, session, "edit_profile", map[string]any{"lesson_topic": ""})
	got = payloadOf[lessonTopicPayload](t, back)
	if !got.Changed || got.Profile == nil || got.Profile.LessonTopic != nil {
		t.Errorf("payload = %+v, want the choice given back to the rule", got)
	}
	wantSaid(t, textOf(t, back), "given back to the rule")

	_, revision := loadKept(t, kept)
	refused := payloadOf[lessonTopicPayload](t, call(t, session, "edit_profile", map[string]any{"lesson_topic": "astronomy.stars"}))
	if refused.Status != "rejected" || refused.Code != "invalid_profile" || len(refused.Problems) != 1 ||
		refused.Problems[0].Field != "lesson_topic" || refused.Problems[0].Code != profile.CodeNotInCatalog ||
		strings.Contains(refused.Problems[0].Rule, "astronomy") {
		t.Errorf("payload = %+v, want lesson_topic refused as not_in_catalog, without repeating it", refused)
	}
	if _, now := loadKept(t, kept); now != revision {
		t.Error("a topic the catalog does not have was written, want nothing written")
	}
}

// Asked by the child or the adult, the model saves the topic too. During the
// trial series the choice is kept and waits; after it, it is what comes next.
func TestTheModelKeepsTheLessonsToATopicWhenAsked(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		student string
		say     []string
	}{
		{"masha", []string{"once the trial series is over: until then each task is on a new topic"}},
		{"olya", []string{"keeps the lessons to Clocks", "Next, the topic chosen for the lessons, Clocks"}},
	} {
		t.Run(tc.student, func(t *testing.T) {
			t.Parallel()

			kept := keptWith(t, tc.student)
			_, session := lesson(t, kept)
			wantSaid(t, textOf(t, call(t, session, "save_profile", map[string]any{"lesson_topic": "time.clocks"})), tc.say...)
			if p, _ := loadKept(t, kept); p.Student.LessonTopic != "time.clocks" {
				t.Errorf("lesson_topic = %q, want time.clocks kept", p.Student.LessonTopic)
			}
		})
	}
}

// Once the trial series is over, the task asked for is on the topic the
// lessons are kept to, chosen by a person — and the line about the request
// says so.
func TestATaskIsAskedOnTheTopicTheLessonsAreKeptTo(t *testing.T) {
	t.Parallel()

	kept := keptTo(t, "olya", "time.clocks")
	h, session := lesson(t, kept)

	asked := call(t, session, "next_task", map[string]any{"language": "en"})
	p, _ := loadKept(t, kept)
	if p.OpenRequest == nil || p.OpenRequest.Brief.TargetConcept != "time.clocks" ||
		p.OpenRequest.TutorMode != profile.TutorPerson {
		t.Fatalf("open request = %+v, want one on time.clocks, chosen by a person", p.OpenRequest)
	}
	wantSaid(t, textOf(t, asked), "keeps the lessons to Clocks")
	lines := linesOf(h, "task_requested")
	if len(lines) != 1 || lines[0].ContextMap()["tutor_mode"] != "person" {
		t.Errorf("task_requested lines = %v, want one saying tutor_mode person", lines)
	}
}

// While the lessons are kept to a topic, the model may name that one — with no
// reason, since it asks for nothing — but not another, which is refused with
// the chosen one named and nothing written.
func TestAModelsTopicBesideTheTopicOfTheLessons(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name      string
		arguments map[string]any
		refused   []string
	}{
		{"another topic", map[string]any{"language": "en", "topic": "logic.ordering", "reason": "a change"},
			[]string{"topic not_one_of"}},
		{"the same topic, with no reason", map[string]any{"language": "en", "topic": "time.clocks"}, nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			kept := keptTo(t, "olya", "time.clocks")
			_, session := lesson(t, kept)
			_, revision := loadKept(t, kept)

			got := payloadOf[requestPayload](t, call(t, session, "next_task", tc.arguments))
			problems := make([]string, 0, len(got.Problems))
			for _, problem := range got.Problems {
				problems = append(problems, problem.Field+" "+problem.Code)
			}
			p, now := loadKept(t, kept)
			switch {
			case tc.refused != nil && (got.Code != "invalid_arguments" || !slices.Equal(problems, tc.refused) || now != revision):
				t.Errorf("next_task = %+v, want %v refused and nothing written", got, tc.refused)
			case tc.refused == nil && (p.OpenRequest == nil || p.OpenRequest.TutorMode != profile.TutorPerson):
				t.Errorf("next_task = %+v and the request %+v, want one opened on the chosen topic", got, p.OpenRequest)
			}
		})
	}
}

// Asked again while its request is open, the topic the lessons are kept to —
// with a reason for it alone — asks for nothing the request does not have, so
// nothing is said to be set aside.
func TestAskingAgainForTheTopicOfTheLessonsSetsNothingAside(t *testing.T) {
	t.Parallel()

	_, session := lesson(t, keptTo(t, "olya", "time.clocks"))
	call(t, session, "next_task", map[string]any{"language": "en"})
	again := textOf(t, call(t, session, "next_task",
		map[string]any{"language": "en", "topic": "time.clocks", "reason": "the child chose clocks"}))
	if !strings.Contains(again, "is already open") || strings.Contains(again, "were not applied") {
		t.Errorf("next_task again = %q, want the request handed back and nothing set aside", again)
	}
}

// A topic chosen while a request on another is open comes with the task after
// it: the request is handed back as it was, and the model is told that its
// topic was set aside and why.
func TestAskingForTheTopicOfTheLessonsOverAnotherRequestSetsItAside(t *testing.T) {
	t.Parallel()

	kept := keptWith(t, "olya")
	_, session := lesson(t, kept)
	call(t, session, "next_task", map[string]any{"language": "en"})
	p, _ := loadKept(t, kept)
	lesson := "time.clocks"
	if p.OpenRequest.Brief.TargetConcept == lesson {
		lesson = "logic.ordering"
	}
	call(t, session, "save_profile", map[string]any{"lesson_topic": lesson})

	again := textOf(t, call(t, session, "next_task", map[string]any{"language": "en", "topic": lesson}))
	wantSaid(t, again, "is already open", "were not applied", "stays on its own topic")
}

// The card of a task is told the choice of the topic once the task is on it
// and the trial series is over: the topic chosen, if any; the topics the
// review's steps are for, as the progress tells them; and the site. A task of
// the trial series carries none, and the card offers no choice on it.
func TestACardIsToldTheChoiceOfTheTopicOnceTheSeriesIsOver(t *testing.T) {
	t.Parallel()

	ordering := "logic.ordering"
	for _, tc := range []struct {
		name    string
		kept    func(t *testing.T) store.Storage
		offered bool
		chosen  *string
	}{
		{"in the trial series", func(t *testing.T) store.Storage { return keptTo(t, "masha", ordering) }, false, nil},
		{"past it, the rule choosing", func(t *testing.T) store.Storage { return keptAs(t, "petya", developing) }, true, nil},
		{"past it, the lessons kept to a topic", func(t *testing.T) store.Storage {
			return keptAs(t, "petya", func(p *profile.Profile) { developing(p); p.Student.LessonTopic = ordering })
		}, true, &ordering},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			kept := tc.kept(t)
			_, session := lesson(t, kept)
			request := raceHandedOut(t, session, kept)

			got := payloadOf[choicePayload](t, call(t, session, "read_task", map[string]any{"request_id": request.ID}))
			if got.Screen != "task" || (got.TopicChoice != nil) != tc.offered {
				t.Fatalf("read_task = %+v, want the task with a choice of the topic %v", got, tc.offered)
			}
			if tc.offered {
				wantChoice(t, &got, tc.chosen, suggestedByTheProgress(t, session))
			}
		})
	}
}

// A catalog the rule can choose no topic from is the service's own fault: the
// card of a task is not drawn without its choice of the topic, nor the
// progress without its review, and the model is told that something went
// wrong inside MathTrail.
func TestNoTopicToChooseFromIsAFailure(t *testing.T) {
	t.Parallel()

	kept := keptAs(t, "petya", developing)
	_, session := lesson(t, kept)
	request := raceHandedOut(t, session, kept)

	parts := allParts(t)
	parts.Store, parts.Content = kept, &content.Content{}
	service, err := mcpserver.NewService(parts)
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}
	h := newHarness(t)
	h.start(t, mcpserver.DevSignIn, slices.Concat(service.ProfileTools(), service.TaskTools())...)
	empty := h.connect(t, "")

	for _, tc := range []struct {
		tool string
		args map[string]any
	}{
		{"read_task", map[string]any{"request_id": request.ID}},
		{"get_progress", nil},
	} {
		wantOurSentence(t, call(t, empty, tc.tool, tc.args), "Something went wrong inside MathTrail.")
	}
}

// developing leaves Enumeration as the review's topic to develop: answered
// often enough to judge, and well below where the child stands overall.
func developing(p *profile.Profile) {
	topic := p.Topics["combinatorics.enumeration"]
	topic.Answers, topic.Correct, topic.Delta = progress.JudgedAnswers+1, 2, -2*rating.CorridorWidth()
	p.Topics["combinatorics.enumeration"] = topic
}

// raceHandedOut asks for the race and hands it in, and is the request whose
// task is now on the child's card.
func raceHandedOut(t *testing.T, session *mcp.ClientSession, kept store.Storage) *profile.OpenRequest {
	t.Helper()

	request := askForTheRace(t, session, kept)
	handed := call(t, session, "submit_task", raceOn(request))
	if p, _ := loadKept(t, kept); p.CurrentTask == nil || p.OpenRequest != nil {
		t.Fatalf("submit_task said %q, want the race on the card", textOf(t, handed))
	}
	return request
}

// wantChoice holds the choice of the topic a card was told to the topic
// chosen, the site, and the topics the progress's review suggests.
func wantChoice(t *testing.T, got *choicePayload, chosen *string, suggested []string) {
	t.Helper()

	choice := got.TopicChoice
	if (choice.Chosen == nil) != (chosen == nil) || (choice.Chosen != nil && *choice.Chosen != *chosen) {
		t.Errorf("chosen = %v, want %v", choice.Chosen, chosen)
	}
	if choice.Site == nil || choice.Site.URL == "" || !slices.Equal(choice.Site.Languages, []string{"en", "ru"}) {
		t.Errorf("site = %+v, want the site's origin and its languages", choice.Site)
	}
	if !slices.Equal(choice.Recommended, suggested) {
		t.Errorf("recommended = %v, want the topics of the review's steps, %v", choice.Recommended, suggested)
	}
}

// suggestedByTheProgress are the topics the steps of the review are for, in
// their order, each once and three at most, as read_progress tells them.
func suggestedByTheProgress(t *testing.T, session *mcp.ClientSession) []string {
	t.Helper()

	read := payloadOf[struct {
		Review *struct {
			Steps []struct {
				Topic string `json:"topic"`
			} `json:"steps"`
		} `json:"review"`
	}](t, call(t, session, "read_progress", nil))
	suggested := []string{}
	if read.Review == nil {
		return suggested
	}
	for _, step := range read.Review.Steps {
		if step.Topic != "" && !slices.Contains(suggested, step.Topic) && len(suggested) < progress.SuggestedTopics {
			suggested = append(suggested, step.Topic)
		}
	}
	return suggested
}

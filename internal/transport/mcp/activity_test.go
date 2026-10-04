package mcpserver_test

import (
	"context"
	"errors"
	"slices"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/profile"
	"github.com/MathTrail/mathtrail-standalone/internal/domain/rating"
	"github.com/MathTrail/mathtrail-standalone/internal/domain/tutor"
	"github.com/MathTrail/mathtrail-standalone/internal/infra/drive/drivetest"
	"github.com/MathTrail/mathtrail-standalone/internal/report"
	"github.com/MathTrail/mathtrail-standalone/internal/store"
	"github.com/MathTrail/mathtrail-standalone/internal/store/memory"
	mcpserver "github.com/MathTrail/mathtrail-standalone/internal/transport/mcp"
)

// The lines a child is counted from: a task handed out, an answer and a topic
// mastered each name the child by the name it is counted under that month,
// and the task handed out says where the family is.

// A task handed out is counted by its child's name for the month, the chat
// host it was handed out in, its language and the grade, the month the child
// started, and where the family is: the country and the region the parent
// gave, unknown when they gave none and other when the file, edited by hand,
// names a code the list does not have — and the country of the sign-in, which
// the development sign-in has none of.
func TestATaskHandedOutIsCountedWhereTheFamilyIs(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name                   string
		country, region        string
		wantCountry, wantState string
	}{
		{"a profile that names no place", "", "", "unknown", "unknown"},
		{"a country and its state", "US", "US-TX", "US", "US-TX"},
		{"a country with no regions", "FR", "", "FR", "unknown"},
		{"codes typed into the file by hand", "ZZ", "US-ZZ", "other", "other"},
		{"a state of another country", "FR", "US-TX", "FR", "other"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			p := profile.New(profile.Student{
				Grade: 2, Pseudonym: "Otter", Interests: []string{"sport"}, Country: tc.country, Region: tc.region,
			}, "test", lessonDay)
			kept := keptAsIs(t, p)
			h, session := lesson(t, kept)
			wantOnTheCard(t, call(t, session, "submit_task", raceOn(askForTheRace(t, session, kept))))
			h.settle()

			fields := theOnlyLine(t, h, "task_accepted")
			for field, want := range map[string]any{
				"learner": learners(t).Of(p.StudentID, lessonDay), "host": "claude", "language": "en",
				"grade": int64(2), "cohort": "2026-09", "country": tc.wantCountry, "region": tc.wantState,
				"signin_country": "unknown",
			} {
				if got := fields[field]; got != want {
					t.Errorf("task_accepted %s = %v, want %v", field, got, want)
				}
			}
			wantCounted(t, "task_accepted", fields)
		})
	}
}

// The country a parent signed in from travels with the account the sign-in
// lets a call in as, and the task handed out is counted by it.
func TestATaskHandedOutIsCountedByTheCountryOfItsSignIn(t *testing.T) {
	t.Parallel()

	account := store.NewAccount(vouchedUser, "", time.Time{})
	account.SignInCountry = "NZ"
	kept := memory.New()
	if _, err := kept.Create(context.Background(), account, profile.New(profile.Student{
		Grade: 2, Pseudonym: "Otter", Interests: []string{"sport"},
	}, "test", lessonDay)); err != nil {
		t.Fatalf("keep the profile: %v", err)
	}
	h := newHarness(t)
	service := lessonService(t, h, kept, &clock{at: lessonDay}, nil)
	h.start(t, mcpserver.BearerSignIn(readerOf(account), metadata), slices.Concat(service.ProfileTools(), service.TaskTools())...)
	session, err := h.connectWith(t, vouchedToken)
	if err != nil {
		t.Fatalf("connectWith() error = %v, want nil", err)
	}

	asked := call(t, session, "next_task", raceChoice)
	p, _, err := kept.Load(context.Background(), account)
	if err != nil || p.OpenRequest == nil {
		t.Fatalf("next_task = %q, and the profile holds no request: %v", textOf(t, asked), err)
	}
	wantOnTheCard(t, call(t, session, "submit_task", raceOn(p.OpenRequest)))
	h.settle()

	if got := theOnlyLine(t, h, "task_accepted")["signin_country"]; got != "NZ" {
		t.Errorf("task_accepted signin_country = %v, want NZ", got)
	}
}

// A child is counted under one name all month, whoever signs in for it: after
// the sealing keys are rotated, the parent's next sign-in gives their account
// another identifier, and the child the same name. The next month the child
// has another name, which the month before's does not lead to.
func TestAChildIsCountedUnderOneNameAMonth(t *testing.T) {
	t.Parallel()

	fake := drivetest.New(t)
	before := instanceOverDriveReading(t, fake, readerOf(store.NewAccount("a-user-before-it", googleToken, time.Time{})))
	after := instanceOverDriveReading(t, fake, readerOf(store.NewAccount("a-user-after-it", googleToken, time.Time{})))

	p := raceOnTheCard(t, rating.TrialAnswers)
	if _, err := before.kept.Create(context.Background(), parent, p); err != nil {
		t.Fatalf("keep the profile: %v", err)
	}
	answerIt(t, before.session, p.CurrentTask.ID, "C", false)
	answerIt(t, after.session, raceAgain(t, after.kept, parent, after.moving.now()), "C", false)
	after.moving.advance(31 * 24 * time.Hour)
	answerIt(t, after.session, raceAgain(t, after.kept, parent, after.moving.now()), "C", false)
	before.h.settle()
	after.h.settle()

	first := theOnlyLine(t, before.h, "answer_recorded")["learner"]
	afterwards := linesOf(after.h, "answer_recorded")
	if len(afterwards) != 2 {
		t.Fatalf("answer_recorded lines after the rotation = %d, want 2", len(afterwards))
	}
	if want := learners(t).Of(p.StudentID, lessonDay); first != want {
		t.Errorf("the child is counted as %v, want %q, the name of its identifier in September", first, want)
	}
	if again := afterwards[0].ContextMap()["learner"]; again != first {
		t.Errorf("after a new sign-in the child is counted as %v, want %v as before it", again, first)
	}
	if next := afterwards[1].ContextMap()["learner"]; next == first {
		t.Errorf("the next month the child is counted as %v, want another name", next)
	}
}

// A topic is counted as mastered by the answer that brings it among the
// topics the progress calls mastered, once each time: mastered, lost to two
// wrong answers in a row, and mastered again is two lines. Every answer says
// how many topics are mastered once it is in, as the progress counts them.
func TestATopicIsCountedAsMasteredEachTimeItIsWon(t *testing.T) {
	t.Parallel()

	p := raceOnTheCard(t, rating.TrialAnswers)
	// A child the race is hard enough for to count towards a topic's run, one
	// right answer short of mastering it.
	p.Ratings.Theta = -1
	p.Topics["logic.ordering"] = profile.Topic{Answers: 4, Correct: 4, TopStreak: profile.MasteryStreak - 1}
	kept := keptAsIs(t, p)
	h, session := lesson(t, kept)

	var shown []int
	answer := func(letter string) {
		t.Helper()
		answerIt(t, session, loadedOf(t, kept).CurrentTask.ID, letter, false)
		shown = append(shown, masteredIn(t, call(t, session, "read_progress", map[string]any{})))
		raceAgain(t, kept, devAccount, lessonDay)
	}
	answer("C")                          // the run completes: mastered
	answer("A")                          // a slip: still mastered
	answer("A")                          // two wrong in a row: lost
	answer("C")                          // a run begins again
	answer("C")                          //
	answer("C")                          // and completes: mastered again
	wantShown := []int{1, 1, 0, 0, 0, 1} // the topics the progress calls mastered, after each answer
	if !slices.Equal(shown, wantShown) {
		t.Fatalf("the progress shows %v topics mastered after each answer, want %v", shown, wantShown)
	}
	h.settle()

	recorded := linesOf(h, "answer_recorded")
	if len(recorded) != len(wantShown) {
		t.Fatalf("answer_recorded lines = %d, want %d", len(recorded), len(wantShown))
	}
	for i := range recorded {
		fields := recorded[i].ContextMap()
		if got := fields["topics_mastered"]; got != int64(wantShown[i]) {
			t.Errorf("answer %d says %v topics are mastered, want %d as the progress shows", i+1, got, wantShown[i])
		}
		wantCounted(t, "answer_recorded", fields)
	}
	mastered := linesOf(h, "topic_mastered")
	if len(mastered) != 2 {
		t.Fatalf("topic_mastered lines = %d, want one for each time the topic was won", len(mastered))
	}
	for i := range mastered {
		fields := mastered[i].ContextMap()
		if fields["topic"] != "logic.ordering" || fields["grade"] != int64(2) ||
			fields["learner"] != learners(t).Of(p.StudentID, lessonDay) {
			t.Errorf("topic_mastered = %v, want the race's topic, the grade and the child's name", fields)
		}
		wantCounted(t, "topic_mastered", fields)
	}
}

// A run that earns a topic its mastery is counted only when it brings the
// topic among those the progress calls mastered: not a run completed below
// the level the topic's tasks come from now, which the progress does not call
// mastered, and not one completed above the level the topic is shown mastered
// at, which it already was.
func TestOnlyATopicTheProgressComesToCallMasteredIsCounted(t *testing.T) {
	t.Parallel()

	shownAt := rating.Grades12
	for _, tc := range []struct {
		name         string
		theta, delta float64
		level        rating.GradeLevel
		difficulty   int
		masteredAt   *rating.GradeLevel
		wantShown    int
	}{
		{"a run completed below the level the topic's tasks come from now", 0.2, 1, rating.Grades12, 4, nil, 0},
		{"a run completed above the level the topic is shown mastered at", -1, 0, rating.Grades34, 1, &shownAt, 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			p := profile.New(profile.Student{Grade: 2, Pseudonym: "Otter", Interests: []string{"sport"}}, "test", lessonDay)
			p.Ratings.Answers, p.Ratings.Theta = rating.TrialAnswers, tc.theta
			topic := profile.Topic{Answers: 4, Correct: 4, TopStreak: profile.MasteryStreak - 1, Delta: tc.delta}
			if tc.masteredAt != nil {
				since := profile.DateOf(lessonDay.AddDate(0, 0, -3))
				topic.MasteredSince, topic.MasteredLevel = &since, tc.masteredAt
			}
			p.Topics["logic.ordering"] = topic
			askForTheRaceAt(t, p, tc.level, tc.difficulty, lessonDay)
			handOutTheRace(t, p, lessonDay)
			kept := keptAsIs(t, p)
			h, session := lesson(t, kept)

			answerIt(t, session, p.CurrentTask.ID, "C", false)
			h.settle()

			if won := loadedOf(t, kept).Topics["logic.ordering"].MasteredLevel; won == nil || *won != tc.level {
				t.Fatalf("the topic is mastered at %v, want the run to have mastered it at %s", won, tc.level)
			}
			if lines := linesOf(h, "topic_mastered"); len(lines) != 0 {
				t.Errorf("topic_mastered lines = %d, want none", len(lines))
			}
			if got := theOnlyLine(t, h, "answer_recorded")["topics_mastered"]; got != int64(tc.wantShown) {
				t.Errorf("answer_recorded topics_mastered = %v, want %d", got, tc.wantShown)
			}
		})
	}
}

// A topic mastered below the level its tasks came from comes back among the
// topics the progress calls mastered when a wrong answer brings the child's
// level in it back down. Nothing was won: the count of topics mastered goes
// up, and no line says the topic was mastered.
func TestATopicBackInViewIsNotCountedAsWon(t *testing.T) {
	t.Parallel()

	p := profile.New(profile.Student{Grade: 2, Pseudonym: "Otter", Interests: []string{"sport"}}, "test", lessonDay)
	p.Ratings.Answers, p.Ratings.Theta = rating.TrialAnswers, -0.65
	since, below := profile.DateOf(lessonDay.AddDate(0, 0, -3)), rating.Grades12
	p.Topics["logic.ordering"] = profile.Topic{Answers: 8, Correct: 7, Delta: 1.9, MasteredSince: &since, MasteredLevel: &below}
	askForTheRaceAt(t, p, rating.Grades12, 1, lessonDay)
	handOutTheRace(t, p, lessonDay)
	kept := keptAsIs(t, p)
	h, session := lesson(t, kept)
	shownBefore := masteredIn(t, call(t, session, "read_progress", map[string]any{}))

	answerIt(t, session, p.CurrentTask.ID, "A", false)
	shownAfter := masteredIn(t, call(t, session, "read_progress", map[string]any{}))
	h.settle()

	if shownBefore != 0 || shownAfter != 1 {
		t.Fatalf("the progress shows %d topics mastered before the answer and %d after, want 0 and 1", shownBefore, shownAfter)
	}
	if lines := linesOf(h, "topic_mastered"); len(lines) != 0 {
		t.Errorf("topic_mastered lines = %d, want none for a topic nothing was won in", len(lines))
	}
	if got := theOnlyLine(t, h, "answer_recorded")["topics_mastered"]; got != int64(1) {
		t.Errorf("answer_recorded topics_mastered = %v, want 1, as the progress shows", got)
	}
}

// raceAgain hands the race out once more, on the model's choice, on the
// profile a store keeps for the account, and is the id of the task on the
// card.
func raceAgain(t *testing.T, kept store.Storage, account store.Account, at time.Time) string {
	t.Helper()

	p, revision, err := kept.Load(context.Background(), account)
	if err != nil {
		t.Fatalf("Load() error = %v, want the profile", err)
	}
	askForTheRaceAt(t, p, rating.Grades12, 2, at)
	handOutTheRace(t, p, at)
	p.Touch("test", at)
	if _, err := kept.Save(context.Background(), account, p, revision); err != nil {
		t.Fatalf("Save() error = %v, want the race on the card", err)
	}
	return p.CurrentTask.ID
}

// askForTheRaceAt opens a request for the race on the profile, at a level and
// a difficulty of the model's choosing.
func askForTheRaceAt(t *testing.T, p *profile.Profile, level rating.GradeLevel, difficulty int, at time.Time) {
	t.Helper()

	loaded, err := shipped()
	if err != nil {
		t.Fatalf("load the content: %v", err)
	}
	brief, mode, err := tutor.Next(p, loaded, tutor.Choice{
		Topic: "logic.ordering", GradeLevel: level, Difficulty: difficulty, Reason: "The child asked for a race.",
	})
	if err != nil {
		t.Fatalf("Next() error = %v", err)
	}
	p.Ask(&brief, mode, "en", at)
}

// readerOf is a sign-in that lets the vouched token in as the account given.
func readerOf(account store.Account) mcpserver.TokenReader {
	return func(_ context.Context, token string) (store.Account, time.Time, error) {
		if token != vouchedToken {
			return store.Account{}, time.Time{}, errors.New(whyRefused)
		}
		return account, time.Now().Add(time.Hour), nil
	}
}

// masteredIn is how many topics a progress payload calls mastered.
func masteredIn(t *testing.T, result *mcp.CallToolResult) int {
	t.Helper()

	payload := payloadOf[progressPayload](t, result)
	count := 0
	for _, topic := range payload.Topics {
		if topic.Mastered {
			count++
		}
	}
	return count
}

// theOnlyLine is the fields of the one line of the event the harness kept.
func theOnlyLine(t *testing.T, h *harness, event string) map[string]any {
	t.Helper()

	lines := linesOf(h, event)
	if len(lines) != 1 {
		t.Fatalf("%s lines = %d, want 1", event, len(lines))
	}
	return lines[0].ContextMap()
}

// wantCounted holds a line a child is counted from to the fields decided for
// its event and to the forms they may hold.
func wantCounted(t *testing.T, event string, fields map[string]any) {
	t.Helper()

	for field, value := range fields {
		if !report.Decided(event, field) || !report.Fits(event, field, value) {
			t.Errorf("a %s line carries %v in %s, which its event may not carry", event, value, field)
		}
	}
}

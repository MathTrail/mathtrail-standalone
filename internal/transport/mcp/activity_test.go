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

// A task handed out says whether it came with a drawing, as a yes or a no and
// never as the drawing, which is the task's own text.
func TestATaskHandedOutSaysWhetherItCameWithADrawing(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name string
		hand func(*profile.OpenRequest) map[string]any
		want bool
	}{
		{"in words alone", raceOn, false},
		{"with a drawing", drawnRaceOn, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			kept := racer(t)
			h, session := lesson(t, kept)
			if handed := call(t, session, "submit_task", tc.hand(askForTheRace(t, session, kept))); handed.IsError {
				t.Fatalf("submit_task failed: %s", textOf(t, handed))
			}
			h.settle()

			fields := theOnlyLine(t, h, "task_accepted")
			if got := fields["drawing"]; got != tc.want {
				t.Errorf("task_accepted drawing = %v, want %v", got, tc.want)
			}
			wantCounted(t, "task_accepted", fields)
		})
	}
}

// drawnRaceOn is the race handed in with its three runners drawn, each named
// in the question by the label the drawing gives it.
func drawnRaceOn(request *profile.OpenRequest) map[string]any {
	race := raceOn(request)
	task, _ := race["task"].(map[string]any)
	task["question"] = "Ann (A), Ben (B) and Kim (K) ran a race. Ben finished before Kim. Ann finished after Kim. " +
		"Who finished first?"
	task["drawing"] = "A  B  K\n●  ●  ●"
	task["drawing_structure"] = map[string]any{
		"kind": "runners",
		"objects": []map[string]string{
			{"id": "ann", "label": "A"}, {"id": "ben", "label": "B"}, {"id": "kim", "label": "K"},
		},
	}
	return race
}

// The country a parent signed in from travels with the account the sign-in
// lets a call in as, and the task handed out is counted by it — unless the
// parent asked for it to be left out of what is counted, and then the line
// carries no country of the sign-in at all.
func TestATaskHandedOutIsCountedByTheCountryOfItsSignIn(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name    string
		leftOut bool
		want    any
	}{
		{"counted", false, "NZ"},
		{"left out by the parent", true, nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			account := store.NewAccount(vouchedUser, "", time.Time{})
			account.SignInCountry = "NZ"
			kept := memory.New()
			if _, err := kept.Create(context.Background(), account, profile.New(profile.Student{
				Grade: 2, Pseudonym: "Otter", Interests: []string{"sport"}, SignInCountryOff: tc.leftOut,
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

			if got := theOnlyLine(t, h, "task_accepted")["signin_country"]; got != tc.want {
				t.Errorf("task_accepted signin_country = %v, want %v", got, tc.want)
			}
		})
	}
}

// The demo account a directory's reviewers sign in to counts no child: its
// task handed out, its answer and its topic mastered are written as for any
// child, and none of them carries the name a child is counted under, which is
// what the counts are made from. Any other account is counted as before.
func TestTheDemoAccountsChildIsCountedNowhere(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name    string
		demo    []string
		counted bool
	}{
		{"the demo account", []string{vouchedUser}, false},
		{"the demo account, signed in before a rotation", []string{"its-identifier-now", vouchedUser}, false},
		{"a parent's beside it", []string{"the-demo-account"}, true},
	} {
		t.Run(tc.name+", a task handed out", func(t *testing.T) {
			t.Parallel()

			h, session, kept := demoLesson(t, tc.demo, profile.New(profile.Student{
				Grade: 2, Pseudonym: "Otter", Interests: []string{"sport"},
			}, "test", lessonDay))
			asked := call(t, session, "next_task", raceChoice)
			p, _, err := kept.Load(context.Background(), store.NewAccount(vouchedUser, "", time.Time{}))
			if err != nil || p.OpenRequest == nil {
				t.Fatalf("next_task = %q, and the profile holds no request: %v", textOf(t, asked), err)
			}
			wantOnTheCard(t, call(t, session, "submit_task", raceOn(p.OpenRequest)))
			h.settle()

			wantNamed(t, "task_accepted", theOnlyLine(t, h, "task_accepted"), tc.counted)
		})
		t.Run(tc.name+", an answer that masters its topic", func(t *testing.T) {
			t.Parallel()

			// A child whose right answer masters the topic, as in the case of
			// a topic counted as mastered.
			p := raceOnTheCard(t, 100)
			p.Ratings.Theta = 2.05
			p.Topics["logic.ordering"] = profile.Topic{Answers: 120, Correct: 120}
			h, session, _ := demoLesson(t, tc.demo, p)
			answerIt(t, session, p.CurrentTask.ID, "C", false)
			h.settle()

			for _, event := range []string{"answer_recorded", "topic_mastered"} {
				wantNamed(t, event, theOnlyLine(t, h, event), tc.counted)
			}
		})
	}
}

// demoLesson serves the tools of a lesson to the account the token signs in,
// whose profile is the one given, on a service told the identifiers of the
// demo account given.
func demoLesson(t *testing.T, demo []string, p *profile.Profile) (*harness, *mcp.ClientSession, store.Storage) {
	t.Helper()

	account := store.NewAccount(vouchedUser, "", time.Time{})
	kept := memory.New()
	if _, err := kept.Create(context.Background(), account, p); err != nil {
		t.Fatalf("keep the profile: %v", err)
	}
	h := newHarness(t)
	service := lessonService(t, h, kept, &clock{at: lessonDay}, nil, func(parts *mcpserver.Parts) { parts.DemoAccounts = demo })
	h.start(t, mcpserver.BearerSignIn(readerOf(account), metadata), slices.Concat(service.ProfileTools(), service.TaskTools())...)
	session, err := h.connectWith(t, vouchedToken)
	if err != nil {
		t.Fatalf("connectWith() error = %v, want nil", err)
	}
	return h, session, kept
}

// wantNamed holds a line that counts a child to naming the child, or to
// naming nobody when it is the demo account's, and to carrying everything else
// its event may either way.
func wantNamed(t *testing.T, event string, fields map[string]any, counted bool) {
	t.Helper()

	if _, named := fields["learner"]; named != counted {
		t.Errorf("%s names its child: %v, want %v", event, named, counted)
	}
	if fields["grade"] != int64(2) {
		t.Errorf("%s grade = %v, want the line written as for any child", event, fields["grade"])
	}
	wantCounted(t, event, fields)
}

// The language a task is written in is read from the profile, which a person
// can edit, and the file is held to no list of languages. A task asked for in
// whatever was typed there by hand — the child's pseudonym, say — is counted
// in the language other, and what was typed is never written out.
func TestATaskInALanguageTypedByHandIsCountedAsOther(t *testing.T) {
	t.Parallel()

	p := profile.New(profile.Student{Grade: 2, Pseudonym: "Otter", Interests: []string{"sport"}}, "test", lessonDay)
	askForTheRaceAt(t, p, rating.Grades12, 2, lessonDay)
	p.OpenRequest.Language = p.Student.Pseudonym
	kept := keptAsIs(t, p)
	h, session := lesson(t, kept)
	if card := payloadOf[handedInPayload](t, call(t, session, "submit_task", raceOn(p.OpenRequest))); card.Task == nil {
		t.Fatalf("submit_task = %+v, want the task on the card", card)
	}
	h.settle()

	fields := theOnlyLine(t, h, "task_accepted")
	if got := fields["language"]; got != "other" {
		t.Errorf("task_accepted language = %v, want other", got)
	}
	wantCounted(t, "task_accepted", fields)
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

	p := raceOnTheCard(t, 100)
	// A child settled in the topic, whose level in it clears the middle of
	// grades 1–2 less how far it may be off, and whom two wrong answers in a
	// row move by too little to keep it from clearing it again.
	p.Ratings.Theta = 2.05
	p.Topics["logic.ordering"] = profile.Topic{Answers: 120, Correct: 120}
	kept := keptAsIs(t, p)
	h, session := lesson(t, kept)

	var shown []int
	answer := func(letter string) {
		t.Helper()
		answerIt(t, session, loadedOf(t, kept).CurrentTask.ID, letter, false)
		shown = append(shown, masteredIn(t, call(t, session, "read_progress", map[string]any{})))
		raceAgain(t, kept, devAccount, lessonDay)
	}
	answer("C")                    // the level clears it: mastered
	answer("A")                    // a slip: still mastered
	answer("A")                    // two wrong in a row: lost
	answer("C")                    // the level clears it still: mastered again
	wantShown := []int{1, 1, 0, 1} // the topics the progress calls mastered, after each answer
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

// A topic mastered is counted only when the answer brings it among the topics
// the progress calls mastered: not when it is mastered below the level its
// tasks come from now, which the progress does not call mastered. A topic the
// progress calls mastered already is never mastered anew by one answer — the
// level it would take clears the middle of a level its tasks do not come from
// yet — so there is no second way for one mastery to be counted twice.
func TestOnlyATopicTheProgressComesToCallMasteredIsCounted(t *testing.T) {
	t.Parallel()

	p := profile.New(profile.Student{Grade: 2, Pseudonym: "Otter", Interests: []string{"sport"}}, "test", lessonDay)
	// A child whose level in the topic is past the tasks of grades 1–2, which
	// it masters the topic at, and whose tasks in it now come from grades 3–4.
	p.Ratings.Answers, p.Ratings.Theta = rating.TrialAnswers, 1.4
	p.Topics["logic.ordering"] = profile.Topic{Answers: 4, Correct: 4, Delta: 1}
	askForTheRaceAt(t, p, rating.Grades12, 4, lessonDay)
	handOutTheRace(t, p, lessonDay)
	kept := keptAsIs(t, p)
	h, session := lesson(t, kept)

	answerIt(t, session, p.CurrentTask.ID, "C", false)
	h.settle()

	switch won := loadedOf(t, kept).Topics["logic.ordering"].MasteredLevel; {
	case won == nil:
		t.Fatalf("the topic is not mastered, want the answer to have mastered it at %s", rating.Grades12)
	case *won != rating.Grades12:
		t.Fatalf("the topic is mastered at %s, want %s", *won, rating.Grades12)
	}
	if lines := linesOf(h, "topic_mastered"); len(lines) != 0 {
		t.Errorf("topic_mastered lines = %d, want none", len(lines))
	}
	if got := theOnlyLine(t, h, "answer_recorded")["topics_mastered"]; got != int64(0) {
		t.Errorf("answer_recorded topics_mastered = %v, want 0, as the progress shows", got)
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

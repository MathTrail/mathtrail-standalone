package profile_test

import (
	"bytes"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/profile"
	"github.com/MathTrail/mathtrail-standalone/internal/domain/rating"
)

// asked is when the tests of this file ask for a task.
var asked = time.Date(2026, 9, 5, 9, 0, 0, 0, time.UTC)

// briefOn is a brief the rule could have written, on this topic.
func briefOn(topic string) profile.Brief {
	return profile.Brief{
		Constraints:     []string{},
		Difficulty:      3,
		ExcludedSkills:  []string{},
		GradeLevel:      rating.Grades12,
		PedagogicalGoal: profile.GoalNewTopic,
		Rationale:       "A topic the child has not met yet.",
		Setting:         "space",
		TargetConcept:   topic,
		TrapsToUse:      []string{"off_by_one", "missed_case"},
	}
}

// written is a task as a model wrote it.
func written() *profile.Written {
	return &profile.Written{
		Wording:             "How many gaps are there between five posts in a row?",
		Options:             map[string]string{"A": "3", "B": "4", "C": "5", "D": "6", "E": "10"},
		Hint:                "Draw the posts and count what is between them.",
		Fingerprint:         "sketch-of-the-gaps",
		InstructionsVersion: "357968db0310",
	}
}

// failingSealer is a sealer whose key is gone: nothing it is handed can be
// sealed, which no real key ring can be made to do on demand.
type failingSealer struct{}

func (failingSealer) Seal([]byte, ...string) (string, error) {
	return "", errors.New("the key is gone")
}
func (failingSealer) Open(string, ...string) ([]byte, error) {
	return nil, errors.New("the key is gone")
}

// A request is what the model hands its task back against: the brief it was
// given, the language the task is to be in, who chose it and when, and no
// attempt spent. A second request replaces the first.
func TestARequestWaitsForTheTaskItAskedFor(t *testing.T) {
	t.Parallel()

	p := parseFixture(t, "dima")
	brief := briefOn("counting.gaps")
	request := p.Ask(&brief, profile.TutorLLM, "pt-BR", asked)

	if p.OpenRequest != request {
		t.Fatal("Ask() returned a request the profile does not hold")
	}
	if !strings.HasPrefix(request.ID, "req_") {
		t.Errorf("the request is %q, want an id of a request", request.ID)
	}
	want := profile.OpenRequest{
		Attempts: 0, Brief: brief, ID: request.ID, Language: "pt-BR", OpenedAt: profile.At(asked), TutorMode: profile.TutorLLM,
	}
	if !reflect.DeepEqual(*request, want) {
		t.Errorf("the request is %+v, want %+v", *request, want)
	}
	if err := p.Validate(); err != nil {
		t.Errorf("Validate() error = %v, want a profile with a request to be one", err)
	}

	again := p.Ask(&brief, profile.TutorRule, "en", asked.Add(time.Minute))
	if again.ID == request.ID || p.OpenRequest != again {
		t.Errorf("a second request is %q and the profile holds %q, want the second to replace %q",
			again.ID, p.OpenRequest.ID, request.ID)
	}
}

// A request is waited for through its window and not a moment longer, and
// only while it has an attempt left. A request dated further ahead than the
// window, or one with every attempt spent, is one only a hand editing the file
// could leave open, and neither is waited for: either would hold every new
// request back.
func TestARequestIsAwaitedOnlyInsideItsWindow(t *testing.T) {
	t.Parallel()

	const window = 15 * time.Minute
	for _, tc := range []struct {
		name     string
		attempts int
		now      time.Time
		awaited  bool
	}{
		{"just opened", 0, asked, true},
		{"a second before the window closes", 2, asked.Add(window - time.Second), true},
		{"as the window closes", 0, asked.Add(window), false},
		{"long after", 0, asked.Add(24 * time.Hour), false},
		{"on a clock a second behind the one that opened it", 0, asked.Add(-time.Second), true},
		{"on a date further ahead than the window", 0, asked.Add(-window - time.Second), false},
		{"with every attempt spent", profile.MaxAttempts, asked, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			request := profile.OpenRequest{Attempts: tc.attempts, OpenedAt: profile.At(asked)}
			if got := request.Awaited(window, tc.now); got != tc.awaited {
				t.Errorf("Awaited() = %v, want %v", got, tc.awaited)
			}
		})
	}
}

// Every refusal spends an attempt, and the last one closes the request: the
// child is handed nothing, the day's failed generations go up, and the day's
// accepted tasks do not — a refusal is not a task.
func TestTheLastRefusalClosesTheRequest(t *testing.T) {
	t.Parallel()

	p := parseFixture(t, "dima")
	p.Daily = profile.Daily{Date: profile.DateOf(asked), Accepted: 2}
	brief := briefOn("counting.gaps")
	p.Ask(&brief, profile.TutorRule, "en", asked)

	for attempt := 1; attempt <= profile.MaxAttempts; attempt++ {
		exhausted, err := p.Refuse(asked)
		if err != nil {
			t.Fatalf("Refuse() error = %v on attempt %d, want nil", err, attempt)
		}
		if last := attempt == profile.MaxAttempts; exhausted != last {
			t.Fatalf("attempt %d: Refuse() = %v, want %v", attempt, exhausted, last)
		}
		if attempt < profile.MaxAttempts && p.OpenRequest.Attempts != attempt {
			t.Errorf("after attempt %d the request counts %d", attempt, p.OpenRequest.Attempts)
		}
	}

	if p.OpenRequest != nil {
		t.Errorf("the request is still open after %d refusals", profile.MaxAttempts)
	}
	if want := (profile.Daily{Date: profile.DateOf(asked), Accepted: 2, Failed: 1}); p.Daily != want {
		t.Errorf("the counters are %+v, want %+v", p.Daily, want)
	}
	if _, err := p.Refuse(asked); !errors.Is(err, profile.ErrNoRequest) {
		t.Errorf("Refuse() with no request error = %v, want %v", err, profile.ErrNoRequest)
	}
}

// An accepted task is handed out standing exactly where the request asked:
// its topic, level and difficulty are the brief's and its language the
// request's, whatever the model might have said. Its answer is sealed, the
// request is closed, its fingerprint is remembered, its topic records the day,
// and it counts as the day's task.
func TestAnIssuedTaskStandsWhereTheRequestAsked(t *testing.T) {
	t.Parallel()

	p := parseFixture(t, "dima")
	p.Daily = profile.Daily{Date: profile.DateOf(asked), Accepted: 2}
	fingerprints := len(p.TaskFingerprints)
	brief := briefOn("counting.gaps")
	p.Ask(&brief, profile.TutorRule, "de-CH", asked)
	sealer := newSealer(t)
	handed := asked.Add(2 * time.Minute)

	task, err := p.Issue(written(), secret(), sealer, handed)
	if err != nil {
		t.Fatalf("Issue() error = %v, want nil", err)
	}

	if task != p.CurrentTask || !strings.HasPrefix(task.ID, "tsk_") {
		t.Fatalf("Issue() = %+v, want the task in flight with an id of a task", task)
	}
	want := profile.CurrentTask{
		Difficulty: brief.Difficulty, Fingerprint: "sketch-of-the-gaps", GradeLevel: brief.GradeLevel,
		Hint: written().Hint, ID: task.ID, InstructionsVersion: "357968db0310", IssuedAt: profile.At(handed),
		Language: "de-CH", Options: written().Options, Sealed: task.Sealed, Topic: "counting.gaps", Wording: written().Wording,
	}
	if !reflect.DeepEqual(*task, want) {
		t.Errorf("the task is %+v, want %+v", *task, want)
	}
	opened, err := p.OpenTask(sealer)
	if err != nil || opened.Answer != secret().Answer || opened.Solution != secret().Solution {
		t.Errorf("OpenTask() = %+v, %v, want the secret that was sealed", opened, err)
	}

	if p.OpenRequest != nil {
		t.Error("the request is still open once its task was handed out")
	}
	if len(p.TaskFingerprints) != fingerprints+1 || p.TaskFingerprints[fingerprints] != "sketch-of-the-gaps" {
		t.Errorf("the fingerprints end with %v, want the task's added last", p.TaskFingerprints[fingerprints-1:])
	}
	if got := p.Topics["counting.gaps"].LastIssued; !got.Equal(profile.DateOf(handed).Time) {
		t.Errorf("the topic was last issued on %v, want the day it was handed out", got)
	}
	if want := (profile.Daily{Date: profile.DateOf(asked), Accepted: 3}); p.Daily != want {
		t.Errorf("the counters are %+v, want %+v", p.Daily, want)
	}
	if err := p.Validate(); err != nil {
		t.Errorf("Validate() error = %v, want the profile with its task to be one", err)
	}

	if _, err := p.Issue(written(), secret(), sealer, handed); !errors.Is(err, profile.ErrNoRequest) {
		t.Errorf("Issue() with no request error = %v, want %v", err, profile.ErrNoRequest)
	}
}

// A task that cannot be sealed is not handed out, and nothing about the
// profile moves: not the task in flight, not the request, not a counter.
func TestATaskThatCannotBeSealedChangesNothing(t *testing.T) {
	t.Parallel()

	p := parseFixture(t, "masha")
	brief := briefOn("counting.gaps")
	p.Ask(&brief, profile.TutorRule, "en", asked)
	before, err := profile.Marshal(p)
	if err != nil {
		t.Fatalf("Marshal() error = %v, want nil", err)
	}

	if _, refused := p.Issue(written(), secret(), failingSealer{}, asked); refused == nil {
		t.Fatal("Issue() error = nil, want the sealer's refusal")
	}

	after, err := profile.Marshal(p)
	if err != nil {
		t.Fatalf("Marshal() error = %v, want nil", err)
	}
	if !bytes.Equal(before, after) {
		t.Errorf("a task that could not be sealed changed the file:\nbefore %s\nafter  %s", before, after)
	}
}

// The fingerprints are a window of their own: the newest joins at the end and
// the oldest leaves once there are as many as are kept.
func TestTheFingerprintsKeepTheLatestTasks(t *testing.T) {
	t.Parallel()

	p := parseFixture(t, "dima")
	p.TaskFingerprints = make([]string, profile.MaxFingerprints)
	for i := range p.TaskFingerprints {
		p.TaskFingerprints[i] = fmt.Sprintf("old-%03d", i)
	}
	brief := briefOn("counting.gaps")
	p.Ask(&brief, profile.TutorRule, "en", asked)

	if _, err := p.Issue(written(), secret(), newSealer(t), asked); err != nil {
		t.Fatalf("Issue() error = %v, want nil", err)
	}

	kept := p.TaskFingerprints
	if len(kept) != profile.MaxFingerprints || kept[0] != "old-001" || kept[len(kept)-1] != "sketch-of-the-gaps" {
		t.Errorf("the fingerprints are %d, from %q to %q; want %d, from old-001 to the new one",
			len(kept), kept[0], kept[len(kept)-1], profile.MaxFingerprints)
	}
}

// A task left without an answer goes into the history as skipped, at the
// moment it was left, and into the count of its topic; it leaves flight. And
// that is all: there was no answer, so the level, the runs, the series and the
// day's counters stay where they were, and the last answer is still the last
// answer.
func TestASkippedTaskIsRecordedAndMovesNothingElse(t *testing.T) {
	t.Parallel()

	p, before := parseFixture(t, "masha"), parseFixture(t, "masha")
	task := before.CurrentTask

	entry, skipped := p.Skip(asked)
	if !skipped {
		t.Fatal("Skip() found no task in flight, want Masha's")
	}
	want := profile.Answer{
		AnsweredAt: profile.At(asked), Difficulty: task.Difficulty, GradeLevel: task.GradeLevel,
		Skipped: true, TaskID: task.ID, Topic: task.Topic,
	}
	if entry != want || p.Recent[len(p.Recent)-1] != want || len(p.Recent) != len(before.Recent)+1 {
		t.Errorf("Skip() = %+v and the window ends with %+v, want %+v added to the end",
			entry, p.Recent[len(p.Recent)-1], want)
	}
	if p.CurrentTask != nil {
		t.Error("the skipped task is still in flight")
	}

	topic := p.Topics[task.Topic]
	if topic.Skipped != 1 {
		t.Errorf("the topic counts %d skipped tasks, want 1", topic.Skipped)
	}
	topic.Skipped = 0
	if !reflect.DeepEqual(topic, before.Topics[task.Topic]) {
		t.Errorf("the topic moved to %+v, want only its count of skipped tasks to move from %+v",
			topic, before.Topics[task.Topic])
	}
	if p.Ratings != before.Ratings || p.Daily != before.Daily ||
		!reflect.DeepEqual(p.TaskFingerprints, before.TaskFingerprints) {
		t.Errorf("a skip moved the ratings to %+v, the counters to %+v or the fingerprints", p.Ratings, p.Daily)
	}
	if last, _ := p.LastAnswer(); last.TaskID != before.Recent[len(before.Recent)-1].TaskID {
		t.Errorf("LastAnswer() = %s, want the answer before the skip", last.TaskID)
	}
	if err := p.Validate(); err != nil {
		t.Errorf("Validate() error = %v, want a profile with a skip to be one", err)
	}

	if _, again := p.Skip(asked); again {
		t.Error("Skip() found a task in flight after the only one was skipped")
	}
}

// A skipped task is no answer of the trial series: the level the next answer
// gives is the one it would give had the task been dropped without a word.
func TestASkippedTaskIsNoAnswerOfTheTrialSeries(t *testing.T) {
	t.Parallel()

	skipping, dropping := parseFixture(t, "masha"), parseFixture(t, "masha")
	if !skipping.Ratings.InTrial() {
		t.Fatal("Masha has finished the trial series, and this needs a child in it")
	}
	skipping.Skip(asked)
	dropping.CurrentTask = nil

	next := rating.Point{GradeLevel: rating.Grades34, Difficulty: 2}
	for _, p := range []*profile.Profile{skipping, dropping} {
		id := answeringAt(t, p, "time.clocks", next)
		if _, err := p.Record(profile.Answered{TaskID: id, Correct: true, At: asked.Add(time.Minute)}); err != nil {
			t.Fatalf("Record() error = %v, want nil", err)
		}
	}

	if skipping.Ratings != dropping.Ratings {
		t.Errorf("after a skip the answer gives %+v, want %+v as without it", skipping.Ratings, dropping.Ratings)
	}
}

// Everything the lesson writes on the way from a request to a skipped task is
// a file the schema describes and this package reads back.
func TestEveryStepOfTheLessonWritesAFileTheSchemaDescribes(t *testing.T) {
	t.Parallel()

	p := parseFixture(t, "dima")
	brief := briefOn("counting.gaps")
	steps := []struct {
		name string
		take func() error
	}{
		{"a request", func() error { p.Ask(&brief, profile.TutorRule, "en", asked); return nil }},
		{"a refusal", func() error { _, err := p.Refuse(asked); return err }},
		{"a task handed out", func() error { _, err := p.Issue(written(), secret(), newSealer(t), asked); return err }},
		{"a task skipped", func() error { p.Skip(asked.Add(time.Minute)); return nil }},
	}
	for _, step := range steps {
		if err := step.take(); err != nil {
			t.Fatalf("%s: error = %v, want nil", step.name, err)
		}
		file, err := profile.Marshal(p)
		if err != nil {
			t.Fatalf("%s: Marshal() error = %v, want nil", step.name, err)
		}
		if err := againstTheSchema(t, file); err != nil {
			t.Errorf("%s: the file breaks the schema: %v", step.name, err)
		}
		if _, err := profile.Parse(file); err != nil {
			t.Errorf("%s: Parse() error = %v, want the file read back", step.name, err)
		}
	}
}

// A file that says null where the topics go — one a person edited by hand —
// still takes a task and its skip: the topic is written into a summary of its
// own rather than into nothing.
func TestAFileWithNoTopicsStillTakesATask(t *testing.T) {
	t.Parallel()

	p := parseFixture(t, "sasha")
	p.Topics = nil
	brief := briefOn("counting.gaps")
	p.Ask(&brief, profile.TutorRule, "en", asked)

	if _, err := p.Issue(written(), secret(), newSealer(t), asked); err != nil {
		t.Fatalf("Issue() error = %v, want nil", err)
	}
	if p.Topics["counting.gaps"].LastIssued.IsZero() {
		t.Errorf("topics = %+v, want the topic handed out recorded", p.Topics)
	}

	p.Topics = nil
	if _, skipped := p.Skip(asked); !skipped || p.Topics["counting.gaps"].Skipped != 1 {
		t.Errorf("topics = %+v, want the skipped task counted", p.Topics)
	}
}

// A task still in flight when another is handed out — which only a file edited
// by hand can hold beside an open request — is not lost: it goes into the
// window as skipped, like any task left without an answer.
func TestATaskStillInFlightIsSkippedByTheOneHandedOut(t *testing.T) {
	t.Parallel()

	p := parseFixture(t, "masha")
	left := p.CurrentTask
	brief := briefOn("counting.gaps")
	p.Ask(&brief, profile.TutorRule, "en", asked)

	issued, err := p.Issue(written(), secret(), newSealer(t), asked)
	if err != nil {
		t.Fatalf("Issue() error = %v, want nil", err)
	}
	last := p.Recent[len(p.Recent)-1]
	if p.CurrentTask != issued || !last.Skipped || last.TaskID != left.ID || p.Topics[left.Topic].Skipped != 1 {
		t.Errorf("the window ends with %+v and the task in flight is %s, want %s skipped and the new task in flight",
			last, p.CurrentTask.ID, left.ID)
	}
}

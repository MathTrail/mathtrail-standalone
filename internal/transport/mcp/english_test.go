package mcpserver_test

import (
	"slices"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/checks"
	"github.com/MathTrail/mathtrail-standalone/internal/domain/profile"
	"github.com/MathTrail/mathtrail-standalone/internal/domain/rating"
	"github.com/MathTrail/mathtrail-standalone/internal/report"
	"github.com/MathTrail/mathtrail-standalone/internal/store"
)

// A lesson in another language than English asks a task for its question in
// English as well, which the reference tasks, written in English, are compared
// with and nothing else reads. The lessons here are the race's, in Russian, the
// language the parent chose, or in English.

// lessonInRussian is a lesson in Russian, the language the parent chose, with
// the race asked for in it: the harness that keeps its log and its traces, the
// session, the profile's store and the request opened.
func lessonInRussian(t *testing.T) (h *harness, session *mcp.ClientSession, kept store.Storage,
	request *profile.OpenRequest,
) {
	t.Helper()

	russian := "ru"
	kept = keptAsIs(t, profile.New(profile.Student{
		Grade: 2, Pseudonym: "Otter", Interests: []string{"sport"}, UILanguage: &russian,
	}, "test", lessonDay))
	h, session = lesson(t, kept)
	return h, session, kept, askForTheRace(t, session, kept)
}

// Without its question in English a task of a lesson in Russian is refused for
// its form, the model told what the member is and where the guide says so, and
// the line of the hand-in says it came without one; with it the task is taken,
// and the line says so. The fingerprint the profile keeps is the question's in
// Russian, and the question in English reaches nothing: no payload, no words
// of a tool, no part of the profile, no span and no line holds it.
func TestALessonInRussianAsksTheTaskForItsQuestionInEnglish(t *testing.T) {
	t.Parallel()

	h, session, kept, request := lessonInRussian(t)
	without := raceInRussianOn(request)
	task, _ := without["task"].(map[string]any)
	delete(task, "question_en")
	refused := call(t, session, "submit_task", without)
	if card := payloadOf[handedInPayload](t, refused); card.Code != string(checks.CodeBadStructure) ||
		len(card.Reasons) != 1 || !strings.Contains(strings.Join(card.Reasons[0].Messages, " "),
		"task.question_en is missing or empty") {
		t.Errorf("submit_task without its question in English = %+v, want it refused for that alone", card)
	}
	wantSaid(t, textOf(t, refused), "with the same names, numbers and conditions",
		"get_package with the same request_id hands it back")
	handed := call(t, session, "submit_task", raceInRussianOn(request))
	if card := payloadOf[handedInPayload](t, handed); card.Screen != "task" {
		t.Fatalf("submit_task with its question in English = %+v, want the task on the card", card)
	}

	h.settle()
	var given []any
	lines := linesOf(h, "task_submitted")
	for i := range lines {
		fields := lines[i].ContextMap()
		given = append(given, fields["question_en_given"])
		for field := range fields {
			if !report.Decided("task_submitted", field) {
				t.Errorf("a task_submitted line carries %s, which its event is not decided to carry", field)
			}
		}
	}
	if !slices.Equal(given, []any{false, true}) {
		t.Errorf("question_en_given = %v, want the hand-in without it and then the one with it", given)
	}

	p, _ := loadKept(t, kept)
	question, _ := task["question"].(string)
	if fingerprints := p.TaskFingerprints; len(fingerprints) == 0 ||
		fingerprints[len(fingerprints)-1] != checks.Fingerprint(question, "ru") {
		t.Errorf("the profile keeps the fingerprints %v, want the last the question's in Russian", fingerprints)
	}
	wantTheQuestionInEnglishNowhere(t, h, p, refused, handed)
}

// wantTheQuestionInEnglishNowhere holds what a lesson in Russian left — the
// results of its calls, the profile, its spans and its lines — to holding none
// of the race's question in English.
func wantTheQuestionInEnglishNowhere(t *testing.T, h *harness, p *profile.Profile, results ...*mcp.CallToolResult) {
	t.Helper()

	secrets := []string{raceQuestion}
	for _, result := range results {
		wantNoneOf(t, "the payload of a hand-in", []string{string(rawPayload(t, result))}, secrets)
		wantNoneOf(t, "the words of a hand-in", []string{textOf(t, result)}, secrets)
	}
	p.CurrentTask.Sealed = "sealed"
	written, err := profile.Marshal(p)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}
	wantNoneOf(t, "the profile", []string{string(written)}, secrets)
	for _, span := range h.spans.Ended() {
		wantNoneOf(t, "span "+span.Name(), spanTexts(span), secrets)
	}
	lines := h.logs.All()
	for i := range lines {
		wantNoneOf(t, "line "+lines[i].Message, withoutPictureOfTheSolution(t, &lines[i]), secrets)
	}
}

// A task of a lesson in Russian whose question in English copies a reference
// task its package shows is refused as a copy of it, though its question in
// Russian shares nothing with any, and the line of the hand-in names the copy
// by the question in English. The model is told to write a task of its own and
// to keep the question in English the task's question.
func TestACopyToldInRussianIsRefusedByItsQuestionInEnglish(t *testing.T) {
	t.Parallel()

	h, session, _, request := lessonInRussian(t)
	race := raceInRussianOn(request)
	task, _ := race["task"].(map[string]any)
	task["question_en"] = exampleQuestionsIn(t, fetchPackage(t, session, request.ID))[0]
	told := textOf(t, call(t, session, "submit_task", race))

	line := submittedLine(t, h)
	if got := namesOf(line["near_duplicate_of"]); !slices.Equal(got, []string{"reference_en"}) {
		t.Errorf("near_duplicate_of = %v, want the copy named by the question in English", got)
	}
	if line["question_en_given"] != true || line["level"] != string(rating.Grades12) {
		t.Errorf("question_en_given = %v and level = %v, want true at the level the race was asked at, %s",
			line["question_en_given"], line["level"], rating.Grades12)
	}
	wantSaid(t, told, "task.question_en is a near-copy of one of the reference tasks, so the task tells it again in "+
		"the lesson's language; write a task of your own rather than a translation of an example, and keep "+
		"task.question_en its question in English, with the same names, numbers and conditions")
}

// A lesson in English asks for no question in English, and the text of one
// sent is not read: one that copies a reference task its package shows costs
// the task nothing, and the line of the hand-in says nothing of it.
func TestALessonInEnglishReadsNoQuestionInEnglish(t *testing.T) {
	t.Parallel()

	kept := racer(t)
	h, session := lesson(t, kept)
	request := askForTheRace(t, session, kept)
	race := raceOn(request)
	task, _ := race["task"].(map[string]any)
	task["question_en"] = exampleQuestionsIn(t, fetchPackage(t, session, request.ID))[0]
	if text := textOf(t, call(t, session, "submit_task", race)); !strings.HasPrefix(text, "Accepted at attempt 1") {
		t.Fatalf("submit_task said %q, want the race accepted", text)
	}
	if given, said := submittedLine(t, h)["question_en_given"]; said {
		t.Errorf("question_en_given = %v, want nothing said of it in a lesson in English", given)
	}
}

// The words that come with a package of a lesson in Russian ask for the
// question in English as well, before the package and the guide in it, and
// those of a lesson in English do not: the words of get_package for a task
// asked for now, and those of prepare_task for the task written ahead.
func TestThePackagesWordsAskForTheQuestionInEnglishWhereTheLessonIsNotInEnglish(t *testing.T) {
	t.Parallel()

	const asked = "Give its question in English as well, in task.question_en, with the same names, numbers and " +
		"conditions and the names in Latin letters, as the guide in the package says"
	_, russian, _, inRussian := lessonInRussian(t)
	english := racer(t)
	_, inEnglish := lesson(t, english)
	for name, words := range map[string]map[string]string{
		"in Russian": packagesWords(t, russian, inRussian, raceInRussianOn(inRussian)),
		"in English": packagesWords(t, inEnglish, askForTheRace(t, inEnglish, english), nil),
	} {
		for tool, said := range words {
			if strings.Contains(said, asked) != (name == "in Russian") {
				t.Errorf("%s, %s says %q, want the question in English asked for in Russian alone", name, tool, said)
			}
		}
	}
}

// packagesWords are the words before the package of a request, as get_package
// gives them, and before the package of the task after it, as prepare_task
// gives them once the task is on the card: handed in as race, or as the race
// in English when none is given.
func packagesWords(t *testing.T, session *mcp.ClientSession, request *profile.OpenRequest,
	race map[string]any,
) map[string]string {
	t.Helper()

	if race == nil {
		race = raceOn(request)
	}
	asked := leadOf(fetchPackage(t, session, request.ID))
	if text := textOf(t, call(t, session, "submit_task", race)); !strings.HasPrefix(text, "Accepted at attempt 1") {
		t.Fatalf("submit_task said %q, want the race accepted", text)
	}
	return map[string]string{"get_package": asked, "prepare_task": leadOf(prepared(t, session, aheadChoice))}
}

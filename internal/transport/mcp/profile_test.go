package mcpserver_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strings"
	"testing"
	"time"
	"unicode/utf16"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	tracenoop "go.opentelemetry.io/otel/trace/noop"
	"go.uber.org/zap"

	"github.com/MathTrail/mathtrail-standalone/internal/config"
	"github.com/MathTrail/mathtrail-standalone/internal/domain/checks"
	"github.com/MathTrail/mathtrail-standalone/internal/domain/profile"
	"github.com/MathTrail/mathtrail-standalone/internal/domain/rating"
	"github.com/MathTrail/mathtrail-standalone/internal/store"
	"github.com/MathTrail/mathtrail-standalone/internal/store/memory"
	mcpserver "github.com/MathTrail/mathtrail-standalone/internal/transport/mcp"
)

// profilePayload is what the tools of the profile hand a card, as a card reads
// it.
type profilePayload struct {
	Screen     string           `json:"screen"`
	Status     string           `json:"status"`
	Code       string           `json:"code"`
	Problems   []problemPayload `json:"problems"`
	LastAnswer *struct {
		TaskID  string `json:"task_id"`
		Correct bool   `json:"correct"`
	} `json:"last_answer"`
	Profile *struct {
		Pseudonym      string   `json:"pseudonym"`
		Grade          int      `json:"grade"`
		Interests      []string `json:"interests"`
		ExcludedSkills []string `json:"excluded_skills"`
		Notes          string   `json:"notes"`
		UILanguage     *string  `json:"ui_language"`
	} `json:"profile"`
	Trial          *trialPayload `json:"trial"`
	Recommendation *struct {
		Topic      string `json:"topic"`
		GradeLevel string `json:"grade_level"`
		Difficulty int    `json:"difficulty"`
		Goal       string `json:"goal"`
	} `json:"recommendation"`
}

// progressPayload is what the tools of the progress hand a card.
type progressPayload struct {
	Screen     string `json:"screen"`
	LastAnswer *struct {
		TaskID string `json:"task_id"`
	} `json:"last_answer"`
	Profile *struct {
		Pseudonym string `json:"pseudonym"`
		Grade     int    `json:"grade"`
	} `json:"profile"`
	Trial   *trialPayload `json:"trial"`
	Overall *struct {
		Rating int             `json:"rating"`
		Rank   int             `json:"rank"`
		Ranks  int             `json:"ranks"`
		Share  int             `json:"share"`
		Grades []gradesPayload `json:"grades"`
		Change *changesPayload `json:"change"`
	} `json:"overall"`
	Topics []topicPayload `json:"topics"`
	Recent []struct {
		Topic string `json:"topic"`
	} `json:"recent"`
	Skipped        int              `json:"skipped"`
	Mistakes       []mistakePayload `json:"mistakes"`
	Recommendation *struct {
		Topic string `json:"topic"`
	} `json:"recommendation"`
	Location *struct {
		File string `json:"file"`
	} `json:"location"`
}

// gradesPayload is a grade level and the ranks it is matched with, as the
// overall standing hands them a card.
type gradesPayload struct {
	GradeLevel string `json:"grade_level"`
	FirstRank  int    `json:"first_rank"`
	LastRank   int    `json:"last_rank"`
}

// topicPayload is one topic of the progress.
type topicPayload struct {
	Topic    string          `json:"topic"`
	Rating   *int            `json:"rating"`
	Rank     *int            `json:"rank"`
	Share    *int            `json:"share"`
	Compared *string         `json:"compared"`
	Answers  int             `json:"answers"`
	Mastered bool            `json:"mastered"`
	Skipped  int             `json:"skipped"`
	Change   *changesPayload `json:"change"`
}

// standsTogether says whether a topic's rating, rank, share and comparison are
// all there or all null, as a card reads them: never one without the others.
func (topic topicPayload) standsTogether() bool {
	there := topic.Rating != nil
	return (topic.Rank != nil) == there && (topic.Share != nil) == there && (topic.Compared != nil) == there
}

// mistakePayload is a mistake on the map of misconceptions.
type mistakePayload struct {
	Trap  string `json:"trap"`
	Times int    `json:"times"`
}

type trialPayload struct {
	Answered int `json:"answered"`
	Of       int `json:"of"`
}

// A host lists the ten tools of the lesson as they are meant: two that draw a
// card, under both keys a host reads — the progress, and the task asked for,
// whose card waits for it and turns into it; the two of the profile, which the
// model reads and writes in words, drawing nothing: the profile has a section
// of the progress; three that only a card calls, kept from the model
// and drawing nothing — the progress opened inside a card, the change the form
// on it sends, and how the task a card waits for stands; the one that hands the
// model what to write a task from, which draws nothing and declares no
// payload, since what it hands over is for the model alone and travels in its
// words; the one that takes the task written, which draws nothing either, the
// card that waits for it being already drawn; and the one that records an
// answer, which the card calls as well as the model and which draws nothing,
// since the card that sent the answer turns to its result. Handing a task in
// is the one call that is not the same twice: each spends an attempt. The five
// that read change nothing; the two that write the profile replace what the
// adult set, and are the only ones a host may take for destroying something;
// the three of the lesson only add to its record. None reaches anything beyond
// the parent's file. Every description is short enough to reach the model
// whole.
func TestTheToolsOfTheLessonAreListedAsTheyAreMeant(t *testing.T) {
	t.Parallel()

	_, session := lesson(t, memory.New())
	listed, err := session.ListTools(t.Context(), nil)
	if err != nil {
		t.Fatalf("ListTools() error = %v", err)
	}
	byName := map[string]*mcp.Tool{}
	for _, tool := range listed.Tools {
		byName[tool.Name] = tool
	}
	if len(byName) != 10 {
		t.Errorf("%d tools are listed, want the ten of the lesson", len(byName))
	}

	for _, want := range []listing{
		{name: "get_profile", readOnly: true, idempotent: true},
		{name: "save_profile", destructive: true, idempotent: true},
		{name: "edit_profile", destructive: true, idempotent: true, widgetOnly: true},
		{name: "get_progress", readOnly: true, idempotent: true, drawsCard: true},
		{name: "read_progress", readOnly: true, idempotent: true, widgetOnly: true},
		{name: "next_task", idempotent: true, drawsCard: true},
		{name: "get_package", readOnly: true, idempotent: true, wordsOnly: true},
		{name: "submit_task"},
		{name: "read_task", readOnly: true, idempotent: true, widgetOnly: true},
		{name: "submit_answer", idempotent: true},
	} {
		t.Run(want.name, func(t *testing.T) {
			t.Parallel()

			tool := byName[want.name]
			if tool == nil {
				t.Fatalf("%s is not listed", want.name)
			}
			wantListedAs(t, tool, want)
		})
	}
}

// The tools of the profile and the progress are described as not being where
// a lesson starts: the progress puts the adult's screen in front of the child,
// and the profile is the adult's business. A task needs only next_task, which
// says when there is no profile, and the progress is shown when someone asks
// for it. Neither tool of the profile draws a card, and both say so: the model
// is the one to tell the adult what was saved, and where the form is.
func TestTheAdultsToolsAreNotWhereALessonStarts(t *testing.T) {
	t.Parallel()

	_, session := lesson(t, memory.New())
	listed, err := session.ListTools(t.Context(), nil)
	if err != nil {
		t.Fatalf("ListTools() error = %v", err)
	}
	described := map[string]string{}
	for _, tool := range listed.Tools {
		described[tool.Name] = tool.Description
	}
	for _, want := range []struct{ name, says string }{
		{"get_profile", "a task needs only next_task"},
		{"get_profile", "It draws no card: the adult sees the profile, and changes it with a form, in the Profile section of the progress"},
		{"save_profile", "No card is drawn: say in a sentence what was saved."},
		{"get_progress", "Call it only when someone asks to see the progress"},
		{"get_progress", "card draws the rank, not the rating's number: say the number yourself"},
		{"get_progress", "that is for the adult, a rough guide and never a school mark or a verdict on the child"},
	} {
		if description, listed := described[want.name]; !listed || !strings.Contains(description, want.says) {
			t.Errorf("%s is described as %q (listed: %v), want it listed and saying %q", want.name, description, listed, want.says)
		}
	}
}

// listing is how a tool is meant to be listed. wordsOnly marks the tool that
// answers the model in words alone, and so declares no payload to check.
type listing struct {
	name        string
	readOnly    bool
	destructive bool
	idempotent  bool
	drawsCard   bool
	widgetOnly  bool
	wordsOnly   bool
}

// hostKeeps is as much of a tool's description as a host is known to pass on to
// the model: Claude Code keeps the first 2,048 characters, counted as
// JavaScript counts them, and cuts the rest.
const hostKeeps = 2048

// wantListedAs holds a listed tool to its listing: its title, its hints, its
// output schema, whether a host draws a card for it and whether the model sees
// it — and its description, which reaches the model whole.
func wantListedAs(t *testing.T, tool *mcp.Tool, want listing) {
	t.Helper()

	if length := len(utf16.Encode([]rune(tool.Description))); length > hostKeeps {
		t.Errorf("the description is %d characters long, and a host may keep only the first %d of it", length, hostKeeps)
	}
	hints := tool.Annotations
	if hints == nil {
		hints = &mcp.ToolAnnotations{}
	}
	if hints.ReadOnlyHint != want.readOnly || hints.IdempotentHint != want.idempotent ||
		hints.DestructiveHint == nil || *hints.DestructiveHint != want.destructive ||
		hints.OpenWorldHint == nil || *hints.OpenWorldHint {
		said, _ := json.Marshal(tool.Annotations)
		t.Errorf("annotations = %s, want read-only %v, destructive %v, idempotent %v and closed to the world",
			said, want.readOnly, want.destructive, want.idempotent)
	}
	if tool.Title == "" || hints.Title != tool.Title {
		t.Errorf("title = %q and the annotations' title = %q, want one a person can read, the same in both",
			tool.Title, hints.Title)
	}

	if hasSchema := tool.OutputSchema != nil; hasSchema == want.wordsOnly {
		t.Errorf("an output schema: %v, want one a host can check a result against unless the tool answers in words alone: %v",
			hasSchema, !want.wordsOnly)
	}
	ui, _ := tool.Meta["ui"].(map[string]any)
	drawn := ui["resourceUri"] == mcpserver.WidgetURI && tool.Meta["ui/resourceUri"] == mcpserver.WidgetURI
	if drawn != want.drawsCard {
		t.Errorf("_meta = %v, want the widget named under both keys: %v", tool.Meta, want.drawsCard)
	}
	visibility, _ := json.Marshal(ui["visibility"])
	if hidden := string(visibility) == `["app"]`; hidden != want.widgetOnly {
		t.Errorf("_meta.ui.visibility = %s, kept from the model: %v, want %v", visibility, hidden, want.widgetOnly)
	}
}

// The model can only fill in the skills a child has not met with ids it has
// been told: the description of save_profile names every skill of the catalog.
func TestSaveProfileNamesEverySkillItTakes(t *testing.T) {
	t.Parallel()

	_, session := lesson(t, memory.New())
	listed, err := session.ListTools(t.Context(), nil)
	if err != nil {
		t.Fatalf("ListTools() error = %v", err)
	}
	loaded, _ := shipped()
	for _, tool := range listed.Tools {
		if tool.Name != "save_profile" {
			continue
		}
		for _, skill := range loaded.Skills() {
			if !strings.Contains(tool.Description, "- "+skill.ID+": ") {
				t.Errorf("the description of save_profile does not name the skill %s", skill.ID)
			}
		}
		return
	}
	t.Fatal("save_profile is not listed")
}

// The limits save_profile tells the model are the limits the profile holds. The
// words of an argument cannot be made of the constants, so this is what keeps
// a cap changed in one place from being announced as it was in the other.
func TestSaveProfileStatesTheCapsTheProfileHolds(t *testing.T) {
	t.Parallel()

	_, session := lesson(t, memory.New())
	listed, err := session.ListTools(t.Context(), nil)
	if err != nil {
		t.Fatalf("ListTools() error = %v", err)
	}
	var schema struct {
		Properties map[string]struct {
			Description string `json:"description"`
		} `json:"properties"`
	}
	for _, tool := range listed.Tools {
		if tool.Name == "save_profile" {
			raw, _ := json.Marshal(tool.InputSchema)
			if err := json.Unmarshal(raw, &schema); err != nil {
				t.Fatalf("the input schema does not read: %v", err)
			}
		}
	}

	for field, want := range map[string]string{
		"pseudonym": fmt.Sprintf("at most %d characters", profile.MaxPseudonym),
		"grade":     fmt.Sprintf("%d to %d", profile.MinGrade, profile.MaxGrade),
		"interests": fmt.Sprintf("at most %d of at most %d characters", profile.MaxInterests, profile.MaxInterest),
		"notes":     fmt.Sprintf("at most %d characters", profile.MaxNotes),
	} {
		if got := schema.Properties[field].Description; !strings.Contains(got, want) {
			t.Errorf("%s is described as %q, want it to say %q", field, got, want)
		}
	}
}

// With no profile yet, every tool that reads one says so in the words the
// model relays and shows the first sign-in, rather than failing.
func TestWithNoProfileEveryReaderSaysSo(t *testing.T) {
	t.Parallel()

	_, session := lesson(t, memory.New())
	for _, tool := range []string{"get_profile", "get_progress", "read_progress"} {
		result := call(t, session, tool, nil)
		if result.IsError {
			t.Fatalf("%s failed: %s", tool, textOf(t, result))
		}
		payload := payloadOf[progressPayload](t, result)
		if payload.Screen != "first_run" || payload.Profile != nil {
			t.Errorf("%s: screen %q with profile %v, want first_run and none", tool, payload.Screen, payload.Profile)
		}
		if text := textOf(t, result); !strings.Contains(text, "There is no profile yet") ||
			!strings.Contains(text, "save_profile") {
			t.Errorf("%s says %q, want it to say there is no profile and how to make one", tool, text)
		}
	}
}

// A profile is made in the chat, so the words of a first sign-in lead the
// conversation that makes it: the adult says first that they are the child's
// parent or tutor, and only then is asked for a pseudonym — never the child's
// name, age or school.
func TestTheFirstSignInLeadsTheConversationThatMakesAProfile(t *testing.T) {
	t.Parallel()

	_, session := lesson(t, memory.New())
	text := textOf(t, call(t, session, "get_profile", nil))
	confirm := strings.Index(text, "First ask the adult to say they are the child's parent or tutor")
	pseudonym := strings.Index(text, "Then ask for a pseudonym — never the child's real name, birth date, age or school")
	create := strings.Index(text, "Then create the profile with save_profile")
	if confirm < 0 || pseudonym <= confirm || create <= pseudonym {
		t.Errorf("the words are %q, want the adult's word first, then the pseudonym alone, then the profile made", text)
	}
}

// A profile is not made without a pseudonym and a grade: the call is refused
// field by field, and nothing is kept.
func TestAProfileIsNotMadeWithoutAPseudonymAndAGrade(t *testing.T) {
	t.Parallel()

	kept := memory.New()
	_, session := lesson(t, kept)

	refused := payloadOf[profilePayload](t, call(t, session, "save_profile", map[string]any{"interests": []string{"space"}}))
	if refused.Status != "rejected" || refused.Code != "invalid_profile" || refused.Screen != "first_run" {
		t.Errorf("status %q, code %q, screen %q; want rejected, invalid_profile, first_run",
			refused.Status, refused.Code, refused.Screen)
	}
	want := []problemPayload{{Field: "grade", Code: "out_of_range"}, {Field: "pseudonym", Code: "required"}}
	if got := withoutRules(refused.Problems); !slices.Equal(got, want) {
		t.Errorf("problems %+v, want %+v", got, want)
	}
	if _, _, err := kept.Load(t.Context(), devAccount); !errors.Is(err, store.ErrNotFound) {
		t.Errorf("Load() after a refusal error = %v, want no profile kept", err)
	}
}

// A profile made of a pseudonym and a grade starts where the grade puts it, in
// the trial series, with a new topic to begin on, and says "none" with an
// empty list rather than with null.
func TestAProfileIsMadeOfAPseudonymAndAGrade(t *testing.T) {
	t.Parallel()

	kept := memory.New()
	_, session := lesson(t, kept)

	made := call(t, session, "save_profile", map[string]any{"pseudonym": "Comet", "grade": 3})
	payload := payloadOf[profilePayload](t, made)
	if payload.Screen != "profile" || payload.Profile == nil || payload.Profile.Pseudonym != "Comet" || payload.Profile.Grade != 3 {
		t.Fatalf("payload = %+v, want the profile of Comet in grade 3", payload)
	}
	if payload.Trial == nil || *payload.Trial != (trialPayload{Answered: 0, Of: rating.TrialAnswers}) {
		t.Errorf("trial = %v, want none of %d done", payload.Trial, rating.TrialAnswers)
	}
	if payload.Recommendation == nil || payload.Recommendation.Goal != string(profile.GoalNewTopic) {
		t.Errorf("recommendation = %+v, want a new topic to start with", payload.Recommendation)
	}
	for _, list := range []string{`"interests":[]`, `"excluded_skills":[]`} {
		if !strings.Contains(string(rawPayload(t, made)), list) {
			t.Errorf("the payload does not say %s", list)
		}
	}

	p, _ := loadKept(t, kept)
	if p.Ratings.Start != rating.Start(3) || p.Student.Interests == nil || p.Student.ExcludedSkills == nil {
		t.Errorf("kept: start %v, interests %#v, skills %#v; want the start of grade 3 and empty lists",
			p.Ratings.Start, p.Student.Interests, p.Student.ExcludedSkills)
	}
}

// A refused edit writes nothing and repeats nothing: the model is told which
// fields broke which rule, and the line of the call says it was refused.
func TestARefusedEditWritesNothingAndRepeatsNothing(t *testing.T) {
	t.Parallel()

	kept := keptWith(t, "masha")
	_, before := loadKept(t, kept)
	h, session := lesson(t, kept)

	notes := strings.Repeat("Loves long sums. ", 40)
	result := call(t, session, "save_profile", map[string]any{"grade": 9, "notes": notes})
	payload := payloadOf[profilePayload](t, result)
	want := []problemPayload{{Field: "grade", Code: "out_of_range"}, {Field: "notes", Code: "too_long"}}
	if got := withoutRules(payload.Problems); payload.Status != "rejected" || !slices.Equal(got, want) {
		t.Errorf("status %q with problems %+v, want rejected with %+v", payload.Status, got, want)
	}
	if payload.Profile == nil || payload.Recommendation == nil || payload.Trial == nil {
		t.Errorf("profile %v, recommendation %v, trial %v; want the profile kept shown whole beside the refusal",
			payload.Profile, payload.Recommendation, payload.Trial)
	}
	if strings.Contains(textOf(t, result), "Loves long sums") {
		t.Errorf("the refusal repeats the notes it refused: %q", textOf(t, result))
	}
	if _, after := loadKept(t, kept); after != before {
		t.Errorf("revision %s after a refusal, want %s: nothing written", after, before)
	}

	h.settle()
	line := h.lineOf(t, "save_profile")
	if got := field(t, line, "outcome"); got != "refused" {
		t.Errorf("outcome = %q, want refused", got)
	}
}

// The grade decides where a child starts, once. Changed later it is a label:
// the start, the level and the trial series stay where the answers put them.
func TestALaterGradeMovesNoRating(t *testing.T) {
	t.Parallel()

	kept := keptWith(t, "masha")
	was, _ := loadKept(t, kept)
	_, session := lesson(t, kept)

	payload := payloadOf[profilePayload](t, call(t, session, "save_profile", map[string]any{"grade": 6}))
	if payload.Profile == nil || payload.Profile.Grade != 6 {
		t.Fatalf("profile = %+v, want grade 6", payload.Profile)
	}
	now, _ := loadKept(t, kept)
	if now.Ratings != was.Ratings || now.Revision != was.Revision+1 {
		t.Errorf("ratings %+v at revision %d, want %+v at %d", now.Ratings, now.Revision, was.Ratings, was.Revision+1)
	}
}

// An edit that asks for what the profile already says writes nothing, so
// that the same call made twice costs one write.
func TestAnEditThatChangesNothingWritesNothing(t *testing.T) {
	t.Parallel()

	kept := keptWith(t, "masha")
	_, before := loadKept(t, kept)
	_, session := lesson(t, kept)

	result := call(t, session, "save_profile", map[string]any{"pseudonym": " Masha ", "interests": []string{"space", "robots"}})
	if !strings.HasPrefix(textOf(t, result), "Nothing changed") {
		t.Errorf("text = %q, want it to say nothing changed", textOf(t, result))
	}
	if _, after := loadKept(t, kept); after != before {
		t.Errorf("revision %s after an edit that changed nothing, want %s", after, before)
	}
}

// The progress is one answer, whoever asks: the model's tool draws it in a
// card of its own, and the card's tool draws it inside the card that asked.
func TestTheProgressIsOneAnswerForTheModelAndTheCard(t *testing.T) {
	t.Parallel()

	_, session := lesson(t, keptWith(t, "olya"))
	forModel := rawPayload(t, call(t, session, "get_progress", nil))
	forCard := rawPayload(t, call(t, session, "read_progress", nil))
	if !bytes.Equal(forModel, forCard) {
		t.Errorf("get_progress and read_progress differ:\n%s\n%s", forModel, forCard)
	}

	var payload progressPayload
	if err := json.Unmarshal(forModel, &payload); err != nil {
		t.Fatalf("the payload does not read: %v", err)
	}
	if payload.Screen != "progress" || payload.Trial != nil || payload.Overall == nil {
		t.Fatalf("screen %q, trial %v, overall %v; want the progress past the trial series", payload.Screen,
			payload.Trial, payload.Overall)
	}
	if payload.Overall.Ranks != rating.Ranks || payload.Overall.Rank < 1 || payload.Overall.Rank > rating.Ranks ||
		payload.Overall.Share < 0 || payload.Overall.Share > 100 {
		t.Errorf("overall = %+v, want a rank out of %d and a share of it", *payload.Overall, rating.Ranks)
	}
	matchesTheLevels(t, payload.Overall.Grades)
	rated := 0
	for _, topic := range payload.Topics {
		if !topic.standsTogether() || (topic.Rating != nil) != (topic.Answers > 0) {
			t.Errorf("topic %+v, want its rating, rank, share and comparison together, and there once it has an answer",
				topic)
		}
		if topic.Rating != nil {
			rated++
		}
	}
	if rated == 0 || rated == len(payload.Topics) {
		t.Errorf("topics = %+v, want the topics met with their ranks and some within reach not met yet", payload.Topics)
	}
}

// matchesTheLevels checks the runs of ranks a progress hands a card: one for
// each level, in ladder order, each the ranks the level is matched with.
func matchesTheLevels(t *testing.T, grades []gradesPayload) {
	t.Helper()

	levels := rating.GradeLevels()
	if len(grades) != len(levels) {
		t.Fatalf("overall grades = %+v, want a run of ranks for each of the %d levels", grades, len(levels))
	}
	for i, level := range levels {
		want := gradesPayload{GradeLevel: string(level), FirstRank: level.FirstRank(), LastRank: level.LastRank()}
		if grades[i] != want {
			t.Errorf("grades[%d] = %+v, want %+v", i, grades[i], want)
		}
	}
}

// A progress kept in memory, nowhere a person could open, says nothing of where
// it is kept: no location for a card to draw the parent's data from, and no
// word of a Drive.
func TestAProgressKeptInMemorySaysNowhere(t *testing.T) {
	t.Parallel()

	_, session := lesson(t, keptWith(t, "olya"))
	result := call(t, session, "get_progress", nil)
	if raw := rawPayload(t, result); bytes.Contains(raw, []byte(`"location"`)) {
		t.Errorf("the payload is %s, want no location in it", raw)
	}
	if text := textOf(t, result); strings.Contains(text, "Google Drive") {
		t.Errorf("the words are %q, want nothing of a Drive", text)
	}
}

// The parent's notes reach the model with the profile, so that it can read
// back what the parent wrote, and never reach the progress: a card that opens
// the progress is what the child looks at.
func TestTheNotesReachTheProfileAndNeverTheProgress(t *testing.T) {
	t.Parallel()

	kept := keptWith(t, "masha")
	p, _ := loadKept(t, kept)
	_, session := lesson(t, kept)

	if got := payloadOf[profilePayload](t, call(t, session, "get_profile", nil)).Profile; got == nil ||
		got.Notes != p.Student.Notes {
		t.Errorf("the profile carries the notes %v, want %q", got, p.Student.Notes)
	}
	for _, tool := range []string{"get_progress", "read_progress"} {
		result := call(t, session, tool, nil)
		if raw := string(rawPayload(t, result)); strings.Contains(raw, `"notes"`) || strings.Contains(raw, "long questions") {
			t.Errorf("%s carries the notes in its payload: %s", tool, raw)
		}
		if strings.Contains(textOf(t, result), "long questions") {
			t.Errorf("%s carries the notes in its words: %s", tool, textOf(t, result))
		}
	}
}

// The prototype's own cases, as this service has them. Masha is three answers
// into the trial series with a failure behind her: the prototype would have
// gone over the topic she failed, and the trial series moves to a new topic
// instead. Her open task is not an answer.
func TestTheProfileOfMasha(t *testing.T) {
	t.Parallel()

	_, session := lesson(t, keptWith(t, "masha"))
	profiled := payloadOf[profilePayload](t, call(t, session, "get_profile", nil))
	if profiled.Recommendation == nil || profiled.Recommendation.Goal != string(profile.GoalNewTopic) {
		t.Errorf("recommendation = %+v, want a new topic in the trial series", profiled.Recommendation)
	}
	if profiled.LastAnswer == nil || profiled.LastAnswer.TaskID != "tsk_masha_03" {
		t.Errorf("last_answer = %+v, want her last answer and not her open task", profiled.LastAnswer)
	}
}

// Masha's progress shows how far her trial series has got and no rating
// anywhere, the topic she mastered as mastered, and her answers the latest
// first.
func TestTheProgressOfMasha(t *testing.T) {
	t.Parallel()

	_, session := lesson(t, keptWith(t, "masha"))
	progressed := payloadOf[progressPayload](t, call(t, session, "get_progress", nil))
	if progressed.Trial == nil || *progressed.Trial != (trialPayload{Answered: 3, Of: rating.TrialAnswers}) ||
		progressed.Overall != nil {
		t.Errorf("trial %v with overall %v, want 3 of %d and no rating", progressed.Trial, progressed.Overall,
			rating.TrialAnswers)
	}
	mastered := false
	for _, topic := range progressed.Topics {
		if topic.Rating != nil || !topic.standsTogether() {
			t.Errorf("%s stands at %+v during the trial series, want nowhere", topic.Topic, topic)
		}
		mastered = mastered || topic.Topic == "logic.knights_liars" && topic.Mastered
	}
	if !mastered {
		t.Errorf("topics = %+v, want logic.knights_liars shown as mastered", progressed.Topics)
	}
	if len(progressed.Recent) != 3 || progressed.Recent[0].Topic != "time.clocks" {
		t.Errorf("recent = %+v, want her three answers, the latest first", progressed.Recent)
	}
}

// A child who has answered nothing yet has the trial series ahead, and the
// topics the rule could set first, none of them met and none with a rank.
func TestTheProgressOfSasha(t *testing.T) {
	t.Parallel()

	_, session := lesson(t, keptWith(t, "sasha"))
	progressed := payloadOf[progressPayload](t, call(t, session, "get_progress", nil))
	if progressed.Trial == nil || progressed.Trial.Answered != 0 || len(progressed.Topics) == 0 ||
		len(progressed.Recent) != 0 || progressed.Skipped != 0 || progressed.LastAnswer != nil {
		t.Errorf("progress = %+v, want the trial series not begun and the topics within reach", progressed)
	}
	for _, topic := range progressed.Topics {
		if topic.Answers != 0 || topic.Skipped != 0 || topic.Rating != nil || !topic.standsTogether() {
			t.Errorf("topic %+v, want one within reach, not met yet and with no rank", topic)
		}
	}
}

// keptAs is a store holding a fixture changed the way a case needs it.
func keptAs(t *testing.T, student string, change func(p *profile.Profile)) store.Storage {
	t.Helper()

	p, _ := loadKept(t, keptWith(t, student))
	change(p)
	kept := memory.New()
	if _, err := kept.Create(context.Background(), devAccount, p); err != nil {
		t.Fatalf("keep the changed fixture of %s: %v", student, err)
	}
	return kept
}

// The words the model relays say what the card shows: the details as they
// are, the last answer and how it went, and what comes next and why — in the
// catalog's names, or by its id a topic the catalog no longer has.
func TestTheWordsForTheModelSayWhatTheCardShows(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name string
		kept store.Storage
		tool string
		want []string
	}{
		{"a right answer", keptWith(t, "olya"), "get_profile",
			[]string{"was right", "The profile of \"Olya\", grade 2.", "Interests: \"horses\", \"drawing\".",
				"Dividing whole numbers with a remainder; Simple fractions",
				"The adult's notes, as the adult wrote them: \"Careful and accurate when counting objects and gaps.",
				"Next, the rule suggests Ordering (Restore an order from comparisons: who stands behind whom, who is older or taller) at level"}},
		{"the topics met, each with what it is", keptWith(t, "masha"), "get_progress",
			[]string{"Knights and liars (Knights always tell the truth and liars always lie; decide which statements are true and who said them), 1 of 1 right"}},
		{"a failure to go over", keptAs(t, "olya", func(p *profile.Profile) {
			p.Ratings.ConsecutiveFailures = 1
			last := &p.Recent[len(p.Recent)-1]
			last.Correct, last.Chosen = false, "B"
			tag := "pt-BR"
			p.Student.UILanguage = &tag
		}), "get_profile",
			[]string{"was wrong", "to go over it again after a mistake", "The lessons are in pt-BR: the tasks, the cards and your words."}},
		{"a topic the catalog no longer has", keptAs(t, "masha", func(p *profile.Profile) {
			p.Recent[len(p.Recent)-1].Topic = "clocks.sundials"
		}), "get_progress",
			[]string{"clocks.sundials wrong", "3 of 5 tasks done"}},
		{"a note and an interest the parent wrote", keptAs(t, "petya", func(p *profile.Profile) {
			p.Student.Notes = "Loves \"space\" and \U0001F469\u200d\U0001F4BB."
			p.Student.Interests = []string{"rockets"}
		}), "get_profile",
			[]string{"The adult's notes, as the adult wrote them: \"Loves \\\"space\\\" and \U0001F469\u200d\U0001F4BB.\".", "Interests: \"rockets\"."}},
		{"a skill the catalog no longer has", keptAs(t, "petya", func(p *profile.Profile) {
			p.Student.ExcludedSkills = []string{"long_division_by_hand"}
		}), "get_profile",
			[]string{"Left out of the tasks: long_division_by_hand."}},
		{"the overall rank and how far through it", keptWith(t, "olya"), "get_progress",
			[]string{"Overall rating 1486, rank 2 of 11, 91% of the way to rank 3."}},
		{"the grades each run of ranks matches", keptWith(t, "olya"), "get_progress",
			[]string{"The card marks under the ranks the grades whose tasks each run of them roughly matches: " +
				"ranks 1 to 4 grades 1-2, ranks 5 to 7 grades 3-4, ranks 8 to 11 grades 5-6; rank 2 is in the run of grades 1-2."}},
		{"the highest rank", keptAs(t, "petya", func(p *profile.Profile) {
			p.Ratings.Theta = 9
		}), "get_progress",
			[]string{"rank 11 of 11, the highest.", "rank 11 is in the run of grades 5-6."}},
		{"a topic's rank beside the overall one", keptAs(t, "petya", func(p *profile.Profile) {
			for id, delta := range map[string]float64{"arithmetic.tricks": 1, "combinatorics.enumeration": -1} {
				kept := p.Topics[id]
				kept.Delta = delta
				p.Topics[id] = kept
			}
		}), "get_progress",
			[]string{"rating 2159, rank 6, ahead of the overall rank", "rating 1811, rank 4, behind the overall rank"}},
		{"the topics within reach not met yet", keptWith(t, "sasha"), "get_progress",
			[]string{"No topic has been answered yet. Not met yet, and within reach now: Ordering, Knights and liars,"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			_, session := lesson(t, tc.kept)
			text := textOf(t, call(t, session, tc.tool, nil))
			for _, want := range tc.want {
				if !strings.Contains(text, want) {
					t.Errorf("%s says %q, want it to say %q", tc.tool, text, want)
				}
			}
		})
	}
}

// unreadable is a store that cannot read the profile it keeps, for a reason
// whose words are not ours to pass on.
type unreadable struct{ store.Storage }

func (unreadable) Load(context.Context, store.Account) (*profile.Profile, store.Revision, error) {
	return nil, "", errors.New("drive: the file of " + pseudonym + " does not parse")
}

// unwritable is a store with no profile that cannot make one.
type unwritable struct{ store.Storage }

func (unwritable) Create(context.Context, store.Account, *profile.Profile) (store.Revision, error) {
	return "", errors.New("drive: the folder of " + pseudonym + " is full")
}

// A store that fails is told in the one sentence of ours, whichever tool met
// it, and never in the store's own words.
func TestAStoreThatFailsIsToldInOurWords(t *testing.T) {
	t.Parallel()

	h, session := lesson(t, unreadable{memory.New()})
	for _, tool := range []string{"get_profile", "get_progress", "read_progress"} {
		wantOurSentence(t, call(t, session, tool, nil), "Something went wrong inside MathTrail.")
	}
	wantOurSentence(t, call(t, session, "save_profile", map[string]any{"grade": 2}),
		"Something went wrong inside MathTrail.")
	wantOurSentence(t, call(t, session, "edit_profile", map[string]any{"grade": 2}),
		"Something went wrong inside MathTrail.")
	wantOurSentence(t, call(t, session, "next_task", map[string]any{"language": "en"}),
		"Something went wrong inside MathTrail.")
	wantOurSentence(t, call(t, session, "submit_task", raceOn(openRace(t))),
		"Something went wrong inside MathTrail.")
	h.settle()
	wantFailed(t, h, "get_profile", "internal")

	_, session = lesson(t, unwritable{memory.New()})
	wantOurSentence(t, call(t, session, "save_profile", map[string]any{"pseudonym": "Comet", "grade": 2}),
		"Something went wrong inside MathTrail.")
}

// conflicted is a store where every save loses to a write made in between.
type conflicted struct{ store.Storage }

func (conflicted) Save(context.Context, store.Account, *profile.Profile, store.Revision) (store.Revision, error) {
	return "", store.ErrConflict
}

// A save that lost to a write made in between is told in words of its own:
// nothing was saved, and the same call made again starts from what won.
func TestASaveThatLostToAnotherWriteIsToldSo(t *testing.T) {
	t.Parallel()

	h, session := lesson(t, conflicted{keptWith(t, "masha")})
	result := call(t, session, "save_profile", map[string]any{"grade": 4})
	wantOurSentence(t, result, "The child's profile was changed somewhere else at the same moment")
	wantOurSentence(t, call(t, session, "next_task", map[string]any{"language": "en"}),
		"The child's profile was changed somewhere else at the same moment")

	h.settle()
	wantFailed(t, h, "save_profile", "conflict")

	// A task handed in writes the profile whichever way its review goes, and
	// a review whose outcome was not kept is not counted in the log either.
	for _, handedIn := range []func(*profile.OpenRequest) map[string]any{raceOn, broken} {
		asked := fuzzProfile(t)
		h, session := lesson(t, conflicted{keptAsIs(t, asked)})
		wantOurSentence(t, call(t, session, "submit_task", handedIn(asked.OpenRequest)),
			"The child's profile was changed somewhere else at the same moment")
		h.settle()
		if counted := len(linesOf(h, "task_submitted")); counted != 0 {
			t.Errorf("task_submitted lines = %d for a review that was not kept, want none", counted)
		}
	}
}

// Nothing a parent typed about the child reaches a span or a line: not the
// pseudonym, not the interests, not the notes.
func TestNothingTheParentTypedReachesASpanOrALine(t *testing.T) {
	t.Parallel()

	const (
		name     = "Mariya-the-Comet"
		interest = "ballet-and-dinosaurs"
		notes    = "Stutters-when-rushed"
	)
	h, session := lesson(t, memory.New())
	call(t, session, "save_profile", map[string]any{"pseudonym": name, "grade": 2, "interests": []string{interest}, "notes": notes})
	call(t, session, "save_profile", map[string]any{"pseudonym": name + name + name, "notes": strings.Repeat(notes, 40)})
	call(t, session, "edit_profile", map[string]any{"interests": []string{interest, interest + interest}})
	call(t, session, "edit_profile", map[string]any{"pseudonym": name + name + name, "interests": []string{strings.Repeat(interest, 3)}})
	for _, tool := range []string{"get_profile", "get_progress", "read_progress"} {
		call(t, session, tool, nil)
	}
	h.settle()

	secrets := []string{name, interest, notes}
	for _, span := range h.spans.Ended() {
		wantNoneOf(t, "span "+span.Name(), spanTexts(span), secrets)
	}
	lines := h.logs.All()
	for i := range lines {
		wantNoneOf(t, "line "+lines[i].Message, lineTexts(t, &lines[i]), secrets)
	}
}

// The tools are not built without any part they work with: a part missing
// would fail in the middle of a child's lesson rather than at the start.
func TestAServiceWithAPartMissingIsNotBuilt(t *testing.T) {
	t.Parallel()

	if _, err := mcpserver.NewService(allParts(t)); err != nil {
		t.Fatalf("NewService() with every part error = %v, want nil", err)
	}
	if _, err := mcpserver.NewService(nil); !errors.Is(err, mcpserver.ErrSettings) {
		t.Errorf("NewService(nil) error = %v, want it refused", err)
	}
	for _, tc := range []struct {
		name string
		drop func(*mcpserver.Parts)
	}{
		{"no store", func(p *mcpserver.Parts) { p.Store = nil }},
		{"no content", func(p *mcpserver.Parts) { p.Content = nil }},
		{"no checks", func(p *mcpserver.Parts) { p.Reviewer = nil }},
		{"no seal", func(p *mcpserver.Parts) { p.Sealer = nil }},
		{"no window", func(p *mcpserver.Parts) { p.Window = 0 }},
		{"no task in a day", func(p *mcpserver.Parts) { p.Daily.Tasks = 0 }},
		{"no failure in a day", func(p *mcpserver.Parts) { p.Daily.Failed = 0 }},
		{"a mistake that repeats once", func(p *mcpserver.Parts) { p.TrapRepeats = 1 }},
		{"a mistake that repeats more often than the window holds", func(p *mcpserver.Parts) {
			p.TrapRepeats = profile.MaxRecent + 1
		}},
		{"no clock", func(p *mcpserver.Parts) { p.Now = nil }},
		{"no version", func(p *mcpserver.Parts) { p.Version = "" }},
		{"no logger", func(p *mcpserver.Parts) { p.Logger = nil }},
		{"no traces", func(p *mcpserver.Parts) { p.Traces = nil }},
		{"no site", func(p *mcpserver.Parts) { p.SiteURL = "" }},
		{"no key to count children under", func(p *mcpserver.Parts) { p.Learners = nil }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			parts := allParts(t)
			tc.drop(parts)
			if _, err := mcpserver.NewService(parts); !errors.Is(err, mcpserver.ErrSettings) {
				t.Errorf("NewService() error = %v, want it refused", err)
			}
		})
	}
}

// allParts are every part the tools of the lesson are built from.
func allParts(t *testing.T) *mcpserver.Parts {
	t.Helper()

	loaded, err := shipped()
	if err != nil {
		t.Fatalf("load the content: %v", err)
	}
	runner, err := sandbox()
	if err != nil {
		t.Fatalf("build the sandbox: %v", err)
	}
	return &mcpserver.Parts{
		Store:       memory.New(),
		Content:     loaded,
		Reviewer:    checks.NewReviewer(loaded, runner, checks.DefaultDrawingLimits()),
		Sealer:      sealer(t),
		Window:      time.Minute,
		Daily:       mcpserver.Daily{Tasks: config.DefaultDailyTasks, Failed: config.DefaultDailyFailed},
		TrapRepeats: config.DefaultTrapRepeats,
		Now:         func() time.Time { return lessonDay },
		Version:     "test",
		Logger:      zap.NewNop(),
		Traces:      tracenoop.NewTracerProvider(),
		Learners:    learners(t),
		SiteURL:     config.DefaultSiteURL,
	}
}

// The map of misconceptions lists the traps the child fell for again among the
// latest answers, by their ids and how many times, the most frequent first, and
// the words name them for the model by what the catalog says of them. Olya fell
// for double_count once; once more puts it on the map.
func TestTheProgressMapsTheMistakesThatRepeat(t *testing.T) {
	t.Parallel()

	kept := keptAs(t, "olya", func(p *profile.Profile) {
		p.Recent = append(p.Recent, profile.Answer{
			AnsweredAt: profile.At(lessonDay), Chosen: "B", Difficulty: 3, GradeLevel: rating.Grades34,
			Pace: profile.PaceNormal, TaskID: "tsk_olya_again", Topic: "combinatorics.enumeration",
			Trap: "double_count",
		})
	})
	_, session := lesson(t, kept)

	result := call(t, session, "get_progress", nil)
	want := []mistakePayload{{Trap: "double_count", Times: 2}}
	if got := payloadOf[progressPayload](t, result).Mistakes; !slices.Equal(got, want) {
		t.Errorf("mistakes = %+v, want %+v", got, want)
	}
	if text := textOf(t, result); !strings.Contains(text,
		"Mistakes that keep coming back among the latest answers, the most frequent first: "+
			"Counted the same option twice (2 times).") {
		t.Errorf("the words are %q, want the mistake named by the catalog", text)
	}
}

// A child with no mistake that repeats, and a parent with no profile yet, get
// an empty map rather than none: the card draws a list, never nothing.
func TestAMapWithNoMistakeIsEmpty(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name string
		kept store.Storage
	}{
		{"olya", keptWith(t, "olya")},
		{"no profile", memory.New()},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			_, session := lesson(t, tc.kept)
			result := call(t, session, "get_progress", nil)
			if raw := string(rawPayload(t, result)); !strings.Contains(raw, `"mistakes":[]`) {
				t.Errorf("the payload is %s, want an empty map", raw)
			}
			if strings.Contains(textOf(t, result), "Mistakes that keep coming back") {
				t.Errorf("the words name mistakes when there are none: %s", textOf(t, result))
			}
		})
	}
}

// The tools only a card calls are never named to the model in the server's
// instructions: it would call one, and a host that keeps them from it would
// refuse the call, or let through a change the adult never made on a form.
// Which tools those are is read from how each is listed.
func TestTheInstructionsNameNoToolOnlyACardCalls(t *testing.T) {
	t.Parallel()

	_, session := lesson(t, memory.New())
	listed, err := session.ListTools(t.Context(), nil)
	if err != nil {
		t.Fatalf("ListTools() error = %v", err)
	}
	loaded, err := shipped()
	if err != nil {
		t.Fatalf("load the content: %v", err)
	}
	text := loaded.ServerInstructions()
	cardOnly := 0
	for _, tool := range listed.Tools {
		ui, _ := tool.Meta["ui"].(map[string]any)
		if visibility, _ := json.Marshal(ui["visibility"]); string(visibility) != `["app"]` {
			continue
		}
		cardOnly++
		if strings.Contains(text, tool.Name) {
			t.Errorf("the server instructions name %s, which only a card calls", tool.Name)
		}
	}
	if cardOnly == 0 {
		t.Error("no tool is listed for a card alone, want the ones only a card calls checked")
	}
}

package mcpserver_test

import (
	"errors"
	"slices"
	"strings"
	"testing"

	"github.com/MathTrail/mathtrail-standalone/internal/store"
	"github.com/MathTrail/mathtrail-standalone/internal/store/memory"
)

// editedPayload is what edit_profile hands back to the card, as the card reads
// it.
type editedPayload struct {
	Screen     string           `json:"screen"`
	Status     string           `json:"status"`
	Code       string           `json:"code"`
	Problems   []problemPayload `json:"problems"`
	Changed    bool             `json:"changed"`
	LastAnswer *struct {
		TaskID string `json:"task_id"`
	} `json:"last_answer"`
	Profile *struct {
		Pseudonym      string   `json:"pseudonym"`
		Grade          int      `json:"grade"`
		Interests      []string `json:"interests"`
		ExcludedSkills []string `json:"excluded_skills"`
		UILanguage     *string  `json:"ui_language"`
	} `json:"profile"`
}

// problemPayload is one field a refusal names: the field, the code of the rule
// it broke, and the rule in words.
type problemPayload struct {
	Field string `json:"field"`
	Code  string `json:"code"`
	Rule  string `json:"rule"`
}

// The form sends only what the adult changed, and only that changes: the rest
// of the details stay as they are kept — a change the model made in the chat
// while the form was open among them — and the parent's notes are neither
// touched nor handed back. The model is told of the change in words, for the
// card to pass on, and the notes are not in them either.
func TestTheFormChangesOnlyWhatItSends(t *testing.T) {
	t.Parallel()

	kept := keptWith(t, "masha")
	was, before := loadKept(t, kept)
	_, session := lesson(t, kept)

	if result := call(t, session, "save_profile", map[string]any{"interests": []string{"comets"}}); result.IsError {
		t.Fatalf("save_profile failed: %s", textOf(t, result))
	}
	result := call(t, session, "edit_profile", map[string]any{"grade": 4})
	edited := payloadOf[editedPayload](t, result)
	if !edited.Changed || edited.Status != "" || edited.Profile == nil || edited.Profile.Grade != 4 ||
		!slices.Equal(edited.Profile.Interests, []string{"comets"}) {
		t.Errorf("payload = %+v, want grade 4 changed beside the interests the chat saved", edited)
	}
	if edited.LastAnswer == nil || edited.LastAnswer.TaskID != "tsk_masha_03" {
		t.Errorf("last_answer = %+v, want Masha's last answer, as every result carries it", edited.LastAnswer)
	}
	if raw := string(rawPayload(t, result)); strings.Contains(raw, `"notes"`) {
		t.Errorf("the payload is %s, want no notes in it", raw)
	}
	words := textOf(t, result)
	if !strings.Contains(words, "The adult changed the child's profile with the form on a card.") ||
		!strings.Contains(words, "The profile of \"Masha\", grade 4.") {
		t.Errorf("the words are %q, want the change told with the details as they now stand", words)
	}
	if strings.Contains(words, "long questions") || strings.Contains(words, "notes") {
		t.Errorf("the words are %q, want nothing of the notes", words)
	}

	now, after := loadKept(t, kept)
	if now.Student.Notes != was.Student.Notes || now.Student.Grade != 4 || now.Revision != was.Revision+2 {
		t.Errorf("kept: notes %q, grade %d, revision %d; want the notes as they were, grade 4 and two writes",
			now.Student.Notes, now.Student.Grade, now.Revision)
	}
	if after == before {
		t.Error("the store's revision did not move, want the change written")
	}
}

// A form only ever changes a profile that is there. Sent for none — deleted
// since the card was drawn — it makes none, and says the profile is gone.
func TestTheFormNeverMakesAProfile(t *testing.T) {
	t.Parallel()

	kept := memory.New()
	_, session := lesson(t, kept)

	gone := payloadOf[editedPayload](t, call(t, session, "edit_profile", map[string]any{"pseudonym": "Comet", "grade": 2}))
	if gone.Status != "stale" || gone.Code != "stale_profile" || gone.Screen != "first_run" || gone.Changed ||
		gone.Profile != nil {
		t.Errorf("payload = %+v, want stale_profile with no profile", gone)
	}
	if _, _, err := kept.Load(t.Context(), devAccount); !errors.Is(err, store.ErrNotFound) {
		t.Errorf("Load() error = %v, want no profile made", err)
	}
}

// The form has no field for the parent's notes, and nothing on a card can
// start a profile over: the protocol refuses either before the tool runs, and
// nothing is written.
func TestTheFormTakesNoNotesAndNoStartingOver(t *testing.T) {
	t.Parallel()

	kept := keptWith(t, "masha")
	_, before := loadKept(t, kept)
	h, session := lesson(t, kept)

	for _, arguments := range []map[string]any{
		{"notes": "Loves comets."},
		{"start_over": true, "pseudonym": "Comet", "grade": 2},
	} {
		if result := call(t, session, "edit_profile", arguments); !result.IsError {
			t.Errorf("edit_profile(%v) was answered %q, want the arguments refused", arguments, textOf(t, result))
		}
	}
	if _, after := loadKept(t, kept); after != before {
		t.Errorf("revision %s after refused arguments, want %s: nothing written", after, before)
	}
	h.settle()
	lines := linesOf(h, "tool_call")
	if len(lines) != 2 {
		t.Fatalf("tool_call lines = %d, want one for each call", len(lines))
	}
	for _, line := range lines {
		if got := line.ContextMap()["outcome"]; got != "invalid" {
			t.Errorf("outcome = %v, want invalid", got)
		}
	}
}

// A change from the form that breaks a rule is refused field by field, each
// by the code of its rule, for the card to say in its own language. Nothing is
// written, and the profile comes back as it is kept.
func TestTheFormIsRefusedFieldByFieldByCode(t *testing.T) {
	t.Parallel()

	kept := keptWith(t, "masha")
	_, before := loadKept(t, kept)
	h, session := lesson(t, kept)

	result := call(t, session, "edit_profile", map[string]any{
		"pseudonym":       " ",
		"grade":           9,
		"interests":       strings.Fields("a b c d e f g h i j k"),
		"excluded_skills": []string{"calculus"},
		"ui_language":     "not a language",
	})
	refused := payloadOf[editedPayload](t, result)
	want := []problemPayload{
		{Field: "excluded_skills", Code: "not_in_catalog"},
		{Field: "grade", Code: "out_of_range"},
		{Field: "interests", Code: "too_many"},
		{Field: "pseudonym", Code: "required"},
		{Field: "ui_language", Code: "not_a_language"},
	}
	if got := withoutRules(refused.Problems); refused.Status != "rejected" || refused.Code != "invalid_profile" ||
		refused.Changed || !slices.Equal(got, want) {
		t.Errorf("status %q, code %q, changed %v, problems %+v; want rejected, invalid_profile, nothing changed and %+v",
			refused.Status, refused.Code, refused.Changed, got, want)
	}
	for _, problem := range refused.Problems {
		if problem.Rule == "" {
			t.Errorf("the problem with %s has no rule in words", problem.Field)
		}
	}
	if refused.Profile == nil || refused.Profile.Grade != 3 || refused.Profile.Pseudonym != "Masha" {
		t.Errorf("profile = %+v, want Masha's as it is kept", refused.Profile)
	}
	if words := textOf(t, result); strings.Contains(words, "calculus") || strings.Contains(words, "not a language") {
		t.Errorf("the refusal repeats what it refused: %q", words)
	}
	if _, after := loadKept(t, kept); after != before {
		t.Errorf("revision %s after a refusal, want %s: nothing written", after, before)
	}
	h.settle()
	if got := field(t, h.lineOf(t, "edit_profile"), "outcome"); got != "refused" {
		t.Errorf("outcome = %q, want refused", got)
	}
}

// withoutRules is the problems with their words left out, for a case to hold
// the fields and the codes to what it wants.
func withoutRules(problems []problemPayload) []problemPayload {
	bare := make([]problemPayload, 0, len(problems))
	for _, problem := range problems {
		bare = append(bare, problemPayload{Field: problem.Field, Code: problem.Code})
	}
	return bare
}

// A form saved with nothing different in it — a pseudonym the same once its
// spaces are dropped — writes nothing, and has nothing to tell the model.
func TestAFormThatChangesNothingWritesNothing(t *testing.T) {
	t.Parallel()

	kept := keptWith(t, "masha")
	_, before := loadKept(t, kept)
	_, session := lesson(t, kept)

	result := call(t, session, "edit_profile", map[string]any{"pseudonym": " Masha "})
	if edited := payloadOf[editedPayload](t, result); edited.Changed || edited.Status != "" {
		t.Errorf("payload = %+v, want nothing changed and nothing refused", edited)
	}
	if words := textOf(t, result); !strings.HasPrefix(words, "Nothing changed") {
		t.Errorf("the words are %q, want them to say nothing changed", words)
	}
	if _, after := loadKept(t, kept); after != before {
		t.Errorf("revision %s after a form that changed nothing, want %s", after, before)
	}
}

// The form chooses a language for the lessons, and gives the choice back to
// the chat with an empty text, as the model does.
func TestTheFormSetsTheLanguageAndGivesItBackToTheChat(t *testing.T) {
	t.Parallel()

	kept := keptWith(t, "masha")
	_, session := lesson(t, kept)

	chosen := payloadOf[editedPayload](t, call(t, session, "edit_profile", map[string]any{"ui_language": "pt-br"}))
	if chosen.Profile == nil || chosen.Profile.UILanguage == nil || *chosen.Profile.UILanguage != "pt-BR" {
		t.Errorf("profile = %+v, want the lessons in pt-BR", chosen.Profile)
	}
	back := payloadOf[editedPayload](t, call(t, session, "edit_profile", map[string]any{"ui_language": ""}))
	if back.Profile == nil || back.Profile.UILanguage != nil || !back.Changed {
		t.Errorf("profile = %+v, changed %v; want the chat's language again", back.Profile, back.Changed)
	}
}

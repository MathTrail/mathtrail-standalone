package mcpserver

import (
	"context"
	"errors"
	"fmt"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/profile"
	"github.com/MathTrail/mathtrail-standalone/internal/store"
)

// codeStaleProfile is a change sent from a card for a profile that is not
// there: deleted since the card was drawn, or never made in the account the
// card's call is signed in with.
const codeStaleProfile = "stale_profile"

// editProfileIn is what the form on a card sends: the details the adult
// changed, and no others, so that a change the model made in the chat while
// the form was open is not written over. A field left out stays as it is. The
// parent's notes and starting over are not among them, and the schema refuses
// either: the notes are the adult's words to the model, said in the chat, and
// starting over is for a file nothing can read, which no card shows.
type editProfileIn struct {
	Pseudonym      *string  `json:"pseudonym,omitempty" jsonschema:"what the child is called: a pseudonym, never a real name"`
	Grade          *int     `json:"grade,omitempty" jsonschema:"the school year, 1 to 6"`
	Interests      []string `json:"interests,omitempty" jsonschema:"what tasks may be dressed in. The list replaces the one kept; an empty list clears it"`
	ExcludedSkills []string `json:"excluded_skills,omitempty" jsonschema:"ids of the skills the child has not met at school yet. The list replaces the one kept; an empty list clears it"`
	UILanguage     *string  `json:"ui_language,omitempty" jsonschema:"the language of the lessons, as a BCP 47 tag. An empty text makes them follow the chat's language"`
	Country        *string  `json:"country,omitempty" jsonschema:"the country the family lives in, as an ISO 3166-1 alpha-2 code. An empty text clears it, and the state with it"`
	Region         *string  `json:"region,omitempty" jsonschema:"for a family in the United States, its state as an ISO 3166-2 code. An empty text clears it"`
}

// edit is the change the form asks for.
func (in *editProfileIn) edit() profile.Edit {
	return profile.Edit{
		Pseudonym:      in.Pseudonym,
		Grade:          in.Grade,
		Interests:      in.Interests,
		ExcludedSkills: in.ExcludedSkills,
		UILanguage:     in.UILanguage,
		Country:        in.Country,
		Region:         in.Region,
	}
}

// editedOut is what edit_profile hands back to the card: the details as they
// stand after the change, or unchanged beside the fields that broke a rule,
// whether anything was written, and the last answer, as every result has it.
// The parent's notes are never among them: the form has no field for them,
// and a card is where a child looks.
type editedOut struct {
	Screen     string       `json:"screen"`
	Status     string       `json:"status,omitempty"`
	Code       string       `json:"code,omitempty"`
	Problems   []problemOut `json:"problems,omitempty"`
	Changed    bool         `json:"changed"`
	LastAnswer *answerLine  `json:"last_answer"`
	Profile    *childOut    `json:"profile"`
}

func (s *Service) editProfileTool() Tool {
	return Define(Spec{
		Name:  "edit_profile",
		Title: "Change the child's profile from the card",
		Description: "Changes the child's details from the form in the progress's Profile section: the fields sent, " +
			"and no other. Only the card calls it.",
		Idempotent: true,
		WidgetOnly: true,
	}, s.editProfile)
}

// editProfile makes the change the form on a card asks for. A form only ever
// changes a profile that is there: one is made in the chat, where the adult
// says they are the child's parent or tutor.
func (s *Service) editProfile(ctx context.Context, account store.Account, in editProfileIn) (Reply[editedOut], error) {
	return afresh(ctx, func() (Reply[editedOut], error) { return s.writeEdit(ctx, account, &in) })
}

// writeEdit is one read of the profile and the write of the change, when
// there is one to write. The words are for the model, which the card hands
// them to once the profile has changed: the model is not called by the form,
// and would otherwise go on with the details it was told before.
func (s *Service) writeEdit(ctx context.Context, account store.Account, in *editProfileIn) (Reply[editedOut], error) {
	p, revision, err := s.store.Load(ctx, account)
	switch {
	case errors.Is(err, store.ErrNotFound):
		return Reply[editedOut]{
			Text:    "Nothing was saved: there is no profile to change.",
			Payload: editedOut{Screen: screenFirstRun, Status: statusStale, Code: codeStaleProfile},
		}, nil
	case err != nil:
		return Reply[editedOut]{}, fmt.Errorf("mcp: read the profile: %w", err)
	}

	edit := in.edit()
	changed, problems, err := s.saveChange(ctx, account, p, revision, &edit)
	if err != nil {
		return Reply[editedOut]{}, err
	}
	reply := Reply[editedOut]{Payload: editedOut{
		Screen: screenProfile, Changed: changed, LastAnswer: lastAnswerOf(p), Profile: childOf(&p.Student),
	}}
	switch {
	case len(problems) > 0:
		var lines string
		reply.Payload.Problems, lines = problemsOf(problems)
		reply.Payload.Status, reply.Payload.Code = statusRejected, codeInvalidProfile
		reply.Text = "Nothing was saved: " + lines + "."
	case changed:
		reply.Text = joined("The adult changed the child's profile with the form on a card.",
			s.detailsText(&p.Student), stillInText(p))
	default:
		reply.Text = "Nothing changed: the profile already says so."
	}
	return reply, nil
}

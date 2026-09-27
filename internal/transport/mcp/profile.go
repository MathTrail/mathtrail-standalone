package mcpserver

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/profile"
	"github.com/MathTrail/mathtrail-standalone/internal/domain/progress"
	"github.com/MathTrail/mathtrail-standalone/internal/store"
)

// codeInvalidProfile is the refusal of details that break a rule of the
// profile. It spends nothing: there is no attempt to lose.
const codeInvalidProfile = "invalid_profile"

// profileOut is what the two tools of the profile hand back: the child's
// details, how far the trial series has got and what comes next — or, when
// the details could not be kept, which fields broke which rule.
//
// The parent's notes are here so that the model can read back to the parent
// what they wrote about the child. The card drawn from the same answer leaves
// them off its screen.
type profileOut struct {
	Screen         string             `json:"screen"`
	Status         string             `json:"status,omitempty"`
	Code           string             `json:"code,omitempty"`
	Problems       []problemOut       `json:"problems,omitempty"`
	LastAnswer     *answerLine        `json:"last_answer"`
	Profile        *detailsOut        `json:"profile"`
	Trial          *trialOut          `json:"trial"`
	Recommendation *recommendationOut `json:"recommendation"`
}

// detailsOut are the child's details as the progress carries them, and the
// parent's notes beside them.
type detailsOut struct {
	childOut
	Notes string `json:"notes"`
}

// problemOut is one field that broke a rule, and the rule. It never repeats
// what the field held.
type problemOut struct {
	Field string `json:"field"`
	Rule  string `json:"rule"`
}

// saveProfileIn is what save_profile takes. Every field is optional, and one
// left out stays as it is. The limits are in the words rather than in the
// schema: the library that checks a schema quotes back the text that broke
// it, and the checks of the profile say what is wrong without doing so.
type saveProfileIn struct {
	Pseudonym      *string  `json:"pseudonym,omitempty" jsonschema:"what the child is called: a pseudonym of at most 32 characters, never a real name. Required to create the profile"`
	Grade          *int     `json:"grade,omitempty" jsonschema:"the school year, 1 to 6. Required to create the profile. It sets where the first tasks start; changed later it is only a label"`
	Interests      []string `json:"interests,omitempty" jsonschema:"what tasks may be dressed in, at most 10 of at most 40 characters each. The list replaces the one kept; an empty list clears it"`
	ExcludedSkills []string `json:"excluded_skills,omitempty" jsonschema:"ids of skills the child has not met at school yet, from the list in this tool's description. The list replaces the one kept; an empty list clears it"`
	Notes          *string  `json:"notes,omitempty" jsonschema:"what the adult wants known about the child, for pitching the words, at most 500 characters. An empty text clears it"`
	UILanguage     *string  `json:"ui_language,omitempty" jsonschema:"the language of the cards as a BCP 47 tag, such as en, ru or pt-BR. An empty text makes the cards follow the chat's language"`
}

// edit is the change the arguments ask for.
func (in *saveProfileIn) edit() profile.Edit {
	return profile.Edit{
		Pseudonym:      in.Pseudonym,
		Grade:          in.Grade,
		Interests:      in.Interests,
		ExcludedSkills: in.ExcludedSkills,
		Notes:          in.Notes,
		UILanguage:     in.UILanguage,
	}
}

func (s *Service) getProfileTool() Tool {
	return Define(Spec{
		Name:  "get_profile",
		Title: "Get the child's profile",
		Description: "Reads the child's profile — the pseudonym, the grade, the interests, the skills left out of " +
			"the tasks, the adult's notes and the language of the cards — and what the next task would be. " +
			"Call it first. When there is no profile yet it says so: ask the adult for a pseudonym and the grade, " +
			"then create the profile with save_profile. Every result carries last_answer, the last answer the child " +
			"gave, maybe on a card without you: read it before you say anything about the current task.",
		ReadOnly:   true,
		Idempotent: true,
		DrawsCard:  true,
	}, s.getProfile)
}

func (s *Service) saveProfileTool() Tool {
	return Define(Spec{
		Name:  "save_profile",
		Title: "Save the child's profile",
		Description: "Creates the child's profile, or changes it. Pass only what changes; a field left out stays " +
			"as it is. To create the profile, pseudonym and grade are required. The pseudonym is what the child is " +
			"called: never a real name, a birth date or a school. The grade only sets where the first tasks start; " +
			"changed later it moves no rating. When a field breaks a rule, nothing is saved and the result names " +
			"the field and the rule.\n\nSkills that can be left out of the tasks, by id:\n" + s.skillList(),
		Idempotent: true,
		DrawsCard:  true,
	}, s.saveProfile)
}

// skillList is the skill catalog as the model needs it to fill in the skills
// a child has not met: one id and what it means on each line.
func (s *Service) skillList() string {
	var lines strings.Builder
	for _, skill := range s.content.Skills() {
		fmt.Fprintf(&lines, "- %s: %s\n", skill.ID, skill.Description)
	}
	return strings.TrimSuffix(lines.String(), "\n")
}

func (s *Service) getProfile(ctx context.Context, account store.Account, _ noArguments) (Reply[profileOut], error) {
	p, _, err := s.store.Load(ctx, account)
	switch {
	case errors.Is(err, store.ErrNotFound):
		return Reply[profileOut]{Text: firstRunText, Payload: profileOut{Screen: screenFirstRun}}, nil
	case err != nil:
		return Reply[profileOut]{}, fmt.Errorf("mcp: read the profile: %w", err)
	}
	return s.profileReply(p, "")
}

// saveProfile creates the profile when there is none and changes it when
// there is. Nothing is written when the details break a rule, and nothing when
// they already say what the call asks for.
func (s *Service) saveProfile(ctx context.Context, account store.Account, in saveProfileIn) (Reply[profileOut], error) {
	edit := in.edit()
	p, revision, err := s.store.Load(ctx, account)
	switch {
	case errors.Is(err, store.ErrNotFound):
		return s.createProfile(ctx, account, &edit)
	case err != nil:
		return Reply[profileOut]{}, fmt.Errorf("mcp: read the profile: %w", err)
	}

	changed, problems := p.Change(&edit, s.content.HasSkill, s.version, s.now())
	if len(problems) > 0 {
		return s.refusal(p, problems)
	}
	if !changed {
		return s.profileReply(p, "Nothing changed: the profile already says so.")
	}
	if _, err := s.store.Save(ctx, account, p, revision); err != nil {
		return Reply[profileOut]{}, fmt.Errorf("mcp: save the profile: %w", err)
	}
	return s.profileReply(p, "Saved.")
}

// createProfile makes the first profile of an account. The grade decides
// where the child starts on the ladder, once, here.
func (s *Service) createProfile(ctx context.Context, account store.Account, edit *profile.Edit) (Reply[profileOut], error) {
	student, problems := profile.NewStudent(edit, s.content.HasSkill)
	if len(problems) > 0 {
		return s.refusal(nil, problems)
	}
	p := profile.New(student, s.version, s.now())
	if _, err := s.store.Create(ctx, account, p); err != nil {
		return Reply[profileOut]{}, fmt.Errorf("mcp: create the profile: %w", err)
	}
	return s.profileReply(p, "The profile is created.")
}

// profileReply is a profile as the two tools of the profile hand it back.
func (s *Service) profileReply(p *profile.Profile, lead string) (Reply[profileOut], error) {
	next, err := progress.Recommend(p, s.content)
	if err != nil {
		return Reply[profileOut]{}, err
	}
	trial := progress.TrialOf(p)
	return Reply[profileOut]{
		Text: joined(lead, s.detailsText(&p.Student), trialLine(trial), s.lastAnswerText(p), s.nextText(&next)),
		Payload: profileOut{
			Screen:         screenProfile,
			LastAnswer:     lastAnswerOf(p),
			Profile:        detailsOf(&p.Student),
			Trial:          trialOf(trial),
			Recommendation: recommendationOf(&next),
		},
	}, nil
}

// refusal is details that could not be kept, told field by field, beside the
// profile kept — shown whole and unchanged, as any other answer about it is —
// or beside the first sign-in when there is none yet.
func (s *Service) refusal(p *profile.Profile, problems []profile.Problem) (Reply[profileOut], error) {
	reply := Reply[profileOut]{Payload: profileOut{Screen: screenFirstRun}}
	if p != nil {
		var err error
		if reply, err = s.profileReply(p, ""); err != nil {
			return Reply[profileOut]{}, err
		}
	}
	lines := make([]string, 0, len(problems))
	for _, problem := range problems {
		reply.Payload.Problems = append(reply.Payload.Problems, problemOut{Field: problem.Field, Rule: problem.Rule})
		lines = append(lines, problem.String())
	}
	reply.Payload.Status, reply.Payload.Code = "rejected", codeInvalidProfile
	reply.Text = "Nothing was saved. Fix these fields and call save_profile again: " + strings.Join(lines, "; ") + "."
	return reply, nil
}

func detailsOf(s *profile.Student) *detailsOut {
	return &detailsOut{childOut: *childOf(s), Notes: s.Notes}
}

// detailsText is the child's details in words, a sentence for each.
func (s *Service) detailsText(student *profile.Student) string {
	language := "The cards follow the chat's language."
	if student.UILanguage != nil {
		language = "The cards are in " + *student.UILanguage + "."
	}
	skills := make([]string, 0, len(student.ExcludedSkills))
	for _, id := range student.ExcludedSkills {
		skills = append(skills, s.skillName(id))
	}
	return joined(
		fmt.Sprintf("The profile of %s, grade %d.", quoted(student.Pseudonym), student.Grade),
		"Interests: "+listed(quotedEach(student.Interests), ", ")+".",
		"Left out of the tasks: "+listed(skills, "; ")+".",
		notesText(student.Notes),
		language,
	)
}

// notesText is the parent's notes as the model reads them, quoted: what the
// parent wrote about the child, not something the service says.
func notesText(notes string) string {
	if notes == "" {
		return "The adult's notes: none."
	}
	return "The adult's notes, as the adult wrote them: " + quoted(notes) + "."
}

// listed is a list in words, or none.
func listed(items []string, separator string) string {
	if len(items) == 0 {
		return "none"
	}
	return strings.Join(items, separator)
}

// joined puts the sentences that say something one after the other.
func joined(sentences ...string) string {
	kept := make([]string, 0, len(sentences))
	for _, sentence := range sentences {
		if sentence != "" {
			kept = append(kept, sentence)
		}
	}
	return strings.Join(kept, " ")
}

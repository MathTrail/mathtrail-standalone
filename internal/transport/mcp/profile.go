package mcpserver

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/country"
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
// what they wrote about the child. Neither tool draws a card; a card drawn
// from this payload all the same — a host may still draw one, as it drew
// one in an earlier chat — leaves the notes off its screen.
type profileOut struct {
	Screen         string             `json:"screen"`
	Status         string             `json:"status,omitempty"`
	Code           string             `json:"code,omitempty"`
	Problems       []problemOut       `json:"problems,omitempty"`
	LastAnswer     *answerLine        `json:"last_answer"`
	Profile        *detailsOut        `json:"profile"`
	Trial          *trialOut          `json:"trial"`
	Recommendation *recommendationOut `json:"recommendation"`
	Location       *locationOut       `json:"location,omitempty"`
}

// detailsOut are the child's details as the progress carries them, and the
// parent's notes beside them.
type detailsOut struct {
	childOut
	Notes string `json:"notes"`
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
	UILanguage     *string  `json:"ui_language,omitempty" jsonschema:"the language of the lessons — the tasks, the cards and your words — as a BCP 47 tag, such as en, ru or pt-BR. An empty text makes them follow the chat's language"`
	Country        *string  `json:"country,omitempty" jsonschema:"the country the family lives in, as an ISO 3166-1 alpha-2 code such as US or FR. It is kept only to count families by country: set it only when the adult says it of their own accord, and never ask for it. An empty text clears it, and the state with it"`
	Region         *string  `json:"region,omitempty" jsonschema:"for a family in the United States, its state as an ISO 3166-2 code such as US-TX, set only when the adult says it of their own accord. An empty text clears it"`
	LessonTopic    *string  `json:"lesson_topic,omitempty" jsonschema:"the topic to keep the lessons to, by its id from the list in the description of next_task: once the trial series is over, every task is on it until the choice is given back. Set it only when the child or the adult asks to keep to one topic; an empty text gives the choice back to the rule"`
	StartOver      bool     `json:"start_over,omitempty" jsonschema:"true only when a result said the profile file cannot be read, was saved by a version of MathTrail this one cannot read, or is in the Google Drive bin, and the adult asked for a new profile instead. The old file is set aside, not deleted, and a new profile starts from the pseudonym and grade given. A profile this version can read is never started over"`
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
		Country:        in.Country,
		Region:         in.Region,
		LessonTopic:    in.LessonTopic,
	}
}

func (s *Service) getProfileTool() Tool {
	return Define(Spec{
		Name:  "get_profile",
		Title: "Get the child's profile",
		Description: "Reads the child's profile — the pseudonym, the grade, the interests, the skills left out of " +
			"the tasks, the adult's notes, the language of the lessons and the country and state the adult may have " +
			"given — and what the next task would be. " +
			"Call it when the adult asks about the profile; a task needs only next_task. It draws no card: the " +
			"adult sees the profile, and changes it with a form, in the Profile section of the progress get_progress shows. " +
			"When there is no profile yet it says so, as next_task does, and how to set one up with save_profile. " +
			"Every result carries last_answer, the last answer the child gave, maybe on a card without you: read it " +
			"before you say anything about the current task.",
		ReadOnly:   true,
		Idempotent: true,
	}, s.getProfile)
}

func (s *Service) saveProfileTool() Tool {
	return Define(Spec{
		Name:  "save_profile",
		Title: "Save the child's profile",
		Description: "Creates the child's profile, or changes it. Pass only what changes; a field left out stays " +
			"as it is. To create the profile, pseudonym and grade are required. The pseudonym is what the child is " +
			"called: never a real name, a birth date or a school. The grade only sets where the first tasks start; " +
			"changed later it moves no rating. No card is drawn: say in a sentence what was saved. When a field " +
			"breaks a rule, nothing is saved and the result names the field and the rule.\n\nSkills that can be " +
			"left out of the tasks, by id:\n" + s.skillList(),
		Idempotent: true,
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

// getProfile reads the child's profile, and says where the adult finds it for
// themselves.
func (s *Service) getProfile(ctx context.Context, account store.Account, _ noArguments) (Reply[profileOut], error) {
	return afresh(ctx, func() (Reply[profileOut], error) { return s.readProfile(ctx, account) })
}

// readProfile is one read of the profile, and of where it is kept.
func (s *Service) readProfile(ctx context.Context, account store.Account) (Reply[profileOut], error) {
	p, _, err := s.store.Load(ctx, account)
	switch {
	case errors.Is(err, store.ErrNotFound):
		return Reply[profileOut]{Text: firstRunText, Payload: profileOut{Screen: screenFirstRun}}, nil
	case err != nil:
		return Reply[profileOut]{}, fmt.Errorf("mcp: read the profile: %w", err)
	}
	location, where, err := s.whereKept(ctx, account)
	if err != nil {
		return Reply[profileOut]{}, err
	}
	reply, err := s.profileReply(p, "")
	if err != nil {
		return Reply[profileOut]{}, err
	}
	reply.Payload.Location = location
	reply.Text = joined(reply.Text, where)
	return reply, nil
}

// saveProfile creates the profile when there is none and changes it when
// there is, and starts it over when the adult asks for that in place of one
// this version cannot read. Nothing is written when the details break a rule, and
// nothing when they already say what the call asks for.
//
//nolint:gocritic // hugeParam: the frame hands every handler its arguments by value
func (s *Service) saveProfile(ctx context.Context, account store.Account, in saveProfileIn) (Reply[profileOut], error) {
	return afresh(ctx, func() (Reply[profileOut], error) { return s.writeProfile(ctx, account, &in) })
}

// writeProfile is one read of the profile and the write the arguments ask of
// what it finds.
func (s *Service) writeProfile(ctx context.Context, account store.Account, in *saveProfileIn) (Reply[profileOut], error) {
	edit := in.edit()
	p, revision, err := s.store.Load(ctx, account)
	switch {
	case errors.Is(err, store.ErrNotFound):
		return s.createProfile(ctx, account, &edit)
	case in.StartOver && (errors.Is(err, store.ErrCorrupted) || errors.Is(err, store.ErrInBin) || errors.Is(err, profile.ErrNewer)):
		return s.startOver(ctx, account, &edit)
	case err != nil:
		return Reply[profileOut]{}, fmt.Errorf("mcp: read the profile: %w", err)
	case in.StartOver:
		// What the adult was told no longer holds — the file was mended, or put
		// back from the bin — and a profile that reads is never replaced.
		return s.profileReply(p, "The profile can be read, so it was not started over, and nothing was saved.")
	}

	changed, problems, err := s.saveChange(ctx, account, p, revision, &edit)
	switch {
	case err != nil:
		return Reply[profileOut]{}, err
	case len(problems) > 0:
		return s.refusal(p, problems)
	case !changed:
		return s.profileReply(p, "Nothing changed: the profile already says so.")
	}
	return s.profileReply(p, joined("Saved.", sayWhatWasSaved))
}

// saveChange makes an edit to the profile read at revision, and writes it.
// It reports whether there was anything to write: there is not when the edit
// breaks a rule, and the problems say which, or when it asks for what the
// profile already says.
func (s *Service) saveChange(ctx context.Context, account store.Account, p *profile.Profile, revision store.Revision,
	edit *profile.Edit,
) (bool, []profile.Problem, error) {
	changed, problems := p.Change(edit, s.content, s.version, s.now())
	if !changed {
		return false, problems, nil
	}
	if _, err := s.store.Save(ctx, account, p, revision); err != nil {
		return false, nil, fmt.Errorf("mcp: save the profile: %w", err)
	}
	return true, nil, nil
}

// sayWhatWasSaved is what the model is told once the profile is written: no
// card shows the profile, so its words are all the adult is shown of it.
const sayWhatWasSaved = "No card shows the profile: tell the adult in a sentence what was saved."

// createProfile makes the first profile of an account. The grade decides
// where the child starts on the ladder, once, here.
func (s *Service) createProfile(ctx context.Context, account store.Account, edit *profile.Edit) (Reply[profileOut], error) {
	student, problems := profile.NewStudent(edit, s.content)
	if len(problems) > 0 {
		return s.refusal(nil, problems)
	}
	p := profile.New(student, s.version, s.now())
	if _, err := s.store.Create(ctx, account, p); err != nil {
		return Reply[profileOut]{}, fmt.Errorf("mcp: create the profile: %w", err)
	}
	return s.profileReply(p, joined("The profile is created.", sayWhatWasSaved))
}

// startOver makes a new profile in place of one nothing can read, one of a
// newer version this one cannot, or one in the bin, as the adult asked: the
// child starts again from the grade given, and the old file is set aside
// rather than deleted.
func (s *Service) startOver(ctx context.Context, account store.Account, edit *profile.Edit) (Reply[profileOut], error) {
	student, problems := profile.NewStudent(edit, s.content)
	if len(problems) > 0 {
		return s.refusal(nil, problems)
	}
	p := profile.New(student, s.version, s.now())
	if _, err := s.store.StartOver(ctx, account, p); err != nil {
		return Reply[profileOut]{}, fmt.Errorf("mcp: start the profile over: %w", err)
	}
	return s.profileReply(p, joined("A new profile is started. The old file was not deleted: it stays in Google "+
		"Drive, renamed as set aside.", sayWhatWasSaved))
}

// profileReply is a profile as the two tools of the profile hand it back: the
// details, with the parent's notes, which only the model is told.
func (s *Service) profileReply(p *profile.Profile, lead string) (Reply[profileOut], error) {
	next, err := progress.Recommend(p, s.content)
	if err != nil {
		return Reply[profileOut]{}, err
	}
	trial := progress.TrialOf(p)
	return Reply[profileOut]{
		Text: joined(lead, s.detailsText(&p.Student), notesText(p.Student.Notes), s.lessonTopicText(p), stillInText(p),
			s.topicStillText(p), trialLine(trial), s.lastAnswerText(p), s.nextText(&next)),
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
	var lines string
	reply.Payload.Problems, lines = problemsOf(problems)
	reply.Payload.Status, reply.Payload.Code = statusRejected, codeInvalidProfile
	reply.Text = "Nothing was saved. Fix these fields and call save_profile again: " + lines + "."
	return reply, nil
}

func detailsOf(s *profile.Student) *detailsOut {
	return &detailsOut{childOut: *childOf(s), Notes: s.Notes}
}

// detailsText is the child's details in words, a sentence for each: every
// detail a card may show, and so never the parent's notes.
func (s *Service) detailsText(student *profile.Student) string {
	language := "The lessons follow the chat's language."
	if chosen, ok := student.ChosenLanguage(); ok {
		language = "The lessons are in " + chosen + ": the tasks, the cards and your words."
	}
	skills := make([]string, 0, len(student.ExcludedSkills))
	for _, id := range student.ExcludedSkills {
		skills = append(skills, s.skillName(id))
	}
	return joined(
		fmt.Sprintf("The profile of %s, grade %d.", quoted(student.Pseudonym), student.Grade),
		"Interests: "+listed(quotedEach(student.Interests), ", ")+".",
		"Left out of the tasks: "+listed(skills, "; ")+".",
		language,
		placeText(student),
	)
}

// placeText says where the family lives, as far as the adult said: the country
// by its code, and a state of the United States by its name as well. A region
// the file pairs with a country it is not part of is given by its code alone.
func placeText(student *profile.Student) string {
	switch {
	case student.Country == "":
		return "No country is given."
	case student.Region == "":
		return "Country: " + student.Country + "."
	}
	if name := country.RegionName(student.Country, student.Region); name != "" {
		return fmt.Sprintf("Country: %s, state: %s (%s).", student.Country, student.Region, name)
	}
	return fmt.Sprintf("Country: %s, region: %s.", student.Country, student.Region)
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

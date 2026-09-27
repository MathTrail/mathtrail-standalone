package profile

import (
	"fmt"
	"slices"
	"strings"
	"unicode"
	"unicode/utf8"
)

// Student is what the parent set up, and the only part of the file a person
// writes directly. There is no name, no birth date and no school here: a
// pseudonym is all the service ever learns about who the child is.
type Student struct {
	// ExcludedSkills are catalog ids that must appear neither in the wording
	// of a task nor in a trap: what this child has not been taught yet.
	ExcludedSkills []string `json:"excluded_skills"`
	// Grade is the school year, 1 to 6. It decides where the child starts on
	// the ladder when the profile is made, and nothing after that: changed
	// later it is a label, and the tasks follow the ratings rather than the
	// grade. The model is still told it, as the child's age.
	Grade int `json:"grade"`
	// Interests are the settings a task is dressed in, rotated through.
	Interests []string `json:"interests"`
	// Notes is free-form context about the child, written by the parent and
	// read by the chat's model as tone and level. It never reaches the rule,
	// so nothing written here can change which task the child is set or what
	// counts as the right answer.
	Notes string `json:"notes"`
	// Pseudonym is what the child is called. It must never reach the text of
	// a task, which is a rule the instructions carry and no code enforces.
	Pseudonym string `json:"pseudonym"`
	// UILanguage overrides the interface language, or nil to follow the chat.
	UILanguage *string `json:"ui_language"`
}

// Problem is one field of the child's details that breaks a rule: the field,
// named as the file names it, and the rule. The rule never repeats what the
// field holds. The details are the parent's own words, and a refusal that
// carried them would carry them into every place a refusal goes.
type Problem struct {
	// Field is the field, as the file and the tool that writes it name it.
	Field string
	// Rule is what the field has to be.
	Rule string
}

// String is the problem as one line: the field, then its rule.
func (p Problem) String() string { return p.Field + ": " + p.Rule }

// problems are the fields of the details that break a rule of the file, in
// the order the file writes them. Every rule of the details is here and only
// here: a file that is read is held to them, and so is an edit before it is
// written.
func (s *Student) problems() []Problem {
	checked := []Problem{
		{Field: "excluded_skills", Rule: excludedSkillsRule(s.ExcludedSkills)},
		{Field: "grade", Rule: gradeRule(s.Grade)},
		{Field: "interests", Rule: interestsRule(s.Interests)},
		{Field: "notes", Rule: notesRule(s.Notes)},
		{Field: "pseudonym", Rule: pseudonymRule(s.Pseudonym)},
		{Field: "ui_language", Rule: languageRule(s.UILanguage)},
	}
	return slices.DeleteFunc(checked, func(p Problem) bool { return p.Rule == "" })
}

// Each rule below says what its field has to be, or nothing when the field is
// what it has to be. A rule may say how far off the field is — a count is not
// what the parent wrote — and never what it holds.

func excludedSkillsRule(skills []string) string {
	if len(skills) > MaxExcludedSkills {
		return fmt.Sprintf("may hold at most %d skills, not %d", MaxExcludedSkills, len(skills))
	}
	return ""
}

func gradeRule(grade int) string {
	if grade < MinGrade || grade > MaxGrade {
		return fmt.Sprintf("must be a school year from %d to %d", MinGrade, MaxGrade)
	}
	return ""
}

func interestsRule(interests []string) string {
	if len(interests) > MaxInterests {
		return fmt.Sprintf("may hold at most %d interests, not %d", MaxInterests, len(interests))
	}
	for _, interest := range interests {
		if count := utf8.RuneCountInString(interest); count == 0 || count > MaxInterest {
			return fmt.Sprintf("must each be 1 to %d characters; one is %d", MaxInterest, count)
		}
	}
	return ""
}

func notesRule(notes string) string {
	if count := utf8.RuneCountInString(notes); count > MaxNotes {
		return fmt.Sprintf("must be at most %d characters, not %d", MaxNotes, count)
	}
	return ""
}

func pseudonymRule(pseudonym string) string {
	switch count := utf8.RuneCountInString(pseudonym); {
	case count == 0:
		return "is required, and is what the child is called: never a real name"
	case count > MaxPseudonym:
		return fmt.Sprintf("must be at most %d characters, not %d", MaxPseudonym, count)
	case strings.ContainsFunc(pseudonym, unicode.IsControl):
		return "must not carry control characters"
	}
	return ""
}

func languageRule(language *string) string {
	switch {
	case language == nil:
		return ""
	case *language == "":
		return "is null to follow the chat's language, never an empty text"
	}
	if count := utf8.RuneCountInString(*language); count > MaxUILanguage {
		return fmt.Sprintf("must be at most %d characters, not %d", MaxUILanguage, count)
	}
	return ""
}

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
	// Country is the country the family lives in, when the parent chose to
	// say, by its code in the list of countries, such as US. It is there to
	// count how many families each country has without knowing who anybody
	// is, and nothing the child is set depends on it. The file is the parent's
	// to edit, so a line counts a code the list does not have as no country of
	// it.
	Country string `json:"country,omitempty"`
	// ExcludedSkills are catalog ids that must appear neither in the wording
	// of a task nor in a trap: what this child has not been taught yet.
	ExcludedSkills []string `json:"excluded_skills"`
	// Grade is the school year, 1 to 6. It decides where the child starts on
	// the ladder when the profile is made, and nothing after that: changed
	// later it is a label, and the tasks follow the ratings rather than the
	// grade. The model is still told it, as the child's age.
	Grade int `json:"grade"`
	// Interests are the settings one task in three is dressed in, taken in
	// turn; the model dresses the others in settings of its own.
	Interests []string `json:"interests"`
	// LessonTopic is the topic of the catalog the child or the adult chose to
	// keep the lessons to, by its id, or empty while the rule chooses. Once the
	// trial series is over, every task is on it until it is given back. The
	// file is the parent's to edit, so a topic the catalog does not have counts
	// as no choice where it is read.
	LessonTopic string `json:"lesson_topic,omitempty"`
	// Notes is free-form context about the child, written by the parent and
	// read by the chat's model as tone and level. It never reaches the rule,
	// so nothing written here can change which task the child is set or what
	// counts as the right answer.
	Notes string `json:"notes"`
	// Pseudonym is what the child is called. It must never reach the text of
	// a task, which is a rule the instructions carry and no code enforces.
	Pseudonym string `json:"pseudonym"`
	// Region is the part of the country the family lives in, when the parent
	// chose to say — for the United States its state, such as US-TX — by its
	// code in ISO 3166-2. Only a country the list names regions of has one.
	Region string `json:"region,omitempty"`
	// SignInCountryOff says the parent asked for the country their browser signs
	// in from to be left out of what is counted. Absent, it is counted, as the
	// privacy policy says.
	SignInCountryOff bool `json:"signin_country_off,omitempty"`
	// UILanguage is the language the parent chose for the lessons — the tasks,
	// the cards and the model's words — or nil to follow the chat. The file
	// keeps the name it was given when it set the cards' language alone.
	UILanguage *string `json:"ui_language"`
}

// ChosenLanguage is the language the parent chose for the lessons, read as any
// language tag is read, and false when they chose none. The file is the
// parent's to edit by hand and is held to the length of a tag alone, so a
// value that names no language counts as no choice: it would otherwise become
// the language a task is written in and checked by.
func (s *Student) ChosenLanguage() (string, bool) {
	if s.UILanguage == nil {
		return "", false
	}
	tag, broken := LanguageTag(*s.UILanguage)
	return tag, tag != "" && broken.Code == ""
}

// LessonLanguage is the language a lesson is held in: the one the parent chose,
// when they chose one, and otherwise the chat's. The task and the words of its
// card always share it: buttons in one language around a task in another
// leave a child reading two languages at once.
func (s *Student) LessonLanguage(chat string) string {
	if chosen, ok := s.ChosenLanguage(); ok {
		return chosen
	}
	return chat
}

// Problem is one field that breaks a rule — of the child's details, or of what
// the model asked a task to be: the field, named as the file and the tools name
// it, and the rule it broke. The rule never repeats what the field holds. The
// details are the parent's own words and a choice is the model's, and a
// refusal that carried either would carry it into every place a refusal goes.
type Problem struct {
	// Field is the field, as the file and the tool that writes it name it.
	Field string
	Broken
}

// Broken is a rule a value breaks: its code, and the rule in words. The code
// is for a card, which says the rule in its own language and from limits of
// its own; the words are for the model, which reads English. The zero Broken
// is no rule broken.
type Broken struct {
	// Code names the rule, one of the codes below.
	Code string
	// Rule is what the value has to be.
	Rule string
}

// The rules a value can break, by the code a refusal names each with. The
// list is closed, so that a card can say a code it is sent in words of its own
// language, and say one it has no words for in words that fit any; a rule
// that fits none of them is a new code here first.
const (
	// CodeRequired is a value that has to be given, and was not.
	CodeRequired = "required"
	// CodeTooLong is a text longer than it may be.
	CodeTooLong = "too_long"
	// CodeTooMany is a list with more entries than it may hold.
	CodeTooMany = "too_many"
	// CodeEntryLength is an entry of a list that is empty, or longer than an
	// entry may be.
	CodeEntryLength = "entry_length"
	// CodeOutOfRange is a number outside the range it has to be in.
	CodeOutOfRange = "out_of_range"
	// CodeNotOneOf is a value that is none of the few it may be.
	CodeNotOneOf = "not_one_of"
	// CodeNotInCatalog is an id the catalog does not have.
	CodeNotInCatalog = "not_in_catalog"
	// CodeNotTaught is a level its topic is not taught at.
	CodeNotTaught = "not_taught"
	// CodeNotALanguage is a text that is no BCP 47 tag of a language.
	CodeNotALanguage = "not_a_language"
	// CodeEmptyText is an empty text where null is what says "none".
	CodeEmptyText = "empty_text"
	// CodeControlCharacters is a text with characters that control rather
	// than show.
	CodeControlCharacters = "control_characters"
)

// Codes are every code a rule can be broken with.
func Codes() []string {
	return []string{
		CodeRequired, CodeTooLong, CodeTooMany, CodeEntryLength, CodeOutOfRange, CodeNotOneOf,
		CodeNotInCatalog, CodeNotTaught, CodeNotALanguage, CodeEmptyText, CodeControlCharacters,
	}
}

// String is the problem as one line: the field, then its rule.
func (p Problem) String() string { return p.Field + ": " + p.Rule }

// problems are the fields of the details that break a rule of the file, in
// the order the file writes them. Every rule of the details is here and only
// here: a file that is read is held to them, and so is an edit before it is
// written.
func (s *Student) problems() []Problem {
	checked := []Problem{
		{Field: "country", Broken: lengthRule(s.Country, MaxCountry)},
		{Field: "excluded_skills", Broken: excludedSkillsRule(s.ExcludedSkills)},
		{Field: "grade", Broken: gradeRule(s.Grade)},
		{Field: "interests", Broken: interestsRule(s.Interests)},
		{Field: "lesson_topic", Broken: lengthRule(s.LessonTopic, MaxLessonTopic)},
		{Field: "notes", Broken: notesRule(s.Notes)},
		{Field: "pseudonym", Broken: pseudonymRule(s.Pseudonym)},
		{Field: "region", Broken: lengthRule(s.Region, MaxRegion)},
		{Field: "ui_language", Broken: languageRule(s.UILanguage)},
	}
	return slices.DeleteFunc(checked, func(p Problem) bool { return p.Code == "" })
}

// Each rule below says what its field has to be, or nothing when the field is
// what it has to be. A rule may say how far off the field is — a count is not
// what the parent wrote — and never what it holds.

func excludedSkillsRule(skills []string) Broken {
	if len(skills) > MaxExcludedSkills {
		return Broken{CodeTooMany, fmt.Sprintf("may hold at most %d skills, not %d", MaxExcludedSkills, len(skills))}
	}
	return Broken{}
}

func gradeRule(grade int) Broken {
	if grade < MinGrade || grade > MaxGrade {
		return Broken{CodeOutOfRange, fmt.Sprintf("must be a school year from %d to %d", MinGrade, MaxGrade)}
	}
	return Broken{}
}

func interestsRule(interests []string) Broken {
	if len(interests) > MaxInterests {
		return Broken{CodeTooMany, fmt.Sprintf("may hold at most %d interests, not %d", MaxInterests, len(interests))}
	}
	for _, interest := range interests {
		if count := utf8.RuneCountInString(interest); count == 0 || count > MaxInterest {
			return Broken{CodeEntryLength, fmt.Sprintf("must each be 1 to %d characters; one is %d", MaxInterest, count)}
		}
	}
	return Broken{}
}

func notesRule(notes string) Broken {
	if count := utf8.RuneCountInString(notes); count > MaxNotes {
		return Broken{CodeTooLong, fmt.Sprintf("must be at most %d characters, not %d", MaxNotes, count)}
	}
	return Broken{}
}

func pseudonymRule(pseudonym string) Broken {
	switch count := utf8.RuneCountInString(pseudonym); {
	case count == 0:
		return Broken{CodeRequired, "is required, and is what the child is called: never a real name"}
	case count > MaxPseudonym:
		return Broken{CodeTooLong, fmt.Sprintf("must be at most %d characters, not %d", MaxPseudonym, count)}
	case strings.ContainsFunc(pseudonym, unicode.IsControl):
		return Broken{CodeControlCharacters, "must not carry control characters"}
	}
	return Broken{}
}

func languageRule(language *string) Broken {
	switch {
	case language == nil:
		return Broken{}
	case *language == "":
		return Broken{CodeEmptyText, "is null to follow the chat's language, never an empty text"}
	}
	if count := utf8.RuneCountInString(*language); count > MaxLanguageTag {
		return Broken{CodeTooLong, fmt.Sprintf("must be at most %d characters, not %d", MaxLanguageTag, count)}
	}
	return Broken{}
}

// lengthRule holds a code to the length of the longest a code can be. The file
// is held to nothing more of a country, a region or the topic of the lessons:
// which codes there are is a rule of the edit, and a code a person typed into
// the file by hand counts as none where it is read.
func lengthRule(code string, most int) Broken {
	if count := utf8.RuneCountInString(code); count > most {
		return Broken{CodeTooLong, fmt.Sprintf("must be at most %d characters, not %d", most, count)}
	}
	return Broken{}
}

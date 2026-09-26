package profile

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

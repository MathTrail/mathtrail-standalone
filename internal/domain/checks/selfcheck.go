package checks

import (
	"fmt"
	"slices"
)

// blockingIssues reports every issue the self-check found and marked
// blocking. An issue is pointed at by its place and its type, never by its
// comment: the model wrote the comment and still has it, and a comment can
// name an option, which a refusal never does.
func (r *reviewer) blockingIssues(check *SelfCheck) (problems []Problem, unchecked string) {
	if check == nil {
		return nil, "the self-check was not checked: that needs self_check"
	}
	for i, issue := range check.Issues {
		if issue.Severity != severityBlocking {
			continue
		}
		problems = append(problems, Problem{Code: CodeSelfCheckBlocking, Message: fmt.Sprintf(
			"self_check.issues.%d%s is blocking; resolve what it describes before handing the task in",
			i, typeNamed(issue))})
	}
	return problems, ""
}

// typeNamed names an issue's type for a refusal when it is a type the format
// has. Anything else is the model's own words, which a refusal does not
// repeat; the structure check reports it.
func typeNamed(issue Issue) string {
	if !slices.Contains(issueTypes, issue.Type) {
		return ""
	}
	return " (" + issue.Type + ")"
}

// minorIssues are the types of the issues the self-check found and let
// through, one for each issue, so that two issues of one type count as two.
// What they said is dropped: the service keeps no text of a task, and a type
// is what they are counted by.
func minorIssues(check *SelfCheck) []string {
	if check == nil {
		return nil
	}
	var types []string
	for _, issue := range check.Issues {
		if issue.Severity == severityMinor && slices.Contains(issueTypes, issue.Type) {
			types = append(types, issue.Type)
		}
	}
	return types
}

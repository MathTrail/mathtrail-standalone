package checks

import (
	"fmt"
	"maps"
	"slices"
	"strings"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/profile"
)

// Catalog is what the structure check has to know about the catalogs the
// service ships. It is declared here, by the side that needs it, and it speaks
// in identifiers.
type Catalog interface {
	// HasTopic says whether the topic catalog has this id.
	HasTopic(id string) bool
	// HasTrap says whether the trap catalog has this id.
	HasTrap(id string) bool
	// HasSkill says whether the skill catalog has this id.
	HasSkill(id string) bool
}

// Structure checks a draft against the format and against the brief the open
// request was opened with: the task is the one that was asked for, the ids it
// names exist, and nothing the child's profile excludes was quietly dropped.
//
// A part the draft could not read is not checked; Decode has already said why.
func Structure(draft Draft, asked *profile.Brief, catalog Catalog) []Problem {
	var problems []Problem
	if draft.Brief != nil {
		problems = append(problems, checkBrief(draft.Brief, asked, catalog)...)
	}
	if draft.Task != nil {
		problems = append(problems, checkTask(draft.Task, catalog)...)
	}
	if draft.SelfCheck != nil {
		problems = append(problems, checkSelfCheck(draft.SelfCheck)...)
	}
	return problems
}

// checkBrief checks the brief handed back. The topic and the difficulty are
// the ones the request recorded: a model that wants others asks for them when
// it asks for the task, and the request then records its choice.
func checkBrief(brief, asked *profile.Brief, catalog Catalog) []Problem {
	var problems []Problem

	switch brief.PedagogicalGoal {
	case profile.GoalReinforce, profile.GoalNewTopic:
	default:
		problems = append(problems, structural("brief.pedagogical_goal must be %q or %q",
			profile.GoalReinforce, profile.GoalNewTopic))
	}

	switch {
	case brief.TargetConcept == "":
		problems = append(problems, structural("brief.target_concept is missing or empty"))
	case !catalog.HasTopic(brief.TargetConcept):
		problems = append(problems, structural("brief.target_concept %q is not a topic in the catalog", brief.TargetConcept))
	case brief.TargetConcept != asked.TargetConcept:
		problems = append(problems, structural("brief.target_concept is not the topic this task was asked for; "+
			"a different topic is chosen when asking for the task, not when handing it in"))
	}

	switch {
	case brief.Difficulty < 1 || brief.Difficulty > 5:
		problems = append(problems, structural("brief.difficulty must be a whole number from 1 to 5"))
	case brief.Difficulty != asked.Difficulty:
		problems = append(problems, structural("brief.difficulty is not the difficulty this task was asked for; "+
			"a different difficulty is chosen when asking for the task, not when handing it in"))
	}

	problems = append(problems, required("brief.setting", brief.Setting)...)
	problems = append(problems, required("brief.rationale", brief.Rationale)...)

	switch {
	case brief.TrapsToUse == nil:
		problems = append(problems, structural("brief.traps_to_use is missing"))
	case len(brief.TrapsToUse) == 0:
		problems = append(problems, structural("brief.traps_to_use must name at least one trap"))
	}
	problems = append(problems, catalogIDs("brief.traps_to_use", "trap", brief.TrapsToUse, catalog.HasTrap)...)

	if brief.ExcludedSkills == nil {
		problems = append(problems, structural("brief.excluded_skills is missing; an empty list says there are none"))
	}
	problems = append(problems, catalogIDs("brief.excluded_skills", "skill", brief.ExcludedSkills, catalog.HasSkill)...)
	for _, skill := range asked.ExcludedSkills {
		if !slices.Contains(brief.ExcludedSkills, skill) {
			problems = append(problems, structural("brief.excluded_skills dropped %q, which the child's profile excludes; "+
				"skills may be added to the list, never removed", skill))
		}
	}

	if brief.Constraints == nil {
		problems = append(problems, structural("brief.constraints is missing; an empty list says there are none"))
	}
	for i, constraint := range brief.Constraints {
		if strings.TrimSpace(constraint) == "" {
			problems = append(problems, structural("brief.constraints.%d is empty", i))
		}
	}
	return problems
}

// catalogIDs checks a list of ids against one catalog: each is known, and none
// is named twice.
func catalogIDs(field, kind string, ids []string, known func(string) bool) []Problem {
	var problems []Problem
	seen := make(map[string]bool, len(ids))
	for _, id := range ids {
		switch {
		case seen[id]:
			problems = append(problems, structural("%s names %q twice", field, id))
		case !known(id):
			problems = append(problems, structural("%s names %q, which is not a %s in the catalog", field, id, kind))
		}
		seen[id] = true
	}
	return problems
}

// checkTask checks the task itself: every text the child or the model relies
// on is there, the child is offered five answers they can tell apart, and
// every wrong one, and only a wrong one, carries the mistake it comes from.
func checkTask(task *Task, catalog Catalog) []Problem {
	var problems []Problem
	problems = append(problems, required("task.core_idea", task.CoreIdea)...)
	problems = append(problems, required("task.design_thought_process", task.DesignThoughtProcess)...)
	problems = append(problems, required("task.question", task.Question)...)
	problems = append(problems, required("task.hint", task.Hint)...)
	problems = append(problems, required("task.solution", task.Solution)...)
	problems = append(problems, checkOptions(task)...)
	problems = append(problems, checkDistractors(task, catalog)...)
	problems = append(problems, checkDrawing(task)...)
	return problems
}

// checkOptions checks that the child is offered five answers they can tell
// apart, with exactly one of them marked correct. Two options that differ only
// in spaces or letter case read the same to a child.
func checkOptions(task *Task) []Problem {
	var problems []Problem
	if !sameKeys(task.Options, optionKeys) {
		problems = append(problems, structural("task.options must hold exactly five options, "+
			"under the five keys of the format"))
	}

	texts := make(map[string]bool, len(task.Options))
	var empty, repeated int
	for _, text := range task.Options {
		folded := strings.ToLower(strings.TrimSpace(text))
		switch {
		case folded == "":
			empty++
		case texts[folded]:
			repeated++
		}
		texts[folded] = true
	}
	if empty > 0 {
		problems = append(problems, structural("task.options has %s", several(empty, "an empty option", "empty options")))
	}
	if repeated > 0 {
		problems = append(problems, structural("task.options has %s once spaces and letter case are ignored",
			several(repeated, "an option that reads the same as another", "options that read the same as others")))
	}

	if !slices.Contains(optionKeys, task.CorrectAnswer) {
		problems = append(problems, structural("task.correct_answer must be one of the five option keys"))
	}
	return problems
}

// checkDistractors checks that every wrong option, and only a wrong option, is
// explained, and that each explanation names a mistake the catalog knows. This
// is what turns a wrong answer into a diagnosis.
func checkDistractors(task *Task, catalog Catalog) []Problem {
	var problems []Problem
	switch {
	case task.Distractors == nil:
		problems = append(problems, structural("task.distractors is missing"))
	case !slices.Contains(optionKeys, task.CorrectAnswer):
		// Which options are the wrong ones is not known; the correct answer's
		// own refusal comes first.
	case !sameKeys(task.Distractors, wrongKeys(task.CorrectAnswer)):
		problems = append(problems, structural("task.distractors must explain exactly the four wrong options, "+
			"each once, and not the correct one"))
	}

	// The entries are keyed by an option's letter and cannot be pointed at, so
	// what is wrong with them is counted; a trap nobody has is named once,
	// however many explanations name it.
	var untold, unnamed int
	unknown := map[string]bool{}
	for _, distractor := range task.Distractors {
		if strings.TrimSpace(distractor.Text) == "" {
			untold++
		}
		switch {
		case distractor.Trap == "":
			unnamed++
		case !catalog.HasTrap(distractor.Trap):
			unknown[distractor.Trap] = true
		}
	}
	for _, trap := range slices.Sorted(maps.Keys(unknown)) {
		problems = append(problems, structural("task.distractors names trap %q, which is not in the catalog", trap))
	}
	if untold > 0 {
		problems = append(problems, structural("task.distractors has %s",
			several(untold, "an explanation with no text", "explanations with no text")))
	}
	if unnamed > 0 {
		problems = append(problems, structural("task.distractors has %s",
			several(unnamed, "an explanation that names no trap", "explanations that name no trap")))
	}
	return problems
}

// wrongKeys are the keys of the four options that are not the correct one.
func wrongKeys(correct string) []string {
	return slices.DeleteFunc(slices.Clone(optionKeys), func(key string) bool { return key == correct })
}

// checkDrawing checks that a picture and its description arrive together, and
// that the description is well formed. Whether the two agree with the wording
// is a later check's; here it only has to be readable as data.
func checkDrawing(task *Task) []Problem {
	hasDrawing, hasStructure := strings.TrimSpace(task.Drawing) != "", task.DrawingStructure != nil
	switch {
	case hasDrawing && !hasStructure:
		return []Problem{structural("task.drawing comes with task.drawing_structure, and it is missing")}
	case hasStructure && !hasDrawing:
		return []Problem{structural("task.drawing_structure describes task.drawing, and it is missing")}
	case !hasStructure:
		return nil
	}

	var problems []Problem
	structure := task.DrawingStructure
	problems = append(problems, required("task.drawing_structure.kind", structure.Kind)...)
	if len(structure.Objects) == 0 {
		problems = append(problems, structural("task.drawing_structure.objects must name at least one object"))
	}
	for i, object := range structure.Objects {
		if strings.TrimSpace(object.ID) == "" {
			problems = append(problems, structural("task.drawing_structure.objects.%d has no id", i))
		}
		if strings.TrimSpace(object.Label) == "" {
			problems = append(problems, structural("task.drawing_structure.objects.%d has no label to look for in the drawing", i))
		}
	}
	for i, relation := range structure.Relations {
		if strings.TrimSpace(relation.Type) == "" || strings.TrimSpace(relation.From) == "" ||
			strings.TrimSpace(relation.To) == "" {
			problems = append(problems, structural("task.drawing_structure.relations.%d needs a type, a from and a to", i))
		}
	}
	return problems
}

// checkSelfCheck checks that the model's own pass over its task is complete:
// every issue is one the format names, every option has its reason, and the
// answer it arrived at is one it could have arrived at.
func checkSelfCheck(check *SelfCheck) []Problem {
	var problems []Problem
	if check.Issues == nil {
		problems = append(problems, structural("self_check.issues is missing; an empty list says there are none"))
	}
	for i, issue := range check.Issues {
		if !slices.Contains(issueTypes, issue.Type) {
			problems = append(problems, structural("self_check.issues.%d has type %q, and the format's types are %s",
				i, issue.Type, strings.Join(issueTypes, ", ")))
		}
		if !slices.Contains(severities, issue.Severity) {
			problems = append(problems, structural("self_check.issues.%d has severity %q, and the format's are %s",
				i, issue.Severity, strings.Join(severities, " and ")))
		}
		if strings.TrimSpace(issue.Comment) == "" {
			problems = append(problems, structural("self_check.issues.%d has no comment", i))
		}
	}

	if !sameKeys(check.OptionCheck, optionKeys) {
		problems = append(problems, structural("self_check.option_check must say why each of the five options "+
			"is right or wrong, under the five keys of the format"))
	}
	empty := 0
	for _, reason := range check.OptionCheck {
		if strings.TrimSpace(reason) == "" {
			empty++
		}
	}
	if empty > 0 {
		problems = append(problems, structural("self_check.option_check has %s", several(empty, "an empty entry", "empty entries")))
	}

	if !slices.Contains(optionKeys, check.FinalAnswer) && check.FinalAnswer != unsolvable {
		problems = append(problems, structural("self_check.final_answer must be one of the five option keys, or %s",
			unsolvable))
	}
	return problems
}

// required reports a text the format requires and the draft left empty.
func required(field, text string) []Problem {
	if strings.TrimSpace(text) == "" {
		return []Problem{structural("%s is missing or empty", field)}
	}
	return nil
}

// sameKeys says whether a map is keyed by exactly these keys.
func sameKeys[V any](entries map[string]V, keys []string) bool {
	if len(entries) != len(keys) {
		return false
	}
	for _, key := range keys {
		if _, present := entries[key]; !present {
			return false
		}
	}
	return true
}

// several says how many of something are wrong, in words that stay right for
// one. An entry keyed by an option's letter cannot be pointed at without the
// letter, so the refusal counts rather than names.
func several(count int, one, many string) string {
	if count == 1 {
		return one
	}
	return fmt.Sprintf("%d %s", count, many)
}

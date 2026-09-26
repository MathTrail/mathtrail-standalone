package checks

// Threshold is the similarity at or above which a question counts as a
// near-duplicate; changing it changes what the product calls a new task.
//
// It is the highest that still catches the same task given new numbers in an
// inflected language, where that change measures 0.74; the prototype's 0.6
// also caught the same task in a new setting, and refused a third of the
// reference tasks as copies of another of their own level on the way. Text
// written without spaces is held to the same number for want of a corpus to
// measure its own against.
const Threshold = 0.7

// NearDuplicate checks that a question is new: that it copies none of the
// reference tasks, and repeats none of the tasks the child has been given.
//
// The reference tasks are compared by their texts, which the binary carries,
// and the child's past tasks by the fingerprints the profile keeps instead of
// theirs. Each is reported once however many it resembles: the model is told
// what to change, not how many times it failed to. The question is cut into
// the shingles the language of the task calls for, as its readability is
// counted, and what it is compared with is cut the same way.
func NearDuplicate(question, language string, references, fingerprints []string) []Problem {
	unit := unitFor(question, language)
	set := shinglesOf(question, unit)

	var problems []Problem
	for _, reference := range references {
		if set.jaccard(shinglesOf(reference, unit)) >= Threshold {
			problems = append(problems, Problem{Code: CodeNearDuplicate, Message: "task.question is a near-copy " +
				"of one of the reference tasks; write a task of your own rather than a variant of an example"})
			break
		}
	}

	sketch := sketchOf(set)
	for _, stored := range fingerprints {
		past, readable := readFingerprint(stored)
		if readable && resemblance(&sketch, &past) >= Threshold {
			problems = append(problems, Problem{Code: CodeNearDuplicate, Message: "task.question repeats a task " +
				"this child has already been given; change the idea and the numbers, not only the wording"})
			break
		}
	}
	return problems
}

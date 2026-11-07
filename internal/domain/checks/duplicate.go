package checks

import (
	"encoding/json"
	"slices"
)

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

// PictureThreshold is the similarity at or above which the picture of a task
// solved from the picture repeats one the child was given. Such a task keeps
// its words and changes its picture, so its picture is what makes it new: of
// the study's pairs of one genre, the elements of their pictures at 0.9
// caught every pair that shares a picture and refused one of 139 that do not.
const PictureThreshold = 0.9

// Earlier is what a task was found to repeat, in the word the log names it by:
// a kind of earlier task, and, of a reference task, whether it was the
// question in English that copied it.
type Earlier string

// The kinds of earlier task.
const (
	// EarlierReference is a reference task of the level the task was asked at.
	EarlierReference Earlier = "reference"
	// EarlierPast is a task the child has been given.
	EarlierPast Earlier = "past"
	// EarlierReferenceInEnglish is a reference task of the level, which the
	// question in English of a task of a lesson in another language copies:
	// the task tells it again in the lesson's language.
	EarlierReferenceInEnglish Earlier = "reference_en"
	// EarlierPastPicture is a task solved from the picture the child has been
	// given, whose picture the task's repeats, whatever its question asks.
	EarlierPastPicture Earlier = "past_picture"
)

// repeatedMessages are what the model is told to change when its question
// repeats an earlier task, by the kind of that task.
var repeatedMessages = map[Earlier]string{
	EarlierReference: "task.question is a near-copy of one of the reference tasks; write a task of your own rather " +
		"than a variant of an example",
	EarlierPast: "task.question repeats a task this child has already been given; change the idea and the " +
		"numbers, not only the wording",
	EarlierReferenceInEnglish: "task.question_en is a near-copy of one of the reference tasks, so the task tells it " +
		"again in the lesson's language; write a task of your own rather than a translation of an example, and " +
		"keep task.question_en its question in English, with the same names, numbers and conditions and the names " +
		"in Latin letters",
	EarlierPastPicture: "task.picture repeats the picture of a task this child has already been given; draw " +
		"another figure, not the same one with another question",
}

// NearDuplicate checks that a question is new: that it copies none of the
// reference tasks, and repeats none of the tasks the child has been given.
//
// The reference tasks are compared by their texts, which the binary carries,
// and the child's past tasks by the fingerprints the profile keeps instead of
// theirs. Each kind is reported once however many of it the question
// resembles: the model is told what to change, not how many times it failed
// to. The question is cut into the shingles the language of the task calls
// for, as its readability is counted, and what it is compared with is cut the
// same way.
func NearDuplicate(question, language string, references, fingerprints []string) []Problem {
	unit := unitFor(question, language)
	set := shinglesOf(question, unit)

	var problems []Problem
	if copiesOneOf(set, unit, references) {
		problems = append(problems, repeats(EarlierReference))
	}

	sketch := sketchOf(set)
	for _, stored := range fingerprints {
		past, readable := readFingerprint(stored)
		if readable && resemblance(&sketch, &past) >= Threshold {
			problems = append(problems, repeats(EarlierPast))
			break
		}
	}
	return problems
}

// RepeatsPicture checks that the picture of a task solved from the picture is
// new: that it repeats the picture of none of the tasks of the kind the child
// has been given, by the sketches the profile keeps of their pictures.
// Another question asked of the same picture is a repeat, since the child
// works on the same figure.
func RepeatsPicture(described json.RawMessage, fingerprints []string) []Problem {
	elements := pictureElements(described)
	if len(elements) == 0 {
		return nil
	}
	sketch := sketchOf(elements)
	for _, stored := range fingerprints {
		past, readable := readFingerprint(stored)
		if readable && resemblance(&sketch, &past) >= PictureThreshold {
			return []Problem{repeats(EarlierPastPicture)}
		}
	}
	return nil
}

// NearCopyInEnglish checks that the question in English of a task of a lesson
// in another language copies none of the reference tasks. They are in English,
// and a copy of one told in the lesson's language shares next to nothing with
// them; its question in English is cut and compared as the question of a lesson
// in English is, at the same threshold, and is reported once however many of
// them it resembles. The child's past tasks are not compared with it: they are
// known by the fingerprints of their questions, cut in their lessons' languages.
func NearCopyInEnglish(questionInEnglish string, references []string) []Problem {
	unit := unitFor(questionInEnglish, englishTag)
	if copiesOneOf(shinglesOf(questionInEnglish, unit), unit, references) {
		return []Problem{repeats(EarlierReferenceInEnglish)}
	}
	return nil
}

// copiesOneOf says whether a question, cut into its shingles, is a near-copy of
// any of these reference tasks, each cut into the same unit.
func copiesOneOf(set shingles, unit string, references []string) bool {
	return slices.ContainsFunc(references, func(reference string) bool {
		return set.jaccard(shinglesOf(reference, unit)) >= Threshold
	})
}

// repeats is the refusal of a question that repeats a kind of earlier task.
func repeats(earlier Earlier) Problem {
	return Problem{Code: CodeNearDuplicate, Message: repeatedMessages[earlier], NearDuplicateOf: earlier}
}

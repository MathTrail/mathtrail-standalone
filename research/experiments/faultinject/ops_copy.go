package main

import (
	"github.com/MathTrail/mathtrail-standalone/internal/domain/checks"
)

// The names and the countable nouns a copy in a new setting changes, each to
// the next one in its list: the names that recur in the reference questions,
// and the objects they count most often.
var (
	settingNames = []string{
		"Ben", "Ann", "Kim", "Tom", "Masha", "Sam", "Max", "Dan", "Eva", "Nina", "Leo", "Kate",
		"Kolya", "Mia", "Lena", "Vera", "Anya", "Anna", "Sasha", "Oleg", "Lily", "Fay", "Ivy",
		"Olga", "Tim", "Rosa", "Rita", "Olya", "Gleb", "Bob", "Zoe", "Uma",
	}
	settingNouns = []string{
		"coins", "balls", "socks", "sweets", "cards", "beads", "marbles", "apples", "books",
		"stamps", "pencils", "gloves", "stickers", "nuts", "shells", "flowers", "eggs", "pears",
		"lemons", "boxes",
	}
	settingSingulars = []string{
		"coin", "ball", "sock", "sweet", "card", "bead", "marble", "apple", "book",
		"stamp", "pencil", "glove", "sticker", "nut", "shell", "flower", "egg", "pear",
		"lemon", "box",
	}
)

// newSetting maps every name and noun of the lists to the next in its list.
var newSetting = func() map[string]string {
	all := cyclic(settingNames)
	for word, next := range cyclic(settingNouns) {
		all[word] = next
	}
	for word, next := range cyclic(settingSingulars) {
		all[word] = next
	}
	return all
}()

var newSettingWords = wordPattern(keysOf(newSetting))

// copyOperators are the copies and repeats: classes D14, D15 and X09.
func copyOperators() []operator {
	duplicate := []checks.Code{checks.CodeNearDuplicate}
	return []operator{
		{ID: "D14a", Class: "D14", Kind: mechanism, Expected: duplicate, Apply: copied(verbatim)},
		{ID: "D14b", Class: "D14", Kind: discovery, Expected: duplicate, Apply: copied(bumpNumbers)},
		{ID: "D14c", Class: "D14", Kind: mechanism, Expected: duplicate, Apply: copied(firstTwoSwapped)},
		{ID: "D15a", Class: "D15", Kind: mechanism, Expected: duplicate, Apply: repeated(verbatim)},
		{ID: "D15b", Class: "D15", Kind: discovery, Expected: duplicate, Apply: repeated(bumpNumbers)},
		{ID: "X09a", Class: "X09", Kind: outOfScope, Apply: copied(inNewSetting)},
		{ID: "X09b", Class: "X09", Kind: outOfScope, Apply: copied(inNewSettingWithNumbers)},
	}
}

// copied puts a question made from another reference task's in place of the
// host's question, leaving the host's options, key and solver. Neither the host
// nor the donor has a drawing: a donor's wording would not name the host's
// labels, and the drawing check would refuse the case for that.
func copied(rewrite func(question string) (string, bool)) func(m *maker) (mutant, bool) {
	return func(m *maker) (mutant, bool) {
		if m.host.HasDrawing() {
			return mutant{}, false
		}
		donor, found := m.copyDonor()
		if !found {
			return mutant{}, false
		}
		question, changed := rewrite(donor.Base.Task.Question)
		if !changed {
			return mutant{}, false
		}
		sub := m.start()
		sub.Task.Question = question
		return mutant{Sub: sub, Source: donor.Base.Task.Question, Donor: donor.ID}, true
	}
}

// copyDonor is the task a copy is made from: another eligible host without a
// drawing, of the same level and topic, or of the same level when the topic has
// no other.
func (m *maker) copyDonor() (*host, bool) {
	sameLevel := func(other *host) bool { return other.Level == m.host.Level && !other.HasDrawing() }
	donors := m.pool.donors(m.host, func(other *host) bool { return sameLevel(other) && other.Topic == m.host.Topic })
	if len(donors) == 0 {
		donors = m.pool.donors(m.host, sameLevel)
	}
	if len(donors) == 0 {
		return nil, false
	}
	return choose(m, donors), true
}

// repeated has the child's history hold a sketch of the host's own question,
// rewritten, and leaves the task itself as it was.
func repeated(rewrite func(question string) (string, bool)) func(m *maker) (mutant, bool) {
	return func(m *maker) (mutant, bool) {
		earlier, changed := rewrite(m.host.Base.Task.Question)
		if !changed {
			return mutant{}, false
		}
		sub := m.start()
		sub.Fingerprints = []string{checks.Fingerprint(earlier, language)}
		return mutant{Sub: sub, Source: earlier}, true
	}
}

// verbatim is the question as it is.
func verbatim(question string) (string, bool) { return question, true }

// firstTwoSwapped swaps the first two sentences of a question of three or
// more.
func firstTwoSwapped(question string) (string, bool) {
	sentences := sentencesOf(question)
	if len(sentences) < 3 {
		return question, false
	}
	sentences[0], sentences[1] = sentences[1], sentences[0]
	return rejoin(sentences), true
}

// inNewSetting changes the names and the objects of a question.
func inNewSetting(question string) (string, bool) {
	return swapWords(question, newSettingWords, newSetting)
}

// inNewSettingWithNumbers changes the names, the objects and the numbers of a
// question; a question with none of either is left alone.
func inNewSettingWithNumbers(question string) (string, bool) {
	moved, renamed := inNewSetting(question)
	if !renamed {
		return question, false
	}
	return bumpNumbers(moved)
}

package checks_test

import (
	"encoding/base64"
	"testing"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/checks"
)

const (
	spaceships = "Four spaceships dock in pairs. How many different pairs can they make?"
	// sevenShips is the same task given new numbers, 0.846 alike in Postgres's
	// own measurement.
	sevenShips = "Seven spaceships dock in pairs. How many different pairs can they make?"
	// kittens is the same task in a new setting, 0.625 alike.
	kittens = "Four kittens play in pairs. How many different pairs can they make?"
	// ships and sevenShipsRussian are the same task in Russian given new
	// numbers, 0.738 alike: an inflected language changes more trigrams for the
	// same edit, and it is what the threshold was set under.
	ships             = "Четыре корабля стыкуются парами. Сколько разных пар получится?"
	sevenShipsRussian = "Семь кораблей стыкуются парами. Сколько разных пар получится?"
	clock             = "A clock strikes 3 times at three o'clock. How long does it strike at six o'clock?"
	robots            = "Five robots shake hands in pairs. How many handshakes are there?"

	// Three tasks written without spaces, measured in bigrams: pears against
	// peaches is 10/14 alike, pears against big peaches 10/15.
	pears      = "小明有三个苹果和两个梨。"
	peaches    = "小明有三个苹果和两个桃。"
	bigPeaches = "小明有三个苹果和两个大桃。"
)

func TestANewQuestionIsNoDuplicate(t *testing.T) {
	t.Parallel()

	references := []string{clock, spaceships}
	past := []string{checks.Fingerprint(clock), checks.Fingerprint(spaceships)}
	if problems := checks.NearDuplicate(robots, references, past); len(problems) != 0 {
		t.Fatalf("NearDuplicate() = %v, want none", problems)
	}
}

func TestACopyOfAReferenceTaskIsRefused(t *testing.T) {
	t.Parallel()

	problems := checks.NearDuplicate(sevenShips, []string{clock, spaceships}, nil)
	if len(problems) != 1 || problems[0].Code != checks.CodeNearDuplicate || !mentions(problems, "reference tasks") {
		t.Fatalf("NearDuplicate() = %v, want one near_duplicate about the reference tasks", problems)
	}
}

// The threshold was set to catch the same task given new numbers in an
// inflected language, which measures lower than it does in English.
func TestNewNumbersInRussianAreStillACopy(t *testing.T) {
	t.Parallel()

	if problems := checks.NearDuplicate(sevenShipsRussian, []string{ships}, nil); len(problems) != 1 {
		t.Fatalf("NearDuplicate() = %v, want the same task with new numbers refused", problems)
	}
}

// The child's past tasks are known only by their fingerprints, and a repeat is
// still recognised through them.
func TestARepeatOfAPastTaskIsRefused(t *testing.T) {
	t.Parallel()

	past := []string{checks.Fingerprint(clock), checks.Fingerprint(spaceships)}
	problems := checks.NearDuplicate(spaceships, nil, past)
	if len(problems) != 1 || problems[0].Code != checks.CodeNearDuplicate || !mentions(problems, "already been given") {
		t.Fatalf("NearDuplicate() = %v, want one near_duplicate about a past task", problems)
	}
}

// Each source is reported once however many of its entries the question
// resembles, and both are reported when both are resembled.
func TestEachSourceIsReportedOnce(t *testing.T) {
	t.Parallel()

	references := []string{spaceships, spaceships, kittens}
	past := []string{checks.Fingerprint(spaceships), checks.Fingerprint(spaceships)}
	if problems := checks.NearDuplicate(spaceships, references, past); len(problems) != 2 {
		t.Fatalf("NearDuplicate() = %v, want exactly one problem per source", problems)
	}
}

// The same task in a new setting is a new task: the threshold was set above
// it on purpose, because the threshold that caught it also refused a third of
// the reference tasks as copies of their neighbours.
func TestTheSameTaskInANewSettingIsNotACopy(t *testing.T) {
	t.Parallel()

	if problems := checks.NearDuplicate(kittens, []string{spaceships}, []string{checks.Fingerprint(spaceships)}); len(problems) != 0 {
		t.Fatalf("NearDuplicate() = %v, want none", problems)
	}
}

// A text written without spaces is compared in character bigrams: a near-copy
// is refused, and a task that shares most of its characters but not enough of
// them is not.
func TestAQuestionWithoutSpacesIsComparedInCharacters(t *testing.T) {
	t.Parallel()

	if problems := checks.NearDuplicate(peaches, []string{pears}, nil); len(problems) != 1 {
		t.Errorf("NearDuplicate() = %v, want the near-copy refused", problems)
	}
	if problems := checks.NearDuplicate(bigPeaches, []string{pears}, nil); len(problems) != 0 {
		t.Errorf("NearDuplicate() = %v, want none under the threshold", problems)
	}
}

// A fingerprint damaged in the file is a repeat that cannot be seen, not a
// reason to refuse a task and not a reason to stop.
func TestADamagedFingerprintIsSkipped(t *testing.T) {
	t.Parallel()

	damaged := []string{"", "not base64 at all!", base64.StdEncoding.EncodeToString([]byte("too short"))}
	if problems := checks.NearDuplicate(spaceships, nil, damaged); len(problems) != 0 {
		t.Fatalf("NearDuplicate() = %v, want none", problems)
	}
	problems := checks.NearDuplicate(spaceships, nil, append(damaged, checks.Fingerprint(spaceships)))
	if len(problems) != 1 {
		t.Fatalf("NearDuplicate() = %v, want the readable fingerprint still compared", problems)
	}
}

// The refusal reaches the widget like any other, so it names neither option
// nor answer — only what to change.
func TestADuplicateRefusalQuotesNoOption(t *testing.T) {
	t.Parallel()

	for _, problem := range checks.NearDuplicate(spaceships, []string{spaceships}, []string{checks.Fingerprint(spaceships)}) {
		if letter.MatchString(problem.Message) {
			t.Errorf("%q quotes an option's letter", problem.Message)
		}
	}
}

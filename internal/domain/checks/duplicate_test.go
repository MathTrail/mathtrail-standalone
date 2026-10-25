package checks_test

import (
	"encoding/base64"
	"slices"
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
	past := []string{checks.Fingerprint(clock, ""), checks.Fingerprint(spaceships, "")}
	if problems := checks.NearDuplicate(robots, "", references, past); len(problems) != 0 {
		t.Fatalf("NearDuplicate() = %v, want none", problems)
	}
}

func TestACopyOfAReferenceTaskIsRefused(t *testing.T) {
	t.Parallel()

	problems := checks.NearDuplicate(sevenShips, "", []string{clock, spaceships}, nil)
	if len(problems) != 1 || problems[0].Code != checks.CodeNearDuplicate || !mentions(problems, "reference tasks") ||
		problems[0].NearDuplicateOf != checks.EarlierReference {
		t.Fatalf("NearDuplicate() = %v, want one near_duplicate of a reference task", problems)
	}
}

// The threshold was set to catch the same task given new numbers in an
// inflected language, which measures lower than it does in English.
func TestNewNumbersInRussianAreStillACopy(t *testing.T) {
	t.Parallel()

	if problems := checks.NearDuplicate(sevenShipsRussian, "", []string{ships}, nil); len(problems) != 1 {
		t.Fatalf("NearDuplicate() = %v, want the same task with new numbers refused", problems)
	}
}

// The child's past tasks are known only by their fingerprints, and a repeat is
// still recognised through them.
func TestARepeatOfAPastTaskIsRefused(t *testing.T) {
	t.Parallel()

	past := []string{checks.Fingerprint(clock, ""), checks.Fingerprint(spaceships, "")}
	problems := checks.NearDuplicate(spaceships, "", nil, past)
	if len(problems) != 1 || problems[0].Code != checks.CodeNearDuplicate || !mentions(problems, "already been given") ||
		problems[0].NearDuplicateOf != checks.EarlierPast {
		t.Fatalf("NearDuplicate() = %v, want one near_duplicate of a past task", problems)
	}
}

// Each source is reported once however many of its entries the question
// resembles, and both are reported when both are resembled, each naming its
// own kind of earlier task: the reference tasks first, as they are compared.
func TestEachSourceIsReportedOnce(t *testing.T) {
	t.Parallel()

	references := []string{spaceships, spaceships, kittens}
	past := []string{checks.Fingerprint(spaceships, ""), checks.Fingerprint(spaceships, "")}
	problems := checks.NearDuplicate(spaceships, "", references, past)
	if got, want := kindsOf(problems), []checks.Earlier{checks.EarlierReference, checks.EarlierPast}; !slices.Equal(got, want) {
		t.Fatalf("NearDuplicate() = %v, want exactly one problem per source, of %v", problems, want)
	}
}

// The same task in a new setting is a new task: the threshold was set above
// it on purpose, because the threshold that caught it also refused a third of
// the reference tasks as copies of their neighbours.
func TestTheSameTaskInANewSettingIsNotACopy(t *testing.T) {
	t.Parallel()

	if problems := checks.NearDuplicate(kittens, "", []string{spaceships}, []string{checks.Fingerprint(spaceships, "")}); len(problems) != 0 {
		t.Fatalf("NearDuplicate() = %v, want none", problems)
	}
}

// A text written without spaces is compared in character bigrams: a near-copy
// is refused, and a task that shares most of its characters but not enough of
// them is not.
func TestAQuestionWithoutSpacesIsComparedInCharacters(t *testing.T) {
	t.Parallel()

	if problems := checks.NearDuplicate(peaches, "", []string{pears}, nil); len(problems) != 1 {
		t.Errorf("NearDuplicate() = %v, want the near-copy refused", problems)
	}
	if problems := checks.NearDuplicate(bigPeaches, "", []string{pears}, nil); len(problems) != 0 {
		t.Errorf("NearDuplicate() = %v, want none under the threshold", problems)
	}
}

// A fingerprint damaged in the file is a repeat that cannot be seen, not a
// reason to refuse a task and not a reason to stop.
func TestADamagedFingerprintIsSkipped(t *testing.T) {
	t.Parallel()

	damaged := []string{"", "not base64 at all!", base64.StdEncoding.EncodeToString([]byte("too short"))}
	if problems := checks.NearDuplicate(spaceships, "", nil, damaged); len(problems) != 0 {
		t.Fatalf("NearDuplicate() = %v, want none", problems)
	}
	problems := checks.NearDuplicate(spaceships, "", nil, append(damaged, checks.Fingerprint(spaceships, "")))
	if len(problems) != 1 {
		t.Fatalf("NearDuplicate() = %v, want the readable fingerprint still compared", problems)
	}
}

// The refusal reaches the widget like any other, so it names neither option
// nor answer — only what to change.
func TestADuplicateRefusalQuotesNoOption(t *testing.T) {
	t.Parallel()

	problems := checks.NearDuplicate(spaceships, "", []string{spaceships}, []string{checks.Fingerprint(spaceships, "")})
	for _, problem := range append(problems, checks.NearCopyInEnglish(spaceships, []string{spaceships})...) {
		if letter.MatchString(problem.Message) {
			t.Errorf("%q quotes an option's letter", problem.Message)
		}
	}
}

// A reference task told again in the lesson's language is a copy its question
// in that language cannot show: tri-34-d3-1 put into Russian word for word
// shares next to nothing with it. Its question in English, written from the
// Russian alone, is 0.80 alike and is caught as a copy of it; a task on the
// same idea that asks something else is 0.47 alike and is not. The translation
// and the two questions in English come from the set such copies were first
// measured on.
func TestACopyToldInAnotherLanguageIsCaughtByItsQuestionInEnglish(t *testing.T) {
	t.Parallel()

	reference := []string{checks.ReferenceQuestion(t, "tri-34-d3-1")}
	const (
		translated = "Фая задумала число. Она умножила его на 5, а потом отняла 7. Получилось 23. Какое число задумала Фая?"
		copied     = "Faya thought of a number. She multiplied it by 5 and then subtracted 7. She got 23. What number did " +
			"Faya think of?"
		another = "Katya thought of a number, added 8 to it, multiplied the result by 3 and got 60. What do you get if " +
			"the number she thought of is multiplied by 2?"
	)
	if problems := checks.NearDuplicate(translated, "ru", reference, nil); len(problems) != 0 {
		t.Errorf("NearDuplicate(the translation) = %v, want none: in Russian the copy cannot be seen", problems)
	}
	problems := checks.NearCopyInEnglish(copied, reference)
	if got, want := kindsOf(problems), []checks.Earlier{checks.EarlierReferenceInEnglish}; !slices.Equal(got, want) ||
		problems[0].Code != checks.CodeNearDuplicate || !mentions(problems, "rather than a translation of an example") {
		t.Errorf("NearCopyInEnglish(its question in English) = %v, want one near_duplicate of %v", problems, want)
	}
	if problems := checks.NearCopyInEnglish(another, reference); len(problems) != 0 {
		t.Errorf("NearCopyInEnglish(a task on the same idea) = %v, want none", problems)
	}
}

// The question in English is cut as English is, into the trigrams of its
// words, and held to the threshold a question in English is held to, whatever
// the lesson's language: this copy of clk-34-d3-1 with new numbers is 0.65
// alike word by word and passes, as the same task with new numbers and new
// words of its plot passes in English. Cut into the characters of a lesson in
// Chinese, it would be 0.73 alike and refused.
func TestTheQuestionInEnglishIsMeasuredAsEnglish(t *testing.T) {
	t.Parallel()

	reference := []string{checks.ReferenceQuestion(t, "clk-34-d3-1")}
	const retold = "A tower clock strikes twice at every half hour. At the start of every hour it strikes as many times " +
		"as the hour it shows, from 1 to 6. How many strikes will it make in 6 hours?"
	if problems := checks.NearDuplicate(retold, "zh", reference, nil); len(problems) != 1 {
		t.Fatalf("NearDuplicate(cut as Chinese) = %v, want the copy refused, so that the case shows what the cut decides",
			problems)
	}
	if problems := checks.NearCopyInEnglish(retold, reference); len(problems) != 0 {
		t.Errorf("NearCopyInEnglish() = %v, want none: word by word it is 0.65 alike", problems)
	}
}

// A Chinese question that names its children Tom and Mary has more Latin
// letters than Chinese ones, and cut into words it hides a repeat: the same
// task given new numbers is 0.63 alike word by word and 0.85 character by
// character. The language of the task decides how it is cut — for the past
// tasks the profile keeps and for the reference tasks alike.
func TestARepeatInChineseWithLatinNamesIsCaught(t *testing.T) {
	t.Parallel()

	first := "Tom和Mary有12个苹果，Tom比Mary多4个，Tom有几个？"
	again := "Tom和Mary有14个苹果，Tom比Mary多2个，Tom有几个？"
	if problems := checks.NearDuplicate(again, "zh", nil, []string{checks.Fingerprint(first, "zh")}); len(problems) != 1 {
		t.Errorf("NearDuplicate(against a past task) = %v, want the repeat refused", problems)
	}
	if problems := checks.NearDuplicate(again, "zh", []string{first}, nil); len(problems) != 1 {
		t.Errorf("NearDuplicate(against a reference task) = %v, want the copy refused", problems)
	}
}

// kindsOf are the kinds of earlier task some problems name, in their order.
func kindsOf(problems []checks.Problem) []checks.Earlier {
	kinds := make([]checks.Earlier, 0, len(problems))
	for _, problem := range problems {
		kinds = append(kinds, problem.NearDuplicateOf)
	}
	return kinds
}

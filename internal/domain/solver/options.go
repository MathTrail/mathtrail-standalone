package solver

import (
	"cmp"
	"strings"
	"unicode"
	"unicode/utf8"

	"golang.org/x/text/unicode/norm"
	"golang.org/x/text/width"
)

// letters label the options, in order. A string rather than an array of five,
// because Go has no constant array and a package-level variable of them would
// be one anybody in this package could write to; the labels of the options are
// nobody's to change.
const letters = "ABCDE"

// Count is how many options a task offers.
const Count = len(letters)

// Letter is the label of the option in that place, counting from zero. A slice
// of the constant rather than a conversion of the byte at it: the slice is a
// second view of bytes that already exist, where the conversion would copy one
// onto the heap every time a letter was named.
func Letter(place int) string { return letters[place : place+1] }

// Letters are the labels of all the options, in order, in a slice the caller
// may keep and change.
func Letters() []string { return strings.Split(letters, "") }

// Place is where a label stands, or -1 when it labels nothing. The length is
// read first: without it a longer string would be found at the place its first
// letter stands.
func Place(letter string) int {
	if len(letter) != 1 {
		return -1
	}
	return strings.IndexByte(letters, letter[0])
}

// Options are the option texts of one task, in the order of the labels.
type Options [Count]string

// relabel is how far the labels move between the two runs of a program: the
// text labelled A comes back labelled C, B comes back D, and so on around the
// five. Two is the smallest shift that leaves no option where it was and no
// pair of options swapped, so a program that repeats an old letter — or one
// that mirrors the list — cannot land on the right one by accident.
const relabel = 2

// Rotated is the same options under shifted labels: the second run of a
// program sees the same five texts, each under a different letter. A program
// that computes an answer and looks it up follows the text and returns the new
// label; a program that writes a letter out by hand returns the old one.
func (o *Options) Rotated() Options {
	var rotated Options
	for place, text := range o {
		rotated[(place+relabel)%Count] = text
	}
	return rotated
}

// Relabelled is where a letter of the first run stands in the second.
func Relabelled(letter string) string {
	place := Place(letter)
	if place < 0 {
		return ""
	}
	return Letter((place + relabel) % Count)
}

// Match is the letters whose option text means the same as a computed value.
//
// All of them, not the first: two options saying one thing is a fault of the
// task, and a verdict that quietly picked one of them would hide it. None of
// them is the most valuable answer this check gives — the computed answer is
// not among the options at all.
func (o *Options) Match(value string) []string {
	matched := make([]string, 0, Count)
	wanted := Key(value)
	for place, text := range o {
		if Key(text) == wanted {
			matched = append(matched, Letter(place))
		}
	}
	return matched
}

// Key is what an option text says, written one way for every text that says
// it: a number in its shortest form, so that six, six point zero, +6 and a
// padded six are all "6", and anything else folded as a card shows it. Two
// texts are one answer exactly when their keys are equal, which is what lets
// a list of options be searched for repeats in one pass. A number's key is a
// number and no other text's key is, so the two kinds never meet.
func Key(text string) string {
	folded := fold(text)
	if !decimal(folded) {
		return folded
	}
	return shortest(folded)
}

// shortest writes a decimal without what does not change its value: a plus
// sign, zeros before the whole part or after the fraction, a point with nothing
// after it, and the sign of a zero. The digits themselves stay as they are, so
// a number longer than a float can hold is still told apart from the next one.
func shortest(number string) string {
	negative := strings.HasPrefix(number, "-")
	whole, fraction, _ := strings.Cut(strings.TrimLeft(number, "+-"), ".")
	written := cmp.Or(strings.TrimLeft(whole, "0"), "0")
	if fraction = strings.TrimRight(fraction, "0"); fraction != "" {
		written += "." + fraction
	}
	if negative && written != "0" {
		written = "-" + written
	}
	return written
}

// unseen are the characters that take no room and change nothing a child reads:
// the soft hyphen, the combining grapheme joiner, the zero-width space, the
// word joiner, the invisible operators of mathematics, the zero-width no-break
// space, the three marks of direction that a text in Arabic or Hebrew puts
// around a number, and the fillers that draw nothing — Hangul's four, which
// are letters to Unicode, and the blank pattern of braille. Other invisible
// characters are kept, because they change what is seen: how the letters of
// Persian join, how an emoji is put together, and the embeddings, isolates and
// overrides of direction, which can show the words of a run in another order,
// "12 34" as "34 12".
const unseen = "\u00ad\u034f\u061c\u115f\u1160\u200b\u200e\u200f\u2060\u2061\u2062\u2063\u2064\u2800\u3164\ufeff\uffa0"

// fold is a text as a child reads it on a card: letter case aside, every run
// of spaces one, without what takes no room on the screen, and with an accent
// the same whichever way it was put together. What takes no room goes first,
// so that it cannot keep an accent apart from its letter, and the accents are
// put together again once the letters under them have lost their case.
// Folding a folded text changes nothing.
//
// A text of plain ASCII has nothing to put together or take away and no case
// but the plain one, so it is only lowered and spaced, which is what most
// options and most computed values are, many times in a run.
func fold(text string) string {
	if ascii(text) {
		return foldASCII(text)
	}
	return foldAny(text)
}

// foldASCII is fold for a text of plain ASCII.
func foldASCII(text string) string { return strings.Join(strings.Fields(strings.ToLower(text)), " ") }

// foldAny is fold for any text at all. A letter or a digit written wide, as
// Chinese and Japanese text writes them, is read at its ordinary width first:
// a wide six is the same six.
func foldAny(text string) string {
	shown := strings.Map(func(r rune) rune {
		if strings.ContainsRune(unseen, r) {
			return -1
		}
		return r
	}, width.Fold.String(text))
	lowered := strings.Map(caseless, norm.NFC.String(shown))
	return strings.Join(strings.Fields(norm.NFC.String(lowered)), " ")
}

// caseless is a letter as its case does not tell it apart: the small letter of
// its capital, so that a final sigma, a sigma and a capital sigma are one
// letter, and so are the three ways of writing a long, short or capital s.
// Going through the capital is what makes the small letter the same one
// whichever form the text began with. A letter whose small letter of its
// capital is another letter altogether — the dotless i of Turkish, whose
// capital is the capital of i — stays itself, so that letters are one exactly
// when strings.EqualFold says they are.
func caseless(r rune) rune {
	lowered := unicode.ToLower(unicode.ToUpper(r))
	for other := unicode.SimpleFold(r); other != r; other = unicode.SimpleFold(other) {
		if other == lowered {
			return lowered
		}
	}
	return r
}

// Blank says whether a text shows nothing on a card: spaces, and characters
// that take no room or only steer the ones around them.
func Blank(text string) bool { return !strings.ContainsFunc(text, shows) }

// shows says whether a character is seen on a card. Of the format characters
// only the signs that Arabic and a few other scripts write in front of a
// number are: the rest take no room or only steer the letters around them.
func shows(r rune) bool {
	switch {
	case unicode.IsSpace(r), strings.ContainsRune(unseen, r):
		return false
	case unicode.Is(unicode.Cf, r):
		return unicode.Is(unicode.Prepended_Concatenation_Mark, r)
	}
	return true
}

// decimal says whether a folded text is the only shape read as a number: a
// whole number or a plain decimal fraction, with a sign or none. Everything
// else stays text, which is what keeps "24/7" from being three and a bit, and
// it is deliberate that no unit is stripped — an option that says twelve
// centimetres is matched by passing those words.
func decimal(text string) bool {
	if text != "" && (text[0] == '+' || text[0] == '-') {
		text = text[1:]
	}
	whole, fraction, pointed := strings.Cut(text, ".")
	return digits(whole) && (!pointed || digits(fraction))
}

// digits says whether a text is one or more of the digits 0 to 9 and nothing
// else.
func digits(text string) bool {
	return text != "" && strings.Trim(text, "0123456789") == ""
}

// ascii says whether a text is plain ASCII throughout.
func ascii(text string) bool {
	for i := range len(text) {
		if text[i] >= utf8.RuneSelf {
			return false
		}
	}
	return true
}

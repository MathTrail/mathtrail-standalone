package checks

import (
	"regexp"
	"slices"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/picture"
	"github.com/MathTrail/mathtrail-standalone/internal/domain/solver"
)

// A picture must not show the right answer: the child sees it before
// answering. What it shows is read as the picture's own reading gives it —
// its labels, times, cells and notes, the numbers it draws for one thing, what
// its marks single out and the last of a run of names — and held to the right
// option as match compares two texts. An option that is one number with words
// beside it, a unit or a noun, is held by its number too, since a picture
// carries no units; a clock by where its hands stand; and a time a picture
// writes by the moment it names, 09:05 as 9:05. The wording is where a task
// gives what it gives: a value it holds itself, as a number standing apart, as
// a word of its own or as a time naming the same moment, is no answer shown.

// showsTheAnswer says whether a picture shows the right option where the
// wording does not give it.
func showsTheAnswer(read picture.Picture, question, right string, decimals picture.Decimals) bool {
	option := answerOf(right, decimals)
	given := wordingOf(question, decimals, read.Labels())
	for _, shown := range read.Shown() {
		if shown.Face != nil && showsOnTheFace(*shown.Face, option, given) {
			return true
		}
		if shown.Face == nil && showsInText(shown.Text, option, given, decimals) {
			return true
		}
	}
	return false
}

// showsOnTheFace says whether a clock's hands stand where the right option
// puts them: at the time it names, or, on the hour, at the hour it names.
func showsOnTheFace(face picture.Time, option answer, given *wording) bool {
	if option.face != nil && *option.face == face && !slices.Contains(given.faces, face) {
		return true
	}
	onTheHour := face.Minute == 0 && option.number != "" && option.number == strconv.Itoa(face.Hour)
	return onTheHour && !slices.Contains(given.numbers, option.number)
}

// showsInText says whether a value a picture writes is the right option, its
// number, or the moment it names.
func showsInText(text string, option answer, given *wording, decimals picture.Decimals) bool {
	shown := keyOf(text, decimals)
	if shown == option.key && !given.holds(option) {
		return true
	}
	if moment, isTime := picture.ReadTime(text); isTime && option.moment != nil && *option.moment == moment &&
		!slices.Contains(given.moments, moment) {
		return true
	}
	return option.number != "" && shown == option.number && !slices.Contains(given.numbers, option.number)
}

// answer is the right option in each form a picture could show it in.
type answer struct {
	// text is the option as it is written, and key the same as match compares
	// two texts.
	text, key string
	// number is the one number the option holds, as match compares it, and
	// empty for an option that holds none, or more than one.
	number string
	// moment is the one time the option names, and face where it stands on a
	// face of twelve hours; both nil for an option that names no time, or more
	// than one.
	moment, face *picture.Time
}

// answerOf is the right option in each form a picture could show it in.
func answerOf(right string, decimals picture.Decimals) answer {
	option := answer{text: strings.TrimSpace(right), key: keyOf(right, decimals)}
	times := timesIn(right)
	if numbers := numbersIn(right, decimals, false); len(numbers) == 1 && len(times) == 0 {
		option.number = numbers[0]
	}
	if len(times) == 1 {
		face := times[0].OnTheFace()
		option.moment, option.face = &times[0], &face
	}
	return option
}

// wording is what a question gives of itself: the numbers standing apart in
// it, the moments its times name, as written and on a face, its text as match
// folds it, and what reading the labels of a picture softly needs.
type wording struct {
	numbers  []string
	moments  []picture.Time
	faces    []picture.Time
	folded   string
	runes    []rune
	declared map[string]bool
	ranges   []span
}

// wordingOf is what a question gives of itself, beside a picture with these
// labels.
func wordingOf(question string, decimals picture.Decimals, labels []string) *wording {
	given := &wording{
		numbers:  numbersIn(question, decimals, true),
		folded:   solver.Key(question),
		runes:    []rune(question),
		declared: declaredOf(labels),
	}
	for _, read := range timesIn(question) {
		given.moments = append(given.moments, read)
		given.faces = append(given.faces, read.OnTheFace())
	}
	given.ranges = rangesOf(given.runes, given.declared)
	return given
}

// holds says whether the wording gives the right option's text itself: as a
// number standing apart where the option is a number, as a word of its own,
// and, where the option is a label of capitals, wherever the soft reading of a
// picture's labels finds it named — among capitals written together, in a
// cell or inside a range — since the wording names every capital label a
// picture shows.
func (w *wording) holds(option answer) bool {
	if option.number == option.key && slices.Contains(w.numbers, option.key) {
		return true
	}
	if asWord(w.folded, option.key) {
		return true
	}
	return capitalsOnly(option.text) && named(w.runes, option.text, w.declared, w.ranges)
}

// asWord says whether a folded text holds another as a word of its own: with
// no Latin letter and no digit joined to it on either side, as the soft
// reading of a picture's labels has it, so that a script written without
// spaces gives B2 in 答案是B2吗.
func asWord(text, word string) bool {
	if word == "" {
		return false
	}
	for from := 0; from < len(text); {
		at := strings.Index(text[from:], word)
		if at < 0 {
			return false
		}
		start, end := from+at, from+at+len(word)
		before, _ := utf8.DecodeLastRuneInString(text[:start])
		after, _ := utf8.DecodeRuneInString(text[end:])
		if !joinsLatin(before) && !joinsLatin(after) {
			return true
		}
		from = start + 1
	}
	return false
}

// keyOf is a text as match compares it, a number read with the lesson's
// decimal mark: 2,5 and 2.5 are one number where the lesson writes a comma.
func keyOf(text string, decimals picture.Decimals) string {
	trimmed := strings.TrimSpace(text)
	if number := numberPattern(decimals).FindString(trimmed); number == trimmed && number != "" {
		return solver.Key(pointed(number, decimals))
	}
	return solver.Key(text)
}

// numbersIn are the numbers a text writes, each as match compares it: a run of
// digits with its sign and decimals, and its thousands where it groups them,
// that no Latin letter or digit stands straight before, as one in the name of
// a cell does, and no digit straight after. A letter of a script written
// without spaces joins nothing: 有12个 writes 12. A minus straight after a
// digit or a Latin letter joins two things, as in 3-5, and the number after it
// is read without it. In an option, a number beside a colon is part of a
// time; in the wording it counts, since a time that gives a number gives it
// all the same.
func numbersIn(text string, decimals picture.Decimals, inWording bool) []string {
	var numbers []string
	for _, at := range numberPattern(decimals).FindAllStringIndex(text, -1) {
		start, end := at[0], at[1]
		before, _ := utf8.DecodeLastRuneInString(text[:start])
		if sign, size := utf8.DecodeRuneInString(text[start:]); (sign == '-' || sign == '−') && joinsLatin(before) {
			start, before = start+size, sign
		}
		after, _ := utf8.DecodeRuneInString(text[end:])
		switch {
		case joinsLatin(before), after != utf8.RuneError && unicode.IsDigit(after):
		case !inWording && (before == ':' || after == ':'):
		default:
			numbers = append(numbers, solver.Key(pointed(text[start:end], decimals)))
		}
	}
	return numbers
}

// The numbers a text writes, by the decimal mark of its language: a grouped
// number first, so that 1,000 is one number where a point marks the decimals.
var (
	pointNumbers = regexp.MustCompile(`[-−]?(?:\d{1,3}(?:,\d{3})+|\d+)(?:\.\d+)?`)
	commaNumbers = regexp.MustCompile(`[-−]?(?:\d{1,3}(?:[. \x{00A0}\x{202F}]\d{3})+|\d+)(?:,\d+)?`)
)

// numberPattern is how a number is written in a language with this decimal
// mark.
func numberPattern(decimals picture.Decimals) *regexp.Regexp {
	if decimals == picture.Comma {
		return commaNumbers
	}
	return pointNumbers
}

// pointed is a number written as match reads one: its sign a hyphen, no
// grouping of its thousands, and a point before its decimals.
func pointed(number string, decimals picture.Decimals) string {
	number = strings.Replace(number, "−", "-", 1)
	if decimals == picture.Comma {
		return strings.NewReplacer(".", "", " ", "", " ", "", " ", "", ",", ".").Replace(number)
	}
	return strings.ReplaceAll(number, ",", "")
}

// timeWritten is a time of day as a text writes one, H:MM or HH:MM.
var timeWritten = regexp.MustCompile(`\d{1,2}:\d{2}`)

// timesIn are the times of day a text writes, each with no digit and no colon
// joined to it, which would make it part of something longer. What stands
// beside a time is looked at, not taken, so that a time straight after
// another, as in 9:00-10:30, is read too.
func timesIn(text string) []picture.Time {
	var times []picture.Time
	for _, at := range timeWritten.FindAllStringIndex(text, -1) {
		before, _ := utf8.DecodeLastRuneInString(text[:at[0]])
		after, _ := utf8.DecodeRuneInString(text[at[1]:])
		if partOfMore(before) || partOfMore(after) {
			continue
		}
		if read, isTime := picture.ReadTime(text[at[0]:at[1]]); isTime {
			times = append(times, read)
		}
	}
	return times
}

// partOfMore says whether a character joined to a time makes it part of
// something longer: a digit, or another colon.
func partOfMore(r rune) bool { return r == ':' || (r >= '0' && r <= '9') }

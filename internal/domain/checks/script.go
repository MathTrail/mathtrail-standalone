package checks

import (
	"slices"
	"strings"
	"unicode"

	"golang.org/x/text/language"
)

// spaceless are the scripts written without spaces between words. A text in
// them is measured in characters rather than words: by the readability check
// for its sentences, by the explanation check for its length, and by the
// duplicate check for its bigrams.
var spaceless = []*unicode.RangeTable{
	unicode.Han, unicode.Hiragana, unicode.Katakana,
	unicode.Thai, unicode.Lao, unicode.Khmer, unicode.Myanmar, unicode.Tibetan,
}

// Spaceless says whether most of the letters of a text belong to the scripts
// written without spaces. The two kinds of script are counted against each
// other rather than script by script, so that Japanese — kanji, hiragana and
// katakana together — is one kind of text however its letters divide.
func Spaceless(text string) bool {
	var letters, without int
	for _, r := range text {
		if !unicode.IsLetter(r) {
			continue
		}
		letters++
		if unicode.In(r, spaceless...) {
			without++
		}
	}
	return 2*without > letters
}

// The units a length is counted in.
const (
	unitWords      = "words"
	unitCharacters = "characters"
)

// spacelessScripts are the scripts written without spaces between words, by the
// codes a language tag names them with.
var spacelessScripts = []string{"Hans", "Hant", "Hani", "Jpan", "Hira", "Kana", "Thai", "Laoo", "Khmr", "Mymr", "Tibt"}

// unitFor is the unit a text of a task is counted in. The language the task was
// asked for in decides it when there is one — by the script that language is
// written in — because the task is written in it: a Chinese question naming
// its points A, B and C, or its children Tom and Mary, is still read a
// character at a time, and an English one quoting a Chinese sign is still read
// a word at a time. Only a task that came with no language is judged by its
// letters.
func unitFor(text, tag string) string {
	if script, named := scriptOf(tag); named {
		if slices.Contains(spacelessScripts, script) {
			return unitCharacters
		}
		return unitWords
	}
	if Spaceless(text) {
		return unitCharacters
	}
	return unitWords
}

// scriptOf is the script a task's language is written in: the one its tag
// names, or the one its language is most likely written in — Chinese in Han
// whether the tag says zh, cmn, yue or lzh. A tag that is no tag, or names no
// language, names no script either.
func scriptOf(tag string) (string, bool) {
	parsed, err := language.Parse(tag)
	if err != nil {
		return "", false
	}
	if base, confidence := parsed.Base(); confidence == language.No || base.String() == "und" {
		return "", false
	}
	script, _ := parsed.Script()
	return script.String(), true
}

// primarySubtag is the language a tag names, lowercased: "zh" of "zh-Hant-TW",
// "en" of "EN_us", and nothing of an empty tag.
func primarySubtag(tag string) string {
	primary, _, _ := strings.Cut(strings.ToLower(strings.TrimSpace(tag)), "-")
	primary, _, _ = strings.Cut(primary, "_")
	return primary
}

// lengthIn is how long a text is in a unit. Words are counted as a child reads
// them; characters are the letters and digits. A mark rides on the letter it
// is written over — the vowels and tones of Thai, Lao, Khmer, Myanmar and
// Tibetan are marks — and counting it too would make a syllable three
// characters long where a child reads one. Punctuation is not counted either.
func lengthIn(text, unit string) int {
	count := 0
	if unit == unitCharacters {
		for _, r := range text {
			if unicode.IsLetter(r) || unicode.IsDigit(r) {
				count++
			}
		}
		return count
	}
	for _, token := range strings.Fields(text) {
		if strings.IndexFunc(token, readAloud) >= 0 {
			count++
		}
	}
	return count
}

// readAloud says whether a character is read out, which is what makes a token
// between spaces a word: "5-litre" is one word, and so are the "+", "=" and
// "-" of "2 + 3 - 1 = 4". Punctuation alone is not — the "?" French sets apart
// with a space, a dash, a guillemet — except the hyphen-minus, which standing
// alone in a task is the minus sign.
func readAloud(r rune) bool { return !unicode.IsPunct(r) || r == '-' }

// reading is a text as two texts are compared by what they say: lowercased,
// and everything that is neither a letter, a digit nor a mark reduced to a
// single space between words.
func reading(text string) string {
	return strings.Join(strings.FieldsFunc(strings.ToLower(text), boundary), " ")
}

// startsWith says whether a text begins with another, as they read: word for
// word where words are separated by spaces, character for character where they
// are not. "Four" does not start "Fourteen ships", and 小明 does start 小明有.
func startsWith(text, start string) bool {
	text, start = reading(text), reading(start)
	if start == "" || !strings.HasPrefix(text, start) {
		return false
	}
	return len(text) == len(start) || Spaceless(start) || text[len(start)] == ' '
}

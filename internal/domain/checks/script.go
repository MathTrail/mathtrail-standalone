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
	if script, confidence := scriptOf(tag); confidence != language.No {
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

// noWritingSystem are the codes a tag may carry in place of a script that name
// none in particular: a script not known, several in common use, one inherited
// from the letter before, none written at all, and notations of symbols and
// mathematics.
var noWritingSystem = []string{"Zzzz", "Zyyy", "Zinh", "Zxxx", "Zsym", "Zsye", "Zmth"}

// scriptOf is the script a task's language is written in, and how sure that
// is: the one its tag names, or the one its language — or failing that its
// place — is most likely written in: Chinese in Han whether the tag says zh,
// cmn, yue or lzh, and so is a language not determined in China. A tag names
// no script, with no confidence at all, when it is no tag, when it names
// nothing at all — "und", or private words alone — and when no script can be
// told from what it does name, or only one of the codes that stand in for
// none.
//
// Whether a tag names anything is read from what it says, before anything is
// inferred: asked for the language of "und", the tag library answers English,
// its likeliest guess, and a Chinese task tagged "und" would be read a word at
// a time.
func scriptOf(tag string) (string, language.Confidence) {
	parsed, err := language.Parse(tag)
	if err != nil {
		return "", language.No
	}
	base, named, region := parsed.Raw()
	if base.String() == "und" && named.String() == "Zzzz" && region.String() == "ZZ" {
		return "", language.No
	}
	script, confidence := parsed.Script()
	if slices.Contains(noWritingSystem, script.String()) {
		return "", language.No
	}
	return script.String(), confidence
}

// usualScripts are the scripts a language of the cards is written in where its
// tag alone does not tell the tag library which, in the order a refusal names
// them: the library only guesses at the script of Serbian, Uzbek, Azerbaijani,
// Filipino, Kurdish, Hausa, Yoruba and Igbo, and is sure of one script of
// Kazakh, Malay, Punjabi and Bosnian, which are written in two. Kazakh is
// written in Cyrillic and in Latin, Malay in Latin and in Jawi, Punjabi in
// Gurmukhi and in Shahmukhi, Hausa in Latin and in Ajami, and Kurmanji in Latin
// and, in Iraq, in Arabic letters.
var usualScripts = map[string][]string{
	"pa":  {"Guru", "Arab"},
	"ms":  {"Latn", "Arab"},
	"ha":  {"Latn", "Arab"},
	"ku":  {"Latn", "Arab"},
	"sr":  {"Cyrl", "Latn"},
	"kk":  {"Cyrl", "Latn"},
	"bs":  {"Latn", "Cyrl"},
	"uz":  {"Latn", "Cyrl"},
	"az":  {"Latn", "Cyrl"},
	"ckb": {"Arab"},
	"fil": {"Latn"},
	"yo":  {"Latn"},
	"ig":  {"Latn"},
}

// lessonLetters are the letters the texts of a task are held to in its
// language, and whether there are any: those of the script its tag names, or
// of the one its language is written in as a rule. A language of the cards
// written in several scripts in common use, named without one, is held to any
// of them — a task in its other script is no fault — and to the one its place
// is written in too: Punjabi of Pakistan, Kazakh of China, Uzbek of Afghanistan
// and Azerbaijani of Iran in Arabic letters. Any other language whose script
// the library only guesses at is held to nothing, but where the lettering says
// a guess is enough: Chinese, whose likely scripts, Simplified and Traditional,
// are one set of characters.
func lessonLetters(tag string) (lettering, bool) {
	script, confidence := scriptOf(tag)
	switch confidence {
	case language.No:
		return lettering{}, false
	case language.Exact:
		named, known := letterings[script]
		return named, known
	}
	if usual, listed := usualScripts[languageOf(tag)]; listed {
		if !slices.Contains(usual, script) {
			usual = append(slices.Clip(usual), script)
		}
		return letteringOf(usual), true
	}
	likely, known := letterings[script]
	if !known || confidence == language.Low && !likely.guessIsEnough {
		return lettering{}, false
	}
	return likely, true
}

// writtenIn is the script a lesson's language is written in, as its tag names
// it or as its language is written as a rule, or nothing where the tag tells
// none.
func writtenIn(tag string) string {
	script, _ := scriptOf(tag)
	return script
}

// languageOf is the language a tag names, as the tag library writes it — "fil"
// of "tl", "he" of "iw" — or nothing where the tag names none it can be sure
// of: "und", whose language the library would only guess at, and a tag that is
// none.
func languageOf(tag string) string {
	parsed, err := language.Parse(tag)
	if err != nil {
		return ""
	}
	base, confidence := parsed.Base()
	if confidence < language.High {
		return ""
	}
	return base.String()
}

// primarySubtag is the language a tag names, lowercased: "zh" of "zh-Hant-TW",
// "en" of "EN_us", and nothing of an empty tag.
func primarySubtag(tag string) string {
	primary, _, _ := strings.Cut(strings.ToLower(strings.TrimSpace(tag)), "-")
	primary, _, _ = strings.Cut(primary, "_")
	return primary
}

// lengthIn is how long a text is in a unit. Words are counted as a child reads
// them, between spaces or the word spaces of Ethiopic; characters are the
// letters and digits. A mark rides on the letter it is written over — the
// vowels and tones of Thai, Lao, Khmer, Myanmar and Tibetan are marks — and
// counting it too would make a syllable three characters long where a child
// reads one. Punctuation is not counted either.
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
	for _, token := range strings.FieldsFunc(text, betweenWords) {
		if strings.IndexFunc(token, readAloud) >= 0 {
			count++
		}
	}
	return count
}

// ethiopicWordspace is the mark Ethiopic once set between words, and Amharic
// still sets there in place of a space.
const ethiopicWordspace = '\u1361'

// betweenWords says whether a character stands between two words: whitespace,
// or the word space of Ethiopic.
func betweenWords(r rune) bool { return unicode.IsSpace(r) || r == ethiopicWordspace }

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

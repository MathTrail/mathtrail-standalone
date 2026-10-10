package checks

import (
	"slices"
	"strings"
	"unicode"

	"golang.org/x/text/unicode/norm"
)

// labelling is how the wording of a lesson names a picture's labels: the words
// its language lists two of them with and the phrases it spans a range with,
// and whether it reads its capitals that are drawn as Latin ones as Latin.
type labelling struct {
	listing listing
	greek   bool
}

// labellingOf is how the wording of a lesson in this language names labels.
func labellingOf(lesson string) labelling {
	return labelling{listing: listingIn(lesson), greek: writtenIn(lesson) == "Grek"}
}

// asLatin is a wording or an option as the labels in it are read. A picture
// writes its labels in Latin capitals, and a lesson written in Greek names its
// points with Greek ones: there a run of Greek capitals standing apart from
// any letter — a point Α, a segment ΑΒ, a cell Β2 — each of them drawn as a
// Latin capital is, is read as those Latin capitals, so that the Α of the
// wording and the A of a picture are one label, as no child could tell them
// apart. A Greek word that opens with such a capital, Από or Μαρία, is a word,
// and a run that holds a capital drawn as none, ΑΒΓ, is left Greek.
func (by labelling) asLatin(text string) string {
	if !by.greek {
		return text
	}
	runes := []rune(text)
	for start := 0; start < len(runes); {
		if !greekCapital(runes[start]) {
			start++
			continue
		}
		end := start + 1
		for end < len(runes) && greekCapital(runes[end]) {
			end++
		}
		if standsApart(runes, start, end) && drawnAsLatin(runes[start:end]) {
			for i := start; i < end; i++ {
				runes[i] = latinTwins[runes[i]]
			}
		}
		start = end
	}
	return string(runes)
}

// greekCapital says whether a character is a capital of the Greek alphabet.
func greekCapital(r rune) bool { return unicode.Is(unicode.Greek, r) && unicode.IsUpper(r) }

// standsApart says whether no letter joins a run of characters to what stands
// on either side of it: a digit may, as the 2 of a cell Β2.
func standsApart(runes []rune, start, end int) bool {
	return (start == 0 || !unicode.IsLetter(runes[start-1])) && (end == len(runes) || !unicode.IsLetter(runes[end]))
}

// drawnAsLatin says whether every capital of a run is drawn as a Latin one.
func drawnAsLatin(run []rune) bool {
	for _, r := range run {
		if _, drawnAlike := latinTwins[r]; !drawnAlike {
			return false
		}
	}
	return true
}

// latinTwins are the Greek capitals drawn as Latin ones, by the Latin capital
// each looks like: Α, Β, Ε, Ζ, Η, Ι, Κ, Μ, Ν, Ο, Ρ, Τ, Υ and Χ.
var latinTwins = map[rune]rune{
	0x391: 'A', 0x392: 'B', 0x395: 'E', 0x396: 'Z', 0x397: 'H', 0x399: 'I', 0x39A: 'K',
	0x39C: 'M', 0x39D: 'N', 0x39F: 'O', 0x3A1: 'P', 0x3A4: 'T', 0x3A5: 'Y', 0x3A7: 'X',
}

// listingWords are the words each language of the cards lists two things with,
// or puts one with the other — and, or, with — by the language a tag names.
// Between two ends of a range they name those two alone, not what lies between
// them. A language's own words alone are read in its lessons, since a word that
// lists in one language may join a range in another: the a of Czech and
// Slovak is "and", and the a of Spanish, Portuguese, Italian and Catalan is
// "to", de A a D. A language written in a second script has its words in that
// one too where they are known: Chinese in both its scripts, Punjabi in
// Shahmukhi, Malay in Jawi, Kurmanji in Arabic letters, and Serbian, Bosnian,
// Uzbek, Kazakh and Azerbaijani in Cyrillic and in Latin. Two labels joined by
// a word missing here, such as one of Ajami, are read as a range: that reading
// may let a picture's label go unnamed, and never refuses a task.
var listingWords = map[string][]string{
	"en":  {"and", "or", "nor", "plus", "with"},
	"zh":  {"和", "与", "與", "及", "或", "或者", "跟", "同", "以及", "还有", "還有"},
	"hi":  {"और", "या", "तथा", "एवं", "व"},
	"es":  {"y", "e", "o", "u", "ni", "con"},
	"ar":  {"و", "أو", "او", "أم"},
	"fr":  {"et", "ou", "avec"},
	"bn":  {"ও", "এবং", "বা", "আর", "অথবা"},
	"pt":  {"e", "ou", "nem", "com"},
	"ru":  {"и", "или", "либо", "с"},
	"ur":  {"اور", "و", "یا"},
	"id":  {"dan", "atau", "serta", "dengan"},
	"de":  {"und", "oder", "sowie", "mit"},
	"ja":  {"と", "や", "か", "または", "および", "及び", "又は"},
	"tr":  {"ve", "veya", "ya", "ile", "yahut"},
	"ko":  {"와", "과", "및", "또는", "이나", "나", "하고", "랑", "이랑"},
	"vi":  {"và", "hoặc", "với", "cùng"},
	"it":  {"e", "ed", "o", "od", "oppure", "con"},
	"fa":  {"و", "یا"},
	"pl":  {"i", "oraz", "lub", "albo", "czy", "z"},
	"uk":  {"і", "й", "та", "або", "чи", "з"},
	"th":  {"และ", "หรือ", "กับ"},
	"nl":  {"en", "of", "met"},
	"pa":  {"ਅਤੇ", "ਤੇ", "ਜਾਂ", "ਨਾਲ", "اتے", "یا", "نال"},
	"te":  {"మరియు", "లేదా", "తో"},
	"mr":  {"आणि", "व", "किंवा", "सह"},
	"ta":  {"மற்றும்", "அல்லது", "உடன்", "யும்", "உம்"},
	"gu":  {"અને", "અથવા", "કે", "સાથે"},
	"fil": {"at", "o", "kasama"},
	"sw":  {"na", "au", "pamoja", "wala"},
	"ms":  {"dan", "atau", "serta", "dengan", "دان", "اتاو", "دڠن"},
	"ro":  {"și", "şi", "sau", "ori", "cu"},
	"cs":  {"a", "i", "nebo", "či", "s", "se", "ani"},
	"hu":  {"és", "vagy", "meg", "valamint"},
	"el":  {"και", "κι", "ή", "με", "ούτε"},
	"sv":  {"och", "eller", "med", "samt"},
	"da":  {"og", "eller", "med", "samt"},
	"nb":  {"og", "eller", "med", "samt"},
	"fi":  {"ja", "tai", "sekä", "kanssa"},
	"uz":  {"va", "yoki", "bilan", "hamda", "ва", "ёки", "билан", "ҳамда"},
	"kk":  {"және", "мен", "бен", "пен", "немесе", "не", "jáne", "men", "ben", "pen", "nemese"},
	"kn":  {"ಮತ್ತು", "ಹಾಗೂ", "ಅಥವಾ", "ಜೊತೆ"},
	"ml":  {"യും", "ഉം", "അല്ലെങ്കിൽ", "അഥവാ", "കൂടെ", "ഒപ്പം"},
	"or":  {"ଓ", "ଏବଂ", "ବା", "କିମ୍ବା", "ସହ", "ସହିତ"},
	"mai": {"आ", "आओर", "वा", "अथवा", "संग"},
	"ps":  {"او", "یا", "سره"},
	"he":  {"ו", "וגם", "או", "עם"},
	"az":  {"və", "ya", "yaxud", "ilə", "вә", "ја", "јахуд", "илә"},
	"ku":  {"û", "an", "yan", "bi", "ligel", "و", "یان"},
	"ckb": {"و", "یان", "یا", "لەگەڵ"},
	"ha":  {"da", "ko", "tare"},
	"yo":  {"àti", "ati", "tàbí", "tabi", "pẹ̀lú", "pelu"},
	"ig":  {"na", "ma"},
	"am":  {"እና", "ና", "ወይም", "ጋር"},
	"zu":  {"na", "no", "ne", "noma", "kanye", "futhi"},
	"sr":  {"и", "или", "са", "с", "ни", "i", "ili", "sa", "s", "ni"},
	"hr":  {"i", "ili", "s", "sa", "te", "ni"},
	"bs":  {"i", "ili", "s", "sa", "ni", "и", "или", "с", "са"},
	"bg":  {"и", "или", "с", "със", "нито"},
	"sk":  {"a", "i", "alebo", "či", "s", "so", "ani"},
	"ca":  {"i", "o", "amb", "ni"},
	"lt":  {"ir", "bei", "ar", "arba", "su"},
	"hy":  {"և", "եւ", "ու", "կամ", "հետ"},
	"ka":  {"და", "ან", "თუ", "თან"},
}

// listsTheEnds are words some language lists with and another joins the ends
// of a range with: the a of Czech and Slovak, "and", which is the a of Spanish,
// Portuguese, Italian and Catalan, "to".
var listsTheEnds = []string{"a"}

// spanningPhrases are the phrases a language spans a range with though a word
// of each lists — "up to and including": Swedish A till och med D, Danish and
// Norwegian til og med, Dutch tot en met, Indonesian and Malay sampai dengan,
// Swiss German bis und mit, and Greek έως και.
var spanningPhrases = map[string][]string{
	"sv": {"till och med"},
	"da": {"til og med"},
	"nb": {"til og med"},
	"nl": {"tot en met"},
	"id": {"sampai dengan"},
	"ms": {"sampai dengan"},
	"de": {"bis und mit"},
	"el": {"έως και", "ως και", "μέχρι και"},
}

// listing is how a language joins two labels in words: the words that list
// them, and the phrases that span a range though a word of them lists, each
// as the words between two ends are compared with it.
type listing struct{ words, phrases []string }

// listings are how each language of the cards joins two labels in words.
var listings = func() map[string]listing {
	read := make(map[string]listing, len(listingWords))
	for language, words := range listingWords {
		read[language] = listing{words: readAll(words), phrases: readAll(spanningPhrases[language])}
	}
	return read
}()

// sharedListing is how a lesson with no listing of its own is read — a lesson
// with no language, or in one the cards do not speak: with the words of every
// language of the cards but those another of them joins a range with, and
// with every phrase that spans one.
var sharedListing = func() listing {
	var shared listing
	for _, own := range listings {
		for _, word := range own.words {
			if !slices.Contains(shared.words, word) && !slices.Contains(listsTheEnds, word) {
				shared.words = append(shared.words, word)
			}
		}
		for _, phrase := range own.phrases {
			if !slices.Contains(shared.phrases, phrase) {
				shared.phrases = append(shared.phrases, phrase)
			}
		}
	}
	return shared
}()

// listingIn is how a lesson in this language joins two labels in words.
// Norwegian, which a host may name by the tag of the language as a whole
// rather than by Bokmål's, is read with the shared listing, which holds
// Bokmål's.
func listingIn(lesson string) listing {
	if own, listed := listings[languageOf(lesson)]; listed {
		return own
	}
	return sharedListing
}

// lists says whether a word between two ends lists them, or is the slash of
// A/D, which lists in any language: as it is written, or without the hyphen
// or the maqaf that joins it to a label — Hebrew writes "and" as ו-D, Zulu as
// no-D, and Tamil as A-யும்.
func (l listing) lists(word string) bool {
	if word == "/" {
		return true
	}
	read := compared(word)
	if bare := strings.Trim(read, wordHyphens); bare != "" {
		read = bare
	}
	return slices.Contains(l.words, read)
}

// spans says whether the words between two ends are a phrase that spans a
// range.
func (l listing) spans(words []string) bool {
	return slices.Contains(l.phrases, compared(strings.Join(words, " ")))
}

// readAll are words as the words between two ends are compared with them.
func readAll(words []string) []string {
	read := make([]string, len(words))
	for i, word := range words {
		read[i] = compared(word)
	}
	return read
}

// compared is a word as it is compared with another: in small letters, its
// accents composed.
func compared(word string) string { return norm.NFC.String(strings.ToLower(word)) }

// wordHyphens are the marks that join a word to the label beside it: the
// hyphen-minus, the hyphen, the hyphen that does not break and the Hebrew
// maqaf.
const wordHyphens = "-‐‑־"

// marksAList says whether a joint holds a mark that lists or reckons, or ends a
// sentence. A colon followed by a small letter is none of them: it joins a case
// ending to the label before it, as Finnish writes A:sta, "from A", and Swedish
// A:s.
func marksAList(joint string) bool {
	runes := []rune(joint)
	for i, r := range runes {
		if r == ':' && i+1 < len(runes) && unicode.IsLower(runes[i+1]) {
			continue
		}
		if strings.ContainsRune(listingMarks, r) || strings.ContainsRune(sentenceEnds, r) {
			return true
		}
	}
	return false
}

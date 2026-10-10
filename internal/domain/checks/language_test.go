package checks_test

import (
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"unicode"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/checks"
)

// liveEnglish is the task a model wrote in English for a lesson in Russian,
// which the service took before the letters of a task were checked.
const liveEnglish = "Three robots sit around a table on a space station: Alpha, Beta, and Gamma. " +
	"Alpha sits to the right of the red seat. Who sits in the red seat?"

// russianQuestion is a question written in Russian.
const russianQuestion = "Четыре корабля стыкуются парами. Сколько пар получится?"

// asking is a task whose question alone has anything to count: its hint, its
// solution and its explanations are empty, and say nothing about the letters
// they are in.
func asking(question string) *checks.Task { return &checks.Task{Question: question} }

// refusedIn says whether the language check refuses a task asked for in a
// language.
func refusedIn(task *checks.Task, language string) bool {
	return len(checks.Language(task, language)) > 0
}

// Each text is held to the letters of the language its lesson is in, counted
// as a child reads it: what a model wrote in another language is refused, and
// a question that names its points in Latin capitals, works with x or names a
// child Tom is not.
func TestATextIsHeldToTheLettersOfItsLesson(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		name, language, question string
		refused                  bool
	}{
		{"Russian", "ru", russianQuestion, false},
		{"English in a lesson in Russian", "ru", liveEnglish, true},
		{"points named in Latin capitals", "ru", "AB = 3, BC = 4. Найди AC.", false},
		{"an unknown called x", "ru", "Найди x, если 2x + 3 = 7.", false},
		{"two names among Russian words", "ru", "Кто выше, Tom или Mary?", false},
		{"two names that outweigh the Russian", "ru", "Tom или Mary?", true},
		{"names and Russian half and half", "ru", "Сравни Tom и Mary.", false},
		{"the orders of three letters", "ru", "Выпишем все порядки: abc, acb, bac, bca, cab, cba. Их 6.", false},
		{"Korean, its script named", "ko-Hang", languagesOfTheCards["ko"], false},
		{"English in a lesson in Korean, its script named", "ko-Hang", liveEnglish, true},
		{"Japanese in hiragana, its script named", "ja-Hira", "ふねは なんくみ できますか？", false},
		{"English in a lesson in Japanese hiragana", "ja-Hira", liveEnglish, true},
		{"Ukrainian", "uk", "Скільки пар можуть утворити чотири кораблі?", false},
		{"English in a lesson in Ukrainian", "uk", liveEnglish, true},
		{"Arabic", "ar", "كم زوجًا يمكن أن تكوّن أربع سفن؟", false},
		{"English in a lesson in Arabic", "ar", liveEnglish, true},
		{"Persian, a word held apart without a space", "fa", "چهار کشتی چند جفت می\u200cسازند؟", false},
		{"Hindi", "hi", "चार जहाज़ कितने जोड़े बना सकते हैं?", false},
		{"English in a lesson in Hindi", "hi", liveEnglish, true},
		{"Bengali", "bn", "চারটি জাহাজ কতগুলো জোড়া তৈরি করতে পারে?", false},
		{"Thai", "th", "เรือสี่ลำจับคู่กันได้กี่คู่", false},
		{"English in a lesson in Thai", "th", liveEnglish, true},
		{"Japanese with the mark of a long vowel", "ja", "ケーキが4つあります。何人で分けられますか？", false},
		{"English in a lesson in Japanese", "ja", liveEnglish, true},
		{"Chinese naming its children in Latin letters", "zh-Hans", "Tom和Mary谁高？", false},
		{"Chinese with more Latin letters than characters", "zh-Hans",
			"Tom、Mary和Lily站成一排。Tom在Mary前面，Lily在Tom前面。谁站在最前面？", false},
		{"Chinese tagged without its script", "zh", "四艘飞船两两对接，可以组成几对？", false},
		{"English in a lesson in Chinese tagged without its script", "zh", liveEnglish, true},
		{"Traditional Chinese", "zh-TW", "四艘太空船兩兩對接，可以組成幾對？", false},
		{"Korean naming its children in Latin letters", "ko", "Tom과 Mary 중 누가 더 큰가요?", false},
		{"English in a lesson in Korean", "ko", liveEnglish, true},
		{"English", "en", liveEnglish, false},
		{"Russian in a lesson in English", "en", russianQuestion, true},
		{"Russian naming its points in Latin capitals, in a lesson in English", "en", "AB = 3, BC = 4. Найди AC.", true},
		{"Greek with its articles of one letter among Latin names", "el", "Ο Tom και η Mary;", false},
		{"Greek in a lesson in Russian, its single letters symbols", "ru", "Ο Tom και η Mary;", true},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			if got := refusedIn(asking(test.question), test.language); got != test.refused {
				t.Errorf("Language(%q, %q) refused %v, want %v", test.question, test.language, got, test.refused)
			}
		})
	}
}

// A lesson in a language that names no script, or one the check can be sure
// of in no language of the cards, holds a task to nothing: Mongolian is
// written in two alphabets, and a guess at one would refuse every task written
// in the other.
func TestALessonHeldToNoScriptRefusesNothing(t *testing.T) {
	t.Parallel()

	for _, test := range []struct{ name, language string }{
		{"no tag", ""},
		{"a language not determined", "und"},
		{"several languages", "mul"},
		{"no language at all", "zxx"},
		{"private words alone", "x-abc"},
		{"several scripts in common use", "und-Zyyy"},
		{"no tag at all", "not a tag"},
		{"Mongolian, its script not named", "mn"},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			if refusedIn(asking(liveEnglish), test.language) {
				t.Errorf("Language(%q) refused, want nothing refused", test.language)
			}
		})
	}
}

// A language of the cards written in several scripts in common use, its tag
// naming none, is held to any of them, and to the one its place is written in
// as well: a task in its other script is no fault. A script the tag names is
// held to alone, whichever of the two it is, and a task in letters none of its
// scripts has is refused.
func TestALanguageWrittenInSeveralScriptsIsHeldToAnyOfThem(t *testing.T) {
	t.Parallel()

	serbianInLatin := "Koliko parova mogu da naprave četiri broda?"
	serbianInCyrillic := languagesOfTheCards["sr"]
	kazakhInLatin := "Tört kemeden neshe jup quraluǵa bolady?"
	punjabiInShahmukhi := "چار جہاز کِنّے جوڑے بنا سکدے نیں؟"
	for _, test := range []struct {
		name, language, question string
		refused                  bool
	}{
		{"Serbian in Latin, its script not named", "sr", serbianInLatin, false},
		{"Serbian in Cyrillic, its script not named", "sr", serbianInCyrillic, false},
		{"Hebrew in a lesson in Serbian", "sr", languagesOfTheCards["he"], true},
		{"Bosnian in Cyrillic, its script not named", "bs", serbianInCyrillic, false},
		{"Uzbek in Cyrillic, its script not named", "uz", "Тўртта кема нечта жуфт ҳосил қила олади?", false},
		{"Uzbek of Afghanistan in Arabic letters", "uz-AF", "تورتته کېمه نېچته جفت جوړولی شي؟", false},
		{"Azerbaijani in Cyrillic, its script not named", "az", "Дөрд ҝәми нечә ҹүт јарада биләр?", false},
		{"Kazakh in Latin, its script not named", "kk", kazakhInLatin, false},
		{"Kazakh of China in Arabic letters", "kk-CN", "تورت كەمە نەشە جۇپ قۇراي الادى؟", false},
		{"Kazakh in Latin, Cyrillic named", "kk-Cyrl", kazakhInLatin, true},
		{"Malay in Jawi, its script not named", "ms", "كاڤل ايمڤت", false},
		{"Hausa in Ajami, its script not named", "ha", "جِرَاغٍ هُدُ زَا سُ اِيَا", false},
		{"Kurmanji in Arabic letters, its script not named", "ku", "چار گەمی دشێن چەند جۆت چێکەن؟", false},
		{"Punjabi in Gurmukhi", "pa", languagesOfTheCards["pa"], false},
		{"Punjabi in Shahmukhi, its script not named", "pa", punjabiInShahmukhi, false},
		{"Punjabi of Pakistan in Shahmukhi", "pa-PK", punjabiInShahmukhi, false},
		{"English in a lesson in Punjabi", "pa", liveEnglish, true},
		{"English in a lesson in Punjabi of Pakistan", "pa-PK", liveEnglish, true},
		{"English in a lesson in Sorani", "ckb", liveEnglish, true},
		{"Serbian in Latin, as named", "sr-Latn", serbianInLatin, false},
		{"Serbian in Cyrillic, Latin named", "sr-Latn", serbianInCyrillic, true},
		{"Serbian in Cyrillic, as named", "sr-Cyrl", serbianInCyrillic, false},
		{"Serbian in Latin, Cyrillic named", "sr-Cyrl", serbianInLatin, true},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			if got := refusedIn(asking(test.question), test.language); got != test.refused {
				t.Errorf("Language(%q, %q) refused %v, want %v", test.question, test.language, got, test.refused)
			}
		})
	}
}

// A text with nothing a child reads as a word of a language — numbers, signs,
// labels, units of measure, a symbol — says nothing about the language it is
// in, and neither does a task that could not be read.
func TestATextWithNothingToCountIsNotRefused(t *testing.T) {
	t.Parallel()

	for _, text := range []string{
		"", "   ", "AB = 3, BC = 4, CD = 5.", "x + y = 12", "2πr", "ＡＢ＝３", "ー", "7 · 8 = 56", "I",
		"5 cm + 3 cm = ? cm", "2 kg 300 g", "1 km 200 m", "45 min", "250 ml", "12 cm²", "9 am", "5 ft", "S = ab",
	} {
		t.Run(text, func(t *testing.T) {
			t.Parallel()

			for _, language := range []string{"ru", "zh", "en"} {
				if refusedIn(asking(text), language) {
					t.Errorf("Language(%q, %q) refused, want nothing to count and nothing refused", text, language)
				}
			}
		})
	}
	if problems := checks.Language(nil, "ru"); problems != nil {
		t.Errorf("Language(nil) = %v, want nothing: a task that could not be read is the structure's to refuse", problems)
	}
}

// languagesOfTheCards are the languages the cards speak, each with a question
// written in it.
var languagesOfTheCards = map[string]string{
	"am":      "አራት መርከቦች ስንት ጥንዶች መፍጠር ይችላሉ?",
	"ar":      "كم زوجًا يمكن أن تكوّن أربع سفن؟",
	"az":      "Dörd gəmi neçə cüt yarada bilər?",
	"bg":      "Колко двойки могат да образуват четири кораба?",
	"bn":      "চারটি জাহাজ কতগুলো জোড়া তৈরি করতে পারে?",
	"bs":      "Koliko parova mogu formirati četiri broda?",
	"ca":      "Quantes parelles poden formar quatre naus?",
	"ckb":     "چوار کەشتی دەتوانن چەند جووت دروست بکەن؟",
	"cs":      "Kolik dvojic mohou vytvořit čtyři lodě?",
	"da":      "Hvor mange par kan fire skibe danne?",
	"de":      "Wie viele Paare können vier Schiffe bilden?",
	"el":      "Πόσα ζευγάρια μπορούν να σχηματίσουν τέσσερα πλοία;",
	"en":      "How many pairs can four ships make?",
	"es":      "¿Cuántas parejas pueden formar cuatro naves?",
	"fa":      "چهار کشتی چند جفت می\u200cسازند؟",
	"fi":      "Kuinka monta paria neljä alusta voi muodostaa?",
	"fil":     "Ilang pares ang mabubuo ng apat na barko?",
	"fr":      "Combien de paires quatre vaisseaux peuvent-ils former ?",
	"gu":      "ચાર જહાજો કેટલી જોડીઓ બનાવી શકે?",
	"ha":      "Jirage huɗu za su iya yin ma'aurata nawa?",
	"he":      "כמה זוגות יכולות ארבע ספינות ליצור?",
	"hi":      "चार जहाज़ कितने जोड़े बना सकते हैं?",
	"hr":      "Koliko parova mogu napraviti četiri broda?",
	"hu":      "Hány párt alkothat négy hajó?",
	"hy":      "Քանի՞ զույգ կարող են կազմել չորս նավերը։",
	"id":      "Berapa pasangan yang bisa dibentuk empat kapal?",
	"ig":      "Ụgbọ mmiri anọ nwere ike ịmepụta abụọ abụọ ole?",
	"it":      "Quante coppie possono formare quattro navi?",
	"ja":      "四せきの船で何組のペアができますか？",
	"ka":      "რამდენ წყვილს შეადგენს ოთხი ხომალდი?",
	"kk":      "Төрт кеме неше жұп құрай алады?",
	"kn":      "ನಾಲ್ಕು ಹಡಗುಗಳು ಎಷ್ಟು ಜೋಡಿಗಳನ್ನು ರಚಿಸಬಹುದು?",
	"ko":      "우주선 네 대로 몇 쌍을 만들 수 있나요?",
	"ku":      "Çar keştî dikarin çend cot çêbikin?",
	"lt":      "Kiek porų gali sudaryti keturi laivai?",
	"mai":     "चारि टा जहाज कतेक जोड़ी बना सकैत अछि?",
	"ml":      "നാല് കപ്പലുകൾക്ക് എത്ര ജോടികൾ ഉണ്ടാക്കാം?",
	"mr":      "चार जहाजे किती जोड्या बनवू शकतात?",
	"ms":      "Berapa pasangan yang boleh dibentuk oleh empat kapal?",
	"nb":      "Hvor mange par kan fire skip danne?",
	"nl":      "Hoeveel paren kunnen vier schepen vormen?",
	"or":      "ଚାରିଟି ଜାହାଜ କେତୋଟି ଯୋଡ଼ି ତିଆରି କରିପାରିବ?",
	"pa":      "ਚਾਰ ਜਹਾਜ਼ ਕਿੰਨੇ ਜੋੜੇ ਬਣਾ ਸਕਦੇ ਹਨ?",
	"pl":      "Ile par mogą utworzyć cztery statki?",
	"ps":      "څلور کښتۍ څو جوړې جوړولی شي؟",
	"pt":      "Quantos pares quatro naves podem formar?",
	"ro":      "Câte perechi pot forma patru nave?",
	"ru":      "Сколько пар могут составить четыре корабля?",
	"sk":      "Koľko dvojíc môžu vytvoriť štyri lode?",
	"sr":      "Колико парова могу да направе четири брода?",
	"sv":      "Hur många par kan fyra skepp bilda?",
	"sw":      "Meli nne zinaweza kuunda jozi ngapi?",
	"ta":      "நான்கு கப்பல்கள் எத்தனை ஜோடிகளை உருவாக்க முடியும்?",
	"te":      "నాలుగు ఓడలు ఎన్ని జతలు ఏర్పరచగలవు?",
	"th":      "เรือสี่ลำจับคู่กันได้กี่คู่",
	"tr":      "Dört gemi kaç çift oluşturabilir?",
	"uk":      "Скільки пар можуть утворити чотири кораблі?",
	"ur":      "چار جہاز کتنے جوڑے بنا سکتے ہیں؟",
	"uz":      "To'rtta kema nechta juft hosil qila oladi?",
	"vi":      "Bốn con tàu có thể tạo thành bao nhiêu cặp?",
	"yo":      "Ọkọ̀ ojú omi mẹ́rin lè ṣe méjì-méjì mélòó?",
	"zh-Hans": "四艘飞船两两对接，可以组成几对？",
	"zh-Hant": "四艘太空船兩兩對接，可以組成幾對？",
	"zu":      "Imikhumbi emine ingakha amapheya amangaki?",
}

// sinhala is a question in a script none of the languages of the cards is
// written in.
const sinhala = "නැව් හතරකට යුගල කීයක් සෑදිය හැකිද?"

// Every language the cards speak holds a task to its own letters: a question
// written in it passes, and one in letters it is not written in is refused.
func TestEveryLanguageTheCardsSpeakIsHeldToItsLetters(t *testing.T) {
	t.Parallel()

	for language, question := range languagesOfTheCards {
		t.Run(language, func(t *testing.T) {
			t.Parallel()

			if refusedIn(asking(question), language) {
				t.Errorf("a question in %s is refused in a lesson in %s, want it let through", language, language)
			}
			if !refusedIn(asking(sinhala), language) {
				t.Errorf("a question in Sinhala is let through in a lesson in %s, want it refused", language)
			}
		})
	}
}

// A lesson in a language written in no Latin letters refuses a task written in
// English, as the model may write one when it copies the reference tasks.
func TestEnglishIsRefusedInEveryLessonWrittenInOtherLetters(t *testing.T) {
	t.Parallel()

	for _, language := range append(slices.Clone(lessonsInOtherLetters), "el") {
		t.Run(language, func(t *testing.T) {
			t.Parallel()

			if !refusedIn(asking(liveEnglish), language) {
				t.Errorf("a task in English is let through in a lesson in %s, want it refused", language)
			}
		})
	}
}

// Every dictionary of the cards is in a language the check holds a task to:
// one added for a language the check did not know would offer the lessons a
// language whose tasks no one reads for their letters.
func TestEveryDictionaryOfTheCardsIsInALanguageTheCheckKnows(t *testing.T) {
	t.Parallel()

	files, err := filepath.Glob(filepath.Join("..", "..", "..", "web", "locales", "*.json"))
	if err != nil || len(files) == 0 {
		t.Fatalf("the cards' dictionaries are %v, %v; want some", files, err)
	}
	for _, file := range files {
		language := strings.TrimSuffix(filepath.Base(file), ".json")
		if _, known := languagesOfTheCards[language]; !known {
			t.Errorf("the cards speak %s, and languagesOfTheCards has no question in it to hold the check to", language)
		}
	}
}

// A refusal names the texts that are in other letters, the language and its
// letters, and what to write; it quotes nothing of the task, and no letter of
// an option.
func TestTheRefusalNamesWhatToMendAndQuotesNothing(t *testing.T) {
	t.Parallel()

	english := validDraft().Task
	for _, test := range []struct {
		name string
		task *checks.Task
		want string
	}{
		{"every text in English", english,
			"task.question, task.hint, task.solution and the explanations in task.distractors are mostly not in " +
				"Cyrillic letters, which ru, the language of the lesson, is written in."},
		{"the hint alone in English", withHint(russianDraft().Task, english.Hint), "task.hint is mostly not in Cyrillic letters"},
		{"the explanations alone in English", withDistractors(russianDraft().Task, english.Distractors),
			"the explanations in task.distractors are mostly not in Cyrillic letters"},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			problems := checks.Language(test.task, "ru")
			if len(problems) != 1 || problems[0].Code != checks.CodeWrongLanguage {
				t.Fatalf("Language() = %v, want one problem under %q", problems, checks.CodeWrongLanguage)
			}
			message := problems[0].Message
			for _, said := range []string{test.want, "Write every text the child reads in ru", "the names in them"} {
				if !strings.Contains(message, said) {
					t.Errorf("the refusal says %q, want it to say %q", message, said)
				}
			}
			repeatsNone(t, []string{message}, textsOf(test.task))
		})
	}
}

// withHint is a task with its hint written anew.
func withHint(task *checks.Task, hint string) *checks.Task {
	task.Hint = hint
	return task
}

// withDistractors is a task with its explanations written anew.
func withDistractors(task *checks.Task, distractors map[string]checks.Distractor) *checks.Task {
	task.Distractors = distractors
	return task
}

// textsOf is every text of a task a child reads.
func textsOf(task *checks.Task) []string {
	texts := []string{task.Question, task.Hint, task.Solution}
	for _, option := range task.Options {
		texts = append(texts, option)
	}
	for _, distractor := range task.Distractors {
		texts = append(texts, distractor.Text)
	}
	return texts
}

// The options are not held to the letters: an option may be a label or a name
// in a task in any language.
func TestTheOptionsAreNotHeldToTheLetters(t *testing.T) {
	t.Parallel()

	task := russianDraft().Task
	task.Options = map[string]string{"A": "Tom", "B": "Mary", "C": "AB", "D": "Kim and Ann", "E": "Nobody"}
	if problems := checks.Language(task, "ru"); len(problems) != 0 {
		t.Errorf("Language() = %v for a Russian task with options in Latin letters, want nothing refused", problems)
	}
}

// The explanations are held to the letters together: one of a few words that
// names two children in Latin letters, refused as a text of its own, is
// outweighed by the rest; four of them are not.
func TestTheExplanationsAreHeldToTheLettersTogether(t *testing.T) {
	t.Parallel()

	const named = "Tom и Kim."
	if !refusedIn(asking(named), "ru") {
		t.Fatalf("Language(%q) as a question let through, want it refused on its own", named)
	}
	task := russianDraft().Task
	task.Distractors["B"] = checks.Distractor{Trap: "missed_case", Text: named}
	if problems := checks.Language(task, "ru"); len(problems) != 0 {
		t.Errorf("Language() = %v with one explanation naming Tom and Kim, want nothing refused", problems)
	}
	for key, distractor := range task.Distractors {
		distractor.Text = named
		task.Distractors[key] = distractor
	}
	problems := checks.Language(task, "ru")
	if len(problems) != 1 || !strings.Contains(problems[0].Message, "the explanations in task.distractors are") {
		t.Errorf("Language() = %v with every explanation naming Tom and Kim, want the explanations refused", problems)
	}
}

// Two languages written in the same letters are not told apart: the letters
// are counted, and no language is guessed at.
func TestLanguagesWrittenInTheSameLettersAreNotToldApart(t *testing.T) {
	t.Parallel()

	for _, test := range []struct{ language, question string }{
		{"uk", russianQuestion},
		{"de", liveEnglish},
		{"ja", languagesOfTheCards["zh-Hans"]},
	} {
		t.Run(test.language, func(t *testing.T) {
			t.Parallel()

			if refusedIn(asking(test.question), test.language) {
				t.Errorf("Language(%q, %q) refused, want letters the lesson shares let through", test.question, test.language)
			}
		})
	}
}

// The texts come from the chat's model and the language from the chat:
// whatever they hold, the check ends in at most one problem of its own code,
// refuses nothing it has no language or no letter to judge, and comes to the
// same whichever letters the explanations sit under.
func FuzzLanguage(f *testing.F) {
	f.Add(liveEnglish, "Who sits next to Alpha?", "Alpha is not in it.", "Beta sits there.", "Gamma does.", "ru")
	f.Add("AB = 3, BC = 4. Найди AC.", "", "", "", "", "ru")
	f.Add("Tom和Mary谁高？", "ー", "ＡＢ", "\xff", "e\u0301", "zh")
	f.Add("x", "π", "2x", "I", "", "zh-Hant-TW")
	f.Add(russianQuestion, "Kim", "", "", "", "sr-Latn")
	f.Add("Ο Tom και η Mary;", "π", "ή", "α", "", "el")
	f.Add(languagesOfTheCards["pa"], "چار جہاز", "", "", "", "pa-PK")
	f.Add(languagesOfTheCards["am"], "ድመት፡ድመት", "", "", "", "am")
	f.Add("", "", "", "", "", "")

	f.Fuzz(func(t *testing.T, question, hint, solution, first, second, language string) {
		task := &checks.Task{Question: question, Hint: hint, Solution: solution, Distractors: map[string]checks.Distractor{
			"A": {Trap: "missed_case", Text: first}, "B": {Trap: "double_count", Text: second},
		}}
		problems := checks.Language(task, language)
		if len(problems) > 1 || len(problems) == 1 && (problems[0].Code != checks.CodeWrongLanguage || problems[0].Message == "") {
			t.Fatalf("Language() = %v, want at most one problem, with a message, under %q", problems, checks.CodeWrongLanguage)
		}
		if len(problems) > 0 && (language == "" || !anyLetter(question, hint, solution, first, second)) {
			t.Fatalf("Language() = %v in %q, want nothing refused without a language or a letter", problems, language)
		}
		elsewhere := &checks.Task{Question: question, Hint: hint, Solution: solution, Distractors: map[string]checks.Distractor{
			"D": {Trap: "missed_case", Text: second}, "E": {Trap: "double_count", Text: first},
		}}
		if again := checks.Language(elsewhere, language); !slices.Equal(messagesOf(again), messagesOf(problems)) {
			t.Fatalf("Language() = %v with the explanations under other letters, want %v", again, problems)
		}
	})
}

// anyLetter says whether any of these texts holds a letter.
func anyLetter(texts ...string) bool {
	return slices.ContainsFunc(texts, func(text string) bool { return strings.IndexFunc(text, unicode.IsLetter) >= 0 })
}

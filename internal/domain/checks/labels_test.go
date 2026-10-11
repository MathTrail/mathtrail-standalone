package checks_test

import (
	"encoding/json"
	"slices"
	"strings"
	"testing"

	"golang.org/x/text/language"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/checks"
)

// labelled is a picture whose labels are these: a table of one label to a
// row, which shows every label it holds and nothing else.
func labelled(labels ...string) json.RawMessage {
	rows := make([][]string, 0, len(labels))
	for _, label := range labels {
		rows = append(rows, []string{label})
	}
	raw, err := json.Marshal(map[string]any{"kind": "table", "rows": rows})
	if err != nil {
		panic(err)
	}
	return raw
}

// matched is what holding a picture to its wording finds, in an English
// lesson, for a task whose options none of the labels is.
func matched(question string, picture json.RawMessage) []checks.Problem {
	return matchedIn("en", question, picture)
}

// matchedIn is what holding a picture to its wording finds, in a lesson in a
// language, for a task whose options none of the labels is.
func matchedIn(lesson, question string, picture json.RawMessage) []checks.Problem {
	task := &checks.Task{
		Question: question, Picture: picture, CorrectAnswer: "A",
		Options: map[string]string{"A": "one hundred", "B": "two", "C": "three", "D": "four", "E": "five"},
	}
	return checks.PictureMatch(task, lesson)
}

func TestAPictureLabelledAsTheWordingNamesMatches(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		name, question string
		picture        json.RawMessage
	}{
		{"points on a number line", "Point A is at 3 and point B is at 7. How far apart are they?", labelled("A", "B")},
		{"a segment named by its ends", "How long is AB, if A is at 3 and B is at 7?", labelled("A", "B")},
		{"boxes numbered in the picture, which no wording has to name", "Box 1 holds 4 balls and box 2 holds 6.",
			labelled("1", "2")},
		{"numbers between two the wording names", "The weights are 1 and 9.", labelled("1", "5", "9")},
		{"ends named only by the segment they close", "The segment AB is 4 cm long.", labelled("A", "B")},
		{"a label of several capitals", "The tower AB stands at the end.", labelled("AB")},
		{"cells named by their row and column", "The robot goes from cell A1 to cell B2.", labelled("A", "B", "1", "2")},
		{"rows and columns named by the first and the last", "Rows are A to D, columns 1 to 3.",
			labelled("A", "B", "C", "D", "1", "2", "3")},
		{"rows named so in another language", "Ряды от A до C.", labelled("A", "B", "C")},
		{"an example of a kind, which has no question", "", labelled("A", "B")},
		{"labels set among the characters of a script without spaces", "点A在3，点B在7。它们相距多远？",
			labelled("A", "B")},
		{"a label opening a question in such a script", "AとBの距離は何センチですか？", labelled("A", "B")},
		{"a label opening an English sentence", "A is at 3 and B is at 7. How far apart are they?", labelled("A", "B")},
		{"a label opening a sentence in Russian", "A находится в 3 см от B.", labelled("A", "B")},
		{"labels opening a sentence in Spanish", "A y B están a 4 cm.", labelled("A", "B")},
		{"labels before the word they name in Turkish", "A kutusunda 5 top, B kutusunda 3 top var.", labelled("A", "B")},
		{"labels with a possessive", "Box A's lid is red and box B's is blue.", labelled("A", "B")},
		{"a label that is the letter I", "Rows G to I are shaded.", labelled("G", "H", "I")},
		{"cells spanned by their corners", "The grid has cells A1 to D4.",
			labelled("A", "B", "C", "D", "1", "2", "3", "4")},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			if problems := matched(test.question, test.picture); len(problems) != 0 {
				t.Errorf("PictureMatch() = %v, want no problems", problems)
			}
		})
	}
}

func TestEachMismatchIsRefused(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		name, question string
		picture        json.RawMessage
		want           string
	}{
		{"a point the wording names and the picture leaves out", "Point D is between A and B.",
			labelled("A", "B"), `task.question names "D"`},
		{"a triangle drawn without its third corner", "The triangle ABC has a right angle at B.",
			labelled("A", "B"), `task.question names "C"`},
		{"a label the wording does not name", "How far apart are the two points?",
			labelled("A", "B"), "task.question does not name 2 of the labels"},
		{"labels standing for names the wording spells otherwise",
			"40 sweets are shared between Anya and Boris in the ratio 2 to 3. How many sweets does Boris get?",
			labelled("A", "B"), "task.question does not name 2 of the labels"},
		{"rows left out, with no range to span them", "Rows A and C are shaded.",
			labelled("A", "B", "C", "D"), "task.question does not name 2 of the labels"},
		{"points between two named points, with no range", "Point A is at 3 and point D is at 9.",
			labelled("A", "B", "C", "D"), "task.question does not name 2 of the labels"},
		{"points too far apart in the wording to be a range",
			"Point A stands well away from the tall tree near point D.",
			labelled("A", "B", "C", "D"), "task.question does not name 2 of the labels"},
		{"ends in two sentences", "Row A is red. Row D is blue.",
			labelled("A", "B", "C", "D"), "task.question does not name 2 of the labels"},
		{"ends set one under the other", "Rows shaded:\nA\nD",
			labelled("A", "B", "C", "D"), "task.question does not name 2 of the labels"},
		{"a capital opening a word with an accent", "Aún quedan 5 litros en el cubo.", labelled("A"),
			"task.question does not name 1 of the labels"},
		{"a range from a capital opening a sentence, with none between", "A shop has three shelves; shelf C holds 9 jars.",
			labelled("A", "B", "C"), "task.question does not name 1 of the labels"},
		{"a capital that is the unit of a number", "The jug holds 3 L of water.", labelled("L"),
			"task.question does not name 1 of the labels"},
		{"a label of several capitals the wording does not name", "The tower stands at the end.", labelled("AB"),
			"task.question does not name 1 of the labels"},
		{"a word in capitals holding a label's letter", "The stone is NOT at P. It is at Q.",
			labelled("P", "Q", "T"), "task.question does not name 1 of the labels"},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			problems := matched(test.question, test.picture)
			if len(problems) != 1 {
				t.Fatalf("PictureMatch() = %v, want exactly one problem", problems)
			}
			if problems[0].Code != checks.CodeDrawingMismatch {
				t.Errorf("code = %q, want %q", problems[0].Code, checks.CodeDrawingMismatch)
			}
			if !strings.Contains(problems[0].Message, test.want) {
				t.Errorf("message = %q, want it to say %q", problems[0].Message, test.want)
			}
		})
	}
}

// A range whose ends the wording names names every label between them, in
// each language of the cards, with the marks any language spans a range by,
// and with the ends joined to a hyphen, an apostrophe, a case ending or a
// script's letters. The Spanish, Portuguese, Italian and Catalan a, "to",
// spans one, though the Czech a lists.
func TestARangeNamesTheLabelsBetweenItsEnds(t *testing.T) {
	t.Parallel()

	for _, test := range []struct{ lesson, question string }{
		{"en", "The rows run from A up to D."},
		{"zh-Hans", "从A到D的行涂了颜色。"},
		{"hi", "A से D तक की पंक्तियाँ रंगी हैं।"},
		{"es", "Las filas de la A a la D están pintadas."},
		{"ar", "الصفوف من A إلى D ملونة."},
		{"fr", "Les rangées de A jusqu'à D sont coloriées."},
		{"bn", "A থেকে D পর্যন্ত সারি রং করা।"},
		{"pt", "As linhas da A à D estão pintadas."},
		{"pt", "As linhas de A a D estão pintadas."},
		{"ru", "Ряды с A по D закрашены."},
		{"ur", "A سے D تک قطاریں رنگی ہیں۔"},
		{"id", "Baris A sampai D diwarnai."},
		{"de", "Die Reihen von A bis D sind bemalt."},
		{"ja", "AからDまでの行に色がある。"},
		{"tr", "A'dan D'ye kadar olan satırlar boyalı."},
		{"ko", "A부터 D까지의 행은 칠해져 있다."},
		{"vi", "Các hàng từ A đến D được tô màu."},
		{"it", "Le righe dalla A alla D sono colorate."},
		{"it", "Le righe da A a D sono colorate."},
		{"fa", "ردیف های A تا D رنگی است."},
		{"pl", "Rzędy od A do D są pomalowane."},
		{"uk", "Ряди від A до D зафарбовані."},
		{"th", "แถว A ถึง D ถูกระบายสี"},
		{"nl", "De rijen van A tot D zijn gekleurd."},
		{"ca", "Les files de l'A a la D estan pintades."},
		{"cs", "Řady od A do D jsou vybarvené."},
		{"sk", "Riadky od A po D sú vyfarbené."},
		{"fi", "Rivit A:sta D:hen on varjostettu."},
		{"sv", "Raderna A till D är skuggade."},
		{"hu", "Az A-tól D-ig tartó sorok színezettek."},
		{"el", "Οι σειρές από A έως D είναι σκιασμένες."},
		{"he", "השורות מ-A עד D צבועות."},
		{"hy", "A-ից մինչև D շարքերը ներկված են։"},
		{"ka", "რიგები A-დან D-მდე შეღებილია."},
		{"sw", "Safu kutoka A hadi D zimetiwa rangi."},
		{"zu", "Imigqa kusuka ku-A kuya ku-D ifakwe umbala."},
		{"sv", "Raderna A till och med D är skuggade."},
		{"da", "Rækkerne A til og med D er skraverede."},
		{"nb", "Radene A til og med D er skravert."},
		{"nl", "De rijen A tot en met D zijn gekleurd."},
		{"id", "Baris A sampai dengan D diwarnai."},
		{"ms", "Baris A sampai dengan D diwarnakan."},
		{"el", "Οι σειρές από A έως και D είναι σκιασμένες."},
		{"", "Raderna A till och med D är skuggade."},
		{"en", "Rows A-D are shaded."},
		{"en", "Rows A\u2010D are shaded."},
		{"en", "Rows A…D are shaded."},
		{"en", "Rows A～D are shaded."},
	} {
		t.Run(test.lesson+" "+test.question, func(t *testing.T) {
			t.Parallel()

			if problems := matchedIn(test.lesson, test.question, labelled("A", "B", "C", "D")); len(problems) != 0 {
				t.Errorf("PictureMatch() = %v in %s, want B and C named by the range", problems, test.lesson)
			}
		})
	}
}

// The Greek semicolon, a raised dot that a keyboard writes as the middle dot,
// and the grave accent as a mark of its own, written after its letter.
const (
	greekAnoTeleia = "\u0387"
	combiningGrave = "\u0300"
)

// listedInEachLanguage are two ends a wording lists in each language of the
// cards, with its word for "and", "or" or "with": written apart, joined to a
// label by a hyphen, or by a case ending.
var listedInEachLanguage = []struct{ lesson, question string }{
	{"am", "ረድፎች A እና D ቀለም ተቀብተዋል።"},
	{"ar", "الصفان A و D ملونان."},
	{"az", "A və D sıraları rənglənib."},
	{"bg", "Редовете A и D са оцветени."},
	{"bn", "সারি A ও D রং করা।"},
	{"bs", "Redovi A i D su obojeni."},
	{"ca", "Les files A i D estan pintades."},
	{"ckb", "ڕیزەکانی A و D ڕەنگکراون."},
	{"cs", "Řady A a D jsou vybarvené."},
	{"da", "Rækkerne A og D er skraverede."},
	{"de", "Die Reihen A und D sind bemalt."},
	{"el", "Οι σειρές A και D είναι σκιασμένες."},
	{"en", "Rows A and D are shaded."},
	{"es", "Las filas A y D están pintadas."},
	{"fa", "ردیف A یا D رنگی است."},
	{"fi", "Rivit A ja D on varjostettu."},
	{"fi", "Rivien A:n ja D:n ruudut on varjostettu."},
	{"fil", "Ang mga hanay A at D ay may kulay."},
	{"fr", "Les rangées A et D sont coloriées."},
	{"gu", "હરોળ A અને D રંગેલી છે."},
	{"ha", "Layukan A da D an yi musu launi."},
	{"he", "השורות A ו-D צבועות."},
	{"hi", "पंक्तियाँ A और D रंगी हैं।"},
	{"hr", "Redovi A i D su obojeni."},
	{"hu", "Az A és D sor színezett."},
	{"hy", "A և D շարքերը ներկված են։"},
	{"id", "Baris A dan D diwarnai."},
	{"ig", "Ahịrị A na D ka e tere agba."},
	{"it", "Le righe A e D sono colorate."},
	{"ja", "AとDの行に色がある。"},
	{"ka", "რიგები A და D შეღებილია."},
	{"kk", "A және D қатарлары боялған."},
	{"kn", "ಸಾಲುಗಳು A ಮತ್ತು D ಬಣ್ಣ ಹಚ್ಚಲಾಗಿದೆ."},
	{"ko", "A와 D 행은 칠해져 있다."},
	{"ku", "Rêzên A û D rengkirî ne."},
	{"lt", "Eilutės A ir D nuspalvintos."},
	{"mai", "पाँती A आ D रंगल अछि।"},
	{"ml", "വരികൾ A-യും D-യും നിറം നൽകിയവയാണ്."},
	{"mr", "रांगा A आणि D रंगवल्या आहेत."},
	{"ms", "Baris A dan D diwarnakan."},
	{"ms", "باريس A دان D دوارناكن."},
	{"nb", "Radene A og D er skravert."},
	{"nl", "De rijen A en D zijn gekleurd."},
	{"no", "Radene A og D er skravert."},
	{"or", "ଧାଡ଼ି A ଓ D ରଙ୍ଗ କରାଯାଇଛି।"},
	{"pa", "ਕਤਾਰਾਂ A ਅਤੇ D ਰੰਗੀਆਂ ਹਨ।"},
	{"pa-PK", "قطاراں A اتے D رنگیاں ہن۔"},
	{"pl", "Rzędy A i D są pomalowane."},
	{"ps", "قطارونه A او D رنګ شوي دي."},
	{"pt", "As linhas A e D estão pintadas."},
	{"ro", "Rândurile A și D sunt colorate."},
	{"ru", "Ряды A и D закрашены."},
	{"sk", "Riadky A a D sú vyfarbené."},
	{"sr", "Редови A и D су обојени."},
	{"sr", "Redovi A i D su obojeni."},
	{"sv", "Raderna A och D är skuggade."},
	{"sw", "Safu A na D zimetiwa rangi."},
	{"ta", "வரிசைகள் A மற்றும் D வண்ணம் தீட்டப்பட்டுள்ளன."},
	{"te", "వరుసలు A మరియు D రంగు వేయబడ్డాయి."},
	{"th", "แถว A และ D ถูกระบายสี"},
	{"tr", "A ve D satırları boyalı."},
	{"uk", "Ряди A і D зафарбовані."},
	{"ur", "قطار A اور D رنگی ہیں۔"},
	{"uz", "A va D qatorlari bo'yalgan."},
	{"vi", "Hàng A và D được tô màu."},
	{"yo", "Àwọn ìlà A àti D ni a kùn."},
	{"zh-Hans", "A和D两行涂了颜色。"},
	{"zh-Hant", "A與D兩行塗了顏色。"},
	{"zu", "Imigqa A no-D ifakwe umbala."},
}

// Two ends the wording lists — with a word for "and", "or" or "with" in each
// language of the cards, or with a mark that lists or reckons in any — name
// those two and nothing between them.
func TestListedEndsNameNothingBetween(t *testing.T) {
	t.Parallel()

	marked := []struct{ lesson, question string }{
		{"en", "Rows A, D are shaded."},
		{"en", "Rows A/D are shaded."},
		{"ru", "Ряды A/D закрашены."},
		{"en", "A + D = 10."},
		{"am", "ረድፎች A፣ D ቀለም ተቀብተዋል።"},
		{"am", "ረድፎች A፡እና፡D ቀለም ተቀብተዋል።"},
		{"hy", "A՝ D շարքերը ներկված են։"},
		{"el", "Οι σειρές A·D είναι σκιασμένες."},
		{"el", "Οι σειρές A" + greekAnoTeleia + "D είναι σκιασμένες."},
		{"el", "Οι σειρές A" + greekQuestionMark + " D είναι σκιασμένες."},
		{"yo", "Àwọn ìlà A a" + combiningGrave + "ti D ni a kùn."},
	}
	for _, test := range append(slices.Clone(listedInEachLanguage), marked...) {
		t.Run(test.lesson+" "+test.question, func(t *testing.T) {
			t.Parallel()

			problems := matchedIn(test.lesson, test.question, labelled("A", "B", "C", "D"))
			if len(problems) != 1 || !strings.Contains(problems[0].Message, "does not name 2 of the labels") {
				t.Errorf("PictureMatch() = %v in %s, want B and C unnamed", problems, test.lesson)
			}
		})
	}
}

// Every language of the cards lists two ends with words of its own, and a
// lesson in one is read with them alone. A lesson with no list of its own, or
// none at all, is read with the words of every language but those another
// joins a range with: the a of Czech lists there, in a lesson in no language,
// as it does not in one in Spanish.
func TestEachLanguageListsWithItsOwnWords(t *testing.T) {
	t.Parallel()

	listing := checks.ListingLanguages()
	for tag := range languagesOfTheCards {
		if base, _ := language.Make(tag).Base(); !slices.Contains(listing, base.String()) {
			t.Errorf("the cards speak %s, and %s lists with no words of its own", tag, base)
		}
	}
	for _, test := range []struct {
		name, lesson, question string
		named                  bool
	}{
		{"Swedish and, in a lesson in English", "en", "Raderna A och D är skuggade.", true},
		{"Swedish and, in a lesson in no language", "", "Raderna A och D är skuggade.", false},
		{"Czech and, in a lesson in no language", "", "Řady A a D jsou vybarvené.", true},
		{"Spanish to, in a lesson in no language", "", "Las filas de la A a la D están pintadas.", true},
		{"Spanish to, in a lesson in Czech", "cs", "Las filas de la A a la D están pintadas.", false},
		{"Bokmål and, in a lesson in Nynorsk", "nn", "Radene A og D er skravert.", false},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			problems := matchedIn(test.lesson, test.question, labelled("A", "B", "C", "D"))
			if named := len(problems) == 0; named != test.named {
				t.Errorf("PictureMatch() named B and C by a range %v in %q, want %v (problems %v)", named, test.lesson, test.named, problems)
			}
		})
	}
}

// A picture labels in Latin capitals, and a lesson written in Greek names its
// points with Greek ones: a run of them standing apart, each drawn as a Latin
// capital is, names the labels it looks like, so that the wording's Α names a
// picture's A. A Greek word opening with such a capital names nothing, nor
// does a run that holds a capital drawn as no Latin one. The labels the
// wording names are looked for in the picture in Latin capitals alone, as in
// any lesson, so an article opening a sentence after a quote is never taken
// for one. A lesson in any other language reads a Greek capital as the letter
// it is.
func TestAGreekLessonReadsItsCapitalsAsTheLatinTheyLookLike(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		name, lesson, question string
		picture                json.RawMessage
		refused                bool
	}{
		{"the wording's Greek capitals, a picture's Latin ones", "el", "Τα σημεία \u0391 και \u0392 απέχουν 4 cm.",
			labelled("A", "B"), false},
		{"a Greek capital no Latin one is drawn as", "el", "Τα σημεία \u0391 και \u0393 απέχουν 4 cm.", labelled("A", "C"), true},
		{"a Greek capital the wording names is not looked for in the picture", "el", "Τα σημεία \u0391 και \u0392 απέχουν 4 cm.", labelled("A"), false},
		{"an article opening a sentence after a Greek question", "el", "Πόσα cm απέχει το \u0391 από το \u0392; \u0397 Μαρία μετράει.", labelled("A", "B"), false},
		{"and after the Greek question mark", "el", "Πόσα cm απέχει το \u0391 από το \u0392" + greekQuestionMark + " \u0397 Μαρία μετράει.", labelled("A", "B"), false},
		{"a liar quoted in a Greek lesson, its articles opening sentences", "el", "\u039f \u0391 λέει: «\u039f \u0392 είναι ψεύτης.» \u039f Γ λέει ότι ψεύδεται.", labelled("A", "B"), false},
		{"Greek words opening with capitals drawn as Latin ones", "el", "Από το σημείο \u0392 ως το τέλος. Τα μήλα είναι 5.", labelled("A", "B", "T"), true},
		{"a run of Greek capitals one of which is drawn as no Latin one", "el", "Το τρίγωνο \u0391\u0392Γ έχει ορθή γωνία.", labelled("A", "B"), true},
		{"the same in a lesson in Russian", "ru", "Точки \u0391 и \u0392 на расстоянии 4 см.", labelled("A", "B"), true},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			if refused := len(matchedIn(test.lesson, test.question, test.picture)) > 0; refused != test.refused {
				t.Errorf("PictureMatch() refused %v in %s, want %v", refused, test.lesson, test.refused)
			}
		})
	}
}

// A label the wording does not name is counted, never quoted: a label may be
// the text of an option, and the refusal reaches the model before the child
// has answered.
func TestAnUnnamedLabelIsNotQuoted(t *testing.T) {
	t.Parallel()

	problems := matched("How many boxes are empty?", labelled("X", "Y"))
	if len(problems) != 1 {
		t.Fatalf("PictureMatch() = %v, want exactly one problem", problems)
	}
	if message := problems[0].Message; strings.Contains(message, `"X"`) || strings.Contains(message, `"Y"`) ||
		!strings.Contains(message, "does not name 2 of the labels") {
		t.Errorf("message = %q, want the two unnamed labels counted and neither quoted", message)
	}
}

// A capital that could as well be a word is taken for one, and numbers and
// quotations are not read at all: a refusal for a label that was never one
// would cost the child an attempt. The picture here labels P and Q only, so
// any other letter read as a label would be refused.
func TestWordsAreNotTakenForLabels(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		name, question string
		labels         []string
	}{
		{"the English article starting a sentence", "A farmer puts a stone at P and another at Q.", nil},
		{"the article starting a quotation", `Ann says: "A stone lies at P." Ben says: "It lies at Q."`, nil},
		{"the article after a colon", "Look at the line: A stone lies at P, a shell at Q.", nil},
		{"the English I", "Then I put a stone at P and a shell at Q.", nil},
		{"a unit after a number", "A 3 L jug stands at P and a 5 L jug at Q.", nil},
		{"a unit after a degree sign", "It is 30 °C at P and 25 °C at Q.", nil},
		{"a word in capitals", "The stone is NOT at P. It is at Q.", nil},
		{"a word in capitals sharing a letter with a label", "The stone is NOT at P. It is at Q, and O is empty.", []string{"O"}},
		{"a capital joined by a hyphen", "Ann wears a T-shirt at P and makes a U-turn at Q.", nil},
		{"a capital joined by an apostrophe", "Mr O'Neil stands at P, and Q is empty.", nil},
		{"numbers", "P is at 3 and Q is at 7, and 12 is the end.", nil},
		{"a quotation", `Ann says: "Ben is at P." Ben is at Q.`, nil},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			labels := append([]string{"P", "Q"}, test.labels...)
			if problems := matched(test.question, labelled(labels...)); len(problems) != 0 {
				t.Errorf("PictureMatch() = %v, want no problems", problems)
			}
		})
	}
}

// A label the wording names over and over is one thing to fix, and it is
// named once.
func TestALabelIsNamedOnce(t *testing.T) {
	t.Parallel()

	problems := matched("Point D is right of A, and D is left of B.", labelled("A", "B"))
	if len(problems) != 1 || !strings.Contains(problems[0].Message, `task.question names "D", which`) {
		t.Errorf("PictureMatch() = %v, want one problem naming D once", problems)
	}
}

// A task with no picture has nothing to hold to its wording, and neither has
// one whose picture is no object, which the structure check refuses.
func TestNothingIsMatchedWithoutAPicture(t *testing.T) {
	t.Parallel()

	for _, picture := range []json.RawMessage{nil, json.RawMessage(`"A and B"`), json.RawMessage(`["A"]`)} {
		if problems := matched("Point D is left of A.", picture); len(problems) != 0 {
			t.Errorf("PictureMatch(%s) = %v, want no problems", picture, problems)
		}
	}
}

// The wording, the picture and the right option all come from the chat's
// model and are untrusted: whatever they hold, the check ends in problems of
// its own code, at most one for each of the four ways it compares them.
func FuzzPictureMatch(f *testing.F) {
	f.Add("The triangle ABC has a right angle at B.", []byte(`{"kind":"table","rows":[["A"],["B"]]}`), "C", "en")
	f.Add(`Ann says: "A stone lies at P."`, []byte(`{"kind":"clock","time":"4:30"}`), "16:30", "de")
	f.Add("\xff'A-B' 3 L °C", []byte(`{"kind":"row","items":[{"label":"1"},{"skip":true},{"label":"9"}]}`), "9", "")
	f.Add("2,5", []byte(`{"kind":"bars","bars":[{"value":"2,5"}],"notes":["A = 2,5"]}`), "2,5 kg", "ru")
	f.Add("Сколько флажков с красной полосой?", []byte(`{"kind":"flags","colors":{"red":"красная","blue":"синяя"},`+
		`"groups":[{"label":"Q","flags":[["red","blue"],["blue","?"]]}]}`), "2", "ru")
	f.Add("ёЁ", []byte(`{"kind":"flags","colors":{"red":"Ё"},"groups":[{"flags":[["red","?"]]}]}`), "1", "")
	f.Add("Řady A a D jsou vybarvené.", []byte(`{"kind":"table","rows":[["A"],["B"],["C"],["D"]]}`), "B", "cs")
	f.Add("השורות A ו-D צבועות.", []byte(`{"kind":"table","rows":[["A"],["D"]]}`), "D", "he")
	f.Add("Rivit A:sta D:hen", []byte(`{"kind":"table","rows":[["A"],["B"],["D"]]}`), "B", "fi")
	f.Add("Οι σειρές Α έως Δ·Ε;", []byte(`{"kind":"table","rows":[["Α"],["B"],["Δ"]]}`), "Β", "el")

	f.Fuzz(func(t *testing.T, question string, picture []byte, right, lesson string) {
		task := &checks.Task{Question: question, Picture: picture, Options: map[string]string{"A": right},
			CorrectAnswer: "A"}
		problems := checks.PictureMatch(task, lesson)
		if len(problems) > 4 {
			t.Fatalf("PictureMatch() = %d problems, want at most four", len(problems))
		}
		for _, problem := range problems {
			if problem.Code != checks.CodeDrawingMismatch || problem.Message == "" {
				t.Fatalf("problem = %+v, want a message under %q", problem, checks.CodeDrawingMismatch)
			}
		}
	})
}

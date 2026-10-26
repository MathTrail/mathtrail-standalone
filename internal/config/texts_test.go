package config_test

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/MathTrail/mathtrail-standalone/internal/config"
)

// inWords is how a text writes a number it spells out, by its language and
// the number: a limit set to another needs its word here before its texts can
// be held to it.
var inWords = map[string]map[int]string{
	"en":      {10: "ten", 50: "fifty"},
	"ru":      {10: "десяти"},
	"es":      {10: "diez"},
	"pt":      {10: "dez"},
	"de":      {10: "zehn"},
	"fr":      {10: "dix"},
	"ja":      {10: "10"},
	"zh-Hans": {10: "十"},
	"ar":      {10: "عشرة"},
	"hi":      {10: "दस"},
}

// The texts a family, a reviewer and a maintainer read say the limits of a
// day the service sets: the tasks a child may be given, and the requests that
// may end with no task written. A limit set to another number names every
// text that still says the old one, and a help page in a language the test
// does not know is named too, so that its limits are held as well.
func TestTheTextsSayTheDailyLimitsTheServiceSets(t *testing.T) {
	t.Parallel()

	tasks, failed := config.DefaultDailyTasks, config.DefaultDailyFailed
	word := func(language string, n int) string {
		t.Helper()
		said, ok := inWords[language][n]
		if !ok {
			t.Fatalf("inWords has no word for %d in %s: write how its texts say it", n, language)
		}
		return said
	}
	helps := map[string][]string{
		"en": {
			fmt.Sprintf("up to %d new tasks a day", tasks),
			fmt.Sprintf("after %s requests in which no task could be written", word("en", failed)),
		},
		"ru": {
			fmt.Sprintf("до %d новых задач в день", tasks),
			fmt.Sprintf("после %s запросов, в которых задачу так и не удалось написать", word("ru", failed)),
		},
		"es": {
			fmt.Sprintf("Cada niño o niña recibe hasta %d ", tasks),
			fmt.Sprintf("tras %s peticiones en las que no se pudo escribir ningún problema", word("es", failed)),
		},
		"pt": {
			fmt.Sprintf("Cada criança recebe até %d ", tasks),
			fmt.Sprintf("depois de %s pedidos em que nenhum problema pôde ser escrito", word("pt", failed)),
		},
		"de": {
			fmt.Sprintf("bis zu %d neue Aufgaben am Tag", tasks),
			fmt.Sprintf("nach %s Anfragen vorbei, bei denen keine Aufgabe geschrieben werden konnte", word("de", failed)),
		},
		"fr": {
			fmt.Sprintf("jusqu’à %d nouveaux problèmes par jour", tasks),
			fmt.Sprintf("la journée se ferme aussi après %s", word("fr", failed)),
		},
		"ja": {
			fmt.Sprintf("一日%d問まで", tasks),
			fmt.Sprintf("問題を書けなかったリクエストが%s回に達した", word("ja", failed)),
		},
		"zh-Hans": {
			fmt.Sprintf("每天最多可以得到%d道新题", tasks),
			fmt.Sprintf("如果有%s次请求都没能写出题目", word("zh-Hans", failed)),
		},
		"ar": {
			fmt.Sprintf("ما يصل إلى %d مسألة جديدة في اليوم", tasks),
			fmt.Sprintf("بعد %s طلبات تعذّرت فيها كتابة أي مسألة", word("ar", failed)),
		},
		"hi": {
			fmt.Sprintf("दिन में %d नए सवाल तक", tasks),
			fmt.Sprintf("ऐसे %s अनुरोधों के बाद भी दिन बंद हो जाता है", word("hi", failed)),
		},
	}
	texts := map[string][]string{
		"plugin/plugin.json": {
			fmt.Sprintf("Up to %d new tasks a day.", tasks),
			fmt.Sprintf("До %d новых задач в день.", tasks),
		},
		"SPEC.md": {
			fmt.Sprintf("| Accepted tasks per day | %d |", tasks),
			fmt.Sprintf("%s holds three of them", word("en", tasks)),
			fmt.Sprintf("| Failed generations per day | %d |", failed),
			fmt.Sprintf("| `MATHTRAIL_DAILY_TASKS` | no | %d |", tasks),
			fmt.Sprintf("| `MATHTRAIL_DAILY_FAILED` | no | %d |", failed),
		},
		"docs/self-hosting.md": {
			fmt.Sprintf("| `MATHTRAIL_DAILY_TASKS` | `%d` |", tasks),
			fmt.Sprintf("| `MATHTRAIL_DAILY_FAILED` | `%d` |", failed),
			fmt.Sprintf("Its %d tasks a day are shared by every reviewer", tasks),
		},
		"docs/catalogs.md": {
			fmt.Sprintf("its %d tasks a day (`MATHTRAIL_DAILY_TASKS`)", tasks),
		},
		"docs/reviewers.md": {
			fmt.Sprintf("**%d new tasks a day** for the demo child", tasks),
		},
		"docs/listing.md": {
			fmt.Sprintf("Limits. Up to %d new tasks a day.", tasks),
			fmt.Sprintf("The demo child has %d new tasks a day", tasks),
		},
		"docs/grants/data-inventory.md": {
			fmt.Sprintf("A child may be given %d tasks a day, and once %d requests of a day", tasks, failed),
		},
		"docs/architecture/03-flows.md": {
			fmt.Sprintf("`MATHTRAIL_DAILY_TASKS`, %s by default", word("en", tasks)),
			fmt.Sprintf("%s in a day by default, `MATHTRAIL_DAILY_FAILED`", word("en", failed)),
		},
		"docs/architecture/04-profile.md": {
			fmt.Sprintf("four days at the most a day gives, %s tasks", word("en", tasks)),
		},
	}

	withHelps(t, texts, helps)
	for file, phrases := range texts {
		wantSaid(t, file, phrases)
	}
}

// withHelps adds to texts the help of every language of the site with the
// phrases helps gives it, naming the help of a language it gives none, and
// the phrases of a language whose help it does not find.
func withHelps(t *testing.T, texts, helps map[string][]string) {
	t.Helper()

	languages, err := os.ReadDir("../../site/content")
	if err != nil {
		t.Fatalf("read the site's languages: %v", err)
	}
	found := map[string]bool{}
	for _, language := range languages {
		help := "site/content/" + language.Name() + "/help.md"
		if _, err := os.Stat("../../" + help); err != nil {
			continue
		}
		found[language.Name()] = true
		phrases, known := helps[language.Name()]
		if !known {
			t.Errorf("%s says the limits of a day in words this test does not know: give its phrases here", help)
			continue
		}
		texts[help] = phrases
	}
	for language := range helps {
		if !found[language] {
			t.Errorf("the help in %s is not where this test reads it, site/content/%s/help.md", language, language)
		}
	}
}

// wantSaid holds the text in file to saying each of the phrases.
func wantSaid(t *testing.T, file string, phrases []string) {
	t.Helper()

	written, err := os.ReadFile("../../" + file)
	if err != nil {
		t.Errorf("read %s: %v", file, err)
		return
	}
	for _, phrase := range phrases {
		if !strings.Contains(string(written), phrase) {
			t.Errorf("%s does not say %q, want the limits of a day as the service sets them", file, phrase)
		}
	}
}

package site_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"go.yaml.in/yaml/v3"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/checks"
)

// card is what the site's data says of a card a page draws, as far as the
// service's checks of a picture read it: the picture, the options and the
// right one, and the picture of the solution with the total under it.
type card struct {
	Picture         json.RawMessage   `json:"picture"`
	Options         map[string]string `json:"options"`
	Correct         string            `json:"correct"`
	SolutionPicture json.RawMessage   `json:"solution_picture"`
	SolutionTotal   string            `json:"solution_total"`
}

// cardPages are the pages that draw a card: the part of the site's data that
// holds the card, and the file of a language's words that holds its question
// and its solution.
var cardPages = []struct{ part, words string }{
	{part: "home", words: "index.yaml"},
	{part: "why", words: "why.yaml"},
}

// cardOf is the card the site's data holds under a part.
func cardOf(t *testing.T, part string) card {
	t.Helper()
	file, err := os.ReadFile("data.json")
	if err != nil {
		t.Fatalf("ReadFile(data.json) error = %v, want the site's data", err)
	}
	var data map[string]json.RawMessage
	if err := json.Unmarshal(file, &data); err != nil {
		t.Fatalf("Unmarshal(data.json) error = %v, want nil", err)
	}
	var holder struct {
		Card *card `json:"card"`
	}
	if err := json.Unmarshal(data[part], &holder); err != nil || holder.Card == nil {
		t.Fatalf("data.json holds no card under %s (error = %v), want the card its page draws", part, err)
	}
	return *holder.Card
}

// said is what a page says of the card it draws, in a language of the site:
// its question and its solution.
type said struct {
	Question string `yaml:"question"`
	Solution string `yaml:"solution"`
}

// saidOf is what a page says of the card it draws, in the words of a language
// of the site.
func saidOf(t *testing.T, language, words string) said {
	t.Helper()
	file, err := os.ReadFile(filepath.Join("content", language, words))
	if err != nil {
		t.Fatalf("ReadFile(%s/%s) error = %v, want the page's words", language, words, err)
	}
	var page struct {
		Task said `yaml:"task"`
	}
	if err := yaml.Unmarshal(file, &page); err != nil {
		t.Fatalf("Unmarshal(%s/%s) error = %v, want nil", language, words, err)
	}
	if page.Task.Question == "" || page.Task.Solution == "" {
		t.Fatalf("%s/%s says no task.question or task.solution, want the question and the solution of its card",
			language, words)
	}
	return page.Task
}

// languagesOf are the languages the site is written in: a folder of words each.
func languagesOf(t *testing.T) []string {
	t.Helper()
	folders, err := os.ReadDir("content")
	if err != nil {
		t.Fatalf("ReadDir(content) error = %v, want the site's words", err)
	}
	var languages []string
	for _, folder := range folders {
		if folder.IsDir() {
			languages = append(languages, folder.Name())
		}
	}
	if len(languages) == 0 {
		t.Fatal("content holds no language, want the site's words")
	}
	return languages
}

// pictureProblems are the problems the service's checks of a picture find in
// a card with what a page says of it, in a language: of its picture, and of
// the picture of its solution with the total under it.
func pictureProblems(held *card, words said, language string) []checks.Problem {
	task := &checks.Task{
		Question:        words.Question,
		Picture:         held.Picture,
		Options:         held.Options,
		CorrectAnswer:   held.Correct,
		Solution:        words.Solution,
		SolutionPicture: held.SolutionPicture,
		SolutionTotal:   held.SolutionTotal,
	}
	return slices.Concat(checks.PictureFormat(task, language), checks.PictureMatch(task, language),
		checks.SolutionPicture(task, language))
}

// A card on a page is a task the service could have handed out, pictures and
// all: its picture is a description of one of the kinds, its labels are the
// question's, and it does not show the answer; the picture of its solution is
// one too, its labels the question's or the solution's, and the total under
// it ends in the answer. The words are the page's own, in each language the
// site is written in, and the pictures are one for every language: a picture
// holds no words.
func TestEveryCardOnTheSiteHasAPictureTheServiceWouldAccept(t *testing.T) {
	t.Parallel()

	for _, page := range cardPages {
		held := cardOf(t, page.part)
		for _, language := range languagesOf(t) {
			t.Run(page.part+"/"+language, func(t *testing.T) {
				t.Parallel()

				if problems := pictureProblems(&held, saidOf(t, language, page.words), language); len(problems) > 0 {
					t.Errorf("the checks of a picture found %v, want none", problems)
				}
				if held.SolutionPicture == nil || held.SolutionTotal == "" {
					t.Error("the card has no picture of its solution or no total under it, want both")
				}
			})
		}
	}
}

// The checks the cards are held to refuse a picture of each kind of mistake
// a card could be changed into, so that a card passing them proves something.
func TestASpoiledCardIsRefusedByTheChecksOfAPicture(t *testing.T) {
	t.Parallel()

	held := cardOf(t, "home")
	words := saidOf(t, "en", "index.yaml")
	for _, spoiled := range []struct {
		name    string
		picture string
		total   string
		code    checks.Code
	}{
		{
			name:    "a member no kind has",
			picture: `{"kind": "row", "items": [{}, {}, {"skip": true}, {}], "posts": 5}`,
			code:    checks.CodeDrawingFormat,
		},
		{
			name:    "a label the question does not name",
			picture: `{"kind": "row", "items": [{"label": "Z"}, {}, {"skip": true}, {}], "gaps": "3"}`,
			code:    checks.CodeDrawingMismatch,
		},
		{
			name:    "the answer, written on the picture",
			picture: `{"kind": "row", "items": [{}, {}, {"skip": true}, {}], "gaps": "3", "span": "` + held.Options[held.Correct] + `"}`,
			code:    checks.CodeDrawingMismatch,
		},
		{
			name:  "a total under the picture of the solution that ends in a wrong option",
			total: "12 ÷ 3 = 4",
			code:  checks.CodeDrawingMismatch,
		},
	} {
		t.Run(spoiled.name, func(t *testing.T) {
			t.Parallel()

			changed := held
			if spoiled.picture != "" {
				changed.Picture = json.RawMessage(spoiled.picture)
			}
			if spoiled.total != "" {
				changed.SolutionTotal = spoiled.total
			}
			problems := pictureProblems(&changed, words, "en")
			if len(problems) == 0 || problems[0].Code != spoiled.code {
				t.Errorf("the checks of a picture found %v, want a refusal of code %s", problems, spoiled.code)
			}
		})
	}
}

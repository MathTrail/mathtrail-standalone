package site_test

import (
	"encoding/json"
	"fmt"
	"maps"
	"os"
	"slices"
	"testing"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/checks"
)

// topicPicture is a picture the site's data draws for a topic: where it
// stands, and its description.
type topicPicture struct {
	where   string
	picture json.RawMessage
}

// topicData is what the site's data says of the topics' drawings: each
// topic's card and first screen, and the drawing beside each of its examples.
type topicData struct {
	Drawings map[string]struct {
		Card json.RawMessage `json:"card"`
		Hero json.RawMessage `json:"hero"`
	} `json:"drawings"`
	Examples map[string][]struct {
		Drawing json.RawMessage `json:"drawing"`
	} `json:"examples"`
}

// pictureIn is the picture a drawing of the site's data is, or nil where the
// drawing is none or a drawing of the site's own. A picture is an object: the
// check of its format has nothing to read in anything else, and passes it.
func pictureIn(t *testing.T, where string, drawing json.RawMessage) json.RawMessage {
	t.Helper()
	if drawing == nil {
		return nil
	}
	var held struct {
		Picture json.RawMessage `json:"picture"`
	}
	if err := json.Unmarshal(drawing, &held); err != nil {
		t.Fatalf("Unmarshal(the drawing of %s) error = %v, want a drawing", where, err)
	}
	if held.Picture == nil {
		return nil
	}
	var members map[string]json.RawMessage
	if err := json.Unmarshal(held.Picture, &members); err != nil || members == nil {
		t.Fatalf("the picture of %s is %s, want a description of one of the kinds", where, held.Picture)
	}
	return held.Picture
}

// topicPicturesOf are the pictures the site's data draws for its topics: their
// cards', their first screens' and their examples', topic by topic.
func topicPicturesOf(t *testing.T) []topicPicture {
	t.Helper()
	file, err := os.ReadFile("data.json")
	if err != nil {
		t.Fatalf("ReadFile(data.json) error = %v, want the site's data", err)
	}
	var data topicData
	if err := json.Unmarshal(file, &data); err != nil {
		t.Fatalf("Unmarshal(data.json) error = %v, want nil", err)
	}
	var pictures []topicPicture
	add := func(where string, drawing json.RawMessage) {
		if picture := pictureIn(t, where, drawing); picture != nil {
			pictures = append(pictures, topicPicture{where: where, picture: picture})
		}
	}
	for _, id := range slices.Sorted(maps.Keys(data.Drawings)) {
		add("the card of "+id, data.Drawings[id].Card)
		add("the first screen of "+id, data.Drawings[id].Hero)
	}
	for _, id := range slices.Sorted(maps.Keys(data.Examples)) {
		for at, example := range data.Examples[id] {
			add(fmt.Sprintf("example %d of %s", at+1, id), example.Drawing)
		}
	}
	if len(pictures) == 0 {
		t.Fatal("data.json draws no picture of a topic, want the pictures of the topics' cards and pages")
	}
	return pictures
}

// Every picture the site draws for a topic is a description the service would
// accept as a task's picture, in each language the site is written in, its
// numbers written with that language's decimal mark. Only its format is held
// to the service's checks, not its match with words: a topic's page does not
// name a picture's labels the way a task's question does, and an example's
// picture may show the answer the example is worked through to.
func TestEveryPictureOfATopicIsInTheFormat(t *testing.T) {
	t.Parallel()

	for _, held := range topicPicturesOf(t) {
		for _, language := range languagesOf(t) {
			t.Run(held.where+"/"+language, func(t *testing.T) {
				t.Parallel()

				task := &checks.Task{Picture: held.picture}
				if problems := checks.PictureFormat(task, language); len(problems) > 0 {
					t.Errorf("the format check found %v, want none", problems)
				}
			})
		}
	}
}

// The format check the topics' pictures are held to refuses a picture of each
// kind of mistake one could be changed into, among them the one the site's
// own reader of a picture lets through, a label in letters other than Latin,
// so that a picture passing it proves something.
func TestASpoiledPictureOfATopicIsRefusedByTheFormatCheck(t *testing.T) {
	t.Parallel()

	for _, spoiled := range []struct {
		name, picture, language string
	}{
		{
			name:     "a member no kind has",
			picture:  `{"kind": "clock", "time": "11:50", "hands": 2}`,
			language: "en",
		},
		{
			name:     "a label in Cyrillic letters",
			picture:  `{"kind": "venn", "sets": [{"label": "А"}, {"label": "B"}], "both": "3"}`,
			language: "ru",
		},
		{
			name:     "a number with a decimal mark its language does not write",
			picture:  `{"kind": "bars", "bars": [{"parts": 4, "value": "2.5"}]}`,
			language: "ru",
		},
	} {
		t.Run(spoiled.name, func(t *testing.T) {
			t.Parallel()

			task := &checks.Task{Picture: json.RawMessage(spoiled.picture)}
			problems := checks.PictureFormat(task, spoiled.language)
			if len(problems) == 0 || problems[0].Code != checks.CodeDrawingFormat {
				t.Errorf("the format check found %v, want a refusal of code %s", problems, checks.CodeDrawingFormat)
			}
		})
	}
}

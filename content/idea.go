package content

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"maps"
	"slices"
	"strings"
	"unicode/utf8"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/profile"
	"github.com/MathTrail/mathtrail-standalone/internal/domain/rating"
)

// The ideas of a topic are the content's own: a list for each level the topic
// is taught at, of a length of its own, each idea the mathematics a task is
// built on rather than its plot, and no two of them one task told twice. A
// package names the one its task is built on, an idea the child has not been
// given in its round, so that the tasks of a topic go through its ideas before
// any comes back — the same list in every chat, where a list the model wrote
// for itself would start again from its favourite idea each time.

const (
	// ideasDir holds one file for each topic, named by the topic's id.
	ideasDir   = "ideas"
	ideaSuffix = ".json"
	// longestIdea is how long an idea may be, in characters: a line that
	// names an idea, not a task.
	longestIdea = 200
	// fewestIdeas is how few ideas a list may hold: every list began with
	// ten, and the files of families name places on them, so a shorter list
	// would leave some of those places with no idea. For the same reason a
	// list that has grown does not shrink back, which the loader cannot see,
	// since it reads a list as it is. The most a list may hold is the
	// furthest place a profile names.
	fewestIdeas = 10
)

// ideaFile is a topic's ideas as its file holds them: the list of each level
// the topic is taught at.
type ideaFile map[rating.GradeLevel][]listedIdea

// listedIdea is one idea of a list: its words, and whether the task built on
// it is solved from the picture. Its file writes an idea in words as a text,
// and one solved from the picture as an object that says so beside its words,
// {"idea": "...", "solved_from_picture": true}, so that the mark stands on the
// line of the idea it marks.
type listedIdea struct {
	text        string
	fromPicture bool
}

// UnmarshalJSON reads an idea written as a text, or as an object that marks it
// solved from the picture. An object that marks nothing, or holds anything
// else, is refused, so that every idea in words is written the one way.
func (i *listedIdea) UnmarshalJSON(raw []byte) error {
	if json.Unmarshal(raw, &i.text) == nil {
		return nil
	}
	var marked struct {
		Idea              string `json:"idea"`
		SolvedFromPicture bool   `json:"solved_from_picture"`
	}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&marked); err != nil {
		return fmt.Errorf("an idea is a text, or an object of idea and solved_from_picture: %w", err)
	}
	if !marked.SolvedFromPicture {
		return errors.New("an idea in words is a text, and an object is an idea marked solved_from_picture: true")
	}
	*i = listedIdea{text: marked.Idea, fromPicture: true}
	return nil
}

// textsOf are the words of the ideas of a list, in its order.
func textsOf(list []listedIdea) []string {
	texts := make([]string, len(list))
	for i, idea := range list {
		texts[i] = idea.text
	}
	return texts
}

// loadIdeas reads the ideas of every topic, a file each, and refuses a list a
// package could not name an idea from: a topic of the catalog without a file,
// a file of no topic, a level the topic is not taught at or one it is taught
// at left out, a list shorter or longer than a list may be, an idea that is
// empty, runs over a line or past its length, or one said twice. Beside the
// ideas come their files, as the version of the instructions counts them: an
// idea is part of what the model is told.
func loadIdeas(src fs.FS, topics []Topic) (map[string]ideaFile, []instructionFile, error) {
	entries, err := fs.ReadDir(src, ideasDir)
	if err != nil {
		return nil, nil, fmt.Errorf("content: read %s: %w", ideasDir, err)
	}
	byID := index(topics, func(t Topic) string { return t.ID })

	ideas := make(map[string]ideaFile, len(topics))
	present := make(map[string]bool, len(entries))
	var (
		files  []instructionFile
		faults []error
	)
	p := &problems{file: ideasDir}
	for _, entry := range entries {
		id, named := strings.CutSuffix(entry.Name(), ideaSuffix)
		topic, known := byID[id]
		if entry.IsDir() || !named || !known {
			p.addf("%s: an idea list is a file named by a topic of the catalog, ending in %s", entry.Name(), ideaSuffix)
			continue
		}
		present[id] = true
		lists, raw, err := loadIdeaFile(src, &topic)
		if err != nil {
			faults = append(faults, err)
			continue
		}
		ideas[id] = lists
		files = append(files, instructionFile{name: ideasDir + "/" + entry.Name(), text: string(raw)})
	}
	for _, topic := range topics {
		if !present[topic.ID] {
			p.addf("topic %q has no file of ideas", topic.ID)
		}
	}
	return ideas, files, errors.Join(append(faults, p.err())...)
}

// loadIdeaFile reads the ideas of one topic and reports everything wrong with
// them at once, together with the bytes they were read from.
func loadIdeaFile(src fs.FS, topic *Topic) (ideaFile, []byte, error) {
	file := ideasDir + "/" + topic.ID + ideaSuffix
	raw, err := fs.ReadFile(src, file)
	if err != nil {
		return nil, nil, fmt.Errorf("content: read %s: %w", file, err)
	}
	var read ideaFile
	if err := decodeBytes(file, raw, &read); err != nil {
		return nil, nil, err
	}

	p := &problems{file: file}
	for _, level := range slices.Sorted(maps.Keys(read)) {
		if !topic.HasLevel(level) {
			p.addf("level %q: %s is not taught at it", level, topic.ID)
		}
	}
	for _, level := range topic.GradeLevels {
		list, listed := read[level]
		if !listed {
			p.addf("level %q: %s is taught at it, and it has no ideas", level, topic.ID)
			continue
		}
		checkIdeas(p, level, textsOf(list))
		if !topic.SolvedFromPicture {
			checkWordIdeas(p, level, list)
		}
	}
	if err := p.err(); err != nil {
		return nil, nil, err
	}
	return read, raw, nil
}

// checkWordIdeas checks that a list of a topic whose tasks are not all solved
// from the picture holds as many ideas in words as a list may hold ideas at
// fewest: a chat not known to draw the card, and the trial series, are given
// those alone, and draw them in rounds of their own.
func checkWordIdeas(p *problems, level rating.GradeLevel, list []listedIdea) {
	words := 0
	for _, idea := range list {
		if !idea.fromPicture {
			words++
		}
	}
	if words < fewestIdeas {
		p.addf("level %q: %d ideas in words, and a list of a topic not all of whose tasks are solved from the "+
			"picture holds at least %d", level, words, fewestIdeas)
	}
}

// checkIdeas checks one level's list: from fewestIdeas to
// profile.MaxIdeasPerList ideas, each a line of its own within the length of
// one, and no two the same once their case and their spaces are set aside.
func checkIdeas(p *problems, level rating.GradeLevel, list []string) {
	if len(list) < fewestIdeas || len(list) > profile.MaxIdeasPerList {
		p.addf("level %q: %d ideas, and a list holds from %d to %d",
			level, len(list), fewestIdeas, profile.MaxIdeasPerList)
	}
	seen := make(map[string]int, len(list))
	for i, idea := range list {
		folded := strings.Join(strings.Fields(strings.ToLower(idea)), " ")
		switch {
		case folded == "":
			p.addf("level %q: idea %d is empty", level, i+1)
		case strings.ContainsAny(idea, "\r\n"):
			p.addf("level %q: idea %d runs over more than a line", level, i+1)
		case utf8.RuneCountInString(idea) > longestIdea:
			p.addf("level %q: idea %d is %d characters long, and an idea is at most %d",
				level, i+1, utf8.RuneCountInString(idea), longestIdea)
		}
		if first, said := seen[folded]; said && folded != "" {
			p.addf("level %q: idea %d says what idea %d says", level, i+1, first)
			continue
		}
		seen[folded] = i + 1
	}
}

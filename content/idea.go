package content

import (
	"errors"
	"fmt"
	"io/fs"
	"maps"
	"slices"
	"strings"
	"unicode/utf8"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/rating"
)

// The ideas of a topic are the content's own: ideasPerList of them for each
// level the topic is taught at, the best known first, each the mathematics a
// task is built on rather than its plot. A package names the one its task is
// built on, after the tasks of the topic the child has left behind, so that
// the tasks of a topic go through its ideas one after another — the same list
// in every chat, where a list the model wrote for itself would start again
// from its favourite idea each time.

const (
	// ideasDir holds one file for each topic, named by the topic's id.
	ideasDir   = "ideas"
	ideaSuffix = ".json"
	// longestIdea is how long an idea may be, in characters: a line that
	// names an idea, not a task.
	longestIdea = 200
)

// ideaFile is a topic's ideas as its file holds them: the list of each level
// the topic is taught at.
type ideaFile map[rating.GradeLevel][]string

// loadIdeas reads the ideas of every topic, a file each, and refuses a list a
// package could not name an idea from: a topic of the catalog without a file,
// a file of no topic, a level the topic is not taught at or one it is taught
// at left out, a list of another length, an idea that is empty, runs over a
// line or past its length, or one said twice. Beside the ideas come their
// files, as the version of the instructions counts them: an idea is part of
// what the model is told.
func loadIdeas(src fs.FS, topics []Topic) (map[string]map[rating.GradeLevel][]string, []instructionFile, error) {
	entries, err := fs.ReadDir(src, ideasDir)
	if err != nil {
		return nil, nil, fmt.Errorf("content: read %s: %w", ideasDir, err)
	}
	byID := index(topics, func(t Topic) string { return t.ID })

	ideas := make(map[string]map[rating.GradeLevel][]string, len(topics))
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
		checkIdeas(p, level, list)
	}
	if err := p.err(); err != nil {
		return nil, nil, err
	}
	return read, raw, nil
}

// checkIdeas checks one level's list: ideasPerList ideas, each a line of its
// own within the length of one, and no two the same once their case and their
// spaces are set aside.
func checkIdeas(p *problems, level rating.GradeLevel, list []string) {
	if len(list) != ideasPerList {
		p.addf("level %q: %d ideas, and a list holds %d", level, len(list), ideasPerList)
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

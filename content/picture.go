package content

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"slices"
	"strings"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/picture"
)

// PictureExample is an example of one kind of picture — a clock, a row of
// posts, a balance — for the model to describe the picture of its own task by,
// rather than from memory: what the kind is for, the topics whose packages
// carry it, and a description of a picture of that kind.
type PictureExample struct {
	// Kind is the kind of picture it is an example of.
	Kind picture.Kind
	// Purpose says what the kind is for and how its members are filled.
	Purpose string
	// Topics are the topics whose packages carry it.
	Topics []string
	// Picture is the example's description, in the format a task's picture is
	// written in.
	Picture json.RawMessage
}

const (
	// picturesDir holds one file for each kind of picture, named by the kind.
	picturesDir   = "pictures"
	pictureSuffix = ".json"
)

// pictureFile is an example of a kind as its file holds it.
type pictureFile struct {
	Purpose string          `json:"purpose"`
	Topics  []string        `json:"topics"`
	Picture json.RawMessage `json:"picture"`
}

// PictureExamples returns the example of every kind of picture, in the order
// the format lists the kinds.
func (c *Content) PictureExamples() []PictureExample {
	examples := make([]PictureExample, 0, len(c.pictures))
	for i := range c.pictures {
		examples = append(examples, c.pictures[i].clone())
	}
	return examples
}

// loadPictures reads the example of every kind of picture, one file each, and
// refuses an example the model could not start from: a kind with no example,
// a file that names no kind, an example that does not say what it is for,
// names no topic or one the catalog does not have, or whose picture is not a
// description of its kind in the format. Beside the examples come their files,
// each under its own path, as the version of the instructions counts them: an
// example is part of what the model is told.
func loadPictures(src fs.FS, topics map[string]Topic) ([]PictureExample, []instructionFile, error) {
	entries, err := fs.ReadDir(src, picturesDir)
	if err != nil {
		return nil, nil, fmt.Errorf("content: read %s: %w", picturesDir, err)
	}

	var (
		examples []PictureExample
		files    []instructionFile
		faults   []error
	)
	p := &problems{file: picturesDir}
	found := map[picture.Kind]bool{}
	for _, entry := range entries {
		name, named := strings.CutSuffix(entry.Name(), pictureSuffix)
		kind := picture.Kind(name)
		if entry.IsDir() || !named || !slices.Contains(picture.Kinds(), kind) {
			p.addf("%s: an example is a file named by the kind of picture it shows, ending in %s",
				entry.Name(), pictureSuffix)
			continue
		}
		example, raw, err := loadPicture(src, kind, topics)
		if err != nil {
			faults = append(faults, err)
			continue
		}
		found[kind] = true
		examples = append(examples, example)
		files = append(files, instructionFile{name: picturesDir + "/" + entry.Name(), text: string(raw)})
	}
	for _, kind := range picture.Kinds() {
		if !found[kind] && !slices.ContainsFunc(entries, func(entry fs.DirEntry) bool {
			return entry.Name() == string(kind)+pictureSuffix
		}) {
			p.addf("there is no example of a %s, and a model shown none describes one from memory", kind)
		}
	}
	slices.SortFunc(examples, func(a, b PictureExample) int {
		return slices.Index(picture.Kinds(), a.Kind) - slices.Index(picture.Kinds(), b.Kind)
	})
	return examples, files, errors.Join(append(faults, p.err())...)
}

// loadPicture reads the example of one kind and reports everything wrong with
// it at once, together with the bytes it was read from.
func loadPicture(src fs.FS, kind picture.Kind, topics map[string]Topic) (PictureExample, []byte, error) {
	file := picturesDir + "/" + string(kind) + pictureSuffix
	raw, err := fs.ReadFile(src, file)
	if err != nil {
		return PictureExample{}, nil, fmt.Errorf("content: read %s: %w", file, err)
	}
	var read pictureFile
	if err := decodeBytes(file, raw, &read); err != nil {
		return PictureExample{}, nil, err
	}

	p := &problems{file: file}
	if strings.TrimSpace(read.Purpose) == "" {
		p.addf("the purpose is empty: say what the kind is for and how its members are filled")
	}
	checkExampleTopics(p, read.Topics, topics)
	if read.Picture == nil {
		p.addf("there is no picture to show the kind by")
	} else if got := picture.KindOf(read.Picture); got != string(kind) {
		p.addf("the picture is a %s, and the file is the example of a %s", got, kind)
	}
	checkPicture(p, "the picture", read.Picture)
	if err := p.err(); err != nil {
		return PictureExample{}, nil, err
	}
	return PictureExample{Kind: kind, Purpose: read.Purpose, Topics: read.Topics, Picture: read.Picture}, raw, nil
}

// checkExampleTopics checks that an example names the topics whose packages
// carry it, each one the catalog has, and each once: an example no package
// carries would never be seen.
func checkExampleTopics(p *problems, named []string, topics map[string]Topic) {
	if len(named) == 0 {
		p.addf("the example names no topic, so no package would carry it")
	}
	seen := make(map[string]bool, len(named))
	for _, topic := range named {
		switch _, known := topics[topic]; {
		case !known:
			p.addf("topic %q is not in the catalog", topic)
		case seen[topic]:
			p.addf("topic %q is named twice", topic)
		}
		seen[topic] = true
	}
}

// clone copies an example together with its topics and its picture, so that
// what a caller is handed shares nothing with the content.
func (e *PictureExample) clone() PictureExample {
	copied := *e
	copied.Topics = slices.Clone(e.Topics)
	copied.Picture = bytes.Clone(e.Picture)
	return copied
}

// packagePicture is an example of a kind of picture as the model is shown it:
// what the kind is for, the limits its pictures are held to, and the example
// itself.
type packagePicture struct {
	Purpose string          `json:"purpose"`
	Limits  string          `json:"limits"`
	Picture json.RawMessage `json:"picture"`
}

// picturesFor are the examples whose topics include this one, in the order of
// the kinds, as a package carries them: a list that is empty rather than null
// for a topic that draws nothing. Each borrows its example's picture instead
// of copying it, as the whole package borrows what it is built from: the
// package is encoded at once, nothing writes to it, and only its encoding
// leaves.
func (c *Content) picturesFor(topic string) []packagePicture {
	examples := make([]packagePicture, 0, len(c.pictures))
	for i := range c.pictures {
		if slices.Contains(c.pictures[i].Topics, topic) {
			examples = append(examples, packagePicture{
				Purpose: c.pictures[i].Purpose, Limits: picture.LimitsOf(c.pictures[i].Kind),
				Picture: c.pictures[i].Picture,
			})
		}
	}
	return examples
}

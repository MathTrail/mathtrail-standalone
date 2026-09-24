package content

import (
	"errors"
	"fmt"
	"io/fs"
	"slices"
	"strings"
)

// Frame is a ready drawing of a subject tasks keep coming back to — a number
// line, a row of posts, a balance — for the model to fill with the numbers and
// labels of its own task rather than draw from memory. It is generic on
// purpose: its labels are there to be kept or renamed, a number goes where it
// has a run of '#', one character to each, and its structure carries no
// values, because it is no task's drawing.
type Frame struct {
	// Name tells it from the other frames.
	Name string
	// Purpose says what it is for and how it is filled.
	Purpose string
	// Topics are the topics whose packages carry it.
	Topics []string
	// Drawing is the frame as the model is shown it, its lines joined by "\n".
	Drawing string
	// Structure describes the drawing as data, the way a task's drawing is
	// described.
	Structure *DrawingStructure
}

const (
	// drawingsDir holds one file for each frame.
	drawingsDir = "drawings"
	frameSuffix = ".json"
)

// frameFile is a frame as its file holds it. The drawing is a list of its
// lines, so that the picture stands in the file as it will be drawn.
type frameFile struct {
	Purpose   string            `json:"purpose"`
	Topics    []string          `json:"topics"`
	Drawing   []string          `json:"drawing"`
	Structure *DrawingStructure `json:"drawing_structure"`
}

// Frames returns every drawing frame, in the order of their names.
func (c *Content) Frames() []Frame {
	frames := make([]Frame, 0, len(c.frames))
	for i := range c.frames {
		frames = append(frames, c.frames[i].clone())
	}
	return frames
}

// loadFrames reads the drawing frames, one file each, and refuses a frame the
// model could not start from: one that does not say what it is for, names no
// topic or one the catalog does not have, draws nothing, or cannot be read as
// data. Beside the frames come their files, each under its own path, as the
// version of the instructions counts them: a frame is part of what the model
// is told.
func loadFrames(src fs.FS, topics map[string]Topic) ([]Frame, []instructionFile, error) {
	entries, err := fs.ReadDir(src, drawingsDir)
	if err != nil {
		return nil, nil, fmt.Errorf("content: read %s: %w", drawingsDir, err)
	}

	var (
		frames []Frame
		files  []instructionFile
		faults []error
	)
	p := &problems{file: drawingsDir}
	for _, entry := range entries {
		name, named := strings.CutSuffix(entry.Name(), frameSuffix)
		if entry.IsDir() || !named || !fileNamePattern.MatchString(name) {
			p.addf("%s: a frame is a file named in lowercase words joined by dashes, ending in %s",
				entry.Name(), frameSuffix)
			continue
		}
		frame, raw, err := loadFrame(src, name, topics)
		if err != nil {
			faults = append(faults, err)
			continue
		}
		frames = append(frames, frame)
		files = append(files, instructionFile{name: drawingsDir + "/" + entry.Name(), text: string(raw)})
	}
	return frames, files, errors.Join(append(faults, p.err())...)
}

// loadFrame reads one frame and reports everything wrong with it at once,
// together with the bytes it was read from.
func loadFrame(src fs.FS, name string, topics map[string]Topic) (Frame, []byte, error) {
	file := drawingsDir + "/" + name + frameSuffix
	raw, err := fs.ReadFile(src, file)
	if err != nil {
		return Frame{}, nil, fmt.Errorf("content: read %s: %w", file, err)
	}
	var read frameFile
	if err := decodeBytes(file, raw, &read); err != nil {
		return Frame{}, nil, err
	}

	p := &problems{file: file}
	if strings.TrimSpace(read.Purpose) == "" {
		p.addf("the purpose is empty: say what the frame is for and how it is filled")
	}
	checkFrameTopics(p, read.Topics, topics)
	checkFrameDrawing(p, read.Drawing)
	checkFrameStructure(p, read.Structure)
	if err := p.err(); err != nil {
		return Frame{}, nil, err
	}

	return Frame{
		Name: name, Purpose: read.Purpose, Topics: read.Topics,
		Drawing: strings.Join(read.Drawing, "\n"), Structure: read.Structure,
	}, raw, nil
}

// checkFrameTopics checks that a frame names the topics whose packages carry
// it, each one the catalog has, and each once: a frame no package carries
// would never be seen.
func checkFrameTopics(p *problems, named []string, topics map[string]Topic) {
	if len(named) == 0 {
		p.addf("the frame names no topic, so no package would carry it")
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

// checkFrameDrawing checks that a frame draws something, and that each of its
// lines is one line: a line break inside one would make the picture in the
// file differ from the picture the model is shown.
func checkFrameDrawing(p *problems, lines []string) {
	if strings.TrimSpace(strings.Join(lines, "")) == "" {
		p.addf("the drawing is empty")
		return
	}
	for i, line := range lines {
		if strings.ContainsAny(line, "\r\n") {
			p.addf("line %d of the drawing breaks in two; the drawing is a list of its lines", i+1)
		}
	}
}

// checkFrameStructure checks that a frame's structure can be read as data and
// holds no values: the values are the numbers of a task, and a frame is the
// drawing before any task has filled it.
func checkFrameStructure(p *problems, structure *DrawingStructure) {
	if structure == nil {
		p.addf("the frame has no drawing_structure, and a drawing cannot be checked against the wording without one")
		return
	}
	checkDrawingStructure(p, "drawing_structure", structure)
	for _, object := range structure.Objects {
		if object.Value != nil {
			p.addf("drawing_structure: drawn object %q has a value, and a frame is no task's drawing", object.ID)
		}
	}
}

// clone copies a frame together with its topics and its structure, so that
// what a caller is handed shares nothing with the content.
func (f *Frame) clone() Frame {
	copied := *f
	copied.Topics = slices.Clone(f.Topics)
	copied.Structure = f.Structure.clone()
	return copied
}

// framesFor are the frames whose topics include this one, in the order of
// their names, as a package carries them: a list that is empty rather than
// null for a topic with none. Each borrows its frame's structure instead of
// copying it, as the whole package borrows what it is built from: the package
// is encoded at once, nothing writes to it, and only its encoding leaves.
func (c *Content) framesFor(topic string) []packageFrame {
	frames := make([]packageFrame, 0, len(c.frames))
	for i := range c.frames {
		if slices.Contains(c.frames[i].Topics, topic) {
			frames = append(frames, packageFrame{
				Purpose: c.frames[i].Purpose, Drawing: c.frames[i].Drawing, Structure: c.frames[i].Structure,
			})
		}
	}
	return frames
}

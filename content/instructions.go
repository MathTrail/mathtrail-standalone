package content

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io/fs"
	"slices"
	"strings"
)

const (
	instructionsDir    = "instructions"
	instructionsSuffix = ".md"
	// The version is long enough to name one wording among all the wordings this
	// service will ever have, and short enough to read in a log line.
	versionLength = 12
)

// instructionFile is one file of instructions for the model: its name, and
// everything it says.
type instructionFile struct {
	name string
	text string
}

// loadInstructions reads every file of instructions and refuses an empty one: a
// file that says nothing still changes the version, which would leave two
// different versions meaning the same thing. A set without the guide is refused
// as well: every package carries it, and a package without it would leave the
// model to guess how a task is written and handed in.
func loadInstructions(src fs.FS) ([]instructionFile, error) {
	entries, err := fs.ReadDir(src, instructionsDir)
	if err != nil {
		return nil, fmt.Errorf("content: read %s: %w", instructionsDir, err)
	}

	p := &problems{file: instructionsDir}
	instructions := make([]instructionFile, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), instructionsSuffix) {
			p.addf("%s: the instructions are files ending in %s, and nothing else", entry.Name(), instructionsSuffix)
			continue
		}
		raw, err := fs.ReadFile(src, instructionsDir+"/"+entry.Name())
		if err != nil {
			return nil, fmt.Errorf("content: read %s/%s: %w", instructionsDir, entry.Name(), err)
		}
		if strings.TrimSpace(string(raw)) == "" {
			p.addf("%s: the file is empty", entry.Name())
			continue
		}
		instructions = append(instructions, instructionFile{name: entry.Name(), text: string(raw)})
	}
	if len(instructions) == 0 {
		p.addf("there are no instructions for the model")
	}
	if !slices.ContainsFunc(entries, func(entry fs.DirEntry) bool { return entry.Name() == guideName }) {
		p.addf("%s: the file is missing, and every package carries it", guideName)
	}
	return instructions, p.err()
}

// instructionsVersion identifies a set of instruction files by their content
// alone. The files are hashed in name order, so the version does not depend on
// the order they were read in; each name and each length goes into the hash
// before the text, so that moving a paragraph from one file to another changes
// the version rather than leaving it where it was.
func instructionsVersion(instructions []instructionFile) string {
	ordered := slices.Clone(instructions)
	slices.SortFunc(ordered, func(a, b instructionFile) int { return strings.Compare(a.name, b.name) })

	sum := sha256.New()
	for _, file := range ordered {
		fmt.Fprintf(sum, "%d:%s\n%d:%s", len(file.name), file.name, len(file.text), file.text)
	}
	return hex.EncodeToString(sum.Sum(nil))[:versionLength]
}

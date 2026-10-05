package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// The page's data file joins the numbers of the page's run with what the
// commit being built adds to them: the commit itself, the facts of the paper
// the page links, and the state of the live numbers. Kept apart from the run,
// the file can be made for every commit from numbers computed once for the
// same build of the bench.

// pageFileCommandName is the word that makes the page's data file.
const pageFileCommandName = "page-file"

// researchSchema is the version of the data file's shape, which its reader
// holds it to.
const researchSchema = 1

// liveComing is the state of live numbers that have not come yet.
const liveComing = "coming"

// researchFile is the page's data file: every number the page shows, from
// one place.
type researchFile struct {
	Schema    int          `json:"schema"`
	BuiltFrom builtFrom    `json:"built_from"`
	Bench     benchBlock   `json:"bench"`
	Product   productBlock `json:"product"`
	Paper     paperBlock   `json:"paper"`
	Live      liveBlock    `json:"live"`
}

// builtFrom is the commit the file was made from, and the commit's own date:
// a wall clock would make every file of the same commit a different one.
type builtFrom struct {
	Commit string `json:"commit"`
	Date   string `json:"date"`
}

// paperBlock is the paper the page names: the commit its numbers were
// computed on, and the files of it the site ships, none until a clean build
// of it is kept.
type paperBlock struct {
	Commit string      `json:"commit"`
	Files  []paperFile `json:"files"`
}

// paperFile is one file of the paper the site ships: its language, where the
// site serves it, and what it is, for the page to show and to vouch for.
type paperFile struct {
	Lang   string `json:"lang"`
	Path   string `json:"path"`
	Pages  int    `json:"pages"`
	Bytes  int64  `json:"bytes"`
	SHA256 string `json:"sha256"`
}

// liveBlock is the state of the live numbers.
type liveBlock struct {
	State string `json:"state"`
}

// The shapes of what the file is made from.
var (
	commitHash = regexp.MustCompile(`^[0-9a-f]{40}([0-9a-f]{24})?$`)
	paperHash  = regexp.MustCompile(`^[0-9a-f]{7,64}$`)
	sha256Hash = regexp.MustCompile(`^[0-9a-f]{64}$`)
	paperPath  = regexp.MustCompile(`^/assets/[a-z0-9.-]+\.pdf$`)
	langTag    = regexp.MustCompile(`^[a-z]{2,3}(-[A-Za-z0-9]{2,8})*$`)
)

// errPaperFacts is a file of the paper's facts the page cannot vouch for.
var errPaperFacts = errors.New("learners page-file: the paper's facts cannot be read")

// pageFileArgs are what the data file is made from: the numbers of the page's
// run and the key of the build they must be of, the commit and its date, and
// the paper's commit and facts.
type pageFileArgs struct {
	numbers, inputs, commit, date, paperCommit, paper, out string
}

// pageFileCommand makes the page's data file, and prints the goals it holds
// as Markdown.
func pageFileCommand(args []string, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("learners page-file", flag.ContinueOnError)
	flags.SetOutput(stderr)
	var a pageFileArgs
	flags.StringVar(&a.numbers, "numbers", "", "the file of the page's numbers")
	flags.StringVar(&a.inputs, "inputs", "", "the key of the build the numbers must be of")
	flags.StringVar(&a.commit, "commit", "", "the commit the file is made from, its full hash")
	flags.StringVar(&a.date, "date", "", "the commit's date, in RFC 3339")
	flags.StringVar(&a.paperCommit, "paper-commit", "", "the commit the paper's numbers were computed on")
	flags.StringVar(&a.paper, "paper", "", "the facts of the paper's files the site ships, if it ships any")
	flags.StringVar(&a.out, "out", "", "the file the data is written to")
	if err := flags.Parse(args); err != nil {
		return 2
	}
	if flags.NArg() > 0 {
		fmt.Fprintf(stderr, "learners page-file: takes flags alone, and was given %q\n", flags.Args())
		return 2
	}
	if a.numbers == "" || a.out == "" {
		fmt.Fprintln(stderr, "learners page-file: -numbers and -out name the files it reads and writes, and both are needed")
		return 2
	}
	file, warning, err := pageFileOf(&a)
	if err == nil {
		err = writeJSON(a.out, file)
	}
	if err != nil {
		fmt.Fprintln(stderr, err)
		if errors.Is(err, errPageFileUsage) {
			return 2
		}
		return 1
	}
	fmt.Fprint(stdout, goalsReport(&file, warning))
	return 0
}

// errPageFileUsage is a flag of a shape no file can be made from, which no
// reading of the numbers would mend.
var errPageFileUsage = errors.New("learners page-file")

// pageFileOf makes the page's data file from what it is given, and says,
// when the paper's PDF is of another commit than the paper's numbers, that it
// is.
func pageFileOf(a *pageFileArgs) (researchFile, string, error) {
	date, err := dateOfArguments(a)
	if err != nil {
		return researchFile{}, "", err
	}
	numbers, err := numbersOf(a)
	if err != nil {
		return researchFile{}, "", err
	}
	paper, warning, err := paperOf(a)
	if err != nil {
		return researchFile{}, "", err
	}
	return researchFile{
		Schema:    researchSchema,
		BuiltFrom: builtFrom{Commit: a.commit, Date: date.UTC().Format(time.RFC3339)},
		Bench:     numbers.Bench, Product: numbers.Product,
		Paper: paper, Live: liveBlock{State: liveComing},
	}, warning, nil
}

// dateOfArguments holds the flags to their shapes, and gives the commit's
// date.
func dateOfArguments(a *pageFileArgs) (time.Time, error) {
	switch {
	case !commitHash.MatchString(a.commit):
		return time.Time{}, fmt.Errorf("%w: -commit is %q, want a commit's full hash", errPageFileUsage, a.commit)
	case !paperHash.MatchString(a.paperCommit):
		return time.Time{}, fmt.Errorf("%w: -paper-commit is %q, want a commit's hash", errPageFileUsage, a.paperCommit)
	case !inputsKey.MatchString(a.inputs):
		return time.Time{}, fmt.Errorf("%w: -inputs is %q, want the key of the build, 64 hexadecimal digits", errPageFileUsage, a.inputs)
	}
	date, err := time.Parse(time.RFC3339, a.date)
	if err != nil {
		return time.Time{}, fmt.Errorf("%w: -date is %q, want the commit's date in RFC 3339", errPageFileUsage, a.date)
	}
	return date, nil
}

// numbersOf reads the numbers of the page's run, refusing numbers of another
// build, and numbers that are not whole, which would otherwise be read as
// zeros.
func numbersOf(a *pageFileArgs) (pageNumbers, error) {
	var numbers pageNumbers
	if err := readStrictJSON(a.numbers, &numbers); err != nil {
		return pageNumbers{}, fmt.Errorf("learners page-file: read the numbers: %w", err)
	}
	if numbers.Bench.Inputs != a.inputs {
		return pageNumbers{}, fmt.Errorf("learners page-file: the numbers are of the build %q, and -inputs names %q: "+
			"numbers of another build are not the page's", numbers.Bench.Inputs, a.inputs)
	}
	if !whole(&numbers) {
		return pageNumbers{}, fmt.Errorf("learners page-file: %s holds not every row of the page's table and the product's counts", a.numbers)
	}
	return numbers, nil
}

// whole says whether numbers hold what a run of the page writes: every row of
// its table, under its rules, and the product's counts.
func whole(n *pageNumbers) bool {
	return n.Bench.Producer == pageProducer && len(n.Bench.Rules) > 0 && len(n.Bench.Rows) == len(pageRows) &&
		n.Product.Topics > 0 && n.Product.Traps > 0 && n.Product.Checks > 0 && n.Product.ReferenceTasks.Total > 0 &&
		n.Product.Options > 0 && n.Product.Corridor.Low < n.Product.Corridor.High
}

// paperOf is the paper the page names, and a word when the PDF the site ships
// is of another commit than the paper's numbers.
func paperOf(a *pageFileArgs) (paperBlock, string, error) {
	if a.paper == "" {
		return paperBlock{Commit: a.paperCommit, Files: []paperFile{}}, "", nil
	}
	paper, err := readPaper(a.paper)
	if err != nil {
		return paperBlock{}, "", err
	}
	warning := ""
	if paper.Commit != a.paperCommit {
		warning = fmt.Sprintf("The PDF on the site is of the paper at %s, and the paper's numbers are now of %s: "+
			"the PDF is due to be built again.", paper.Commit, a.paperCommit)
	}
	return paper, warning, nil
}

// readPaper reads the facts of the paper's files the site ships, refusing
// any it could not vouch for on the page.
func readPaper(path string) (paperBlock, error) {
	var paper paperBlock
	if err := readStrictJSON(path, &paper); err != nil {
		return paperBlock{}, fmt.Errorf("%w: %w", errPaperFacts, err)
	}
	if !paperHash.MatchString(paper.Commit) {
		return paperBlock{}, fmt.Errorf("%w: %s names the commit %q, want a commit's hash", errPaperFacts, path, paper.Commit)
	}
	if len(paper.Files) == 0 {
		return paperBlock{}, fmt.Errorf("%w: %s lists no file", errPaperFacts, path)
	}
	for _, f := range paper.Files {
		if !langTag.MatchString(f.Lang) || !paperPath.MatchString(f.Path) || f.Pages < 1 || f.Bytes < 1 || !sha256Hash.MatchString(f.SHA256) {
			return paperBlock{}, fmt.Errorf("%w: %s has a file of no language, place, pages, size or hash: %+v", errPaperFacts, path, f)
		}
	}
	return paper, nil
}

// readStrictJSON reads one JSON value from a file into v, refusing a field v
// does not have: a file of another shape is not read as if it were this one.
func readStrictJSON(path string, v any) error {
	data, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		return err
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err = decoder.Decode(v); err != nil {
		return fmt.Errorf("%s: %w", path, err)
	}
	if decoder.More() {
		return fmt.Errorf("%s: more than one value", path)
	}
	return nil
}

// goalsReport is the page's table of goals as Markdown: what the run was,
// every row under each rule with its mark, and what the page says of the
// paper.
func goalsReport(f *researchFile, warning string) string {
	var b strings.Builder
	bench := &f.Bench
	fmt.Fprintf(&b, "# The student model on the page \"Research\"\n\n")
	rules := make([]string, 0, len(bench.Rules))
	for _, r := range bench.Rules {
		rules = append(rules, roleNames[r.Role]+" ("+r.ID+")")
	}
	fmt.Fprintf(&b, "%s: %d children a cell, %d answers each, seed %d, experiment %s. "+
		"Computed by the build %s, on the commit %s. Each number has its %s interval, and a goal's mark is read on it.\n\n",
		strings.Join(rules, ", "), bench.Children, bench.Answers, bench.Seed, bench.Experiment,
		short(bench.Inputs), short(f.BuiltFrom.Commit), percentOf(bench.Interval))
	b.WriteString("| Row | Generator | Measure | Earlier rule | Service | Goal | Ceiling |\n|---|---|---|---|---|---|---|\n")
	for i := range bench.Rows {
		row := &bench.Rows[i]
		measure := row.Metric
		if row.ReadAs == "size" {
			measure += ", its size"
		}
		fmt.Fprintf(&b, "| %s | %s | %s | %s | %s | %s | %s |\n", row.ID, row.Generator, measure,
			valueText(row.Values.Earlier), valueText(row.Values.Service), boundText(row), ceilingText(row.Values.Ceiling))
	}
	if len(f.Paper.Files) == 0 {
		fmt.Fprintf(&b, "\nThe paper's numbers are of the commit %s. The site ships no PDF of it yet.\n", f.Paper.Commit)
	} else {
		langs := make([]string, 0, len(f.Paper.Files))
		for _, file := range f.Paper.Files {
			langs = append(langs, file.Lang)
		}
		fmt.Fprintf(&b, "\nThe site ships the paper's PDF in %s, built at the commit %s.\n", strings.Join(langs, ", "), f.Paper.Commit)
	}
	if warning != "" {
		fmt.Fprintf(&b, "\n**%s**\n", warning)
	}
	return b.String()
}

// roleNames are the parts the page's rules play, as the table's reader knows
// them.
var roleNames = map[string]string{"service": "The service", "earlier": "the rule before it", "ceiling": "the ceiling"}

// short is a hash cut to the length people read it at.
func short(hash string) string { return hash[:min(len(hash), 12)] }

// valueText is a rule's number with its interval, and its mark when it has
// one.
func valueText(v pageValue) string {
	text := number(v.Value) + " [" + number(v.Low) + ", " + number(v.High) + "]"
	if v.Mark != nil {
		text += " " + strings.ReplaceAll(*v.Mark, "_", " ")
	}
	return text
}

// boundText is a row's goal, or a dash for a row shown for context.
func boundText(row *pageRowData) string {
	if row.Bound == nil {
		return noNumber
	}
	side := "at most"
	if row.Better == higher {
		side = "at least"
	}
	text := side + " " + number(row.Bound.Value)
	if row.Bound.Own {
		text += ", read off the service's own"
	}
	return text
}

// ceilingText is a row's ceiling and whose it is, or a dash for none.
func ceilingText(c *pageCeiling) string {
	if c == nil {
		return noNumber
	}
	return c.Of + " " + number(c.Value) + " [" + number(c.Low) + ", " + number(c.High) + "]"
}

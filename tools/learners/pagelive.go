package main

import (
	"cmp"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"slices"
	"strconv"
	"time"

	"github.com/MathTrail/mathtrail-standalone/internal/report"
)

// The live numbers of the page are how the answers of real children came out
// against the chance promised, in the latest month counted whole, as the
// public views of the counts kept for years show them. A snapshot of those
// views is committed once a month, and the page's file takes it in with each
// range named by the numbers it spans, since the page writes every number from
// data and the views name a range in words. A snapshot that shows a cell the
// views' rule hides is refused, so that neither a hand that edits it nor a
// query of the wrong views can put on the page what the rule kept back.

// The states the live numbers are in.
const (
	// liveComing is the state before a month is counted whole.
	liveComing = "coming"
	// liveTooFew is the state of a month whose children were too few for any
	// of its numbers to be shown.
	liveTooFew = "too_few"
	// liveReady is the state of a month whose numbers are shown.
	liveReady = "ready"
)

// liveRule is the rule a cell of the live numbers is shown by: the children
// and the answers it stands on at least, and the multiple its counts are
// rounded to.
type liveRule struct {
	Learners  int `json:"learners"`
	Answers   int `json:"answers"`
	RoundedTo int `json:"rounded_to"`
}

// privacyRule is the rule of the public views the snapshot is read from, as
// their SQL writes it.
var privacyRule = liveRule{Learners: 10, Answers: 30, RoundedTo: 5}

// liveCounts are the children and the answers a cell stands on.
type liveCounts struct {
	Learners int `json:"learners"`
	Answers  int `json:"answers"`
}

// liveWeighed is how a cell's answers came out against the chance promised:
// the chance promised on average, and the share answered right.
type liveWeighed struct {
	PromisedMean float64 `json:"promised_mean"`
	CorrectShare float64 `json:"correct_share"`
}

// liveMargin is how far a cell's answers came out from the chance promised:
// right less the chance, on average, and its standard error, counted by child.
type liveMargin struct {
	CameTrueLessPromised float64 `json:"came_true_less_promised"`
	StandardError        float64 `json:"standard_error"`
}

// liveTotal is a month's answers over every range of chance.
type liveTotal struct {
	liveCounts
	liveWeighed
}

// liveSnapshot is the snapshot as the views give it: the month, its answers
// over every range of chance, and the ranges shown, named as the views name
// them. Before a month is counted whole there is no month, and a month of too
// few children has no total.
type liveSnapshot struct {
	Month   *string          `json:"month"`
	Total   *liveTotal       `json:"total"`
	Chances []snapshotChance `json:"chances"`
	KeptUp  []snapshotKeptUp `json:"kept_up"`
}

// snapshotChance is a range of the chance promised, as the views give it.
type snapshotChance struct {
	Chance string `json:"chance"`
	liveCounts
	liveWeighed
}

// snapshotKeptUp is a range of the child's answers, as the views give it.
type snapshotKeptUp struct {
	AnswersRange string `json:"answers_range"`
	liveCounts
	liveMargin
}

// liveBlock is the live numbers as the page reads them: their state, the rule
// they are shown by, and, once a month is counted whole, its numbers, each
// range by what it spans, from the lowest. What a state has no value for is
// null or empty, so that every state has the same keys.
type liveBlock struct {
	State   string       `json:"state"`
	Rule    liveRule     `json:"rule"`
	Month   *string      `json:"month"`
	Total   *liveTotal   `json:"total"`
	Chances []liveChance `json:"chances"`
	KeptUp  []liveKeptUp `json:"kept_up"`
}

// liveChance is a range of the chance promised, from its lowest chance to its
// highest.
type liveChance struct {
	From float64 `json:"from"`
	To   float64 `json:"to"`
	liveCounts
	liveWeighed
}

// liveKeptUp is a range of the child's answers, from its first answer to its
// last, with no last for the range of every answer past the others.
type liveKeptUp struct {
	First int  `json:"first"`
	Last  *int `json:"last"`
	liveCounts
	liveMargin
}

// answersRange is a range of the child's answers as the service names it:
// its first answer and its last, or its first and a plus for every answer
// past the others.
var answersRange = regexp.MustCompile(`^([1-9]\d{0,8})(?:-([1-9]\d{0,8})|\+)$`)

// errLiveSnapshot is a snapshot of the live numbers the page cannot vouch for.
var errLiveSnapshot = errors.New("learners page-file: the snapshot of the live numbers cannot be taken")

// liveOf is the page's live numbers from the snapshot at path, or those still
// to come when there is none.
func liveOf(path string) (liveBlock, error) {
	if path == "" {
		return shownNothing(liveComing, nil), nil
	}
	var s liveSnapshot
	if err := readWholeJSON(path, &s); err != nil {
		return liveBlock{}, fmt.Errorf("%w: %w", errLiveSnapshot, err)
	}
	live, err := liveBlockOf(&s)
	if err != nil {
		return liveBlock{}, fmt.Errorf("%w: %s: %w", errLiveSnapshot, path, err)
	}
	return live, nil
}

// readWholeJSON reads one JSON value from a file into v, as readStrictJSON
// does, and refuses one that leaves out a key of v's: a number left out would
// be read as nothing, and shown as a number.
func readWholeJSON(path string, v any) error {
	data, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		return err
	}
	return decodeWholeJSON(path, data, v)
}

// decodeWholeJSON decodes data, the bytes of the file at path, into v as
// readWholeJSON reads them. The value read, written again, holds every key of
// v's, and the file's own bytes hold the same only when they left none out.
func decodeWholeJSON(path string, data []byte, v any) error {
	if err := decodeStrictJSON(path, data, v); err != nil {
		return err
	}
	read, err := json.Marshal(v)
	if err != nil {
		return err
	}
	var given, kept any
	if err := json.Unmarshal(data, &given); err != nil {
		return err
	}
	if err := json.Unmarshal(read, &kept); err != nil {
		return err
	}
	if !reflect.DeepEqual(given, kept) {
		return fmt.Errorf("%s does not read back as it is written: a key of its shape is left out, set to null or spelled otherwise", path)
	}
	return nil
}

// liveBlockOf is the live numbers of a snapshot, refused when it shows what
// the views' rule hides or what no month could give.
func liveBlockOf(s *liveSnapshot) (liveBlock, error) {
	ranges := len(s.Chances) > 0 || len(s.KeptUp) > 0
	switch {
	case s.Month == nil && (s.Total != nil || ranges):
		return liveBlock{}, errors.New("it shows numbers of no month")
	case s.Month == nil:
		return shownNothing(liveComing, nil), nil
	}
	if _, err := time.Parse("2006-01", *s.Month); err != nil {
		return liveBlock{}, fmt.Errorf("its month is %q, want a year and a month, such as 2026-11", *s.Month)
	}
	switch {
	case s.Total == nil && ranges:
		return liveBlock{}, errors.New("it shows ranges of a month it shows no total of")
	case s.Total == nil:
		return shownNothing(liveTooFew, s.Month), nil
	}
	month := s.Total.liveCounts
	if err := shownByTheRule("the month", month, month); err != nil {
		return liveBlock{}, err
	}
	if err := weighedWithin("the month", s.Total.liveWeighed, 0, 1); err != nil {
		return liveBlock{}, err
	}
	chances, err := chancesOf(s.Chances, month)
	if err != nil {
		return liveBlock{}, err
	}
	keptUp, err := keptUpOf(s.KeptUp, month)
	if err != nil {
		return liveBlock{}, err
	}
	return liveBlock{State: liveReady, Rule: privacyRule, Month: s.Month, Total: s.Total, Chances: chances, KeptUp: keptUp}, nil
}

// shownNothing is the live numbers of a state that shows none of them.
func shownNothing(state string, month *string) liveBlock {
	return liveBlock{State: state, Rule: privacyRule, Month: month, Chances: []liveChance{}, KeptUp: []liveKeptUp{}}
}

// shownByTheRule refuses a cell the views' rule would not show: one of fewer
// children or answers than the rule needs, one whose counts are not rounded
// as the rule rounds them, or one that counts more than its month.
func shownByTheRule(cell string, c, month liveCounts) error {
	switch {
	case c.Learners < privacyRule.Learners || c.Answers < privacyRule.Answers:
		return fmt.Errorf("%s stands on %d children and %d answers, and the rule shows nothing under %d and %d",
			cell, c.Learners, c.Answers, privacyRule.Learners, privacyRule.Answers)
	case c.Learners%privacyRule.RoundedTo != 0 || c.Answers%privacyRule.RoundedTo != 0:
		return fmt.Errorf("%s counts %d children and %d answers, which are not rounded to %d",
			cell, c.Learners, c.Answers, privacyRule.RoundedTo)
	case c.Learners > month.Learners || c.Answers > month.Answers:
		return fmt.Errorf("%s counts %d children and %d answers, more than its month's %d and %d",
			cell, c.Learners, c.Answers, month.Learners, month.Answers)
	}
	return nil
}

// weighedWithin refuses a chance promised on average outside the chances its
// cell holds, and a share right outside nothing to all.
func weighedWithin(cell string, w liveWeighed, from, to float64) error {
	if w.PromisedMean < from || w.PromisedMean > to {
		return fmt.Errorf("%s promises %v on average, outside the chances from %v to %v it holds", cell, w.PromisedMean, from, to)
	}
	if w.CorrectShare < 0 || w.CorrectShare > 1 {
		return fmt.Errorf("%s has %v of its answers right, which is no share", cell, w.CorrectShare)
	}
	return nil
}

// chancesOf are the ranges of chance a snapshot shows, from the lowest, each
// by the chances it spans. A range the report counts in under no such name,
// one that is there twice, and one the rule would not show are refused.
func chancesOf(shown []snapshotChance, month liveCounts) ([]liveChance, error) {
	ranges := report.ChanceRanges()
	chances := make([]liveChance, 0, len(shown))
	for _, c := range shown {
		at := slices.IndexFunc(ranges, func(r report.ChanceRange) bool { return r.Name == c.Chance })
		if at < 0 {
			return nil, fmt.Errorf("the range of chance %q is none the report counts in", c.Chance)
		}
		from, to := float64(ranges[at].From)/100, float64(ranges[at].To)/100
		if slices.ContainsFunc(chances, func(seen liveChance) bool { return seen.From == from }) {
			return nil, fmt.Errorf("the range of chance %s is there twice", c.Chance)
		}
		cell := "the range of chance " + c.Chance
		if err := shownByTheRule(cell, c.liveCounts, month); err != nil {
			return nil, err
		}
		if err := weighedWithin(cell, c.liveWeighed, from, to); err != nil {
			return nil, err
		}
		chances = append(chances, liveChance{From: from, To: to, liveCounts: c.liveCounts, liveWeighed: c.liveWeighed})
	}
	slices.SortFunc(chances, func(a, b liveChance) int { return cmp.Compare(a.From, b.From) })
	return chances, nil
}

// keptUpOf are the ranges of the child's answers a snapshot shows, from the
// first answer, each by the answers it spans. A range of another name, one
// the rule would not show, one of a margin no answers give, and ranges that
// overlap are refused.
func keptUpOf(shown []snapshotKeptUp, month liveCounts) ([]liveKeptUp, error) {
	keptUp := make([]liveKeptUp, 0, len(shown))
	for _, k := range shown {
		first, last, err := answersRangeOf(k.AnswersRange)
		if err != nil {
			return nil, err
		}
		cell := "the range of answers " + k.AnswersRange
		if err := shownByTheRule(cell, k.liveCounts, month); err != nil {
			return nil, err
		}
		if err := marginWithin(cell, k.liveMargin); err != nil {
			return nil, err
		}
		keptUp = append(keptUp, liveKeptUp{First: first, Last: last, liveCounts: k.liveCounts, liveMargin: k.liveMargin})
	}
	slices.SortFunc(keptUp, func(a, b liveKeptUp) int { return cmp.Compare(a.First, b.First) })
	for i := 1; i < len(keptUp); i++ {
		before := keptUp[i-1]
		if before.Last == nil || *before.Last >= keptUp[i].First {
			return nil, fmt.Errorf("the ranges of answers from %d and from %d overlap", before.First, keptUp[i].First)
		}
	}
	return keptUp, nil
}

// answersRangeOf is the first and the last answer of a range of the child's
// answers, read from its name, with no last for the range of every answer
// past the others. A name of another shape, and a range that ends before it
// begins, are refused.
func answersRangeOf(name string) (first int, last *int, err error) {
	found := answersRange.FindStringSubmatch(name)
	if found == nil {
		return 0, nil, fmt.Errorf("the range of answers %q is not named as the service names one", name)
	}
	// The pattern holds nine digits at most, which every int holds.
	first, _ = strconv.Atoi(found[1])
	if found[2] == "" {
		return first, nil, nil
	}
	end, _ := strconv.Atoi(found[2])
	if end < first {
		return 0, nil, fmt.Errorf("the range of answers %s ends before it begins", name)
	}
	return first, &end, nil
}

// marginWithin refuses a margin of right less the chance past one either way,
// which no answers give, and a standard error under nothing.
func marginWithin(cell string, m liveMargin) error {
	if m.CameTrueLessPromised < -1 || m.CameTrueLessPromised > 1 {
		return fmt.Errorf("%s came out %v from its promise, past one either way", cell, m.CameTrueLessPromised)
	}
	if m.StandardError < 0 {
		return fmt.Errorf("%s has a standard error of %v, under nothing", cell, m.StandardError)
	}
	return nil
}

// liveReport is what the page's table of goals says of the live numbers.
func liveReport(l *liveBlock) string {
	switch l.State {
	case liveComing:
		return "\nThe live numbers are still to come: no snapshot of them names a month counted whole.\n"
	case liveTooFew:
		return fmt.Sprintf("\nThe live numbers of %s show nothing: too few children answered for the rule of %d children and %d answers.\n",
			*l.Month, l.Rule.Learners, l.Rule.Answers)
	}
	return fmt.Sprintf("\nThe live numbers are of %s: %d children and %d answers, with %d of the %d ranges of chance and %d ranges of the child's answers shown.\n",
		*l.Month, l.Total.Learners, l.Total.Answers, len(l.Chances), len(report.ChanceRanges()), len(l.KeptUp))
}

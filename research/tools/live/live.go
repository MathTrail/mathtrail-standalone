package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"reflect"
	"strconv"
	"strings"
	"time"
)

// The states the paper's sentence on real use is written in, the number it
// branches on as \stat{live}{state}.
const (
	// stateComing is before a month is counted whole, or with no snapshot.
	stateComing = 0
	// stateTooFew is a month counted whole whose children were too few for
	// its total to be shown.
	stateTooFew = 1
	// stateShown is a month counted whole that shows its total.
	stateShown = 2
)

// rule is what the public views show a month's total behind: the least
// children and answers it stands on, and the multiple its counts are rounded
// to, as their SQL writes it.
var rule = struct{ Learners, Answers, RoundedTo int }{Learners: 10, Answers: 30, RoundedTo: 5}

// snapshot is the file the site's page "Research" draws its live numbers
// from. The paper prints its month and its total; the ranges are read only to
// know whether it shows any.
type snapshot struct {
	Month   *string           `json:"month"`
	Total   *total            `json:"total"`
	Chances []json.RawMessage `json:"chances"`
	KeptUp  []json.RawMessage `json:"kept_up"`
}

// total is how a month's weighed answers came out: the children and the
// answers behind them, the chance the rule promised on average, and the share
// answered right.
type total struct {
	Learners     *int     `json:"learners"`
	Answers      *int     `json:"answers"`
	PromisedMean *float64 `json:"promised_mean"`
	CorrectShare *float64 `json:"correct_share"`
}

// Numbers are what the paper prints of a snapshot: the state always, the
// month from stateTooFew on, the total at stateShown.
type Numbers struct {
	State        int
	Year         int
	Month        int
	Learners     int
	Answers      int
	PromisedMean float64
	CorrectShare float64
}

// Parse reads a snapshot strictly: every key it holds in every state and no
// other, a month written as a year and a month, and a total only of a month
// and only as the public views would show it. A snapshot edited by hand thus
// cannot put into the paper what the rule keeps back, and one whose shape
// changed is refused rather than read as nothing.
func Parse(data []byte) (Numbers, error) {
	var s snapshot
	if err := decodeWhole(data, &s); err != nil {
		return Numbers{}, err
	}
	ranges := len(s.Chances) > 0 || len(s.KeptUp) > 0
	switch {
	case s.Month == nil && (s.Total != nil || ranges):
		return Numbers{}, errors.New("it shows numbers of no month")
	case s.Month == nil:
		return Numbers{State: stateComing}, nil
	}
	month, err := time.Parse("2006-01", *s.Month)
	if err != nil {
		return Numbers{}, fmt.Errorf("its month is %q, want a year and a month, such as 2026-11", *s.Month)
	}
	n := Numbers{State: stateTooFew, Year: month.Year(), Month: int(month.Month())}
	switch {
	case s.Total == nil && ranges:
		return Numbers{}, errors.New("it shows ranges of a month it shows no total of")
	case s.Total == nil:
		return n, nil
	}
	t := s.Total
	if t.Learners == nil || t.Answers == nil || t.PromisedMean == nil || t.CorrectShare == nil {
		return Numbers{}, errors.New("its total holds a null")
	}
	if err := shownByTheRule(*t.Learners, *t.Answers); err != nil {
		return Numbers{}, err
	}
	if !isShare(*t.PromisedMean) || !isShare(*t.CorrectShare) {
		return Numbers{}, fmt.Errorf("its total promised %v and came right in %v, and both are shares, from 0 to 1", *t.PromisedMean, *t.CorrectShare)
	}
	n.State = stateShown
	n.Learners, n.Answers = *t.Learners, *t.Answers
	n.PromisedMean, n.CorrectShare = *t.PromisedMean, *t.CorrectShare
	return n, nil
}

// shownByTheRule refuses a total the public views would not show: one behind
// fewer children or answers than the rule needs, or one whose counts are not
// rounded as the rule rounds them.
func shownByTheRule(learners, answers int) error {
	switch {
	case learners < rule.Learners || answers < rule.Answers:
		return fmt.Errorf("its total stands on %d children and %d answers, and the public views show a month only behind %d children and %d answers or more",
			learners, answers, rule.Learners, rule.Answers)
	case learners%rule.RoundedTo != 0 || answers%rule.RoundedTo != 0:
		return fmt.Errorf("its total counts %d children and %d answers, and the public views round every count to %d",
			learners, answers, rule.RoundedTo)
	}
	return nil
}

// isShare is whether a value can be a share of answers.
func isShare(v float64) bool {
	return v >= 0 && v <= 1
}

// decodeWhole decodes one JSON value into s, refusing a key s does not know,
// a key of s's left out, which would read as nothing, and anything after the
// value. A key left out is found by writing the value read back: it then holds
// that key, set to null, where the file's own bytes do not.
func decodeWhole(data []byte, s *snapshot) error {
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	if err := dec.Decode(s); err != nil {
		return fmt.Errorf("read the snapshot: %w", err)
	}
	if err := dec.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return errors.New("read the snapshot: it holds more than one value")
	}
	read, err := json.Marshal(s)
	if err != nil {
		return fmt.Errorf("read the snapshot: %w", err)
	}
	var given, kept any
	if err := json.Unmarshal(data, &given); err != nil {
		return fmt.Errorf("read the snapshot: %w", err)
	}
	if err := json.Unmarshal(read, &kept); err != nil {
		return fmt.Errorf("read the snapshot: %w", err)
	}
	if !reflect.DeepEqual(given, kept) {
		return errors.New("read the snapshot: a key of its shape is left out")
	}
	return nil
}

// Render writes the numbers as the key=value lines the paper's macros read:
// the state always, the month once there is one, the total once it is shown.
func Render(n Numbers) string {
	var b strings.Builder
	b.WriteString("# The service's real use, as its latest monthly snapshot of public totals shows it.\n")
	line := func(key, value string) { fmt.Fprintf(&b, "%s=%s\n", key, value) }
	line("state", strconv.Itoa(n.State))
	if n.State == stateComing {
		return b.String()
	}
	line("year", strconv.Itoa(n.Year))
	line("month", strconv.Itoa(n.Month))
	if n.State == stateTooFew {
		return b.String()
	}
	line("learners", strconv.Itoa(n.Learners))
	line("answers", strconv.Itoa(n.Answers))
	line("promised_mean", strconv.FormatFloat(n.PromisedMean, 'f', -1, 64))
	line("correct_share", strconv.FormatFloat(n.CorrectShare, 'f', -1, 64))
	return b.String()
}

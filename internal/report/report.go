// Package report adds up the lines the service writes into the numbers they
// are written for: how many tasks were asked for, handed in, accepted and
// refused, at which attempt and why, how long the chat's model took to write
// one, whether the chance of a right answer a task was handed out at came
// true, and whether the estimate of a child keeps up as the answers pile up,
// which limits were reached, and how the tools answered — by the version of
// the instructions a task was written to and by the chat host that called.
//
// It reads the service's own lines, one JSON object each: a log of a local run
// as the service wrote it, or the payloads of the entries Cloud Logging keeps,
// with what Cloud Logging moved out of them put back. The report is Markdown.
package report

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"time"
)

// Run reads the lines from in, adds them up and writes the report to out.
func Run(in io.Reader, out io.Writer) error {
	read, err := readAll(in)
	if err != nil {
		return err
	}
	if err := write(out, tally(read)); err != nil {
		return fmt.Errorf("report: write the report: %w", err)
	}
	return nil
}

// line is as much of one line of the service's log as the report reads. Each
// event fills the fields it has and leaves the others empty.
type line struct {
	Time                time.Time `json:"time"`
	Message             string    `json:"message"`
	RequestID           string    `json:"request_id"`
	SpanID              string    `json:"logging.googleapis.com/spanId"`
	InstructionsVersion string    `json:"instructions_version"`

	// Outcome is how a tool call ended, or how an attempt at a task did.
	Outcome string `json:"outcome"`

	// A tool call: which tool, which chat host, and how long it took.
	Tool       string `json:"tool"`
	Client     string `json:"client"`
	DurationMS whole  `json:"duration_ms"`

	// A request for a task, whether it was one already open.
	AlreadyOpen bool `json:"already_open"`

	// An attempt at a task: which of a request's attempts it was, the check it
	// is counted by, and every check it failed.
	Attempt whole    `json:"attempt"`
	Primary string   `json:"primary"`
	Failed  []string `json:"failed"`

	// An accepted task: the attempts it took, and the seconds from the request
	// to the task.
	Attempts            whole `json:"attempts"`
	SecondsSinceRequest whole `json:"seconds_since_request"`

	// A limit reached.
	Limit string `json:"limit"`

	// An answer: the child's account, whether it was right and came after the
	// hint, the chance its task was handed out at, who chose the task, which
	// answer of the trial series it was, and, after the series, which of the
	// child's answers it was, as a range.
	User          string   `json:"user"`
	Correct       bool     `json:"correct"`
	HintUsed      bool     `json:"hint_used"`
	Chance        *float64 `json:"chance"`
	TutorMode     string   `json:"tutor_mode"`
	Trial         whole    `json:"trial"`
	AnswersBucket string   `json:"answers_bucket"`
}

// whole is a whole number as a line carries it. The service writes 2, and
// Cloud Logging, which keeps every number of a line as a double, gives it back
// as 2.0.
type whole int64

func (w *whole) UnmarshalJSON(raw []byte) error {
	var read float64
	if err := json.Unmarshal(raw, &read); err != nil {
		return err
	}
	*w = whole(read)
	return nil
}

// call is which tool call a line was written in: the span of the call, which
// every line of the call names, or, where a line names no span — the service
// names one only when it knows its project — the request the call came in.
// The span is the service's own, so two calls never share it; a request's id
// may be the one its client sent, and a client may send the same one twice.
func (l *line) call() string {
	if l.SpanID != "" {
		return l.SpanID
	}
	return l.RequestID
}

// input is what the report was given: the lines of the service's it read, how
// many other lines there were — the runtime's own words when a process ends,
// say, or a line cut short — and how many lines of the service's it could not
// read, which the report owns up to rather than leaving out as another's.
type input struct {
	lines              []line
	others, unreadable int
}

// readAll reads the input a line at a time, to its end.
func readAll(in io.Reader) (*input, error) {
	read := &input{}
	reader := bufio.NewReader(in)
	for {
		text, err := reader.ReadBytes('\n')
		if text = bytes.TrimSpace(text); len(text) > 0 {
			read.add(text)
		}
		switch {
		case errors.Is(err, io.EOF):
			return read, nil
		case err != nil:
			return nil, fmt.Errorf("report: read the log: %w", err)
		}
	}
}

// add takes in one line of the input: a line of the service's, one of the
// service's that does not read as the report expects, or another's.
func (in *input) add(text []byte) {
	var read line
	if err := json.Unmarshal(text, &read); err == nil && read.Message != "" {
		in.lines = append(in.lines, read)
		return
	}
	var named struct {
		Message string `json:"message"`
	}
	if json.Unmarshal(text, &named) == nil && named.Message != "" {
		in.unreadable++
		return
	}
	in.others++
}

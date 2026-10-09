// Package report adds up the lines the service writes into the numbers they
// are written for: how many tasks were asked for, handed in, accepted and
// refused, at which attempt and why, how long the chat's model took to write
// one, whether the chance of a right answer a task was handed out at came
// true, whether the estimate of a child keeps up as the answers pile up — the
// same child's later answers set against its earlier ones —, how soon a topic
// a child was shown as mastered is taken back, which limits were reached, how
// the tools answered and how long they took beside their calls to Drive, what
// became of the traces, and how busy the busiest minute was — by the version
// of the instructions a task was written to and by the chat host that called.
// It also holds every line to the rules of the log: an event the service is
// decided to write, the fields decided for it, and nothing shaped like an
// email address.
//
// It reads the service's own lines, one JSON object each: a log of a local run
// as the service wrote it, or the payloads of the entries Cloud Logging keeps,
// with what Cloud Logging moved out of them put back and the instance that
// wrote each beside it. The report is Markdown.
package report

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"slices"
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

	// The trace a line was written in, and whether it was kept, which the
	// service names only when it knows its project; the instance that wrote
	// the line, which the reader of the platform's log names.
	Trace    string `json:"logging.googleapis.com/trace"`
	Sampled  *bool  `json:"logging.googleapis.com/trace_sampled"`
	Instance string `json:"instance"`

	// A request's line: the route it was served by, and how long it took, in
	// seconds.
	Route    string  `json:"route"`
	Duration float64 `json:"duration"`

	// A failure: what failed, in the words of whatever failed.
	Error string `json:"error"`

	// The telemetry as it was built: whether it exports, and the share of the
	// traces it keeps.
	Export      bool     `json:"export"`
	SampleRatio *float64 `json:"sample_ratio"`

	// Outcome is how a tool call ended, or how an attempt at a task did.
	Outcome string `json:"outcome"`

	// A tool call: which tool, which chat host, how long it took, the screen
	// its answer drew and the task's request it was a call of, which a line
	// written before the service named them leaves unsaid.
	Tool        string `json:"tool"`
	Client      string `json:"client"`
	DurationMS  whole  `json:"duration_ms"`
	Screen      string `json:"screen"`
	TaskRequest string `json:"task_request"`

	// A request for a task, whether it was one already open, and whether it
	// was for a task written ahead.
	AlreadyOpen bool `json:"already_open"`
	Ahead       bool `json:"ahead"`

	// An attempt at a task: which of a request's attempts it was, the check it
	// is counted by, and every check it failed.
	Attempt whole    `json:"attempt"`
	Primary string   `json:"primary"`
	Failed  []string `json:"failed"`
	// A hand-in: how large it was, part by part, in bytes, which a line written
	// before the service measured them leaves unsaid; what the format no longer
	// reads that came all the same; and the fields read as they were meant.
	TaskBytes      *whole   `json:"task_bytes"`
	SelfCheckBytes *whole   `json:"self_check_bytes"`
	SolverBytes    *whole   `json:"solver_bytes"`
	CoreIdeaBytes  *whole   `json:"core_idea_bytes"`
	TotalBytes     *whole   `json:"total_bytes"`
	Retired        []string `json:"retired"`
	Mended         []string `json:"mended"`

	// An accepted task: the attempts it took, the seconds from the request to
	// the task, and the kind of its picture, or none. A line written before
	// pictures says only whether the task came with a text drawing, and one
	// written before that says neither.
	Attempts            whole  `json:"attempts"`
	SecondsSinceRequest whole  `json:"seconds_since_request"`
	Picture             string `json:"picture"`
	Drawing             *bool  `json:"drawing"`
	// A task handed out: whether it was written ahead and kept, and whether a
	// card took it into its own place, which only a line from a release that let
	// a card do so says.
	Ready  bool `json:"ready"`
	ByCard bool `json:"by_card"`
	// A task written ahead and let go: why, and whether it had been written
	// and kept or was still being written.
	Reason  string `json:"reason"`
	Written bool   `json:"written"`

	// A limit reached.
	Limit string `json:"limit"`

	// An answer: the child's account, the topic of its task, whether it was
	// right and came after the hint, the chance its task was handed out at,
	// who chose the task, which answer of the trial series it was, and, after
	// the series, which of the child's answers it was, as a range. A topic
	// mastered names the topic too.
	User          string   `json:"user"`
	Topic         string   `json:"topic"`
	Correct       bool     `json:"correct"`
	HintUsed      bool     `json:"hint_used"`
	Chance        *float64 `json:"chance"`
	TutorMode     string   `json:"tutor_mode"`
	Trial         whole    `json:"trial"`
	AnswersBucket string   `json:"answers_bucket"`

	// A line a child is counted from: the name the child is counted under
	// that month, and, on a task handed out, the chat host it was handed out
	// in. Whether the log repeated a line the report has read already is the
	// report's own note, and no field of the line.
	Learner string `json:"learner"`
	Host    string `json:"host"`
	repeat  bool
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

// request is which request a line was written in, whatever span it names: the
// request's id and the trace it was in together, empty when the line names
// neither. A call to Drive has a span of its own inside the tool call's, and is
// tied to the call through the request. Either alone could be another
// request's as well: a client may send the same id twice, and every request
// it sends with one parent joins the same trace.
func (l *line) request() string {
	if l.RequestID == "" && l.Trace == "" {
		return ""
	}
	return l.RequestID + " " + l.Trace
}

// input is what the report was given: the lines of the service's it read, how
// many other lines there were — the runtime's own words when a process ends,
// say, or a line cut short — and how many lines of the service's it could not
// read, which the report owns up to rather than leaving out as another's; and
// every way the lines of the service's broke the rules of the log.
type input struct {
	lines              []line
	others, unreadable int
	breaches           map[breach]int
	// counted are the lines a child is counted from, each by a digest of its
	// text: the log may hand one over twice, and a repeat is the same text
	// again.
	counted map[[sha256.Size]byte]struct{}
}

// readAll reads the input a line at a time, to its end.
func readAll(in io.Reader) (*input, error) {
	read := &input{breaches: map[breach]int{}, counted: map[[sha256.Size]byte]struct{}{}}
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
// service's that does not read as the report expects, or another's. Every line
// of the service's is held to the rules of the log, read as the report
// expects it or not.
func (in *input) add(text []byte) {
	var fields map[string]any
	if json.Unmarshal(text, &fields) != nil {
		in.others++
		return
	}
	if message, isText := fields["message"].(string); !isText || message == "" {
		in.others++
		return
	}
	for _, broken := range audit(fields) {
		in.breaches[broken]++
	}
	var read line
	if json.Unmarshal(text, &read) != nil {
		in.unreadable++
		return
	}
	if slices.Contains(countedFrom, read.Message) {
		digest := sha256.Sum256(text)
		_, read.repeat = in.counted[digest]
		in.counted[digest] = struct{}{}
	}
	in.lines = append(in.lines, read)
}

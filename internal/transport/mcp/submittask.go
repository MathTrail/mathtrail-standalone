package mcpserver

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.uber.org/zap"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/checks"
	"github.com/MathTrail/mathtrail-standalone/internal/domain/profile"
	"github.com/MathTrail/mathtrail-standalone/internal/domain/solver"
	"github.com/MathTrail/mathtrail-standalone/internal/store"
)

// The refusals of submit_task that are about the request rather than the task.
const (
	// codeStaleRequest is a task handed in for no request that is open. It
	// spends no attempt: there is no request to spend one of.
	codeStaleRequest = "stale_request"
	// codeAttemptsExhausted is the refusal that spent the last attempt and
	// closed the request.
	codeAttemptsExhausted = "attempts_exhausted"
)

// submitTaskIn is what submit_task takes. The three parts of the task are
// taken as whatever JSON arrived and read by the checks, in words of their
// own: the library that holds arguments to a schema quotes back what broke it,
// and the answer of this tool is drawn on a card.
type submitTaskIn struct {
	RequestID string `json:"request_id" jsonschema:"the id of the request next_task opened"`
	Brief     any    `json:"brief" jsonschema:"the brief of the package as you received it, as a JSON object: you may change its setting, traps_to_use and constraints, and say why in its rationale"`
	Task      any    `json:"task" jsonschema:"the task as a JSON object, written as the guide in the package describes"`
	Solver    string `json:"solver" jsonschema:"the Starlark program that proves the answer, written as the guide in the package describes"`
	SelfCheck any    `json:"self_check" jsonschema:"your own check of the task as a JSON object, written as the guide in the package describes"`
}

// submission is the task as the checks read it: its three parts as the JSON
// they arrived as, and the program.
func (in *submitTaskIn) submission() (*checks.Submission, error) {
	parts := make([]json.RawMessage, 0, 3)
	for _, part := range []any{in.Brief, in.Task, in.SelfCheck} {
		raw, err := partJSON(part)
		if err != nil {
			return nil, fmt.Errorf("mcp: read the task: %w", err)
		}
		parts = append(parts, raw)
	}
	return &checks.Submission{Brief: parts[0], Task: parts[1], SelfCheck: parts[2], Solver: in.Solver}, nil
}

// partJSON is a part of the task as the JSON it is. A client may send a part
// whose schema names no type as a string holding the JSON rather than as the
// JSON itself, and the model cannot choose otherwise: such a string is read for
// the JSON it holds, decoded into the same plain values as a part sent as JSON
// — a number as a float — so that the checks see the same task either way. A
// string that opens as an object or a list but does not decode is handed on as
// it is, so that the checks say its JSON is broken rather than that it is no
// object; any other string stays the string it is, for the checks to refuse.
func partJSON(part any) (json.RawMessage, error) {
	text, isText := part.(string)
	if !isText {
		return json.Marshal(part)
	}
	var held any
	if json.Unmarshal([]byte(text), &held) == nil {
		return json.Marshal(held)
	}
	if opened := strings.TrimSpace(text); strings.HasPrefix(opened, "{") || strings.HasPrefix(opened, "[") {
		return json.RawMessage(text), nil
	}
	return json.Marshal(text)
}

// handedInOut is what submit_task hands back: the task on the child's card, or
// why there is none — every reason at once — and always who the card is for.
// Nothing in it gives the answer away: not the letter, not the explanations of
// the wrong options, not the solution.
type handedInOut struct {
	Screen       string      `json:"screen"`
	Status       string      `json:"status,omitempty"`
	Code         string      `json:"code,omitempty"`
	Reasons      []reasonOut `json:"reasons,omitempty"`
	Unchecked    []string    `json:"unchecked,omitempty"`
	Attempt      int         `json:"attempt,omitempty"`
	AttemptsLeft *int        `json:"attempts_left,omitempty"`
	LastAnswer   *answerLine `json:"last_answer"`
	Child        *childLine  `json:"child"`
	Task         *cardOut    `json:"task"`
}

// reasonOut is one check a task failed, and everything it found. No message
// names an answer letter or quotes an option.
type reasonOut struct {
	Code     string   `json:"code"`
	Messages []string `json:"messages"`
}

// childLine is who a card is for: the name the child goes by and the grade, for
// the top line of the card and its badge, and the language the parent chose
// for the cards, which the card speaks in place of the chat's — null when they
// chose none. It stands beside the task and never inside its text.
type childLine struct {
	Pseudonym  string  `json:"pseudonym"`
	Grade      int     `json:"grade"`
	UILanguage *string `json:"ui_language"`
}

func childLineOf(s *profile.Student) *childLine {
	return &childLine{Pseudonym: s.Pseudonym, Grade: s.Grade, UILanguage: s.UILanguage}
}

// cardOut is the task as the child's card shows it: the wording, the drawing,
// the five options and the hint, which the card keeps folded until the child
// opens it. What would give the answer away is sealed in the profile and never
// here.
type cardOut struct {
	ID       string            `json:"id"`
	Topic    string            `json:"topic"`
	Language string            `json:"language"`
	Question string            `json:"question"`
	Drawing  string            `json:"drawing"`
	Options  map[string]string `json:"options"`
	Hint     string            `json:"hint"`
}

func cardOf(task *profile.CurrentTask) *cardOut {
	return &cardOut{
		ID:       task.ID,
		Topic:    task.Topic,
		Language: task.Language,
		Question: task.Wording,
		Drawing:  task.Drawing,
		Options:  task.Options,
		Hint:     task.Hint,
	}
}

func (s *Service) submitTaskTool() Tool {
	return Define(Spec{
		Name:  "submit_task",
		Title: "Hand in a written task",
		Description: "Hands in the task you wrote for the open request, with its request id, to be checked and put " +
			"on the child's card. Each call spends one of three attempts. A refusal names every reason at once: fix " +
			"them all and hand the task in again with the same request id. When the task is accepted the card shows " +
			"it without its answer: never say which option is right before the child has answered, whatever the " +
			"child asks. Every result carries last_answer, the last answer the child gave, maybe on a card without " +
			"you.",
		DrawsCard: true,
	}, s.submitTask)
}

// submitTask takes a written task through the checks and, when it passes, puts
// it on the child's card with its answer sealed. The slow half of the checks —
// the runs of the solver — comes before the profile is read, so that nothing
// slow stands between the read and the write.
func (s *Service) submitTask(ctx context.Context, account store.Account, in submitTaskIn) (Reply[handedInOut], error) {
	started := time.Now()
	submission, err := in.submission()
	if err != nil {
		return Reply[handedInOut]{}, err
	}
	examined, err := s.examine(ctx, submission)
	if err != nil {
		return Reply[handedInOut]{}, err
	}
	for _, run := range examined.Runs() {
		s.events.write(ctx, account, eventSolverRun,
			zap.String("status", string(run.Status)),
			zap.Uint64("steps", run.Steps),
			zap.Int64("duration_ms", run.Duration.Milliseconds()),
		)
	}
	return afresh(ctx, func() (Reply[handedInOut], error) { return s.review(ctx, account, &in, examined, started) })
}

// review is one read of the profile, the judgement of the examined task
// against what it holds, and the write of how the review went.
func (s *Service) review(ctx context.Context, account store.Account, in *submitTaskIn, examined checks.Examined, started time.Time) (Reply[handedInOut], error) {
	p, revision, err := s.store.Load(ctx, account)
	switch {
	case errors.Is(err, store.ErrNotFound):
		return Reply[handedInOut]{
			Text:    "There is no profile yet, so no task was asked for and nothing was checked. " + firstRunText,
			Payload: handedInOut{Screen: screenFirstRun, Status: statusStale, Code: codeStaleRequest},
		}, nil
	case err != nil:
		return Reply[handedInOut]{}, fmt.Errorf("mcp: read the profile: %w", err)
	}

	now := s.now()
	request := p.OpenRequest
	if request == nil || request.ID != in.RequestID || !request.Awaited(s.window, now) {
		return s.stale(p), nil
	}
	outcome, err := s.judge(ctx, examined, checks.Against{
		Asked: &request.Brief, Language: request.Language, Fingerprints: p.TaskFingerprints,
	})
	if err != nil {
		return Reply[handedInOut]{}, err
	}
	done := &reviewed{
		account: account, profile: p, revision: revision, outcome: &outcome, attempt: request.Attempts + 1,
		now: now, started: started,
	}

	if !outcome.Accepted() {
		return s.refuse(ctx, done)
	}
	return s.hand(ctx, done, in.Solver)
}

// examine reads the task and runs its solver inside a span of its own, so that
// the runs are recorded under it and it under the call. A failure here is the
// sandbox's, not the task's, and it spends no attempt.
func (s *Service) examine(ctx context.Context, submission *checks.Submission) (checks.Examined, error) {
	ctx, span := s.tracer.Start(ctx, "examine_task")
	defer span.End()

	examined, err := s.reviewer.Examine(ctx, submission)
	if err != nil {
		span.SetStatus(codes.Error, "the task could not be examined")
		return checks.Examined{}, fmt.Errorf("mcp: examine the task: %w", err)
	}
	return examined, nil
}

// judge runs every check against what the profile holds, inside a span of its
// own that says how the review ended in words from a closed list.
func (s *Service) judge(ctx context.Context, examined checks.Examined, against checks.Against) (checks.Outcome, error) {
	_, span := s.tracer.Start(ctx, "judge_task")
	defer span.End()

	outcome, err := s.reviewer.Judge(examined, against)
	if err != nil {
		span.SetStatus(codes.Error, "the task could not be judged")
		return checks.Outcome{}, fmt.Errorf("mcp: judge the task: %w", err)
	}
	event := outcome.Event()
	span.SetAttributes(
		attribute.String("mathtrail.review.outcome", event.Outcome),
		attribute.String("mathtrail.review.primary", string(event.Primary)),
	)
	return outcome, nil
}

// stale is a task handed in for no request that is open — one that was
// accepted already, ran out of attempts, was replaced or waited too long — or
// for another request than the open one. Nothing is judged, spent or written.
// The card shows the task the child is working on, when there is one: a task
// handed in twice, the answer to the first gone astray, must not turn the card
// the child is working on into a wait. A task that has had its answer is done
// with, and the card waits for the next.
func (s *Service) stale(p *profile.Profile) Reply[handedInOut] {
	lead := "The request_id is not the open request's: that request was accepted already, ran out of attempts, " +
		"was replaced by a newer one or waited too long. Nothing was checked or spent."
	task := p.InFlight()
	if task == nil {
		return Reply[handedInOut]{
			Text: lead + " Ask for a new task with next_task.",
			Payload: handedInOut{
				Screen: screenWaiting, Status: statusStale, Code: codeStaleRequest,
				LastAnswer: lastAnswerOf(p), Child: childLineOf(&p.Student),
			},
		}
	}
	reply := onTheCard(p, task, fmt.Sprintf("%s Task %s is on the child's card: wait for the child's "+
		"answer, and ask for a new task only when the child wants another.", lead, task.ID))
	reply.Payload.Status, reply.Payload.Code = statusStale, codeStaleRequest
	return reply
}

// reviewed is a task handed in and judged, with what writing the outcome takes:
// whose profile it goes into, as it was read, the review itself, the attempt
// the task spent and the moment of the call.
type reviewed struct {
	account  store.Account
	profile  *profile.Profile
	revision store.Revision
	outcome  *checks.Outcome
	attempt  int
	now      time.Time
	// started is when the call began, for how long the review took.
	started time.Time
}

// record writes the profile the review changed, over the revision it was read
// at, and then the line about the review: an attempt is counted in the log
// only once the file counts it, so that a write lost to another is not counted
// twice when the task is handed in again.
func (s *Service) record(ctx context.Context, done *reviewed) error {
	done.profile.Touch(s.version, done.now)
	if _, err := s.store.Save(ctx, done.account, done.profile, done.revision); err != nil {
		return fmt.Errorf("mcp: save the profile: %w", err)
	}
	event := done.outcome.Event()
	s.events.write(ctx, done.account, eventTaskSubmitted,
		submittedFields(done.attempt, &event, time.Since(done.started))...)
	return nil
}

// refuse spends the attempt a task failed its checks on, and says every reason
// at once. The last attempt closes the request with nothing handed out.
func (s *Service) refuse(ctx context.Context, done *reviewed) (Reply[handedInOut], error) {
	p, outcome, attempt := done.profile, done.outcome, done.attempt
	requestID := p.OpenRequest.ID
	exhausted, err := p.Refuse(done.now)
	if err != nil {
		return Reply[handedInOut]{}, fmt.Errorf("mcp: refuse the task: %w", err)
	}
	if err := s.record(ctx, done); err != nil {
		return Reply[handedInOut]{}, err
	}

	left := profile.MaxAttempts - attempt
	payload := handedInOut{
		Screen: screenWaiting, Status: statusRejected, Code: string(outcome.Primary()),
		Reasons: reasonsOf(outcome), Unchecked: outcome.Unchecked, Attempt: attempt, AttemptsLeft: &left,
		LastAnswer: lastAnswerOf(p), Child: childLineOf(&p.Student),
	}
	lead := fmt.Sprintf("Refused, attempt %d of %d: nothing was handed to the child. Fix every reason below and "+
		"hand the task in again with submit_task and request_id %s; %d left.", attempt, profile.MaxAttempts, requestID, left)
	if exhausted {
		payload.Code = codeAttemptsExhausted
		lead = fmt.Sprintf("Refused, and that was the last of %d attempts: request %s is closed and nothing was "+
			"handed to the child. Tell the child this task did not work out, and ask for a new one with next_task.",
			profile.MaxAttempts, requestID)
	}
	return Reply[handedInOut]{Text: lead + "\n" + reasonsText(outcome), Payload: payload}, nil
}

// hand puts an accepted task on the child's card: the task the request asked
// for, with everything that gives its answer away sealed, and the request
// closed.
func (s *Service) hand(ctx context.Context, done *reviewed, program string) (Reply[handedInOut], error) {
	p, task := done.profile, done.outcome.Draft.Task
	request, left := *p.OpenRequest, p.InFlight()
	issued, err := p.Issue(&profile.Written{
		Wording:             task.Question,
		Drawing:             task.Drawing,
		Options:             task.Options,
		Hint:                task.Hint,
		Fingerprint:         checks.Fingerprint(task.Question, request.Language),
		InstructionsVersion: s.content.InstructionsVersion(),
	}, secretOf(task, program), s.sealer, done.now)
	if err != nil {
		return Reply[handedInOut]{}, fmt.Errorf("mcp: hand the task out: %w", err)
	}
	if err := s.record(ctx, done); err != nil {
		return Reply[handedInOut]{}, err
	}

	if left != nil {
		s.events.write(ctx, done.account, eventTaskSkipped, s.skippedFields(left.Topic, left.GradeLevel, left.Difficulty)...)
	}
	s.events.write(ctx, done.account, eventTaskAccepted,
		zap.String("topic", issued.Topic),
		zap.String("level", string(issued.GradeLevel)),
		zap.Int("difficulty", issued.Difficulty),
		zap.Int("attempts", done.attempt),
		zap.Int64("seconds_since_request", int64(done.now.Sub(request.OpenedAt.Time)/time.Second)),
	)
	reply := onTheCard(p, issued, fmt.Sprintf("Accepted at attempt %d: task %s is on the child's card. Without a "+
		"card, read out the question, the drawing and the options A to E below, and give the hint only when the "+
		"child asks for it. Never say which option is right before the child has answered; record the answer with "+
		"submit_answer.", done.attempt, issued.ID))
	reply.Payload.Attempt = done.attempt
	return reply, nil
}

// secretOf is everything about a task that would give its answer away, as it
// is sealed: the letter, what every wrong option leads to, the solution and
// the program that proved it.
func secretOf(task *checks.Task, program string) profile.TaskSecret {
	distractors := make(map[string]profile.Distractor, len(task.Distractors))
	for letter, wrong := range task.Distractors {
		distractors[letter] = profile.Distractor{Text: wrong.Text, Trap: wrong.Trap}
	}
	return profile.TaskSecret{
		Answer:      task.CorrectAnswer,
		Distractors: distractors,
		Solution:    task.Solution,
		Solver:      program,
	}
}

// onTheCard is the task the child has, as the card shows it and as the words
// read it out where there is no card.
func onTheCard(p *profile.Profile, task *profile.CurrentTask, lead string) Reply[handedInOut] {
	return Reply[handedInOut]{
		Text: lead + "\n\n" + taskWords(task),
		Payload: handedInOut{
			Screen: screenTask, LastAnswer: lastAnswerOf(p), Child: childLineOf(&p.Student), Task: cardOf(task),
		},
	}
}

// taskWords is the task in words, for the model to read out: the question, the
// drawing, the five options and the hint, as the model wrote them. The texts
// stand in quotes, and the drawing in a block of its own, as it is to be shown.
func taskWords(task *profile.CurrentTask) string {
	var words strings.Builder
	fmt.Fprintf(&words, "Question: %s\n", quoted(task.Wording))
	if drawing := strings.TrimSuffix(task.Drawing, "\n"); drawing != "" {
		fence := fenceFor(drawing)
		fmt.Fprintf(&words, "Drawing:\n%s\n%s\n%s\n", fence, drawing, fence)
	}
	words.WriteString("Options:")
	for _, letter := range solver.Letters() {
		fmt.Fprintf(&words, " %s) %s", letter, quoted(task.Options[letter]))
	}
	fmt.Fprintf(&words, "\nHint, only when the child asks: %s", quoted(task.Hint))
	return words.String()
}

// fenceFor is a fence no drawing can close early: longer than any run of
// backticks inside it, and never shorter than three.
func fenceFor(drawing string) string {
	longest, run := 0, 0
	for _, r := range drawing {
		if r != '`' {
			run = 0
			continue
		}
		run++
		longest = max(longest, run)
	}
	return strings.Repeat("`", max(3, longest+1))
}

// reasonsOf are the checks a task failed, each with everything it found.
func reasonsOf(outcome *checks.Outcome) []reasonOut {
	reasons := outcome.Reasons()
	out := make([]reasonOut, 0, len(reasons))
	for _, reason := range reasons {
		out = append(out, reasonOut{Code: string(reason.Code), Messages: reason.Messages})
	}
	return out
}

// reasonsText is every reason a task was refused, in the order the checks
// report them, and every check that could not run yet.
func reasonsText(outcome *checks.Outcome) string {
	lines := make([]string, 0, len(outcome.Problems)+len(outcome.Unchecked))
	for _, reason := range outcome.Reasons() {
		lines = append(lines, "- "+string(reason.Code)+": "+strings.Join(reason.Messages, "; "))
	}
	for _, sentence := range outcome.Unchecked {
		lines = append(lines, "- not checked yet: "+sentence)
	}
	return strings.Join(lines, "\n")
}

// submittedFields are what the line about a review keeps of it: codes, counts
// and costs, and not a word of the task.
func submittedFields(attempt int, event *checks.Submitted, took time.Duration) []zap.Field {
	failed := make([]string, 0, len(event.Failed))
	for _, code := range event.Failed {
		failed = append(failed, string(code))
	}
	return []zap.Field{
		zap.Int("attempt", attempt),
		zap.String("outcome", event.Outcome),
		zap.String("primary", string(event.Primary)),
		zap.Strings("failed", failed),
		zap.Strings("minor_issues", event.MinorIssues),
		zap.Int64("duration_ms", took.Milliseconds()),
		zap.Uint64("solver_steps", event.SolverSteps),
		zap.Int64("solver_ms", event.SolverTime.Milliseconds()),
	}
}

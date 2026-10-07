package mcpserver_test

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"go.uber.org/zap/zapcore"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/profile"
	"github.com/MathTrail/mathtrail-standalone/internal/domain/rating"
	"github.com/MathTrail/mathtrail-standalone/internal/store"
)

// holding is a store whose writes go through until the case arms it; then the
// next write is held just before it reaches the store, until the case lets it
// on.
type holding struct {
	store.Storage
	armed   atomic.Bool
	once    sync.Once
	reached chan struct{}
	letGo   chan struct{}
	release func()
}

func holdingOver(kept store.Storage) *holding {
	h := &holding{Storage: kept, reached: make(chan struct{}), letGo: make(chan struct{})}
	h.release = sync.OnceFunc(func() { close(h.letGo) })
	return h
}

func (h *holding) Save(ctx context.Context, account store.Account, p *profile.Profile, expected store.Revision) (store.Revision, error) {
	if h.armed.Load() {
		first := false
		h.once.Do(func() { first = true })
		if first {
			close(h.reached)
			<-h.letGo
		}
	}
	return h.Storage.Save(ctx, account, p, expected)
}

// heldLesson is a lesson over a store that holds a write once armed. The write
// is let on before the server closes, so that a case that fails while it is
// held does not leave the server waiting for it.
func heldLesson(t *testing.T, kept store.Storage) (*harness, *mcp.ClientSession, *holding) {
	t.Helper()

	held := holdingOver(kept)
	h, session := lesson(t, held)
	t.Cleanup(held.release)
	return h, session, held
}

// answeredWhileHeld makes a call and is its answer, which has to arrive while
// the write it leaves for after the answer is still held.
func answeredWhileHeld(t *testing.T, session *mcp.ClientSession, tool string, args any) *mcp.CallToolResult {
	t.Helper()

	done := make(chan *mcp.CallToolResult, 1)
	go func() {
		result, err := session.CallTool(t.Context(), &mcp.CallToolParams{Name: tool, Arguments: args})
		if err != nil {
			result = nil
		}
		done <- result
	}()
	select {
	case result := <-done:
		if result == nil {
			t.Fatalf("%s failed at the protocol", tool)
		}
		return result
	case <-time.After(5 * time.Second):
		t.Fatalf("%s did not answer while its write was held, want the answer before the write", tool)
		return nil
	}
}

// lateLine is the one line of a write after the answer a tool made.
func lateLine(t *testing.T, h *harness, tool string) map[string]any {
	t.Helper()

	var found []map[string]any
	lines := h.logs.FilterMessage("write_after_answer").All()
	for i := range lines {
		if fields := lines[i].ContextMap(); fields["tool"] == tool {
			fields["level"] = lines[i].Level
			found = append(found, fields)
		}
	}
	if len(found) != 1 {
		t.Fatalf("write_after_answer lines of %s = %d, want 1", tool, len(found))
	}
	return found[0]
}

// lateOutcome is how the one write after the answer a tool made went.
func lateOutcome(t *testing.T, h *harness, tool string) string {
	t.Helper()
	return fmt.Sprint(lateLine(t, h, tool)["outcome"])
}

// A task kept is on the card the moment next_task answers: the hand-out is
// written after the answer, and the lines of it once the file holds it.
func TestAKeptTaskComesBeforeItsHandOutIsWritten(t *testing.T) {
	t.Parallel()

	h, session, held := heldLesson(t, racer(t))
	race, relay := keepTheRelay(t, session, held)
	accepted := len(linesOf(h, "task_accepted"))
	held.armed.Store(true)

	card := payloadOf[requestPayload](t, answeredWhileHeld(t, session, "next_task", map[string]any{"language": "en"}))
	if card.Screen != "task" || card.Task == nil || card.Task.ID != relay {
		t.Fatalf("next_task = %+v, want the relay on the card at once", card)
	}
	<-held.reached
	if p, _ := loadKept(t, held.Storage); p.ReadyTask == nil || p.CurrentTask.ID != profile.TaskIDFor(race.ID) {
		t.Errorf("before the write the file holds %+v on the card and %+v kept, want it as it was", p.CurrentTask, p.ReadyTask)
	}
	if got := len(linesOf(h, "task_accepted")); got != accepted {
		t.Errorf("task_accepted lines before the write = %d, want %d: a line waits for the file", got, accepted)
	}

	held.release()
	h.settle()
	p, _ := loadKept(t, held.Storage)
	if p.CurrentTask == nil || p.CurrentTask.ID != relay || p.ReadyTask != nil {
		t.Errorf("after the write the file holds %+v on the card and %+v kept, want the relay handed out", p.CurrentTask, p.ReadyTask)
	}
	acceptedLines := linesOf(h, "task_accepted")
	if got := len(acceptedLines); got != accepted+1 {
		t.Fatalf("task_accepted lines after the write = %d, want %d", got, accepted+1)
	}
	if outcome := lateOutcome(t, h, "next_task"); outcome != "written" {
		t.Errorf("the hand-out's write went %q, want written", outcome)
	}
	// The lines written after the answer are the call's: they name its span,
	// which ties them to the call and to the host it came from.
	var asked []string
	tools := h.toolLines()
	for i := range tools {
		if tools[i].ContextMap()["tool"] == "next_task" {
			asked = append(asked, field(t, &tools[i], spanKey))
		}
	}
	call := asked[len(asked)-1]
	if got := field(t, &acceptedLines[len(acceptedLines)-1], spanKey); got != call {
		t.Errorf("the hand-out's task_accepted names span %q, want the call's, %q", got, call)
	}
	if got := fmt.Sprint(lateLine(t, h, "next_task")[spanKey]); got != call {
		t.Errorf("the hand-out's write_after_answer names span %q, want the call's, %q", got, call)
	}
}

// spanKey is the field a line names its span in.
const spanKey = "logging.googleapis.com/spanId"

// An answer is told the moment submit_answer answers: it is written after the
// answer, and its line once the file holds it.
func TestAnAnswerIsToldBeforeItIsWritten(t *testing.T) {
	t.Parallel()

	p := raceOnTheCard(t, rating.TrialAnswers)
	h, session, held := heldLesson(t, keptAsIs(t, p))
	held.armed.Store(true)

	told := answered(t, answeredWhileHeld(t, session, "submit_answer", map[string]any{"task_id": p.CurrentTask.ID, "answer": "C"}))
	if told.Choice != "C" || told.AlreadyAnswered {
		t.Fatalf("submit_answer = %+v, want C recorded", told)
	}
	<-held.reached
	if before, _ := loadKept(t, held.Storage); before.CurrentTask.Answered != nil {
		t.Errorf("before the write the file holds the answer %+v, want none yet", before.CurrentTask.Answered)
	}
	if lines := linesOf(h, "answer_recorded"); len(lines) != 0 {
		t.Errorf("answer_recorded lines before the write = %d, want none", len(lines))
	}

	held.release()
	h.settle()
	if after, _ := loadKept(t, held.Storage); after.CurrentTask.Answered == nil || after.CurrentTask.Answered.Choice != "C" {
		t.Errorf("after the write the file holds the answer %+v, want C", after.CurrentTask.Answered)
	}
	if lines := linesOf(h, "answer_recorded"); len(lines) != 1 {
		t.Errorf("answer_recorded lines after the write = %d, want 1", len(lines))
	}
	if outcome := lateOutcome(t, h, "submit_answer"); outcome != "written" {
		t.Errorf("the answer's write went %q, want written", outcome)
	}
}

// A call that comes while a write after the answer is still on its way, as the
// model's prepare_task comes right after next_task, waits for it, and finds
// what the answer said: the task kept handed out, and the next one to write.
func TestACallWaitsForTheWriteBeforeIt(t *testing.T) {
	t.Parallel()

	h, session, held := heldLesson(t, racer(t))
	keepTheRelay(t, session, held)
	held.armed.Store(true)
	answeredWhileHeld(t, session, "next_task", map[string]any{"language": "en"})
	<-held.reached

	done := make(chan string, 1)
	go func() {
		result, err := session.CallTool(t.Context(), &mcp.CallToolParams{Name: "prepare_task", Arguments: aheadChoice})
		if err != nil || len(result.Content) != 1 {
			done <- ""
			return
		}
		text, _ := result.Content[0].(*mcp.TextContent)
		done <- text.Text
	}()
	// Long enough for the call to have read the file, were it not waiting.
	time.Sleep(300 * time.Millisecond)
	held.release()
	words := <-done
	if words == "" || strings.Contains(words, "written already") {
		t.Errorf("prepare_task said %q, want the package of the next task, read once the hand-out was written", words)
	}
	h.settle()
	if p, _ := loadKept(t, held.Storage); p.OpenRequest == nil || !p.OpenRequest.Ahead {
		t.Errorf("the file holds the request %+v, want one for the next task, written ahead", p.OpenRequest)
	}
}

// The answer leaves as an event of the stream as soon as it is ready, and the
// request ends only once the write after it is done.
func TestTheAnswerLeavesBeforeTheRequestEnds(t *testing.T) {
	t.Parallel()

	h, session, held := heldLesson(t, racer(t))
	keepTheRelay(t, session, held)
	held.armed.Store(true)

	req, err := http.NewRequestWithContext(t.Context(), http.MethodPost, h.server.URL+"/mcp",
		strings.NewReader(legacyCall("next_task", `{"language":"en"}`)))
	if err != nil {
		t.Fatalf("NewRequest() error = %v, want nil", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json, text/event-stream")
	resp, err := h.server.Client().Do(req)
	if err != nil {
		t.Fatalf("POST /mcp error = %v, want nil", err)
	}
	defer func() { _ = resp.Body.Close() }()
	reader := bufio.NewReader(resp.Body)
	var event strings.Builder
	for !strings.HasSuffix(event.String(), "\n\n") {
		line, err := reader.ReadString('\n')
		if err != nil {
			t.Fatalf("reading the answer's event: %v; read %q", err, event.String())
		}
		event.WriteString(line)
	}
	if !strings.Contains(event.String(), `"screen":"task"`) {
		t.Errorf("the first event is %q, want the card with the task kept", event.String())
	}

	ended := make(chan error, 1)
	go func() {
		_, err := io.ReadAll(reader)
		ended <- err
	}()
	<-held.reached
	select {
	case <-ended:
		t.Fatal("the request ended while its write was held, want it held open until the write is done")
	case <-time.After(100 * time.Millisecond):
	}
	held.release()
	select {
	case err := <-ended:
		if err != nil {
			t.Errorf("the rest of the answer: %v, want its end", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("the request did not end once its write was let on")
	}
}

// overtaking is a store whose first write, once armed, finds another
// instance's write in the file just before it: the other writer's change is
// made first, through the store itself, and the held write then finds the
// file moved on.
type overtaking struct {
	store.Storage
	armed atomic.Bool
	once  sync.Once
	edit  func(*profile.Profile)
}

func (o *overtaking) Save(ctx context.Context, account store.Account, p *profile.Profile, expected store.Revision) (store.Revision, error) {
	if o.armed.Load() {
		o.once.Do(func() {
			theirs, revision, err := o.Load(ctx, account)
			if err != nil {
				return
			}
			o.edit(theirs)
			theirs.Touch("another-instance", lessonDay)
			_, _ = o.Storage.Save(ctx, account, theirs, revision)
		})
	}
	return o.Storage.Save(ctx, account, p, expected)
}

// A hand-out whose file another instance wrote first is made again on what is
// there now, and both changes stay.
func TestAHandOutOvertakenIsMadeAgainOnTheFile(t *testing.T) {
	t.Parallel()

	kept := &overtaking{Storage: racer(t), edit: func(p *profile.Profile) { p.Student.Notes = "Likes puzzles." }}
	h, session := lesson(t, kept)
	_, relay := keepTheRelay(t, session, kept)
	kept.armed.Store(true)

	call(t, session, "next_task", map[string]any{"language": "en"})
	h.settle()
	p, _ := loadKept(t, kept.Storage)
	if p.CurrentTask == nil || p.CurrentTask.ID != relay || p.Student.Notes != "Likes puzzles." {
		t.Errorf("the file holds %+v on the card and the notes %q, want the relay and the other instance's notes",
			p.CurrentTask, p.Student.Notes)
	}
	if outcome := lateOutcome(t, h, "next_task"); outcome != "remade" {
		t.Errorf("the hand-out's write went %q, want remade", outcome)
	}
	if got := len(linesOf(h, "task_accepted")); got != 2 {
		t.Errorf("task_accepted lines = %d, want the race's and the relay's", got)
	}
}

// A hand-out whose task another instance let go first cannot be made: the
// card shows a task the file no longer keeps, and the line says so, as a
// warning.
func TestAHandOutWhoseTaskWasLetGoIsLost(t *testing.T) {
	t.Parallel()

	kept := &overtaking{Storage: racer(t), edit: func(p *profile.Profile) { p.DropReady() }}
	h, session := lesson(t, kept)
	_, relay := keepTheRelay(t, session, kept)
	kept.armed.Store(true)

	call(t, session, "next_task", map[string]any{"language": "en"})
	h.settle()
	if p, _ := loadKept(t, kept.Storage); p.CurrentTask != nil && p.CurrentTask.ID == relay {
		t.Errorf("the file holds the relay on the card, want it let go as the other instance left it")
	}
	line := lateLine(t, h, "next_task")
	if line["outcome"] != "lost" || line["level"] != zapcore.WarnLevel {
		t.Errorf("the hand-out's line = %v, want lost, as a warning", line)
	}
}

// failingLater is a store whose writes fail once armed, as a Drive with no
// room left does.
type failingLater struct {
	store.Storage
	armed atomic.Bool
}

func (f *failingLater) Save(ctx context.Context, account store.Account, p *profile.Profile, expected store.Revision) (store.Revision, error) {
	if f.armed.Load() {
		return "", store.ErrStorageFull
	}
	return f.Storage.Save(ctx, account, p, expected)
}

// A write after the answer that fails is said in one warning, in the words a
// call's line uses, and with nothing the child could be known by or the answer
// read from.
func TestAWriteAfterTheAnswerThatFailsIsOneWarning(t *testing.T) {
	t.Parallel()

	p := raceOnTheCard(t, rating.TrialAnswers)
	kept := &failingLater{Storage: keptAsIs(t, p)}
	h, session := lesson(t, kept)
	kept.armed.Store(true)

	call(t, session, "submit_answer", map[string]any{"task_id": p.CurrentTask.ID, "answer": "C"})
	h.settle()
	line := lateLine(t, h, "submit_answer")
	if line["outcome"] != "failed" || line["error"] != "storage_full" || line["level"] != zapcore.WarnLevel {
		t.Errorf("the answer's line = %v, want failed for storage_full, as a warning", line)
	}
	if loaded, _ := shipped(); line["instructions_version"] != loaded.InstructionsVersion() {
		t.Errorf("the answer's line names the instructions %v, want those the lesson's lines name, %v",
			line["instructions_version"], loaded.InstructionsVersion())
	}
	for key, value := range line {
		switch key {
		case "answer", "choice", "task_id", "topic":
			t.Errorf("the line carries %s = %v, want nothing of the task or the answer", key, value)
		}
		text := fmt.Sprint(value)
		for _, secret := range []string{p.Student.Pseudonym, p.CurrentTask.Wording} {
			if strings.Contains(text, secret) {
				t.Errorf("the line's %s carries %q", key, secret)
			}
		}
	}
	if lines := linesOf(h, "answer_recorded"); len(lines) != 0 {
		t.Errorf("answer_recorded lines = %d, want none for an answer the file does not hold", len(lines))
	}
}

// An answer whose profile was deleted for good before its write is told
// recorded, since the call answers before it writes, and lost: its line says
// so, as a warning, and nothing counts it.
func TestAnAnswerWhoseProfileWasDeletedBeforeItsWriteIsLost(t *testing.T) {
	t.Parallel()

	p := raceOnTheCard(t, rating.TrialAnswers)
	h, session := lesson(t, &vanishing{Storage: keptAsIs(t, p)})
	answered(t, answerIt(t, session, p.CurrentTask.ID, "C", false))

	h.settle()
	if line := lateLine(t, h, "submit_answer"); line["outcome"] != "lost" || line["level"] != zapcore.WarnLevel {
		t.Errorf("the answer's write = %v, want lost, as a warning", line)
	}
	if lines := linesOf(h, "answer_recorded"); len(lines) != 0 {
		t.Errorf("answer_recorded lines = %d for an answer the file does not hold, want none", len(lines))
	}
}

// unreachable is a store whose first write, once armed, cannot reach the file
// and writes nothing, and whose first read after it lags behind, as Drive may
// for a moment.
type unreachable struct {
	store.Storage
	armed, lagging atomic.Bool
}

func (u *unreachable) Save(ctx context.Context, account store.Account, p *profile.Profile, expected store.Revision) (store.Revision, error) {
	if u.armed.CompareAndSwap(true, false) {
		u.lagging.Store(true)
		return "", fmt.Errorf("%w: no answer in time", store.ErrUnavailable)
	}
	return u.Storage.Save(ctx, account, p, expected)
}

func (u *unreachable) Load(ctx context.Context, account store.Account) (*profile.Profile, store.Revision, error) {
	if u.lagging.CompareAndSwap(true, false) {
		return nil, "", store.ErrBehind
	}
	return u.Storage.Load(ctx, account)
}

// A write after the answer that could not reach the file wrote nothing, so it
// is made again on a fresh read, which is read once more when it lags behind,
// and the answer is written on the second try.
func TestAWriteThatCouldNotReachTheFileIsMadeAgain(t *testing.T) {
	t.Parallel()

	p := raceOnTheCard(t, rating.TrialAnswers)
	kept := &unreachable{Storage: keptAsIs(t, p)}
	h, session := lesson(t, kept)
	kept.armed.Store(true)

	call(t, session, "submit_answer", map[string]any{"task_id": p.CurrentTask.ID, "answer": "C"})
	h.settle()
	if line := lateLine(t, h, "submit_answer"); line["outcome"] != "remade" || fmt.Sprint(line["attempts"]) != "2" {
		t.Errorf("the answer's line = %v, want remade, on the second try", line)
	}
	if after, _ := loadKept(t, kept.Storage); after.CurrentTask.Answered == nil || after.CurrentTask.Answered.Choice != "C" {
		t.Errorf("the file holds the answer %+v, want C", after.CurrentTask.Answered)
	}
	if lines := linesOf(h, "answer_recorded"); len(lines) != 1 {
		t.Errorf("answer_recorded lines = %d, want 1", len(lines))
	}
}

// panicking is a store whose writes panic once armed, as a defect would.
type panicking struct {
	store.Storage
	armed atomic.Bool
}

func (k *panicking) Save(ctx context.Context, account store.Account, p *profile.Profile, expected store.Revision) (store.Revision, error) {
	if k.armed.Load() {
		panic("the write broke")
	}
	return k.Storage.Save(ctx, account, p, expected)
}

// A write after the answer that panics fails alone, as a call that panics
// does: its line says so, as an error with the stack, and the instance and the
// account's next call go on.
func TestAWriteAfterTheAnswerThatPanicsFailsAlone(t *testing.T) {
	t.Parallel()

	p := raceOnTheCard(t, rating.TrialAnswers)
	kept := &panicking{Storage: keptAsIs(t, p)}
	h, session := lesson(t, kept)
	kept.armed.Store(true)

	call(t, session, "submit_answer", map[string]any{"task_id": p.CurrentTask.ID, "answer": "C"})
	progress := payloadOf[map[string]any](t, call(t, session, "get_progress", nil))
	h.settle()
	if progress["screen"] != "progress" {
		t.Errorf("get_progress after the write broke = %v, want the progress", progress)
	}
	line := lateLine(t, h, "submit_answer")
	if line["outcome"] != "failed" || line["error"] != "panic" || line["level"] != zapcore.ErrorLevel {
		t.Errorf("the answer's line = %v, want failed for a panic, as an error", line)
	}
	if line["stack"] == nil || line["panic"] != "a value of type string" {
		t.Errorf("the answer's line = %v, want the stack and the type of what the panic was raised with", line)
	}
	if lines := linesOf(h, "answer_recorded"); len(lines) != 0 {
		t.Errorf("answer_recorded lines = %d, want none for an answer the file does not hold", len(lines))
	}
}

// unheard is a store whose first write, once armed, lands in the file and is
// answered as Drive out of reach all the same, as an upload Drive failed on
// may be.
type unheard struct {
	store.Storage
	armed atomic.Bool
}

func (u *unheard) Save(ctx context.Context, account store.Account, p *profile.Profile, expected store.Revision) (store.Revision, error) {
	revision, err := u.Storage.Save(ctx, account, p, expected)
	if err == nil && u.armed.CompareAndSwap(true, false) {
		return "", fmt.Errorf("%w: no answer to the upload", store.ErrUnavailable)
	}
	return revision, err
}

// An upload that landed though Drive never answered it is this write, found
// by the fresh read at the number it gave the profile: the answer counts once,
// as written, and is not made again.
func TestAWriteThatLandedUnheardCountsAsWritten(t *testing.T) {
	t.Parallel()

	p := raceOnTheCard(t, rating.TrialAnswers)
	kept := &unheard{Storage: keptAsIs(t, p)}
	h, session := lesson(t, kept)
	kept.armed.Store(true)

	call(t, session, "submit_answer", map[string]any{"task_id": p.CurrentTask.ID, "answer": "C"})
	h.settle()
	if line := lateLine(t, h, "submit_answer"); line["outcome"] != "written" || fmt.Sprint(line["attempts"]) != "1" {
		t.Errorf("the answer's line = %v, want written, on the first try", line)
	}
	if lines := linesOf(h, "answer_recorded"); len(lines) != 1 {
		t.Errorf("answer_recorded lines = %d, want 1", len(lines))
	}
}

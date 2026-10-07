package session_test

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/jsonrpc"
	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/MathTrail/mathtrail-standalone/tools/load/servicetest"
	"github.com/MathTrail/mathtrail-standalone/tools/load/session"
)

// says is a tool that answers every call with the result given.
func says(result *mcp.CallToolResult) mcp.ToolHandler {
	return func(context.Context, *mcp.CallToolRequest) (*mcp.CallToolResult, error) { return result, nil }
}

// words is a result's content of one text.
func words(text string) []mcp.Content { return []mcp.Content{&mcp.TextContent{Text: text}} }

// childOf is a child of a run against the target, closed when the case ends.
func childOf(t *testing.T, target session.Target, timeout time.Duration, name string) *session.Child {
	t.Helper()

	service := session.Open(target, timeout)
	child := service.Child(name)
	t.Cleanup(func() {
		if err := child.Close(); err != nil {
			t.Errorf("Close() error = %v, want nil", err)
		}
		service.Close()
	})
	return child
}

// Every answer is told apart by what it says: an answer by being one, a refusal
// by the status and code its payload carries, a failure by the sentence it
// begins with — and two of those sentences, a sandbox with no slot and a call
// a pace held back, by name.
func TestAnAnswerIsToldApartByWhatItSays(t *testing.T) {
	t.Parallel()

	target := servicetest.Fake(t, map[string]mcp.ToolHandler{
		"answer": says(&mcp.CallToolResult{Content: words("done"), StructuredContent: map[string]any{"said": "done"}}),
		"words":  says(&mcp.CallToolResult{Content: words("Request req_1 is open.")}),
		"reject": says(&mcp.CallToolResult{Content: words("refused"),
			StructuredContent: map[string]any{"status": "rejected", "code": "solver_error"}}),
		"stale": says(&mcp.CallToolResult{Content: words("stale"),
			StructuredContent: map[string]any{"status": "stale", "code": "stale_request"}}),
		"limit": says(&mcp.CallToolResult{Content: words("tomorrow"),
			StructuredContent: map[string]any{"status": "limited", "code": "limit_reached"}}),
		"busy": says(&mcp.CallToolResult{IsError: true, Content: words("MathTrail is checking too many tasks " +
			"right now, so this task was not checked and no attempt was spent. Hand in the same task again.")}),
		"paced": says(&mcp.CallToolResult{IsError: true, Content: words("MathTrail received too many calls in " +
			"a short time, so this call was not made. Wait a moment, then make the same call again.")}),
		"fail": says(&mcp.CallToolResult{IsError: true, Content: words("Something went wrong inside MathTrail. " +
			"Try the same request again in a moment.")}),
	}, nil)
	child := childOf(t, target, 5*time.Second, "ann")

	for tool, want := range map[string]string{
		"answer": "ok",
		"words":  "ok",
		"reject": "rejected:solver_error",
		"stale":  "stale:stale_request",
		"limit":  "limited:limit_reached",
		"busy":   "failure:busy",
		"paced":  "limited:paced",
		"fail":   "failure:Something went wrong inside MathTrail.",
	} {
		t.Run(tool, func(t *testing.T) {
			answer := child.Call(t.Context(), tool, map[string]any{})
			if got := answer.Kind.String(); got != want {
				t.Errorf("Call(%s).Kind = %q, want %q (error %v)", tool, got, want, answer.Err)
			}
			if answer.Status != http.StatusOK {
				t.Errorf("Call(%s).Status = %d, want %d", tool, answer.Status, http.StatusOK)
			}
		})
	}

	// The protocol answers a call of a tool it does not have with its error,
	// and at this version with a status of 400 too: the status tells it, and
	// the error goes with it.
	t.Run("a tool the service does not have", func(t *testing.T) {
		answer := child.Call(t.Context(), "missing", map[string]any{})
		if got := answer.Kind.String(); got != "http:400(-32602)" {
			t.Errorf("Call(missing).Kind = %q, want http:400(-32602)", got)
		}
	})
}

// A call that waits while another call of the child opens the session ends
// at its own deadline, not at the end of a handshake it is only waiting for.
func TestACallWaitingForItsSessionEndsAtItsOwnDeadline(t *testing.T) {
	t.Parallel()

	release := make(chan struct{})
	target := servicetest.Fake(t, map[string]mcp.ToolHandler{
		"answer": says(&mcp.CallToolResult{Content: words("done")}),
	}, func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Header.Get("Mcp-Method") == "server/discover" {
				<-release
			}
			next.ServeHTTP(w, r)
		})
	})
	child := childOf(t, target, 10*time.Second, "ann")

	var opening sync.WaitGroup
	var opened session.Answer
	opening.Go(func() { opened = child.Call(context.Background(), "answer", map[string]any{}) })
	time.Sleep(50 * time.Millisecond)

	hurried, cancel := context.WithTimeout(t.Context(), 100*time.Millisecond)
	defer cancel()
	answer := child.Call(hurried, "answer", map[string]any{})
	// The handshake is let go of only now, so that the call waiting for it
	// has had to end on its own.
	close(release)
	opening.Wait()

	if answer.Kind.String() != "no_answer:timeout" || answer.Took() > 2*time.Second {
		t.Errorf("the waiting call: %q after %v, want no_answer:timeout at about 100ms", answer.Kind, answer.Took())
	}
	if opened.Kind != session.Answered {
		t.Errorf("the call that opened the session: %q (%v), want an answer once the handshake was let go", opened.Kind, opened.Err)
	}
}

// A call is told by what its own request got back, however many calls of the
// child are under way beside it.
func TestEachCallKeepsTheStatusOfItsOwnRequest(t *testing.T) {
	t.Parallel()

	release := make(chan struct{})
	target := servicetest.Fake(t, map[string]mcp.ToolHandler{
		"answer": says(&mcp.CallToolResult{Content: words("done")}),
		"slow": func(context.Context, *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			<-release
			return &mcp.CallToolResult{Content: words("done")}, nil
		},
	}, func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Header.Get("Mcp-Name") == "throttled" {
				w.WriteHeader(http.StatusTooManyRequests)
				return
			}
			next.ServeHTTP(w, r)
		})
	})
	child := childOf(t, target, 5*time.Second, "ann")
	// The session is opened before the two calls, so that neither pays for it.
	if first := child.Call(t.Context(), "answer", map[string]any{}); first.Kind != session.Answered {
		t.Fatalf("the first call: %q (%v), want an answer", first.Kind, first.Err)
	}

	var slow session.Answer
	var both sync.WaitGroup
	both.Go(func() { slow = child.Call(context.Background(), "slow", map[string]any{}) })
	throttled := child.Call(t.Context(), "throttled", map[string]any{})
	close(release)
	both.Wait()

	if throttled.Kind.String() != "http:429" || throttled.Status != http.StatusTooManyRequests {
		t.Errorf("the throttled call: %q with status %d, want http:429 with 429", throttled.Kind, throttled.Status)
	}
	if slow.Kind != session.Answered || slow.Status != http.StatusOK {
		t.Errorf("the call beside it: %q with status %d (%v), want an answer with 200", slow.Kind, slow.Status, slow.Err)
	}
}

// A status that leaves the session no use to the library puts it out of use:
// the next call opens another, and the reconnect is counted.
func TestASessionTheServiceBrokeIsOpenedAgainAndCounted(t *testing.T) {
	t.Parallel()

	var once sync.Once
	target := servicetest.Fake(t, map[string]mcp.ToolHandler{
		"answer": says(&mcp.CallToolResult{Content: words("done")}),
	}, func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			broken := false
			if r.Header.Get("Mcp-Name") == "answer" {
				once.Do(func() { broken = true })
			}
			if broken {
				http.Error(w, "gone", http.StatusNotFound)
				return
			}
			next.ServeHTTP(w, r)
		})
	})
	child := childOf(t, target, 5*time.Second, "ann")

	if first := child.Call(t.Context(), "answer", map[string]any{}); first.Kind.String() != "http:404" {
		t.Fatalf("the call the service broke the session on: %q (%v), want http:404", first.Kind, first.Err)
	}
	if again := child.Call(t.Context(), "answer", map[string]any{}); again.Kind != session.Answered {
		t.Errorf("the call after it: %q (%v), want an answer through a new session", again.Kind, again.Err)
	}
	if got := child.Reconnects(); got != 1 {
		t.Errorf("Reconnects() = %d, want 1", got)
	}
}

// A call nobody answers ends at its deadline, told as no answer in time, and
// long before it would end otherwise.
func TestACallNobodyAnswersIsCutOffAtItsTimeout(t *testing.T) {
	t.Parallel()

	never := make(chan struct{})
	target := servicetest.Fake(t, map[string]mcp.ToolHandler{
		"hang": func(context.Context, *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			<-never
			return &mcp.CallToolResult{Content: words("too late")}, nil
		},
	}, nil)
	// The handler is let go of before the listener closes, since closing it
	// waits for every request under way.
	t.Cleanup(func() { close(never) })
	const timeout = 300 * time.Millisecond
	child := childOf(t, target, timeout, "ann")

	answer := child.Call(t.Context(), "hang", map[string]any{})
	if answer.Kind.String() != "no_answer:timeout" {
		t.Errorf("Kind = %q (%v), want no_answer:timeout", answer.Kind, answer.Err)
	}
	if took := answer.Took(); took < timeout || took > 5*time.Second {
		t.Errorf("Took() = %v, want the call cut off at about %v", took, timeout)
	}
}

// Every request of a child is signed in by the child's name, and sent to the
// host the service knows itself by rather than the one it is reached at.
func TestEveryRequestCarriesTheChildsNameAndTheServedHost(t *testing.T) {
	t.Parallel()

	var mu sync.Mutex
	seen := map[string]bool{}
	target := servicetest.Fake(t, map[string]mcp.ToolHandler{
		"answer": says(&mcp.CallToolResult{Content: words("done")}),
	}, func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			mu.Lock()
			seen[r.Host+" "+r.Header.Get("Authorization")] = true
			mu.Unlock()
			next.ServeHTTP(w, r)
		})
	})
	target.Host = "localhost:8080"
	child := childOf(t, target, 5*time.Second, "ann")

	if answer := child.Call(t.Context(), "answer", map[string]any{}); answer.Kind != session.Answered {
		t.Fatalf("Call() = %q (%v), want an answer", answer.Kind, answer.Err)
	}
	mu.Lock()
	defer mu.Unlock()
	if len(seen) != 1 || !seen["localhost:8080 Bearer ann"] {
		t.Errorf("requests went out as %v, want every one to localhost:8080 as Bearer ann", seen)
	}
}

// A child signed in with an account a parent signed in sends the access token
// the account gives at each request — a renewed one once it is renewed — and
// never the name the account is kept under. One whose account cannot give a
// token sends nothing, and its call is told as unsigned.
func TestAChildSignedInSendsTheAccountsToken(t *testing.T) {
	t.Parallel()

	var mu sync.Mutex
	var bearers []string
	target := servicetest.Fake(t, map[string]mcp.ToolHandler{
		"answer": says(&mcp.CallToolResult{Content: words("done")}),
	}, func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			mu.Lock()
			bearers = append(bearers, r.Header.Get("Authorization"))
			mu.Unlock()
			next.ServeHTTP(w, r)
		})
	})
	service := session.Open(target, 5*time.Second)
	defer service.Close()

	tokens := []string{"access-1", "access-2"}
	given := 0
	signedIn := service.SignedIn("parent", func() (string, error) {
		mu.Lock()
		defer mu.Unlock()
		token := tokens[min(given, len(tokens)-1)]
		given++
		return token, nil
	})
	defer func() { _ = signedIn.Close() }()
	if answer := signedIn.Call(t.Context(), "answer", map[string]any{}); answer.Kind != session.Answered {
		t.Fatalf("Call() = %q (%v), want an answer", answer.Kind, answer.Err)
	}
	mu.Lock()
	sent := slices.Clone(bearers)
	mu.Unlock()
	if len(sent) < 2 || sent[0] != "Bearer access-1" || sent[len(sent)-1] != "Bearer access-2" {
		t.Errorf("requests went out as %q, want each with the token the account gave then", sent)
	}

	nobody := service.SignedIn("nobody", func() (string, error) { return "", nil })
	defer func() { _ = nobody.Close() }()
	mu.Lock()
	bearers = nil
	mu.Unlock()
	_ = nobody.Call(t.Context(), "answer", map[string]any{})
	mu.Lock()
	unsigned := slices.Clone(bearers)
	mu.Unlock()
	if len(unsigned) == 0 || slices.ContainsFunc(unsigned, func(sent string) bool { return sent != "" }) {
		t.Errorf("an account with no token sent %q, want requests with no Authorization at all", unsigned)
	}

	lapsed := service.SignedIn("lapsed", func() (string, error) { return "", errors.New("the refresh token has ended") })
	defer func() { _ = lapsed.Close() }()
	before := service.Requests()
	answer := lapsed.Call(t.Context(), "answer", map[string]any{})
	if answer.Kind.String() != "no_answer:unsigned" || service.Requests() != before {
		t.Errorf("Call() = %q after %d requests more, want no_answer:unsigned and nothing sent",
			answer.Kind, service.Requests()-before)
	}
}

// A request counts once the service has answered it: one that never reached
// the service is none the platform bills for.
func TestOnlyRequestsTheServiceAnsweredAreCounted(t *testing.T) {
	t.Parallel()

	answering := servicetest.Fake(t, map[string]mcp.ToolHandler{
		"answer": says(&mcp.CallToolResult{Content: words("done")}),
	}, nil)
	gone := httptest.NewServer(http.NotFoundHandler())
	gone.Close()

	for _, test := range []struct {
		name   string
		target session.Target
		want   int
	}{
		// The handshake, and the call.
		{"a service that answers", answering, 2},
		{"a service nobody listens at", session.Target{URL: gone.URL}, 0},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			service := session.Open(test.target, 5*time.Second)
			defer service.Close()
			child := service.Child("ann")
			defer func() { _ = child.Close() }()

			answer := child.Call(t.Context(), "answer", map[string]any{})
			if got := service.Requests(); got != test.want {
				t.Errorf("Requests() = %d after a call told as %q, want %d", got, answer.Kind, test.want)
			}
		})
	}
}

// A fetch is told by its status: served, held back by a pace, a status the
// service should not give — a redirect among them, which is not followed — or
// nobody there to answer. It goes to the host the service knows itself by,
// from the address given, signed in by nobody, and counts once answered.
func TestAFetchIsToldByItsStatus(t *testing.T) {
	t.Parallel()

	var mu sync.Mutex
	var seen []string
	mux := http.NewServeMux()
	mux.HandleFunc("/document", func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		seen = append(seen, r.Host+" from "+r.Header.Get("X-Forwarded-For")+" as "+r.Header.Get("Authorization"))
		mu.Unlock()
		_, _ = io.WriteString(w, "{}")
	})
	mux.HandleFunc("/paced", func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusTooManyRequests) })
	mux.HandleFunc("/down", func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusServiceUnavailable) })
	mux.HandleFunc("/moved", func(w http.ResponseWriter, r *http.Request) { http.Redirect(w, r, "/document", http.StatusFound) })
	answering := httptest.NewServer(mux)
	defer answering.Close()
	gone := httptest.NewServer(http.NotFoundHandler())
	gone.Close()

	for _, test := range []struct {
		target        session.Target
		path, want    string
		wantRequests  int
		answeredByWho string
	}{
		{session.Target{URL: answering.URL, Host: "localhost:8080"}, "/document", "ok", 1, "a service that serves it"},
		{session.Target{URL: answering.URL}, "/paced", "limited:paced", 1, "a pace"},
		{session.Target{URL: answering.URL}, "/down", "http:503", 1, "a service that is down"},
		{session.Target{URL: answering.URL}, "/moved", "http:302", 1, "a redirect, not followed"},
		{session.Target{URL: gone.URL}, "/document", "no_answer:refused", 0, "nobody"},
	} {
		t.Run(test.answeredByWho, func(t *testing.T) {
			service := session.Open(test.target, 5*time.Second)
			defer service.Close()

			answer := service.Fetch(t.Context(), test.path, "203.0.113.7")
			if answer.Kind.String() != test.want || service.Requests() != test.wantRequests {
				t.Errorf("Fetch(%q) = %q (%v) after %d requests, want %q after %d",
					test.path, answer.Kind, answer.Err, service.Requests(), test.want, test.wantRequests)
			}
		})
	}
	mu.Lock()
	defer mu.Unlock()
	if len(seen) != 1 || seen[0] != "localhost:8080 from 203.0.113.7 as " {
		t.Errorf("the document was asked for as %q, want once, of localhost:8080 from 203.0.113.7 and signed in by nobody", seen)
	}
}

// A handshake the library finished at an older version, after the service
// refused the newest, is not a session of the service's: the call is told as
// such, and nothing is called through it.
func TestAHandshakeThatFellBackIsNoSession(t *testing.T) {
	t.Parallel()

	target := servicetest.Fake(t, map[string]mcp.ToolHandler{
		"answer": says(&mcp.CallToolResult{Content: words("done")}),
	}, func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Header.Get("Mcp-Method") != "server/discover" {
				next.ServeHTTP(w, r)
				return
			}
			refuseAsAPace(t, w, r)
		})
	})
	child := childOf(t, target, 5*time.Second, "ann")

	answer := child.Call(t.Context(), "answer", map[string]any{})
	if answer.Kind.Class != session.Handshake || answer.Result != nil {
		t.Errorf("Call() = %q with result %v, want the handshake told and nothing called", answer.Kind, answer.Result)
	}
	if got := child.Reconnects(); got != 0 {
		t.Errorf("Reconnects() = %d, want 0: no session was ever opened to be opened again", got)
	}
}

// A message a pace held back with an error of the protocol is told as paced,
// as a tool call held back by a pace is; the same code with other words is an
// error of the protocol.
func TestAPacesErrorOfTheProtocolIsToldAsPaced(t *testing.T) {
	t.Parallel()

	target := servicetest.Fake(t, map[string]mcp.ToolHandler{
		"answer": says(&mcp.CallToolResult{Content: words("done")}),
	}, func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Header.Get("Mcp-Name") == "held" {
				refuseAsAPace(t, w, r)
				return
			}
			next.ServeHTTP(w, r)
		})
	})
	child := childOf(t, target, 5*time.Second, "ann")

	if answer := child.Call(t.Context(), "held", map[string]any{}); answer.Kind != session.Paced {
		t.Errorf("Call(held).Kind = %q (%v), want %q", answer.Kind, answer.Err, session.Paced)
	}
}

// A call the run abandons, because it was asked to stop, is told as stopped:
// nothing the service said or failed to say.
func TestACallTheRunAbandonedIsToldAsStopped(t *testing.T) {
	t.Parallel()

	target := servicetest.Fake(t, map[string]mcp.ToolHandler{
		"answer": says(&mcp.CallToolResult{Content: words("done")}),
	}, nil)
	child := childOf(t, target, 5*time.Second, "ann")
	stopped, stop := context.WithCancel(t.Context())
	stop()

	if answer := child.Call(stopped, "answer", map[string]any{}); answer.Kind.Class != session.Stopped {
		t.Errorf("Call().Kind = %q (%v), want the call told as stopped", answer.Kind, answer.Err)
	}
}

// refuseAsAPace answers a request of the protocol the way a pace refuses one
// that is not a tool call: an error of the protocol's, too many requests.
func refuseAsAPace(t *testing.T, w http.ResponseWriter, r *http.Request) {
	t.Helper()

	body, err := io.ReadAll(r.Body)
	if err != nil {
		t.Errorf("read the request: %v", err)
		return
	}
	message, err := jsonrpc.DecodeMessage(body)
	request, isRequest := message.(*jsonrpc.Request)
	if err != nil || !isRequest {
		t.Errorf("the request is %T (%v), want a request of the protocol", message, err)
		return
	}
	refusal, err := jsonrpc.EncodeMessage(&jsonrpc.Response{
		ID: request.ID, Error: &jsonrpc.Error{Code: -32000, Message: "too many requests; try again in a moment"},
	})
	if err != nil {
		t.Errorf("encode the refusal: %v", err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(refusal)
}

// The service's own sentence for a call a pace held back is told by name. An
// account's pace counts the handshake too, so at six calls a minute the
// handshake and one call go through and the second call is held back.
func TestThePacedRefusalIsToldApart(t *testing.T) {
	t.Parallel()

	target := servicetest.Start(t, "MATHTRAIL_RATE_USER_PER_MIN=6")
	child := childOf(t, target, 10*time.Second, "greedy")

	if first := child.Call(t.Context(), "get_profile", map[string]any{}); first.Kind != session.Answered {
		t.Fatalf("the first call: %q (%v), want an answer", first.Kind, first.Err)
	}
	if again := child.Call(t.Context(), "get_profile", map[string]any{}); again.Kind != session.Paced {
		t.Errorf("the call past the pace: %q, %q, want %q", again.Kind, again.Text(), session.Paced)
	}
}

// The service's own sentence for a task the sandbox had no slot for is told by
// name. One slot, held by a solver only its clock stops, and a second task
// handed in meanwhile: one of the two is turned away.
func TestTheBusyRefusalIsToldApart(t *testing.T) {
	t.Parallel()

	target := servicetest.Start(t,
		"MATHTRAIL_SOLVER_CONCURRENCY=1", "MATHTRAIL_SOLVER_TIMEOUT=1s", "MATHTRAIL_SOLVER_WAIT=50ms",
		"MATHTRAIL_SOLVER_STEPS=1000000000000")
	// A task for no open request, whose solver runs before anything else is
	// looked at, and runs until its clock stops it.
	stale := map[string]any{
		"request_id": "req_00000000-0000-0000-0000-000000000000",
		"task":       map[string]any{"options": map[string]string{"A": "4", "B": "5", "C": "6", "D": "8", "E": "12"}},
		"solver":     "def solve(options):\n    while True:\n        pass\n    return []\n",
		"self_check": map[string]any{},
	}
	one, another := childOf(t, target, 10*time.Second, "one"), childOf(t, target, 10*time.Second, "another")

	answers := make([]session.Answer, 2)
	var both sync.WaitGroup
	for i, child := range []*session.Child{one, another} {
		both.Go(func() { answers[i] = child.Call(context.Background(), "submit_task", stale) })
	}
	both.Wait()

	busy := 0
	for _, answer := range answers {
		if answer.Kind == session.Busy {
			busy++
		}
	}
	if busy != 1 {
		t.Errorf("answers %q and %q, want one of them %q", answers[0].Kind, answers[1].Kind, session.Busy)
	}
	if !strings.Contains(answers[0].Text()+answers[1].Text(), "no attempt was spent") {
		t.Errorf("answers %q and %q, want the busy one to say no attempt was spent", answers[0].Text(), answers[1].Text())
	}
}

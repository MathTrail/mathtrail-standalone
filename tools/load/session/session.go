// Package session reaches the MathTrail service the way a chat host does:
// through the protocol library's own client, one session for each child of a
// run, every call with a deadline of its own and every answer told apart by
// what it says. Nothing here writes a message of the protocol by hand.
package session

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Protocol is the one version of the protocol the service speaks. A session
// that ends up at another has not reached the service the way a host does.
const Protocol = "2026-07-28"

// clientName is what the tool calls itself when it connects: a name no chat
// host goes by, so the service counts its calls apart from theirs.
const clientName = "mathtrail-load"

// Target is the service as the tool reaches it.
type Target struct {
	// URL is where the tool connects: scheme, host and port.
	URL string
	// Host is the host the service knows itself by, sent with every request,
	// since the service answers no other. Empty sends the URL's own.
	Host string
}

// Service is the service under load as every child of a run reaches it: one
// pool of connections, one deadline for each call, and a count of the
// requests they sent.
type Service struct {
	target   Target
	pool     *http.Transport
	timeout  time.Duration
	requests atomic.Int64
}

// Requests is how many requests the children of the run have sent the
// service: every call, and every request the library sent on its own — the
// handshakes, a cancellation. The platform bills for each of them.
func (s *Service) Requests() int { return int(s.requests.Load()) }

// Open is the service at the target, each of whose calls may take as long as
// the timeout given and no longer.
func Open(target Target, timeout time.Duration) *Service {
	// The standard transport's settings — its dialer, its timeouts, HTTP/2 —
	// are the ones a host's client would have, where they can be had.
	pool := &http.Transport{}
	if standard, isStandard := http.DefaultTransport.(*http.Transport); isStandard {
		pool = standard.Clone()
	}
	// A run keeps many calls in flight at once, and a pool that holds only a
	// few idle connections would open and close one for nearly every call,
	// which is load of the tool's own making. Every call goes to one host, so
	// the pool as a whole keeps as many as that host may.
	pool.MaxIdleConnsPerHost = 1024
	pool.MaxIdleConns = pool.MaxIdleConnsPerHost
	return &Service{target: target, pool: pool, timeout: timeout}
}

// Close lets go of the connections the children of a run left open.
func (s *Service) Close() { s.pool.CloseIdleConnections() }

// Child is a child of the run, signed in by the name given: the development
// sign-in makes the account dev-<name> of it, with a profile and a pace of its
// own.
func (s *Service) Child(name string) *Child {
	return &Child{
		name:    name,
		service: s,
		client:  mcp.NewClient(&mcp.Implementation{Name: clientName, Version: "1"}, nil),
		http:    &http.Client{Transport: signing{name: name, host: s.target.Host, next: s.pool, sent: &s.requests}},
	}
}

// Child is one child of a run: the name they are signed in by, and the session
// their calls go through, opened when a call first needs it and again after a
// call found it broken.
type Child struct {
	name    string
	service *Service
	client  *mcp.Client
	http    *http.Client

	mu sync.Mutex
	// session is the session new calls go through.
	session *mcp.ClientSession
	// opening is closed when the session being opened is ready or has failed,
	// and nil while none is being opened. The calls that need the session in
	// the meantime wait on it, each only as long as its own deadline allows.
	opening chan struct{}
	// retired are the sessions calls found broken. They are not closed until
	// the child is, so that a call of the child's still under way on one ends
	// the way the service ends it rather than the way a close would.
	retired []*mcp.ClientSession
	// opened is how many sessions at the service's version have been opened.
	opened int
}

// Name is the name the child is signed in by.
func (c *Child) Name() string { return c.name }

// Reconnects is how many times the child's session had to be opened again
// after a call found it broken.
func (c *Child) Reconnects() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return max(c.opened-1, 0)
}

// Close ends every session of the child's.
func (c *Child) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	var failed []error
	for _, session := range append(c.retired, c.session) {
		if session == nil {
			continue
		}
		if err := session.Close(); err != nil {
			failed = append(failed, err)
		}
	}
	c.session, c.retired = nil, nil
	if err := errors.Join(failed...); err != nil {
		return fmt.Errorf("session: close the sessions of %s: %w", c.name, err)
	}
	return nil
}

// Call calls a tool once, as the child, and tells what came back. The session
// is opened first when the child has none, and the call and the opening share
// the one deadline.
func (c *Child) Call(ctx context.Context, tool string, arguments any) Answer {
	ctx, cancel := context.WithTimeout(ctx, c.service.timeout)
	defer cancel()
	seen := &worst{}
	ctx = context.WithValue(ctx, worstKey{}, seen)

	answer := Answer{Tool: tool, Started: time.Now()}
	session, err := c.open(ctx)
	if err == nil {
		answer.Result, err = session.CallTool(ctx, &mcp.CallToolParams{Name: tool, Arguments: arguments})
	}
	answer.Ended = time.Now()
	answer.Status = seen.status()
	answer.Err = err
	answer.Kind = kindOf(answer.Result, err, answer.Status)
	if session != nil && broke(err, answer.Status) {
		c.retire(session)
	}
	return answer
}

// open is the child's session: the one already open, or a new one. One call
// opens it while the others that need it wait, and a call that runs out of
// time while it waits ends then, as unanswered. A session the handshake left
// at another version than the service's is closed again at once and counts as
// no session.
func (c *Child) open(ctx context.Context) (*mcp.ClientSession, error) {
	for {
		c.mu.Lock()
		if c.session != nil {
			session := c.session
			c.mu.Unlock()
			return session, nil
		}
		if c.opening == nil {
			c.opening = make(chan struct{})
			c.mu.Unlock()
			return c.connect(ctx)
		}
		opening := c.opening
		c.mu.Unlock()
		select {
		case <-opening:
			// Ready, or failed: look again, and open one if it failed.
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
}

// connect opens a session for the child and puts it in use, and lets every
// call waiting for it go on, whether it was opened or not.
func (c *Child) connect(ctx context.Context) (*mcp.ClientSession, error) {
	session, err := c.client.Connect(ctx, &mcp.StreamableClientTransport{
		Endpoint:   strings.TrimSuffix(c.service.target.URL, "/") + "/mcp",
		HTTPClient: c.http,
		// A service without sessions has no stream to open outside a call, and
		// the library's own retries would hide what the service answered.
		DisableStandaloneSSE: true,
		MaxRetries:           -1,
	}, &mcp.ClientSessionOptions{ProtocolVersion: Protocol})
	if err == nil {
		if version := session.InitializeResult().ProtocolVersion; version != Protocol {
			_ = session.Close()
			session, err = nil, fellBack{version: version}
		}
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	close(c.opening)
	c.opening = nil
	if err != nil {
		return nil, err
	}
	c.session = session
	c.opened++
	return session, nil
}

// retire takes the session a call found broken out of use, so that the next
// call opens another; a session opened since in its place stays in use.
func (c *Child) retire(broken *mcp.ClientSession) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.session == broken {
		c.retired = append(c.retired, broken)
		c.session = nil
	}
}

// Answer is one call as the child saw it.
type Answer struct {
	// Tool is the tool called.
	Tool string
	// Kind is what the answer came to.
	Kind Kind
	// Status is the highest HTTP status the requests of the call got back, and
	// zero when none got an answer.
	Status int
	// Result is what the tool said, nil when the call got no result.
	Result *mcp.CallToolResult
	// Started and Ended are when the call went out and when it came back.
	Started, Ended time.Time
	// Err is why the call got no result, nil when it got one.
	Err error
}

// Took is how long the call took.
func (a *Answer) Took() time.Duration { return a.Ended.Sub(a.Started) }

// Text is the words of the result, empty when there is none.
func (a *Answer) Text() string {
	if a.Result == nil {
		return ""
	}
	var words []string
	for _, block := range a.Result.Content {
		if text, isText := block.(*mcp.TextContent); isText {
			words = append(words, text.Text)
		}
	}
	return strings.Join(words, "\n")
}

// Payload reads the machine-readable part of the result into what the caller
// expects of it.
func (a *Answer) Payload(into any) error {
	if a.Result == nil || a.Result.StructuredContent == nil {
		return errors.New("session: the answer carries no payload")
	}
	raw, err := json.Marshal(a.Result.StructuredContent)
	if err != nil {
		return fmt.Errorf("session: encode the payload: %w", err)
	}
	if err := json.Unmarshal(raw, into); err != nil {
		return fmt.Errorf("session: read the payload: %w", err)
	}
	return nil
}

// fellBack is a handshake the library finished at an older version, after the
// service refused the one the tool asked for.
type fellBack struct{ version string }

func (f fellBack) Error() string {
	return fmt.Sprintf("session: the handshake ended at %q, not at %q", f.version, Protocol)
}

// signing sends every request of a child signed in by their name, to the host
// the service knows itself by, counts it, and keeps the worst status a call's
// requests got back.
type signing struct {
	name, host string
	next       http.RoundTripper
	sent       *atomic.Int64
}

func (s signing) RoundTrip(r *http.Request) (*http.Response, error) {
	r = r.Clone(r.Context())
	r.Header.Set("Authorization", "Bearer "+s.name)
	if s.host != "" {
		r.Host = s.host
	}
	resp, err := s.next.RoundTrip(r)
	if err == nil {
		// Counted once the service has answered: a request that never reached
		// it is none the platform bills for.
		s.sent.Add(1)
		if seen, watched := r.Context().Value(worstKey{}).(*worst); watched {
			seen.saw(resp.StatusCode)
		}
	}
	return resp, err
}

// worstKey is where a call's context carries what its requests got back.
type worstKey struct{}

// worst is the highest HTTP status the requests of one call got back. A call
// is more than one request when a session is opened for it, and the one that
// failed is the one worth telling.
type worst struct {
	mu   sync.Mutex
	code int
}

func (w *worst) saw(code int) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.code = max(w.code, code)
}

func (w *worst) status() int {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.code
}

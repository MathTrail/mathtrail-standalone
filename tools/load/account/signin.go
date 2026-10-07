package account

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/modelcontextprotocol/go-sdk/auth"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/modelcontextprotocol/go-sdk/oauthex"
	"golang.org/x/oauth2"

	"github.com/MathTrail/mathtrail-standalone/tools/load/session"
)

// clientName is what the load calls itself to the service: the consent screen
// shows it to the parent.
const clientName = "MathTrail load"

// callbackPath is where the service sends the browser back to, on this
// computer.
const callbackPath = "/callback"

// signInTime is how long a parent has to sign in once the address is shown,
// which is as long as the service keeps a sign-in under way.
const signInTime = 10 * time.Minute

// Browser is the parent signing in, as the sign-in reaches them.
type Browser struct {
	// Show shows the parent the address their browser is to open.
	Show func(address string)
	// Pasted is where the parent may paste the address their browser was
	// sent back to, one to a line, when it could not reach this computer; nil
	// when nothing can be pasted.
	Pasted io.Reader
}

// SignIn signs an account in on the service at the URL given, the way a chat
// host signs a parent in, and keeps it under the name given in the directory.
// It is the protocol library's own sign-in: the service's endpoint asks for
// one, the load registers itself as a client, and the parent opens the
// address shown, signs in with Google and allows the load on the service's
// consent screen. The service then sends the browser back to this computer,
// on the port given — any free one for zero — or the parent pastes the
// address it was sent to. A session opens once the tokens are in hand, so an
// account is kept only when the service has let it in.
func SignIn(ctx context.Context, dir, name, service string, browser Browser, port int) (*Account, error) {
	where, err := shelfOf(dir, name)
	if err != nil {
		return nil, err
	}
	back, err := listen(ctx, port)
	if err != nil {
		return nil, err
	}
	defer back.close()
	if browser.Pasted != nil {
		go back.read(browser.Pasted)
	}

	// The tokens arrive inside the library's own call, which hands them
	// over here.
	var signedIn atomic.Pointer[handedOver]
	handler, err := auth.NewAuthorizationCodeHandler(&auth.AuthorizationCodeHandlerConfig{
		DynamicClientRegistrationConfig: &auth.DynamicClientRegistrationConfig{
			Metadata: &oauthex.ClientRegistrationMetadata{
				RedirectURIs:            []string{back.address},
				TokenEndpointAuthMethod: "none",
				GrantTypes:              []string{"authorization_code", "refresh_token"},
				ResponseTypes:           []string{"code"},
				ClientName:              clientName,
			},
		},
		AuthorizationCodeFetcher: func(ctx context.Context, args *auth.AuthorizationArgs) (*auth.AuthorizationResult, error) {
			return back.signIn(ctx, browser.Show, args.URL)
		},
		NewTokenSource: func(ctx context.Context, config *oauth2.Config, token *oauth2.Token) (oauth2.TokenSource, error) {
			k := kept{Service: serviceOf(service), ClientID: config.ClientID, TokenURL: config.Endpoint.TokenURL, Token: token}
			if !k.whole() {
				return nil, errors.New("account: the service gave no token to renew the account with")
			}
			tokens := &latest{next: config.TokenSource(ctx, token), token: token}
			signedIn.Store(&handedOver{kept: k, tokens: tokens})
			return tokens, nil
		},
		Client: client(),
	})
	if err != nil {
		return nil, fmt.Errorf("account: set up the sign-in: %w", err)
	}

	ctx, cancel := context.WithTimeout(ctx, signInTime)
	defer cancel()
	opened, err := mcp.NewClient(&mcp.Implementation{Name: clientName, Version: "1"}, nil).Connect(ctx,
		&mcp.StreamableClientTransport{
			Endpoint:             serviceOf(service) + "/mcp",
			HTTPClient:           client(),
			OAuthHandler:         handler,
			DisableStandaloneSSE: true,
			MaxRetries:           -1,
		}, &mcp.ClientSessionOptions{ProtocolVersion: session.Protocol})
	if err != nil {
		return nil, fmt.Errorf("account: sign %s in on %s: %w", name, serviceOf(service), err)
	}
	_ = opened.Close()
	got := signedIn.Load()
	if got == nil {
		return nil, fmt.Errorf("account: %s let the session open with no sign-in: it is no deployed service", serviceOf(service))
	}
	k := got.kept
	k.Token = got.tokens.last()
	if unkept := where.write(&k); unkept != nil {
		return nil, unkept
	}
	return k.account(renewing(), where), nil
}

// handedOver is what the library hands over once the parent has signed in:
// the account as it is to be kept, and the tokens the session opens with.
type handedOver struct {
	kept   kept
	tokens *latest
}

// latest is the tokens of a sign-in under way, and the latest of them, which
// the account is kept with once a session has opened: one the library renewed
// meanwhile is the one the service holds good.
type latest struct {
	next oauth2.TokenSource

	mu    sync.Mutex
	token *oauth2.Token
}

func (l *latest) Token() (*oauth2.Token, error) {
	token, err := l.next.Token()
	if err != nil {
		return nil, err
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	l.token = token
	return token, nil
}

// last is the latest token the library was given.
func (l *latest) last() *oauth2.Token {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.token
}

// errOnce is a sign-in asked for after the one the parent went through.
var errOnce = errors.New("account: the service asked for a sign-in again after the parent's")

// cameBack is what the browser was sent back with: a code to exchange for the
// tokens, or why the sign-in did not go through.
type cameBack struct {
	result *auth.AuthorizationResult
	err    error
}

// way is the way back from the browser: the address on this computer the
// service sends the browser to, and what the browser — or the parent, pasting
// the address — brought back first.
type way struct {
	address string
	server  *http.Server
	once    sync.Once
	back    chan cameBack

	mu sync.Mutex
	// sent is whether the parent was sent to sign in, and with what state,
	// which the service sends back with the browser; refused is why the
	// sign-in they went through did not let the load in.
	sent    bool
	state   string
	refused error
}

// signIn sends the parent to the address given, the first time it is asked,
// and is what the browser brought back. The library asks again whenever a
// request is refused for want of a sign-in, and after a sign-in that went
// wrong it does: a parent signs in once, so every time after that is answered
// with what the parent's own sign-in came to.
func (w *way) signIn(ctx context.Context, show func(address string), address string) (*auth.AuthorizationResult, error) {
	sending, err := url.Parse(address)
	if err != nil {
		return nil, fmt.Errorf("account: read the address of the sign-in: %w", err)
	}
	w.mu.Lock()
	if w.sent {
		defer w.mu.Unlock()
		if w.refused != nil {
			return nil, w.refused
		}
		return nil, errOnce
	}
	w.sent, w.state = true, sending.Query().Get("state")
	w.mu.Unlock()

	show(address)
	result, err := w.wait(ctx)
	if err != nil {
		w.mu.Lock()
		w.refused = err
		w.mu.Unlock()
	}
	return result, err
}

// listen opens the way back on the port given of this computer's loopback,
// where only this computer reaches it.
func listen(ctx context.Context, port int) (*way, error) {
	var config net.ListenConfig
	listener, err := config.Listen(ctx, "tcp", net.JoinHostPort("127.0.0.1", strconv.Itoa(port)))
	if err != nil {
		return nil, fmt.Errorf("account: listen for the browser on port %d: %w", port, err)
	}
	w := &way{
		address: "http://" + listener.Addr().String() + callbackPath,
		back:    make(chan cameBack, 1),
	}
	w.server = &http.Server{Handler: http.HandlerFunc(w.serve), ReadHeaderTimeout: timeout}
	go func() { _ = w.server.Serve(listener) }()
	return w, nil
}

// close closes the way back.
func (w *way) close() { _ = w.server.Close() }

// serve takes the browser sent back, and tells the parent what to do next.
func (w *way) serve(rw http.ResponseWriter, r *http.Request) {
	if r.URL.Path != callbackPath {
		http.NotFound(rw, r)
		return
	}
	query := r.URL.Query()
	if !w.awaited(query) {
		// Anything else on this computer may ask here too, and only the
		// sign-in under way is waited for.
		http.Error(rw, "This is not the sign-in the load is waiting for.", http.StatusBadRequest)
		return
	}
	came := cameBackWith(query)
	// The page goes out whole before the sign-in hears what the browser
	// brought: once it has, it may be over, and close the way back, before a
	// page written after it reaches the browser. The code is still to be
	// exchanged and the session still to open, so only a refusal is certain
	// here, and how the sign-in ends is the terminal's to say.
	page := "Close this tab and go back to the terminal: it says whether the sign-in went through.\n"
	if came.err != nil {
		page = "The sign-in did not go through. Go back to the terminal to see why.\n"
	}
	rw.Header().Set("Content-Type", "text/plain; charset=utf-8")
	rw.Header().Set("Content-Length", strconv.Itoa(len(page)))
	_, _ = io.WriteString(rw, page)
	_ = http.NewResponseController(rw).Flush()
	w.bring(came)
}

// read takes an address pasted, one to a line, and brings back the first that
// is the address the browser of the sign-in under way was sent back to.
func (w *way) read(pasted io.Reader) {
	lines := bufio.NewScanner(pasted)
	for lines.Scan() {
		address, err := url.Parse(strings.TrimSpace(lines.Text()))
		if err != nil {
			continue
		}
		if query := address.Query(); w.awaited(query) {
			w.bring(cameBackWith(query))
			return
		}
	}
}

// awaited says whether an address the browser came back to is the one the
// sign-in under way waits for: the service sends it back with the state the
// parent was sent with, a code or a refusal alike.
func (w *way) awaited(query url.Values) bool {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.state != "" && query.Get("state") == w.state
}

// bring brings back what the browser came back with, the first time only: the
// browser and a paste may both bring the same.
func (w *way) bring(came cameBack) {
	w.once.Do(func() { w.back <- came })
}

// wait is what was brought back, once it is.
func (w *way) wait(ctx context.Context) (*auth.AuthorizationResult, error) {
	select {
	case came := <-w.back:
		return came.result, came.err
	case <-ctx.Done():
		return nil, fmt.Errorf("account: nobody signed in within %v: %w", signInTime, ctx.Err())
	}
}

// cameBackWith is what the parameters of the address the browser was sent
// back to say. A refusal names its reason in the protocol's own words.
func cameBackWith(query url.Values) cameBack {
	switch {
	case query.Get("error") != "":
		return cameBack{err: fmt.Errorf("account: the sign-in was refused: %s", query.Get("error"))}
	case query.Get("code") == "":
		return cameBack{err: errors.New("account: the address the browser came back to carries no code")}
	}
	return cameBack{result: &auth.AuthorizationResult{
		Code:  query.Get("code"),
		State: query.Get("state"),
		Iss:   query.Get("iss"),
	}}
}

package cimd

import (
	"context"
	"crypto/x509"
	"errors"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/netip"
	"slices"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// fakeClient is a server that answers as a client's address would, and
// counts the connections and the requests that reached it: a case that must
// reach nothing reads both as zero.
type fakeClient struct {
	*httptest.Server
	connections atomic.Int32
	closed      atomic.Int32
	requests    atomic.Int32
}

// newFakeClient serves an answer over TLS on the loopback, under the
// certificate the test library issues for example.com and its subdomains.
func newFakeClient(t *testing.T, answer http.HandlerFunc) *fakeClient {
	t.Helper()

	client := &fakeClient{}
	client.Server = httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		client.requests.Add(1)
		answer(w, r)
	}))
	client.Config.ConnState = func(_ net.Conn, state http.ConnState) {
		switch state {
		case http.StateNew:
			client.connections.Add(1)
		case http.StateClosed, http.StateHijacked:
			client.closed.Add(1)
		default:
		}
	}
	client.StartTLS()
	t.Cleanup(client.Close)
	return client
}

// dial connects to the address asked for on the port the fake server listens
// on, in place of https's own: that is the one port a client's address may
// have, and not one a test can listen on.
func (c *fakeClient) dial(ctx context.Context, network, address string) (net.Conn, error) {
	host, _, err := net.SplitHostPort(address)
	if err != nil {
		return nil, err
	}
	_, port, err := net.SplitHostPort(c.Listener.Addr().String())
	if err != nil {
		return nil, err
	}
	return (&net.Dialer{}).DialContext(ctx, network, net.JoinHostPort(host, port))
}

// servingItself answers with a document that names itself by the address it
// was asked at, as a client's document does, kept for as long as the header
// says.
func servingItself(cacheControl string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if cacheControl != "" {
			w.Header().Set("Cache-Control", cacheControl)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(claudeLike("https://" + r.Host + r.URL.RequestURI())))
	}
}

// clock is the time a case sets.
type clock struct {
	mu sync.Mutex
	at time.Time
}

func (c *clock) now() time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.at
}

func (c *clock) advance(by time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.at = c.at.Add(by)
}

// publicOrLoopback is the service's own rule with one exception: the loopback
// the fake servers live on.
func publicOrLoopback(addr netip.Addr) bool {
	return public(addr) || addr.Unmap().IsLoopback()
}

// fetcherFor builds a fetcher for a case: every name resolves to the
// addresses given, what may be dialled is what allowed allows, the fake
// server answers at each of them, its certificate is believed, and an attempt
// takes at most timeout.
func fetcherFor(t *testing.T, server *fakeClient, allowed func(netip.Addr) bool, timeout time.Duration,
	resolved ...netip.Addr) (*fetcher, *clock) {
	t.Helper()

	moment := &clock{at: someDay}
	fetcher := newFetcher(reach{
		resolve: func(context.Context, string) ([]netip.Addr, error) { return resolved, nil },
		allowed: allowed,
		dial:    server.dial,
		roots:   rootsOf(server),
	}, timeout, moment.now)
	t.Cleanup(fetcher.client.CloseIdleConnections)
	return fetcher, moment
}

var loopback = netip.MustParseAddr("127.0.0.1")

// A document at a public address is fetched, read for what the sign-in needs,
// and kept: the next request for it reaches nothing until its time is up.
func TestADocumentIsFetchedAndKept(t *testing.T) {
	t.Parallel()

	server := newFakeClient(t, servingItself("public, max-age=300"))
	fetcher, moment := fetcherFor(t, server, publicOrLoopback, time.Second, loopback)
	address := "https://client.example.com/oauth/mcp-oauth-client-metadata"

	first, err := fetcher.Fetch(t.Context(), address)
	if err != nil {
		t.Fatalf("Fetch() error = %v, want nil", err)
	}
	if first.Cached || first.Document.ClientID != address || first.Document.ClientName != "Claude" {
		t.Errorf("Fetch() = %+v, want Claude's document from the network", first)
	}

	moment.advance(5*time.Minute - time.Second)
	if again, err := fetcher.Fetch(t.Context(), address); err != nil || !again.Cached {
		t.Errorf("Fetch() again = %+v, %v; want the kept document", again, err)
	}
	moment.advance(time.Second)
	if later, err := fetcher.Fetch(t.Context(), address); err != nil || later.Cached {
		t.Errorf("Fetch() past its time = %+v, %v; want it fetched again", later, err)
	}
	if got := server.requests.Load(); got != 2 {
		t.Errorf("%d requests reached the client, want 2: one, then one after the document's time", got)
	}
}

// No address a stranger's name could stand for inside somebody's own network
// is ever dialled: the fetch is refused before a connection is made, whatever
// the name resolves to and however the address is written.
func TestNoTrapIsDialled(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name     string
		resolved []string
	}{
		{"the loopback", []string{"127.0.0.1"}},
		{"the loopback of IPv6", []string{"::1"}},
		{"the loopback written in IPv6", []string{"::ffff:127.0.0.1"}},
		{"a private network", []string{"10.0.0.1"}},
		{"another private network", []string{"192.168.1.1"}},
		{"the platform's metadata server", []string{"169.254.169.254"}},
		{"the metadata server written in IPv6", []string{"::ffff:169.254.169.254"}},
		{"link-local", []string{"fe80::1"}},
		{"unique-local", []string{"fd20:ce::254"}},
		{"the carrier's NAT", []string{"100.64.0.1"}},
		{"NAT64 into a private network", []string{"64:ff9b::a00:1"}},
		{"the unspecified address", []string{"0.0.0.0"}},
		{"a public address beside a private one", []string{"192.0.2.1", "127.0.0.1"}},
		{"a private address beside a public one", []string{"127.0.0.1", "192.0.2.1"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			server := newFakeClient(t, servingItself(""))
			resolved := make([]netip.Addr, 0, len(tc.resolved))
			for _, address := range tc.resolved {
				resolved = append(resolved, netip.MustParseAddr(address))
			}
			fetcher, _ := fetcherFor(t, server, public, time.Second, resolved...)

			_, err := fetcher.Fetch(t.Context(), clientAt)
			if !errors.Is(err, ErrAddress) {
				t.Errorf("Fetch() error = %v, want ErrAddress", err)
			}
			if got := server.connections.Load(); got != 0 {
				t.Errorf("%d connections reached the server, want none", got)
			}
		})
	}
}

// An address a client identifier may not be is refused before its name is
// even looked up.
func TestAnAddressThatIsNotAClientsIsNotLookedUp(t *testing.T) {
	t.Parallel()

	server := newFakeClient(t, servingItself(""))
	var lookups atomic.Int32
	fetcher := newFetcher(reach{
		resolve: func(context.Context, string) ([]netip.Addr, error) {
			lookups.Add(1)
			return []netip.Addr{loopback}, nil
		},
		allowed: publicOrLoopback,
		dial:    server.dial,
	}, time.Second, (&clock{at: someDay}).now)

	for _, address := range []string{
		"http://client.example.com/meta",
		"https://client.example.com",
		"https://client.example.com/",
		"https://user:secret@client.example.com/meta",
		"https://client.example.com/meta#part",
		"https://client.example.com/a/../meta",
		"https://client.example.com:8443/meta",
	} {
		if _, err := fetcher.Fetch(t.Context(), address); !errors.Is(err, ErrClientURL) {
			t.Errorf("Fetch(%q) error = %v, want ErrClientURL", address, err)
		}
	}
	if lookups.Load() != 0 || server.connections.Load() != 0 {
		t.Errorf("%d lookups and %d connections, want none", lookups.Load(), server.connections.Load())
	}
}

// A redirect is an answer, not a direction: it is not followed, wherever it
// points, and the document is refused.
func TestARedirectIsNotFollowed(t *testing.T) {
	t.Parallel()

	server := newFakeClient(t, func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "https://other.example.com/meta", http.StatusFound)
	})
	fetcher, _ := fetcherFor(t, server, publicOrLoopback, time.Second, loopback)

	if _, err := fetcher.Fetch(t.Context(), clientAt); !errors.Is(err, ErrRedirect) {
		t.Errorf("Fetch() error = %v, want ErrRedirect", err)
	}
	// Every name reaches the fake server, so a redirect followed would be one
	// more request there.
	if got := server.requests.Load(); got != 1 {
		t.Errorf("%d requests reached the server, want 1: the one at the address", got)
	}
}

// An answer that could only be the same the second time is not asked for
// again; one that might not be is asked for exactly once more.
func TestOnlyAFailureThatMightPassIsTriedAgain(t *testing.T) {
	t.Parallel()

	stalling := func(w http.ResponseWriter, r *http.Request) { <-r.Context().Done() }
	failing := func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusInternalServerError) }
	missing := func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusNotFound) }
	oversized := func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"client_name":"` + strings.Repeat("a", maxDocument) + `"}`))
	}
	declaredOversized := func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Length", "70000")
		_, _ = w.Write([]byte(strings.Repeat(" ", 70000)))
	}

	for _, tc := range []struct {
		name     string
		answer   http.HandlerFunc
		want     error
		requests int32
	}{
		{"a server that never answers", stalling, ErrUnreachable, 2},
		{"a server that fails", failing, ErrStatus, 2},
		{"no document there", missing, ErrStatus, 1},
		{"a document past the limit", oversized, ErrTooLarge, 1},
		{"a document declared past the limit", declaredOversized, ErrTooLarge, 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			server := newFakeClient(t, tc.answer)
			fetcher, _ := fetcherFor(t, server, publicOrLoopback, time.Second, loopback)

			if _, err := fetcher.Fetch(t.Context(), clientAt); !errors.Is(err, tc.want) {
				t.Errorf("Fetch() error = %v, want %v", err, tc.want)
			}
			if got := server.requests.Load(); got != tc.requests {
				t.Errorf("%d requests reached the client, want %d", got, tc.requests)
			}
		})
	}
}

// A connection whose certificate does not name the address is not believed,
// and not tried again: the certificate would be the same.
func TestACertificateForAnotherNameIsNotBelieved(t *testing.T) {
	t.Parallel()

	server := newFakeClient(t, servingItself(""))
	fetcher, _ := fetcherFor(t, server, publicOrLoopback, time.Second, loopback)

	_, err := fetcher.Fetch(t.Context(), "https://client.example.org/oauth/metadata")
	if !errors.Is(err, ErrUnreachable) {
		t.Errorf("Fetch() error = %v, want ErrUnreachable", err)
	}
	if server.connections.Load() != 1 || server.requests.Load() != 0 {
		t.Errorf("%d connections and %d requests, want one connection and no request",
			server.connections.Load(), server.requests.Load())
	}
}

// A document that is not the client's is refused and never kept: the next
// request for it is a fetch again.
func TestADocumentThatIsNotTheClientsIsNeverKept(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name string
		body func(self string) string
	}{
		{"another client's", func(string) string { return claudeLike("https://client.example.com/other") }},
		{"no redirect", func(self string) string { return `{"client_id":"` + self + `","client_name":"A"}` }},
		{"a secret", func(self string) string {
			return `{"client_id":"` + self + `","client_name":"A","redirect_uris":["https://a.example/cb"],"client_secret":"s"}`
		}},
		{"a shared secret to sign in with", func(self string) string {
			return `{"client_id":"` + self + `","client_name":"A","redirect_uris":["https://a.example/cb"],` +
				`"token_endpoint_auth_method":"client_secret_basic"}`
		}},
		{"not JSON", func(string) string { return "<html>a client</html>" }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			server := newFakeClient(t, func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Cache-Control", "max-age=3600")
				_, _ = w.Write([]byte(tc.body("https://" + r.Host + r.URL.RequestURI())))
			})
			fetcher, _ := fetcherFor(t, server, publicOrLoopback, time.Second, loopback)
			address := clientAt

			for range 2 {
				if _, err := fetcher.Fetch(t.Context(), address); !errors.Is(err, ErrDocument) {
					t.Errorf("Fetch() error = %v, want ErrDocument", err)
				}
			}
			if got := server.requests.Load(); got != 2 {
				t.Errorf("%d requests reached the client, want 2: nothing refused is kept", got)
			}
		})
	}
}

// rootsOf believes the certificate the fake servers are issued.
func rootsOf(server *fakeClient) *x509.CertPool {
	roots := x509.NewCertPool()
	roots.AddCert(server.Certificate())
	return roots
}

// A fetch leaves no connection open to the host a stranger chose: the
// document is kept, and a connection kept beside it would only hold a socket.
func TestNoConnectionOutlivesAFetch(t *testing.T) {
	t.Parallel()

	server := newFakeClient(t, servingItself(""))
	fetcher, _ := fetcherFor(t, server, publicOrLoopback, time.Second, loopback)
	if _, err := fetcher.Fetch(t.Context(), clientAt); err != nil {
		t.Fatalf("Fetch() error = %v, want nil", err)
	}

	deadline := time.Now().Add(2 * time.Second)
	for server.closed.Load() < server.connections.Load() && time.Now().Before(deadline) {
		time.Sleep(10 * time.Millisecond)
	}
	if opened, closed := server.connections.Load(), server.closed.Load(); closed != opened {
		t.Errorf("%d connections opened and %d closed after the fetch, want every one closed", opened, closed)
	}
}

// An address that never answers — a route that swallows every packet — takes
// its share of the attempt and no more, and the next address is dialled in
// the time that is left.
func TestAnAddressThatNeverAnswersLeavesTheNextItsTurn(t *testing.T) {
	t.Parallel()

	server := newFakeClient(t, servingItself(""))
	silent := netip.MustParseAddr("192.0.2.1")
	fetcher := newFetcher(reach{
		resolve: func(context.Context, string) ([]netip.Addr, error) { return []netip.Addr{silent, loopback}, nil },
		allowed: publicOrLoopback,
		dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			if strings.HasPrefix(address, silent.String()+":") {
				<-ctx.Done()
				return nil, ctx.Err()
			}
			return server.dial(ctx, network, address)
		},
		roots: rootsOf(server),
	}, 4*time.Second, (&clock{at: someDay}).now)
	t.Cleanup(fetcher.client.CloseIdleConnections)

	if _, err := fetcher.Fetch(t.Context(), clientAt); err != nil {
		t.Errorf("Fetch() error = %v, want the document from the address that answers", err)
	}
}

// However many addresses a name stands for, the first few are dialled, in the
// order the resolver gave them, and on https's own port: a stranger's name
// cannot have the service knock on a long list of addresses of its choosing.
func TestOnlyTheFirstFewAddressesOfANameAreDialled(t *testing.T) {
	t.Parallel()

	many := make([]netip.Addr, 0, 64)
	for i := range 64 {
		many = append(many, netip.AddrFrom4([4]byte{192, 0, 2, byte(i + 1)}))
	}
	var mu sync.Mutex
	var dialled []string
	fetcher := newFetcher(reach{
		resolve: func(context.Context, string) ([]netip.Addr, error) { return many, nil },
		allowed: public,
		dial: func(_ context.Context, _, address string) (net.Conn, error) {
			mu.Lock()
			defer mu.Unlock()
			dialled = append(dialled, address)
			return nil, errors.New("connection refused")
		},
	}, time.Second, (&clock{at: someDay}).now)

	if _, err := fetcher.Fetch(t.Context(), clientAt); !errors.Is(err, ErrUnreachable) {
		t.Errorf("Fetch() error = %v, want ErrUnreachable", err)
	}
	first := []string{"192.0.2.1:443", "192.0.2.2:443", "192.0.2.3:443", "192.0.2.4:443"}
	mu.Lock()
	defer mu.Unlock()
	if want := slices.Concat(first, first); !slices.Equal(dialled, want) {
		t.Errorf("dialled %v, want %v: the first four, once in each of the two attempts", dialled, want)
	}
}

// Nothing a fetch starts outlives the attempt that started it: not a lookup
// that never answers, nor a connection to a server that takes it and never
// says a word. The HTTP client lets a dial go on after its request has given
// up, so each would otherwise hold a goroutine — and the second a socket too
// — for as long as the stranger's name liked.
func TestNothingAFetchStartsOutlivesItsAttempt(t *testing.T) {
	t.Parallel()

	const attempt = 200 * time.Millisecond
	for _, tc := range []struct {
		name  string
		reach func(t *testing.T, started, over *atomic.Int32) reach
	}{
		{"a lookup that never answers", func(_ *testing.T, started, over *atomic.Int32) reach {
			return reach{
				resolve: func(ctx context.Context, _ string) ([]netip.Addr, error) {
					started.Add(1)
					defer over.Add(1)
					<-ctx.Done()
					return nil, ctx.Err()
				},
				allowed: publicOrLoopback,
				dial:    (&net.Dialer{}).DialContext,
			}
		}},
		{"a server that takes the connection and never says a word", func(t *testing.T, started, over *atomic.Int32) reach {
			silent := silentServer(t, started, over)
			return reach{
				resolve: func(context.Context, string) ([]netip.Addr, error) { return []netip.Addr{loopback}, nil },
				allowed: publicOrLoopback,
				dial: func(ctx context.Context, network, _ string) (net.Conn, error) {
					return (&net.Dialer{}).DialContext(ctx, network, silent)
				},
			}
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var started, over atomic.Int32
			fetcher := newFetcher(tc.reach(t, &started, &over), attempt, (&clock{at: someDay}).now)
			t.Cleanup(fetcher.client.CloseIdleConnections)

			if _, err := fetcher.Fetch(t.Context(), clientAt); !errors.Is(err, ErrUnreachable) {
				t.Errorf("Fetch() error = %v, want ErrUnreachable", err)
			}
			deadline := time.Now().Add(attempt + time.Second)
			for over.Load() < started.Load() && time.Now().Before(deadline) {
				time.Sleep(10 * time.Millisecond)
			}
			if started, over := started.Load(), over.Load(); started == 0 || over != started {
				t.Errorf("%d started and %d over after the fetch, want every one over", started, over)
			}
		})
	}
}

// silentServer takes connections and never says a word. It counts each one
// it takes, and each one the other side lets go of, and returns its address.
func silentServer(t *testing.T, taken, letGo *atomic.Int32) string {
	t.Helper()

	listener, err := (&net.ListenConfig{}).Listen(t.Context(), "tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	t.Cleanup(func() { _ = listener.Close() })
	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			taken.Add(1)
			go func() {
				defer func() { _ = conn.Close() }()
				_, _ = io.Copy(io.Discard, conn)
				letGo.Add(1)
			}()
		}
	}()
	return listener.Addr().String()
}

// A name that cannot be reached for now — no answer from the resolver, no
// address, or no address that takes the connection — is unreachable, and it
// is tried once more, since the next lookup or dial may find it.
func TestANameThatCannotBeReachedIsTriedAgain(t *testing.T) {
	t.Parallel()

	refused := func(context.Context, string, string) (net.Conn, error) {
		return nil, errors.New("connection refused")
	}
	for _, tc := range []struct {
		name    string
		resolve func(context.Context, string) ([]netip.Addr, error)
		dial    func(context.Context, string, string) (net.Conn, error)
	}{
		{"the resolver fails", func(context.Context, string) ([]netip.Addr, error) {
			return nil, errors.New("no such host")
		}, refused},
		{"the name stands for nothing", func(context.Context, string) ([]netip.Addr, error) {
			return nil, nil
		}, refused},
		{"no address takes the connection", func(context.Context, string) ([]netip.Addr, error) {
			return []netip.Addr{netip.MustParseAddr("192.0.2.1"), netip.MustParseAddr("192.0.2.2")}, nil
		}, refused},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var lookups atomic.Int32
			fetcher := newFetcher(reach{
				resolve: func(ctx context.Context, host string) ([]netip.Addr, error) {
					lookups.Add(1)
					return tc.resolve(ctx, host)
				},
				allowed: public,
				dial:    tc.dial,
			}, time.Second, (&clock{at: someDay}).now)

			if _, err := fetcher.Fetch(t.Context(), clientAt); !errors.Is(err, ErrUnreachable) {
				t.Errorf("Fetch() error = %v, want ErrUnreachable", err)
			}
			if got := lookups.Load(); got != 2 {
				t.Errorf("%d lookups, want 2: the first attempt and one more", got)
			}
		})
	}
}

// A server that breaks the document off halfway has sent no document, and it
// is asked once more.
func TestADocumentBrokenOffIsTriedAgain(t *testing.T) {
	t.Parallel()

	server := newFakeClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Length", "1000")
		_, _ = w.Write([]byte(`{"client_id":`))
		if hijacker, ok := w.(http.Hijacker); ok {
			if conn, _, err := hijacker.Hijack(); err == nil {
				_ = conn.Close()
			}
		}
	})
	fetcher, _ := fetcherFor(t, server, publicOrLoopback, time.Second, loopback)

	if _, err := fetcher.Fetch(t.Context(), clientAt); !errors.Is(err, ErrUnreachable) {
		t.Errorf("Fetch() error = %v, want ErrUnreachable", err)
	}
	if got := server.requests.Load(); got != 2 {
		t.Errorf("%d requests reached the client, want 2", got)
	}
}

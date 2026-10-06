package httpserver_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	metricnoop "go.opentelemetry.io/otel/metric/noop"
	tracenoop "go.opentelemetry.io/otel/trace/noop"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"

	"github.com/MathTrail/mathtrail-standalone/internal/apierror"
	"github.com/MathTrail/mathtrail-standalone/internal/ratelimit/ratelimittest"
	httpserver "github.com/MathTrail/mathtrail-standalone/internal/transport/http"
)

// roomy are paces no case of routing comes near: those cases are about where
// a request goes, and the paces have cases of their own.
func roomy(t testing.TB) httpserver.Limits {
	t.Helper()
	return httpserver.Limits{PerAddress: ratelimittest.Roomy(t), Instance: ratelimittest.Roomy(t)}
}

// paces are the paces of an address and of the instance, each so many requests
// a minute, on a clock that stands still.
func paces(t testing.TB, perAddress, instance int) httpserver.Limits {
	t.Helper()
	return httpserver.Limits{PerAddress: ratelimittest.Keyed(t, perAddress), Instance: ratelimittest.Shared(t, instance)}
}

// pacedRouter is the router held to the paces given, with every line it
// writes kept for the case to read.
type pacedRouter struct {
	handler http.Handler
	lines   *observer.ObservedLogs
}

func newPacedRouter(t *testing.T, limits httpserver.Limits) *pacedRouter {
	t.Helper()

	core, lines := observer.New(zapcore.DebugLevel)
	router, err := httpserver.NewRouter(publicURL, endpoints(), limits, zap.New(core), httpserver.Observability{
		Traces: tracenoop.NewTracerProvider(),
		Meters: metricnoop.NewMeterProvider(),
		Flush:  noDelivery,
	})
	if err != nil {
		t.Fatalf("NewRouter() error = %v, want nil", err)
	}
	return &pacedRouter{handler: router, lines: lines}
}

// ask sends one request to the path given, as the platform hands it on: with
// the hops it came through in X-Forwarded-For, the last one the platform's.
func (p *pacedRouter) ask(t *testing.T, method, path, forwardedFor string) *httptest.ResponseRecorder {
	t.Helper()

	req := httptest.NewRequestWithContext(t.Context(), method, path, http.NoBody)
	if forwardedFor != "" {
		req.Header.Set("X-Forwarded-For", forwardedFor)
	}
	rec := httptest.NewRecorder()
	p.handler.ServeHTTP(rec, req)
	return rec
}

// hits are the lines of the ceilings reached, by the ceiling each names.
func (p *pacedRouter) hits() []string {
	var named []string
	hits := p.lines.FilterMessage("limit_hit").All()
	for i := range hits {
		limit, _ := hits[i].ContextMap()["limit"].(string)
		named = append(named, limit)
	}
	return named
}

// An address past its pace at the sign-in is refused before the endpoint runs:
// 429, when to ask again, no cache to keep it, and the refusal in the shape
// every refusal of the router has. The ceiling reached is written down once for
// the flood, and never with the address.
func TestAnAddressPastItsPaceIsRefusedAtTheSignIn(t *testing.T) {
	t.Parallel()

	const household = "203.0.113.7"
	p := newPacedRouter(t, paces(t, 3, 1<<20))

	if rec := p.ask(t, http.MethodPost, "/oauth/token", household); rec.Body.String() != reachedToken {
		t.Fatalf("the first request: status %d, body %q, want the token endpoint reached", rec.Code, rec.Body.String())
	}
	rec := p.ask(t, http.MethodPost, "/oauth/token", household)
	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("the second request: status = %d, want %d", rec.Code, http.StatusTooManyRequests)
	}
	if got := rec.Header().Get("Retry-After"); got != "20" {
		t.Errorf("Retry-After = %q, want 20: three a minute is one every twenty seconds", got)
	}
	if got := rec.Header().Get("Cache-Control"); got != "no-store" {
		t.Errorf("Cache-Control = %q, want no-store", got)
	}
	var refusal apierror.Response
	if err := json.Unmarshal(rec.Body.Bytes(), &refusal); err != nil || refusal.Code != apierror.CodeTooManyRequests {
		t.Errorf("body = %s, want the router's refusal with code %s", rec.Body.String(), apierror.CodeTooManyRequests)
	}
	p.ask(t, http.MethodPost, "/oauth/token", household)

	if hits := p.hits(); len(hits) != 1 || hits[0] != "ip_rate" {
		t.Errorf("limit_hit lines = %v, want one, naming ip_rate, for the whole run", hits)
	}
	if warned := p.lines.FilterMessage("limit_hit").FilterLevelExact(zapcore.WarnLevel).Len(); warned != 1 {
		t.Errorf("limit_hit warnings = %d, want the line a warning", warned)
	}
	p.wantNoLineCarrying(t, household)
}

// wantNoLineCarrying holds every line the router wrote to leaving out the text
// given.
func (p *pacedRouter) wantNoLineCarrying(t *testing.T, text string) {
	t.Helper()

	lines := p.lines.All()
	for i := range lines {
		for key, value := range lines[i].ContextMap() {
			if written, isText := value.(string); isText && strings.Contains(written, text) {
				t.Errorf("the %s line carries %q in %q", lines[i].Message, text, key)
			}
		}
	}
}

// One address past its pace costs no other address anything.
func TestOneAddressPastItsPaceCostsNoOtherAddressAnything(t *testing.T) {
	t.Parallel()

	p := newPacedRouter(t, paces(t, 3, 1<<20))
	p.ask(t, http.MethodPost, "/oauth/token", "203.0.113.7")
	p.ask(t, http.MethodPost, "/oauth/token", "203.0.113.7")

	if rec := p.ask(t, http.MethodPost, "/oauth/token", "198.51.100.4"); rec.Body.String() != reachedToken {
		t.Errorf("another address: status %d, body %q, want the token endpoint reached", rec.Code, rec.Body.String())
	}
}

// Whatever a client writes into X-Forwarded-For stands before the hop the
// platform appends, and is never read: a client that forges the header buys no
// fresh allowance, while another hop at the end is another address.
func TestAForgedForwardedForBuysNoFreshAllowance(t *testing.T) {
	t.Parallel()

	p := newPacedRouter(t, paces(t, 3, 1<<20))
	p.ask(t, http.MethodGet, "/oauth/authorize", "10.0.0.1, 203.0.113.7")

	for _, forged := range []string{"10.0.0.2, 203.0.113.7", "198.51.100.4,203.0.113.7", "203.0.113.7"} {
		if rec := p.ask(t, http.MethodGet, "/oauth/authorize", forged); rec.Code != http.StatusTooManyRequests {
			t.Errorf("X-Forwarded-For %q: status = %d, want %d: the platform's hop is the same", forged, rec.Code, http.StatusTooManyRequests)
		}
	}
	if rec := p.ask(t, http.MethodGet, "/oauth/authorize", "203.0.113.7, 198.51.100.4"); rec.Body.String() != reachedAuthorize {
		t.Errorf("another last hop: status %d, body %q, want the authorization request reached", rec.Code, rec.Body.String())
	}
}

// An IPv6 household is counted by the network it is given, so that its many
// addresses share one allowance; another network is another household.
func TestAnIPv6HouseholdSharesOneAllowance(t *testing.T) {
	t.Parallel()

	p := newPacedRouter(t, paces(t, 3, 1<<20))
	p.ask(t, http.MethodGet, "/oauth/callback", "2001:db8:1:2::1")

	if rec := p.ask(t, http.MethodGet, "/oauth/callback", "2001:db8:1:2:aaaa:bbbb:cccc:dddd"); rec.Code != http.StatusTooManyRequests {
		t.Errorf("another address of the same /64: status = %d, want %d", rec.Code, http.StatusTooManyRequests)
	}
	if rec := p.ask(t, http.MethodGet, "/oauth/callback", "2001:db8:1:3::1"); rec.Body.String() != reachedCallback {
		t.Errorf("another /64: status %d, body %q, want Google's answer reached", rec.Code, rec.Body.String())
	}
}

// The instance's pace holds every address together: past it, an address that
// has spent nothing of its own is refused too, the ceiling named is the
// instance's, and the request it turned away is given back to the address.
func TestTheInstancesPaceHoldsEveryAddressTogether(t *testing.T) {
	t.Parallel()

	limits := paces(t, 3, 3)
	p := newPacedRouter(t, limits)
	p.ask(t, http.MethodPost, "/oauth/register", "203.0.113.7")

	if rec := p.ask(t, http.MethodPost, "/oauth/register", "198.51.100.4"); rec.Code != http.StatusTooManyRequests {
		t.Errorf("another address past the instance's pace: status = %d, want %d", rec.Code, http.StatusTooManyRequests)
	}
	if hits := p.hits(); len(hits) != 1 || hits[0] != "instance_rate" {
		t.Errorf("limit_hit lines = %v, want one, naming instance_rate", hits)
	}
	if got := limits.PerAddress.Take("198.51.100.4"); !got.Allowed {
		t.Errorf("the address's own pace after = %+v, want the request the instance turned away given back", got)
	}
}

// Every endpoint of the sign-in, and each of its documents, is held to the
// pace; the probe, the MCP endpoint and a path nobody declared are not.
func TestOnlyTheSignInIsHeldToThePaceOfAnAddress(t *testing.T) {
	t.Parallel()

	for _, route := range []struct{ method, path string }{
		{http.MethodGet, "/.well-known/oauth-protected-resource/mcp"},
		{http.MethodGet, "/.well-known/oauth-authorization-server"},
		{http.MethodGet, httpserver.ChallengePath},
		{http.MethodPost, "/oauth/register"},
		{http.MethodGet, "/oauth/authorize"},
		{http.MethodPost, "/oauth/consent"},
		{http.MethodGet, "/oauth/callback"},
		{http.MethodGet, "/oauth/drive"},
		{http.MethodPost, "/oauth/token"},
		{http.MethodPost, "/oauth/revoke"},
	} {
		p := newPacedRouter(t, paces(t, 3, 1<<20))
		p.ask(t, route.method, route.path, "203.0.113.7")
		if rec := p.ask(t, route.method, route.path, "203.0.113.7"); rec.Code != http.StatusTooManyRequests {
			t.Errorf("%s %s past the pace: status = %d, want %d", route.method, route.path, rec.Code, http.StatusTooManyRequests)
		}
	}

	p := newPacedRouter(t, paces(t, 3, 3))
	p.ask(t, http.MethodPost, "/oauth/register", "203.0.113.7")
	p.ask(t, http.MethodPost, "/oauth/register", "203.0.113.7")
	for _, route := range []struct{ method, path string }{
		{http.MethodGet, "/health"},
		{http.MethodPost, "/mcp"},
		{http.MethodGet, "/no-such-path"},
	} {
		if rec := p.ask(t, route.method, route.path, "203.0.113.7"); rec.Code == http.StatusTooManyRequests {
			t.Errorf("%s %s after the sign-in's pace was spent: status = %d, want it not held to that pace", route.method, route.path, rec.Code)
		}
	}
}

// The token endpoint is where a lesson renews its access, every few minutes, so
// a flood that has spent the pace the sign-in takes from everybody does not hold
// it back: it answers any address that has spent nothing of its own, while a
// registration from the same address is refused by the instance.
func TestARenewalIsNotHeldToTheSignInsPace(t *testing.T) {
	t.Parallel()

	p := newPacedRouter(t, paces(t, 1<<20, 3))
	p.ask(t, http.MethodPost, "/oauth/register", "203.0.113.7")

	if rec := p.ask(t, http.MethodPost, "/oauth/token", "198.51.100.4"); rec.Code == http.StatusTooManyRequests {
		t.Errorf("a renewal past the sign-in's pace: status = %d, want it answered", rec.Code)
	}
	if rec := p.ask(t, http.MethodPost, "/oauth/register", "198.51.100.4"); rec.Code != http.StatusTooManyRequests {
		t.Errorf("a registration past the sign-in's pace: status = %d, want %d", rec.Code, http.StatusTooManyRequests)
	}
}

// A router given no pace to hold the sign-in to is refused where it is built.
func TestARouterWithoutItsPacesIsRefused(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name    string
		missing func(*httpserver.Limits)
		want    string
	}{
		{"no pace of an address", func(l *httpserver.Limits) { l.PerAddress = nil }, "PerAddress"},
		{"no pace of the instance", func(l *httpserver.Limits) { l.Instance = nil }, "Instance"},
	} {
		limits := roomy(t)
		tc.missing(&limits)
		_, err := httpserver.NewRouter(publicURL, endpoints(), limits, zap.NewNop(), httpserver.Observability{
			Traces: tracenoop.NewTracerProvider(),
			Meters: metricnoop.NewMeterProvider(),
			Flush:  noDelivery,
		})
		if !errors.Is(err, httpserver.ErrLimits) || !strings.Contains(err.Error(), tc.want) {
			t.Errorf("%s: NewRouter() error = %v, want ErrLimits naming %s", tc.name, err, tc.want)
		}
	}
}

// A parent's browser past its pace is shown the sign-in's own page, which says
// so in the parent's language; a host's server is answered as the router
// answers a program. Both are told when to ask again.
func TestAPersonIsShownAPageAndAProgramIsAnswered(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		method, path string
		page         bool
	}{
		{http.MethodGet, "/oauth/authorize", true},
		{http.MethodPost, "/oauth/consent", true},
		{http.MethodGet, "/oauth/callback", true},
		{http.MethodGet, "/oauth/drive", true},
		{http.MethodPost, "/oauth/register", false},
		{http.MethodPost, "/oauth/token", false},
		{http.MethodPost, "/oauth/revoke", false},
		{http.MethodGet, "/.well-known/oauth-authorization-server", false},
	} {
		p := newPacedRouter(t, paces(t, 3, 1<<20))
		p.ask(t, tc.method, tc.path, "203.0.113.7")
		rec := p.ask(t, tc.method, tc.path, "203.0.113.7")

		if rec.Code != http.StatusTooManyRequests || rec.Header().Get("Retry-After") == "" {
			t.Errorf("%s %s past the pace: status %d, Retry-After %q, want 429 and when to ask again",
				tc.method, tc.path, rec.Code, rec.Header().Get("Retry-After"))
		}
		if shown := rec.Body.String() == reachedBusy; shown != tc.page {
			t.Errorf("%s %s past the pace: the busy page shown = %t, want %t", tc.method, tc.path, shown, tc.page)
		}
	}
}

// The documents cost nothing to answer, so they are not held to the pace the
// sign-in takes from everybody: past it, an address that has spent nothing of
// its own still reads them.
func TestTheDocumentsAreNotHeldToTheSignInsPace(t *testing.T) {
	t.Parallel()

	p := newPacedRouter(t, paces(t, 1<<20, 3))
	p.ask(t, http.MethodPost, "/oauth/register", "203.0.113.7")

	if rec := p.ask(t, http.MethodGet, "/.well-known/oauth-authorization-server", "198.51.100.4"); rec.Body.String() != reachedServerMetadata {
		t.Errorf("a document past the sign-in's pace: status %d, body %q, want it read", rec.Code, rec.Body.String())
	}
	if rec := p.ask(t, http.MethodPost, "/oauth/register", "198.51.100.4"); rec.Code != http.StatusTooManyRequests {
		t.Errorf("a registration past the sign-in's pace: status = %d, want %d", rec.Code, http.StatusTooManyRequests)
	}
}

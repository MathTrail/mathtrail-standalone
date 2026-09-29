package oauthserver

import (
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"
)

// plainPagePolicy is the policy of a page whose form, if it has one, goes
// nowhere but here.
const plainPagePolicy = "default-src 'none'; style-src 'unsafe-inline'; form-action 'self'; frame-ancestors 'none'"

// Until the host is known and the address it gave is one it registered, a
// request that cannot be answered stops at a page and sends the parent
// nowhere; the page names the error for whoever has to fix the host.
func TestARequestWithNoTrustedAddressStopsAtAPage(t *testing.T) {
	t.Parallel()

	twice := func(name string) func(string) string {
		return func(address string) string { return address + "&" + name + "=x" }
	}
	for _, tc := range []struct {
		name    string
		changed url.Values
		spoil   func(string) string
		status  int
		code    string
		reason  string
	}{
		{name: "a query that does not parse", spoil: func(a string) string { return a + "&%zz" },
			status: http.StatusBadRequest, code: "invalid_request", reason: "invalid_request"},
		{name: "no response type", changed: url.Values{"response_type": {""}},
			status: http.StatusBadRequest, code: "invalid_request", reason: "invalid_request"},
		{name: "a token asked for", changed: url.Values{"response_type": {"token"}},
			status: http.StatusBadRequest, code: "unsupported_response_type", reason: "unsupported_response_type"},
		{name: "no challenge", changed: url.Values{"code_challenge": {""}},
			status: http.StatusBadRequest, code: "invalid_request", reason: "invalid_pkce"},
		{name: "no method", changed: url.Values{"code_challenge_method": {""}},
			status: http.StatusBadRequest, code: "invalid_request", reason: "invalid_pkce"},
		{name: "the plain method", changed: url.Values{"code_challenge_method": {"plain"}, "code_challenge": {hostVerifier}},
			status: http.StatusBadRequest, code: "invalid_request", reason: "invalid_pkce"},
		{name: "a challenge of the wrong length", changed: url.Values{"code_challenge": {"c2hvcnQ"}},
			status: http.StatusBadRequest, code: "invalid_request", reason: "invalid_pkce"},
		{name: "a challenge that is not base64url", changed: url.Values{"code_challenge": {strings.Repeat("+", 43)}},
			status: http.StatusBadRequest, code: "invalid_request", reason: "invalid_pkce"},
		{name: "a client named twice", spoil: twice("client_id"),
			status: http.StatusBadRequest, code: "invalid_request", reason: "invalid_request"},
		{name: "an address back named twice", spoil: twice("redirect_uri"),
			status: http.StatusBadRequest, code: "invalid_request", reason: "invalid_request"},
		{name: "no client", changed: url.Values{"client_id": {""}},
			status: http.StatusBadRequest, code: "invalid_client", reason: "invalid_client"},
		{name: "a client nobody issued", changed: url.Values{"client_id": {"mt1.d.x.y"}},
			status: http.StatusBadRequest, code: "invalid_client", reason: "invalid_client"},
		{name: "a client whose document cannot be fetched", changed: url.Values{"client_id": {"https://client.example/meta"}},
			status: http.StatusBadRequest, code: "invalid_client", reason: "invalid_client"},
		{name: "no address back", changed: url.Values{"redirect_uri": {""}},
			status: http.StatusBadRequest, code: "invalid_request", reason: "invalid_redirect_uri"},
		{name: "an address the host never registered", changed: url.Values{"redirect_uri": {"https://attacker.example/cb"}},
			status: http.StatusBadRequest, code: "invalid_request", reason: "invalid_redirect_uri"},
		{name: "the registered address with a slash added", changed: url.Values{"redirect_uri": {hostRedirect + "/"}},
			status: http.StatusBadRequest, code: "invalid_request", reason: "invalid_redirect_uri"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			h := newSignIn(t)
			address := h.authorizeURL(h.register(hostRedirect, hostName), tc.changed)
			if tc.spoil != nil {
				address = tc.spoil(address)
			}
			stopped := h.browser(t).get(address)
			if stopped.status != tc.status || stopped.location != "" || !strings.Contains(stopped.body, tc.code) {
				t.Errorf("GET /oauth/authorize = %d to %q, want %d and a page naming %s", stopped.status, stopped.location, tc.status, tc.code)
			}
			if got := stopped.header.Get("Content-Security-Policy"); got != plainPagePolicy {
				t.Errorf("the page's policy = %q, want %q", got, plainPagePolicy)
			}
			if lines := h.lines(eventAuthReject); len(lines) != 1 || lines[0]["step"] != stepAuthorize || lines[0]["reason"] != tc.reason {
				t.Errorf("auth_reject lines = %v, want one at authorize for %s", lines, tc.reason)
			}
		})
	}
}

// ungrantable is a request from a known host, at an address it registered,
// that this server cannot give what it asks for: what it changes of a good
// request, the state the host hears back, and the error and the reason it is
// refused with.
type ungrantable struct {
	name    string
	changed func(served string) url.Values
	state   string
	error   string
	reason  string
}

// Once the host and its address are known, and the parent has approved the
// host there in this browser, a request this server cannot give what it asks
// for sends the parent back to the host, which hears why in the protocol's
// words, with its own state and this issuer.
func TestARequestThatCannotBeGrantedIsSentBackToAnApprovedHost(t *testing.T) {
	t.Parallel()

	for _, tc := range ungrantables() {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			h := newSignIn(t)
			clientID := h.register(hostRedirect, hostName)
			parent := h.browser(t)
			h.toGoogle(parent, clientID)
			back := parent.get(h.authorizeURL(clientID, tc.changed(h.served.URL)))
			answer := answerAt(t, back)
			if back.status != http.StatusFound || answer.Get("error") != tc.error ||
				answer.Get("state") != tc.state || answer.Get("iss") != h.served.URL {
				t.Errorf("GET /oauth/authorize = %d with %v, want %s with the host's state and this issuer", back.status, answer, tc.error)
			}
			if lines := h.lines(eventAuthAuthorize); len(lines) != 2 || lines[1]["outcome"] != "refused" || lines[1]["reason"] != tc.reason {
				t.Errorf("auth_authorize lines = %v, want the second one refused for %s", lines, tc.reason)
			}
		})
	}
}

// Until the parent has approved the host in this browser, the same request
// stops at a page, and sends the parent nowhere: anybody may register a host
// at an address of their own choosing, and a link to this server must not be a
// way to send a parent there before they have seen a screen of it.
func TestARequestThatCannotBeGrantedStopsAtAPageUntilTheHostIsApproved(t *testing.T) {
	t.Parallel()

	for _, tc := range ungrantables() {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			h := newSignIn(t)
			stopped := h.browser(t).get(h.authorizeURL(h.register(hostRedirect, hostName), tc.changed(h.served.URL)))
			if stopped.status != http.StatusBadRequest || stopped.location != "" || !strings.Contains(stopped.body, tc.error) {
				t.Errorf("GET /oauth/authorize = %d to %q, want 400 and a page naming %s", stopped.status, stopped.location, tc.error)
			}
			if lines := h.lines(eventAuthReject); len(lines) != 1 || lines[0]["step"] != stepAuthorize || lines[0]["reason"] != tc.reason {
				t.Errorf("auth_reject lines = %v, want one at authorize for %s", lines, tc.reason)
			}
			if lines := h.lines(eventAuthAuthorize); len(lines) != 0 {
				t.Errorf("auth_authorize lines = %v, want none: the request stopped at a page", lines)
			}
		})
	}
}

// A host on the parent's own computer hears why its request cannot be granted
// at its address, approved or not: an address there sends the parent nowhere
// else, and a program waiting at it would otherwise wait for an answer that
// never comes.
func TestARequestThatCannotBeGrantedIsSentBackToTheParentsComputer(t *testing.T) {
	t.Parallel()

	const here = "http://127.0.0.1:33418/cb"
	h := newSignIn(t)
	asked := h.authorizeURL(h.register(here, hostName), url.Values{"redirect_uri": {here}, "scope": {"mcp admin"}})
	back := h.browser(t).get(asked)

	address, err := url.Parse(back.location)
	if err != nil || back.status != http.StatusFound || !strings.HasPrefix(back.location, here+"?") ||
		address.Query().Get("error") != "invalid_scope" {
		t.Errorf("GET /oauth/authorize = %d to %q, want the refusal sent back to %s", back.status, back.location, here)
	}
}

// ungrantables are the requests a known host may send from an address it
// registered that this server cannot give what they ask for.
func ungrantables() []ungrantable {
	longState := strings.Repeat("s", maxHostState+1)
	return []ungrantable{
		{"another resource", func(string) url.Values {
			return url.Values{"resource": {"https://other.example/mcp"}}
		}, hostState, "invalid_target", "invalid_target"},
		{"a resource of another path", func(served string) url.Values {
			return url.Values{"resource": {served + "/other"}}
		}, hostState, "invalid_target", "invalid_target"},
		{"two resources", func(served string) url.Values {
			return url.Values{"resource": {served + "/mcp", "https://other.example/mcp"}}
		}, hostState, "invalid_target", "invalid_target"},
		{"another scope", func(string) url.Values {
			return url.Values{"scope": {"mcp admin"}}
		}, hostState, "invalid_scope", "invalid_scope"},
		{"a state too long to carry", func(string) url.Values {
			return url.Values{"state": {longState}}
		}, longState, "invalid_request", "state_too_long"},
		{"a state given twice", func(string) url.Values {
			return url.Values{"state": {hostState, "another"}}
		}, hostState, "invalid_request", "invalid_request"},
		{"a scope given twice", func(string) url.Values {
			return url.Values{"scope": {"mcp", "mcp"}}
		}, hostState, "invalid_request", "invalid_request"},
	}
}

// What a request may leave out, or say in the forms the protocols allow, it
// is given: the resource here when it names none, with a trailing slash or
// without; the one scope when it asks for none.
func TestARequestMayLeaveOutWhatHasOneAnswer(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name     string
		changed  func(served string) url.Values
		resource string
		scope    string
	}{
		{"no resource", func(string) url.Values { return url.Values{"resource": {""}} }, "absent", "mcp"},
		{"the resource with a slash", func(served string) url.Values {
			return url.Values{"resource": {served + "/mcp/"}}
		}, "given", "mcp"},
		{"no scope", func(string) url.Values { return url.Values{"scope": {""}} }, "given", "none"},
		{"no state", func(string) url.Values { return url.Values{"state": {""}} }, "given", "mcp"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			h := newSignIn(t)
			screen := h.browser(t).get(h.authorizeURL(h.register(hostRedirect, hostName), tc.changed(h.served.URL)))
			if screen.status != http.StatusOK {
				t.Fatalf("GET /oauth/authorize status = %d, want the consent screen; body %q", screen.status, screen.body)
			}
			request, err := h.server.flow.openFlight(requestOn(t, screen))
			if err != nil {
				t.Fatalf("the screen's request does not open: %v", err)
			}
			if request.Resource != h.served.URL+"/mcp" || request.Scope != "mcp" {
				t.Errorf("the request is for %q and %q, want the resource here and its one scope", request.Resource, request.Scope)
			}
			lines := h.lines(eventAuthAuthorize)
			if len(lines) != 1 || lines[0]["outcome"] != "consent" || lines[0]["resource"] != tc.resource || lines[0]["scope"] != tc.scope {
				t.Errorf("auth_authorize lines = %v, want one with resource %s and scope %s", lines, tc.resource, tc.scope)
			}
		})
	}
}

// The consent screen's answer counts only from the browser the sign-in began
// in, only while the sign-in is under way, and only as one of its two
// answers.
func TestAConsentIsHeldToItsBrowser(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name   string
		form   func(t *testing.T, h *signIn, request string) (*browser, url.Values)
		status int
		reason string
	}{
		{"another browser", func(t *testing.T, h *signIn, request string) (*browser, url.Values) {
			return h.browser(t), url.Values{"request": {request}, "decision": {"allow"}}
		}, http.StatusBadRequest, "cookie"},
		{"a request changed on its way", func(_ *testing.T, _ *signIn, request string) (*browser, url.Values) {
			return nil, url.Values{"request": {request[:len(request)-2] + flipped(request[len(request)-2:])}, "decision": {"allow"}}
		}, http.StatusBadRequest, "unknown_request"},
		{"no request", func(*testing.T, *signIn, string) (*browser, url.Values) {
			return nil, url.Values{"decision": {"allow"}}
		}, http.StatusBadRequest, "unknown_request"},
		{"a sign-in past its ten minutes", func(_ *testing.T, h *signIn, request string) (*browser, url.Values) {
			h.clock.advance(flightLifetime + time.Second)
			return nil, url.Values{"request": {request}, "decision": {"allow"}}
		}, http.StatusBadRequest, "expired"},
		{"neither answer", func(_ *testing.T, _ *signIn, request string) (*browser, url.Values) {
			return nil, url.Values{"request": {request}, "decision": {"maybe"}}
		}, http.StatusBadRequest, "invalid_request"},
		{"a form too large to read", func(_ *testing.T, _ *signIn, request string) (*browser, url.Values) {
			return nil, url.Values{"request": {request}, "decision": {"allow"}, "padding": {strings.Repeat("p", maxConsentForm)}}
		}, http.StatusRequestEntityTooLarge, "too_large"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			h := newSignIn(t)
			parent := h.browser(t)
			_, request := h.toConsent(parent, h.register(hostRedirect, hostName))
			sender, form := tc.form(t, h, request)
			if sender == nil {
				sender = parent
			}
			answer := sender.post(h.served.URL+"/oauth/consent", form)
			if answer.status != tc.status || answer.location != "" {
				t.Errorf("POST /oauth/consent = %d to %q, want %d and a page", answer.status, answer.location, tc.status)
			}
			if lines := h.lines(eventAuthReject); len(lines) != 1 || lines[0]["step"] != stepConsent || lines[0]["reason"] != tc.reason {
				t.Errorf("auth_reject lines = %v, want one at consent for %s", lines, tc.reason)
			}
			if lines := h.lines(eventAuthConsent); len(lines) != 0 {
				t.Errorf("auth_consent lines = %v, want none: the parent's answer did not count", lines)
			}
		})
	}
}

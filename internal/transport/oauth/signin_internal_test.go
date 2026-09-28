package oauthserver

import (
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/MathTrail/mathtrail-standalone/internal/infra/googleauth/googletest"
)

// A parent goes from the host's request through the consent screen and Google
// to the host, which is handed a code; the code opens into exactly what the
// exchange of it has to match, and into Google's grant for the account.
func TestAParentSignsInFromTheHostToTheCode(t *testing.T) {
	t.Parallel()

	h := newSignIn(t)
	clientID := h.register(hostRedirect, hostName)
	parent := h.browser(t)

	screen, request := h.toConsent(parent, clientID)
	policy := screen.header.Get("Content-Security-Policy")
	if !strings.Contains(policy, "form-action 'self' "+h.google.URL+" https://host.example;") {
		t.Errorf("the consent screen's policy = %q, want its form let on to Google and to the host alone", policy)
	}
	if parent.cookie("/oauth/consent", csrfCookie) == "" || parent.cookie("/", csrfCookie) != "" {
		t.Error("the browser holds no sign-in cookie for /oauth, or holds one beyond it")
	}

	allowed := parent.post(h.served.URL+"/oauth/consent", url.Values{"request": {request}, "decision": {"allow"}})
	google, err := url.Parse(allowed.location)
	if err != nil || allowed.status != http.StatusSeeOther || google.Query().Get("state") != request {
		t.Fatalf("POST /oauth/consent = %d to %q, want 303 to Google carrying the request as its state", allowed.status, allowed.location)
	}

	answer := answerAt(t, parent.get(h.google.Allow(allowed.location)))
	if answer.Get("state") != hostState || answer.Get("iss") != h.served.URL || answer.Has("error") {
		t.Fatalf("the host was sent %v, want a code with its own state and this issuer", answer)
	}
	if parent.cookie("/oauth/callback", csrfCookie) != "" {
		t.Error("the sign-in cookie outlived the sign-in")
	}

	code := h.opened(answer.Get("code"))
	want := grantCode{
		User:               h.ring.UserID("google-sub:" + googletest.Subject),
		Client:             digestOf(clientID),
		RedirectURI:        hostRedirect,
		Challenge:          googletest.ChallengeOf(hostVerifier),
		Resource:           h.served.URL + "/mcp",
		Scope:              "mcp",
		GoogleAccessToken:  googletest.AccessToken,
		GoogleAccessExpiry: testDay.Add(googletest.ExpiresIn * time.Second).Unix(),
		GoogleRefreshToken: googletest.RefreshToken,
		IssuedAt:           testDay.Unix(),
	}
	if code != want {
		t.Errorf("the code opens into %+v, want %+v", code, want)
	}
}

// Every step of a sign-in leaves a line, and no line carries the host's state,
// the request, the code, a token or who signed in at Google: only which host,
// which way, and how it ended.
func TestASignInLeavesItsLinesAndNothingOfItsSecrets(t *testing.T) {
	t.Parallel()

	h := newSignIn(t)
	clientID := h.register(hostRedirect, hostName)
	parent := h.browser(t)
	answer := answerAt(t, parent.get(h.google.Allow(h.toGoogle(parent, clientID))))

	for event, want := range map[string]map[string]any{
		eventAuthAuthorize: {"outcome": "consent", "registration": registrationDCR, "redirect_host": "host.example", "resource": "given", "scope": "mcp"},
		eventAuthConsent:   {"outcome": "allowed", "registration": registrationDCR, "redirect_host": "host.example"},
		eventAuthCallback:  {"outcome": "ok", "registration": registrationDCR, "user": h.ring.UserID("google-sub:" + googletest.Subject)},
	} {
		lines := h.lines(event)
		if len(lines) != 1 {
			t.Errorf("%s lines = %v, want one", event, lines)
			continue
		}
		for field, value := range want {
			if lines[0][field] != value {
				t.Errorf("%s %s = %v, want %v", event, field, lines[0][field], value)
			}
		}
	}
	noLineCarries(t, h, hostState, answer.Get("code"), googletest.AccessToken, googletest.RefreshToken, googletest.Subject, clientID)
}

// noLineCarries fails for every field of every line that carries one of the
// values given.
func noLineCarries(t *testing.T, h *signIn, values ...string) {
	t.Helper()

	entries := h.logs.All()
	for i := range entries {
		for field, carried := range entries[i].ContextMap() {
			text, ok := carried.(string)
			for _, value := range values {
				if ok && strings.Contains(text, value) {
					t.Errorf("%s %s carries %q", entries[i].Message, field, value)
				}
			}
		}
	}
}

// A browser whose parent approved a host before goes from the host's request
// straight on to Google.
func TestAnApprovedHostGoesStraightToGoogle(t *testing.T) {
	t.Parallel()

	h := newSignIn(t)
	clientID := h.register(hostRedirect, hostName)
	parent := h.browser(t)
	h.toGoogle(parent, clientID)

	again := parent.get(h.authorizeURL(clientID, nil))
	if again.status != http.StatusFound || !strings.HasPrefix(again.location, h.google.URL+"/auth?") {
		t.Fatalf("GET /oauth/authorize again = %d to %q, want 302 to Google", again.status, again.location)
	}
	if lines := h.lines(eventAuthAuthorize); len(lines) != 2 || lines[1]["outcome"] != "to_google" {
		t.Errorf("auth_authorize lines = %v, want the second one gone on to Google", lines)
	}

	// Another browser, or the same host sending the parent somewhere else, is
	// shown the screen again.
	if other := h.browser(t).get(h.authorizeURL(clientID, nil)); other.status != http.StatusOK {
		t.Errorf("another browser: status = %d, want the consent screen", other.status)
	}
	elsewhere := h.register("https://elsewhere.example/cb", hostName)
	if other := parent.get(h.authorizeURL(elsewhere, url.Values{"redirect_uri": {"https://elsewhere.example/cb"}})); other.status != http.StatusOK {
		t.Errorf("another address back: status = %d, want the consent screen", other.status)
	}
}

// A parent who declines on the consent screen is sent back to the host, which
// hears it was declined; the sign-in is over.
func TestAParentWhoDeclinesSendsTheHostAway(t *testing.T) {
	t.Parallel()

	h := newSignIn(t)
	parent := h.browser(t)
	_, request := h.toConsent(parent, h.register(hostRedirect, hostName))

	declined := parent.post(h.served.URL+"/oauth/consent", url.Values{"request": {request}, "decision": {"deny"}})
	answer := answerAt(t, declined)
	if declined.status != http.StatusSeeOther || answer.Get("error") != "access_denied" ||
		answer.Get("state") != hostState || answer.Get("iss") != h.served.URL {
		t.Errorf("POST /oauth/consent declined = %d with %v, want 303 with access_denied, the state and the issuer", declined.status, answer)
	}
	if parent.cookie("/oauth/consent", csrfCookie) != "" || parent.cookie("/oauth/consent", consentCookie) != "" {
		t.Error("declining left a cookie behind, want the sign-in's taken back and no approval")
	}
	if lines := h.lines(eventAuthConsent); len(lines) != 1 || lines[0]["outcome"] != "denied" {
		t.Errorf("auth_consent lines = %v, want one denied", lines)
	}
}

// A parent who declines at Google is sent back to the host, which hears it was
// declined.
func TestAParentWhoDeclinesAtGoogleSendsTheHostAway(t *testing.T) {
	t.Parallel()

	h := newSignIn(t)
	parent := h.browser(t)
	answer := answerAt(t, parent.get(h.google.Deny(h.toGoogle(parent, h.register(hostRedirect, hostName)))))

	if answer.Get("error") != "access_denied" || answer.Get("state") != hostState || answer.Get("iss") != h.served.URL {
		t.Errorf("the host was sent %v, want access_denied with its state and this issuer", answer)
	}
	if lines := h.lines(eventAuthCallback); len(lines) != 1 || lines[0]["outcome"] != "denied" || lines[0]["reason"] != "access_denied" {
		t.Errorf("auth_callback lines = %v, want one denied at Google", lines)
	}
}

// Google's answer that cannot be trusted to come from the browser the sign-in
// began in, or from a sign-in of this server's at all, stops at a page: the
// host hears nothing, since nothing says the parent at this browser is the
// one it would be told about.
func TestACallbackRefusesWhatItCannotTrust(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name   string
		spoil  func(t *testing.T, h *signIn, parent *browser, back string) (*browser, string)
		status int
		reason string
	}{
		{
			name: "a request changed on its way",
			spoil: func(_ *testing.T, _ *signIn, parent *browser, back string) (*browser, string) {
				return parent, respelled(back, "state", func(state string) string { return state[:len(state)-2] + flipped(state[len(state)-2:]) })
			},
			status: http.StatusBadRequest, reason: "unknown_request",
		},
		{
			name: "no request at all",
			spoil: func(_ *testing.T, _ *signIn, parent *browser, back string) (*browser, string) {
				return parent, respelled(back, "state", func(string) string { return "" })
			},
			status: http.StatusBadRequest, reason: "unknown_request",
		},
		{
			name: "a browser with no cookie",
			spoil: func(t *testing.T, h *signIn, _ *browser, back string) (*browser, string) {
				return h.browser(t), back
			},
			status: http.StatusBadRequest, reason: "cookie",
		},
		{
			name: "another sign-in's browser",
			spoil: func(t *testing.T, h *signIn, _ *browser, back string) (*browser, string) {
				other := h.browser(t)
				h.toConsent(other, h.register(hostRedirect, hostName))
				return other, back
			},
			status: http.StatusBadRequest, reason: "cookie",
		},
		{
			name: "a sign-in past its ten minutes",
			spoil: func(_ *testing.T, h *signIn, parent *browser, back string) (*browser, string) {
				h.clock.advance(flightLifetime + time.Second)
				return parent, back
			},
			status: http.StatusBadRequest, reason: "expired",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			h := newSignIn(t)
			parent := h.browser(t)
			back := h.google.Allow(h.toGoogle(parent, h.register(hostRedirect, hostName)))
			parent, back = tc.spoil(t, h, parent, back)

			refused := parent.get(back)
			if refused.status != tc.status || refused.location != "" {
				t.Errorf("GET /oauth/callback = %d to %q, want %d and no redirect", refused.status, refused.location, tc.status)
			}
			if lines := h.lines(eventAuthReject); len(lines) != 1 || lines[0]["reason"] != tc.reason || lines[0]["step"] != stepCallback {
				t.Errorf("auth_reject lines = %v, want one at the callback for %s", lines, tc.reason)
			}
			if lines := h.lines(eventAuthCallback); len(lines) != 0 {
				t.Errorf("auth_callback lines = %v, want none: the sign-in never got that far", lines)
			}
		})
	}
}

// Whatever Google gives back that is not a sign-in the service can keep, the
// host hears of it at its own address, in the protocol's words.
func TestACallbackTellsTheHostWhatWentWrong(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name    string
		answer  googletest.Answer
		back    func(back string) string
		error   string
		outcome string
		reason  string
	}{
		{name: "an expired code", answer: googletest.Answer{Status: http.StatusBadRequest, ErrorCode: "invalid_grant"},
			error: "server_error", outcome: "failed", reason: "code_refused"},
		{name: "Google failing", answer: googletest.Answer{Status: http.StatusServiceUnavailable},
			error: "server_error", outcome: "failed", reason: "google_unavailable"},
		{name: "an identity not proven", answer: googletest.Answer{Nonce: "another-nonce"},
			error: "server_error", outcome: "failed", reason: "identity"},
		{name: "no lasting access", answer: googletest.Answer{NoRefresh: true},
			error: "server_error", outcome: "failed", reason: "no_refresh"},
		{name: "our own client refused", answer: googletest.Answer{Status: http.StatusUnauthorized, ErrorCode: "invalid_client"},
			error: "server_error", outcome: "failed", reason: "client_refused"},
		{name: "the Drive permission unticked", answer: googletest.Answer{Scope: "openid"},
			error: "access_denied", outcome: "denied", reason: "no_drive"},
		{name: "Google's own error", back: func(back string) string {
			return respelled(respelled(back, "code", func(string) string { return "" }), "error", func(string) string { return "server_error" })
		}, error: "server_error", outcome: "failed", reason: "google_error"},
		{name: "no code", back: func(back string) string {
			return respelled(back, "code", func(string) string { return "" })
		}, error: "server_error", outcome: "failed", reason: "no_code"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			h := newSignIn(t)
			parent := h.browser(t)
			back := h.google.Allow(h.toGoogle(parent, h.register(hostRedirect, hostName)))
			h.google.Misbehave(&tc.answer)
			if tc.back != nil {
				back = tc.back(back)
			}

			answer := answerAt(t, parent.get(back))
			if answer.Get("error") != tc.error || answer.Get("error_description") == "" || answer.Has("code") ||
				answer.Get("state") != hostState || answer.Get("iss") != h.served.URL {
				t.Errorf("the host was sent %v, want %s explained, with its state and this issuer", answer, tc.error)
			}
			endedAs(t, h, tc.outcome, tc.reason, tc.back == nil && tc.reason != "no_drive")
		})
	}
}

// The parent who unticked the one file the service keeps in their Drive has
// not allowed the sign-in, and the host is told why in words it can show.
func TestASignInWithoutTheDriveSaysWhy(t *testing.T) {
	t.Parallel()

	h := newSignIn(t)
	parent := h.browser(t)
	back := h.google.Allow(h.toGoogle(parent, h.register(hostRedirect, hostName)))
	h.google.Misbehave(&googletest.Answer{Scope: "openid"})

	answer := answerAt(t, parent.get(back))
	if answer.Get("error") != "access_denied" || !strings.Contains(answer.Get("error_description"), "Google Drive") {
		t.Errorf("the host was sent %v, want access_denied that names Google Drive", answer)
	}
}

// The grant Google gave without the Drive is left as it is. Ending it at
// Google would end every grant of the parent's at this service, the chats
// they have already connected included, for a sign-in they only declined.
func TestAGrantWithoutTheDriveIsLeftAtGoogle(t *testing.T) {
	t.Parallel()

	h := newSignIn(t)
	parent := h.browser(t)
	back := h.google.Allow(h.toGoogle(parent, h.register(hostRedirect, hostName)))
	h.google.Misbehave(&googletest.Answer{Scope: "openid"})

	if answer := answerAt(t, parent.get(back)); answer.Get("error") != "access_denied" {
		t.Errorf("the host was sent %v, want access_denied", answer)
	}
	if got := h.google.Revocations(); len(got) != 0 {
		t.Errorf("Google was asked to end the grant with %q, want the grant left alone", got)
	}
	endedAs(t, h, "denied", "no_drive", false)
}

// Where no Google sign-in is configured, the consent screen is still shown,
// and allowing stops at a page that says signing in is not set up; nothing is
// remembered of the approval.
func TestASignInWithoutGoogleStopsAtAPage(t *testing.T) {
	t.Parallel()

	h := newSignInWithoutGoogle(t)
	parent := h.browser(t)
	screen, request := h.toConsent(parent, h.register(hostRedirect, hostName))
	if policy := screen.header.Get("Content-Security-Policy"); !strings.Contains(policy, "form-action 'self' https://host.example;") {
		t.Errorf("the consent screen's policy = %q, want its form let on to the host alone", policy)
	}

	stopped := parent.post(h.served.URL+"/oauth/consent", url.Values{"request": {request}, "decision": {"allow"}})
	if stopped.status != http.StatusServiceUnavailable || stopped.location != "" || !strings.Contains(stopped.body, "temporarily_unavailable") {
		t.Errorf("POST /oauth/consent = %d to %q, want 503 and a page", stopped.status, stopped.location)
	}
	if parent.cookie("/oauth/authorize", consentCookie) != "" {
		t.Error("an approval was remembered for a sign-in that could not go on")
	}
}

// respelled is an address with one parameter of its query rewritten; an empty
// value takes the parameter out.
func respelled(address, name string, rewrite func(string) string) string {
	parsed, err := url.Parse(address)
	if err != nil {
		return address
	}
	query := parsed.Query()
	if value := rewrite(query.Get(name)); value != "" {
		query.Set(name, value)
	} else {
		query.Del(name)
	}
	parsed.RawQuery = query.Encode()
	return parsed.String()
}

// The sign-in's cookies are sent over HTTPS alone, to its own endpoints alone,
// never read by a page's script, and never on a request another site makes
// but a top-level visit: the one that ties a sign-in to its browser lasts the
// ten minutes of a sign-in, and the one that remembers approvals half a year.
func TestTheSignInsCookiesStayWithIt(t *testing.T) {
	t.Parallel()

	h := newSignIn(t)
	parent := h.browser(t)
	screen, request := h.toConsent(parent, h.register(hostRedirect, hostName))
	allowed := parent.post(h.served.URL+"/oauth/consent", url.Values{"request": {request}, "decision": {"allow"}})

	for _, tc := range []struct {
		answer reply
		name   string
		maxAge int
	}{
		{screen, csrfCookie, int(flightLifetime / time.Second)},
		{allowed, consentCookie, int(approvalLifetime / time.Second)},
	} {
		cookie := setCookie(t, tc.answer, tc.name)
		if !cookie.Secure || !cookie.HttpOnly || cookie.SameSite != http.SameSiteLaxMode || cookie.Path != cookiePath ||
			cookie.Domain != "" || cookie.MaxAge != tc.maxAge {
			t.Errorf("%s = %+v, want Secure, HttpOnly, SameSite=Lax, Path=%s, no domain, Max-Age=%d",
				tc.name, cookie, cookiePath, tc.maxAge)
		}
	}

	back := parent.get(h.google.Allow(allowed.location))
	if dropped := setCookie(t, back, csrfCookie); dropped.MaxAge >= 0 || dropped.Path != cookiePath {
		t.Errorf("%s at the end of the sign-in = %+v, want it taken back", csrfCookie, dropped)
	}
}

// setCookie is the cookie of the name an answer sets.
func setCookie(t *testing.T, answer reply, name string) *http.Cookie {
	t.Helper()

	for _, line := range answer.header.Values("Set-Cookie") {
		cookie, err := http.ParseSetCookie(line)
		if err != nil {
			t.Fatalf("the answer set %q, which is no cookie: %v", line, err)
		}
		if cookie.Name == name {
			return cookie
		}
	}
	t.Fatalf("the answer sets no %s", name)
	return nil
}

// endedAs fails unless the sign-in left one line of its end, with the outcome
// and the reason given, telling what went wrong exactly when it should.
func endedAs(t *testing.T, h *signIn, outcome, reason string, told bool) {
	t.Helper()

	lines := h.lines(eventAuthCallback)
	if len(lines) != 1 || lines[0]["outcome"] != outcome || lines[0]["reason"] != reason {
		t.Fatalf("auth_callback lines = %v, want one %s for %s", lines, outcome, reason)
	}
	if _, said := lines[0]["error"]; said != told {
		t.Errorf("auth_callback line = %v, want what went wrong told: %v", lines[0], told)
	}
}

// A host that carries a mark to turn text around is written escaped, on the
// consent screen and in the line alike, so that neither reads otherwise than
// it was sent.
func TestAHostThatTurnsTextAroundIsWrittenEscaped(t *testing.T) {
	t.Parallel()

	const turning = "https://a%E2%80%AEb.example/cb"
	escaped := "a" + string(rune(92)) + "u202eb.example"
	h := newSignIn(t)
	screen := h.browser(t).get(h.authorizeURL(h.register(turning, hostName), url.Values{"redirect_uri": {turning}}))

	if screen.status != http.StatusOK || !strings.Contains(screen.body, "<bdi>"+escaped+"</bdi>") {
		t.Errorf("GET /oauth/authorize = %d, want the consent screen naming the host as %s", screen.status, escaped)
	}
	if strings.ContainsRune(screen.body, rune(0x202E)) {
		t.Error("the consent screen carries the mark that turns text around")
	}
	if lines := h.lines(eventAuthAuthorize); len(lines) != 1 || lines[0]["redirect_host"] != escaped {
		t.Errorf("auth_authorize lines = %v, want the host written as %s", lines, escaped)
	}
}

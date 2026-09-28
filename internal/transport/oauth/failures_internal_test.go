package oauthserver

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"go.uber.org/zap/zapcore"

	"github.com/MathTrail/mathtrail-standalone/internal/infra/googleauth/googletest"
)

// A step of a sign-in that fails on this side stops at the page of a failure
// of the server's own, and leaves an error line that says which step; nothing
// is issued, and nobody is sent anywhere.
func TestASignInThatFailsHereSaysSo(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name  string
		spoil func(h *signIn)
		step  func(t *testing.T, h *signIn, parent *browser) reply
		where string
	}{
		{"the request cannot be sealed", func(h *signIn) { h.server.flow.flights = brokenSealer{} },
			func(t *testing.T, h *signIn, parent *browser) reply {
				return parent.get(h.authorizeURL(h.register(hostRedirect, hostName), nil))
			}, stepAuthorize},
		{"the approval cannot be sealed", func(h *signIn) { h.server.flow.consents = brokenSealer{} },
			func(t *testing.T, h *signIn, parent *browser) reply {
				_, request := h.toConsent(parent, h.register(hostRedirect, hostName))
				return parent.post(h.served.URL+"/oauth/consent", url.Values{"request": {request}, "decision": {"allow"}})
			}, stepConsent},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			h := newSignIn(t)
			tc.spoil(h)
			failed := tc.step(t, h, h.browser(t))
			if failed.status != http.StatusInternalServerError || failed.location != "" || !strings.Contains(failed.body, "server_error") {
				t.Errorf("status = %d to %q, want 500 and the page of a failure", failed.status, failed.location)
			}
			lines := h.logs.FilterMessage(eventAuthReject).All()
			if len(lines) != 1 || lines[0].Level != zapcore.ErrorLevel || lines[0].ContextMap()["step"] != tc.where {
				t.Errorf("auth_reject lines = %v, want one error at %s", lines, tc.where)
			}
		})
	}
}

// A code that cannot be sealed is a failure of the server's: the host hears
// it at its own address, with no code, and the line of the end is an error.
func TestACodeThatCannotBeSealedIsAFailureOfOurs(t *testing.T) {
	t.Parallel()

	h := newSignIn(t)
	parent := h.browser(t)
	back := h.google.Allow(h.toGoogle(parent, h.register(hostRedirect, hostName)))
	h.server.flow.codes = brokenSealer{}

	answer := answerAt(t, parent.get(back))
	if answer.Get("error") != "server_error" || answer.Has("code") {
		t.Errorf("the host was sent %v, want server_error and no code", answer)
	}
	lines := h.logs.FilterMessage(eventAuthCallback).All()
	if len(lines) != 1 || lines[0].Level != zapcore.ErrorLevel || lines[0].ContextMap()["reason"] != "internal" {
		t.Errorf("auth_callback lines = %v, want one error of ours", lines)
	}
}

// Where no Google sign-in is configured, nothing goes on to Google: not a
// browser that approved the host before, and not an answer that claims to
// come back from Google.
func TestNothingGoesOnToAGoogleThatIsNotThere(t *testing.T) {
	t.Parallel()

	h := newSignInWithoutGoogle(t)
	clientID := h.register(hostRedirect, hostName)
	parent := h.browser(t)
	screen, request := h.toConsent(parent, clientID)
	if screen.status != http.StatusOK {
		t.Fatalf("GET /oauth/authorize status = %d, want the consent screen", screen.status)
	}

	answer := answerAt(t, parent.get(h.served.URL+CallbackPath+"?"+url.Values{"state": {request}, "code": {"a-code"}}.Encode()))
	if answer.Get("error") != "server_error" || answer.Has("code") {
		t.Errorf("the host was sent %v, want server_error and no code", answer)
	}
	if lines := h.lines(eventAuthCallback); len(lines) != 1 || lines[0]["reason"] != "unconfigured" {
		t.Errorf("auth_callback lines = %v, want one failed for want of Google", lines)
	}

	opened, err := h.server.flow.openFlight(request)
	if err != nil {
		t.Fatalf("openFlight() error = %v, want nil", err)
	}
	approving := h.browser(t)
	approving.client.Jar.SetCookies(mustParse(t, h.served.URL+"/oauth/"), approvedCookie(t, h, &opened))
	stopped := approving.get(h.authorizeURL(clientID, nil))
	if stopped.status != http.StatusServiceUnavailable || stopped.location != "" {
		t.Errorf("GET /oauth/authorize approved = %d to %q, want 503 and a page", stopped.status, stopped.location)
	}
}

// approvedCookie is the consent cookie of a browser that approved a request's
// host.
func approvedCookie(t *testing.T, h *signIn, request *flight) []*http.Cookie {
	t.Helper()

	recorded := httptest.NewRecorder()
	fromParent, err := http.NewRequestWithContext(t.Context(), http.MethodPost, h.served.URL+"/oauth/consent", http.NoBody)
	if err != nil {
		t.Fatalf("NewRequest() error = %v, want nil", err)
	}
	if err := h.server.flow.approve(recorded, fromParent, request); err != nil {
		t.Fatalf("approve() error = %v, want nil", err)
	}
	var cookies []*http.Cookie
	for _, line := range recorded.Header().Values("Set-Cookie") {
		cookie, err := http.ParseSetCookie(line)
		if err != nil {
			t.Fatalf("the approval is no cookie: %v", err)
		}
		cookies = append(cookies, cookie)
	}
	return cookies
}

func mustParse(t *testing.T, address string) *url.URL {
	t.Helper()

	parsed, err := url.Parse(address)
	if err != nil {
		t.Fatalf("url.Parse(%q) error = %v, want nil", address, err)
	}
	return parsed
}

// A consent form that does not parse is no answer of the parent's.
func TestAConsentFormThatDoesNotParseIsRefused(t *testing.T) {
	t.Parallel()

	h := newSignIn(t)
	form := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/oauth/consent", strings.NewReader("request=%zz"))
	form.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	answer := httptest.NewRecorder()
	h.server.Consent.ServeHTTP(answer, form)

	if answer.Code != http.StatusBadRequest || answer.Header().Get("Location") != "" {
		t.Errorf("POST /oauth/consent = %d to %q, want 400 and a page", answer.Code, answer.Header().Get("Location"))
	}
	if lines := h.lines(eventAuthReject); len(lines) != 1 || lines[0]["reason"] != "invalid_request" {
		t.Errorf("auth_reject lines = %v, want one for invalid_request", lines)
	}
}

// Google refusing the service's own client is a failure of ours, not the
// parent's: every sign-in would end the same way, so its line is an error.
func TestAClientGoogleRefusesIsAFailureOfOurs(t *testing.T) {
	t.Parallel()

	h := newSignIn(t)
	parent := h.browser(t)
	back := h.google.Allow(h.toGoogle(parent, h.register(hostRedirect, hostName)))
	h.google.Misbehave(&googletest.Answer{Status: http.StatusUnauthorized, ErrorCode: "invalid_client"})

	answer := answerAt(t, parent.get(back))
	if answer.Get("error") != "server_error" || answer.Has("code") {
		t.Errorf("the host was sent %v, want server_error and no code", answer)
	}
	lines := h.logs.FilterMessage(eventAuthCallback).All()
	if len(lines) != 1 || lines[0].Level != zapcore.ErrorLevel || lines[0].ContextMap()["reason"] != "client_refused" {
		t.Errorf("auth_callback lines = %v, want one error: our client was refused", lines)
	}
}

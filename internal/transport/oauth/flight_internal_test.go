package oauthserver

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"slices"
	"strconv"
	"testing"
	"time"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"

	"github.com/MathTrail/mathtrail-standalone/internal/infra/seal"
)

// flowAt is the sign-in of a server under test, whose clock reads what the
// case sets.
func flowAt(t *testing.T, now *time.Time) *flow {
	t.Helper()

	known, _ := knownClients(t, ringOf(t, 'k'), &documents{})
	screens := pagesOf(t)
	ring := ringOf(t, 'k')
	return &flow{
		issuer:   testIssuer,
		resource: testIssuer + resourcePath,
		scope:    "mcp",
		clients:  known,
		flights:  ring.For(seal.PurposeState),
		consents: ring.For(seal.PurposeConsent),
		codes:    ring.For(seal.PurposeCode),
		userID:   ring.UserID,
		pages:    screens,
		events:   known.events,
		now:      func() time.Time { return *now },
	}
}

// aRequest is a request under way, begun when the case says.
func aRequest(started time.Time) *flight {
	return &flight{
		Client:       digestOf("https://claude.ai/oauth/mcp-oauth-client-metadata"),
		Registration: registrationCIMD,
		RedirectURI:  "https://claude.ai/api/mcp/auth_callback",
		State:        "host-state",
		Challenge:    "E9Melhoa2OwvFrEMTJguCHaoeK1t8URWbuGJSstw-cM",
		Resource:     testIssuer + resourcePath,
		Scope:        "mcp",
		Verifier:     "a-verifier",
		Nonce:        "a-nonce",
		Cookie:       digestOf("a-cookie"),
		StartedAt:    started.Unix(),
	}
}

// A request opens while its sign-in may still be under way: not after its ten
// minutes, and not when it says it began further ahead of this clock than
// two clocks may disagree by.
func TestARequestOpensWhileItIsUnderWay(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name  string
		after time.Duration
		want  error
	}{
		{"just begun", 0, nil},
		{"at its last second", flightLifetime, nil},
		{"past its ten minutes", flightLifetime + time.Second, errLateFlight},
		{"from a clock a little ahead", -clockSkew, nil},
		{"from a clock too far ahead", -clockSkew - time.Second, errLateFlight},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			now := testDay
			f := flowAt(t, &now)
			sealed, err := f.sealFlight(aRequest(testDay))
			if err != nil {
				t.Fatalf("sealFlight() error = %v, want nil", err)
			}
			now = testDay.Add(tc.after)
			if _, err := f.openFlight(sealed); !errors.Is(err, tc.want) {
				t.Errorf("openFlight() error = %v, want %v", err, tc.want)
			}
		})
	}
}

// Nothing but a request this server sealed, for itself, as a request, opens as
// one: not a code, not another issuer's request, not one changed since.
func TestOnlyARequestOpensAsOne(t *testing.T) {
	t.Parallel()

	now := testDay
	f := flowAt(t, &now)
	ring := ringOf(t, 'k')
	sealed, err := f.sealFlight(aRequest(testDay))
	if err != nil {
		t.Fatalf("sealFlight() error = %v, want nil", err)
	}
	asCode, err := ring.Seal(seal.PurposeCode, []byte(`{"started_at":0}`), testIssuer)
	if err != nil {
		t.Fatalf("Seal() error = %v, want nil", err)
	}
	elsewhere, err := ring.Seal(seal.PurposeState, []byte(`{"started_at":0}`), "https://other.example")
	if err != nil {
		t.Fatalf("Seal() error = %v, want nil", err)
	}
	notARequest, err := ring.Seal(seal.PurposeState, []byte(`[]`), testIssuer)
	if err != nil {
		t.Fatalf("Seal() error = %v, want nil", err)
	}

	for name, value := range map[string]string{
		"nothing":                  "",
		"a code":                   asCode,
		"another issuer's request": elsewhere,
		"one changed since":        sealed[:len(sealed)-2] + flipped(sealed[len(sealed)-2:]),
		"not a request":            notARequest,
	} {
		if _, err := f.openFlight(value); !errors.Is(err, errForeignFlight) {
			t.Errorf("openFlight(%s) error = %v, want %v", name, err, errForeignFlight)
		}
	}
}

// A request comes back to its browser only with the cookie whose digest it
// carries.
func TestARequestKnowsItsBrowser(t *testing.T) {
	t.Parallel()

	request := aRequest(testDay)
	for name, cookie := range map[string]*http.Cookie{
		"its own cookie":       {Name: csrfCookie, Value: "a-cookie"},
		"another cookie":       {Name: csrfCookie, Value: "another-cookie"},
		"the cookie elsewhere": {Name: consentCookie, Value: "a-cookie"},
		"no cookie":            nil,
	} {
		back := httptest.NewRequestWithContext(t.Context(), http.MethodGet, CallbackPath, http.NoBody)
		if cookie != nil {
			back.AddCookie(cookie)
		}
		if got, want := fromThisBrowser(back, request), name == "its own cookie"; got != want {
			t.Errorf("fromThisBrowser(%s) = %v, want %v", name, got, want)
		}
	}
}

// Digests tell lists of parts apart however the parts are cut.
func TestDigestsTellPartsApart(t *testing.T) {
	t.Parallel()

	seen := map[string][]string{}
	for _, parts := range [][]string{{"ab", "c"}, {"a", "bc"}, {"abc"}, {"abc", ""}, {"", "abc"}, {}} {
		digest := digestOf(parts...)
		if other, clash := seen[digest]; clash {
			t.Errorf("digestOf(%q) = digestOf(%q)", parts, other)
		}
		seen[digest] = parts
	}
}

// An answer is added after whatever query the client's address has, which is
// kept as it was.
func TestAnAnswerKeepsTheAddressesOwnQuery(t *testing.T) {
	t.Parallel()

	for address, want := range map[string]string{
		"https://host.example/cb":           "https://host.example/cb?code=c",
		"https://host.example/cb?":          "https://host.example/cb?code=c",
		"https://host.example/cb?b=2&a=%41": "https://host.example/cb?b=2&a=%41&code=c",
		"https://host.example/cb?a=1&":      "https://host.example/cb?a=1&code=c",
	} {
		if got := withQuery(address, "code=c"); got != want {
			t.Errorf("withQuery(%q) = %q, want %q", address, got, want)
		}
	}
}

// approvalsAfter is what a browser holds after approving requests one after
// another, a day apart, beginning on the day given.
func approvalsAfter(t *testing.T, f *flow, now *time.Time, requests []*flight) []*http.Cookie {
	t.Helper()

	var held []*http.Cookie
	for _, request := range requests {
		approve := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/oauth/consent", http.NoBody)
		for _, cookie := range held {
			approve.AddCookie(cookie)
		}
		answer := httptest.NewRecorder()
		if err := f.approve(answer, approve, request); err != nil {
			t.Fatalf("approve() error = %v, want nil", err)
		}
		held = answer.Result().Cookies()
		*now = now.Add(24 * time.Hour)
	}
	return held
}

// fromBrowser is a request that carries the cookies given.
func fromBrowser(t *testing.T, cookies []*http.Cookie) *http.Request {
	t.Helper()

	request := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/oauth/authorize", http.NoBody)
	for _, cookie := range cookies {
		request.AddCookie(cookie)
	}
	return request
}

// A browser remembers the newest approvals the cookie has room for, each for
// half a year, and the approval of one host at one address is no approval of
// another.
func TestABrowserRemembersItsNewestApprovals(t *testing.T) {
	t.Parallel()

	now := testDay
	f := flowAt(t, &now)
	var requests []*flight
	for i := range maxApprovals + 1 {
		request := aRequest(testDay)
		request.Client = digestOf("client " + strconv.Itoa(i))
		requests = append(requests, request)
	}
	browser := fromBrowser(t, approvalsAfter(t, f, &now, requests))

	if f.approved(browser, requests[0]) {
		t.Error("the oldest of more approvals than the cookie keeps is still remembered")
	}
	for _, request := range requests[1:] {
		if !f.approved(browser, request) {
			t.Errorf("an approval among the newest %d is forgotten", maxApprovals)
		}
	}
	elsewhere := *requests[1]
	elsewhere.RedirectURI = "https://elsewhere.example/cb"
	if f.approved(browser, &elsewhere) {
		t.Error("approving a host approved it for another address back")
	}

	now = testDay.Add(approvalLifetime + 2*24*time.Hour)
	if f.approved(browser, requests[1]) {
		t.Error("an approval older than half a year is still remembered")
	}
	if !f.approved(browser, requests[maxApprovals]) {
		t.Error("the newest approval, younger than half a year, is forgotten")
	}
}

// A cookie this server did not seal, for itself, as approvals, holds none.
func TestOnlyApprovalsSealedHereCount(t *testing.T) {
	t.Parallel()

	now := testDay
	f := flowAt(t, &now)
	request := aRequest(testDay)
	approval := `[{"f":"` + fingerprint(request) + `","t":` + strconv.FormatInt(testDay.Unix(), 10) + `}]`
	ring := ringOf(t, 'k')
	elsewhere, err := ring.Seal(seal.PurposeConsent, []byte(approval), "https://other.example")
	if err != nil {
		t.Fatalf("Seal() error = %v, want nil", err)
	}
	asRequest, err := ring.Seal(seal.PurposeState, []byte(approval), testIssuer)
	if err != nil {
		t.Fatalf("Seal() error = %v, want nil", err)
	}
	notApprovals, err := ring.Seal(seal.PurposeConsent, []byte(`{"f":"`+fingerprint(request)+`"}`), testIssuer)
	if err != nil {
		t.Fatalf("Seal() error = %v, want nil", err)
	}
	for name, value := range map[string]string{
		"written by hand":         approval,
		"sealed for another host": elsewhere,
		"sealed as a request":     asRequest,
		"sealed, of no approvals": notApprovals,
	} {
		if f.approved(fromBrowser(t, []*http.Cookie{{Name: consentCookie, Value: value}}), request) {
			t.Errorf("a cookie %s approved the host", name)
		}
	}
}

// genFlight produces a request under way with every field chosen freely.
func genFlight() gopter.Gen {
	return gen.SliceOfN(10, gen.AnyString()).Map(func(fields []string) *flight {
		return &flight{
			Client: fields[0], Registration: fields[1], RedirectURI: fields[2], State: fields[3],
			Challenge: fields[4], Resource: fields[5], Scope: fields[6], Verifier: fields[7],
			Nonce: fields[8], Cookie: fields[9], StartedAt: testDay.Unix(),
		}
	})
}

func TestRequestsAndApprovalsHoldTheirProperties(t *testing.T) {
	t.Parallel()

	properties := gopter.NewProperties(nil)

	properties.Property("a sealed request opens as itself", prop.ForAll(
		func(request *flight) bool {
			now := testDay
			f := flowAt(t, &now)
			sealed, err := f.sealFlight(request)
			if err != nil {
				return false
			}
			opened, err := f.openFlight(sealed)
			return err == nil && opened == *request
		},
		genFlight(),
	))

	properties.Property("a browser remembers every approval it made, up to what it keeps", prop.ForAll(
		func(clients []string) bool {
			now := testDay
			f := flowAt(t, &now)
			var requests []*flight
			for _, client := range clients {
				request := aRequest(testDay)
				request.Client = client
				requests = append(requests, request)
			}
			browser := fromBrowser(t, approvalsAfter(t, f, &now, requests))
			kept := f.approvals(browser)
			newest := requests[max(0, len(requests)-maxApprovals):]
			return len(kept) <= maxApprovals && !slices.ContainsFunc(newest, func(r *flight) bool { return !f.approved(browser, r) })
		},
		gen.SliceOfN(maxApprovals+5, gen.Identifier()),
	))

	properties.TestingRun(t)
}

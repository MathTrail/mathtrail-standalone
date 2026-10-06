package app_test

import (
	"context"
	"encoding/json"
	"io"
	"maps"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"regexp"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"

	"github.com/MathTrail/mathtrail-standalone/internal/app"
	"github.com/MathTrail/mathtrail-standalone/internal/config"
	"github.com/MathTrail/mathtrail-standalone/internal/domain/profile"
	"github.com/MathTrail/mathtrail-standalone/internal/domain/solver"
	"github.com/MathTrail/mathtrail-standalone/internal/infra/drive/drivetest"
	"github.com/MathTrail/mathtrail-standalone/internal/infra/geoip/geoiptest"
	"github.com/MathTrail/mathtrail-standalone/internal/infra/googleauth/googletest"
	mcpserver "github.com/MathTrail/mathtrail-standalone/internal/transport/mcp"
)

// A family's whole way through the service, played against the service as it
// is built: the parent's browser, the chat host's server and the host's client
// of the protocol, for which the chat's model writes the race. Google's sign-in
// and Drive are stand-ins, and every line the service writes is kept. So is
// everything the family sends or is issued on the way that a line of the log
// must never carry.

// Who the family is to the service, and what they send it.
const (
	// hostRedirect is where the host's sign-in comes back to, and hostName
	// the name the host gives itself to the protocol, which a line names only
	// by the family of hosts it belongs to.
	hostRedirect = "https://claude.ai/api/mcp/auth_callback"
	hostName     = "claude-ai"
	hostState    = "the-host's-own-state"
	hostVerifier = "the-host-verifier-of-forty-three-characters-at-least"
	// The addresses the parent's browser and the host's servers send from, as
	// the platform in front of the service names them.
	familyAddress = "203.0.113.77"
	hostAddress   = "198.51.100.23"
	// parentEmail is the parent's address, which a host may pass on as a hint
	// and a client may name as its contact.
	parentEmail = "parent.of.zoya@example.com"
	// The child, as the parent describes them to the service.
	pseudonym = "Zoya-Quillfeather"
	interest  = "origami-and-comets"
	notes     = "Loses heart after two wrong answers; the aunt is at aunt.zoya@example.org"
	// reason is why the chat's model chose the race itself.
	reason = "Zoya-Quillfeather asked for a race."
	// wrongAnswer is the letter the child picks: the reversed comparison.
	wrongAnswer = "A"
	// strangersToken is a bearer credential this service never issued.
	strangersToken = "not-a-token-of-ours@example.net"
)

// consentRequest is the sealed request the consent screen posts back.
var consentRequest = regexp.MustCompile(`name="request" value="([^"]+)"`)

// openedRequest is the request next_task's words say is open.
var openedRequest = regexp.MustCompile(`Request (\S+) is open`)

// family is one family's way through the service.
type family struct {
	t      *testing.T
	served *httptest.Server
	google *googletest.Server
	drive  *drivetest.Drive
	logs   *observer.ObservedLogs
	// browser is the parent's browser: it keeps the cookies it is given, and
	// follows no redirect by itself, so that every step is seen.
	browser *http.Client
	// clientID, access and refresh are what the host holds once the parent
	// has signed in.
	clientID, access, refresh string
	session                   *mcp.ClientSession
	// secrets are what no line may carry anywhere in it, and words what no
	// line may carry as a value of its own: an option of the race or a letter
	// of an answer, too short to look for inside other text.
	secrets, words []string
}

// newFamily serves the service as it is built, reaching stand-ins for Google's
// sign-in and Drive, to a family that has not signed in yet.
func newFamily(t *testing.T) *family {
	t.Helper()
	return newFamilyWith(t, func(*config.Config) {})
}

// newFamilyWith is newFamily, on a configuration a case changes.
func newFamilyWith(t *testing.T, change func(*config.Config)) *family {
	t.Helper()

	f := &family{t: t, google: googletest.New(t, time.Now), drive: drivetest.New(t)}
	f.keep(parentEmail, pseudonym, interest, notes, reason, strangersToken, familyAddress, hostAddress,
		hostName, hostState, hostVerifier, googletest.ChallengeOf(hostVerifier), sealKey,
		googletest.ClientSecret, googletest.Subject, googletest.AccessToken, googletest.RefreshToken,
		googletest.RenewedAccessToken, googletest.RotatedRefreshToken)
	f.keep(raceWords()...)
	f.words = append(slices.Collect(maps.Values(raceOptions)), profile.DontKnow)
	f.words = append(f.words, solver.Letters()...)

	// The address is known once the listener is, and the service is built for
	// the address it is served at.
	f.served = httptest.NewUnstartedServer(nil)
	cfg := testConfig()
	cfg.PublicURL = "https://" + f.served.Listener.Addr().String()
	cfg.GoogleClientID, cfg.GoogleClientSecret = googletest.ClientID, googletest.ClientSecret
	// The parent's browser comes back from Google at an address the database of
	// countries puts in one; the children are counted under a key of the
	// deployment's, which no line may carry either.
	cfg.CountryDB = geoiptest.Write(t)
	cfg.LearnerKey = learnerKey
	f.keep(learnerKey, geoiptest.City)
	// Named, so that every line names the trace it belongs to.
	cfg.GCPProjectID = "a-project"
	// A lesson played at the pace of a test, and a day with room for one task,
	// so that the lesson reaches the day's limit.
	cfg.RateUserPerMin = 600
	cfg.DailyTasks = 1
	change(cfg)

	core, logs := observer.New(zapcore.DebugLevel)
	f.logs = logs
	container, err := app.NewContainerReaching(t.Context(), cfg, zap.New(core), f.google.Endpoints(), f.drive.Root())
	if err != nil {
		t.Fatalf("NewContainerReaching() error = %v, want nil", err)
	}
	t.Cleanup(func() { container.Close(context.Background()) })
	f.served.Config.Handler = container.Router
	f.served.StartTLS()
	t.Cleanup(f.served.Close)

	jar, err := cookiejar.New(nil)
	if err != nil {
		t.Fatalf("cookiejar.New() error = %v, want nil", err)
	}
	f.browser = &http.Client{
		Transport:     sentFrom{address: familyAddress, next: f.served.Client().Transport},
		Jar:           jar,
		CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse },
	}
	return f
}

// keep adds values no line may carry.
func (f *family) keep(values ...string) {
	for _, value := range values {
		if value != "" {
			f.secrets = append(f.secrets, value)
		}
	}
}

// signIn is the parent signing in: the host registers itself, the parent
// allows it on the consent screen and at Google, and the host's server trades
// the code the parent brings back for the tokens. On the way a client that
// names itself by an address on a private network is turned away before its
// document is asked for.
func (f *family) signIn() {
	f.t.Helper()

	f.register()
	request := f.toConsent()
	allowed := f.browse(http.MethodPost, "/oauth/consent", url.Values{"request": {request}, "decision": {"allow"}})
	toGoogle, err := url.Parse(allowed.location)
	if err != nil || allowed.status != http.StatusSeeOther || !strings.HasPrefix(allowed.location, f.google.URL) {
		f.t.Fatalf("POST /oauth/consent = %d to %q, want 303 to Google", allowed.status, allowed.location)
	}
	f.keep(toGoogle.Query().Get("state"), toGoogle.Query().Get("code_challenge"), toGoogle.Query().Get("nonce"))

	fromGoogle := f.google.Allow(allowed.location)
	called, err := url.Parse(fromGoogle)
	if err != nil {
		f.t.Fatalf("Google sends the parent back to %q, which does not parse: %v", fromGoogle, err)
	}
	f.keep(called.Query().Get("code"))
	back := f.browse(http.MethodGet, fromGoogle, nil)
	toHost, err := url.Parse(back.location)
	if err != nil || toHost.Query().Get("code") == "" {
		f.t.Fatalf("the callback = %d to %q, want the parent sent back to the host with a code", back.status, back.location)
	}
	f.exchange(toHost.Query().Get("code"))
}

// signInAsReviewer is a directory's reviewer signing in as the demo account:
// the host registers itself, the reviewer types a password that is not the
// reviewers' and is stopped, then types theirs on the same consent screen and
// is sent back to the host, which trades the code for the tokens. Google is
// asked for nothing but a renewal of the demo account's grant.
func (f *family) signInAsReviewer(password string) {
	f.t.Helper()

	f.register()
	request := f.toConsent()
	mistyped := password + "x"
	f.keep(password, mistyped)
	if stopped := f.browse(http.MethodPost, "/oauth/consent", url.Values{
		"request": {request}, "decision": {"reviewer"}, "password": {mistyped},
	}); stopped.status != http.StatusForbidden {
		f.t.Fatalf("a password not the reviewers': status %d, want the page that says so", stopped.status)
	}
	back := f.browse(http.MethodPost, "/oauth/consent", url.Values{
		"request": {request}, "decision": {"reviewer"}, "password": {password},
	})
	toHost, err := url.Parse(back.location)
	if err != nil || toHost.Query().Get("code") == "" {
		f.t.Fatalf("POST /oauth/consent = %d to %q, want the reviewer sent back to the host with a code", back.status, back.location)
	}
	f.exchange(toHost.Query().Get("code"))
}

// register is the host registering itself with the service. On the way a
// client that names itself by an address on a private network is turned away
// before its document is asked for.
func (f *family) register() {
	f.t.Helper()

	var registered struct {
		ClientID string `json:"client_id"`
	}
	f.asHost("/oauth/register", "application/json", `{"redirect_uris":["`+hostRedirect+`"],`+
		`"client_name":"Claude for `+parentEmail+`","contacts":["`+parentEmail+`"]}`, http.StatusCreated, &registered)
	f.clientID = registered.ClientID
	f.keep(f.clientID)

	if refused := f.browse(http.MethodGet, f.authorizeURL("https://10.0.0.1/"+parentEmail+"/client"), nil); refused.status != http.StatusBadRequest {
		f.t.Fatalf("a client on a private network: status %d, want the page that refuses it", refused.status)
	}
}

// toConsent takes the browser from the host's request to the consent screen,
// and is the sealed request the screen posts back.
func (f *family) toConsent() string {
	f.t.Helper()

	screen := f.browse(http.MethodGet, f.authorizeURL(f.clientID), nil)
	request := consentRequest.FindStringSubmatch(screen.body)
	if screen.status != http.StatusOK || request == nil {
		f.t.Fatalf("GET /oauth/authorize = %d, want the consent screen with a request to post back", screen.status)
	}
	f.keep(request[1])
	return request[1]
}

// exchange is the host's server trading the code the browser brought back for
// the tokens.
func (f *family) exchange(code string) {
	f.t.Helper()

	f.keep(code)
	var tokens struct {
		Access  string `json:"access_token"`
		Refresh string `json:"refresh_token"`
	}
	f.asHost("/oauth/token", "application/x-www-form-urlencoded", url.Values{
		"grant_type": {"authorization_code"}, "code": {code}, "code_verifier": {hostVerifier},
		"client_id": {f.clientID}, "redirect_uri": {hostRedirect}, "resource": {f.served.URL + "/mcp"},
	}.Encode(), http.StatusOK, &tokens)
	f.access, f.refresh = tokens.Access, tokens.Refresh
	f.keep(f.access, f.refresh)
}

// authorizeURL is where the host sends the parent to sign in, as the client
// given: a code with the challenge of the host's verifier, for the service's
// resource and its one scope — and the parent's address as a hint, which a
// host may add.
func (f *family) authorizeURL(clientID string) string {
	return "/oauth/authorize?" + url.Values{
		"response_type": {"code"}, "client_id": {clientID}, "redirect_uri": {hostRedirect},
		"code_challenge": {googletest.ChallengeOf(hostVerifier)}, "code_challenge_method": {"S256"},
		"state": {hostState}, "resource": {f.served.URL + "/mcp"}, "scope": {"mcp"},
		"login_hint": {parentEmail},
	}.Encode()
}

// presentAStrangersToken is a request to the lessons with a bearer credential
// this service never issued, which signs nobody in.
func (f *family) presentAStrangersToken() {
	f.t.Helper()

	request, err := http.NewRequestWithContext(f.t.Context(), http.MethodPost, f.served.URL+"/mcp",
		strings.NewReader(`{"jsonrpc":"2.0","id":1,"method":"tools/list"}`))
	if err != nil {
		f.t.Fatalf("NewRequest() error = %v, want nil", err)
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Accept", "application/json, text/event-stream")
	answer, err := (&http.Client{Transport: sentFrom{address: hostAddress, token: strangersToken, next: f.served.Client().Transport}}).Do(request)
	if err != nil {
		f.t.Fatalf("POST /mcp: %v", err)
	}
	_ = answer.Body.Close()
	if answer.StatusCode != http.StatusUnauthorized {
		f.t.Fatalf("POST /mcp with a stranger's token: status %d, want %d", answer.StatusCode, http.StatusUnauthorized)
	}
}

// holdALesson is a lesson, as the host's client of the protocol, the chat's
// model and the card go through it: the tools listed and the widget's page
// read, the child's profile made, a race written and answered wrongly, the
// answer told again, the progress read, and then the day's limit, since the
// day has room for one task.
func (f *family) holdALesson() {
	t := f.t
	t.Helper()

	client := mcp.NewClient(&mcp.Implementation{Name: hostName, Version: "1.0.0"}, nil)
	session, err := client.Connect(t.Context(), &mcp.StreamableClientTransport{
		Endpoint:             f.served.URL + "/mcp",
		HTTPClient:           &http.Client{Transport: sentFrom{address: hostAddress, token: f.access, next: f.served.Client().Transport}},
		DisableStandaloneSSE: true,
		MaxRetries:           -1,
	}, nil)
	if err != nil {
		t.Fatalf("Connect() error = %v, want nil", err)
	}
	f.session = session
	if _, err := session.ListTools(t.Context(), nil); err != nil {
		t.Fatalf("ListTools() error = %v, want nil", err)
	}
	if _, err := session.ReadResource(t.Context(), &mcp.ReadResourceParams{URI: mcpserver.WidgetURI}); err != nil {
		t.Fatalf("ReadResource() error = %v, want the widget's page", err)
	}

	if first, _ := f.call("get_profile", map[string]any{}); first.Screen != "first_run" {
		t.Fatalf("get_profile shows %q, want the first sign-in", first.Screen)
	}
	made, _ := f.call("save_profile", map[string]any{
		"pseudonym": pseudonym, "grade": 2, "interests": []string{interest}, "notes": notes,
	})
	if made.Screen != "profile" {
		t.Fatalf("save_profile shows %q, want the profile made", made.Screen)
	}

	answer := map[string]any{"task_id": f.writeTheRace(), "answer": wrongAnswer}
	if told, _ := f.call("submit_answer", answer); told.Result == nil || told.Result.Correct {
		t.Fatalf("the answer %s is told as %+v, want a wrong answer", wrongAnswer, told.Result)
	}
	if again, _ := f.call("submit_answer", answer); again.Result == nil || !again.Result.AlreadyAnswered {
		t.Fatalf("the same answer again is told as %+v, want the recorded one told again", again.Result)
	}
	f.call("get_progress", map[string]any{})
	f.call("read_progress", map[string]any{})

	if limited, _ := f.call("next_task", map[string]any{"language": "en"}); limited.Status != "limited" {
		t.Fatalf("a second task in the day is %q, want it held to the day's limit", limited.Status)
	}
}

// writeTheRace is a race asked for, and is the id of the task the child is
// handed: the card asked for it draws, the package fetched, the card told the
// task is being written, the race asked for again before it is written — which
// hands the same request back —, refused once and accepted, and the card told
// it is on it.
func (f *family) writeTheRace() (taskID string) {
	t := f.t
	t.Helper()

	_, words := f.call("next_task", map[string]any{
		"language": "en", "topic": "logic.ordering", "grade_level": "1-2", "difficulty": 2, "reason": reason,
	})
	requestID := opened(t, words)
	_, packed := f.call("get_package", map[string]any{"request_id": requestID})
	brief := briefOf(t, packed)
	awaited := map[string]any{"request_id": requestID}
	if waiting, _ := f.call("read_task", awaited); waiting.Screen != "coming" {
		t.Fatalf("the card is told %q, want the task being written", waiting.Screen)
	}
	if _, again := f.call("next_task", map[string]any{"language": "en"}); !strings.Contains(again, "Request "+requestID+" is already open") {
		t.Fatalf("next_task asked again says %q, want the request still open handed back", again)
	}
	if refused, _ := f.call("submit_task", race(requestID, brief, false)); refused.Status != "rejected" {
		t.Fatalf("a race with no hint is %q, want it refused", refused.Status)
	}
	handed, _ := f.call("submit_task", race(requestID, brief, true))
	if handed.Screen != "task" || handed.Task == nil {
		t.Fatalf("the race shows %q, want it on the child's card", handed.Screen)
	}
	if shown, _ := f.call("read_task", awaited); shown.Screen != "task" || shown.Task == nil || shown.Task.ID != handed.Task.ID {
		t.Fatalf("the card is told %+v, want the race on it", shown)
	}
	return handed.Task.ID
}

// card is as much of a tool's payload as says how a step went.
type card struct {
	Screen string `json:"screen"`
	Status string `json:"status"`
	Task   *struct {
		ID string `json:"id"`
	} `json:"task"`
	Result *struct {
		Correct         bool `json:"correct"`
		AlreadyAnswered bool `json:"already_answered"`
	} `json:"result"`
}

// call calls a tool as the chat's model does, and is what it showed on the
// card and said in words. A call that failed stops the lesson.
func (f *family) call(tool string, arguments any) (shown card, words string) {
	f.t.Helper()

	result, err := f.session.CallTool(f.t.Context(), &mcp.CallToolParams{Name: tool, Arguments: arguments})
	if err != nil {
		f.t.Fatalf("CallTool(%s) error = %v, want an answer", tool, err)
	}
	for _, block := range result.Content {
		if text, isText := block.(*mcp.TextContent); isText {
			words += text.Text
		}
	}
	if result.IsError {
		f.t.Fatalf("%s failed: %s", tool, words)
	}
	if result.StructuredContent != nil {
		payload, err := json.Marshal(result.StructuredContent)
		if err != nil || json.Unmarshal(payload, &shown) != nil {
			f.t.Fatalf("the payload of %s does not read: %v", tool, err)
		}
	}
	return shown, words
}

// opened is the request next_task's words open.
func opened(t *testing.T, words string) (requestID string) {
	t.Helper()

	id := openedRequest.FindStringSubmatch(words)
	if id == nil {
		t.Fatalf("next_task says %q, want a request opened", words)
	}
	return id[1]
}

// briefOf is the brief of the package get_package's words carry, as a model
// reads it to hand back with the task.
func briefOf(t *testing.T, words string) json.RawMessage {
	t.Helper()

	_, pack, found := strings.Cut(words, "Package:\n")
	var contents struct {
		Brief json.RawMessage `json:"brief"`
	}
	if !found || json.Unmarshal([]byte(pack), &contents) != nil || contents.Brief == nil {
		t.Fatalf("get_package says %q, want the package of the request", words)
	}
	return contents.Brief
}

// race is the race handed in for a request, the brief handed back as it was
// received — or, not whole, without its hint.
func race(requestID string, brief json.RawMessage, whole bool) map[string]any {
	task := raceTask()
	if !whole {
		task["hint"] = ""
	}
	return map[string]any{
		"request_id": requestID, "brief": brief, "task": task, "solver": raceSolver, "self_check": raceSelfCheck(),
	}
}

// renewAndEnd is the host's server renewing the access the parent gave, and
// then ending it with the renewed refresh token.
func (f *family) renewAndEnd() {
	f.t.Helper()

	var renewed struct {
		Access  string `json:"access_token"`
		Refresh string `json:"refresh_token"`
	}
	f.asHost("/oauth/token", "application/x-www-form-urlencoded", url.Values{
		"grant_type": {"refresh_token"}, "refresh_token": {f.refresh}, "client_id": {f.clientID},
		"resource": {f.served.URL + "/mcp"},
	}.Encode(), http.StatusOK, &renewed)
	f.keep(renewed.Access, renewed.Refresh)
	f.asHost("/oauth/revoke", "application/x-www-form-urlencoded", url.Values{
		"token": {renewed.Refresh}, "client_id": {f.clientID},
	}.Encode(), http.StatusOK, nil)
}

// lines are every line the service wrote. The listener is closed first, which
// waits for every request to be over: a line can be written after the answer
// it tells of has left. Where the child's profile lies in the parent's Drive is
// kept then too, and the identifier the profile gives the child, which the name
// a child is counted under is derived from and must never stand in for.
func (f *family) lines() []observer.LoggedEntry {
	f.t.Helper()

	if f.session != nil {
		_ = f.session.Close()
	}
	f.served.Close()
	// A parent's sign-in reaches Drive with the token Google gave it, and a
	// reviewer's with the one Google renewed the demo account's grant for.
	for _, token := range []string{googletest.AccessToken, googletest.RenewedAccessToken} {
		for _, file := range f.drive.Files(token) {
			f.keep(file.ID)
			if kept, err := profile.Parse(file.Content); err == nil {
				f.keep(kept.StudentID)
			}
		}
	}
	return f.logs.All()
}

// page is an answer as the parent's browser gets it.
type page struct {
	status   int
	location string
	body     string
}

// browse sends a request from the parent's browser: an address of the service
// by its path, or anywhere else by its whole address. The value of every
// cookie it is given is kept.
func (f *family) browse(method, address string, form url.Values) page {
	f.t.Helper()

	if strings.HasPrefix(address, "/") {
		address = f.served.URL + address
	}
	body := io.Reader(http.NoBody)
	if form != nil {
		body = strings.NewReader(form.Encode())
	}
	request, err := http.NewRequestWithContext(f.t.Context(), method, address, body)
	if err != nil {
		f.t.Fatalf("NewRequest(%s %s) error = %v, want nil", method, address, err)
	}
	if form != nil {
		request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}
	answer, err := f.browser.Do(request)
	if err != nil {
		f.t.Fatalf("%s %s: %v", method, address, err)
	}
	defer func() { _ = answer.Body.Close() }()
	read, err := io.ReadAll(answer.Body)
	if err != nil {
		f.t.Fatalf("reading the answer to %s %s: %v", method, address, err)
	}
	for _, cookie := range answer.Cookies() {
		f.keep(cookie.Value)
	}
	return page{status: answer.StatusCode, location: answer.Header.Get("Location"), body: string(read)}
}

// asHost posts to the service as the host's own server does — from its own
// address, with no browser and no cookie — and reads the JSON it answers into
// answer, when there is one to read. An answer of another status stops the
// family.
func (f *family) asHost(path, contentType, body string, status int, answer any) {
	f.t.Helper()

	request, err := http.NewRequestWithContext(f.t.Context(), http.MethodPost, f.served.URL+path, strings.NewReader(body))
	if err != nil {
		f.t.Fatalf("NewRequest(%s) error = %v, want nil", path, err)
	}
	request.Header.Set("Content-Type", contentType)
	reply, err := (&http.Client{Transport: sentFrom{address: hostAddress, next: f.served.Client().Transport}}).Do(request)
	if err != nil {
		f.t.Fatalf("POST %s: %v", path, err)
	}
	defer func() { _ = reply.Body.Close() }()
	read, err := io.ReadAll(reply.Body)
	if err != nil {
		f.t.Fatalf("reading the answer to POST %s: %v", path, err)
	}
	if reply.StatusCode != status {
		f.t.Fatalf("POST %s = %d %s, want %d", path, reply.StatusCode, read, status)
	}
	if answer != nil {
		if err := json.Unmarshal(read, answer); err != nil {
			f.t.Fatalf("POST %s answered %s, which does not read: %v", path, read, err)
		}
	}
}

// sentFrom sends every request as coming from the address given, the way the
// platform in front of the service names it, and with the bearer token given
// when there is one.
type sentFrom struct {
	address, token string
	next           http.RoundTripper
}

func (s sentFrom) RoundTrip(r *http.Request) (*http.Response, error) {
	r = r.Clone(r.Context())
	r.Header.Set("X-Forwarded-For", s.address)
	if s.token != "" {
		r.Header.Set("Authorization", "Bearer "+s.token)
	}
	return s.next.RoundTrip(r)
}

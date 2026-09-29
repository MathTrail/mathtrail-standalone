package app_test

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"reflect"
	"runtime"
	"runtime/debug"
	"slices"
	"strings"
	"testing"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest"
	"go.uber.org/zap/zaptest/observer"

	"github.com/MathTrail/mathtrail-standalone/internal/app"
	"github.com/MathTrail/mathtrail-standalone/internal/config"
	"github.com/MathTrail/mathtrail-standalone/internal/domain/profile"
	"github.com/MathTrail/mathtrail-standalone/internal/infra/seal"
	"github.com/MathTrail/mathtrail-standalone/internal/ratelimit"
	"github.com/MathTrail/mathtrail-standalone/internal/store"
	mcpserver "github.com/MathTrail/mathtrail-standalone/internal/transport/mcp"
	"github.com/MathTrail/mathtrail-standalone/internal/widget"
)

// The server must stop because its context was cancelled — the same way a
// signal reaches it — and it must stop without an error.
func TestServerStopsOnContextCancel(t *testing.T) {
	t.Parallel()

	container := newTestContainer(t)
	server := app.NewServer(container)
	if err := server.Listen(t.Context()); err != nil {
		t.Fatalf("Listen() error = %v, want nil", err)
	}

	ctx, cancel := context.WithCancel(t.Context())
	stopped := make(chan error, 1)
	go func() { stopped <- server.Run(ctx) }()

	// The server answers before the shutdown, so the test proves that it
	// stopped rather than that it never started.
	waitForHealthz(t, server.Addr())

	cancel()

	select {
	case err := <-stopped:
		if err != nil {
			t.Fatalf("Run() error = %v, want nil", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("Run() did not return within 5s of the context being cancelled")
	}

	resp, err := http.Get("http://" + server.Addr() + "/health") //nolint:noctx // the request is expected to fail
	if err == nil {
		_ = resp.Body.Close()
		t.Error("the server still answers after shutdown, want a refused connection")
	}
}

// Addr is read from one goroutine while Run binds the listener on another,
// which is how a caller learns the port the kernel picked. Run under the race
// detector, this is the test that says the field is guarded.
func TestAddrIsSafeWhileTheServerStarts(t *testing.T) {
	t.Parallel()

	server := app.NewServer(newTestContainer(t))

	ctx, cancel := context.WithCancel(t.Context())
	stopped := make(chan error, 1)
	go func() { stopped <- server.Run(ctx) }()

	deadline := time.Now().Add(5 * time.Second)
	addr := ""
	for addr == "" && time.Now().Before(deadline) {
		addr = server.Addr()
		runtime.Gosched()
	}
	if addr == "" {
		t.Fatal("Addr() stayed empty for 5s, want the bound address")
	}

	cancel()
	select {
	case err := <-stopped:
		if err != nil {
			t.Fatalf("Run() error = %v, want nil", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("Run() did not return within 5s of the context being cancelled")
	}
}

// Run returns only once the goroutine it serves on has finished, with every line
// of its own log written: a line written after it returned would land in a
// test that has already ended, or in a process that is already on its way out.
func TestRunLogsEverythingBeforeItReturns(t *testing.T) {
	t.Parallel()

	core, logs := observer.New(zapcore.InfoLevel)
	container := newTestContainer(t)
	container.Logger = zap.New(core)
	server := app.NewServer(container)
	if err := server.Listen(t.Context()); err != nil {
		t.Fatalf("Listen() error = %v, want nil", err)
	}

	// Cancelled before Run starts, so it turns to shutting down at once: the
	// moment a goroutine of its own is most likely still to be on its way.
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	if err := server.Run(ctx); err != nil {
		t.Fatalf("Run() error = %v, want nil", err)
	}

	var logged []string
	for _, entry := range logs.All() {
		logged = append(logged, entry.Message)
	}
	if want := []string{"listening", "shutdown requested", "stopped"}; !slices.Equal(logged, want) {
		t.Errorf("logged by the time Run returned: %q, want %q", logged, want)
	}
}

// A request that never finishes holds the server for the drain's share of the
// way out and no longer: the close that follows has to fit in the rest before
// the platform stops the process.
func TestTheDrainLeavesTheCloseItsShare(t *testing.T) {
	t.Parallel()

	cfg := testConfig()
	cfg.ShutdownTimeout = 4 * time.Second
	container := containerFrom(t, cfg)
	core, logs := observer.New(zapcore.InfoLevel)
	container.Logger = zap.New(core)

	entered, released := make(chan struct{}), make(chan struct{})
	t.Cleanup(func() { close(released) })
	container.Router = http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		close(entered)
		<-released
	})
	server := app.NewServer(container)
	if err := server.Listen(t.Context()); err != nil {
		t.Fatalf("Listen() error = %v, want nil", err)
	}

	ctx, cancel := context.WithCancel(t.Context())
	stopped := make(chan error, 1)
	go func() { stopped <- server.Run(ctx) }()
	go func() {
		resp, err := http.Get("http://" + server.Addr() + "/stuck") //nolint:noctx // cut off by the shutdown
		if err == nil {
			_ = resp.Body.Close()
		}
	}()
	<-entered

	started := time.Now()
	cancel()
	select {
	case err := <-stopped:
		// The stop was asked for and it happened: the request it cut off is a
		// warning, not a failure of the process.
		if err != nil {
			t.Errorf("Run() error = %v, want nil", err)
		}
		if warned := logs.FilterMessage("stopped").FilterField(zap.Bool("cut_short", true)).All(); len(warned) != 1 ||
			warned[0].Level != zapcore.WarnLevel {
			t.Errorf("stop lines = %v, want one warning that the drain was cut short", warned)
		}
	case <-time.After(2 * cfg.ShutdownTimeout):
		t.Fatalf("Run() did not return within %v of the context being cancelled", 2*cfg.ShutdownTimeout)
	}

	// Each line halfway to its neighbour, so that the drain is told apart from
	// the close's quarter below it and from the whole way out above it, a
	// second away and half a second away.
	elapsed := time.Since(started)
	if least := (cfg.CloseTimeout() + cfg.DrainTimeout()) / 2; elapsed < least {
		t.Errorf("the drain took %v, want at least %v: it waits out its share of %v",
			elapsed, least, cfg.DrainTimeout())
	}
	if most := (cfg.DrainTimeout() + cfg.ShutdownTimeout) / 2; elapsed >= most {
		t.Errorf("the drain took %v, want less than %v: it is done within its share of %v",
			elapsed, most, cfg.DrainTimeout())
	}
}

// A port that is already taken is reported when the server binds it, not
// swallowed by the goroutine that serves on it.
func TestListenReportsATakenPort(t *testing.T) {
	t.Parallel()

	first := app.NewServer(newTestContainer(t))
	if err := first.Listen(t.Context()); err != nil {
		t.Fatalf("Listen() error = %v, want nil", err)
	}
	// Nothing below reads the first server again, and the listener of a server
	// nothing can reach is closed by the finalizer of its own descriptor as
	// soon as a collection happens to run. That would hand the port back and
	// leave the second server binding a free one.
	defer runtime.KeepAlive(first)

	cfg := testConfig()
	cfg.Port = portOf(t, first.Addr())
	if err := app.NewServer(containerFrom(t, cfg)).Listen(t.Context()); err == nil {
		t.Error("Listen() error = nil, want a refusal on a port that is taken")
	}
}

// The content is read and checked while the container is built, so that a
// catalog or a reference task nobody could use stops the process instead of
// reaching a child.
func TestContainerCarriesTheCheckedContent(t *testing.T) {
	t.Parallel()

	embedded := newTestContainer(t).Content
	if embedded == nil {
		t.Fatal("the container carries no content")
	}
	if _, ok := embedded.Topic("counting.gaps"); !ok {
		t.Error("the content carries no topic catalog")
	}
	if embedded.InstructionsVersion() == "" {
		t.Error("the content carries no version of the instructions")
	}
}

// The line that says the sandbox is built says too what the instance gave the
// runtime: the processors it schedules on, and the soft limit on its heap, so
// that the log of a deployment shows whether the size it was given arrived.
//
// Not parallel: the soft limit belongs to the whole process. It is set here as
// the environment of a deployment sets it, and put back before any test that
// runs in parallel starts.
func TestContainerSaysWhatTheInstanceGaveTheRuntime(t *testing.T) {
	const told = 921 << 20
	before := debug.SetMemoryLimit(told)
	t.Cleanup(func() { debug.SetMemoryLimit(before) })

	core, logs := observer.New(zapcore.InfoLevel)
	container, err := app.NewContainer(t.Context(), testConfig(), zap.New(core))
	if err != nil {
		t.Fatalf("NewContainer() error = %v, want nil", err)
	}
	t.Cleanup(func() { container.Close(context.Background()) })

	built := logs.FilterMessage("solver sandbox built").All()
	if len(built) != 1 {
		t.Fatalf("got %d lines saying the sandbox was built, want 1", len(built))
	}
	fields := built[0].ContextMap()
	if got, want := fields["gomaxprocs"], int64(runtime.GOMAXPROCS(0)); got != want {
		t.Errorf("gomaxprocs: got %v, want %v", got, want)
	}
	if got := fields["memory_limit"]; got != int64(told) {
		t.Errorf("memory_limit: got %v, want %d", got, int64(told))
	}
}

// The key ring is built while the container is, so a key the service could not
// seal with stops the process instead of surfacing at the first sign-in.
func TestContainerCarriesTheKeyRing(t *testing.T) {
	t.Parallel()

	ring := newTestContainer(t).Seal
	if ring == nil {
		t.Fatal("the container carries no key ring")
	}

	value, err := ring.Seal(seal.PurposeTaskAnswer, []byte(`{"answer":"C"}`), "OX1sT9")
	if err != nil {
		t.Fatalf("Seal() error = %v, want nil", err)
	}
	if _, err := ring.Open(seal.PurposeTaskAnswer, value, "OX1sT9"); err != nil {
		t.Errorf("Open() error = %v, want the ring the container built to be usable", err)
	}
}

// The container serves the MCP endpoint behind its sign-in: with the
// development sign-in switched on a client is answered, and without it — the
// way a deployment always runs — nobody is let in.
func TestContainerServesTheEndpointToSomebodySignedInOnly(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		devAuth bool
		want    int
	}{
		{name: "development sign-in", devAuth: true, want: http.StatusOK},
		{name: "no sign-in", devAuth: false, want: http.StatusUnauthorized},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			cfg := testConfig()
			cfg.DevAuth = tc.devAuth
			container := containerFrom(t, cfg)

			req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/mcp",
				strings.NewReader(`{"jsonrpc":"2.0","id":1,"method":"tools/list"}`))
			req.Host = "localhost" // the configured public URL's host
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Accept", "application/json, text/event-stream")
			rec := httptest.NewRecorder()
			container.Router.ServeHTTP(rec, req)

			if rec.Code != tc.want {
				t.Errorf("status = %d, want %d; body = %s", rec.Code, tc.want, rec.Body.String())
			}
		})
	}
}

// Profiles are kept where the sign-in can reach them. Under the development
// sign-in, whose account carries no Google token, they stay in the memory of
// the process, where nobody could open them. Under the real one they go to the
// parent's Drive, which an account without a token cannot reach: it is
// refused before anything is asked of anyone, and not told that it has no
// profile.
func TestContainerKeepsProfilesWhereTheSignInReaches(t *testing.T) {
	t.Parallel()

	dev := testConfig()
	dev.DevAuth = true
	inMemory := containerFrom(t, dev).Store
	devAccount := store.NewAccount(mcpserver.DevAccount, "", time.Time{})
	p := profile.New(profile.Student{ExcludedSkills: []string{}, Grade: 3, Interests: []string{"space"}, Pseudonym: "Mia"},
		"app_test", time.Now())
	if _, err := inMemory.Create(t.Context(), devAccount, p); err != nil {
		t.Fatalf("Create() under the development sign-in error = %v, want nil", err)
	}
	if location, err := inMemory.Export(t.Context(), devAccount); err != nil || !reflect.DeepEqual(location, store.Location{}) {
		t.Errorf("Export() under the development sign-in = %+v, %v, want the memory's zero location", location, err)
	}

	inDrive := containerFrom(t, testConfig()).Store
	_, _, err := inDrive.Load(t.Context(), store.NewAccount("a-parent", "", time.Time{}))
	if err == nil || errors.Is(err, store.ErrNotFound) {
		t.Errorf("Load() for an account without a token under the real sign-in: error = %v, want a refusal that is not %v",
			err, store.ErrNotFound)
	}
}

// A client the endpoint refuses is led to the sign-in, and every step of the
// way is served under the service's own name: the refusal names the resource's
// metadata, that document names the endpoint as the resource and the service
// as its authorization server, whose metadata names a registration endpoint
// that registers.
func TestContainerLeadsARefusedClientToTheSignIn(t *testing.T) {
	t.Parallel()

	container := containerFrom(t, testConfig())
	refused := serveOne(t, container, http.MethodPost, "/mcp", `{"jsonrpc":"2.0","id":1,"method":"tools/list"}`)
	metadata := "http://localhost/.well-known/oauth-protected-resource/mcp"
	if got, want := refused.Header().Get("WWW-Authenticate"), `Bearer resource_metadata="`+metadata+`", scope="mcp"`; got != want {
		t.Fatalf("WWW-Authenticate = %q, want %q", got, want)
	}

	var resource struct {
		Resource             string   `json:"resource"`
		AuthorizationServers []string `json:"authorization_servers"`
	}
	decodeAnswer(t, serveOne(t, container, http.MethodGet, strings.TrimPrefix(metadata, "http://localhost"), ""), &resource)
	if resource.Resource != "http://localhost/mcp" || !slices.Equal(resource.AuthorizationServers, []string{"http://localhost"}) {
		t.Fatalf("the resource's metadata is %+v, want the endpoint and the service", resource)
	}

	var server struct {
		RegistrationEndpoint string `json:"registration_endpoint"`
	}
	decodeAnswer(t, serveOne(t, container, http.MethodGet, "/.well-known/oauth-authorization-server", ""), &server)
	registered := serveOne(t, container, http.MethodPost, strings.TrimPrefix(server.RegistrationEndpoint, "http://localhost"),
		`{"redirect_uris":["https://claude.ai/api/mcp/auth_callback"],"client_name":"Claude"}`)
	if registered.Code != http.StatusCreated || !strings.Contains(registered.Body.String(), `"client_id":"mt1.d.`) {
		t.Errorf("a registration at %s: status %d, body %s; want a client registered", server.RegistrationEndpoint,
			registered.Code, registered.Body.String())
	}
}

// serveOne sends the container one request under its own name.
func serveOne(t *testing.T, container *app.Container, method, path, body string) *httptest.ResponseRecorder {
	t.Helper()

	req := httptest.NewRequestWithContext(t.Context(), method, path, strings.NewReader(body))
	req.Host = "localhost" // the configured public URL's host
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json, text/event-stream")
	rec := httptest.NewRecorder()
	container.Router.ServeHTTP(rec, req)
	return rec
}

// serveForm posts a form to the container's router, as a host's server posts
// to the sign-in's endpoints.
func serveForm(t *testing.T, container *app.Container, path, form string) *httptest.ResponseRecorder {
	t.Helper()

	req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, path, strings.NewReader(form))
	req.Host = "localhost" // the configured public URL's host
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	container.Router.ServeHTTP(rec, req)
	return rec
}

// decodeAnswer reads a JSON answer, which has to be a 200.
func decodeAnswer(t *testing.T, rec *httptest.ResponseRecorder, into any) {
	t.Helper()

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body = %s", rec.Code, http.StatusOK, rec.Body.String())
	}
	if err := json.Unmarshal(rec.Body.Bytes(), into); err != nil {
		t.Fatalf("the answer is not JSON: %v; body = %s", err, rec.Body.String())
	}
}

// The container serves the widget the binary carries — the build, or the
// placeholder where none ran — as the page every card is drawn by.
func TestContainerServesTheWidgetItCarries(t *testing.T) {
	t.Parallel()

	cfg := testConfig()
	cfg.DevAuth = true
	container := containerFrom(t, cfg)

	req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/mcp", strings.NewReader(
		`{"jsonrpc":"2.0","id":1,"method":"resources/read","params":{"uri":"`+mcpserver.WidgetURI+`"}}`))
	req.Host = "localhost" // the configured public URL's host
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json, text/event-stream")
	rec := httptest.NewRecorder()
	container.Router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body = %s", rec.Code, http.StatusOK, rec.Body.String())
	}
	var answer struct {
		Result struct {
			Contents []struct {
				Text string `json:"text"`
			} `json:"contents"`
		} `json:"result"`
	}
	if err := json.Unmarshal(jsonRPCMessage(t, rec), &answer); err != nil {
		t.Fatalf("the answer's message is not JSON: %v; body = %s", err, rec.Body.String())
	}
	contents := answer.Result.Contents
	if len(contents) != 1 {
		t.Fatalf("contents = %d, want the page alone", len(contents))
	}
	if got, want := contents[0].Text, widget.Page(); got != want {
		t.Errorf("page served = %d bytes starting %q, want the %d bytes the binary carries, starting %q",
			len(got), got[:min(len(got), 40)], len(want), want[:min(len(want), 40)])
	}
}

// The container serves the tools of the lesson over the store it built: a
// development account sees the six of them, a first call finds no profile yet,
// and no task can be asked for without one.
func TestContainerServesTheToolsOfTheLesson(t *testing.T) {
	t.Parallel()

	cfg := testConfig()
	cfg.DevAuth = true
	container := containerFrom(t, cfg)

	var listed struct {
		Result struct {
			Tools []struct {
				Name string `json:"name"`
			} `json:"tools"`
		} `json:"result"`
	}
	askEndpoint(t, container, `{"jsonrpc":"2.0","id":1,"method":"tools/list"}`, &listed)
	var names []string
	for _, tool := range listed.Result.Tools {
		names = append(names, tool.Name)
	}
	slices.Sort(names)
	want := []string{
		"get_profile", "get_progress", "next_task", "read_progress", "save_profile", "submit_answer", "submit_task",
	}
	if !slices.Equal(names, want) {
		t.Errorf("tools = %v, want %v", names, want)
	}

	// Nothing is kept yet: the profile shows the first sign-in on its screen,
	// and a task asked for says there is no profile to write one for, in the
	// words alone it answers in.
	for _, call := range []struct {
		tool, arguments string
		inWords         bool
	}{
		{tool: "get_profile", arguments: `{}`},
		{tool: "next_task", arguments: `{"language":"en"}`, inWords: true},
	} {
		payload, words := callTool(t, container, call.tool, call.arguments)
		switch {
		case call.inWords && (payload != nil || !strings.Contains(words, "There is no profile yet")):
			t.Errorf("%s answers with the payload %+v and says %q, want the first sign-in in words alone",
				call.tool, payload, words)
		case !call.inWords && (payload == nil || payload.Screen != "first_run"):
			t.Errorf("%s shows %+v, want the first_run screen: nothing is kept yet", call.tool, payload)
		}
	}
}

// screenPayload is as much of a tool's payload as says which screen a card
// draws.
type screenPayload struct {
	Screen string `json:"screen"`
}

// callTool calls a tool through the container's MCP endpoint, and reads its
// payload, nil when it has none, and the first block of its words.
func callTool(t *testing.T, container *app.Container, tool, arguments string) (payload *screenPayload, words string) {
	t.Helper()

	var answered struct {
		Result struct {
			Content []struct {
				Text string `json:"text"`
			} `json:"content"`
			StructuredContent *screenPayload `json:"structuredContent"`
		} `json:"result"`
	}
	askEndpoint(t, container, `{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"`+tool+
		`","arguments":`+arguments+`}}`, &answered)
	if len(answered.Result.Content) > 0 {
		words = answered.Result.Content[0].Text
	}
	return answered.Result.StructuredContent, words
}

// askEndpoint sends one message to the container's MCP endpoint, as a client
// of the protocol's older versions would, and reads the answer into answer.
func askEndpoint(t *testing.T, container *app.Container, message string, answer any) {
	t.Helper()

	req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/mcp", strings.NewReader(message))
	req.Host = "localhost" // the configured public URL's host
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json, text/event-stream")
	rec := httptest.NewRecorder()
	container.Router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body = %s", rec.Code, http.StatusOK, rec.Body.String())
	}
	if err := json.Unmarshal(jsonRPCMessage(t, rec), answer); err != nil {
		t.Fatalf("the answer's message is not JSON: %v; body = %s", err, rec.Body.String())
	}
}

// jsonRPCMessage is the message an answer of the MCP endpoint carries, whether
// it came as a JSON body or as the data of the one event of a stream.
func jsonRPCMessage(t *testing.T, rec *httptest.ResponseRecorder) []byte {
	t.Helper()

	body := rec.Body.Bytes()
	if !strings.HasPrefix(rec.Header().Get("Content-Type"), "text/event-stream") {
		return body
	}
	for line := range strings.SplitSeq(string(body), "\n") {
		if data, isData := strings.CutPrefix(line, "data: "); isData {
			return []byte(data)
		}
	}
	t.Fatalf("the stream carries no message; body = %s", body)
	return nil
}

func TestContainerRefusesAKeyItCannotRead(t *testing.T) {
	t.Parallel()

	cfg := testConfig()
	cfg.SealKeyCurrent = "not a key"

	container, err := app.NewContainer(t.Context(), cfg, zaptest.NewLogger(t))
	if !errors.Is(err, seal.ErrKey) {
		t.Fatalf("NewContainer() error = %v, want %v", err, seal.ErrKey)
	}
	if container != nil {
		t.Error("NewContainer() returned a container alongside an error, want nothing")
	}
}

func newTestContainer(t *testing.T) *app.Container {
	t.Helper()

	return containerFrom(t, testConfig())
}

// containerFrom builds a container from a configuration a case has changed,
// and closes it when the case is over.
func containerFrom(t *testing.T, cfg *config.Config) *app.Container {
	t.Helper()

	container, err := app.NewContainer(t.Context(), cfg, zaptest.NewLogger(t))
	if err != nil {
		t.Fatalf("NewContainer() error = %v, want nil", err)
	}
	t.Cleanup(func() { container.Close(context.Background()) })
	return container
}

func testConfig() *config.Config {
	return &config.Config{
		Port:               "0", // the kernel picks a free one
		PublicURL:          "http://localhost",
		LogLevel:           "debug",
		LogFormat:          "console",
		ReadHeaderTimeout:  config.DefaultReadHeaderTimeout,
		ReadTimeout:        config.DefaultReadTimeout,
		WriteTimeout:       config.DefaultWriteTimeout,
		IdleTimeout:        config.DefaultIdleTimeout,
		ShutdownTimeout:    time.Second,
		SealKeyCurrent:     sealKey,
		SolverSteps:        config.DefaultSolverSteps,
		SolverTimeout:      config.DefaultSolverTimeout,
		SolverConcurrency:  config.DefaultSolverConcurrency,
		SolverWait:         config.DefaultSolverWait,
		RequestWindow:      config.DefaultRequestWindow,
		RateUserPerMin:     config.DefaultRateUserPerMin,
		RateIPPerMin:       config.DefaultRateIPPerMin,
		RateInstancePerMin: config.DefaultRateInstancePerMin,
		DailyTasks:         config.DefaultDailyTasks,
		DailyFailed:        config.DefaultDailyFailed,
		DriveTimeout:       config.DefaultDriveTimeout,
		SiteURL:            config.DefaultSiteURL,
	}
}

// sealKey is a key for the tests of the wiring: the container only has to be
// able to read it, and nothing here seals anything with it.
var sealKey = base64.StdEncoding.EncodeToString(bytes.Repeat([]byte("mathtrail"), 4)[:32])

func portOf(t *testing.T, addr string) string {
	t.Helper()

	_, port, err := net.SplitHostPort(addr)
	if err != nil {
		t.Fatalf("address %q has no port: %v", addr, err)
	}
	return port
}

func waitForHealthz(t *testing.T, addr string) {
	t.Helper()

	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		resp, err := http.Get("http://" + addr + "/health") //nolint:noctx // a probe with its own deadline loop
		if err == nil {
			_ = resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				return
			}
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("no healthy answer from %s within 5s", addr)
}

// The instance keeps a pace at each of its two doors apart: a flood at the
// sign-in, which anybody reaches, spends nothing of the pace the lessons are
// held to, so a child in the middle of one is not held back by it.
func TestAFloodAtTheSignInDoesNotHoldBackTheLessons(t *testing.T) {
	t.Parallel()

	cfg := testConfig()
	cfg.DevAuth = true
	cfg.RateInstancePerMin = 3
	container := containerFrom(t, cfg)

	serveOne(t, container, http.MethodPost, "/oauth/register", `{}`)
	if flooded := serveOne(t, container, http.MethodPost, "/oauth/register", `{}`); flooded.Code != http.StatusTooManyRequests {
		t.Fatalf("the sign-in past its pace: status = %d, want %d", flooded.Code, http.StatusTooManyRequests)
	}
	lesson := serveOne(t, container, http.MethodPost, "/mcp", `{"jsonrpc":"2.0","id":1,"method":"tools/list"}`)
	if lesson.Code != http.StatusOK || strings.Contains(lesson.Body.String(), "too many requests") {
		t.Errorf("a lesson after the flood at the sign-in: status %d, body %s, want it answered", lesson.Code, lesson.Body.String())
	}
	// A lesson renews its access at the sign-in's token endpoint every few
	// minutes, so the flood must not hold that back either. This renewal is
	// refused for what it carries, which is the endpoint's own answer, not the
	// instance's pace.
	renewal := serveForm(t, container, "/oauth/token", "grant_type=refresh_token&refresh_token=spent&client_id=https://client.example/c")
	if renewal.Code == http.StatusTooManyRequests {
		t.Errorf("a renewal after the flood at the sign-in: status %d, want it answered by the endpoint", renewal.Code)
	}
}

// A pace nothing could be let in at stops the container where it is built,
// whichever pace it is.
func TestContainerRefusesAPaceNothingCouldBeLetInAt(t *testing.T) {
	t.Parallel()

	for name, spoil := range map[string]func(*config.Config){
		"an account's":   func(cfg *config.Config) { cfg.RateUserPerMin = 0 },
		"an address's":   func(cfg *config.Config) { cfg.RateIPPerMin = 0 },
		"the instance's": func(cfg *config.Config) { cfg.RateInstancePerMin = 0 },
	} {
		cfg := testConfig()
		spoil(cfg)
		container, err := app.NewContainer(t.Context(), cfg, zaptest.NewLogger(t))
		if !errors.Is(err, ratelimit.ErrSettings) || container != nil {
			t.Errorf("%s pace of none: NewContainer() = %v, %v, want no container and ErrSettings", name, container, err)
		}
	}
}

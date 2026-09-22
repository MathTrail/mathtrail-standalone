package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"

	"github.com/MathTrail/mathtrail-standalone/internal/transport/http/middleware"
)

// The sign-in carries codes and tokens in query strings, and a log line is the
// one place they must not end up.
func TestSecretsInTheQueryAreMasked(t *testing.T) {
	t.Parallel()

	logs, router := routerWithObservedLogs(t)
	router.GET("/oauth/callback", func(c *gin.Context) { c.Status(http.StatusBadRequest) })

	serve(t, router, http.MethodGet, "/oauth/callback?code=secret-code&state=secret-state&error=access_denied")

	line := onlyLine(t, logs)
	query := field(t, line, "query")
	for _, secret := range []string{"secret-code", "secret-state"} {
		if strings.Contains(query, secret) {
			t.Errorf("query = %q, want %q masked", query, secret)
		}
	}
	if !strings.Contains(query, "access_denied") {
		t.Errorf("query = %q, want the non-secret parameters kept", query)
	}
}

// A probe that answers normally is not worth a line; a probe that fails is.
func TestSuccessfulProbesAreNotLogged(t *testing.T) {
	t.Parallel()

	logs, router := routerWithObservedLogs(t)
	router.GET("/health", func(c *gin.Context) { c.Status(http.StatusOK) })

	serve(t, router, http.MethodGet, "/health")

	if logs.Len() != 0 {
		t.Errorf("a successful probe wrote %d lines, want none", logs.Len())
	}
}

func TestFailuresAreLoggedWithTheirRequestID(t *testing.T) {
	t.Parallel()

	logs, router := routerWithObservedLogs(t)
	router.GET("/health", func(c *gin.Context) { c.Status(http.StatusServiceUnavailable) })

	serve(t, router, http.MethodGet, "/health")

	line := onlyLine(t, logs)
	if line.Level != zapcore.ErrorLevel {
		t.Errorf("level = %v, want %v for a 5xx", line.Level, zapcore.ErrorLevel)
	}
	if field(t, line, "request_id") == "" {
		t.Error("request_id is empty, want the id the request carried")
	}
}

// One request must not take the process down: another child is in the middle
// of a task on the same instance.
func TestAPanicBecomesAnAnswer(t *testing.T) {
	t.Parallel()

	logs, router := routerWithObservedLogs(t)
	router.GET("/boom", func(_ *gin.Context) { panic("boom") })

	rec := serve(t, router, http.MethodGet, "/boom")

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusInternalServerError)
	}
	if !strings.Contains(rec.Body.String(), "INTERNAL_ERROR") {
		t.Errorf("body = %s, want a structured error", rec.Body.String())
	}
	if rec.Body.String() != "" && strings.Contains(rec.Body.String(), "boom") {
		t.Errorf("body = %s, want the panic's text kept out of the answer", rec.Body.String())
	}
	if logs.FilterMessage("panic").Len() != 1 {
		t.Error("the panic was not logged once")
	}
}

func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)
	os.Exit(m.Run())
}

func routerWithObservedLogs(t *testing.T) (*observer.ObservedLogs, *gin.Engine) {
	t.Helper()

	core, logs := observer.New(zapcore.DebugLevel)
	logger := zap.New(core)

	router := gin.New()
	router.Use(middleware.RequestID())
	router.Use(middleware.ZapRecovery(logger))
	router.Use(middleware.ZapLogger(logger))
	return logs, router
}

func serve(t *testing.T, router *gin.Engine, method, target string) *httptest.ResponseRecorder {
	t.Helper()

	req := httptest.NewRequestWithContext(t.Context(), method, target, http.NoBody)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	return rec
}

func onlyLine(t *testing.T, logs *observer.ObservedLogs) *observer.LoggedEntry {
	t.Helper()

	entries := logs.FilterMessage("http_request").All()
	if len(entries) != 1 {
		t.Fatalf("http_request lines = %d, want exactly 1", len(entries))
	}
	return &entries[0]
}

func field(t *testing.T, line *observer.LoggedEntry, key string) string {
	t.Helper()

	value, found := line.ContextMap()[key]
	if !found {
		t.Fatalf("the line has no %q field: %v", key, line.ContextMap())
	}
	text, ok := value.(string)
	if !ok {
		t.Fatalf("%q = %T, want a string", key, value)
	}
	return text
}

// A parameter name may be percent-encoded, so a search of the raw query text
// finds nothing while the handler — and anything else that parses — sees a
// secret. This is the test that stops a shortcut being added to maskQuery.
func TestPercentEncodedSecretNamesAreMasked(t *testing.T) {
	t.Parallel()

	logs, router := routerWithObservedLogs(t)
	router.GET("/oauth/callback", func(c *gin.Context) { c.Status(http.StatusBadRequest) })

	serve(t, router, http.MethodGet, "/oauth/callback?%63ode=super-secret&%73tate=also-secret")

	query := field(t, onlyLine(t, logs), "query")
	for _, secret := range []string{"super-secret", "also-secret"} {
		if strings.Contains(query, secret) {
			t.Errorf("query = %q, want %q masked even with an encoded parameter name", query, secret)
		}
	}
}

// An answer with no body has a body of zero, not the -1 the writer starts at.
func TestAnEmptyBodyIsLoggedAsZero(t *testing.T) {
	t.Parallel()

	logs, router := routerWithObservedLogs(t)
	router.GET("/gone", func(c *gin.Context) { c.Status(http.StatusNotFound) })

	serve(t, router, http.MethodGet, "/gone")

	size, found := onlyLine(t, logs).ContextMap()["body_size"]
	if !found {
		t.Fatal("the line has no body_size field")
	}
	if size != int64(0) {
		t.Errorf("body_size = %v, want 0", size)
	}
}

// A handler that panicked halfway through its answer has already sent a status
// and headers, and a second one would corrupt what the client is reading.
func TestAPanicAfterTheAnswerHasStartedDoesNotWriteTwice(t *testing.T) {
	t.Parallel()

	logs, router := routerWithObservedLogs(t)
	router.GET("/half", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"first": "part"})
		panic("boom")
	})

	rec := serve(t, router, http.MethodGet, "/half")

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want the %d the handler already sent", rec.Code, http.StatusOK)
	}
	if strings.Contains(rec.Body.String(), "INTERNAL_ERROR") {
		t.Errorf("body = %s, want no second answer appended to the first", rec.Body.String())
	}
	if logs.FilterMessage("panic").Len() != 1 {
		t.Error("the panic was not logged once")
	}
}

// net/http documents ErrAbortHandler as the way to drop a request silently,
// and the framework aborts on it without calling our recovery at all. This
// test is what notices if that ever changes and a silent abort starts being
// answered with 500.
func TestASilentAbortIsNotTreatedAsAPanic(t *testing.T) {
	t.Parallel()

	logs, router := routerWithObservedLogs(t)
	router.GET("/abort", func(_ *gin.Context) { panic(http.ErrAbortHandler) })

	rec := serve(t, router, http.MethodGet, "/abort")

	if logs.FilterMessage("panic").Len() != 0 {
		t.Error("a silent abort was logged as a panic")
	}
	if strings.Contains(rec.Body.String(), "INTERNAL_ERROR") {
		t.Errorf("body = %s, want no answer to a request that asked for none", rec.Body.String())
	}
}

// An id the client dictates is echoed back and written into every line about
// the request, so what a client may dictate is kept to what an id can be.
func TestOnlyAUsableRequestIDIsKept(t *testing.T) {
	t.Parallel()

	kept := []struct {
		name string
		id   string
	}{
		{"a uuid", "01a0c467-8008-75c3-825b-37f804bf11bb"},
		{"a hex trace id", "4bf92f3577b34da6a3ce929d0e0e4736"},
		{"a traceparent", "00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01"},
	}
	for _, tc := range kept {
		t.Run("kept: "+tc.name, func(t *testing.T) {
			t.Parallel()

			if got := echoedRequestID(t, tc.id); got != tc.id {
				t.Errorf("X-Request-ID = %q, want the id the client sent", got)
			}
		})
	}

	replaced := []struct {
		name string
		id   string
	}{
		{"a newline", "abc\r\nX-Injected: yes"},
		{"a control byte", "abc\x00def"},
		{"a quote", `abc"def`},
		{"a space", "two words"},
		{"letters outside ASCII", "идентификатор"},
		{"longer than the limit", strings.Repeat("a", 65)},
	}
	for _, tc := range replaced {
		t.Run("replaced: "+tc.name, func(t *testing.T) {
			t.Parallel()

			got := echoedRequestID(t, tc.id)
			if got == tc.id {
				t.Errorf("X-Request-ID = %q, want it replaced", got)
			}
			if !usableLooking(got) {
				t.Errorf("X-Request-ID = %q, want a plain identifier", got)
			}
		})
	}
}

func echoedRequestID(t *testing.T, sent string) string {
	t.Helper()

	router := gin.New()
	router.Use(middleware.RequestID())
	router.GET("/x", func(c *gin.Context) { c.Status(http.StatusOK) })

	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/x", http.NoBody)
	req.Header.Set(middleware.RequestIDHeader, sent)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	return rec.Header().Get(middleware.RequestIDHeader)
}

func usableLooking(id string) bool {
	if id == "" {
		return false
	}
	for _, r := range id {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9':
		case r == '-', r == '_', r == '.', r == ':':
		default:
			return false
		}
	}
	return true
}

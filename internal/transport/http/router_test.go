package httpserver_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	metricnoop "go.opentelemetry.io/otel/metric/noop"
	tracenoop "go.opentelemetry.io/otel/trace/noop"

	"go.uber.org/zap/zaptest"

	"github.com/MathTrail/mathtrail-standalone/internal/apierror"
	httpserver "github.com/MathTrail/mathtrail-standalone/internal/transport/http"
	"github.com/MathTrail/mathtrail-standalone/internal/transport/http/middleware"
)

// The framework keeps its mode in a package variable. A test binary sets it
// once, here, rather than letting a constructor decide it for every other test
// running beside it.
func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)
	os.Exit(m.Run())
}

// A probe that asks with HEAD is told the service is up.
func TestHealthAnswersHead(t *testing.T) {
	t.Parallel()

	if rec := call(t, http.MethodHead, "/health"); rec.Code != http.StatusOK {
		t.Errorf("HEAD /health status = %d, want %d", rec.Code, http.StatusOK)
	}
}

func TestHealth(t *testing.T) {
	t.Parallel()

	rec := call(t, http.MethodGet, "/health")

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if got := rec.Header().Get("Cache-Control"); got != "no-store" {
		t.Errorf("Cache-Control = %q, want %q", got, "no-store")
	}

	var body map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("body is not JSON: %v; body = %s", err, rec.Body.String())
	}
	if body["status"] != "ok" {
		t.Errorf("status field = %q, want %q", body["status"], "ok")
	}
	if body["version"] == "" {
		t.Error("version field is empty, want the build's version")
	}
}

func TestEveryAnswerCarriesARequestID(t *testing.T) {
	t.Parallel()

	rec := call(t, http.MethodGet, "/health")

	if rec.Header().Get(middleware.RequestIDHeader) == "" {
		t.Errorf("%s is empty, want an id on every answer", middleware.RequestIDHeader)
	}
}

func TestAClientsRequestIDIsKept(t *testing.T) {
	t.Parallel()

	router := newRouter(t)
	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/health", http.NoBody)
	req.Header.Set(middleware.RequestIDHeader, "from-the-caller")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if got := rec.Header().Get(middleware.RequestIDHeader); got != "from-the-caller" {
		t.Errorf("%s = %q, want the id the client sent", middleware.RequestIDHeader, got)
	}
}

func TestRefusals(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		method     string
		path       string
		wantStatus int
		wantCode   string
	}{
		{
			name:       "unknown path",
			method:     http.MethodGet,
			path:       "/nope",
			wantStatus: http.StatusNotFound,
			wantCode:   apierror.CodeNotFound,
		},
		{
			name:       "wrong method",
			method:     http.MethodPost,
			path:       "/health",
			wantStatus: http.StatusMethodNotAllowed,
			wantCode:   apierror.CodeMethodNotAllowed,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			rec := call(t, tc.method, tc.path)

			if rec.Code != tc.wantStatus {
				t.Fatalf("status = %d, want %d", rec.Code, tc.wantStatus)
			}
			var body apierror.Response
			if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
				t.Fatalf("body is not an apierror.Response: %v; body = %s", err, rec.Body.String())
			}
			if body.Code != tc.wantCode {
				t.Errorf("code = %q, want %q", body.Code, tc.wantCode)
			}
			if body.Message == "" {
				t.Error("message is empty, want a sentence that can be relayed")
			}
		})
	}
}

func call(t *testing.T, method, path string) *httptest.ResponseRecorder {
	t.Helper()

	router := newRouter(t)
	req := httptest.NewRequestWithContext(t.Context(), method, path, http.NoBody)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	return rec
}

// A 405 must say which methods would have worked (RFC 9110). The framework
// fills the header in; this is the test that notices if it stops.
func TestA405SaysWhichMethodsWork(t *testing.T) {
	t.Parallel()

	rec := call(t, http.MethodPost, "/health")

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusMethodNotAllowed)
	}
	if got, want := rec.Header().Get("Allow"), "GET, HEAD"; got != want {
		t.Errorf("Allow = %q, want %q", got, want)
	}
}

// newRouter builds the router with telemetry that keeps nothing. These cases
// are about routing, and providers that recorded what they were given would
// only add noise to them.
func newRouter(t *testing.T) http.Handler {
	t.Helper()

	router, err := httpserver.NewRouter(httpserver.NewHealthHandler(), zaptest.NewLogger(t), httpserver.Observability{
		Traces: tracenoop.NewTracerProvider(),
		Meters: metricnoop.NewMeterProvider(),
		Flush:  func(context.Context, bool) error { return nil },
	})
	if err != nil {
		t.Fatalf("NewRouter() error = %v, want nil", err)
	}
	return router
}

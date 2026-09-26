package httpserver_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	metricnoop "go.opentelemetry.io/otel/metric/noop"
	tracenoop "go.opentelemetry.io/otel/trace/noop"
	"go.uber.org/zap"

	httpserver "github.com/MathTrail/mathtrail-standalone/internal/transport/http"
)

// The service answers under its own name only: a request under any other gets
// an empty 404 whatever it asked for, so that there is one issuer and one set
// of metadata. The platform's probe is the exception, since it asks by the
// platform's own address.
func TestOnlyTheServicesOwnHostIsServed(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		method     string
		host       string
		path       string
		wantStatus int
		wantReach  bool
	}{
		{name: "own host", method: http.MethodPost, host: "example.com", path: "/mcp", wantStatus: http.StatusOK, wantReach: true},
		{name: "own host written otherwise", method: http.MethodPost, host: "EXAMPLE.com:80", path: "/mcp", wantStatus: http.StatusOK, wantReach: true},
		{name: "another host", method: http.MethodPost, host: "service-abc.a.run.app", path: "/mcp", wantStatus: http.StatusNotFound},
		{name: "another host, unknown path", method: http.MethodGet, host: "service-abc.a.run.app", path: "/nope", wantStatus: http.StatusNotFound},
		{name: "another host, wrong method", method: http.MethodPost, host: "service-abc.a.run.app", path: "/health", wantStatus: http.StatusNotFound},
		{name: "another host, own port elsewhere", method: http.MethodPost, host: "example.com:8443", path: "/mcp", wantStatus: http.StatusNotFound},
		{name: "probe under another host", method: http.MethodGet, host: "service-abc.a.run.app", path: "/health", wantStatus: http.StatusOK},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			req := httptest.NewRequestWithContext(t.Context(), tc.method, tc.path, http.NoBody)
			req.Host = tc.host
			rec := send(t, req)

			if rec.Code != tc.wantStatus {
				t.Fatalf("status = %d, want %d", rec.Code, tc.wantStatus)
			}
			if reachedIt := rec.Body.String() == reached; reachedIt != tc.wantReach {
				t.Errorf("reached the endpoint = %t, want %t", reachedIt, tc.wantReach)
			}
			if tc.wantStatus == http.StatusNotFound && rec.Body.Len() != 0 {
				t.Errorf("body = %q, want an empty 404 for a host that is not ours", rec.Body.String())
			}
		})
	}
}

// A browser can be made to send a request on somebody's behalf from a page of
// any origin, and it says which in the Origin header. The endpoint refuses
// every origin but the service's own; a request with no Origin at all — every
// chat host's own client — is not a browser's, and passes.
func TestABrowserOnAnotherOriginIsRefused(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		origin     string
		wantStatus int
		wantReach  bool
	}{
		{name: "no origin", origin: "", wantStatus: http.StatusOK, wantReach: true},
		{name: "own origin", origin: "http://example.com", wantStatus: http.StatusOK, wantReach: true},
		{name: "own origin written otherwise", origin: "HTTP://Example.COM:80", wantStatus: http.StatusOK, wantReach: true},
		{name: "another origin", origin: "https://attacker.example", wantStatus: http.StatusForbidden},
		{name: "own host over another scheme", origin: "https://example.com", wantStatus: http.StatusForbidden},
		{name: "a sandboxed page", origin: "null", wantStatus: http.StatusForbidden},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/mcp", http.NoBody)
			if tc.origin != "" {
				req.Header.Set("Origin", tc.origin)
			}
			rec := send(t, req)

			if rec.Code != tc.wantStatus {
				t.Fatalf("status = %d, want %d", rec.Code, tc.wantStatus)
			}
			if reachedIt := rec.Body.String() == reached; reachedIt != tc.wantReach {
				t.Errorf("reached the endpoint = %t, want %t", reachedIt, tc.wantReach)
			}
			if tc.wantStatus == http.StatusForbidden && rec.Body.Len() != 0 {
				t.Errorf("body = %q, want an empty 403", rec.Body.String())
			}
		})
	}
}

// Every answer tells a browser to reach the service over HTTPS only, to
// believe the type it declares, and to send no referrer onward — refusals
// included, since they are answers too.
func TestEveryAnswerCarriesTheSecurityHeaders(t *testing.T) {
	t.Parallel()

	want := map[string]string{
		"Strict-Transport-Security": "max-age=31536000",
		"X-Content-Type-Options":    "nosniff",
		"Referrer-Policy":           "no-referrer",
	}
	cases := []struct {
		name   string
		method string
		host   string
		origin string
		path   string
	}{
		{name: "probe", method: http.MethodGet, path: "/health"},
		{name: "unknown path", method: http.MethodGet, path: "/nope"},
		{name: "wrong method", method: http.MethodPost, path: "/health"},
		{name: "another host", method: http.MethodPost, host: "service-abc.a.run.app", path: "/mcp"},
		{name: "another origin", method: http.MethodPost, origin: "https://attacker.example", path: "/mcp"},
		{name: "the endpoint", method: http.MethodPost, path: "/mcp"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			req := httptest.NewRequestWithContext(t.Context(), tc.method, tc.path, http.NoBody)
			if tc.host != "" {
				req.Host = tc.host
			}
			if tc.origin != "" {
				req.Header.Set("Origin", tc.origin)
			}
			rec := send(t, req)

			for header, value := range want {
				if got := rec.Header().Get(header); got != value {
					t.Errorf("%s = %q, want %q", header, got, value)
				}
			}
		})
	}
}

// The endpoint takes both methods the protocol names and decides itself what
// each means; any other method is refused before it, with the two named.
func TestTheEndpointTakesGetAndPostOnly(t *testing.T) {
	t.Parallel()

	for _, method := range []string{http.MethodGet, http.MethodPost} {
		if rec := call(t, method, "/mcp"); rec.Body.String() != reached {
			t.Errorf("%s /mcp did not reach the endpoint: status %d", method, rec.Code)
		}
	}

	rec := call(t, http.MethodDelete, "/mcp")
	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("DELETE /mcp status = %d, want %d", rec.Code, http.StatusMethodNotAllowed)
	}
	if got, want := rec.Header().Get("Allow"), "GET, POST"; got != want {
		t.Errorf("Allow = %q, want %q", got, want)
	}
}

// A router given no address to be, or no endpoint to serve, refuses to be
// built and says what is missing, rather than serving a route that fails on
// its first request.
func TestARouterWithSomethingMissingIsRefused(t *testing.T) {
	t.Parallel()

	withoutMCP := endpoints()
	withoutMCP.MCP = nil
	withoutHealth := endpoints()
	withoutHealth.Health = nil

	cases := []struct {
		name      string
		publicURL string
		endpoints httpserver.Endpoints
		want      string
	}{
		{name: "no host", publicURL: "not a url", endpoints: endpoints(), want: "public URL has no host"},
		{name: "not an address", publicURL: "http://[::1", endpoints: endpoints(), want: "public URL is not an address"},
		{name: "no MCP endpoint", publicURL: publicURL, endpoints: withoutMCP, want: "MCP"},
		{name: "no probe", publicURL: publicURL, endpoints: withoutHealth, want: "Health"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			_, err := httpserver.NewRouter(tc.publicURL, tc.endpoints, zap.NewNop(), httpserver.Observability{
				Traces: tracenoop.NewTracerProvider(),
				Meters: metricnoop.NewMeterProvider(),
				Flush:  noDelivery,
			})
			if !errors.Is(err, httpserver.ErrEndpoints) {
				t.Fatalf("NewRouter() error = %v, want it to refuse", err)
			}
			if !strings.Contains(err.Error(), tc.want) {
				t.Errorf("error = %q, want it to name %q", err, tc.want)
			}
		})
	}
}

// send serves one request with a router that keeps no telemetry.
func send(t *testing.T, req *http.Request) *httptest.ResponseRecorder {
	t.Helper()

	rec := httptest.NewRecorder()
	newRouter(t).ServeHTTP(rec, req)
	return rec
}

// A path with a slash added is not another name for the same endpoint. It is
// an unknown path like any other, answered by the router with every rule in
// place, rather than redirected by the framework before any rule runs.
func TestAPathWithASlashAddedIsNotAnotherName(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		method string
		host   string
		path   string
		want   int
	}{
		{method: http.MethodPost, host: "example.com", path: "/mcp/", want: http.StatusNotFound},
		{method: http.MethodGet, host: "example.com", path: "/health/", want: http.StatusNotFound},
		{method: http.MethodPost, host: "service-abc.a.run.app", path: "/mcp/", want: http.StatusNotFound},
	} {
		req := httptest.NewRequestWithContext(t.Context(), tc.method, tc.path, http.NoBody)
		req.Host = tc.host
		rec := send(t, req)

		if rec.Code != tc.want {
			t.Errorf("%s %s under %s: status = %d, want %d", tc.method, tc.path, tc.host, rec.Code, tc.want)
		}
		if got := rec.Header().Get("Location"); got != "" {
			t.Errorf("%s %s: redirected to %q, want no redirect", tc.method, tc.path, got)
		}
		if got := rec.Header().Get("Strict-Transport-Security"); got == "" {
			t.Errorf("%s %s: the answer carries no Strict-Transport-Security, want every rule in place", tc.method, tc.path)
		}
	}
}

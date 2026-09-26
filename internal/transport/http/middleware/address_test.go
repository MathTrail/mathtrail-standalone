package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/MathTrail/mathtrail-standalone/internal/transport/http/middleware"
)

// addressRouter serves one route and the probe behind the address checks, for
// the service at the given public URL.
func addressRouter(t *testing.T, publicURL string) *gin.Engine {
	t.Helper()

	public, err := url.Parse(publicURL)
	if err != nil {
		t.Fatalf("url.Parse(%q) error = %v", publicURL, err)
	}
	router := gin.New()
	router.Use(middleware.SecurityHeaders(), middleware.Host(public))
	router.POST("/mcp", middleware.Origin(public), func(c *gin.Context) { c.Status(http.StatusOK) })
	router.GET("/health", func(c *gin.Context) { c.Status(http.StatusOK) })
	return router
}

// A host is the service's own when it names the same host, in any case, with
// or without the port its scheme takes by default; any other name is refused,
// except for the probe.
func TestHostServesTheServicesOwnNameOnly(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name   string
		public string
		host   string
		path   string
		want   int
	}{
		{name: "the same host", public: "https://mcp.example", host: "mcp.example", path: "/mcp", want: http.StatusOK},
		{name: "in another case", public: "https://mcp.example", host: "MCP.Example", path: "/mcp", want: http.StatusOK},
		{name: "with the default port", public: "https://mcp.example", host: "mcp.example:443", path: "/mcp", want: http.StatusOK},
		{name: "with another port", public: "https://mcp.example", host: "mcp.example:8443", path: "/mcp", want: http.StatusNotFound},
		{name: "a port of its own", public: "http://localhost:8080", host: "localhost:8080", path: "/mcp", want: http.StatusOK},
		{name: "without its port", public: "http://localhost:8080", host: "localhost", path: "/mcp", want: http.StatusNotFound},
		{name: "an address", public: "http://[::1]:80", host: "[::1]", path: "/mcp", want: http.StatusOK},
		{name: "another name", public: "https://mcp.example", host: "service.a.run.app", path: "/mcp", want: http.StatusNotFound},
		{name: "the probe under another name", public: "https://mcp.example", host: "service.a.run.app", path: "/health", want: http.StatusOK},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			method := http.MethodPost
			if tc.path == "/health" {
				method = http.MethodGet
			}
			req := httptest.NewRequestWithContext(t.Context(), method, tc.path, http.NoBody)
			req.Host = tc.host
			rec := httptest.NewRecorder()
			addressRouter(t, tc.public).ServeHTTP(rec, req)

			if rec.Code != tc.want {
				t.Errorf("status = %d, want %d", rec.Code, tc.want)
			}
			if tc.want == http.StatusNotFound && rec.Body.Len() != 0 {
				t.Errorf("body = %q, want an empty refusal", rec.Body.String())
			}
		})
	}
}

// Only a browser sends an Origin, and only the service's own is let through.
func TestOriginLetsTheServicesOwnPagesAndNoBrowserThrough(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name   string
		origin string
		want   int
	}{
		{name: "no origin", origin: "", want: http.StatusOK},
		{name: "the service's own", origin: "https://mcp.example", want: http.StatusOK},
		{name: "the same, written otherwise", origin: "HTTPS://MCP.example:443", want: http.StatusOK},
		{name: "another scheme", origin: "http://mcp.example", want: http.StatusForbidden},
		{name: "another host", origin: "https://attacker.example", want: http.StatusForbidden},
		{name: "a sandboxed page", origin: "null", want: http.StatusForbidden},
		{name: "not an address at all", origin: "::::", want: http.StatusForbidden},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/mcp", http.NoBody)
			req.Host = "mcp.example"
			if tc.origin != "" {
				req.Header.Set("Origin", tc.origin)
			}
			rec := httptest.NewRecorder()
			addressRouter(t, "https://mcp.example").ServeHTTP(rec, req)

			if rec.Code != tc.want {
				t.Errorf("status = %d, want %d", rec.Code, tc.want)
			}
			if tc.want == http.StatusForbidden && rec.Body.Len() != 0 {
				t.Errorf("body = %q, want an empty refusal", rec.Body.String())
			}
		})
	}
}

// The headers every answer carries are on a refusal too, since they are set
// before anything that could refuse.
func TestTheSecurityHeadersAreOnARefusalToo(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/mcp", http.NoBody)
	req.Host = "service.a.run.app"
	rec := httptest.NewRecorder()
	addressRouter(t, "https://mcp.example").ServeHTTP(rec, req)

	for header, want := range map[string]string{
		"Strict-Transport-Security": "max-age=31536000",
		"X-Content-Type-Options":    "nosniff",
		"Referrer-Policy":           "no-referrer",
	} {
		if got := rec.Header().Get(header); got != want {
			t.Errorf("%s = %q, want %q", header, got, want)
		}
	}
}

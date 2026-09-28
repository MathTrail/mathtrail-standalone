// Package servicetest stands in for the MathTrail service in the tests of the
// load tool. Start runs the real one inside the test: the real container
// behind a listener of its own, with the development sign-in and its profiles
// in memory, as a developer's machine runs it. Fake runs a service of the
// protocol library's own, for a case that needs answers the real one does not
// give on demand. Either is reached the way any service is, through a target.
package servicetest

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"

	"github.com/MathTrail/mathtrail-standalone/internal/app"
	"github.com/MathTrail/mathtrail-standalone/internal/config"
	"github.com/MathTrail/mathtrail-standalone/tools/load/session"
)

// quiet puts the router's framework in its test mode once, before the first
// service of the process is built: it is one setting for the whole process.
var quiet sync.Once

// Start starts the service with the environment given on top of the
// development sign-in and a sealing key made for the test, and stops it when
// the test is over. What it returns is where the service is reached, and the
// host it knows itself by.
func Start(t testing.TB, env ...string) session.Target {
	t.Helper()
	quiet.Do(func() { gin.SetMode(gin.TestMode) })

	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		t.Fatalf("make a sealing key: %v", err)
	}
	environ := append([]string{
		"MATHTRAIL_DEV_AUTH=true",
		"MATHTRAIL_SEAL_KEY_CURRENT=" + base64.StdEncoding.EncodeToString(key),
	}, env...)
	cfg, err := config.LoadFrom(environ)
	if err != nil {
		t.Fatalf("config.LoadFrom(%q) error = %v, want nil", env, err)
	}
	public, err := url.Parse(cfg.PublicURL)
	if err != nil {
		t.Fatalf("the public URL %q does not parse: %v", cfg.PublicURL, err)
	}

	container, err := app.NewContainer(context.Background(), cfg, zaptest.NewLogger(t, zaptest.Level(zap.WarnLevel)))
	if err != nil {
		t.Fatalf("app.NewContainer() error = %v, want nil", err)
	}
	server := httptest.NewServer(container.Router)
	t.Cleanup(func() {
		server.Close()
		container.Close(context.Background())
	})
	return session.Target{URL: server.URL, Host: public.Host}
}

// Fake runs a service of the protocol library's own with the tools given, each
// taking whatever it is handed, behind whatever the case puts in front of it,
// and without sessions as the real one is.
func Fake(t testing.TB, tools map[string]mcp.ToolHandler, front func(http.Handler) http.Handler) session.Target {
	t.Helper()

	server := mcp.NewServer(&mcp.Implementation{Name: "fake", Version: "1"}, nil)
	for name, handler := range tools {
		server.AddTool(&mcp.Tool{Name: name, InputSchema: json.RawMessage(`{"type":"object"}`)}, handler)
	}
	var endpoint http.Handler = mcp.NewStreamableHTTPHandler(
		func(*http.Request) *mcp.Server { return server }, &mcp.StreamableHTTPOptions{Stateless: true})
	if front != nil {
		endpoint = front(endpoint)
	}
	mux := http.NewServeMux()
	mux.Handle("/mcp", endpoint)
	listener := httptest.NewServer(mux)
	t.Cleanup(listener.Close)
	return session.Target{URL: listener.URL}
}

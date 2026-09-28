package oauthserver_test

import (
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"

	"github.com/MathTrail/mathtrail-standalone/internal/infra/cimd"
	"github.com/MathTrail/mathtrail-standalone/internal/infra/seal"
	oauthserver "github.com/MathTrail/mathtrail-standalone/internal/transport/oauth"
)

// issuer is the address the server under test is.
const issuer = "https://mcp.example"

// site is the site the pages under test link to.
const site = "https://site.example"

// someDay is when these cases happen.
var someDay = time.Date(2026, 9, 27, 12, 0, 0, 0, time.UTC)

// keyRing is a ring with one key, the same in every case.
func keyRing(t *testing.T) *seal.KeyRing {
	t.Helper()

	ring, err := seal.NewKeyRing(base64.StdEncoding.EncodeToString(bytes.Repeat([]byte("k"), seal.KeySize)), "")
	if err != nil {
		t.Fatalf("NewKeyRing() error = %v, want nil", err)
	}
	return ring
}

// noDocuments fetches nothing: the cases here never ask for a client's
// document, and one that did would be told it is unreachable.
type noDocuments struct{}

func (noDocuments) Fetch(context.Context, string) (cimd.Fetched, error) {
	return cimd.Fetched{}, cimd.ErrUnreachable
}

// settings are what a server under test is built from, with its lines kept.
func settings(t *testing.T) (*oauthserver.Settings, *observer.ObservedLogs) {
	t.Helper()

	core, logs := observer.New(zapcore.InfoLevel)
	return &oauthserver.Settings{
		PublicURL: issuer,
		Scope:     "mcp",
		Seal:      keyRing(t),
		Documents: noDocuments{},
		Logger:    zap.New(core),
		SiteURL:   site,
		Now:       func() time.Time { return someDay },
	}, logs
}

// serve builds a server and mounts its handlers at their paths, as the router
// does.
func serve(t *testing.T) (*httptest.Server, *observer.ObservedLogs) {
	t.Helper()

	given, logs := settings(t)
	server, err := oauthserver.New(given)
	if err != nil {
		t.Fatalf("New() error = %v, want nil", err)
	}
	routes := http.NewServeMux()
	routes.Handle("GET /.well-known/oauth-protected-resource/mcp", server.ResourceMetadata)
	routes.Handle("GET /.well-known/oauth-authorization-server", server.ServerMetadata)
	routes.Handle("POST /oauth/register", server.Register)
	served := httptest.NewServer(routes)
	t.Cleanup(served.Close)
	return served, logs
}

// A refusal of the resource names where a sign-in begins: the resource's own
// metadata, under the issuer.
func TestTheResourcesRefusalLeadsToItsMetadata(t *testing.T) {
	t.Parallel()

	given, _ := settings(t)
	server, err := oauthserver.New(given)
	if err != nil {
		t.Fatalf("New() error = %v, want nil", err)
	}
	if want := issuer + "/.well-known/oauth-protected-resource/mcp"; server.ResourceMetadataURL != want {
		t.Errorf("ResourceMetadataURL = %q, want %q", server.ResourceMetadataURL, want)
	}
}

// A server given a part it cannot work without, or an issuer that is not a
// scheme and a host alone, refuses to be built and says what is wrong.
func TestAServerWithSomethingMissingIsRefused(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name  string
		spoil func(*oauthserver.Settings)
		want  string
	}{
		{"no scope", func(s *oauthserver.Settings) { s.Scope = "" }, "Scope"},
		{"no key ring", func(s *oauthserver.Settings) { s.Seal = nil }, "Seal"},
		{"no documents", func(s *oauthserver.Settings) { s.Documents = nil }, "Documents"},
		{"no logger", func(s *oauthserver.Settings) { s.Logger = nil }, "Logger"},
		{"no clock", func(s *oauthserver.Settings) { s.Now = nil }, "Now"},
		{"no issuer", func(s *oauthserver.Settings) { s.PublicURL = "" }, "PublicURL"},
		{"an issuer with a path", func(s *oauthserver.Settings) { s.PublicURL = issuer + "/tenant" }, "PublicURL"},
		{"an issuer with a trailing slash", func(s *oauthserver.Settings) { s.PublicURL = issuer + "/" }, "PublicURL"},
		{"an issuer with a query", func(s *oauthserver.Settings) { s.PublicURL = issuer + "?a=b" }, "PublicURL"},
		{"an issuer with an empty query", func(s *oauthserver.Settings) { s.PublicURL = issuer + "?" }, "PublicURL"},
		{"a site with an empty fragment", func(s *oauthserver.Settings) { s.SiteURL = site + "#" }, "SiteURL"},
		{"no site", func(s *oauthserver.Settings) { s.SiteURL = "" }, "SiteURL"},
		{"a site with a path", func(s *oauthserver.Settings) { s.SiteURL = site + "/en/" }, "SiteURL"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			given, _ := settings(t)
			tc.spoil(given)
			_, err := oauthserver.New(given)
			if !errors.Is(err, oauthserver.ErrSettings) || !strings.Contains(err.Error(), tc.want) {
				t.Errorf("New() error = %v, want ErrSettings naming %s", err, tc.want)
			}
		})
	}
}

package oauthserver

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"slices"
	"strings"
	"testing"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"

	"github.com/MathTrail/mathtrail-standalone/internal/infra/cimd"
	"github.com/MathTrail/mathtrail-standalone/internal/infra/seal"
)

// testIssuer is the address the server under test is.
const testIssuer = "https://mcp.example"

// testSite is the site the pages under test link to.
const testSite = "https://site.example"

// testDay is when these cases happen.
var testDay = time.Date(2026, 9, 27, 12, 0, 0, 0, time.UTC)

// ringOf is a ring with one key made of the byte given.
func ringOf(t testing.TB, fill byte) *seal.KeyRing {
	t.Helper()

	ring, err := seal.NewKeyRing(base64.StdEncoding.EncodeToString(bytes.Repeat([]byte{fill}, seal.KeySize)), "")
	if err != nil {
		t.Fatalf("NewKeyRing() error = %v, want nil", err)
	}
	return ring
}

// documents answers every fetch with what a case sets, and remembers what it
// was asked for.
type documents struct {
	fetched cimd.Fetched
	err     error
	asked   []string
}

func (d *documents) Fetch(_ context.Context, clientID string) (cimd.Fetched, error) {
	d.asked = append(d.asked, clientID)
	return d.fetched, d.err
}

// knownClients are the clients of a server under test, over the key ring and
// the documents given, with their lines kept.
func knownClients(t testing.TB, ring *seal.KeyRing, fetcher cimd.Fetcher) (*clients, *observer.ObservedLogs) {
	t.Helper()

	core, logs := observer.New(zapcore.InfoLevel)
	server, err := New(&Settings{
		PublicURL: testIssuer,
		Scope:     "mcp",
		Seal:      ring,
		Documents: fetcher,
		Logger:    zap.New(core),
		SiteURL:   testSite,
		Now:       func() time.Time { return testDay },
	})
	if err != nil {
		t.Fatalf("New() error = %v, want nil", err)
	}
	return server.clients, logs
}

// A client registered here is its own identifier: the identifier opens into
// the registration, and nothing had to be kept for it.
func TestARegisteredClientIsItsIdentifier(t *testing.T) {
	t.Parallel()

	known, _ := knownClients(t, ringOf(t, 'k'), &documents{})
	clientID, err := known.register(registration{
		RedirectURIs: []string{"https://chatgpt.com/connector/oauth/abc", "http://127.0.0.1:33418/"},
		ClientName:   "ChatGPT",
		IssuedAt:     testDay.Unix(),
	})
	if err != nil {
		t.Fatalf("register() error = %v, want nil", err)
	}

	client, err := known.resolve(t.Context(), clientID)
	if err != nil {
		t.Fatalf("resolve() error = %v, want the registered client", err)
	}
	if client.ID != clientID || client.Name != "ChatGPT" || client.Registration != registrationDCR ||
		!slices.Equal(client.RedirectURIs, []string{"https://chatgpt.com/connector/oauth/abc", "http://127.0.0.1:33418/"}) {
		t.Errorf("resolve() = %+v, want the registration as it was made", client)
	}
}

// Nothing but an identifier this server issued, for itself, opens as a
// registered client: not a string nobody issued, not one changed since, not a
// value sealed for another purpose or another issuer, not one sealed under a
// key this server does not have.
func TestAForgedIdentifierIsNoClient(t *testing.T) {
	t.Parallel()

	ring := ringOf(t, 'k')
	known, _ := knownClients(t, ring, &documents{})
	record, err := json.Marshal(registration{RedirectURIs: []string{"https://a.example/cb"}, IssuedAt: testDay.Unix()})
	if err != nil {
		t.Fatalf("encoding a registration: %v", err)
	}
	issued, err := known.register(registration{RedirectURIs: []string{"https://a.example/cb"}, IssuedAt: testDay.Unix()})
	if err != nil {
		t.Fatalf("register() error = %v, want nil", err)
	}
	sealed := func(ring *seal.KeyRing, purpose seal.Purpose, plaintext []byte, issuer string) string {
		value, err := ring.Seal(purpose, plaintext, issuer)
		if err != nil {
			t.Fatalf("Seal() error = %v, want nil", err)
		}
		return value
	}

	for _, tc := range []struct {
		name     string
		clientID string
	}{
		{"nothing", ""},
		{"a name", "claude"},
		{"the shape of one", "mt1.d." + ring.CurrentKeyID() + ".AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"},
		{"one changed since", issued[:len(issued)-2] + flipped(issued[len(issued)-2:])},
		{"an access token", sealed(ring, seal.PurposeAccess, record, testIssuer)},
		{"a sealed answer", sealed(ring, seal.PurposeTaskAnswer, record, testIssuer)},
		{"another issuer's", sealed(ring, seal.PurposeClient, record, "https://other.example")},
		{"no issuer's", sealed(ring, seal.PurposeClient, record, "")},
		{"another key's", sealed(ringOf(t, 'o'), seal.PurposeClient, record, testIssuer)},
		{"a registration of nothing", sealed(ring, seal.PurposeClient, []byte(`{"redirect_uris":[]}`), testIssuer)},
		{"not a registration", sealed(ring, seal.PurposeClient, []byte(`[]`), testIssuer)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if client, err := known.resolve(t.Context(), tc.clientID); !errors.Is(err, errUnknownClient) {
				t.Errorf("resolve() = %+v, %v; want errUnknownClient", client, err)
			}
		})
	}
}

// flipped is two characters of base64url replaced by two others.
func flipped(pair string) string {
	if pair == "AA" {
		return "BB"
	}
	return "AA"
}

// A client that names itself by its document is the client the document
// describes, less the addresses a sign-in may never send the parent to.
func TestAClientIsWhatItsDocumentSays(t *testing.T) {
	t.Parallel()

	const address = "https://claude.ai/oauth/mcp-oauth-client-metadata"
	fetcher := &documents{fetched: cimd.Fetched{Document: cimd.Document{
		ClientID:   address,
		ClientName: "Claude",
		RedirectURIs: []string{
			"https://claude.ai/api/mcp/auth_callback",
			"http://claude.ai/api/mcp/auth_callback",
			"http://localhost:3000/callback",
			"com.claude.app:/oauth",
		},
	}}}
	known, logs := knownClients(t, ringOf(t, 'k'), fetcher)

	client, err := known.resolve(t.Context(), address)
	if err != nil {
		t.Fatalf("resolve() error = %v, want Claude", err)
	}
	if client.ID != address || client.Name != "Claude" || client.Registration != registrationCIMD ||
		!slices.Equal(client.RedirectURIs, []string{"https://claude.ai/api/mcp/auth_callback", "http://localhost:3000/callback"}) {
		t.Errorf("resolve() = %+v, want Claude with its two addresses a parent may be sent to", client)
	}
	if !slices.Equal(fetcher.asked, []string{address}) {
		t.Errorf("the fetcher was asked for %v, want the client's address once", fetcher.asked)
	}

	lines := logs.FilterMessage("cimd_fetch").All()
	if len(lines) != 1 {
		t.Fatalf("%d cimd_fetch lines, want 1", len(lines))
	}
	fields := lines[0].ContextMap()
	if fields["host"] != "claude.ai" || fields["cached"] != false || fields["outcome"] != "ok" || fields["duration_ms"] != int64(0) {
		t.Errorf("the line is %v, want a fetch from claude.ai that went well", fields)
	}
}

// A document that leaves no address a parent may be sent back to, or names
// its client past what can be shown, is no client.
func TestADocumentThatCannotBeUsedIsNoClient(t *testing.T) {
	t.Parallel()

	const address = "https://client.example.com/meta"
	for _, tc := range []struct {
		name     string
		document cimd.Document
	}{
		{"only addresses a parent may not be sent to", cimd.Document{
			ClientID: address, ClientName: "A", RedirectURIs: []string{"http://a.example/cb", "com.example:/cb"},
		}},
		{"a name too long to show", cimd.Document{
			ClientID: address, ClientName: strings.Repeat("я", maxClientName+1), RedirectURIs: []string{"https://a.example/cb"},
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			known, _ := knownClients(t, ringOf(t, 'k'), &documents{fetched: cimd.Fetched{Document: tc.document}})
			if client, err := known.resolve(t.Context(), address); !errors.Is(err, errUnknownClient) {
				t.Errorf("resolve() = %+v, %v; want errUnknownClient", client, err)
			}
		})
	}
}

// Every way a fetch of a client's document can end is no client, and leaves
// a line in the words of its own closed list — with the host it was asked of,
// except when the address was not a client's at all.
func TestAFailedFetchIsNoClientAndSaysHowItEnded(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		err     error
		outcome string
		host    string
	}{
		{cimd.ErrClientURL, "refused_url", ""},
		{cimd.ErrAddress, "refused_address", "client.example.com"},
		{cimd.ErrRedirect, "redirect", "client.example.com"},
		{cimd.ErrStatus, "status", "client.example.com"},
		{cimd.ErrTooLarge, "too_large", "client.example.com"},
		{cimd.ErrUnreachable, "unreachable", "client.example.com"},
		{cimd.ErrDocument, "invalid_document", "client.example.com"},
		{context.Canceled, "unreachable", "client.example.com"},
	} {
		t.Run(tc.outcome, func(t *testing.T) {
			t.Parallel()

			known, logs := knownClients(t, ringOf(t, 'k'), &documents{err: tc.err})
			if client, err := known.resolve(t.Context(), "https://client.example.com/meta"); !errors.Is(err, errUnknownClient) || !errors.Is(err, tc.err) {
				t.Errorf("resolve() = %+v, %v; want errUnknownClient for %v", client, err, tc.err)
			}
			lines := logs.FilterMessage("cimd_fetch").All()
			if len(lines) != 1 || lines[0].ContextMap()["outcome"] != tc.outcome || lines[0].ContextMap()["host"] != tc.host {
				t.Errorf("the lines are %v, want one with outcome %s and host %q", lines, tc.outcome, tc.host)
			}
		})
	}
}

// A document a kept copy answered for says so in its line.
func TestAKeptDocumentSaysSoInItsLine(t *testing.T) {
	t.Parallel()

	known, logs := knownClients(t, ringOf(t, 'k'), &documents{fetched: cimd.Fetched{
		Document: cimd.Document{ClientID: "https://a.example/meta", ClientName: "A", RedirectURIs: []string{"https://a.example/cb"}},
		Cached:   true,
	}})
	if _, err := known.resolve(t.Context(), "https://a.example/meta"); err != nil {
		t.Fatalf("resolve() error = %v, want the client", err)
	}
	if lines := logs.FilterMessage("cimd_fetch").All(); len(lines) != 1 || lines[0].ContextMap()["cached"] != true {
		t.Errorf("the lines are %v, want one that says a kept copy answered", lines)
	}
}

// A sign-in may send the parent back only to an address the client registered
// exactly, byte for byte: nothing that merely looks the same, starts the same
// or means the same.
func TestARedirectMatchesOnlyExactly(t *testing.T) {
	t.Parallel()

	client := &Client{RedirectURIs: []string{"https://claude.ai/api/mcp/auth_callback", "http://localhost:3000/callback"}}
	for _, tc := range []struct {
		uri       string
		redirects bool
	}{
		{"https://claude.ai/api/mcp/auth_callback", true},
		{"http://localhost:3000/callback", true},

		{"https://claude.ai/api/mcp/auth_callback/", false},
		{"https://Claude.ai/api/mcp/auth_callback", false},
		{"HTTPS://claude.ai/api/mcp/auth_callback", false},
		{"https://claude.ai:443/api/mcp/auth_callback", false},
		{"https://claude.ai/api/mcp/auth_callback?next=1", false},
		{"https://claude.ai/api/mcp/auth_call", false},
		{"https://claude.ai/api/mcp/auth%5Fcallback", false},
		{"https://claude.ai/api/mcp/auth_callback.evil.example", false},
		{"http://localhost:3001/callback", false},
		{"http://127.0.0.1:3000/callback", false},
		{"", false},
	} {
		if got := client.Redirects(tc.uri); got != tc.redirects {
			t.Errorf("Redirects(%q) = %v, want %v", tc.uri, got, tc.redirects)
		}
	}
}

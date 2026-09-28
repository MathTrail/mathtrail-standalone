package mcpserver_test

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/MathTrail/mathtrail-standalone/internal/store"
	mcpserver "github.com/MathTrail/mathtrail-standalone/internal/transport/mcp"
)

// The token the reader of these cases vouches for, and whom it signs in.
const (
	vouchedToken = "a-token-the-server-issued"
	vouchedUser  = "a-user-of-sixteen"
	// whyRefused is why the reader refuses any other token: the reader's own
	// words, which the client must never hear.
	whyRefused = "the reader's own reason"
	metadata   = "https://mcp.example/.well-known/oauth-protected-resource/mcp"
)

// readerEndingIn is a reader that vouches for one token, whose account ends at
// the time given from now, and refuses every other.
func readerEndingIn(left time.Duration) mcpserver.TokenReader {
	return func(_ context.Context, token string) (store.Account, time.Time, error) {
		if token != vouchedToken {
			return store.Account{}, time.Time{}, errors.New(whyRefused)
		}
		return store.NewAccount(vouchedUser, "a-google-access-token", time.Time{}), time.Now().Add(left), nil
	}
}

// bearing is a client that sends every request with the token given.
type bearing struct{ token string }

func (b bearing) RoundTrip(r *http.Request) (*http.Response, error) {
	r = r.Clone(r.Context())
	r.Header.Set("Authorization", "Bearer "+b.token)
	return http.DefaultTransport.RoundTrip(r)
}

// connectWith is a client of the protocol that signs its every request with a
// token.
func (h *harness) connectWith(t *testing.T, token string) (*mcp.ClientSession, error) {
	t.Helper()
	return h.connectAs(t, "claude-code", token)
}

// connectAs is connectWith, for a client that gives itself the name given.
func (h *harness) connectAs(t *testing.T, name, token string) (*mcp.ClientSession, error) {
	t.Helper()

	client := mcp.NewClient(&mcp.Implementation{Name: name, Version: "1.0.0"}, nil)
	session, err := client.Connect(t.Context(), &mcp.StreamableClientTransport{
		Endpoint:             h.server.URL + "/mcp",
		HTTPClient:           &http.Client{Transport: bearing{token: token}},
		DisableStandaloneSSE: true,
		MaxRetries:           -1,
	}, nil)
	if err == nil {
		t.Cleanup(func() { _ = session.Close() })
	}
	return session, err
}

// A request with no token, or with one the reader does not vouch for, is not
// let in. The refusal is the one a client expects of a service that wants a
// token — 401, and a challenge naming where the resource's metadata is and the
// scope — and says nothing of why the reader refused. No tool runs.
func TestARequestWithoutAVouchedTokenIsRefused(t *testing.T) {
	t.Parallel()

	h := serve(t, mcpserver.BearerSignIn(readerEndingIn(time.Hour), metadata))
	for _, credential := range []string{"", "Bearer a-token-nobody-issued", "Basic " + vouchedToken} {
		header := http.Header{}
		if credential != "" {
			header.Set("Authorization", credential)
		}
		resp := h.post(t, legacyCall("echo", `{"say":"hello"}`), header)

		if resp.status != http.StatusUnauthorized {
			t.Errorf("with credential %q: status = %d, want %d", credential, resp.status, http.StatusUnauthorized)
		}
		if got, want := resp.header.Get("WWW-Authenticate"), `Bearer resource_metadata="`+metadata+`", scope="mcp"`; got != want {
			t.Errorf("with credential %q: WWW-Authenticate = %q, want %q", credential, got, want)
		}
		if got, want := resp.header.Get("Cache-Control"), "no-store, no-transform"; got != want {
			t.Errorf("with credential %q: Cache-Control = %q, want %q", credential, got, want)
		}
		if strings.Contains(string(resp.body), whyRefused) {
			t.Errorf("with credential %q: the answer %q tells why the token was refused", credential, resp.body)
		}
	}
	if _, err := h.connectWith(t, "a-token-nobody-issued"); err == nil {
		t.Error("Connect() error = nil, want the client refused")
	}

	h.settle()
	if lines := h.toolLines(); len(lines) != 0 {
		t.Errorf("tool_call lines = %d, want none: no tool may run for nobody", len(lines))
	}
}

// A request with a token the reader vouches for is let in as the account the
// token signs in: the tool acts for it, and its line names it.
func TestAVouchedTokenSignsTheRequestInAsItsAccount(t *testing.T) {
	t.Parallel()

	h := serve(t, mcpserver.BearerSignIn(readerEndingIn(time.Hour), metadata))
	session, err := h.connectWith(t, vouchedToken)
	if err != nil {
		t.Fatalf("Connect() error = %v, want the client let in", err)
	}
	result := call(t, session, "echo", map[string]any{"say": "hello"})
	if payload, _ := result.StructuredContent.(map[string]any); payload["for"] != vouchedUser {
		t.Errorf("the tool acted for %v, want %q", payload["for"], vouchedUser)
	}

	h.settle()
	if got := field(t, h.lineOf(t, "echo"), "user"); got != vouchedUser {
		t.Errorf("user = %q, want %q", got, vouchedUser)
	}
}

// A token is judged to end as the reader judges it: the minute two clocks may
// disagree by is allowed past the end the reader gave, and no more.
func TestATokenEndsWhenTheReaderSaysWithAMinuteOfSkew(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name   string
		left   time.Duration
		status int
	}{
		{"half a minute past its end", -30 * time.Second, http.StatusOK},
		{"two minutes past its end", -2 * time.Minute, http.StatusUnauthorized},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			h := serve(t, mcpserver.BearerSignIn(readerEndingIn(tc.left), metadata))
			resp := h.post(t, legacyCall("echo", `{"say":"hello"}`), http.Header{"Authorization": {"Bearer " + vouchedToken}})
			if resp.status != tc.status {
				t.Errorf("status = %d, want %d", resp.status, tc.status)
			}
		})
	}
}

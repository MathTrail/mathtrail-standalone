package oauthserver

import (
	"context"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/auth"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/modelcontextprotocol/go-sdk/oauthex"
	tracenoop "go.opentelemetry.io/otel/trace/noop"
	"go.uber.org/zap"

	"github.com/MathTrail/mathtrail-standalone/internal/infra/cimd"
	"github.com/MathTrail/mathtrail-standalone/internal/infra/googleauth/googletest"
	"github.com/MathTrail/mathtrail-standalone/internal/ratelimit/ratelimittest"
	"github.com/MathTrail/mathtrail-standalone/internal/store"
	mcpserver "github.com/MathTrail/mathtrail-standalone/internal/transport/mcp"
)

// hostDocument is where the host of these cases publishes the document it
// names itself by.
const hostDocument = "https://host.example/oauth/client.json"

// nothing is what the tool of these cases takes.
type nothing struct{}

// whom is what the tool of these cases says: whom the call acts for.
type whom struct {
	User string `json:"user"`
}

// serveResource serves the MCP endpoint of the case behind the check of the
// access tokens this server issues, with one tool, which says whom a call
// acts for.
func (h *signIn) serveResource(t *testing.T) {
	t.Helper()

	endpoint, err := mcpserver.NewHandler(&mcpserver.Settings{
		Instructions:        "Says whom a call acts for.",
		InstructionsVersion: "test",
		Version:             "test",
		SignIn:              mcpserver.BearerSignIn(h.server.Account, h.server.ResourceMetadataURL),
		Limits:              mcpserver.Limits{PerAccount: ratelimittest.Roomy(t), Instance: ratelimittest.Roomy(t)},
		Traces:              tracenoop.NewTracerProvider(),
		Logger:              zap.NewNop(),
		Widget:              "<!doctype html><title>the widget</title>",
		Origin:              "https://mcp.example",
		Site:                "https://site.example",
	}, mcpserver.Define(mcpserver.Spec{Name: "whoami", Title: "Who am I", Description: "Says whom a call acts for.", Effect: mcpserver.Reads},
		func(_ context.Context, account store.Account, _ nothing) (mcpserver.Reply[whom], error) {
			return mcpserver.Reply[whom]{Text: account.ID, Payload: whom{User: account.ID}}, nil
		}))
	if err != nil {
		t.Fatalf("mcpserver.NewHandler() error = %v, want nil", err)
	}
	h.routes.Handle("/mcp", endpoint)
}

// parentsOf is the parent at the host of a case: each time the host sends
// them to sign in, they go through the consent screen and Google in a browser
// of their own, allow everything asked, and the host is handed what the
// browser brought back. It counts how many times they were sent.
func (h *signIn) parentsOf(t *testing.T, signIns *int) auth.AuthorizationCodeFetcher {
	return func(_ context.Context, args *auth.AuthorizationArgs) (*auth.AuthorizationResult, error) {
		*signIns++
		parent := h.browser(t)
		back := answerAt(t, parent.get(h.google.Allow(h.toGoogleAt(parent, args.URL))))
		return &auth.AuthorizationResult{Code: back.Get("code"), State: back.Get("state"), Iss: back.Get("iss")}, nil
	}
}

// whoAmI calls the tool of the case, and is whom it said the call acted for.
func whoAmI(t *testing.T, session *mcp.ClientSession) string {
	t.Helper()

	result, err := session.CallTool(t.Context(), &mcp.CallToolParams{Name: "whoami", Arguments: map[string]any{}})
	if err != nil {
		t.Fatalf("CallTool(whoami) error = %v, want an answer", err)
	}
	if result.IsError || len(result.Content) != 1 {
		t.Fatalf("CallTool(whoami) = %+v, want one text", result)
	}
	text, _ := result.Content[0].(*mcp.TextContent)
	if text == nil {
		t.Fatalf("CallTool(whoami) = %+v, want a text", result.Content[0])
	}
	return text.Text
}

// A host signs a parent in with the protocol library's own client, as a chat
// host does. Refused at the resource, it finds the resource's metadata from
// the refusal and the server's from that, names itself — by its own document,
// or by registering — sends the parent through the consent screen and Google,
// trades the code for its tokens and calls a tool as the parent's account. It
// renews its tokens as they end. When the grant ends — the host disconnects,
// or the parent takes the access back in their Google account — the next
// renewal is refused, the host is refused at the resource, and it sends the
// parent to sign in again on its own.
func TestAHostSignsInRenewsAndSignsInAgain(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name string
		// registration is how the host names itself, and how the grant ends.
		registration string
	}{
		{"a host that names itself by its document and disconnects", registrationCIMD},
		{"a host that registers, whose parent takes the access back", registrationDCR},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			h := newSignInNow(t)
			h.serveResource(t)
			// Google's access tokens end ten seconds past the margin an access
			// token keeps before them, so that every token the host holds ends
			// ten seconds after it is issued and each call renews it.
			h.google.Misbehave(&googletest.Answer{Lifetime: int(googleMargin/time.Second) + 10})
			signIns := 0
			session, tokens := h.connectHost(t, tc.registration, &signIns)

			user := h.ring.UserID("google-drive:" + googletest.PermissionID)
			if got := whoAmI(t, session); got != user || signIns != 1 {
				t.Fatalf("after one sign-in the tool acted for %q, having signed in %d times; want %q once", got, signIns, user)
			}
			if h.google.Renewals() == 0 {
				t.Errorf("Google renewed nothing, want the host's tokens renewed as they end")
			}

			h.endGrant(t, tc.registration, tokens)
			if got := whoAmI(t, session); got != user || signIns != 2 {
				t.Errorf("once the grant ended the tool acted for %q, having signed in %d times; want %q, signed in again", got, signIns, user)
			}
			refreshes := h.lines(eventAuthRefresh)
			if !anyLine(refreshes, "reason", "grant_ended") {
				t.Errorf("auth_refresh lines = %v, want one refused as the grant ended", refreshes)
			}
		})
	}
}

// disconnect is the host disconnecting: it ends its grant with the refresh
// token it holds, naming itself by its document.
func (h *signIn) disconnect(t *testing.T, tokens *auth.AuthorizationCodeHandler) {
	t.Helper()

	source, err := tokens.TokenSource(t.Context())
	if err != nil || source == nil {
		t.Fatalf("TokenSource() = %v, %v; want the host's tokens", source, err)
	}
	held, err := source.Token()
	if err != nil {
		t.Fatalf("Token() error = %v, want the host's tokens", err)
	}
	answer := h.revoke(t, held.RefreshToken, hostDocument, url.Values{"token_type_hint": {"refresh_token"}})
	if answer.status != http.StatusOK {
		t.Fatalf("POST /oauth/revoke = %d %v, want 200", answer.status, answer.fields)
	}
}

// anyLine reports whether any of the lines has the value in the field.
func anyLine(lines []map[string]any, field, value string) bool {
	for _, line := range lines {
		if line[field] == value {
			return true
		}
	}
	return false
}

// connectHost is a host connecting to the resource of the case with the
// protocol library's own client, named the way given — by its own document,
// or by registering — and signing the parent in when the resource refuses it.
func (h *signIn) connectHost(t *testing.T, registration string, signIns *int) (*mcp.ClientSession, *auth.AuthorizationCodeHandler) {
	t.Helper()

	config := &auth.AuthorizationCodeHandlerConfig{
		RedirectURL:              hostRedirect,
		AuthorizationCodeFetcher: h.parentsOf(t, signIns),
		Client:                   h.served.Client(),
	}
	if registration == registrationCIMD {
		h.documents.fetched = cimd.Fetched{Document: cimd.Document{
			ClientID: hostDocument, ClientName: hostName, RedirectURIs: []string{hostRedirect},
		}}
		h.documents.err = nil
		config.ClientIDMetadataDocumentConfig = &auth.ClientIDMetadataDocumentConfig{URL: hostDocument}
	} else {
		config.DynamicClientRegistrationConfig = &auth.DynamicClientRegistrationConfig{
			Metadata: &oauthex.ClientRegistrationMetadata{RedirectURIs: []string{hostRedirect}, ClientName: hostName},
		}
	}
	tokens, err := auth.NewAuthorizationCodeHandler(config)
	if err != nil {
		t.Fatalf("NewAuthorizationCodeHandler() error = %v, want nil", err)
	}
	client := mcp.NewClient(&mcp.Implementation{Name: "a-chat-host", Version: "1.0.0"}, nil)
	session, err := client.Connect(t.Context(), &mcp.StreamableClientTransport{
		Endpoint:             h.served.URL + "/mcp",
		HTTPClient:           h.served.Client(),
		OAuthHandler:         tokens,
		DisableStandaloneSSE: true,
		MaxRetries:           -1,
	}, nil)
	if err != nil {
		t.Fatalf("Connect() error = %v, want the host signed in", err)
	}
	t.Cleanup(func() { _ = session.Close() })
	return session, tokens
}

// endGrant ends the grant of a host: a host that names itself by its document
// disconnects, and the parent of one that registered takes the access back in
// their Google account.
func (h *signIn) endGrant(t *testing.T, registration string, tokens *auth.AuthorizationCodeHandler) {
	t.Helper()

	if registration == registrationCIMD {
		h.disconnect(t, tokens)
		return
	}
	h.revokeAtGoogle(t, googletest.RefreshToken)
}

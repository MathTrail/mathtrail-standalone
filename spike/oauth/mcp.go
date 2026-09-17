// Throwaway T04 spike: the protected MCP server behind /mcp. Two tools only —
// this spike is about the OAuth dance, not protocol/widget mechanics (T03
// already covers those). Deleted in T17.
package main

import (
	"context"
	"fmt"
	"sync/atomic"

	"github.com/modelcontextprotocol/go-sdk/auth"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// currentIssuer caches the last request's "https://" + Host, set by the /mcp
// wrapper in main.go on every request. Tool handlers only get a
// *mcp.CallToolRequest, not the originating *http.Request, so this is the
// simplest way to give whoamiAdvancedHandler an issuer to build a
// resource_metadata URL from — fine for a single-tunnel spike where the host
// never changes mid-run.
var currentIssuer atomic.Value // string

func issuerForTools() string {
	if v, ok := currentIssuer.Load().(string); ok {
		return v
	}
	return ""
}

type whoamiArgs struct{}

type whoamiResult struct {
	ClientID  string   `json:"clientId"`
	Scopes    []string `json:"scopes"`
	ExpiresAt string   `json:"expiresAt"`
}

func whoamiHandler(ctx context.Context, _ *mcp.CallToolRequest, _ whoamiArgs) (*mcp.CallToolResult, whoamiResult, error) {
	info := auth.TokenInfoFromContext(ctx)
	if info == nil {
		// Shouldn't happen: RequireBearerToken already rejected unauthenticated
		// requests before the MCP layer ever sees them. Defensive only.
		return &mcp.CallToolResult{IsError: true, Content: []mcp.Content{&mcp.TextContent{Text: "no token info in context"}}}, whoamiResult{}, nil
	}
	return nil, whoamiResult{ClientID: info.UserID, Scopes: info.Scopes, ExpiresAt: info.Expiration.Format("15:04:05")}, nil
}

// requiredScopeAdvanced is the scope this spike's /authorize never grants
// (oauth.go always issues plain "mcp"), so any call needing it demonstrates
// the insufficient-scope / step-up path deterministically. Two tools exercise
// it two different, incompatible ways — see main.go's requireToolScope for
// why "whoami_advanced" never actually reaches its own Go handler when the
// scope is missing.
const requiredScopeAdvanced = "mcp:advanced"

// toolScopeRequirements drives main.go's requireToolScope middleware: a tool
// listed here is rejected at the HTTP layer (403 + WWW-Authenticate) before
// the MCP server ever sees the call, if the bearer token lacks the scope.
var toolScopeRequirements = map[string]string{
	"whoami_advanced": requiredScopeAdvanced,
}

type whoamiAdvancedArgs struct{}

// whoamiAdvancedHandler only runs when requireToolScope already let the call
// through, i.e. the token actually carries mcp:advanced — which never happens
// in this spike (oauth.go always grants "mcp"). It exists for symmetry/
// completeness, not because a live run is expected to hit it.
//
// This is the *core* MCP spec's mechanism (authorization#scope-challenge-handling):
// "HTTP 403 Forbidden ... WWW-Authenticate header with ... error=insufficient_scope".
// A live Claude run on the first version of this spike (which used the
// ChatGPT-only _meta mechanism below for this exact tool) reported it got a
// plain tool error with no re-auth prompt, and pointed at this spec section
// as what it actually expected — see whoamiAdvancedMetaErrorHandler for the
// mechanism that confused it.
func whoamiAdvancedHandler(_ context.Context, _ *mcp.CallToolRequest, _ whoamiAdvancedArgs) (*mcp.CallToolResult, any, error) {
	return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: "you have mcp:advanced, congratulations"}}}, nil, nil
}

type whoamiAdvancedMetaErrorArgs struct{}

// whoamiAdvancedMetaErrorHandler is the ChatGPT Apps SDK's documented
// mechanism instead of the core spec's: a normal (HTTP 200) CallToolResult
// carrying _meta["mcp/www_authenticate"] with isError: true. Kept as its own
// tool, separate from whoami_advanced, so both mechanisms stay independently
// testable — a real host may expect one, the other, or (per the Claude run
// above) neither for the "wrong" tool.
func whoamiAdvancedMetaErrorHandler(ctx context.Context, _ *mcp.CallToolRequest, _ whoamiAdvancedMetaErrorArgs) (*mcp.CallToolResult, any, error) {
	info := auth.TokenInfoFromContext(ctx)
	hasScope := false
	if info != nil {
		for _, s := range info.Scopes {
			if s == requiredScopeAdvanced {
				hasScope = true
				break
			}
		}
	}
	if hasScope {
		return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: "you have mcp:advanced, congratulations"}}}, nil, nil
	}

	challenge := fmt.Sprintf(
		`Bearer resource_metadata="%s/.well-known/oauth-protected-resource", error="insufficient_scope", scope="%s", error_description="mcp:advanced is required for this tool"`,
		issuerForTools(), requiredScopeAdvanced,
	)
	return &mcp.CallToolResult{
		IsError: true,
		Meta:    mcp.Meta{"mcp/www_authenticate": []string{challenge}},
		Content: []mcp.Content{&mcp.TextContent{Text: "insufficient_scope: mcp:advanced is required"}},
	}, nil, nil
}

func newMCPServer() *mcp.Server {
	s := mcp.NewServer(&mcp.Implementation{Name: "mathtrail-spike-oauth", Version: "0.1.0"}, &mcp.ServerOptions{
		Instructions: "Spike server for T04: exercises the MCP OAuth 2.1 dance. Not product code.",
	})

	mcp.AddTool(s, &mcp.Tool{
		Name:        "whoami",
		Description: "Return the client_id and scopes carried by the bearer token used for this call.",
		Meta: mcp.Meta{
			"securitySchemes": []map[string]any{{"type": "oauth2", "scopes": []string{"mcp"}}},
		},
	}, whoamiHandler)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "whoami_advanced",
		Description: "Like whoami, but requires the mcp:advanced scope, which this spike's authorization server never grants. Rejected at the HTTP layer (403 + WWW-Authenticate), per the core MCP spec's scope-challenge-handling section.",
		Meta: mcp.Meta{
			"securitySchemes": []map[string]any{{"type": "oauth2", "scopes": []string{"mcp:advanced"}}},
		},
	}, whoamiAdvancedHandler)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "whoami_advanced_metaerror",
		Description: "Same idea as whoami_advanced, but signals insufficient scope via _meta[\"mcp/www_authenticate\"] in a normal CallToolResult, per ChatGPT's Apps SDK auth guidance rather than the core MCP spec.",
		Meta: mcp.Meta{
			"securitySchemes": []map[string]any{{"type": "oauth2", "scopes": []string{"mcp:advanced"}}},
		},
	}, whoamiAdvancedMetaErrorHandler)

	return s
}

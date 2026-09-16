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

type whoamiAdvancedArgs struct{}

// whoamiAdvancedHandler demonstrates the step-up / insufficient-scope path
// (SPEC: authorization#scope-challenge-handling): this spike's /authorize
// always grants only "mcp" (oauth.go), so a call here always needs a scope
// the token doesn't have. Rather than a transport-level 401/403, this reports
// the failure inside a normal (HTTP 200) CallToolResult carrying
// _meta["mcp/www_authenticate"], per ChatGPT's Apps SDK auth guidance — the
// RUN.md T04 checklist item this tool exists to exercise.
func whoamiAdvancedHandler(ctx context.Context, _ *mcp.CallToolRequest, _ whoamiAdvancedArgs) (*mcp.CallToolResult, any, error) {
	const requiredScope = "mcp:advanced"
	info := auth.TokenInfoFromContext(ctx)
	hasScope := false
	if info != nil {
		for _, s := range info.Scopes {
			if s == requiredScope {
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
		issuerForTools(), requiredScope,
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
		Description: "Like whoami, but requires the mcp:advanced scope, which this spike's authorization server never grants — demonstrates the insufficient-scope / step-up error path.",
		Meta: mcp.Meta{
			"securitySchemes": []map[string]any{{"type": "oauth2", "scopes": []string{"mcp:advanced"}}},
		},
	}, whoamiAdvancedHandler)

	return s
}

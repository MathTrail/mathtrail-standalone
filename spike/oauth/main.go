// Command spike-oauth is throwaway T04 spike code: a stub OAuth 2.1
// authorization server that approves everyone without Google, protecting a
// minimal MCP server, to see how Claude and ChatGPT actually walk the OAuth
// dance (registration, authorize, token, refresh, 401 mid-conversation).
// Deleted in T17; nothing here follows product or Go-standards conventions
// from CLAUDE.md — no real login, no encryption (T04's own brief).
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"

	"github.com/modelcontextprotocol/go-sdk/auth"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	srv := &oauthServer{store: newStore(), logger: logger}
	mcpServer := newMCPServer()

	streamable := mcp.NewStreamableHTTPHandler(
		func(*http.Request) *mcp.Server { return mcpServer },
		&mcp.StreamableHTTPOptions{
			Stateless: true,
			Logger:    logger,
			// Requests arrive from cloudflared via 127.0.0.1 carrying the real
			// public tunnel hostname as Host — exactly what the SDK's DNS-rebinding
			// check exists to block. Disabling it is what makes the dynamic-issuer
			// trick (main.go's mcpHandler, oauth.go's issuer()) possible without a
			// startup flag. Spike-only: production (T47-T49) runs on its own real
			// domain and keeps this protection on.
			DisableLocalhostProtection: true,
		},
	)

	mux := http.NewServeMux()
	mux.HandleFunc("/.well-known/oauth-protected-resource", srv.protectedResourceMetadata)
	mux.HandleFunc("/.well-known/oauth-authorization-server", srv.authServerMetadata)
	mux.HandleFunc("/register", srv.register)
	mux.HandleFunc("/authorize", srv.authorize)
	mux.HandleFunc("/token", srv.token)
	mux.HandleFunc("/mcp", srv.mcpHandler(requireToolScope(streamable)))
	mux.HandleFunc("/spike/revoke", srv.revoke)
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8082"
	}
	addr := ":" + port
	logger.Info("spike_oauth_listening", "addr", addr)
	if err := http.ListenAndServe(addr, mux); err != nil { //nolint:gosec // spike: no timeouts needed
		logger.Error("server_failed", "error", err.Error())
		os.Exit(1)
	}
}

// mcpHandler wraps the streamable handler with bearer-token enforcement,
// rebuilding auth.RequireBearerTokenOptions on every request so its
// ResourceMetadataURL always points at *this* request's own Host — see
// issuer() in oauth.go for why that can't be a fixed startup value.
func (s *oauthServer) mcpHandler(streamable http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		currentIssuer.Store(issuer(r))
		opts := &auth.RequireBearerTokenOptions{
			ResourceMetadataURL: issuer(r) + "/.well-known/oauth-protected-resource",
			Scopes:              []string{"mcp"},
		}
		auth.RequireBearerToken(s.verifyToken, opts)(streamable).ServeHTTP(w, r)
	}
}

// jsonrpcToolCallPeek is just enough of a tools/call request body to read the
// tool name; every other field is left to the real MCP server to parse.
type jsonrpcToolCallPeek struct {
	Method string `json:"method"`
	Params struct {
		Name string `json:"name"`
	} `json:"params"`
}

// requireToolScope sits between RequireBearerToken and the streamable
// handler. RequireBearerToken already checked the token has "mcp" (the one
// scope this spike's /authorize ever grants); this adds a *per-tool* check
// for tools in mcp.go's toolScopeRequirements, rejecting with the core MCP
// spec's mechanism for insufficient scope — an HTTP 403 carrying
// WWW-Authenticate (authorization#scope-challenge-handling) — rather than
// letting the call reach a CallToolResult. A live Claude run against the
// first version of this spike (which only had the ChatGPT-style
// _meta["mcp/www_authenticate"] mechanism, now whoami_advanced_metaerror)
// reported getting a plain tool error with no re-auth prompt, and named this
// exact mechanism as what it expected instead.
func requireToolScope(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
		if err != nil {
			http.Error(w, "failed to read request body", http.StatusBadRequest)
			return
		}
		r.Body = io.NopCloser(bytes.NewReader(body))

		var peek jsonrpcToolCallPeek
		if json.Unmarshal(body, &peek) == nil && peek.Method == "tools/call" {
			if requiredScope, tracked := toolScopeRequirements[peek.Params.Name]; tracked {
				info := auth.TokenInfoFromContext(r.Context())
				if !tokenHasScope(info, requiredScope) {
					writeInsufficientScope(w, r, requiredScope)
					return
				}
			}
		}
		next.ServeHTTP(w, r)
	})
}

func tokenHasScope(info *auth.TokenInfo, scope string) bool {
	if info == nil {
		return false
	}
	for _, s := range info.Scopes {
		if s == scope {
			return true
		}
	}
	return false
}

func writeInsufficientScope(w http.ResponseWriter, r *http.Request, scope string) {
	resourceMetadataURL := issuer(r) + "/.well-known/oauth-protected-resource"
	w.Header().Set("WWW-Authenticate", fmt.Sprintf(
		`Bearer error="insufficient_scope", scope=%q, resource_metadata=%q, error_description="this tool requires additional scope"`,
		scope, resourceMetadataURL,
	))
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusForbidden)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"error":             "insufficient_scope",
		"error_description": "this tool requires additional scope",
		"scope":             scope,
	})
}

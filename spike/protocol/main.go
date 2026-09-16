// Command spike-protocol is throwaway T03 spike code: it exercises MCP 2026-07-28 over
// stateless Streamable HTTP and an MCP Apps widget against real Claude and ChatGPT hosts.
// Deleted in T17; nothing here follows product or Go-standards conventions from CLAUDE.md.
package main

import (
	"context"
	_ "embed"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"sync/atomic"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

//go:embed embed/widget.html
var widgetHTML string

const widgetURI = "ui://spike/widget"

// spikeDrawing is a monospace text drawing, 30 columns wide, per T03.
// Every row (border and content) is exactly 30 runes — a live Claude run
// caught a first version where the content rows were one column wider than
// the border, so this is spelled out explicitly rather than eyeballed.
const spikeDrawing = "" +
	"┌────────────────────────────┐\n" +
	"│ 2 + 2 = ?                  │\n" +
	"│ A) 3   B) 4   C) 5   D) 22 │\n" +
	"└────────────────────────────┘"

type echoArgs struct {
	Message string `json:"message" jsonschema:"text to echo back"`
}

type echoResult struct {
	Echo       string `json:"echo"`
	ReceivedAt string `json:"receivedAt"`
}

// echoHandler is a plain tool: visible to the model, returns both content and
// structuredContent, no widget attached.
func echoHandler(_ context.Context, _ *mcp.CallToolRequest, args echoArgs) (*mcp.CallToolResult, echoResult, error) {
	out := echoResult{Echo: args.Message, ReceivedAt: time.Now().UTC().Format(time.RFC3339)}
	res := &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("echo: %s", args.Message)}},
	}
	return res, out, nil
}

type widgetPokeArgs struct {
	Note string `json:"note,omitempty" jsonschema:"optional note from the widget"`
}

type widgetPokeResult struct {
	Counter   int    `json:"counter"`
	Note      string `json:"note,omitempty"`
	ServerNow string `json:"serverNow"`
}

// widgetPokeHandler is an app-only tool (_meta.ui.visibility: ["app"]): the host
// must hide it from the model's tool list and only allow the widget to call it.
func widgetPokeHandler(counter *atomic.Int64) mcp.ToolHandlerFor[widgetPokeArgs, widgetPokeResult] {
	return func(_ context.Context, _ *mcp.CallToolRequest, args widgetPokeArgs) (*mcp.CallToolResult, widgetPokeResult, error) {
		n := counter.Add(1)
		out := widgetPokeResult{Counter: int(n), Note: args.Note, ServerNow: time.Now().UTC().Format(time.RFC3339)}
		return nil, out, nil
	}
}

type showWidgetArgs struct{}

type showWidgetResult struct {
	Greeting string `json:"greeting"`
	Drawing  string `json:"drawing"`
}

// showWidgetHandler is a tool carrying _meta.ui.resourceUri: the host should
// render the spike widget for its result.
func showWidgetHandler(_ context.Context, _ *mcp.CallToolRequest, _ showWidgetArgs) (*mcp.CallToolResult, showWidgetResult, error) {
	out := showWidgetResult{Greeting: "MathTrail spike widget", Drawing: spikeDrawing}
	return nil, out, nil
}

func widgetResourceHandler(_ context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
	return &mcp.ReadResourceResult{
		Contents: []*mcp.ResourceContents{{
			URI:      req.Params.URI,
			MIMEType: "text/html;profile=mcp-app",
			Text:     widgetHTML,
		}},
	}, nil
}

// versionedRequest is satisfied by every *mcp.ServerRequest[P] instantiation;
// asserting Request against it recovers the per-request protocol version and
// clientInfo the SDK parses from _meta on 2026-07-28 requests (T03: log both).
type versionedRequest interface {
	ProtocolVersion() string
	ClientInfo() *mcp.Implementation
}

func logRequestMiddleware(logger *slog.Logger) mcp.Middleware {
	return func(next mcp.MethodHandler) mcp.MethodHandler {
		return func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
			attrs := []any{"method", method}
			if v, ok := req.(versionedRequest); ok {
				attrs = append(attrs, "protocol_version", v.ProtocolVersion())
				if ci := v.ClientInfo(); ci != nil {
					attrs = append(attrs, "client_name", ci.Name, "client_version", ci.Version)
				}
			}
			logger.Info("mcp_request", attrs...)
			return next(ctx, method, req)
		}
	}
}

func newServer(logger *slog.Logger) *mcp.Server {
	opts := &mcp.ServerOptions{
		Instructions: "Spike server for T03: exercises the MCP 2026-07-28 stateless transport and an MCP Apps widget. Not product code.",
		Logger:       logger,
	}
	// "only 2026-07-28" mode: enabled after the first unrestricted measurement, per T03.
	if os.Getenv("STRICT_PROTOCOL") == "1" {
		opts.SupportedProtocolVersions = []string{"2026-07-28"}
	}

	s := mcp.NewServer(&mcp.Implementation{Name: "mathtrail-spike-protocol", Version: "0.1.0"}, opts)
	s.AddReceivingMiddleware(logRequestMiddleware(logger))

	mcp.AddTool(s, &mcp.Tool{
		Name:        "echo",
		Description: "Echo a message back as both content and structuredContent. Plain test tool, no widget.",
	}, echoHandler)

	var pokeCounter atomic.Int64
	mcp.AddTool(s, &mcp.Tool{
		Name:        "widget_poke",
		Description: "App-only tool: callable from the widget via callServerTool, hidden from the model's tool list.",
		Meta: mcp.Meta{
			"ui": map[string]any{"visibility": []string{"app"}},
		},
	}, widgetPokeHandler(&pokeCounter))

	mcp.AddTool(s, &mcp.Tool{
		Name:        "show_widget",
		Description: "Return a short greeting and a text drawing; the host should render the spike widget for the result.",
		Meta: mcp.Meta{
			"ui": map[string]any{"resourceUri": widgetURI},
		},
	}, showWidgetHandler)

	s.AddResource(&mcp.Resource{
		URI:      widgetURI,
		Name:     "spike-widget",
		MIMEType: "text/html;profile=mcp-app",
		Meta: mcp.Meta{
			// No external domains: the widget is a single embedded file (T03 "без внешних доменов").
			"ui": map[string]any{"csp": map[string]any{}},
		},
	}, widgetResourceHandler)

	return s
}

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	server := newServer(logger)

	streamable := mcp.NewStreamableHTTPHandler(
		func(*http.Request) *mcp.Server { return server },
		&mcp.StreamableHTTPOptions{
			Stateless: true,
			Logger:    logger,
		},
	)

	mux := http.NewServeMux()
	mux.Handle("/mcp", streamable)
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}
	addr := ":" + port
	logger.Info("spike_protocol_listening", "addr", addr, "strict_protocol", os.Getenv("STRICT_PROTOCOL") == "1")
	if err := http.ListenAndServe(addr, mux); err != nil { //nolint:gosec // spike: no timeouts needed
		logger.Error("server_failed", "error", err.Error())
		os.Exit(1)
	}
}

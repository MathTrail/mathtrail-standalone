// Package mcpserver serves the MCP endpoint: the protocol over stateless
// Streamable HTTP, the sign-in in front of it, and the frame every tool call
// passes through on its way in and out.
package mcpserver

import (
	"errors"
	"fmt"
	"net/http"
	"regexp"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

// maxRequestBody is the most a request to the endpoint may carry. A submitted
// task with its solver is some twenty kilobytes, so a megabyte leaves room for
// any honest request and none for a flood.
const maxRequestBody = 1 << 20

// toolNamePattern is what the protocol lets a tool be called: one to 128
// letters, digits, dots, dashes and underscores.
var toolNamePattern = regexp.MustCompile(`^[A-Za-z0-9_.-]{1,128}$`)

// Settings are what the endpoint is built from.
type Settings struct {
	// Instructions are what the endpoint tells a model about the service when
	// it connects.
	Instructions string
	// InstructionsVersion names the wording the model was given, so that the
	// line every call leaves says which one it had read.
	InstructionsVersion string
	// Version is the build serving the endpoint.
	Version string
	// SignIn decides whose account a request acts for, or refuses it.
	SignIn SignIn
	// Traces records a span for every tool call.
	Traces trace.TracerProvider
	// Logger writes the line every tool call leaves.
	Logger *zap.Logger
	// ProjectID names the project a line's trace is filed under.
	ProjectID string
}

// ErrSettings is returned when the endpoint is given settings it cannot be
// served with; callers branch on it with errors.Is.
var ErrSettings = errors.New("mcp: settings")

// validate refuses settings the endpoint could not be served with, naming what
// is missing. A missing sign-in above all: an endpoint built without one would
// answer anybody.
func (s *Settings) validate() error {
	switch {
	case s == nil:
		return fmt.Errorf("%w: settings must be given", ErrSettings)
	case s.SignIn == nil:
		return fmt.Errorf("%w: SignIn must be set", ErrSettings)
	case s.Traces == nil:
		return fmt.Errorf("%w: Traces must be set", ErrSettings)
	case s.Logger == nil:
		return fmt.Errorf("%w: Logger must be set", ErrSettings)
	case s.Instructions == "":
		return fmt.Errorf("%w: Instructions must be set", ErrSettings)
	case s.InstructionsVersion == "":
		// Every call's line and span name the wording the model had read; an
		// empty version would leave no call tied to one.
		return fmt.Errorf("%w: InstructionsVersion must be set", ErrSettings)
	}
	return nil
}

// NewHandler builds the endpoint as it is served: the protocol, the tools, the
// frame around every call and the sign-in in front of all of it.
//
// The protocol is served without sessions. Every request carries what the
// server needs to answer it, so any instance can answer any request and none
// has anything to keep between them. The set of protocol versions is left as
// the library has it: a host that probes with an older version before it
// settles on the newest is answered rather than refused.
//
// The library logs nothing here. Its lines would not be held to the fields this
// service allows in a line, and what they would report — a request the
// protocol refused — the request's own line already counts.
func NewHandler(settings *Settings, tools ...Tool) (http.Handler, error) {
	if err := settings.validate(); err != nil {
		return nil, err
	}

	server := mcp.NewServer(
		&mcp.Implementation{Name: "mathtrail", Title: "MathTrail", Version: settings.Version},
		&mcp.ServerOptions{
			Instructions: settings.Instructions,
			// Declared rather than left to the default, which also advertises
			// logging: a feature the protocol has deprecated and this service
			// never offered. Tools are declared even while there are none, so
			// that a client asking for the list is told it is empty.
			Capabilities: &mcp.ServerCapabilities{Tools: &mcp.ToolCapabilities{}},
		},
	)

	names := make(map[string]struct{}, len(tools))
	for _, tool := range tools {
		name := tool.name()
		// The library takes a name it would refuse, and a name already taken,
		// and says so to nobody: it keeps the first, replaces the second, and a
		// host may drop either.
		if !toolNamePattern.MatchString(name) {
			return nil, fmt.Errorf("%w: %q is not a name a tool can have", ErrSettings, name)
		}
		if _, taken := names[name]; taken {
			return nil, fmt.Errorf("%w: two tools are named %q", ErrSettings, name)
		}
		names[name] = struct{}{}
		if err := tool.add(server); err != nil {
			return nil, err
		}
	}
	server.AddReceivingMiddleware(newBoundary(settings, names).wrap)

	protocol := mcp.NewStreamableHTTPHandler(
		func(*http.Request) *mcp.Server { return server },
		&mcp.StreamableHTTPOptions{
			Stateless:           true,
			MaxRequestBodyBytes: maxRequestBody,
		},
	)
	return neverStored(settings.SignIn(protocol)), nil
}

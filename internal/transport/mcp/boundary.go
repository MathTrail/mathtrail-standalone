package mcpserver

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/modelcontextprotocol/go-sdk/jsonrpc"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"

	"github.com/MathTrail/mathtrail-standalone/internal/logger"
	"github.com/MathTrail/mathtrail-standalone/internal/telemetry"
)

// tracerScope is what the spans of tool calls are recorded under, so that a
// reader can tell them from the ones a library produced.
const tracerScope = "github.com/MathTrail/mathtrail-standalone/internal/transport/mcp"

// How a tool call ended, as its line and its span say it.
const (
	// outcomeOK is a call the tool answered.
	outcomeOK = "ok"
	// outcomeRefused is an answer too, one the model is meant to act on: the
	// task failed its checks, a limit was reached, the request went stale.
	outcomeRefused = "refused"
	// outcomeFailed is a call the tool could not do, for a reason of ours.
	outcomeFailed = "failed"
	// outcomeInvalid is a call the protocol refused before any tool ran: a
	// tool nobody defined, or arguments that do not fit the tool's schema.
	outcomeInvalid = "invalid"
)

// Why a call was invalid.
const (
	kindProtocol  = "protocol"
	kindArguments = "arguments"
)

// boundary is where every request the protocol hands the server passes, on its
// way to a handler and back.
//
// A tool's own frame is not enough to see a call whole, because much of what
// happens to one happens where no tool runs: the library refuses arguments
// that do not fit a tool's schema and a tool nobody defined before the tool is
// reached, and it fails on a payload that does not fit the tool's own output
// schema after the tool has returned. A call is observed here once, whichever
// of those it met. A panic anywhere below is caught here too: the library runs
// handlers on goroutines of its own, which the HTTP layer's recovery never
// sees, and an uncaught panic there ends the process for every child using it.
type boundary struct {
	tracer              trace.Tracer
	logger              *zap.Logger
	projectID           string
	instructionsVersion string
	tools               map[string]struct{}
}

func newBoundary(s *Settings, tools map[string]struct{}) *boundary {
	return &boundary{
		tracer:              s.Traces.Tracer(tracerScope),
		logger:              s.Logger,
		projectID:           s.ProjectID,
		instructionsVersion: s.InstructionsVersion,
		tools:               tools,
	}
}

// wrap puts the boundary around the handler of every method.
//
// A request that is not a tool call is passed on as it is, and so is what the
// library answers it with: every such request is the library's own to handle,
// and its errors describe the request — a cursor that does not decode, a
// method nobody offers — in the protocol's terms. What the boundary adds to
// every request is the last word on a panic: one below, in the library, or in
// the boundary's own work around a tool call, is answered as an internal error
// and written as a line, rather than ending the process.
func (b *boundary) wrap(next mcp.MethodHandler) mcp.MethodHandler {
	return func(ctx context.Context, method string, req mcp.Request) (result mcp.Result, err error) {
		defer func() {
			if recovered := recover(); recovered != nil {
				b.logger.Error("mcp_panic", b.panicFields(ctx, method, recovered)...)
				result, err = nil, internalError()
			}
		}()
		if call, isCall := req.(*mcp.CallToolRequest); isCall {
			return b.toolCall(ctx, method, call, next)
		}
		return next(ctx, method, req)
	}
}

// panicFields are the fields of the line a panic outside a tool leaves. They
// are read from the deferred function that recovered, while the panic is still
// on the stack.
func (b *boundary) panicFields(ctx context.Context, method string, recovered any) []zap.Field {
	fields := []zap.Field{
		logger.PanicValue(recovered),
		logger.PanicStack(),
		zap.String("method", methodLabel(method)),
		zap.String("request_id", logger.RequestID(ctx)),
	}
	if account, signedIn := accountFrom(ctx); signedIn {
		fields = append(fields, zap.String("user", account.ID))
	}
	return append(fields, telemetry.LogFields(ctx, b.projectID)...)
}

// toolCall runs one tool call inside a span of its own and leaves one line
// saying how it ended.
func (b *boundary) toolCall(ctx context.Context, method string, req *mcp.CallToolRequest, next mcp.MethodHandler) (result mcp.Result, err error) {
	call := b.begin(ctx, req)
	defer func() {
		if recovered := recover(); recovered != nil {
			result, err = call.panicked(recovered), nil
		}
		b.finish(call, result, err)
	}()

	result, err = next(call.ctx, method, req)
	return settle(result, err)
}

// observedCall is what the boundary knows about one tool call while it runs.
type observedCall struct {
	// ctx carries the call's span, so that whatever the tool does is recorded
	// inside it.
	ctx      context.Context
	span     trace.Span
	start    time.Time
	tool     string
	protocol string
	client   string
	user     string
	panic    []zap.Field
}

// begin opens the call's span as a child of the request's.
//
// The protocol lets a client send a trace context of its own inside the call,
// and the conventions for tracing it suggest taking that as the parent. It is
// not taken: one request is one tree — the request, the tool, what the tool
// did — and which traces are kept is this service's decision, which a client
// that asked for every call to be traced would otherwise be making.
//
// What is known before the call runs is set here and how it ended in finish.
// Every value is a word from a closed list, so neither half carries anything a
// caller wrote — which matters for the second half, since the filter that
// blanks a caller's details from a trace reads a span only as it starts.
func (b *boundary) begin(ctx context.Context, req *mcp.CallToolRequest) *observedCall {
	tool := b.toolLabel(toolName(req))
	protocol := protocolLabel(req.ProtocolVersion())
	spanned, span := b.tracer.Start(ctx, "tools/call "+tool,
		trace.WithSpanKind(trace.SpanKindServer),
		trace.WithAttributes(
			attribute.String("mcp.method.name", "tools/call"),
			attribute.String("gen_ai.operation.name", "execute_tool"),
			attribute.String("gen_ai.tool.name", tool),
			attribute.String("mcp.protocol.version", protocol),
			attribute.String("mathtrail.instructions_version", b.instructionsVersion),
		),
	)

	call := &observedCall{
		ctx:      spanned,
		span:     span,
		start:    time.Now(),
		tool:     tool,
		protocol: protocol,
		client:   clientFamily(req.ClientInfo()),
	}
	if account, signedIn := accountFrom(ctx); signedIn {
		call.user = account.ID
	}
	return call
}

// panicked answers a call whose handling panicked, as a failure of ours, and
// writes down what it panicked with and where. It is called from the deferred
// function that recovered, while the panic is still on the stack.
func (c *observedCall) panicked(recovered any) mcp.Result {
	c.panic = []zap.Field{logger.PanicValue(recovered), logger.PanicStack()}
	return failed(&failure{kind: kindPanic, sentence: sentenceInternal})
}

// finish closes the call's span and writes its line.
func (b *boundary) finish(call *observedCall, result mcp.Result, err error) {
	ended := judge(result, err)

	call.span.SetAttributes(attribute.String("mathtrail.tool.outcome", ended.outcome))
	if ended.status != "" {
		call.span.SetAttributes(attribute.String("mathtrail.tool.status", ended.status))
	}
	if ended.kind != "" {
		call.span.SetAttributes(attribute.String("error.type", ended.kind))
	}
	// Only a failure of ours marks the span. A refusal is an answer, and a
	// call the protocol refused is the caller's mistake; neither is something
	// that went wrong here. The error itself is never recorded on the span:
	// its text is whatever the code that made it had in hand.
	if ended.outcome == outcomeFailed {
		call.span.SetStatus(codes.Error, "the tool call failed")
	}
	call.span.End()

	fields := b.callFields(call, ended)
	switch ended.outcome {
	case outcomeFailed:
		b.logger.Error("tool_call", fields...)
	case outcomeInvalid:
		b.logger.Warn("tool_call", fields...)
	default:
		b.logger.Info("tool_call", fields...)
	}
}

// callFields are the fields of a tool call's line. Each is a closed word, a
// number or an identifier this service made; nothing the caller wrote.
func (b *boundary) callFields(call *observedCall, ended verdict) []zap.Field {
	fields := []zap.Field{
		zap.String("tool", call.tool),
		zap.String("outcome", ended.outcome),
		zap.Int64("duration_ms", time.Since(call.start).Milliseconds()),
		zap.String("instructions_version", b.instructionsVersion),
		zap.String("protocol_version", call.protocol),
		zap.String("client", call.client),
		zap.String("request_id", logger.RequestID(call.ctx)),
	}
	if ended.status != "" {
		fields = append(fields, zap.String("status", ended.status))
	}
	if ended.kind != "" {
		fields = append(fields, zap.String("error", ended.kind))
	}
	if call.user != "" {
		fields = append(fields, zap.String("user", call.user))
	}
	fields = append(fields, call.panic...)
	return append(fields, telemetry.LogFields(call.ctx, b.projectID)...)
}

// verdict is how a call ended.
type verdict struct {
	outcome string
	// status is a refusal's status, the one its payload gives.
	status string
	// kind is why a call failed or was invalid.
	kind string
}

// judge reads how a call ended from what the protocol is about to send.
func judge(result mcp.Result, err error) verdict {
	if err != nil {
		// The protocol has one word for a fault on this side, and that one is
		// ours; every other error it answers with is the caller's mistake.
		var wire *jsonrpc.Error
		if errors.As(err, &wire) && wire.Code == jsonrpc.CodeInternalError {
			return verdict{outcome: outcomeFailed, kind: kindInternal}
		}
		return verdict{outcome: outcomeInvalid, kind: kindProtocol}
	}
	answered, isAnswer := result.(*mcp.CallToolResult)
	if !isAnswer || answered == nil {
		return verdict{outcome: outcomeOK}
	}
	if answered.IsError {
		var ours *failure
		if errors.As(answered.GetError(), &ours) {
			return verdict{outcome: outcomeFailed, kind: ours.kind}
		}
		// The library refused the arguments before any tool saw them.
		return verdict{outcome: outcomeInvalid, kind: kindArguments}
	}
	if status, refused := refusal(answered); refused {
		return verdict{outcome: outcomeRefused, status: status}
	}
	return verdict{outcome: outcomeOK}
}

// refusal reads the status a result gives itself at the top of its payload,
// when that status is one of the refusals a model can act on. That field is
// what makes a result a refusal, for the model and the widget as much as here,
// so it is read rather than reported a second time by the tool.
func refusal(result *mcp.CallToolResult) (string, bool) {
	payload, isJSON := result.StructuredContent.(json.RawMessage)
	if !isJSON {
		return "", false
	}
	var top struct {
		Status string `json:"status"`
	}
	if json.Unmarshal(payload, &top) != nil {
		return "", false
	}
	switch top.Status {
	case "rejected", "limited", "stale":
		return top.Status, true
	}
	return "", false
}

// settle keeps the protocol's own refusals, in their own words, and replaces
// anything else that went wrong below the boundary with a failure of ours. The
// library sends an error with its whole text as the message — the text of
// whatever wrapped a protocol error included — and that text is not ours to
// send: a payload that did not fit its own schema, for one, is described in
// the words of the payload.
func settle(result mcp.Result, err error) (mcp.Result, error) {
	if err == nil {
		return result, nil
	}
	var wire *jsonrpc.Error
	if errors.As(err, &wire) {
		return nil, wire
	}
	return failed(explain(err)), nil
}

// failed is the result of a call that failed: the failure's sentence, marked as
// an error for the model to read, with the failure itself kept for the line.
func failed(f *failure) *mcp.CallToolResult {
	var result mcp.CallToolResult
	result.SetError(f)
	return &result
}

// internalError is what a request that is not a tool call is answered with when
// something of ours failed under it.
func internalError() error {
	return &jsonrpc.Error{Code: jsonrpc.CodeInternalError, Message: "internal error"}
}

// toolName is the tool a call asks for.
func toolName(req *mcp.CallToolRequest) string {
	if req.Params == nil {
		return ""
	}
	return req.Params.Name
}

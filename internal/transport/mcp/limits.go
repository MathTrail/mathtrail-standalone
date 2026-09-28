package mcpserver

import (
	"context"
	"slices"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/jsonrpc"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"go.uber.org/zap"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/profile"
	"github.com/MathTrail/mathtrail-standalone/internal/ratelimit"
)

// The ceilings a call can be held to, as the line of one reached names it.
const (
	limitAccount     = "user_rate"
	limitInstance    = "instance_rate"
	limitDailyTasks  = "daily_tasks"
	limitDailyFailed = "daily_failed"
)

// eventLimitHit is the line of a ceiling reached.
const eventLimitHit = "limit_hit"

// Limits are the paces every message of a signed-in account is held to before
// anything is done for it: the account's own, and the instance's, which it
// shares with everybody.
type Limits struct {
	// PerAccount counts the messages of each account apart.
	PerAccount ratelimit.Limiter
	// Instance counts the messages of everybody together.
	Instance ratelimit.Limiter
}

// ceilings are the paces in the order a message is held to them. The
// account's own comes first, so that an account in a runaway is stopped by its
// own ceiling before it spends what the instance has for everybody; a message
// the instance turns away costs the account nothing.
func (l Limits) ceilings() []ratelimit.Ceiling {
	return []ratelimit.Ceiling{
		{Name: limitAccount, Limiter: l.PerAccount},
		{Name: limitInstance, Limiter: l.Instance},
	}
}

// admitted counts one message of the account against the paces, and says
// whether it may go ahead. A ceiling reached is written down once for a flood.
func (b *boundary) admitted(ctx context.Context, user string) bool {
	refusedBy, verdict := ratelimit.Admit(user, b.ceilings...)
	if !verdict.Allowed && verdict.Began {
		limitHit(ctx, b.logger, b.projectID, user, refusedBy)
	}
	return verdict.Allowed
}

// counted says whether a message is held to the paces. A notification is not:
// it asks for no answer, so a refusal would only lose it without a word, and
// it asks for no work.
func counted(method string) bool {
	return !strings.HasPrefix(method, "notifications/")
}

// limitHit writes the line of a ceiling reached, as a warning whichever it is:
// what reaches one is worth a look — a runaway, or a topic the models cannot
// write a task on. It names the ceiling and whom it held back, and anything the
// ceiling adds, such as the counter of a day; never an address.
func limitHit(ctx context.Context, logger *zap.Logger, projectID, user, limit string, fields ...zap.Field) {
	logger.Warn(eventLimitHit, slices.Concat(
		[]zap.Field{zap.String("limit", limit)},
		fields,
		callerFields(ctx, user, projectID),
	)...)
}

// sentenceHeldBack is what the model is told of a call a pace held back.
// Nothing went wrong and nothing was done, and a moment's wait is all it
// takes. It names no number and no limit: a number invites arithmetic, and the
// child has nothing to do about a limit.
const sentenceHeldBack = "MathTrail received too many calls in a short time, so this call was not made. " +
	"Wait a moment, then make the same call again."

// heldBack is the answer to a call a pace held back. It is told the way a
// failure is — one sentence, marked as an error, and no payload — because it is
// decided before any tool runs, and a payload of the shape a tool declares
// cannot be made without the tool. It is a refusal all the same: nothing here
// went wrong, and the model's part is to wait.
type heldBack struct{}

func (heldBack) Error() string { return sentenceHeldBack }

// heldBackCall is the result of a call a pace held back.
func heldBackCall() *mcp.CallToolResult {
	var result mcp.CallToolResult
	result.SetError(heldBack{})
	return &result
}

// codeTooManyMessages is the protocol's error for a message a pace held back
// that is not a tool call: a host listing the tools or reading the widget's
// page. The protocol names no error for it, so it is the first of the codes
// JSON-RPC leaves to a server for errors of its own.
const codeTooManyMessages = -32000

// tooManyMessages answers such a message in the protocol's own terms, since no
// model reads it: the host sees it, and asks again.
func tooManyMessages() error {
	return &jsonrpc.Error{Code: codeTooManyMessages, Message: "too many requests; try again in a moment"}
}

// Daily are the ceilings of a child's day: how many tasks may be accepted, and
// how many requests may end with the model out of attempts, before next_task
// opens no more that day.
type Daily struct {
	Tasks  int
	Failed int
}

// reached is the first ceiling the day's counters stand at, if any: its name
// as a line gives it, and the counter's value.
func (d Daily) reached(today profile.Daily) (limit string, count int, reached bool) {
	switch {
	case today.Accepted >= d.Tasks:
		return limitDailyTasks, today.Accepted, true
	case today.Failed >= d.Failed:
		return limitDailyFailed, today.Failed, true
	}
	return "", 0, false
}

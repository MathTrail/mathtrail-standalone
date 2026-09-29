package session

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"
	"syscall"

	"github.com/modelcontextprotocol/go-sdk/jsonrpc"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Class is the family an answer belongs to.
type Class string

// The families of answers. Rejected, stale and limited are the service's own
// refusals, named by the status its payload carries; the rest are how a call
// can end without one.
const (
	// OK is an answer the tool gave as asked.
	OK Class = "ok"
	// Rejected, Stale and Limited are refusals the model is meant to act on.
	Rejected Class = "rejected"
	Stale    Class = "stale"
	Limited  Class = "limited"
	// Failure is a result marked as an error: a failure of the service's, told
	// in a sentence of its own.
	Failure Class = "failure"
	// JSONRPC is an error of the protocol rather than a result.
	JSONRPC Class = "jsonrpc"
	// HTTP is a request the service answered with a status of 400 or more.
	HTTP Class = "http"
	// NoAnswer is a call that got nothing back in time, or nothing at all.
	NoAnswer Class = "no_answer"
	// Mismatch is an answer that belongs to another call: a card that shows
	// another question or another child.
	Mismatch Class = "mismatch"
	// Handshake is a session that ended up at another version of the protocol.
	Handshake Class = "handshake"
	// Stopped is a call the run abandoned when it was asked to stop: nothing
	// the service said or failed to say.
	Stopped Class = "stopped"
)

// Kind is what an answer came to, in words a report can count by: a family,
// and within it the code, the sentence or the status that tells it apart.
type Kind struct {
	Class  Class
	Detail string
}

// String is the kind as a report writes it, such as "rejected:solver_error".
func (k Kind) String() string {
	if k.Detail == "" {
		return string(k.Class)
	}
	return string(k.Class) + ":" + k.Detail
}

// The kinds a scenario looks for by name.
var (
	// Answered is a call the tool answered as asked.
	Answered = Kind{Class: OK}
	// Busy is a task the sandbox had no slot for in time: never checked, and
	// costing no attempt.
	Busy = Kind{Class: Failure, Detail: "busy"}
	// Paced is a call or a fetch a pace held back before anything was done
	// for it.
	Paced = Kind{Class: Limited, Detail: "paced"}
	// Mismatched is an answer that belongs to another call.
	Mismatched = Kind{Class: Mismatch}
)

// The sentences the service begins two answers with, which say more than the
// families of their results: a task turned away for want of a slot, and a
// call held back by a pace. Both are results marked as errors, with no payload
// to say which they are.
const (
	busySentence  = "MathTrail is checking too many tasks right now"
	pacedSentence = "MathTrail received too many calls in a short time"
)

// A message other than a tool call that a pace holds back is refused with an
// error of the protocol: this code, the first a server may use for errors of
// its own, and a message that says which of them it is.
const (
	pacedCode    = -32000
	pacedMessage = "too many requests"
)

// rejectedByTransport is the code the protocol library puts on a request its
// own transport got no message back for — a status it reads as passing, a
// connection that never reached the service. It is the library's error rather
// than the service's, so the call is told by the status or by the network.
const rejectedByTransport = -32005

// kindOf tells what a call came to, from its result, the error in place of one
// and the worst status its requests got back.
func kindOf(result *mcp.CallToolResult, err error, status int) Kind {
	if err == nil && result != nil {
		return kindOfResult(result)
	}

	var rpc *jsonrpc.Error
	var handshake fellBack
	protocolError := errors.As(err, &rpc) && rpc.Code != rejectedByTransport
	switch {
	case errors.As(err, &handshake):
		return Kind{Class: Handshake, Detail: handshake.version}
	case status >= 400:
		// A status of 400 or more is told by its status, whatever error of the
		// protocol it carried, and that error goes with it.
		detail := strconv.Itoa(status)
		if protocolError {
			detail += "(" + strconv.FormatInt(rpc.Code, 10) + ")"
		}
		return Kind{Class: HTTP, Detail: detail}
	case protocolError && rpc.Code == pacedCode && strings.HasPrefix(rpc.Message, pacedMessage):
		return Paced
	case protocolError:
		return Kind{Class: JSONRPC, Detail: strconv.FormatInt(rpc.Code, 10)}
	case errors.Is(err, context.Canceled):
		return Kind{Class: Stopped}
	case errors.Is(err, context.DeadlineExceeded):
		return Kind{Class: NoAnswer, Detail: "timeout"}
	case errors.Is(err, syscall.ECONNREFUSED):
		return Kind{Class: NoAnswer, Detail: "refused"}
	case errors.Is(err, syscall.ECONNRESET), errors.Is(err, io.EOF), errors.Is(err, io.ErrUnexpectedEOF):
		return Kind{Class: NoAnswer, Detail: "reset"}
	case errors.Is(err, mcp.ErrConnectionClosed):
		return Kind{Class: NoAnswer, Detail: "closed"}
	}
	return Kind{Class: NoAnswer, Detail: "other"}
}

// kindOfStatus tells what a fetch came to, from the status it got and the
// error that ended it. The service's pace answers a request it held back with
// 429, and it is told as the pace a call is held back by; any other status of
// 400 or more, or one that is neither served nor refused, is told by itself.
func kindOfStatus(status int, err error) Kind {
	switch {
	case status == http.StatusTooManyRequests:
		return Paced
	case status >= http.StatusBadRequest:
		return Kind{Class: HTTP, Detail: strconv.Itoa(status)}
	case err != nil:
		return kindOf(nil, err, status)
	case status >= http.StatusOK && status < http.StatusMultipleChoices:
		return Answered
	}
	return Kind{Class: HTTP, Detail: strconv.Itoa(status)}
}

// kindOfResult tells a result apart: a failure by the sentence it begins with,
// a refusal by the status and code of its payload, and anything else as an
// answer.
func kindOfResult(result *mcp.CallToolResult) Kind {
	if result.IsError {
		text := ""
		if len(result.Content) > 0 {
			if block, isText := result.Content[0].(*mcp.TextContent); isText {
				text = block.Text
			}
		}
		switch {
		case strings.HasPrefix(text, busySentence):
			return Busy
		case strings.HasPrefix(text, pacedSentence):
			return Paced
		}
		return Kind{Class: Failure, Detail: firstSentence(text)}
	}

	payload, _ := result.StructuredContent.(map[string]any)
	status, _ := payload["status"].(string)
	if status == "" {
		return Answered
	}
	code, _ := payload["code"].(string)
	return Kind{Class: Class(status), Detail: code}
}

// firstSentence is a text up to the end of its first sentence: every failure
// the service tells has a first sentence of its own, and the rest says what to
// do about it.
func firstSentence(text string) string {
	if end := strings.Index(text, ". "); end >= 0 {
		return text[:end+1]
	}
	return text
}

// broke says whether a call's error means its session may serve no further
// call: the library closed it, or the service answered with a status it may
// have closed it over. An error of the protocol, a call that ran out of time
// and a connection that never reached the service leave it as it was.
func broke(err error, status int) bool {
	if err == nil {
		return false
	}
	var rpc *jsonrpc.Error
	if errors.As(err, &rpc) || errors.Is(err, context.DeadlineExceeded) {
		return false
	}
	return errors.Is(err, mcp.ErrConnectionClosed) || status >= 400
}

package middleware

import (
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"runtime"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/MathTrail/mathtrail-standalone/internal/apierror"
)

// panicKey holds, for the length of one request, what the request panicked
// with and where, so that the request's own line in the log says it. It is a
// type of its own, so that nothing else a request carries can be taken for it.
type panicKey struct{}

// maxStackDepth is how many frames of a panicked request's stack are read.
const maxStackDepth = 64

// ZapRecovery turns a panic of a handler into a plain JSON answer. One request
// must not take the process down: another child is in the middle of a task on
// the same instance.
//
// It stands next to the handlers and writes no line of its own. What the
// request panicked with, and where, goes into the request's line, which the
// request log writes when it sees the answer: one fault, one line, and the
// panicked request counted and traced like any other.
//
// A handler that panics with http.ErrAbortHandler, which is how net/http is
// told to end a response unfinished, is passed on as it asked: the server
// then drops the request silently, and nothing above takes it for an answer.
func ZapRecovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		recovering(c, func(recovered any) {
			c.Set(panicKey{}, []zap.Field{panicField(recovered), panicStack()})
		})
	}
}

// LastResort catches what ZapRecovery cannot: a panic in the middleware above
// it, whose request never reaches the request log. It writes the line itself,
// and answers and passes an abort on as ZapRecovery does.
func LastResort(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		recovering(c, func(recovered any) {
			logger.Error("panic",
				panicField(recovered),
				panicStack(),
				zap.String("method", c.Request.Method),
				zap.String("path", c.Request.URL.Path),
				zap.String("request_id", RequestIDFrom(c)),
			)
		})
	}
}

// recovering runs the handlers after it and catches a panic of theirs: an
// abort is passed on to the server, and anything else is written down by
// record and answered.
func recovering(c *gin.Context, record func(recovered any)) {
	defer func() {
		recovered := recover()
		switch {
		case recovered == nil:
			return
		case aborted(recovered):
			panic(http.ErrAbortHandler)
		}
		record(recovered)
		answerPanic(c)
	}()
	c.Next()
}

// aborted says whether a panic is a handler asking net/http to end its
// response unfinished, however it was wrapped on the way. It is passed on bare,
// because that is the one value the server knows to drop without a word.
func aborted(recovered any) bool {
	err, isError := recovered.(error)
	return isError && errors.Is(err, http.ErrAbortHandler)
}

// panicStack is the stack of the request that panicked, from the code that
// panicked down. What stands above that code is left off: the recovery that
// caught the panic, and the runtime's own frames — one for a panic called by
// name, a few more for a fault the runtime raised itself — so that the first
// line a reader sees is the line that failed.
func panicStack() zap.Field {
	callers := make([]uintptr, maxStackDepth)
	frames := runtime.CallersFrames(callers[:runtime.Callers(1, callers)])

	var stack strings.Builder
	panicking, reached := false, false
	for {
		frame, more := frames.Next()
		switch {
		case reached:
		case frame.Function == "runtime.gopanic":
			panicking = true
		case panicking && !strings.HasPrefix(frame.Function, "runtime."):
			reached = true
		}
		if reached {
			fmt.Fprintf(&stack, "%s\n\t%s:%d\n", frame.Function, frame.File, frame.Line)
		}
		if !more {
			return zap.String("stack", stack.String())
		}
	}
}

// answerPanic answers a request that panicked. A handler that panicked halfway
// through its answer has already sent a status and headers, and writing a
// second one corrupts what the client is reading: all that is left then is to
// stop.
func answerPanic(c *gin.Context) {
	if c.Writer.Written() {
		c.Abort()
		return
	}
	c.AbortWithStatusJSON(http.StatusInternalServerError, apierror.Response{
		Code:    apierror.CodeInternal,
		Message: "an unexpected error occurred",
	})
}

// panicField names what the request panicked with, and none of what it
// carried. A fault of the runtime is put in the runtime's own words, found
// inside whatever wrapped it, because those words hold types and numbers and
// no data. Anything else a handler panicked with is a value of ours or a
// library's, which may hold a task, an answer or a profile, so only its type is
// written: the stack beside it shows where it came from.
func panicField(recovered any) zap.Field {
	var fault runtime.Error
	if err, isError := recovered.(error); isError && errors.As(err, &fault) && ofTheRuntime(fault) {
		return zap.String("panic", fault.Error())
	}
	return zap.String("panic", fmt.Sprintf("a value of type %T", recovered))
}

// ofTheRuntime says whether a fault was made by the Go runtime itself. Any type
// can call itself a runtime error by having the method, and one that is not
// the runtime's may say anything at all.
func ofTheRuntime(fault runtime.Error) bool {
	kind := reflect.TypeOf(fault)
	for kind.Kind() == reflect.Pointer {
		kind = kind.Elem()
	}
	return kind.PkgPath() == "runtime"
}

package middleware

import (
	"context"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.opentelemetry.io/otel/metric/noop"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"

	"github.com/MathTrail/mathtrail-standalone/internal/telemetry"
)

// sampledKey holds, for the length of one request, whether its trace is being
// kept. See Tracing for why it has to be written down at all.
const sampledKey = "trace_sampled"

// Tracing is a request's whole relationship with the tracer: it joins the
// trace the request arrived in, and it delivers what the request left behind
// before the request is over.
//
// It is three handlers rather than one because of the order the framework runs
// them in, and they only work side by side, which is why they are built
// together and registered in one line:
//
//   - the first one delivers, and it is first because a handler registered
//     earlier finishes later — which is the only position from which the span
//     has already been closed and is therefore ready to send;
//   - the second one is the instrumentation, which opens the span and closes
//     it again;
//   - the third one writes down whether the trace is being kept, because the
//     instrumentation puts the request's original context back when it is
//     done: by the time the first handler runs, the request no longer
//     remembers the span it just had.
//
// Every request offers a delivery, and says whether its trace was kept: its
// spans are sent only then, and the measurements whenever they are due, which
// is how a quiet instance whose requests keep no trace still sends them.
func Tracing(tracers trace.TracerProvider, flush func(context.Context, bool) error, logger *zap.Logger) gin.HandlersChain {
	return gin.HandlersChain{
		deliver(flush, logger),
		otelgin.Middleware(telemetry.ServiceName,
			otelgin.WithTracerProvider(tracers),
			// The standard header and nothing else. Baggage would carry
			// whatever a caller put in it into every span we record.
			otelgin.WithPropagators(propagation.TraceContext{}),
			// The instrumentation's own measurements duplicate the ones this
			// service defines, under different names, and each series costs
			// bytes of a monthly allowance.
			otelgin.WithMeterProvider(noop.NewMeterProvider()),
			otelgin.WithGinFilter(worthATrace),
		),
		rememberSampling(),
	}
}

// deliver sends what the request left behind, once the span above it has
// closed. A collector that is slow or gone costs a line in the log, and time:
// the answer has been written but a small one is not sent until the handlers
// return, so the caller waits for this delivery when it sends anything — the
// spans of a request whose trace is kept, the measurements once an interval —
// up to the deadline the delivery brings. It happens here all the same,
// because once the answer is sent the instance may have no processor left to
// send anything with.
func deliver(flush func(context.Context, bool) error, logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Deferred, so that a request dropped on the way out — the abort a
		// handler asked for, passed through here to the server — still sends
		// what it recorded: its spans are the ones worth reading.
		defer deliverDue(c, flush, logger)
		c.Next()
	}
}

// deliverDue asks for what is due: the request's spans when its trace is kept,
// and whatever else the telemetry holds that is due to leave.
func deliverDue(c *gin.Context, flush func(context.Context, bool) error, logger *zap.Logger) {
	// A delivery that panics costs a line, not the request: its answer is
	// written, and a request dropped on the way out stays dropped.
	defer func() {
		if recovered := recover(); recovered != nil {
			logger.Error("telemetry_flush_panicked", panicField(recovered), zap.String("request_id", RequestIDFrom(c)))
		}
	}()
	// Detached from the request on purpose. A caller that hung up has cancelled
	// its context, and handing that to the delivery would end it before it
	// began — which costs not a line in a log but the spans themselves, because
	// a batch that failed to send is dropped rather than kept for the next
	// attempt. The delivery brings its own deadline.
	if err := flush(context.WithoutCancel(c.Request.Context()), c.GetBool(sampledKey)); err != nil {
		logger.Warn("telemetry_flush_failed",
			zap.Error(err),
			zap.String("request_id", RequestIDFrom(c)),
		)
	}
}

// rememberSampling records the decision while the span is still the request's,
// which is the only moment anything upstream can be told about it.
func rememberSampling() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set(sampledKey, trace.SpanContextFromContext(c.Request.Context()).IsSampled())
		c.Next()
	}
}

// worthATrace keeps the probes out. They arrive constantly, they say the same
// thing every time, and a trace of one is a span nobody will ever open.
func worthATrace(c *gin.Context) bool {
	return !isProbe(c.Request.URL.Path)
}

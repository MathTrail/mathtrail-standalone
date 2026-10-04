package telemetry

import (
	"context"
	"encoding/binary"
	"math"
	"math/rand/v2"
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

// tracesPerShare is how many traces a property draws for one share: enough
// that a fair draw lands within a twentieth of the share.
const tracesPerShare = 4000

// The share kept is the share configured whichever half of the identifier
// changes from one trace to the next. The platform's front end fills the right
// half once for a run of requests and the left half anew for each; a client
// that makes its identifiers as the newer convention of the standard asks does
// the opposite. What a request arrives saying of its trace is drawn as well,
// since it must not count either way.
func TestTheShareKeptIsTheShareConfigured(t *testing.T) {
	t.Parallel()

	properties := gopter.NewProperties(nil)
	for _, half := range []struct {
		name     string
		changing int
	}{
		{name: "the left half", changing: 0},
		{name: "the right half", changing: 8},
	} {
		properties.Property("with "+half.name+" alone changing, the share kept is the share configured", prop.ForAll(
			func(ratio float64, fixed uint64) bool {
				return math.Abs(keptOf(ratio, half.changing, fixed)-ratio) <= chance(ratio)
			},
			gen.Float64Range(0.02, 0.98), gen.UInt64(),
		))
	}
	properties.TestingRun(t)
}

// keptOf is the share of tracesPerShare traces the sampler keeps when one half
// of their identifiers is the same for all, fixed, and the half at changing is
// drawn anew for each, from a source the case decides.
func keptOf(ratio float64, changing int, fixed uint64) float64 {
	decide := sampler(ratio)
	draw := rand.New(rand.NewPCG(fixed, math.Float64bits(ratio)))
	kept := 0
	for range tracesPerShare {
		var id trace.TraceID
		binary.BigEndian.PutUint64(id[changing:changing+8], draw.Uint64())
		binary.BigEndian.PutUint64(id[8-changing:16-changing], fixed)
		var flags trace.TraceFlags
		if draw.IntN(2) == 1 {
			flags = trace.FlagsSampled
		}
		arrived := trace.ContextWithRemoteSpanContext(context.Background(), trace.NewSpanContext(trace.SpanContextConfig{
			TraceID:    id,
			SpanID:     trace.SpanID{1},
			TraceFlags: flags,
			Remote:     true,
		}))
		if decide.ShouldSample(sdktrace.SamplingParameters{ParentContext: arrived, TraceID: id}).Decision == sdktrace.RecordAndSample {
			kept++
		}
	}
	return float64(kept) / tracesPerShare
}

// chance is how far from the share configured a fair draw of tracesPerShare
// traces may land: six standard errors, which a fair sampler passes all but
// once in hundreds of millions of draws.
func chance(ratio float64) float64 {
	return 6 * math.Sqrt(ratio*(1-ratio)/tracesPerShare)
}

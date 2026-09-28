package middleware_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"

	"github.com/MathTrail/mathtrail-standalone/internal/apierror"
	"github.com/MathTrail/mathtrail-standalone/internal/ratelimit"
	"github.com/MathTrail/mathtrail-standalone/internal/ratelimit/ratelimittest"
	"github.com/MathTrail/mathtrail-standalone/internal/transport/http/middleware"
)

// What the handler behind the pace answers, so that a case can tell a request
// let through from one refused.
const (
	reachedHandler = "reached the handler"
	shownBusyPage  = "the busy page"
)

// paced is a router with two routes held to the ceilings given: one answering
// a refusal as a program is answered, one with a page for a person.
func paced(t *testing.T, ceilings ...ratelimit.Ceiling) (*gin.Engine, *observer.ObservedLogs) {
	t.Helper()

	core, lines := observer.New(zapcore.DebugLevel)
	logger := zap.New(core)
	page := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
		_, _ = w.Write([]byte(shownBusyPage))
	})
	reached := func(c *gin.Context) { c.String(http.StatusOK, reachedHandler) }

	router := gin.New()
	router.GET("/program", middleware.Limit(logger, "a-project", middleware.TooMany, ceilings...), reached)
	router.GET("/person", middleware.Limit(logger, "a-project", middleware.Page(page), ceilings...), reached)
	return router, lines
}

// A request past its pace is refused before the handler runs: 429, when to ask
// again in whole seconds, no cache to keep it, and the answer each route
// gives — the router's refusal to a program, the page to a person. The
// ceiling reached is written down once for the flood, as a warning.
func TestARequestPastItsPaceIsAnsweredAsItsRouteAnswers(t *testing.T) {
	t.Parallel()

	for _, route := range []string{"/program", "/person"} {
		router, lines := paced(t, ratelimit.Ceiling{Name: "ip_rate", Limiter: ratelimittest.Keyed(t, 3)})
		if rec := serve(t, router, http.MethodGet, route); rec.Body.String() != reachedHandler {
			t.Fatalf("%s, the first request: status %d, body %q, want the handler reached", route, rec.Code, rec.Body.String())
		}
		rec := serve(t, router, http.MethodGet, route)
		serve(t, router, http.MethodGet, route)

		if rec.Code != http.StatusTooManyRequests {
			t.Errorf("%s past the pace: status = %d, want %d", route, rec.Code, http.StatusTooManyRequests)
		}
		if got := rec.Header().Get("Retry-After"); got != "20" {
			t.Errorf("%s: Retry-After = %q, want 20: three a minute is one every twenty seconds", route, got)
		}
		if got := rec.Header().Get("Cache-Control"); got != "no-store" {
			t.Errorf("%s: Cache-Control = %q, want no-store", route, got)
		}
		wantAnsweredAs(t, route, rec.Body.Bytes())

		hits := lines.FilterMessage("limit_hit").FilterLevelExact(zapcore.WarnLevel).All()
		if len(hits) != 1 || hits[0].ContextMap()["limit"] != "ip_rate" {
			t.Errorf("%s: limit_hit warnings = %v, want one naming ip_rate for the flood", route, hits)
		}
	}
}

// wantAnsweredAs holds the body of a refusal to what its route answers with.
func wantAnsweredAs(t *testing.T, route string, body []byte) {
	t.Helper()

	if route == "/person" {
		if string(body) != shownBusyPage {
			t.Errorf("%s: body = %q, want the busy page", route, body)
		}
		return
	}
	var refusal apierror.Response
	if err := json.Unmarshal(body, &refusal); err != nil || refusal.Code != apierror.CodeTooManyRequests {
		t.Errorf("%s: body = %s, want the router's refusal with code %s", route, body, apierror.CodeTooManyRequests)
	}
}

// A request the later ceiling turns away is given back to the earlier one, and
// the line names the ceiling that refused.
func TestARequestTheLaterCeilingRefusesCostsTheEarlierNothing(t *testing.T) {
	t.Parallel()

	own := ratelimittest.Keyed(t, 3)
	everybody := ratelimittest.Shared(t, 3)
	everybody.Take("somebody else")
	router, lines := paced(t,
		ratelimit.Ceiling{Name: "ip_rate", Limiter: own},
		ratelimit.Ceiling{Name: "instance_rate", Limiter: everybody},
	)

	if rec := serve(t, router, http.MethodGet, "/program"); rec.Code != http.StatusTooManyRequests {
		t.Fatalf("status = %d, want %d: the instance's pace is spent", rec.Code, http.StatusTooManyRequests)
	}
	if hits := lines.FilterMessage("limit_hit").All(); len(hits) != 1 || hits[0].ContextMap()["limit"] != "instance_rate" {
		t.Errorf("limit_hit lines = %v, want one naming instance_rate", hits)
	}
	if got := own.Take("192.0.2.1"); !got.Allowed {
		t.Errorf("the address's own pace after = %+v, want the refused request given back", got)
	}
}

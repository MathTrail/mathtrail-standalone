package report

import (
	"slices"
	"strings"
	"testing"
)

// A call's time without Drive is its time less that of the calls to Drive made
// in its request — tied to it by the request's id and its trace together, so
// that two requests of one trace keep apart — and never less than nothing. A
// call to Drive whose line names no request is taken off no call.
func TestACallsTimeWithoutDriveIsItsOwn(t *testing.T) {
	t.Parallel()

	c := tallied(t,
		`{"message":"drive_call","duration_ms":150,"request_id":"r1","logging.googleapis.com/trace":"t1","logging.googleapis.com/spanId":"d1"}`,
		`{"message":"drive_call","duration_ms":180,"request_id":"r1","logging.googleapis.com/trace":"t1","logging.googleapis.com/spanId":"d2"}`,
		`{"message":"tool_call","tool":"get_progress","outcome":"ok","client":"claude","duration_ms":400,"request_id":"r1",`+
			`"logging.googleapis.com/trace":"t1","logging.googleapis.com/spanId":"c1"}`,
		`{"message":"drive_call","duration_ms":60,"request_id":"r2"}`,
		`{"message":"tool_call","tool":"read_task","outcome":"ok","client":"claude","duration_ms":90,"request_id":"r2"}`,
		`{"message":"drive_call","duration_ms":120,"request_id":"r3"}`,
		`{"message":"drive_call","duration_ms":120,"request_id":"r3"}`,
		`{"message":"tool_call","tool":"save_profile","outcome":"ok","client":"claude","duration_ms":200,"request_id":"r3"}`,
		`{"message":"tool_call","tool":"get_profile","outcome":"ok","client":"claude","duration_ms":7,"request_id":"r4"}`,
		// Two requests a client sent with one parent, which share the trace.
		`{"message":"drive_call","duration_ms":100,"request_id":"r5","logging.googleapis.com/trace":"shared"}`,
		`{"message":"tool_call","tool":"next_task","outcome":"ok","client":"claude","duration_ms":150,"request_id":"r5",`+
			`"logging.googleapis.com/trace":"shared"}`,
		`{"message":"drive_call","duration_ms":40,"request_id":"r6","logging.googleapis.com/trace":"shared"}`,
		`{"message":"tool_call","tool":"submit_answer","outcome":"ok","client":"claude","duration_ms":90,"request_id":"r6",`+
			`"logging.googleapis.com/trace":"shared"}`,
		// Lines that name no request at all.
		`{"message":"drive_call","duration_ms":500}`,
		`{"message":"tool_call","tool":"get_package","outcome":"ok","client":"claude","duration_ms":12}`,
	)
	for tool, want := range map[string]int64{
		"get_progress": 70, "read_task": 30, "save_profile": 0, "get_profile": 7, "next_task": 50, "submit_answer": 50,
		"get_package": 12,
	} {
		called := c.tools[toolOf{host: "claude", tool: tool}]
		if called == nil || !slices.Equal(called.own, []int64{want}) {
			t.Errorf("%s without Drive = %+v, want [%d]", tool, called, want)
		}
	}
}

// A request's trace is counted kept or dropped from its own line, and a
// delivery that failed is counted for the request it belongs to — kept,
// dropped, or one its own line does not tell: a line not read, or an id two
// requests of both kinds came with — and as late when it ran out of time. The
// share the telemetry was built to keep is the newest build's, whatever order
// the lines come in, and any other failure of the telemetry is counted beside.
func TestTracesAreCountedByWhetherTheirRequestKeptThem(t *testing.T) {
	t.Parallel()

	c := tallied(t,
		`{"message":"telemetry built","time":"2026-10-02T00:00:00Z","export":true,"sample_ratio":0.1}`,
		`{"message":"telemetry built","time":"2026-10-01T00:00:00Z","export":true,"sample_ratio":1}`,
		`{"message":"http_request","route":"/mcp","request_id":"k","logging.googleapis.com/trace_sampled":true}`,
		`{"message":"http_request","route":"/mcp","request_id":"d1","logging.googleapis.com/trace_sampled":false}`,
		`{"message":"http_request","route":"/mcp","request_id":"d2","logging.googleapis.com/trace_sampled":false}`,
		`{"message":"http_request","route":"/mcp","request_id":"twice","logging.googleapis.com/trace_sampled":true}`,
		`{"message":"http_request","route":"/mcp","request_id":"twice","logging.googleapis.com/trace_sampled":false}`,
		`{"message":"http_request","route":"/mcp","request_id":"local"}`,
		`{"message":"telemetry_flush_failed","error":"traces export: context deadline exceeded","request_id":"k"}`,
		`{"message":"telemetry_flush_failed","error":"metrics export: 503","request_id":"d1"}`,
		`{"message":"telemetry_flush_failed","error":"context deadline exceeded","request_id":"gone"}`,
		`{"message":"telemetry_flush_failed","error":"metrics export: 503","request_id":"twice"}`,
		`{"message":"telemetry_failed","error":"exporter export timeout"}`,
		`{"message":"telemetry_flush_panicked","panic":"a value of type string"}`,
	)
	tr := c.traces
	if tr.kept != 2 || tr.dropped != 3 || tr.configured == nil || *tr.configured != 0.1 || tr.broke != 2 {
		t.Errorf("traces = kept %d, dropped %d, configured %v, broke %d; want 2, 3, 0.1, 2",
			tr.kept, tr.dropped, tr.configured, tr.broke)
	}
	for whose, want := range map[string][2]int{ofKept: {1, 1}, ofDropped: {1, 0}, ofUntold: {2, 1}} {
		if got := [2]int{tr.failed[whose], tr.late[whose]}; got != want {
			t.Errorf("deliveries of %s failed and late = %v, want %v", whose, got, want)
		}
	}
}

// The busiest minute is counted by the clock, in the minute a call or a request
// arrived rather than the one its line was written in as it ended: an
// account's tool calls, an instance's requests at the MCP endpoint and
// elsewhere apart, and all the instances' requests at the endpoint together. A line with no time is left
// out of every minute, and one no reader named an instance for is none of the
// instances counted.
func TestTheBusiestMinuteIsCountedByTheClock(t *testing.T) {
	t.Parallel()

	c := tallied(t,
		`{"message":"tool_call","time":"2026-09-30T08:15:01Z","user":"u1"}`,
		`{"message":"tool_call","time":"2026-09-30T08:15:59Z","user":"u1"}`,
		`{"message":"tool_call","time":"2026-09-30T08:16:00Z","user":"u1"}`,
		`{"message":"tool_call","time":"2026-09-30T08:15:30Z","user":"u2"}`,
		`{"message":"tool_call","user":"u1"}`,
		// A call and a request that ended in the next minute, but arrived in this.
		`{"message":"tool_call","time":"2026-09-30T08:16:00.200Z","duration_ms":400,"user":"u1"}`,
		`{"message":"http_request","time":"2026-09-30T08:16:00.500Z","duration":2,"route":"/mcp","instance":"i-1"}`,
		`{"message":"http_request","time":"2026-09-30T08:15:01Z","route":"/mcp","instance":"i-1"}`,
		`{"message":"http_request","time":"2026-09-30T08:15:02Z","route":"/mcp","instance":"i-1"}`,
		`{"message":"http_request","time":"2026-09-30T08:15:03Z","route":"/mcp","instance":"i-2"}`,
		`{"message":"http_request","time":"2026-09-30T08:15:04Z","route":"/oauth/token","instance":"i-2"}`,
		`{"message":"http_request","time":"2026-09-30T08:17:00Z","route":"/mcp"}`,
	)
	b := &c.busy
	for what, counted := range map[string][2]int{
		"tool calls of one account":    {mostOf(b.ofAccount), 3},
		"MCP requests on one instance": {mostOf(b.mcp), 3},
		"other requests":               {mostOf(b.others), 1},
		"MCP requests on all":          {mostOf(acrossInstances(b.mcp)), 4},
		"instances":                    {b.instances(), 2},
		"tool calls":                   {b.toolCalls, 6},
		"MCP requests":                 {b.mcpRequests, 5},
	} {
		if counted[0] != counted[1] {
			t.Errorf("%s = %d, want %d", what, counted[0], counted[1])
		}
	}
}

// The build of the telemetry in force is the newest, whether it exports or
// not: a newer build that exports nothing leaves no share kept to tell.
func TestTheNewestBuildOfTheTelemetryIsTheOneInForce(t *testing.T) {
	t.Parallel()

	c := tallied(t,
		`{"message":"telemetry built","time":"2026-10-02T00:00:00Z","export":false}`,
		`{"message":"telemetry built","time":"2026-10-01T00:00:00Z","export":true,"sample_ratio":0.1}`,
	)
	if !c.traces.built || c.traces.configured != nil {
		t.Errorf("traces = built %v, configured %v; want built, with no share kept", c.traces.built, c.traces.configured)
	}
	if about := c.tracesAbout(); !strings.Contains(about, "exports nothing") {
		t.Errorf("tracesAbout() = %q, want it to say the newest build exports nothing", about)
	}
}

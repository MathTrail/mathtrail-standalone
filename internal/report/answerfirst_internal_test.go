package report

import "testing"

// A call that began on another instance while a write after the answer was
// under way read the file as it was before the write, and is counted beside
// the write; a write whose line names no instance, as a run on this machine
// writes it, has no other instance to count a call on.
func TestCallsElsewhereAreCountedOnlyForAWriteThatNamesItsInstance(t *testing.T) {
	t.Parallel()

	// It began 300 milliseconds into the write, on an instance of its own.
	const call = `{"message":"tool_call","tool":"get_progress","time":"2026-09-30T09:00:01.000Z","duration_ms":200,` +
		`"user":"u1","instance":"i-2"}`
	for _, test := range []struct {
		name, instance string
		want           int
	}{
		{"a write on an instance", `,"instance":"i-1"`, 1},
		{"a write of a run on this machine", "", 0},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			write := `{"message":"write_after_answer","tool":"submit_answer","outcome":"written",` +
				`"time":"2026-09-30T09:00:01.500Z","duration_ms":1000,"instructions_version":"v1","user":"u1"` +
				test.instance + `}`
			late := tallied(t, call, write).late[lateOf{version: "v1", tool: "submit_answer"}]
			if late == nil || late.elsewhere != test.want {
				t.Errorf("the writes after the answer = %+v, want %d calls elsewhere meanwhile", late, test.want)
			}
		})
	}
}

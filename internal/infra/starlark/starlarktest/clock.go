package starlarktest

import "time"

// Clock is the wall clock a test gives one run where the service gives limit:
// the same without the race detector, and ten times over with it. The detector
// costs the interpreter several times the time, and the machine the checks run
// on shares its few cores among the tests of every package at once: a program
// that takes a fraction of a second alone has taken more than two seconds
// there, and a run meant to end in an answer or a refusal has run out of the
// clock instead. The steps a run may take are not scaled. They hold a program
// to the service's limits on any machine, while the clock under the detector
// measures the machine. A test about the clock itself sets its own.
func Clock(limit time.Duration) time.Duration {
	return limit * raceSlowdown
}

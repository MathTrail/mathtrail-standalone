//go:build !race

package starlarktest

// raceSlowdown is one without the race detector: a test gives a run the
// service's own clock.
const raceSlowdown = 1

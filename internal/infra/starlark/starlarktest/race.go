//go:build race

package starlarktest

// raceSlowdown is how many times the service's clock a test gives a run under
// the race detector.
const raceSlowdown = 10

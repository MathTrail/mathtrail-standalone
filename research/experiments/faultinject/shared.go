package main

import "github.com/MathTrail/mathtrail-standalone/research/experiments/reviewing"

// The tasks the experiment hands in, and how it reviews them, are shared with
// the other experiments that hand tasks to the checks.
type (
	host       = reviewing.Host
	submission = reviewing.Submission
	verdict    = reviewing.Verdict
)

// language is what every case is written and judged in.
const language = reviewing.Language

package middleware

// probePaths answer a machine rather than a person. The platform asks for them
// constantly and the answer is the same every time.
var probePaths = map[string]struct{}{
	"/health": {},
}

// isProbe reports whether a path is one of those.
//
// Three middlewares ask, and all three leave a probe out for the same reason:
// there are far more of them than there are children, so a probe in the log
// buries the lines that mean something, a probe in a trace is a span nobody
// will ever open, and a probe in the counters makes "requests answered" a
// measure of how often the platform checked rather than of how much the
// service was used.
func isProbe(path string) bool {
	_, probe := probePaths[path]
	return probe
}

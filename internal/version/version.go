// Package version holds the build identity of the binary.
//
// The values are placeholders in a plain `go build` and are set at link time,
// so that a running instance can say exactly which commit it came from:
//
//	go build -ldflags "-X github.com/MathTrail/mathtrail-standalone/internal/version.Version=1.2.3 ..."
//
// The symbol path is spelled out because a wrong one is silent: the linker
// writes nothing, the build succeeds, and the binary reports "dev" forever.
// What catches that is the startup line and the health endpoint, both of which
// print these values on a real build.
package version

// Version is the release or tag the binary was built from.
var Version = "dev"

// Commit is the git commit the binary was built from.
var Commit = "unknown"

// Date is the build timestamp, in RFC 3339.
var Date = "unknown"

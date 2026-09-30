// Command report adds up the service's log into the numbers it is written
// for, reading the service's own lines on standard input and writing the
// report in Markdown on standard output.
package main

import (
	"fmt"
	"os"

	"github.com/MathTrail/mathtrail-standalone/internal/report"
)

func main() {
	if err := report.Run(os.Stdin, os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

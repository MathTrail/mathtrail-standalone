// Command sitecheck inspects a built site and refuses the ones that should not
// be published.
package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/MathTrail/mathtrail-standalone/internal/site/check"
)

func main() {
	log.SetFlags(0)
	log.SetPrefix("sitecheck: ")

	dir := flag.String("dir", "site/dist", "directory holding the built site")
	base := flag.String("base", "", "origin the site is published on, without a trailing slash")
	reference := flag.String("reference-locale", "en", "locale every other one must match")
	maxPage := flag.Int64("max-page-bytes", 300*1024, "cap on a page together with the files it loads")
	flag.Parse()

	if err := run(os.Stdout, *dir, check.Options{
		BaseURL:         *base,
		ReferenceLocale: *reference,
		MaxPageBytes:    *maxPage,
	}); err != nil {
		log.Fatal(err)
	}
}

// run checks the built site in dir and reports every finding before failing, so
// that one run tells a person everything there is to fix rather than the first
// thing.
func run(out io.Writer, dir string, opt check.Options) error {
	findings, err := check.Run(os.DirFS(dir), opt)
	if err != nil {
		return err
	}
	for _, finding := range findings {
		fmt.Fprintln(out, finding)
	}
	if len(findings) > 0 {
		return fmt.Errorf("%d problems in %s", len(findings), dir)
	}
	fmt.Fprintf(out, "%s is publishable\n", dir)
	return nil
}

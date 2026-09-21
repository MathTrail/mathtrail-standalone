// Command sitegen renders the site's sources into the directory a static host
// serves, and can serve that directory while somebody is working on it.
package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/MathTrail/mathtrail-standalone/internal/site/render"
)

// The preview server is a development tool, but a server with no header timeout
// is a server that can be held open, whatever it serves.
const readHeaderTimeout = 5 * time.Second

func main() {
	log.SetFlags(0)
	log.SetPrefix("sitegen: ")

	source := flag.String("source", "site", "directory holding content, templates and assets")
	out := flag.String("out", "site/dist", "directory the rendered site is written to")
	base := flag.String("base", "", "origin the site is published on, without a trailing slash")
	siteName := flag.String("name", "MathTrail", "product name for the masthead and sharing previews")
	reference := flag.String("reference-locale", "en", "locale every other one is measured against")
	addr := flag.String("serve", "", "address to serve the rendered site on, such as :8080")
	flag.Parse()

	if err := run(*source, *out, *addr, render.Options{
		BaseURL:         *base,
		SiteName:        *siteName,
		ReferenceLocale: *reference,
	}); err != nil {
		log.Fatal(err)
	}
}

// run renders the site and, when an address is given, serves what it rendered.
func run(source, out, addr string, opt render.Options) error {
	files, err := render.Build(os.DirFS(source), opt)
	if err != nil {
		return err
	}
	if err := render.Write(out, files); err != nil {
		return err
	}
	fmt.Printf("%d files in %s\n", len(files), out)

	if addr == "" {
		return nil
	}
	fmt.Printf("serving %s on http://localhost%s/\n", out, addr)
	server := &http.Server{
		Addr:              addr,
		Handler:           http.FileServer(http.Dir(out)),
		ReadHeaderTimeout: readHeaderTimeout,
	}
	return server.ListenAndServe()
}

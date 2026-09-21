package check_test

import (
	"os"
	"testing"
	"testing/fstest"

	"github.com/MathTrail/mathtrail-standalone/internal/site/check"
	"github.com/MathTrail/mathtrail-standalone/internal/site/render"
)

// TestTheSiteInThisRepositoryIsPublishable renders the real sources and holds
// them to the same rules the published site is held to, so a broken template, a
// dead link or a missing translation fails here rather than on the host.
func TestTheSiteInThisRepositoryIsPublishable(t *testing.T) {
	t.Parallel()

	const origin = "https://example.test"

	files, err := render.Build(os.DirFS("../../../site"), render.Options{
		BaseURL:         origin,
		SiteName:        "MathTrail",
		ReferenceLocale: "en",
	})
	if err != nil {
		t.Fatalf("Build() error = %v, want no error", err)
	}

	built := make(fstest.MapFS, len(files))
	for _, file := range files {
		built[file.Path] = &fstest.MapFile{Data: file.Data}
	}

	findings, err := check.Run(built, check.Options{
		BaseURL:         origin,
		ReferenceLocale: "en",
		MaxPageBytes:    300 * 1024,
	})
	if err != nil {
		t.Fatalf("Run() error = %v, want no error", err)
	}
	for _, finding := range findings {
		t.Errorf("the site is not publishable: %s", finding)
	}
}

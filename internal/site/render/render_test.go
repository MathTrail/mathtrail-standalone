package render_test

import (
	"sort"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/MathTrail/mathtrail-standalone/internal/site/render"
)

const pageTemplate = `<html lang="{{ .Lang }}" dir="{{ .Dir }}">` +
	`<title>{{ .Title }}</title>` +
	`<link rel="canonical" href="{{ .Canonical }}">` +
	`{{ range .Alternates }}<link rel="alternate" hreflang="{{ .Hreflang }}" href="{{ .URL }}">{{ end }}` +
	`{{ range .Languages }}<a href="{{ .URL }}">{{ .Name }}{{ if .Current }}*{{ end }}</a>{{ end }}` +
	`<nav aria-label="{{ .Text.language }}">` +
	`{{ range .Documents }}<a class="doc" href="{{ .URL }}">{{ .Label }}</a>{{ end }}</nav>` +
	`{{ .Body }}</html>`

const rootTemplate = `<html lang="{{ .Lang }}"><title>{{ .Title }}</title>` +
	`<link rel="canonical" href="{{ .Canonical }}">` +
	`{{ range .Alternates }}<link rel="alternate" hreflang="{{ .Hreflang }}" href="{{ .URL }}">{{ end }}` +
	`<nav aria-label="{{ .Text.language }}"></nav></html>`

func options() render.Options {
	return render.Options{
		BaseURL:         "https://example.test",
		SiteName:        "Example",
		ReferenceLocale: "en",
	}
}

// source builds a site with two locales, where only the reference locale has
// the second page — the shape every cross-linking rule has to survive.
func source() fstest.MapFS {
	text := func(title string) *fstest.MapFile {
		return &fstest.MapFile{Data: []byte("---\ntitle: " + title + "\ndescription: About " + title + "\n---\n\n# " + title + "\n")}
	}
	return fstest.MapFS{
		"content/en/index.md":   text("Home"),
		"content/en/privacy.md": text("Privacy"),
		"content/ru/index.md":   text("Главная"),
		"strings/en.json":       &fstest.MapFile{Data: []byte(`{"language":"Language","privacy":"Privacy"}`)},
		"strings/ru.json":       &fstest.MapFile{Data: []byte(`{"language":"Язык","privacy":"Приватность"}`)},
		"templates/page.html":   &fstest.MapFile{Data: []byte(pageTemplate)},
		"templates/root.html":   &fstest.MapFile{Data: []byte(rootTemplate)},
		"assets/style.css":      &fstest.MapFile{Data: []byte("body{color:#000}")},
	}
}

func build(t *testing.T, fsys fstest.MapFS, opt render.Options) map[string]string {
	t.Helper()

	files, err := render.Build(fsys, opt)
	if err != nil {
		t.Fatalf("Build() error = %v, want no error", err)
	}
	out := make(map[string]string, len(files))
	for _, file := range files {
		out[file.Path] = string(file.Data)
	}
	return out
}

func TestBuildWritesEveryFileAHostNeeds(t *testing.T) {
	t.Parallel()

	got := build(t, source(), options())

	want := []string{
		".nojekyll",
		"CNAME",
		"assets/style.css",
		"en/index.html",
		"en/privacy/index.html",
		"index.html",
		"robots.txt",
		"ru/index.html",
		"sitemap.xml",
	}
	var paths []string
	for path := range got {
		paths = append(paths, path)
	}
	sort.Strings(paths)

	if strings.Join(paths, " ") != strings.Join(want, " ") {
		t.Errorf("built files = %v, want %v", paths, want)
	}
}

func TestBuildIsByteIdenticalTwice(t *testing.T) {
	t.Parallel()

	first := build(t, source(), options())
	second := build(t, source(), options())

	for path, data := range first {
		if second[path] != data {
			t.Errorf("%s differs between two builds of the same sources", path)
		}
	}
}

func TestPageCarriesItsOwnAddressAndItsTranslations(t *testing.T) {
	t.Parallel()

	got := build(t, source(), options())

	tests := []struct {
		name string
		path string
		want []string
	}{
		{
			name: "a locale front page points at every language and at the apex",
			path: "en/index.html",
			want: []string{
				`lang="en"`,
				`<link rel="canonical" href="https://example.test/en/">`,
				`<link rel="alternate" hreflang="en" href="https://example.test/en/">`,
				`<link rel="alternate" hreflang="ru" href="https://example.test/ru/">`,
				`<link rel="alternate" hreflang="x-default" href="https://example.test/">`,
			},
		},
		{
			name: "a page only one locale has falls back to that locale",
			path: "en/privacy/index.html",
			want: []string{
				`<link rel="canonical" href="https://example.test/en/privacy/">`,
				`<link rel="alternate" hreflang="x-default" href="https://example.test/en/privacy/">`,
			},
		},
		{
			name: "the apex is a page of its own, in the reference language",
			path: "index.html",
			want: []string{
				`lang="en"`,
				`<link rel="canonical" href="https://example.test/">`,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			for _, want := range tt.want {
				if !strings.Contains(got[tt.path], want) {
					t.Errorf("%s does not contain %q\ngot: %s", tt.path, want, got[tt.path])
				}
			}
		})
	}
}

func TestLanguageSwitcherSkipsLocalesWithoutThePage(t *testing.T) {
	t.Parallel()

	got := build(t, source(), options())

	if strings.Contains(got["en/privacy/index.html"], `href="/ru/privacy/"`) {
		t.Errorf("the switcher offers a Russian privacy page the site does not have:\n%s", got["en/privacy/index.html"])
	}
	if !strings.Contains(got["en/index.html"], `href="/ru/"`) {
		t.Errorf("the switcher does not offer Russian on a page that has it:\n%s", got["en/index.html"])
	}
}

func TestMetadataFilesFollowTheBaseURL(t *testing.T) {
	t.Parallel()

	got := build(t, source(), options())

	if got["CNAME"] != "example.test\n" {
		t.Errorf("CNAME = %q, want %q", got["CNAME"], "example.test\n")
	}
	if !strings.Contains(got["robots.txt"], "Sitemap: https://example.test/sitemap.xml") {
		t.Errorf("robots.txt does not name the sitemap:\n%s", got["robots.txt"])
	}
	for _, want := range []string{
		"<loc>https://example.test/</loc>",
		"<loc>https://example.test/en/privacy/</loc>",
		`<xhtml:link rel="alternate" hreflang="ru" href="https://example.test/ru/"/>`,
	} {
		if !strings.Contains(got["sitemap.xml"], want) {
			t.Errorf("sitemap.xml does not contain %q\ngot: %s", want, got["sitemap.xml"])
		}
	}
}

func TestBuildRefusesWhatItCannotPublishCorrectly(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		mutate  func(fstest.MapFS, *render.Options)
		wantErr string
	}{
		{
			name:    "a base URL with a trailing slash would double every address",
			mutate:  func(_ fstest.MapFS, opt *render.Options) { opt.BaseURL = "https://example.test/" },
			wantErr: "ends in a slash",
		},
		{
			name:    "a base URL with a path is not an origin",
			mutate:  func(_ fstest.MapFS, opt *render.Options) { opt.BaseURL = "https://example.test/site" },
			wantErr: "more than a scheme and a host",
		},
		{
			name:    "a missing base URL leaves every canonical wrong",
			mutate:  func(_ fstest.MapFS, opt *render.Options) { opt.BaseURL = "" },
			wantErr: "base URL is empty",
		},
		{
			name:    "a reference locale with no content cannot be fallen back to",
			mutate:  func(_ fstest.MapFS, opt *render.Options) { opt.ReferenceLocale = "de" },
			wantErr: `reference locale "de" has no content`,
		},
		{
			name: "a text with no front matter has no title to render",
			mutate: func(fsys fstest.MapFS, _ *render.Options) {
				fsys["content/en/index.md"] = &fstest.MapFile{Data: []byte("# Home\n")}
			},
			wantErr: "front matter must open",
		},
		{
			name: "a locale with no text would publish an empty language",
			mutate: func(fsys fstest.MapFS, _ *render.Options) {
				delete(fsys, "content/ru/index.md")
				fsys["content/ru/notes.txt"] = &fstest.MapFile{Data: []byte("hello")}
			},
			wantErr: `locale "ru" holds no text`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			fsys, opt := source(), options()
			tt.mutate(fsys, &opt)

			_, err := render.Build(fsys, opt)
			if err == nil {
				t.Fatalf("Build() error = nil, want one containing %q", tt.wantErr)
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("Build() error = %q, want one containing %q", err, tt.wantErr)
			}
		})
	}
}

package check_test

import (
	"strings"
	"testing"
	"testing/fstest"

	"github.com/MathTrail/mathtrail-standalone/internal/site/check"
)

const base = "https://example.test"

func options() check.Options {
	return check.Options{BaseURL: base, ReferenceLocale: "en", MaxPageBytes: 4096}
}

// page writes a page that breaks none of the rules, so that a test can break
// exactly one of them and see only that.
func page(address string, alternates ...string) string {
	var b strings.Builder
	b.WriteString(`<!DOCTYPE html><html lang="` + langOf(address) + `" dir="ltr"><head>`)
	b.WriteString(`<title>Title</title><meta name="description" content="Description">`)
	b.WriteString(`<link rel="canonical" href="` + base + address + `">`)
	for _, alt := range alternates {
		hreflang, href, _ := strings.Cut(alt, " ")
		b.WriteString(`<link rel="alternate" hreflang="` + hreflang + `" href="` + base + href + `">`)
	}
	b.WriteString(`<link rel="stylesheet" href="/assets/style.css">`)
	b.WriteString(`</head><body><a href="https://github.com/example">Source</a></body></html>`)
	return b.String()
}

// langOf is the language a page at this address is expected to declare.
func langOf(address string) string {
	if address == "/" {
		return "en"
	}
	return strings.Split(strings.Trim(address, "/"), "/")[0]
}

// site is a small but complete site: an apex, two locales, one stylesheet.
func site() fstest.MapFS {
	home := []string{"en /en/", "ru /ru/", "x-default /"}
	return fstest.MapFS{
		"index.html":       &fstest.MapFile{Data: []byte(page("/", home...))},
		"en/index.html":    &fstest.MapFile{Data: []byte(page("/en/", home...))},
		"ru/index.html":    &fstest.MapFile{Data: []byte(page("/ru/", home...))},
		"assets/style.css": &fstest.MapFile{Data: []byte("body{color:#000}")},
	}
}

func run(t *testing.T, fsys fstest.MapFS, opt check.Options) []check.Finding {
	t.Helper()

	findings, err := check.Run(fsys, opt)
	if err != nil {
		t.Fatalf("Run() error = %v, want no error", err)
	}
	return findings
}

func TestRunPassesASiteWithNothingWrong(t *testing.T) {
	t.Parallel()

	if findings := run(t, site(), options()); len(findings) != 0 {
		t.Errorf("Run() = %v, want no findings", findings)
	}
}

func TestRunNamesWhatIsBroken(t *testing.T) {
	t.Parallel()

	home := []string{"en /en/", "ru /ru/", "x-default /"}

	tests := []struct {
		name     string
		mutate   func(fstest.MapFS, *check.Options)
		wantRule string
		wantIn   string
	}{
		{
			name: "a link to a page that was never built",
			mutate: func(fsys fstest.MapFS, _ *check.Options) {
				fsys["en/index.html"] = &fstest.MapFile{
					Data: []byte(strings.Replace(page("/en/", home...), "</body>", `<a href="/en/privacy/">Privacy</a></body>`, 1)),
				}
			},
			wantRule: check.RuleLink,
			wantIn:   "en/privacy/index.html",
		},
		{
			name: "a stylesheet loaded from somebody else's domain",
			mutate: func(fsys fstest.MapFS, _ *check.Options) {
				fsys["en/index.html"] = &fstest.MapFile{
					Data: []byte(strings.Replace(page("/en/", home...), `href="/assets/style.css"`, `href="https://cdn.example.com/style.css"`, 1)),
				}
			},
			wantRule: check.RuleExternal,
			wantIn:   "another origin",
		},
		{
			name: "a font pulled in from inside the stylesheet",
			mutate: func(fsys fstest.MapFS, _ *check.Options) {
				fsys["assets/style.css"] = &fstest.MapFile{Data: []byte("@import url(https://fonts.example.com/x.css);")}
			},
			wantRule: check.RuleExternal,
			wantIn:   "stylesheet mentions",
		},
		{
			name: "a page with no description",
			mutate: func(fsys fstest.MapFS, _ *check.Options) {
				fsys["ru/index.html"] = &fstest.MapFile{
					Data: []byte(strings.Replace(page("/ru/", home...), `<meta name="description" content="Description">`, "", 1)),
				}
			},
			wantRule: check.RuleHead,
			wantIn:   "no description",
		},
		{
			name: "a canonical that points at another page",
			mutate: func(fsys fstest.MapFS, _ *check.Options) {
				fsys["ru/index.html"] = &fstest.MapFile{
					Data: []byte(strings.Replace(page("/ru/", home...), base+"/ru/\">", base+"/en/\">", 1)),
				}
			},
			wantRule: check.RuleHead,
			wantIn:   "canonical is",
		},
		{
			name: "a lang that disagrees with the address",
			mutate: func(fsys fstest.MapFS, _ *check.Options) {
				fsys["ru/index.html"] = &fstest.MapFile{
					Data: []byte(strings.Replace(page("/ru/", home...), `lang="ru"`, `lang="en"`, 1)),
				}
			},
			wantRule: check.RuleHead,
			wantIn:   "disagrees with the locale",
		},
		{
			name: "a translation the page never points at",
			mutate: func(fsys fstest.MapFS, _ *check.Options) {
				fsys["en/index.html"] = &fstest.MapFile{Data: []byte(page("/en/", "en /en/", "x-default /"))}
			},
			wantRule: check.RuleHead,
			wantIn:   `no alternate for "ru"`,
		},
		{
			name: "a locale that lost a page the reference locale has",
			mutate: func(fsys fstest.MapFS, _ *check.Options) {
				fsys["en/privacy/index.html"] = &fstest.MapFile{
					Data: []byte(page("/en/privacy/", "en /en/privacy/", "x-default /en/privacy/")),
				}
			},
			wantRule: check.RuleTranslation,
			wantIn:   `locale "ru" is missing`,
		},
		{
			name: "a page that costs the reader more than the budget",
			mutate: func(_ fstest.MapFS, opt *check.Options) {
				opt.MaxPageBytes = 64
			},
			wantRule: check.RuleWeight,
			wantIn:   "the budget is 64",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			fsys, opt := site(), options()
			tt.mutate(fsys, &opt)

			findings := run(t, fsys, opt)
			if len(findings) == 0 {
				t.Fatalf("Run() = no findings, want one of rule %q containing %q", tt.wantRule, tt.wantIn)
			}
			for _, finding := range findings {
				if finding.Rule == tt.wantRule && strings.Contains(finding.String(), tt.wantIn) {
					return
				}
			}
			t.Errorf("Run() = %v, want a finding of rule %q containing %q", findings, tt.wantRule, tt.wantIn)
		})
	}
}

func TestRunRefusesToJudgeASiteItCannotRead(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		fsys    fstest.MapFS
		opt     check.Options
		wantErr string
	}{
		{
			name:    "no base URL to measure addresses against",
			fsys:    site(),
			opt:     check.Options{ReferenceLocale: "en"},
			wantErr: "base URL is empty",
		},
		{
			name:    "an empty directory is not a site that passed",
			fsys:    fstest.MapFS{},
			opt:     options(),
			wantErr: "empty",
		},
		{
			name:    "a reference locale that was never built",
			fsys:    site(),
			opt:     check.Options{BaseURL: base, ReferenceLocale: "de"},
			wantErr: `reference locale "de" has no front page`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, err := check.Run(tt.fsys, tt.opt)
			if err == nil {
				t.Fatalf("Run() error = nil, want one containing %q", tt.wantErr)
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("Run() error = %q, want one containing %q", err, tt.wantErr)
			}
		})
	}
}

func FuzzRun(f *testing.F) {
	f.Add(page("/en/", "en /en/", "x-default /"))
	f.Add("<html><body><a href=")
	f.Add("")

	f.Fuzz(func(t *testing.T, document string) {
		fsys := site()
		fsys["en/index.html"] = &fstest.MapFile{Data: []byte(document)}

		// A page the model or a future builder produced may be anything at all;
		// the only promise is that judging it never takes the process down.
		if _, err := check.Run(fsys, options()); err != nil {
			t.Errorf("Run() error = %v, want findings rather than an error", err)
		}
	})
}

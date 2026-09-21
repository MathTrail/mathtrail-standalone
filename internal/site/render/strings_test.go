package render_test

import (
	"strings"
	"testing"
	"testing/fstest"

	"github.com/MathTrail/mathtrail-standalone/internal/site/render"
)

func TestFooterSpeaksTheLanguageOfThePage(t *testing.T) {
	t.Parallel()

	got := build(t, source(), options())

	tests := []struct {
		name string
		path string
		want string
	}{
		{
			name: "English page, English label",
			path: "en/index.html",
			want: `<a class="doc" href="/en/privacy/">Privacy</a>`,
		},
		{
			name: "Russian page, Russian label pointing at the Russian page",
			path: "ru/index.html",
			want: `<nav aria-label="Язык">`,
		},
		{
			name: "the apex speaks the reference language",
			path: "index.html",
			want: `Language`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if !strings.Contains(got[tt.path], tt.want) {
				t.Errorf("%s does not contain %q\ngot: %s", tt.path, tt.want, got[tt.path])
			}
		})
	}
}

// A locale that does not have a page must not be offered it in the footer: a
// link into another language is worse than one link fewer.
func TestFooterOffersOnlyThePagesThisLocaleHas(t *testing.T) {
	t.Parallel()

	got := build(t, source(), options())

	if strings.Contains(got["ru/index.html"], `class="doc"`) {
		t.Errorf("the Russian footer offers a page only English has:\n%s", got["ru/index.html"])
	}
}

func TestBuildRefusesTextItCannotPublishCorrectly(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		mutate  func(fstest.MapFS)
		wantErr string
	}{
		{
			name: "a locale missing a name the reference locale has",
			mutate: func(fsys fstest.MapFS) {
				fsys["strings/ru.json"] = &fstest.MapFile{Data: []byte(`{"language":"Язык"}`)}
			},
			wantErr: `ru.json has no "privacy"`,
		},
		{
			name: "a locale carrying a name nothing renders",
			mutate: func(fsys fstest.MapFS) {
				fsys["strings/ru.json"] = &fstest.MapFile{
					Data: []byte(`{"language":"Язык","privacy":"Приватность","extra":"?"}`),
				}
			},
			wantErr: `ru.json has "extra"`,
		},
		{
			name: "a page with no label in the reference locale",
			mutate: func(fsys fstest.MapFS) {
				fsys["strings/en.json"] = &fstest.MapFile{Data: []byte(`{"language":"Language","privacy":"Privacy"}`)}
				fsys["strings/ru.json"] = &fstest.MapFile{Data: []byte(`{"language":"Язык","privacy":"Приватность"}`)}
				fsys["content/en/terms.md"] = &fstest.MapFile{
					Data: []byte("---\ntitle: Terms\ndescription: About Terms\n---\n\n# Terms\n"),
				}
			},
			wantErr: `no label for the page "terms"`,
		},
		{
			name: "a locale with no text file at all",
			mutate: func(fsys fstest.MapFS) {
				delete(fsys, "strings/ru.json")
			},
			wantErr: "read strings/ru.json",
		},
		{
			name: "a text file that is not JSON",
			mutate: func(fsys fstest.MapFS) {
				fsys["strings/ru.json"] = &fstest.MapFile{Data: []byte("язык: Язык")}
			},
			wantErr: "parse strings/ru.json",
		},
		{
			name: "an empty text file",
			mutate: func(fsys fstest.MapFS) {
				fsys["strings/ru.json"] = &fstest.MapFile{Data: []byte(`{}`)}
			},
			wantErr: "holds no text",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			fsys := source()
			tt.mutate(fsys)

			_, err := render.Build(fsys, options())
			if err == nil {
				t.Fatalf("Build() error = nil, want one containing %q", tt.wantErr)
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("Build() error = %q, want one containing %q", err, tt.wantErr)
			}
		})
	}
}

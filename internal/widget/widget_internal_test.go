package widget

import (
	"strings"
	"testing"
	"testing/fstest"
)

// placeholderMarker is how a check that reads a served page tells the
// placeholder from the widget: it survives the escaping a page gets inside a
// JSON answer.
const placeholderMarker = "widget-placeholder"

func TestPageChoosesTheBuildWhenThereIsOne(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name  string
		files fstest.MapFS
		want  string
	}{
		{
			name: "built and placeholder",
			files: fstest.MapFS{
				built:       {Data: []byte("the widget")},
				placeholder: {Data: []byte("the placeholder")},
			},
			want: "the widget",
		},
		{
			name:  "placeholder only",
			files: fstest.MapFS{placeholder: {Data: []byte("the placeholder")}},
			want:  "the placeholder",
		},
		{
			name:  "neither",
			files: fstest.MapFS{},
			want:  "",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if got := page(tc.files); got != tc.want {
				t.Errorf("page() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestThePlaceholderSaysWhatItIs(t *testing.T) {
	t.Parallel()

	html, err := pages.ReadFile(placeholder)
	if err != nil {
		t.Fatalf("reading the embedded placeholder: %v", err)
	}
	if !strings.Contains(string(html), placeholderMarker) {
		t.Errorf("the placeholder carries no %q, want the marker a check of a served page looks for", placeholderMarker)
	}
	if strings.Contains(string(html), "ui/initialize") {
		t.Error("the placeholder names ui/initialize, want that word only in a built widget, which speaks the protocol")
	}
}

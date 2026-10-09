package checks

import (
	"testing"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/picture"
)

func TestALessonsLanguageSaysHowItWritesDecimals(t *testing.T) {
	t.Parallel()
	tests := []struct {
		lesson string
		want   picture.Decimals
	}{
		{"en", picture.Point},
		{"en-US", picture.Point},
		{"en-ZA", picture.Comma},
		{"ru", picture.Comma},
		{"de", picture.Comma},
		{"de-CH", picture.Point},
		{"fr", picture.Comma},
		{"es", picture.Comma},
		{"es-MX", picture.Point},
		{"pt-BR", picture.Comma},
		{"uk", picture.Comma},
		{"zh-Hans", picture.Point},
		{"ja", picture.Point},
		{"hi", picture.Point},
		{"ar", picture.Point},
		{"fa", picture.Point},
		{"und", picture.Point},
		{"", picture.Point},
		{"not a language", picture.Point},
	}
	for _, test := range tests {
		if got := decimalsOf(test.lesson); got != test.want {
			t.Errorf("decimalsOf(%q) = %v, want %v", test.lesson, got, test.want)
		}
	}
}

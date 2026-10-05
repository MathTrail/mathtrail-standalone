package widget_test

import (
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/MathTrail/mathtrail-standalone/internal/widget"
)

// Whether or not a build has run, the package has a page to serve.
func TestPageIsAnHTMLDocument(t *testing.T) {
	t.Parallel()

	page := widget.Page()
	if !strings.HasPrefix(strings.ToLower(page), "<!doctype html>") {
		t.Errorf("Page() starts %q, want an HTML document", page[:min(len(page), 40)])
	}
}

// The tokens are custom properties for both themes: the light ones at the
// root, and the dark ones for a viewer who prefers them.
func TestTokensDefineBothThemes(t *testing.T) {
	t.Parallel()

	tokens := widget.Tokens()
	for _, want := range []string{":root", "--surface:", "--text:", "--accent:", "--font-sans:", "prefers-color-scheme: dark", `[data-theme="dark"]`} {
		if !strings.Contains(tokens, want) {
			t.Errorf("Tokens() has no %q, want the design tokens of both themes", want)
		}
	}
}

// The ramp a topic's rank is drawn in stands out from the card at 3 to 1 at
// least in either theme, as a part of a picture a reader needs must; each step
// stands apart from the one before on the colour wheel, so that two ranks a
// step apart never look alike; and it is the same for a viewer who prefers the
// dark theme as for a host that names it.
func TestTheRampStandsOutFromTheCard(t *testing.T) {
	t.Parallel()

	tokens := widget.Tokens()
	light := tokensOf(t, tokens, `:root, [data-theme="light"] {`)
	dark := tokensOf(t, tokens, `[data-theme="dark"] {`)
	preferred := tokensOf(t, tokens, `:root:not([data-theme="light"]) {`)
	for step := 1; step <= 5; step++ {
		name := fmt.Sprintf("ramp-%d", step)
		for _, theme := range []struct {
			name   string
			tokens map[string]string
		}{{"light", light}, {"dark", dark}} {
			colour, surface := theme.tokens[name], theme.tokens["surface"]
			if got := contrast(t, colour, surface); got < 3 {
				t.Errorf("--%s in the %s theme is %q, %.2f to 1 against the card %s, want 3 to 1 at least",
					name, theme.name, colour, got, surface)
			}
			if step == 1 {
				continue
			}
			before := theme.tokens[fmt.Sprintf("ramp-%d", step-1)]
			if apart := hue(t, colour) - hue(t, before); apart < 8 {
				t.Errorf("--%s in the %s theme is %.1f° on from the step before it, %q to %q, want 8° at least",
					name, theme.name, apart, before, colour)
			}
		}
		if preferred[name] != dark[name] {
			t.Errorf("--%s is %q for a viewer who prefers the dark theme and %q for a host that names it, want one",
				name, preferred[name], dark[name])
		}
	}
}

// How a rank moved is written in words beside the course as well as drawn in
// stripes on it, so a gain and a step back stand out from the card as text
// must, at 4.5 to 1 at least, and from the course's empty track as a part of a
// picture must, at 3 to 1 — in either theme, and the same for a viewer who
// prefers the dark theme as for a host that names it.
func TestTheMovesStandOutFromTheCardAndTheCourse(t *testing.T) {
	t.Parallel()

	tokens := widget.Tokens()
	light := tokensOf(t, tokens, `:root, [data-theme="light"] {`)
	dark := tokensOf(t, tokens, `[data-theme="dark"] {`)
	preferred := tokensOf(t, tokens, `:root:not([data-theme="light"]) {`)
	for _, name := range []string{"gain", "loss"} {
		for _, theme := range []struct {
			name   string
			tokens map[string]string
		}{{"light", light}, {"dark", dark}} {
			colour := theme.tokens[name]
			for _, against := range []struct {
				what, token string
				least       float64
			}{{"the card", "surface", 4.5}, {"the course's track", "border", 3}} {
				if got := contrast(t, colour, theme.tokens[against.token]); got < against.least {
					t.Errorf("--%s in the %s theme is %q, %.2f to 1 against %s %s, want %.1f to 1 at least",
						name, theme.name, colour, got, against.what, theme.tokens[against.token], against.least)
				}
			}
		}
		if preferred[name] != dark[name] {
			t.Errorf("--%s is %q for a viewer who prefers the dark theme and %q for a host that names it, want one",
				name, preferred[name], dark[name])
		}
	}
}

// Each theme shows one of the logo's two drawings: the light theme the one on
// a white tile, and the dark theme the site's icon itself, alike for a host
// that names it and for a viewer who prefers it.
func TestEachThemeShowsOneDrawingOfTheLogo(t *testing.T) {
	t.Parallel()

	tokens := widget.Tokens()
	for _, theme := range []struct {
		name, opening, shown, hidden string
	}{
		{"light", `:root, [data-theme="light"] {`, "mark-light", "mark-dark"},
		{"dark", `[data-theme="dark"] {`, "mark-dark", "mark-light"},
		{"preferred dark", `:root:not([data-theme="light"]) {`, "mark-dark", "mark-light"},
	} {
		t.Run(theme.name, func(t *testing.T) {
			t.Parallel()

			properties := tokensOf(t, tokens, theme.opening)
			if got := properties[theme.shown]; got != "block" {
				t.Errorf("--%s in the %s theme is %q, want block", theme.shown, theme.name, got)
			}
			if got := properties[theme.hidden]; got != "none" {
				t.Errorf("--%s in the %s theme is %q, want none", theme.hidden, theme.name, got)
			}
		})
	}
}

// declaration is one custom property of the tokens: its name and its value.
var declaration = regexp.MustCompile(`--([a-z0-9-]+):\s*([^;]+);`)

// tokensOf are the custom properties of the block of tokens that opens with
// opening, by their names.
func tokensOf(t *testing.T, tokens, opening string) map[string]string {
	t.Helper()

	_, body, found := strings.Cut(tokens, opening)
	if !found {
		t.Fatalf("the tokens have no block %q", opening)
	}
	body, _, _ = strings.Cut(body, "}")
	properties := map[string]string{}
	for _, match := range declaration.FindAllStringSubmatch(body, -1) {
		properties[match[1]] = strings.TrimSpace(match[2])
	}
	return properties
}

// contrast is how far apart two colours written #rrggbb stand, from 1 for one
// colour to 21 for black on white.
func contrast(t *testing.T, one, other string) float64 {
	t.Helper()

	lighter, darker := luminance(t, one), luminance(t, other)
	if darker > lighter {
		lighter, darker = darker, lighter
	}
	return (lighter + 0.05) / (darker + 0.05)
}

// luminance is how light a colour written #rrggbb is, as a contrast is worked
// out from it.
func luminance(t *testing.T, colour string) float64 {
	t.Helper()

	red, green, blue := channels(t, colour)
	linear := func(value float64) float64 {
		if value <= 0.03928 {
			return value / 12.92
		}
		return math.Pow((value+0.055)/1.055, 2.4)
	}
	return 0.2126*linear(red) + 0.7152*linear(green) + 0.0722*linear(blue)
}

// hue is where a colour written #rrggbb stands on the colour wheel, in degrees
// from red.
func hue(t *testing.T, colour string) float64 {
	t.Helper()

	red, green, blue := channels(t, colour)
	high, low := max(red, green, blue), min(red, green, blue)
	if high == low {
		return 0
	}
	spread := high - low
	var sixth float64
	switch high {
	case red:
		sixth = math.Mod((green-blue)/spread, 6)
	case green:
		sixth = (blue-red)/spread + 2
	default:
		sixth = (red-green)/spread + 4
	}
	return math.Mod(sixth*60+360, 360)
}

// channels are the red, green and blue of a colour written #rrggbb, each from
// 0 to 1.
func channels(t *testing.T, colour string) (red, green, blue float64) {
	t.Helper()

	if len(colour) != 7 || colour[0] != '#' {
		t.Fatalf("the colour %q is not written #rrggbb", colour)
	}
	var values [3]float64
	for at := range values {
		channel, err := strconv.ParseUint(colour[1+2*at:3+2*at], 16, 8)
		if err != nil {
			t.Fatalf("the colour %q is not written #rrggbb: %v", colour, err)
		}
		values[at] = float64(channel) / 255
	}
	return values[0], values[1], values[2]
}

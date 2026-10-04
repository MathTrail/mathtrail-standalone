package site_test

import (
	"encoding/json"
	"maps"
	"os"
	"slices"
	"testing"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/checks"
)

// A card a page draws is a card of the widget, and its drawing is one the
// service would let through: no wider and no taller than a phone's card
// holds, and made only of characters every monospaced font draws alike. A
// drawing written in the letters of one language, or wider than the card, is
// a card no chat would ever show, and it stops here rather than on the page.
func TestEveryCardOnTheSiteHasADrawingTheChecksAccept(t *testing.T) {
	t.Parallel()

	drawings := cardDrawingsOf(t)
	for _, page := range slices.Sorted(maps.Keys(drawings)) {
		t.Run(page, func(t *testing.T) {
			t.Parallel()

			drawing := drawings[page]
			if problems := checks.DrawingFormat(drawing, checks.DefaultDrawingLimits()); len(problems) > 0 {
				t.Errorf("DrawingFormat(%q) = %v, want no problems", drawing, problems)
			}
		})
	}
}

// cardDrawingsOf are the drawings of the cards the site's data gives its
// pages, by the part of the data that holds each: every part with a card, so
// that a card a page takes up later is held to the checks without a word
// added here. A card names its drawing, as the site's reader of a card
// requires: a drawing under another name would go unread here.
func cardDrawingsOf(t *testing.T) map[string]string {
	t.Helper()

	file, err := os.ReadFile("data.json")
	if err != nil {
		t.Fatalf("ReadFile(data.json) error = %v, want the site's data", err)
	}
	var parts map[string]json.RawMessage
	if err := json.Unmarshal(file, &parts); err != nil {
		t.Fatalf("Unmarshal(data.json) error = %v, want nil", err)
	}
	drawings := map[string]string{}
	for name, part := range parts {
		var holder struct {
			Card *struct {
				Drawing *string `json:"drawing"`
			} `json:"card"`
		}
		// A part that is no object, such as the list of the groups, holds no card.
		if json.Unmarshal(part, &holder) != nil || holder.Card == nil {
			continue
		}
		if holder.Card.Drawing == nil {
			t.Fatalf("data.json gives the card of %s no drawing, which the site's reader of a card requires", name)
		}
		drawings[name] = *holder.Card.Drawing
	}
	if len(drawings) == 0 {
		t.Fatal("data.json holds no card, so nothing here was tested")
	}
	return drawings
}

package checks_test

import (
	"reflect"
	"strings"
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/checks"
)

// The properties here name what the drawing checks hold for any drawing, not
// only for the frames. Raising a limit never adds a refusal, so a calibration
// that loosens the limits lets through everything it let through before. And
// a drawing that shows every label it declares, under a wording that names no
// other, is never refused, whatever ordinary words stand around the labels.

// genDrawing is a drawing made of what drawings are made of — lines, labels,
// numbers, spaces, wide characters and line breaks — and of a few characters
// a drawing may not use.
func genDrawing() gopter.Gen {
	piece := gen.OneConstOf("─", "│", "┼", "A", "12", " ", "      ", "猫", "◽", "\n", "\t", "\u00A0", " \n")
	return gen.SliceOf(piece).Map(func(pieces []string) string { return strings.Join(pieces, "") })
}

// genLimits are limits of a drawing, each from least to most.
func genLimits(least, most int) gopter.Gen {
	widths, heights, spaceRuns := gen.IntRange(least, most), gen.IntRange(least, most), gen.IntRange(least, most)
	return gopter.CombineGens(widths, heights, spaceRuns).Map(func(values []any) checks.DrawingLimits {
		width, _ := values[0].(int)
		height, _ := values[1].(int)
		spaceRun, _ := values[2].(int)
		return checks.DrawingLimits{Width: width, Height: height, SpaceRun: spaceRun}
	})
}

// labelledWording is the labels of a drawing, and a wording that names them.
type labelledWording struct {
	labels  []string
	wording string
}

// genLabelledWording is a few labels, each once, and a wording that names
// every one of them among sentences that name none: an article, a unit, a
// pronoun, a word in capitals and a quotation.
func genLabelledWording() gopter.Gen {
	naming := gen.OneConstOf("Point _ is on the line.", "It is 3 cm from _ to the end.", "Put a stone at _.")
	ordinary := gen.OneConstOf("A farmer has a 3 L jug.", "Then I count again.", "It is NOT far.", `Ann says: "A knight."`)
	labels := gen.SliceOf(gen.OneConstOf("A", "B", "C", "D", "P", "Q", "X", "Y")).
		Map(func(drawn []string) []string {
			var distinct []string
			for _, label := range drawn {
				if !strings.Contains(strings.Join(distinct, ""), label) {
					distinct = append(distinct, label)
				}
			}
			return distinct
		}).
		SuchThat(func(distinct []string) bool { return len(distinct) > 0 })

	return labels.FlatMap(func(value any) gopter.Gen {
		named, _ := value.([]string)
		return gen.SliceOfN(len(named), gopter.CombineGens(naming, ordinary)).Map(func(pairs [][]any) labelledWording {
			var wording strings.Builder
			for i, pair := range pairs {
				template, _ := pair[0].(string)
				sentence, _ := pair[1].(string)
				wording.WriteString(strings.Replace(template, "_", named[i], 1) + " " + sentence + " ")
			}
			return labelledWording{labels: named, wording: wording.String()}
		})
	}, reflect.TypeOf(labelledWording{}))
}

func TestTheDrawingChecksHoldTheirProperties(t *testing.T) {
	t.Parallel()

	properties := gopter.NewProperties(nil)

	properties.Property("raising a limit never adds a refusal", prop.ForAll(
		func(drawing string, limits, raise checks.DrawingLimits) bool {
			raised := checks.DrawingLimits{
				Width:    limits.Width + raise.Width,
				Height:   limits.Height + raise.Height,
				SpaceRun: limits.SpaceRun + raise.SpaceRun,
			}
			return len(checks.DrawingFormat(drawing, raised)) <= len(checks.DrawingFormat(drawing, limits))
		},
		genDrawing(), genLimits(1, 40), genLimits(0, 20),
	))

	properties.Property("a drawing that shows what it declares, under a wording that names no more, is never refused",
		prop.ForAll(
			func(named labelledWording) bool {
				drawing := strings.Join(named.labels, "───")
				return len(checks.DrawingMatch(named.wording, drawing, labelled(named.labels...))) == 0
			},
			genLabelledWording(),
		))

	properties.TestingRun(t)
}

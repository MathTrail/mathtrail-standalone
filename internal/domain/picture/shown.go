package picture

import (
	"slices"
	"strconv"
)

// present are the labels a picture gives, each once, in the order it gives
// them; a label left out is the empty text, and is not one.
func present(labels ...string) []string {
	var kept []string
	for _, label := range labels {
		if label != "" && !slices.Contains(kept, label) {
			kept = append(kept, label)
		}
	}
	return kept
}

// texts are values a picture shows as the description writes them; a value
// left out is the empty text, and shows nothing.
func texts(values ...string) []Shown {
	var shown []Shown
	for _, value := range values {
		if value != "" {
			shown = append(shown, Shown{Text: value})
		}
	}
	return shown
}

// number is a number the card draws for one thing, written in digits.
func number(value int) Shown {
	return Shown{Text: strconv.Itoa(value)}
}

// capitalRuns are the runs of Latin capitals among some texts, each once: the
// labels a note names.
func capitalRuns(notes []string) []string {
	var runs []string
	for _, note := range notes {
		for _, token := range noteTokens(note, Point) {
			if capitals(token) {
				runs = append(runs, token)
			}
		}
	}
	return present(runs...)
}

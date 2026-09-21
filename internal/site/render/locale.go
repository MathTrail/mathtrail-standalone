package render

import "strings"

// rightToLeft holds the languages whose scripts run right to left. A page in one
// of them carries dir="rtl", and the stylesheet does the rest through logical
// properties.
var rightToLeft = map[string]bool{
	"ar": true,
	"fa": true,
	"he": true,
	"ur": true,
}

// direction reports the writing direction of a locale as an HTML dir value.
func direction(locale string) string {
	if rightToLeft[primaryLanguage(locale)] {
		return "rtl"
	}
	return "ltr"
}

// primaryLanguage returns the language subtag of a locale, so that a regional
// tag is judged by the language it belongs to.
func primaryLanguage(locale string) string {
	if i := strings.IndexByte(locale, '-'); i >= 0 {
		return locale[:i]
	}
	return locale
}

// languageNames gives each locale its name in its own language, which is the
// only name a reader looking for their language can recognise.
var languageNames = map[string]string{
	"en": "English",
	"ru": "Русский",
}

// languageName returns the endonym of a locale, falling back to the tag itself
// so that a new locale appears in the switcher before anyone names it.
func languageName(locale string) string {
	if name, ok := languageNames[locale]; ok {
		return name
	}
	return locale
}

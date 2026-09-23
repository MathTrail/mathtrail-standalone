package profile

import (
	"encoding/json"
	"fmt"
	"time"
)

// dateLayout is how a calendar day is written, and timeLayout how a moment is.
const (
	dateLayout = "2006-01-02"
	timeLayout = time.RFC3339
)

// Time is a moment in the file: always UTC, always whole seconds. The type
// carries that rather than every caller, so the file cannot pick up a local
// offset or a tail of nanoseconds from whichever clock happened to write it.
type Time struct {
	time.Time
}

// At is a moment as the file keeps it.
func At(t time.Time) Time { return Time{t.UTC().Truncate(time.Second)} }

// MarshalJSON writes the moment as RFC 3339 in UTC.
func (t Time) MarshalJSON() ([]byte, error) {
	return []byte(`"` + At(t.Time).Format(timeLayout) + `"`), nil
}

// UnmarshalJSON reads RFC 3339 and keeps it as the file keeps it.
func (t *Time) UnmarshalJSON(raw []byte) error {
	var text string
	if err := json.Unmarshal(raw, &text); err != nil {
		return fmt.Errorf("profile: a moment is a string: %w", err)
	}
	moment, err := time.Parse(timeLayout, text)
	if err != nil {
		return fmt.Errorf("profile: %q is not a moment in RFC 3339", text)
	}
	*t = At(moment)
	return nil
}

// Date is a calendar day, with no hour and no zone. The days in this file are
// the ones a person counts — when a topic was last issued, when it was
// mastered, which day the limits are counting — and none of them means a
// moment in time.
type Date struct {
	time.Time
}

// DateOf is the UTC day a moment falls on.
func DateOf(t time.Time) Date {
	utc := t.UTC()
	return Date{time.Date(utc.Year(), utc.Month(), utc.Day(), 0, 0, 0, 0, time.UTC)}
}

// MarshalJSON writes the day as a plain calendar date.
func (d Date) MarshalJSON() ([]byte, error) {
	return []byte(`"` + d.Format(dateLayout) + `"`), nil
}

// UnmarshalJSON reads a plain calendar date.
func (d *Date) UnmarshalJSON(raw []byte) error {
	var text string
	if err := json.Unmarshal(raw, &text); err != nil {
		return fmt.Errorf("profile: a day is a string: %w", err)
	}
	day, err := time.Parse(dateLayout, text)
	if err != nil {
		return fmt.Errorf("profile: %q is not a day as %s", text, dateLayout)
	}
	*d = Date{day}
	return nil
}

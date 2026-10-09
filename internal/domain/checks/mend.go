package checks

import (
	"bytes"
	"encoding/json"
	"math"
	"regexp"
	"strconv"
	"strings"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/solver"
)

// A draft is read as the model meant it where what it wrote can mean one thing
// alone: a letter in another case or between spaces, an option written as a
// number, no issues or no picture written as null. Each is set right before the checks read
// the draft, and named in it by its field, never by its value. A member
// misspelt, a value missing, or anything that could mean two things is still
// refused: read past, it would be a field the model believes it filled in.

// The fields a mend is named by.
const (
	mendedCorrectAnswer = "task.correct_answer"
	mendedOptions       = "task.options"
	mendedDistractors   = "task.distractors"
	mendedOptionCheck   = "self_check.option_check"
	mendedFinalAnswer   = "self_check.final_answer"
	mendedIssues        = "self_check.issues"
	mendedPicture       = "task.picture"
)

// plainNumber is an option written as a number that reads the same as text:
// digits, with a sign and a decimal part at most. A number written any other
// way — with an exponent, say — is not read as the text it would be.
var plainNumber = regexp.MustCompile(`^-?\d+(\.\d+)?$`)

// exactLimit is where a number stops being sure to hold its value: past 2^53 a
// float holds only some of the whole numbers, and a reader of JSON that reads
// every number as a float, as the protocol's library does, may have changed
// one on its way here.
const exactLimit = 1 << 53

// optionsAsText is the task with every option written as a plain number
// written as its text, digit for digit, and whether there was one. A number
// past exactLimit is left as it came, for the reading of the draft to refuse,
// since its digits may not be the model's. A task whose options do not read
// as an object is passed on as it came, for the reading of the draft to judge.
func optionsAsText(raw json.RawMessage) (json.RawMessage, bool) {
	var members map[string]json.RawMessage
	if json.Unmarshal(raw, &members) != nil {
		return raw, false
	}
	var options map[string]json.RawMessage
	if json.Unmarshal(members["options"], &options) != nil {
		return raw, false
	}
	mended := false
	for key, value := range options {
		written := string(bytes.TrimSpace(value))
		if !plainNumber.MatchString(written) {
			continue
		}
		if value, err := strconv.ParseFloat(written, 64); err != nil || math.Abs(value) >= exactLimit {
			continue
		}
		text, err := json.Marshal(written)
		if err != nil {
			return raw, false
		}
		options[key], mended = text, true
	}
	if !mended {
		return raw, false
	}
	return setMember(members, "options", options, raw)
}

// mendPicture reads a picture written as null, the way a card's payload
// writes no picture, or as an empty text, as no picture, and says whether it
// did.
func mendPicture(task *Task) bool {
	if task == nil || task.Picture == nil {
		return false
	}
	var text string
	blank := json.Unmarshal(task.Picture, &text) == nil && strings.TrimSpace(text) == ""
	if !absent(task.Picture) && !blank {
		return false
	}
	task.Picture = nil
	return true
}

// issuesListed is the self-check with issues written as null written as an
// empty list, and whether they were. Issues left out are not listed: a
// missing member is the model's to write.
func issuesListed(raw json.RawMessage) (json.RawMessage, bool) {
	var members map[string]json.RawMessage
	if json.Unmarshal(raw, &members) != nil {
		return raw, false
	}
	if issues, held := members["issues"]; !held || !bytes.Equal(bytes.TrimSpace(issues), []byte("null")) {
		return raw, false
	}
	return setMember(members, "issues", []any{}, raw)
}

// setMember is an object with one member set to value, or what came, as it
// came, when it cannot be written.
func setMember(members map[string]json.RawMessage, name string, value any, came json.RawMessage) (json.RawMessage, bool) {
	written, err := json.Marshal(value)
	if err != nil {
		return came, false
	}
	members[name] = written
	object, err := json.Marshal(members)
	if err != nil {
		return came, false
	}
	return object, true
}

// mendLetters sets right the letters of a draft written in another case or
// between spaces — the answer, the self-check's verdict and the keys of the
// options, the explanations and the self-check's options — and names the
// fields it set right.
func mendLetters(draft *Draft) []string {
	var mended []string
	if task := draft.Task; task != nil {
		if letter, isLetter := letterOf(task.CorrectAnswer); isLetter && letter != task.CorrectAnswer {
			task.CorrectAnswer, mended = letter, append(mended, mendedCorrectAnswer)
		}
		if keyed, changed := rekeyed(task.Options); changed {
			task.Options, mended = keyed, append(mended, mendedOptions)
		}
		if keyed, changed := rekeyed(task.Distractors); changed {
			task.Distractors, mended = keyed, append(mended, mendedDistractors)
		}
	}
	if check := draft.SelfCheck; check != nil {
		if keyed, changed := rekeyed(check.OptionCheck); changed {
			check.OptionCheck, mended = keyed, append(mended, mendedOptionCheck)
		}
		verdict := strings.ToUpper(strings.TrimSpace(check.FinalAnswer))
		if verdict != check.FinalAnswer && (solver.Place(verdict) >= 0 || verdict == unsolvable) {
			check.FinalAnswer, mended = verdict, append(mended, mendedFinalAnswer)
		}
	}
	return mended
}

// letterOf is a letter of the options as it was written, in any case and
// between spaces, and whether it is one.
func letterOf(written string) (string, bool) {
	letter := strings.ToUpper(strings.TrimSpace(written))
	return letter, solver.Place(letter) >= 0
}

// rekeyed is entries under their letters set right, and whether any was. Set
// right is nothing when a key is no letter at all, or when two keys would come
// to one letter: what was meant is no longer beyond doubt.
func rekeyed[V any](entries map[string]V) (map[string]V, bool) {
	keyed := make(map[string]V, len(entries))
	changed := false
	for key, value := range entries {
		letter, isLetter := letterOf(key)
		if _, taken := keyed[letter]; !isLetter || taken {
			return entries, false
		}
		keyed[letter] = value
		changed = changed || letter != key
	}
	return keyed, changed
}

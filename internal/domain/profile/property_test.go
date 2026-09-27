package profile_test

import (
	"bytes"
	"encoding/json"
	"maps"
	"strings"
	"testing"
	"time"
	"unicode"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/profile"
	"github.com/MathTrail/mathtrail-standalone/internal/domain/rating"
)

// The properties here name what has to hold for every child rather than for
// the five in the fixtures. Two things in this package have that shape: the
// file, which must come back as it went in whatever the parent typed into it,
// and the sealed block, which must leave no trace of itself anywhere else —
// the whole point of sealing is that the rest of the file says nothing about
// what is inside it.

// genText produces what a parent might type: any script, any length, with the
// control characters and the length the file refuses taken out, because a
// profile past a cap is refused rather than written.
func genText(limit int) gopter.Gen {
	return gen.AnyString().Map(func(text string) string {
		text = strings.Map(func(r rune) rune {
			if unicode.IsControl(r) {
				return -1
			}
			return r
		}, text)
		if runes := []rune(text); len(runes) > limit {
			text = string(runes[:limit])
		}
		return text
	})
}

func TestTheFileHoldsItsProperties(t *testing.T) {
	t.Parallel()

	properties := gopter.NewProperties(nil)

	properties.Property("what a parent typed comes back as they typed it", prop.ForAll(
		func(pseudonym, notes, interest string) bool {
			if pseudonym == "" || interest == "" {
				return true // a profile with neither is refused, and says so
			}
			p := parseFixture(t, "dima")
			p.Student.Pseudonym, p.Student.Notes = pseudonym, notes
			p.Student.Interests = []string{interest}

			written, err := profile.Marshal(p)
			if err != nil {
				return false
			}
			read, err := profile.Parse(written)
			if err != nil {
				return false
			}
			return read.Student.Pseudonym == pseudonym &&
				read.Student.Notes == notes &&
				len(read.Student.Interests) == 1 && read.Student.Interests[0] == interest
		},
		genText(profile.MaxPseudonym), genText(profile.MaxNotes), genText(profile.MaxInterest),
	))

	properties.Property("every object is written with its keys in order", prop.ForAll(
		func(pseudonym string, trap string, times int) bool {
			if pseudonym == "" || trap == "" {
				return true
			}
			p := parseFixture(t, "dima")
			p.Student.Pseudonym = pseudonym
			topic := p.Topics["counting.gaps"]
			topic.Traps[trap] = times
			topic.Answers += times
			p.Topics["counting.gaps"] = topic

			written, err := profile.Marshal(p)
			if err != nil {
				return false
			}
			checkSorted(t, "", written)
			return !t.Failed()
		},
		genText(profile.MaxPseudonym), gen.Identifier(), gen.IntRange(1, 9),
	))

	properties.TestingRun(t)
}

// Whatever is sealed, nothing else in the file moves with it. A file whose
// open part changed shape with the answer — a longer wording for a longer
// explanation, a field that appears only sometimes — would give away by its
// outline what the sealing is there to hide.
func TestNothingOutsideTheSealedBlockFollowsIt(t *testing.T) {
	t.Parallel()

	properties := gopter.NewProperties(nil)

	// Two generators rather than one used twice, so that each argument reads
	// as the sealed value it stands for.
	before, after := genText(200), genText(200)

	properties.Property("the file differs only where the sealed value does", prop.ForAll(
		func(first, second string) bool {
			// The prefix is what a sealed value always carries, and it keeps a
			// generated one from colliding with a word elsewhere in the file.
			first, second = "mt1.t.kQ7fWx."+first, "mt1.t.kQ7fWx."+second
			if first == second {
				return true
			}

			written := make([][]byte, 2)
			for i, sealed := range []string{first, second} {
				p := parseFixture(t, "masha")
				p.CurrentTask.Sealed = sealed
				out, err := profile.Marshal(p)
				if err != nil {
					return false
				}
				written[i] = out
			}

			// Put the second sealed value where the first one was: if nothing
			// else followed it, the two files are now the same file.
			was, now := asWritten(first), asWritten(second)
			return bytes.Equal(bytes.Replace(written[0], was, now, 1), written[1])
		},
		before, after,
	))

	properties.TestingRun(t)
}

// asWritten is a string as the profile writes it. The profile leaves the
// characters of HTML alone and encoding/json does not, so a value read back
// out of a written file has to be looked for in the spelling the file uses.
func asWritten(text string) []byte {
	var out bytes.Buffer
	encoder := json.NewEncoder(&out)
	encoder.SetEscapeHTML(false)
	if err := encoder.Encode(text); err != nil {
		return nil
	}
	return bytes.TrimRight(out.Bytes(), "\n")
}

// An answer moves the child once. Sent again any number of times, whatever
// those say, the profile stays exactly as the first left it, and every later
// telling is the first one told again. And "I don't know" is a wrong answer
// that took no trap: whatever the child and wherever the task stood, it moves
// the levels and the runs exactly as a wrong letter does, and the map of the
// child's mistakes does not count it.
func TestAnAnswerHoldsItsProperties(t *testing.T) {
	t.Parallel()

	properties := gopter.NewProperties(nil)
	children := gen.OneConstOf("dima", "masha", "olya", "petya", "sasha")
	choices := gen.OneConstOf("A", "B", "C", "D", "E", profile.DontKnow)
	levels := gen.OneConstOf(rating.Grades12, rating.Grades34, rating.Grades56)
	difficulties := gen.IntRange(profile.MinDifficulty, profile.MaxDifficulty)
	const topic = "counting.gaps"

	properties.Property("an answer is recorded once, whatever is sent again", prop.ForAll(
		func(child, first string, again []string, level rating.GradeLevel, difficulty int) bool {
			p := parseFixture(t, child)
			id := answeringAt(t, p, topic, rating.Point{GradeLevel: level, Difficulty: difficulty})
			recorded := give(t, p, id, first, issued.Add(time.Minute))
			before, err := profile.Marshal(p)
			if err != nil {
				return false
			}

			// Told again is told without what only recording knew.
			want := recorded
			want.Probability, want.Pace, want.Mastered, want.Unmastered, want.Again = 0, "", false, false, true
			for _, choice := range again {
				if give(t, p, id, choice, issued.Add(time.Hour)) != want {
					return false
				}
			}
			after, err := profile.Marshal(p)
			return err == nil && bytes.Equal(before, after)
		},
		children, choices, gen.SliceOfN(3, choices), levels, difficulties,
	))

	properties.Property("I don't know moves the child as a wrong letter does, and counts no trap", prop.ForAll(
		func(child string, level rating.GradeLevel, difficulty int, hint bool) bool {
			unsure, mistaken := parseFixture(t, child), parseFixture(t, child)
			traps := maps.Clone(unsure.Topics[topic].Traps)
			point := rating.Point{GradeLevel: level, Difficulty: difficulty}
			for p, choice := range map[*profile.Profile]string{unsure: profile.DontKnow, mistaken: wrongLetter} {
				id := answeringAt(t, p, topic, point)
				if _, err := p.Record(profile.Answered{TaskID: id, Choice: choice, HintUsed: hint, At: issued.Add(time.Minute)},
					newSealer(t)); err != nil {
					return false
				}
			}

			doubtful, wrong := unsure.Topics[topic], mistaken.Topics[topic]
			last := unsure.Recent[len(unsure.Recent)-1]
			return unsure.Ratings == mistaken.Ratings && doubtful.Delta == wrong.Delta &&
				doubtful.TopStreak == wrong.TopStreak && doubtful.WrongStreak == wrong.WrongStreak &&
				maps.Equal(doubtful.Traps, traps) && wrong.Traps[wrongTrap] == traps[wrongTrap]+1 &&
				last.Confused && !last.Correct && last.Chosen == "" && last.Trap == ""
		},
		children, levels, difficulties, gen.Bool(),
	))

	properties.TestingRun(t)
}

package profile_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"maps"
	"slices"
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
					newSealer(t), everyLevel); err != nil {
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

// How an answer went is told for a card of its own: whatever the child, the
// choice and where the task stood, it is told as it was recorded, telling it
// moves nothing in the file, and before the answer the task is never opened.
func TestTellingAnAnswerHoldsItsProperties(t *testing.T) {
	t.Parallel()

	properties := gopter.NewProperties(nil)
	const topic = "counting.gaps"

	properties.Property("told as recorded, moving nothing, and never opened before the answer", prop.ForAll(
		func(child, choice string, level rating.GradeLevel, difficulty int) bool {
			p := parseFixture(t, child)
			id := answeringAt(t, p, topic, rating.Point{GradeLevel: level, Difficulty: difficulty})
			if _, err := p.Told(id, opensNothing{Sealer: newSealer(t), t: t}); !errors.Is(err, profile.ErrNotAnswered) {
				return false
			}
			recorded := give(t, p, id, choice, issued.Add(time.Minute))
			before, err := profile.Marshal(p)
			if err != nil {
				return false
			}
			told, err := p.Told(id, newSealer(t))
			after, marshalled := profile.Marshal(p)
			return err == nil && told == toldOnly(&recorded) && marshalled == nil && bytes.Equal(before, after)
		},
		gen.OneConstOf("dima", "masha", "olya", "petya", "sasha"),
		gen.OneConstOf("A", "B", "C", "D", "E", profile.DontKnow),
		gen.OneConstOf(rating.Grades12, rating.Grades34, rating.Grades56),
		gen.IntRange(profile.MinDifficulty, profile.MaxDifficulty),
	))

	properties.TestingRun(t)
}

// In a file of the cautious estimate's, however a topic goes, it is declared
// mastered exactly where the rule puts the child: on a right and unaided
// answer, with five answers in the topic, at the level the cautious estimate
// clears among the topic's own levels no higher than the task's, whenever that
// level is above the one held, and since the day of the answer. Two wrong
// answers in a row take it away, and nothing else does.
func TestMasteryHoldsItsProperties(t *testing.T) {
	t.Parallel()

	properties := gopter.NewProperties(nil)
	levels := gen.OneConstOf(rating.Grades12, rating.Grades34, rating.Grades56)
	difficulties := gen.IntRange(profile.MinDifficulty, profile.MaxDifficulty)
	rights, hints := gen.Bool(), gen.Bool()
	steps := gopter.CombineGens(levels, difficulties, rights, hints).Map(func(values []any) answerStep {
		level, _ := values[0].(rating.GradeLevel)
		difficulty, _ := values[1].(int)
		correct, _ := values[2].(bool)
		hint, _ := values[3].(bool)
		return answerStep{point: rating.Point{GradeLevel: level, Difficulty: difficulty}, correct: correct, hint: hint}
	})
	// The levels a topic can be taught at: a run of neighbouring levels, as
	// every topic of the catalog is.
	taughtAt := gen.OneConstOf(
		[]rating.GradeLevel{rating.Grades12},
		[]rating.GradeLevel{rating.Grades56},
		[]rating.GradeLevel{rating.Grades34, rating.Grades56},
		everyLevel,
	)

	properties.Property("declared where the cautious estimate clears, held until two failures in a row", prop.ForAll(
		func(theta float64, answers int, taught []rating.GradeLevel, lesson []answerStep) bool {
			p := settled(t, theta, answers, 0, 0)
			for number, step := range lesson {
				before := p.Topics[masteredTopic]
				at := issued.Add(time.Duration(number+1) * time.Hour)
				recorded := answerTaughtAt(t, p, step.point, number+1, step.correct, step.hint, taught)
				if !keptByTheRule(p, &before, &recorded, &ruledAnswer{answerStep: step, taught: taught, day: profile.DateOf(at)}) {
					return false
				}
			}
			return true
		},
		gen.Float64Range(-2, 7), gen.IntRange(0, 400), taughtAt, gen.SliceOfN(30, steps),
	))

	properties.TestingRun(t)
}

// answerStep is one answer of a generated lesson: where its task stood, and
// whether the child got it right and took the hint.
type answerStep struct {
	point         rating.Point
	correct, hint bool
}

// ruledAnswer is an answer as the rule is held to it: the step, the levels its
// topic is taught at, and the day it came.
type ruledAnswer struct {
	answerStep
	taught []rating.GradeLevel
	day    profile.Date
}

// keptByTheRule says whether what an answer did to the mastery of its topic is
// what the rule asks for: a loss after two wrong answers in a row from
// mastery; a declaration on a right and unaided answer with five answers in
// the topic, at the level the cautious estimate clears among the topic's
// levels, when that level is above the one held, since the day of the answer;
// and otherwise nothing moved. Which level the estimate clears is the rating's
// to say, and its own properties hold it.
func keptByTheRule(p *profile.Profile, before *profile.Topic, recorded *profile.Recorded, a *ruledAnswer) bool {
	after := p.Topics[masteredTopic]
	lost := before.MasteredSince != nil && after.WrongStreak >= profile.MasteryLostAfter
	earned, clears := rating.MasteredAt(p.Ratings.Theta+after.Delta, p.Ratings.Answers, after.Answers, a.point.GradeLevel, a.taught)
	earns := clears && a.correct && !a.hint && after.Answers >= profile.MasteryAnswers &&
		(before.MasteredLevel == nil || before.MasteredLevel.Shift() < earned.Shift())
	switch {
	case recorded.Unmastered:
		return lost && after.MasteredSince == nil && after.MasteredLevel == nil
	case recorded.Mastered:
		return earns && after.MasteredLevel != nil && *after.MasteredLevel == earned &&
			after.MasteredSince != nil && after.MasteredSince.Equal(a.day.Time)
	default:
		return !lost && !earns && sameMastery(before, &after)
	}
}

// sameMastery says whether a topic is mastered at the same level since the
// same day as before, or is still not mastered.
func sameMastery(before, after *profile.Topic) bool {
	if before.MasteredSince == nil || after.MasteredSince == nil {
		return before.MasteredSince == nil && after.MasteredSince == nil &&
			before.MasteredLevel == nil && after.MasteredLevel == nil
	}
	return before.MasteredSince.Equal(after.MasteredSince.Time) && *before.MasteredLevel == *after.MasteredLevel
}

// However a lesson goes — tasks asked for now or ahead, attempts turned down,
// tasks kept, handed out, let go and left, seals lost, time passing — the task
// of a request only moves forward: from being written to kept ready, to the
// card or to not coming; from kept to the card or to not coming; and from the
// card to not coming, never back. A card that has been told its task will not
// come stops asking on the strength of it. A task is being written only inside
// its request's window, with its count of attempts turned down never falling;
// and at any moment one task at most is being written, one at most is kept and
// one at most is on the card, and none is kept while one is being written.
func TestTheTaskOfARequestOnlyMovesForward(t *testing.T) {
	t.Parallel()

	properties := gopter.NewProperties(nil)
	properties.Property("a task never moves back, and one at a time is written, kept or shown", prop.ForAll(
		func(steps []int) bool {
			walk := walkALesson(t)
			for _, step := range steps {
				walk.take(step)
				if !walk.movedForward() {
					return false
				}
			}
			return true
		},
		gen.SliceOf(gen.IntRange(0, 11)),
	))

	properties.TestingRun(t)
}

// requestWindow is how long the lessons of these properties wait for a task.
const requestWindow = 15 * time.Minute

// forward is where the task of a request may move from where it stands.
var forward = map[profile.TaskState][]profile.TaskState{
	profile.TaskBeingWritten: {profile.TaskBeingWritten, profile.TaskKept, profile.TaskOnTheCard, profile.TaskNotComing},
	profile.TaskKept:         {profile.TaskKept, profile.TaskOnTheCard, profile.TaskNotComing},
	profile.TaskOnTheCard:    {profile.TaskOnTheCard, profile.TaskNotComing},
	profile.TaskNotComing:    {profile.TaskNotComing},
}

// seenTask is what was last said of the task of a request, and when the
// request was opened.
type seenTask struct {
	state   profile.TaskState
	refused int
	opened  time.Time
}

// lessonWalk is a profile taken through a lesson a step at a time, with what
// was last said of the task of every request it asked for.
type lessonWalk struct {
	p      *profile.Profile
	sealer profile.Sealer
	brief  profile.Brief
	now    time.Time
	last   map[string]seenTask
}

// walkALesson starts a lesson on a profile with no request open and no task
// on the card.
func walkALesson(t *testing.T) *lessonWalk {
	t.Helper()

	p := parseFixture(t, "dima")
	p.OpenRequest, p.CurrentTask = nil, nil
	return &lessonWalk{p: p, sealer: newSealer(t), brief: briefOn("counting.gaps"), now: asked, last: map[string]seenTask{}}
}

// take does one thing a lesson does: ask for a task, turn an attempt down,
// hand the task out, take it off the card, lose its seal, let a minute or six
// pass, ask for the next task ahead, keep it, hand out the task kept, wait for
// the task being written ahead once the child asks for it, or let go of the
// task kept. An attempt is turned down or a task kept or handed out only while
// the request is still awaited, as the tool that takes a task makes sure: a
// task handed in for any other request is stale, and changes nothing. A task
// is written ahead only while nothing else is being written, as the tool that
// asks for one makes sure.
func (w *lessonWalk) take(step int) {
	open := w.p.OpenRequest
	awaited := open != nil && open.Awaited(requestWindow, w.now)
	if step >= firstStepAhead {
		w.takeAhead(step, open, awaited)
		return
	}
	switch step {
	case 0:
		w.asked(w.p.Ask(&w.brief, profile.TutorRule, "en", w.now))
	case 1:
		if awaited {
			_, _ = w.p.Refuse(w.now)
		}
	case 2:
		if awaited && !open.Ahead {
			_, _ = w.p.Issue(written(), secret(), w.sealer, w.now)
		}
	case 3:
		w.p.Skip(w.now)
	case 4:
		w.p.DiscardTask()
	case 5:
		w.now = w.now.Add(time.Minute)
	case 6:
		w.now = w.now.Add(6 * time.Minute)
	}
}

// firstStepAhead is the first of the steps that write a task ahead.
const firstStepAhead = 7

// takeAhead does one thing a lesson does with a task written ahead, open being
// the request open and awaited whether it is still waited for.
func (w *lessonWalk) takeAhead(step int, open *profile.OpenRequest, awaited bool) {
	switch step {
	case 7:
		if !awaited {
			if request, err := w.p.AskAhead(&w.brief, profile.TutorRule, "en", "", w.now); err == nil {
				w.asked(request)
			}
		}
	case 8:
		if awaited && open.Ahead {
			_, _ = w.p.Keep(written(), secret(), w.sealer, w.now)
		}
	case 9:
		_, _ = w.p.HandOutReady(w.now)
	case 10:
		if awaited && open.Ahead {
			w.p.Skip(w.now)
			open.Await(w.now)
			// Its window starts again: the child waits for it from now on.
			w.asked(open)
		}
	case 11:
		w.p.DropReady()
	}
}

// asked notes a request just opened as being written.
func (w *lessonWalk) asked(request *profile.OpenRequest) {
	w.last[request.ID] = seenTask{state: profile.TaskBeingWritten, opened: w.now}
}

// movedForward says whether the task of every request asked for moved only
// forward since the last step, was written only inside its window with no
// fewer attempts turned down, and whether one task at most is being written,
// one at most is kept and one at most is on the card, with none kept while one
// is being written.
func (w *lessonWalk) movedForward() bool {
	writing, kept, shown := 0, 0, 0
	for id, before := range w.last {
		state, refused := w.p.TaskFor(id, requestWindow, w.now)
		if !slices.Contains(forward[before.state], state) {
			return false
		}
		switch state {
		case profile.TaskBeingWritten:
			if refused < before.refused || w.now.Sub(before.opened) >= requestWindow {
				return false
			}
			writing++
		case profile.TaskKept:
			kept++
		case profile.TaskOnTheCard:
			shown++
		}
		w.last[id] = seenTask{state: state, refused: refused, opened: before.opened}
	}
	if w.p.ReadyTask != nil && w.p.OpenRequest != nil {
		return false
	}
	return writing <= 1 && kept <= 1 && shown <= 1
}

// Whatever an edit holds, its refusal names each field it found wrong once,
// by a code of the closed list, and with the rule in words. A card shows one
// message under a field, from a code it has words for; a field named twice
// would show two, and a code it was never told of none it could say.
func TestARefusalHoldsItsProperties(t *testing.T) {
	t.Parallel()

	properties := gopter.NewProperties(nil)
	known := profile.Codes()

	properties.Property("each field once, by a code of the list, with its rule", prop.ForAll(
		func(pseudonym string, grade int, interests, excluded []string, notes, language, topic string) bool {
			edit := profile.Edit{
				Pseudonym: &pseudonym, Grade: &grade, Interests: interests, ExcludedSkills: excluded,
				Notes: &notes, UILanguage: &language, LessonTopic: &topic,
			}
			_, problems := profile.NewStudent(&edit, shipped)
			named := map[string]bool{}
			for _, problem := range problems {
				if named[problem.Field] || !slices.Contains(known, problem.Code) || problem.Rule == "" {
					return false
				}
				named[problem.Field] = true
			}
			return true
		},
		gen.AnyString(), gen.IntRange(profile.MinGrade-2, profile.MaxGrade+2),
		gen.SliceOf(gen.AnyString()), gen.SliceOf(gen.OneGenOf(gen.OneConstOf("fractions", "negative_numbers", ""), gen.Identifier())),
		gen.OneGenOf(gen.AnyString(), gen.Const(strings.Repeat("ж", profile.MaxNotes+1))),
		gen.OneConstOf("", " ", "en", "pt-br", "und", "x-private", strings.Repeat("a", profile.MaxLanguageTag+1)),
		gen.OneConstOf("", " ", "time.clocks", " logic.ordering ", "astronomy.stars", strings.Repeat("t", profile.MaxLessonTopic+1)),
	))

	properties.TestingRun(t)
}

// Whatever clock a family keeps and whenever a task is given, the counters of
// its day hold to the family's next midnight and not a second past it: the
// last second of the family's day still counts the task, and the first of the
// next counts none. And whatever clocks a card tells later, in whatever order,
// no day ends sooner or later than the midnight of the clock it began under.
func TestTheFamilysDayHoldsItsProperties(t *testing.T) {
	t.Parallel()

	properties := gopter.NewProperties(nil)
	moments := gen.Int64Range(time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC).Unix(), time.Date(2030, 1, 1, 0, 0, 0, 0, time.UTC).Unix())
	quarters := gen.IntRange(-12*4, 14*4)

	properties.Property("the counters hold to the family's midnight and no further", prop.ForAll(
		func(quarters int, given int64) bool {
			offset := time.Duration(quarters*15) * time.Minute
			var p profile.Profile
			p.SetClock(quarters * 15)
			at := time.Unix(given, 0).UTC()
			p.CountAccepted(at)

			shown := at.Add(offset)
			midnight := time.Date(shown.Year(), shown.Month(), shown.Day()+1, 0, 0, 0, 0, time.UTC).Add(-offset)
			return p.Daily.Today(midnight.Add(-time.Second)).Accepted == 1 &&
				p.Daily.Today(midnight).Accepted == 0
		},
		quarters, moments,
	))

	properties.Property("a clock told moves the end of no day", prop.ForAll(
		func(first int, told []int, given, later int64, count int) bool {
			var p profile.Profile
			p.SetClock(first * 15)
			at := time.Unix(given, 0).UTC()
			for range count {
				p.CountAccepted(at)
			}
			for _, quarters := range told {
				p.SetClock(quarters * 15)
			}

			offset := time.Duration(first*15) * time.Minute
			shown := at.Add(offset)
			ends := time.Date(shown.Year(), shown.Month(), shown.Day()+1, 0, 0, 0, 0, time.UTC).Add(-offset)
			then := at.Add(time.Duration(later) * time.Second)
			want := 0
			if then.Before(ends) {
				want = count
			}
			return p.Daily.Today(then).Accepted == want
		},
		quarters, gen.SliceOf(quarters), moments, gen.Int64Range(0, 48*60*60), gen.IntRange(0, 30),
	))

	properties.TestingRun(t)
}

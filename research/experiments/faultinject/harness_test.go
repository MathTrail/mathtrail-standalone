package main

import (
	"context"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/MathTrail/mathtrail-standalone/content"
	"github.com/MathTrail/mathtrail-standalone/internal/domain/checks"
	"github.com/MathTrail/mathtrail-standalone/internal/domain/rating"
	"github.com/MathTrail/mathtrail-standalone/internal/domain/solver"
	"github.com/MathTrail/mathtrail-standalone/internal/infra/starlark"
	"github.com/MathTrail/mathtrail-standalone/research/experiments/reviewing"
)

// bench is what the harness's tests share: the content, a sandbox, and the
// hosts of grades 5–6 that the checks accept as they are. That level is the
// one with drawings and with hints of its own, so every operator has a host
// among them.
type bench struct {
	shipped *content.Content
	runner  solver.Runner
	pool    *pool
}

var loadBench = sync.OnceValues(func() (*bench, error) {
	shipped, err := content.Load()
	if err != nil {
		return nil, err
	}
	// The tests run in parallel and share one sandbox of one slot, so a run
	// may wait for the slot longer than the service lets it; the wait is no
	// part of what is tested.
	limits := reviewing.SandboxLimits
	limits.Wait = time.Minute
	runner, err := starlark.New(limits)
	if err != nil {
		return nil, err
	}
	var eligible []*host
	for _, h := range reviewing.Hosts(shipped) {
		if h.Level != rating.Grades56 {
			continue
		}
		v, err := reviewing.Review(context.Background(), runner, reviewing.Without(shipped, h.Base.Task.Question), &h.Base)
		if err != nil {
			return nil, err
		}
		if !v.Refused() && len(v.Unchecked) == 0 {
			eligible = append(eligible, h)
		}
	}
	return &bench{shipped: shipped, runner: runner, pool: &pool{content: shipped, hosts: eligible}}, nil
})

func testBench(t *testing.T) *bench {
	t.Helper()
	b, err := loadBench()
	if err != nil {
		t.Fatalf("the bench: %v", err)
	}
	return b
}

// firstCase is the case an operator makes of the first host of the bench it
// applies to.
func firstCase(t *testing.T, b *bench, op *operator) (*host, mutant) {
	t.Helper()
	for _, h := range b.pool.hosts {
		if made, applies := op.Apply(newMaker(h, b.pool, op.ID)); applies {
			return h, made
		}
	}
	t.Fatalf("%s applies to no eligible host of grades 5–6", op.ID)
	return nil, mutant{}
}

// An operator makes the same case of the same host every time.
func TestACaseIsRepeatable(t *testing.T) {
	t.Parallel()
	b := testBench(t)
	for _, op := range operators() {
		h, first := firstCase(t, b, &op)
		again, _ := op.Apply(newMaker(h, b.pool, op.ID))
		if !first.Sub.SameAs(&again.Sub) {
			t.Errorf("%s made two different cases of %s", op.ID, h.ID)
		}
	}
}

// Every operator changes the submission, and changes it as it says.
func TestEveryOperatorChangesTheTaskAsItSays(t *testing.T) {
	t.Parallel()
	b := testBench(t)
	for _, op := range operators() {
		t.Run(op.ID, func(t *testing.T) {
			t.Parallel()
			h, made := firstCase(t, b, &op)
			if made.Sub.SameAs(&h.Base) {
				t.Fatalf("%s left %s as it was", op.ID, h.ID)
			}
			describe, known := descriptions[op.ID]
			if !known {
				describe = descriptions[strings.SplitN(op.ID, "-", 2)[0]]
			}
			if describe == nil {
				t.Fatalf("no description of what %s does", op.ID)
			}
			if wrong := describe(b, &h.Base, &made); wrong != "" {
				t.Errorf("%s on %s: %s", op.ID, h.ID, wrong)
			}
		})
	}
}

// description says what is wrong with a case, given the host it came from, or
// nothing when the case is what its operator says it makes.
type description func(b *bench, before *submission, made *mutant) string

func unless(holds bool, wrong string) string {
	if holds {
		return ""
	}
	return wrong
}

// changedLetters are the letters whose option or explanation differs.
func changedExplanations(before, after *submission) []string {
	var changed []string
	for _, letter := range solver.Letters() {
		if before.Task.Distractors[letter] != after.Task.Distractors[letter] {
			changed = append(changed, letter)
		}
	}
	return changed
}

var descriptions = map[string]description{
	"D01a": func(_ *bench, before *submission, m *mutant) string {
		task := &m.Sub.Task
		_, oldExplained := task.Distractors[before.Task.CorrectAnswer]
		_, keyExplained := task.Distractors[task.CorrectAnswer]
		return unless(task.CorrectAnswer != before.Task.CorrectAnswer && m.Sub.SelfCheck.FinalAnswer == task.CorrectAnswer &&
			oldExplained && !keyExplained && len(task.Distractors) == 4, "the key did not move with its explanation and self-check")
	},
	"D02a": func(_ *bench, before *submission, m *mutant) string {
		right := before.Task.Options[before.Task.CorrectAnswer]
		for _, letter := range solver.Letters() {
			text := m.Sub.Task.Options[letter]
			if letter != before.Task.CorrectAnswer && text != right && solver.Key(text) == solver.Key(right) {
				return ""
			}
		}
		return "no wrong option says the right answer another way"
	},
	"D02b": func(_ *bench, before *submission, m *mutant) string {
		words, _ := inWords(before.Task.Options[before.Task.CorrectAnswer])
		options := m.Sub.Options()
		return unless(slices.Contains(options[:], words), "no option is the answer in words")
	},
	"D03a": func(_ *bench, before *submission, m *mutant) string {
		key := before.Task.CorrectAnswer
		return unless(m.Sub.Task.CorrectAnswer == key && m.Sub.Task.Options[key] != before.Task.Options[key], "the right option was not replaced")
	},
	"D04a": solverHolds("injected = 1 // 0"),
	"D04b": func(_ *bench, _ *submission, m *mutant) string {
		return unless(strings.HasSuffix(m.Sub.Solver, "\nsolve(\n"), "the solver does not end with a line that does not parse")
	},
	"D04c": func(_ *bench, _ *submission, m *mutant) string {
		return unless(!strings.Contains(m.Sub.Solver, "def solve(") && strings.Contains(m.Sub.Solver, "def solve_injected("), "solve was not renamed")
	},
	"D04d": solverHolds(`return "A"`),
	"D04e": solverHolds(`injected = "%5d" % 1`),
	"D05a": solverHolds("for injected in range(100000000):"),
	"D06a": func(_ *bench, before *submission, m *mutant) string {
		return solverHolds("return [\""+before.Task.CorrectAnswer+"\"]")(nil, before, m)
	},
	"D07a": func(_ *bench, before *submission, m *mutant) string {
		answer := m.Sub.SelfCheck.FinalAnswer
		return unless(answer != before.Task.CorrectAnswer && solver.Place(answer) >= 0, "the self-check does not arrive at a wrong option")
	},
	"D07b": func(_ *bench, _ *submission, m *mutant) string {
		return unless(m.Sub.SelfCheck.FinalAnswer == "UNSOLVABLE", "the self-check does not call the task unsolvable")
	},
	"D08a": func(_ *bench, _ *submission, m *mutant) string {
		issues := m.Sub.SelfCheck.Issues
		return unless(len(issues) == 1 && issues[0].Severity == "blocking" && slices.Contains(issueTypes, issues[0].Type), "no blocking issue of a known type")
	},
	"D09a": func(_ *bench, before *submission, m *mutant) string {
		texts := func(s *submission) []string {
			return []string{s.Task.Question, s.Task.Hint, s.Task.Solution, s.Task.CoreIdea, s.Task.DesignThoughtProcess, s.Brief.Setting, s.Brief.Rationale}
		}
		emptiedNow := 0
		for i, text := range texts(&m.Sub) {
			if text == "" && texts(before)[i] != "" {
				emptiedNow++
			}
		}
		return unless(emptiedNow == 1, "not exactly one required text was emptied")
	},
	"D09b": func(_ *bench, _ *submission, m *mutant) string {
		return unless(len(m.Sub.Task.Options) == 4 && len(m.Sub.Task.Distractors) == 3, "an option and its explanation were not removed")
	},
	"D09c": func(_ *bench, _ *submission, m *mutant) string {
		return unless(m.Sub.Task.CorrectAnswer == "F", "the key is not F")
	},
	"D09d": func(_ *bench, _ *submission, m *mutant) string {
		_, explained := m.Sub.Task.Distractors[m.Sub.Task.CorrectAnswer]
		return unless(explained && len(m.Sub.Task.Distractors) == 5, "the key's option is not explained")
	},
	"D09e": func(_ *bench, _ *submission, m *mutant) string {
		return unless(len([]rune(m.Sub.Task.Solution)) > longestSolution, "the solution is not past its limit")
	},
	"D09f": func(_ *bench, _ *submission, m *mutant) string {
		return unless(len(m.Sub.SelfCheck.OptionCheck) == 4, "no option lost its reason")
	},
	"D09g": func(_ *bench, _ *submission, m *mutant) string {
		issues := m.Sub.SelfCheck.Issues
		return unless(len(issues) == 1 && issues[0].Type == "injected", "no issue of an unknown type")
	},
	"D10a": func(_ *bench, _ *submission, m *mutant) string {
		return unless(m.Sub.Brief.TargetConcept != m.Sub.Asked.TargetConcept, "the brief's topic is the request's")
	},
	"D10b": func(_ *bench, _ *submission, m *mutant) string {
		return unless(m.Sub.Brief.GradeLevel != m.Sub.Asked.GradeLevel && m.Sub.Brief.GradeLevel.Known(), "the brief's level is not another level")
	},
	"D10c": func(_ *bench, _ *submission, m *mutant) string {
		return unless(m.Sub.Brief.Difficulty != m.Sub.Asked.Difficulty, "the brief's difficulty is the request's")
	},
	"D10d": func(_ *bench, _ *submission, m *mutant) string {
		return unless(len(m.Sub.Asked.ExcludedSkills) == 1 && len(m.Sub.Brief.ExcludedSkills) == 0, "no excluded skill was dropped")
	},
	"D11a": func(_ *bench, before *submission, m *mutant) string {
		changed := changedExplanations(before, &m.Sub)
		return unless(len(changed) == 1 && m.Sub.Task.Distractors[changed[0]].Trap == "injected_trap", "no explanation names an unknown trap")
	},
	"D12a": explanationBecomes(func(_ *bench, before *submission, _ string) string { return firstWords(before.Task.Solution, 6) }),
	"D12b": explanationBecomes(func(_ *bench, before *submission, _ string) string { return firstWords(before.Task.Hint, 6) }),
	"D12c": func(_ *bench, before *submission, m *mutant) string {
		changed := changedExplanations(before, &m.Sub)
		if len(changed) != 1 {
			return "not exactly one explanation changed"
		}
		text := m.Sub.Task.Distractors[changed[0]].Text
		for _, letter := range solver.Letters() {
			if letter != changed[0] && before.Task.Distractors[letter].Text == text && text != "" {
				return ""
			}
		}
		return "the explanation is not another's"
	},
	"D12d": explanationBecomes(func(b *bench, before *submission, letter string) string {
		description, _ := b.shipped.TrapDescription(before.Task.Distractors[letter].Trap)
		return description
	}),
	"D12e": explanationBecomes(func(_ *bench, before *submission, letter string) string {
		return firstWords(before.Task.Distractors[letter].Text, 2)
	}),
	"D12f": explanationBecomes(func(*bench, *submission, string) string { return "" }),
	"D13a": func(_ *bench, before *submission, m *mutant) string {
		allowed := checks.ReadabilityLimitsFor(m.Sub.Brief.GradeLevel).SentenceWords + wordsPastLimit
		long := slices.ContainsFunc(sentencesOf(m.Sub.Task.Question), func(s string) bool { return wordsIn(s) >= allowed })
		fewer := len(sentencesOf(m.Sub.Task.Question)) < len(sentencesOf(before.Task.Question))
		return unless(long && fewer, "no sentence was joined past the level's limit")
	},
	"D13b": func(_ *bench, _ *submission, m *mutant) string {
		return unless(strings.Contains(m.Sub.Task.Question, longWords), "the sentence of long words is missing")
	},
	"D14a": copyOf(func(question string) string { return question }),
	"D14b": copyOf(func(question string) string { bumped, _ := bumpNumbers(question); return bumped }),
	"D14c": copyOf(func(question string) string { swapped, _ := firstTwoSwapped(question); return swapped }),
	"D15a": historyHolds(func(question string) string { return question }),
	"D15b": historyHolds(func(question string) string { bumped, _ := bumpNumbers(question); return bumped }),
	"D16a": drawingLines(func(lines []string) bool { return len(lines) == tooManyLines }),
	"D16b": drawingLines(func(lines []string) bool {
		return slices.ContainsFunc(lines, func(line string) bool { return len([]rune(line)) == tooManyCells })
	}),
	"D16c": drawingHolds(strings.Repeat(" ", tooManySpaces)),
	"D16d": drawingLines(func(lines []string) bool { return strings.HasSuffix(lines[0], "  ") }),
	"D17a": drawingHolds("\u200b"),
	"D17b": drawingHolds("\t"),
	"D17c": drawingHolds("\u00a0"),
	"D17d": drawingHolds("╭"),
	"D18a": func(_ *bench, _ *submission, m *mutant) string {
		return unless(strings.Contains(m.Sub.Task.Question, strayWording), "the wording does not name the stray point")
	},
	"D18b": func(_ *bench, _ *submission, m *mutant) string {
		objects := m.Sub.Task.DrawingStructure.Objects
		return unless(objects[len(objects)-1] == checks.DrawingObject{ID: undrawnObject, Label: undrawnLabel}, "no undrawn object was declared")
	},
	"D18c": func(_ *bench, before *submission, m *mutant) string {
		renamed := 0
		for i, object := range m.Sub.Task.DrawingStructure.Objects {
			if object.Label == undrawnLabel && before.Task.DrawingStructure.Objects[i].Label != undrawnLabel {
				renamed++
			}
		}
		return unless(renamed == 1 && m.Sub.Task.Drawing == before.Task.Drawing, "not exactly one label was renamed in the structure only")
	},
	"X01a": questionChanged(func(before, after string) bool { return strings.Contains(strings.ToLower(after), "about") }),
	"X01b": questionChanged(func(before, after string) bool { return strings.Contains(after, " or ") }),
	"X02a": questionChanged(func(before, after string) bool { return len(sentencesOf(after)) == len(sentencesOf(before))-1 }),
	"X03a": questionChanged(func(before, after string) bool { return before != after }),
	"X04a": func(_ *bench, before *submission, m *mutant) string {
		changed := changedExplanations(before, &m.Sub)
		if len(changed) != 2 {
			return "not exactly two explanations changed"
		}
		first, second := changed[0], changed[1]
		swapped := m.Sub.Task.Distractors[first].Trap == before.Task.Distractors[second].Trap &&
			m.Sub.Task.Distractors[second].Trap == before.Task.Distractors[first].Trap
		return unless(swapped, "the two traps were not swapped")
	},
	"X04b": func(b *bench, before *submission, m *mutant) string {
		changed := changedExplanations(before, &m.Sub)
		return unless(len(changed) == 1 && b.shipped.HasTrap(m.Sub.Task.Distractors[changed[0]].Trap), "no explanation names another trap of the catalog")
	},
	"X05a": func(_ *bench, before *submission, m *mutant) string {
		right := strings.TrimSpace(before.Task.Options[before.Task.CorrectAnswer])
		return unless(m.Sub.Task.Hint == "The answer is "+right+".", "the hint does not say the answer")
	},
	"X06a": questionChanged(func(_, after string) bool { return strings.HasPrefix(after, pseudonym) }),
	"X07a": swappedTask(func(before *submission, m *mutant) bool { return m.Sub.Task.Question != before.Task.Question }),
	"X08a": swappedTask(func(before *submission, m *mutant) bool { return m.Sub.Task.Question != before.Task.Question }),
	"X09a": func(_ *bench, _ *submission, m *mutant) string {
		return unless(m.Source != "" && m.Sub.Task.Question != m.Source && m.Donor != "", "the copy in a new setting is not the donor's question changed")
	},
	"X09b": func(_ *bench, _ *submission, m *mutant) string {
		renamed, _ := inNewSetting(m.Source)
		return unless(m.Sub.Task.Question != renamed && m.Sub.Task.Question != m.Source, "the numbers of the copy were not changed")
	},
}

func solverHolds(line string) description {
	return func(_ *bench, _ *submission, m *mutant) string {
		return unless(strings.Contains(m.Sub.Solver, "def solve(options):\n    "+line), "the solver's body does not start with "+line)
	}
}

func explanationBecomes(text func(b *bench, before *submission, letter string) string) description {
	return func(b *bench, before *submission, m *mutant) string {
		changed := changedExplanations(before, &m.Sub)
		if len(changed) != 1 {
			return "not exactly one explanation changed"
		}
		return unless(m.Sub.Task.Distractors[changed[0]].Text == text(b, before, changed[0]), "the explanation is not what the operator makes")
	}
}

func copyOf(rewrite func(question string) string) description {
	return func(_ *bench, _ *submission, m *mutant) string {
		return unless(m.Donor != "" && m.Source != "" && m.Sub.Task.Question == rewrite(m.Source), "the question is not the donor's, rewritten")
	}
}

func historyHolds(rewrite func(question string) string) description {
	return func(_ *bench, before *submission, m *mutant) string {
		earlier := rewrite(before.Task.Question)
		return unless(slices.Equal(m.Sub.Fingerprints, []string{checks.Fingerprint(earlier, language)}) && m.Sub.Task.Question == before.Task.Question,
			"the history does not hold the sketch of the question, rewritten, or the task changed")
	}
}

func drawingLines(holds func(lines []string) bool) description {
	return func(_ *bench, _ *submission, m *mutant) string {
		return unless(holds(strings.Split(m.Sub.Task.Drawing, "\n")), "the drawing's lines are not as the operator makes them")
	}
}

func drawingHolds(text string) description {
	return func(_ *bench, before *submission, m *mutant) string {
		return unless(strings.Contains(m.Sub.Task.Drawing, text) && !strings.Contains(before.Task.Drawing, text), "the drawing does not hold what was inserted")
	}
}

func questionChanged(holds func(before, after string) bool) description {
	return func(_ *bench, before *submission, m *mutant) string {
		return unless(m.Sub.Task.Question != before.Task.Question && holds(before.Task.Question, m.Sub.Task.Question), "the question is not changed as the operator says")
	}
}

func swappedTask(holds func(before *submission, m *mutant) bool) description {
	return func(_ *bench, before *submission, m *mutant) string {
		kept := m.Sub.Brief.TargetConcept == before.Brief.TargetConcept && m.Sub.Brief.Difficulty == before.Brief.Difficulty &&
			m.Sub.Asked.TargetConcept == before.Asked.TargetConcept
		return unless(kept && holds(before, m) && m.Sub.SelfCheck.FinalAnswer == m.Sub.Task.CorrectAnswer &&
			slices.Equal(m.LeaveOut, []string{m.Sub.Task.Question}), "the task was not swapped under the host's brief")
	}
}

// checkOperators name, for each check, operators expected to be refused by it
// alone, in the order they are tried.
var checkOperators = map[checks.Code][]string{
	checks.CodeBadStructure:           {"D09a-core_idea"},
	checks.CodeDistractorExplanations: {"D12e"},
	checks.CodeDrawingFormat:          {"D16d"},
	checks.CodeDrawingMismatch:        {"D18b"},
	checks.CodeReadability:            {"D13b", "D13a"},
	checks.CodeSolverError:            {"D04a"},
	checks.CodeSolverDisagrees:        {"D07a"},
	checks.CodeSelfCheckBlocking:      {"D08a"},
	checks.CodeNearDuplicate:          {"D15a"},
}

// The checks judge on their own: a case carrying two defects, each refused by
// a different check, gets exactly the problems of the two cases that carry one
// each. That is what lets the cases a single check catches be read off the
// outcomes rather than found by switching a check off.
func TestTheChecksJudgeIndependently(t *testing.T) {
	t.Parallel()
	b := testBench(t)
	byID := map[string]*operator{}
	all := operators()
	for i := range all {
		byID[all[i].ID] = &all[i]
	}
	for i, first := range codeOrder {
		for _, second := range codeOrder[i+1:] {
			t.Run(string(first)+"+"+string(second), func(t *testing.T) {
				t.Parallel()
				if !independentSomewhere(t, b, byID, first, second) {
					t.Errorf("no host of grades 5–6 takes defects for both %s and %s", first, second)
				}
			})
		}
	}
}

// independentSomewhere tries the operators of two checks on the hosts of the
// bench until one host takes both, and checks the outcomes there. It says
// whether any host did.
func independentSomewhere(t *testing.T, b *bench, byID map[string]*operator, first, second checks.Code) bool {
	t.Helper()
	for _, firstID := range checkOperators[first] {
		for _, secondID := range checkOperators[second] {
			// The near-duplicate check is the last of codeOrder, so its operator,
			// a repeat of the child's history, always goes in second: the history
			// then holds a sketch of the question the case ends with.
			one, other := byID[firstID], byID[secondID]
			for _, h := range b.pool.hosts {
				if checkedOnHost(t, b, h, one, other) {
					return true
				}
			}
		}
	}
	return false
}

// checkedOnHost puts two defects into one host, alone and together, and checks
// the union. It says whether the host took both.
func checkedOnHost(t *testing.T, b *bench, h *host, one, other *operator) bool {
	t.Helper()
	alone, applies := one.Apply(newMaker(h, b.pool, one.ID))
	if !applies {
		return false
	}
	carried := &host{ID: h.ID, Topic: h.Topic, Level: h.Level, Difficulty: h.Difficulty, Base: alone.Sub}
	both, applies := other.Apply(newMaker(carried, b.pool, other.ID))
	if !applies {
		return false
	}
	second, applies := other.Apply(newMaker(h, b.pool, other.ID))
	if !applies {
		return false
	}
	leftOut := slices.Concat([]string{h.Base.Task.Question}, alone.LeaveOut, second.LeaveOut)
	problems := func(sub *submission) []string {
		outcome, _, err := reviewing.Judged(context.Background(), b.runner, reviewing.Without(b.shipped, leftOut...), sub)
		if err != nil {
			t.Fatalf("judge: %v", err)
		}
		found := make([]string, 0, len(outcome.Problems))
		for _, problem := range outcome.Problems {
			found = append(found, string(problem.Code)+": "+problem.Message)
		}
		slices.Sort(found)
		return found
	}
	want := slices.Compact(slices.Sorted(slices.Values(slices.Concat(problems(&alone.Sub), problems(&second.Sub)))))
	if got := problems(&both.Sub); !slices.Equal(got, want) {
		t.Errorf("%s and %s on %s: together %q, want the two alone %q", one.ID, other.ID, h.ID, got, want)
	}
	return true
}

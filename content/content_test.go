package content_test

import (
	"encoding/json"
	"regexp"
	"slices"
	"strings"
	"testing"

	"github.com/MathTrail/mathtrail-standalone/content"
)

// The catalogs are closed lists: adding to one is a change to the spec and to
// every profile written before it, so the tests name every id rather than
// counting them.
var (
	wantTopics = []string{
		"logic.ordering",
		"logic.knights_liars",
		"combinatorics.enumeration",
		"counting.gaps",
		"time.clocks",
		"time.calendar",
		"pigeonhole.basic",
		"parity.alternation",
		"arithmetic.tricks",
		"algorithms.weighing_pouring",
		"fractions.parts",
		"percent.basic",
		"ratio.sharing",
		"geometry.grid",
		"number.divisibility",
		"logic.sets",
		"games.strategy",
	}
	wantTraps = []string{
		"off_by_one",
		"missed_case",
		"double_count",
		"ignored_not",
		"answered_other_question",
		"stopped_early",
		"wrong_operation",
		"reversed_relation",
		"best_case_not_worst",
		"trusted_statement",
		"time_unit_mixup",
		"number_from_text",
		"wrong_parity",
		"ignored_condition",
		"percent_wrong_base",
		"part_whole_swap",
		"ratio_total_confusion",
		"remainder_vs_quotient",
		"area_perimeter_swap",
		"first_move_assumed",
	}
	wantSkills = []string{
		"addition",
		"subtraction",
		"multiplication",
		"division",
		"division_with_remainder",
		"fractions",
		"order_of_operations",
		"even_odd",
		"numbers_above_20",
		"numbers_above_100",
		"numbers_above_1000",
		"hours_minutes",
		"calendar_months",
		"money",
		"unit_conversion",
		"fractions_arithmetic",
		"decimals",
		"percentages",
		"ratios",
		"negative_numbers",
		"powers_squares",
		"area_perimeter",
		"coordinates",
		"averages",
		"divisibility_rules",
	}
)

// loaded is the content of the binary, read once for every test that only reads
// it.
func loaded(t *testing.T) *content.Content {
	t.Helper()

	c, err := content.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	return c
}

func TestCatalogsAreTheClosedLists(t *testing.T) {
	t.Parallel()
	c := loaded(t)

	got := make([]string, 0, len(c.Topics()))
	for _, topic := range c.Topics() {
		got = append(got, topic.ID)
	}
	if !slices.Equal(got, wantTopics) {
		t.Errorf("topics:\ngot  %v\nwant %v", got, wantTopics)
	}

	got = got[:0]
	for _, trap := range c.Traps() {
		got = append(got, trap.ID)
	}
	if !slices.Equal(got, wantTraps) {
		t.Errorf("traps:\ngot  %v\nwant %v", got, wantTraps)
	}

	got = got[:0]
	for _, skill := range c.Skills() {
		got = append(got, skill.ID)
	}
	if !slices.Equal(got, wantSkills) {
		t.Errorf("skills:\ngot  %v\nwant %v", got, wantSkills)
	}
}

func TestEveryTopicIsOfferedToTheOldestChildren(t *testing.T) {
	t.Parallel()

	for _, topic := range loaded(t).Topics() {
		if !topic.HasLevel(content.Level56) {
			t.Errorf("topic %s is offered at %v, want %s among them", topic.ID, topic.GradeLevels, content.Level56)
		}
	}
}

func TestLookupFindsAnEntryAndMissesWhatIsNotThere(t *testing.T) {
	t.Parallel()
	c := loaded(t)

	topic, ok := c.Topic("counting.gaps")
	if !ok || topic.Name == "" {
		t.Errorf("Topic(counting.gaps) = %+v, %v; want the catalog entry", topic, ok)
	}
	if _, ok := c.Topic("patterns.sequences"); ok {
		t.Error("Topic(patterns.sequences) found a topic that is deliberately not in the catalog")
	}
	if _, ok := c.Trap("off_by_one"); !ok {
		t.Error("Trap(off_by_one) found nothing")
	}
	if _, ok := c.Trap("forgot_a_case"); ok {
		t.Error("Trap(forgot_a_case) found a trap nobody wrote")
	}
	if _, ok := c.Skill("division_with_remainder"); !ok {
		t.Error("Skill(division_with_remainder) found nothing")
	}
	if _, ok := c.Skill("calculus"); ok {
		t.Error("Skill(calculus) found a skill nobody wrote")
	}
}

// Grades 1-4 carry five reference tasks for every topic, level and difficulty.
// Three of them go into the package for the model, and the rotation that stops
// a child seeing the same examples twice has nothing to rotate without the
// other two.
func TestEveryCellOfGradesOneToFourHasFiveTasks(t *testing.T) {
	t.Parallel()
	c := loaded(t)

	type cell struct {
		topic      string
		level      string
		difficulty int
	}
	count := map[cell]int{}
	for _, example := range c.Examples() {
		count[cell{example.Topic, example.GradeLevel, example.Difficulty}]++
	}

	const want = 5
	for _, topic := range c.Topics() {
		for _, level := range []string{content.Level12, content.Level34} {
			if !topic.HasLevel(level) {
				continue
			}
			for difficulty := 1; difficulty <= 5; difficulty++ {
				key := cell{topic.ID, level, difficulty}
				if got := count[key]; got != want {
					t.Errorf("%s at %s, difficulty %d: got %d reference tasks, want %d",
						key.topic, key.level, key.difficulty, got, want)
				}
			}
		}
	}
}

func TestSchemasAreInTheBinary(t *testing.T) {
	t.Parallel()
	c := loaded(t)

	for _, name := range []string{"brief.json", "task.json", "self_check.json"} {
		raw, ok := c.Schema(name)
		if !ok {
			t.Errorf("Schema(%s) found nothing", name)
			continue
		}
		var document map[string]any
		if err := json.Unmarshal(raw, &document); err != nil {
			t.Errorf("Schema(%s) is not a JSON document: %v", name, err)
		}
	}
	if _, ok := c.Schema("profile.json"); ok {
		t.Error("Schema(profile.json) found a schema this service does not have")
	}
}

func TestInstructionsAreInTheBinary(t *testing.T) {
	t.Parallel()
	c := loaded(t)

	for _, name := range []string{"mcp_instructions.md", "task_writing.md"} {
		text, ok := c.Instruction(name)
		if !ok || text == "" {
			t.Errorf("Instruction(%s) = %q, %v; want the text of the file", name, text, ok)
		}
	}
	if _, ok := c.Instruction("play_session.md"); ok {
		t.Error("Instruction(play_session.md) found a file this service does not ship")
	}
}

// tools are the six tools the server offers, named as the server names them.
var tools = []string{"get_profile", "save_profile", "get_progress", "next_task", "submit_task", "submit_answer"}

// The instructions the server hands the model walk it through every tool it
// offers, each by the name the server gives it. A host may put its own prefix
// before that name, and nothing in them depends on its being absent.
func TestTheServerInstructionsNameEveryTool(t *testing.T) {
	t.Parallel()

	text, _ := loaded(t).Instruction("mcp_instructions.md")
	for _, tool := range tools {
		if !strings.Contains(text, "`"+tool+"`") {
			t.Errorf("the server instructions never name %s", tool)
		}
	}
}

// Nothing of the prototype's instructions comes back with them: no tool takes a
// student id, since the profile is the one the token belongs to, and the
// prototype's own tools are gone.
func TestTheServerInstructionsNameNothingThePrototypeHad(t *testing.T) {
	t.Parallel()

	text, _ := loaded(t).Instruction("mcp_instructions.md")
	for _, gone := range []string{"student_id", "get_next_task", "get_student_profile", "solver_code", "taskgen"} {
		if strings.Contains(text, gone) {
			t.Errorf("the server instructions still name %s", gone)
		}
	}
}

// The version names one wording among all the wordings this service will ever
// have, and it is written into a log line, so it stays short and stays the same
// for the same content.
func TestInstructionsVersionIsShortAndSteady(t *testing.T) {
	t.Parallel()

	version := loaded(t).InstructionsVersion()
	if !regexp.MustCompile(`^[0-9a-f]{12}$`).MatchString(version) {
		t.Errorf("got version %q, want twelve hexadecimal characters", version)
	}
	if again := loaded(t).InstructionsVersion(); again != version {
		t.Errorf("got version %q on the second read, want %q", again, version)
	}
}

// A caller that sorts or appends to what it was handed must not change what the
// next caller is handed.
func TestWhatIsHandedOutIsACopy(t *testing.T) {
	t.Parallel()
	c := loaded(t)

	topics := c.Topics()
	topics[0] = content.Topic{ID: "edited.by.a.caller"}
	if again := c.Topics(); again[0].ID == "edited.by.a.caller" {
		t.Error("editing the topics handed out changed the catalog")
	}

	// The levels of a topic are a slice inside a copied value, which is the
	// kind of sharing a copy of the outer slice does nothing about.
	levels := c.Topics()[0].GradeLevels
	levels[0] = "9-10"
	if again := c.Topics(); again[0].GradeLevels[0] == "9-10" {
		t.Error("editing the levels of a topic handed out changed the catalog")
	}

	// A reference task holds its options in a map, and a map is shared however
	// many times the task around it is copied.
	example := c.Examples()[0]
	example.Options["A"] = "edited by a caller"
	example.Distractors["C"] = content.Distractor{Trap: "off_by_one", Text: "edited by a caller"}
	if again := c.Examples()[0]; again.Options["A"] == "edited by a caller" {
		t.Error("editing the options of a reference task handed out changed the content")
	}
	if again := c.Examples()[0]; again.Distractors["C"].Text == "edited by a caller" {
		t.Error("editing the explanations of a reference task handed out changed the content")
	}

	schema, _ := c.Schema("task.json")
	schema[0] = ' '
	if again, _ := c.Schema("task.json"); again[0] == ' ' {
		t.Error("editing the schema handed out changed the schema")
	}
}

// The count is what the startup log reports, and a number that drifts from the
// tasks behind it is a lie in the one place a deployment is inspected from.
func TestExampleCountMatchesTheTasks(t *testing.T) {
	t.Parallel()
	c := loaded(t)

	if got, want := c.ExampleCount(), len(c.Examples()); got != want {
		t.Errorf("ExampleCount() = %d, want %d", got, want)
	}
}

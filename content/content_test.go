package content_test

import (
	"encoding/json"
	"regexp"
	"slices"
	"strings"
	"testing"
	"unicode/utf16"

	"github.com/MathTrail/mathtrail-standalone/content"
	"github.com/MathTrail/mathtrail-standalone/internal/domain/profile"
	"github.com/MathTrail/mathtrail-standalone/internal/domain/rating"
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

// Every trap of the catalog says what an adult can do about it, in a sentence
// of its own: the review of the progress hands it on as it is written.
func TestEveryTrapAdvisesTheAdult(t *testing.T) {
	t.Parallel()

	c := loaded(t)
	for _, id := range c.TrapIDs() {
		advice, ok := c.TrapAdvice(id)
		if !ok || advice == "" || !strings.HasSuffix(advice, ".") {
			t.Errorf("TrapAdvice(%s) = %q, %v; want a sentence of advice", id, advice, ok)
		}
		if description, _ := c.TrapDescription(id); advice == description {
			t.Errorf("TrapAdvice(%s) = %q, the mistake again; want what to do about it", id, advice)
		}
	}
	want := "Before answering, draw a quick sketch: the posts as dots and the gaps between them, then count both."
	if advice, _ := c.TrapAdvice("off_by_one"); advice != want {
		t.Errorf("TrapAdvice(off_by_one) = %q, want %q", advice, want)
	}
	if advice, ok := c.TrapAdvice("forgot_a_case"); ok || advice != "" {
		t.Errorf("TrapAdvice(forgot_a_case) = %q, %v; want nothing for a trap nobody wrote", advice, ok)
	}
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

// A parent may leave every skill of the catalog out of the tasks, and no more:
// the cap of the profile is the size of the catalog. A catalog that grows
// without the cap would leave some child's skills unsayable.
func TestAProfileMayLeaveOutEverySkillOfTheCatalog(t *testing.T) {
	t.Parallel()

	if got := len(loaded(t).Skills()); got != profile.MaxExcludedSkills {
		t.Errorf("the catalog has %d skills and a profile may leave out %d, want the two the same",
			got, profile.MaxExcludedSkills)
	}
}

func TestEveryTopicIsOfferedToTheOldestChildren(t *testing.T) {
	t.Parallel()

	for _, topic := range loaded(t).Topics() {
		if !topic.HasLevel(rating.Grades56) {
			t.Errorf("topic %s is offered at %v, want %s among them", topic.ID, topic.GradeLevels, rating.Grades56)
		}
	}
}

// A topic builds only on topics a child can have met by the time it is taught:
// a base first taught in a later grade would point a younger child at a topic
// they cannot be set yet.
func TestATopicBuildsOnlyOnTopicsTaughtNoLater(t *testing.T) {
	t.Parallel()
	c := loaded(t)

	for _, topic := range c.Topics() {
		for _, id := range topic.BuildsOn {
			base, ok := c.Topic(id)
			if !ok {
				t.Fatalf("topic %s builds on %s, which is not in the catalog", topic.ID, id)
			}
			if firstGrade(base.GradeLevels) > firstGrade(topic.GradeLevels) {
				t.Errorf("topic %s is first taught in grade %d and builds on %s, first taught in grade %d",
					topic.ID, firstGrade(topic.GradeLevels), base.ID, firstGrade(base.GradeLevels))
			}
		}
	}
}

// The map of topics on the site has three layers — foundations, techniques and
// hard problems — and a topic's layer is the longest chain of bases below it. A
// link that made a chain longer would need a layer the map does not draw.
func TestTheMapOfTopicsHasThreeLayers(t *testing.T) {
	t.Parallel()

	topics := loaded(t).Topics()
	layers := layersOf(topics)
	sizes := make([]int, 3)
	for _, topic := range topics {
		layer := layers[topic.ID]
		if layer >= len(sizes) {
			t.Errorf("topic %s is in layer %d, want one of the map's %d", topic.ID, layer, len(sizes))
			continue
		}
		sizes[layer]++
	}
	t.Logf("topics in each layer, foundations first: %v", sizes)
}

// firstGrade is the youngest school year any of these levels is taught in.
func firstGrade(levels []rating.GradeLevel) int {
	grade := 0
	for _, level := range levels {
		if first := level.FirstGrade(); grade == 0 || first < grade {
			grade = first
		}
	}
	return grade
}

// layersOf works out every topic's layer: 0 for a topic that builds on none,
// and one more than its highest base otherwise. The loader has refused a
// circle, so the walk ends.
func layersOf(topics []content.Topic) map[string]int {
	byID := make(map[string]content.Topic, len(topics))
	for _, topic := range topics {
		byID[topic.ID] = topic
	}
	layers := make(map[string]int, len(topics))
	var layerOf func(id string) int
	layerOf = func(id string) int {
		if layer, ok := layers[id]; ok {
			return layer
		}
		layer := 0
		for _, base := range byID[id].BuildsOn {
			layer = max(layer, layerOf(base)+1)
		}
		layers[id] = layer
		return layer
	}
	for _, topic := range topics {
		layerOf(topic.ID)
	}
	return layers
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
		level      rating.GradeLevel
		difficulty int
	}
	count := map[cell]int{}
	for _, example := range c.Examples() {
		count[cell{example.Topic, example.GradeLevel, example.Difficulty}]++
	}

	const want = 5
	for _, topic := range c.Topics() {
		for _, level := range []rating.GradeLevel{rating.Grades12, rating.Grades34} {
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

// Grades 5–6 carry nine reference tasks for a topic, three at each of the
// middle difficulties: the rule aims at the middle of the corridor, and a
// request for difficulty 1 or 5 is shown the nearest of them, labelled with its
// own difficulty. The tasks arrive topic by topic, so a topic with none yet is
// named rather than refused.
func TestEveryTopicOfGradesFiveAndSixHasThreeTasksAtEachMiddleDifficulty(t *testing.T) {
	t.Parallel()
	c := loaded(t)

	count := tasksByTopicAndDifficulty(c, rating.Grades56)
	var waiting []string
	for _, topic := range c.Topics() {
		tasks, written := count[topic.ID]
		if !written {
			waiting = append(waiting, topic.ID)
			continue
		}
		for difficulty := 1; difficulty <= 5; difficulty++ {
			want := 0
			if difficulty >= 2 && difficulty <= 4 {
				want = 3
			}
			if got := tasks[difficulty]; got != want {
				t.Errorf("%s at %s, difficulty %d: got %d reference tasks, want %d",
					topic.ID, rating.Grades56, difficulty, got, want)
			}
		}
	}
	t.Logf("no reference tasks at %s yet: %v", rating.Grades56, waiting)
}

// tasksByTopicAndDifficulty counts the reference tasks of one level, by topic
// and then by difficulty. A topic with none at the level is not in it.
func tasksByTopicAndDifficulty(c *content.Content, level rating.GradeLevel) map[string]map[int]int {
	count := map[string]map[int]int{}
	examples := c.Examples()
	for i := range examples {
		example := &examples[i]
		if example.GradeLevel != level {
			continue
		}
		if count[example.Topic] == nil {
			count[example.Topic] = map[int]int{}
		}
		count[example.Topic][example.Difficulty]++
	}
	return count
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

// tools are the tools the model sees, named as the server names them. A tool
// only a card calls is not among them: the model is never told of it.
var tools = []string{
	"get_profile", "save_profile", "get_progress", "next_task", "get_package", "prepare_task", "submit_task", "submit_answer",
}

// The instructions the server hands the model walk it through every tool it
// offers, each by the name the server gives it. A host may put its own prefix
// before that name, and nothing in them depends on its being absent.
func TestTheServerInstructionsNameEveryTool(t *testing.T) {
	t.Parallel()

	text := loaded(t).ServerInstructions()
	for _, tool := range tools {
		if !strings.Contains(text, "`"+tool+"`") {
			t.Errorf("the server instructions never name %s", tool)
		}
	}
}

// hostKeeps is as much of a server's instructions as a host is known to pass on
// to the model: Claude Code keeps the first 2,048 characters, counted as
// JavaScript counts them, and cuts the rest.
const hostKeeps = 2048

// A host that cuts the instructions keeps their beginning. What must hold in
// every lesson therefore stands first, in a section of its own, and ends within
// what such a host keeps; the sections after it say how, and a host may lose
// them.
func TestWhatAlwaysHoldsIsKeptByAHostThatCuts(t *testing.T) {
	t.Parallel()

	always := throughAlways(t, loaded(t).ServerInstructions())
	if kept := len(utf16.Encode([]rune(always))); kept > hostKeeps {
		t.Errorf(`the server instructions up to the end of "## Always" are %d characters long, and a host may keep only the first %d`,
			kept, hostKeeps)
	}
}

// throughAlways is the server instructions from their start to the end of
// "## Always", which must be their first section.
func throughAlways(t *testing.T, text string) string {
	t.Helper()

	first := strings.Index(text, "\n## ")
	if first < 0 || !strings.HasPrefix(text[first+1:], "## Always\n") {
		t.Fatal(`the first section of the server instructions is not "## Always"`)
	}
	if next := strings.Index(text[first+1:], "\n## "); next >= 0 {
		return text[:first+1+next]
	}
	return text
}

// hostReadsFirst is as much of a server's instructions as ChatGPT asks to hold
// what matters most, counted as JavaScript counts characters.
const hostReadsFirst = 512

// Who runs the lesson is said before anything else: the adult types every
// message, and the child, beside them, answers on the card and never in the
// chat, so the model talks with the adult.
func TestTheInstructionsOpenWithWhoRunsTheLesson(t *testing.T) {
	t.Parallel()

	units := utf16.Encode([]rune(loaded(t).ServerInstructions()))
	opening := string(utf16.Decode(units[:min(len(units), hostReadsFirst)]))
	for _, says := range []string{
		"An adult, a parent or a tutor, runs the lesson",
		"types every message",
		"answers each task on its card",
		"never types in the chat",
		"you talk with the adult",
	} {
		if !strings.Contains(opening, says) {
			t.Errorf("the first %d characters of the server instructions do not say %q", hostReadsFirst, says)
		}
	}
}

// Every rule a lesson keeps whatever the host stands in "## Always", where a
// host that cuts the instructions still finds it. Each rule is found by the
// words that carry it.
func TestAlwaysSaysEveryRuleOfALesson(t *testing.T) {
	t.Parallel()

	always := throughAlways(t, loaded(t).ServerInstructions())
	for _, rule := range []struct{ that, says string }{
		{"the model talks to the adult", "Talk to the adult"},
		{"what the child hears is worded for the adult to read out", "for the adult to read out"},
		{"nothing shows whether the child is a boy or a girl", "whether the child is a boy or a girl"},
		{"not even a pseudonym that has a gender of its own", "not even the pseudonym"},
		{"the child is spoken of by the pseudonym and never as he or she", "speak of the child by the pseudonym, never as he or she"},
		{"the child is known by a pseudonym", "known by a pseudonym alone"},
		{"the answer waits for the child's", "keep the answer, the solution and the explanations to yourself"},
		{"the answer waits even when the adult asks for it", "even when the adult asks"},
		{"the hint comes only when asked for", "give the hint only when asked"},
		{"a task on the card gets no word of the model's own", "say nothing of a task on the card"},
		{"the adult gives the child's answer in the chat", "When the adult gives the child's answer in the chat"},
		{"an answer is recorded before it is explained", "before you explain anything"},
		{`"I don't know" is an answer`, "\"I don't know\" is the answer `?`"},
		{"a wrong answer is explained from its trap", "from its trap's text"},
		{"the adult decides when the next task comes", "The adult decides when the next task comes"},
		{"the last answer is read from a tool", "read the last recorded answer in its result"},
		{"MathTrail's tools are for its own tasks alone", "never for homework, another subject or a task past grade 6"},
	} {
		if !strings.Contains(always, rule.says) {
			t.Errorf(`"## Always" does not say that %s: it has no %q`, rule.that, rule.says)
		}
	}
}

// The tools' descriptions say what each tool does and leave how to behave to
// the instructions, so the rules a description no longer carries are said
// there.
func TestTheInstructionsSayWhatTheDescriptionsLeaveOut(t *testing.T) {
	t.Parallel()

	text := loaded(t).ServerInstructions()
	for _, rule := range []struct{ that, says string }{
		{"a step back is told gently", "say a step back gently"},
		{"the model says the rating's number", "say the number yourself"},
		{"the grades under the ranks are no school mark", "never a school mark or a verdict on the child"},
		{"a topic's page is passed on as the result names it", "pass a page on only as the result names it"},
		{"the progress is told encouragingly", "Present it encouragingly"},
		{"the country is never asked for", "never ask for them"},
		{"homework is not MathTrail's", "MathTrail is not for homework"},
	} {
		if !strings.Contains(text, rule.says) {
			t.Errorf("the server instructions do not say that %s: they have no %q", rule.that, rule.says)
		}
	}
}

// Nothing of the prototype's instructions comes back with them: no tool takes a
// student id, since the profile is the one the token belongs to, and the
// prototype's own tools are gone.
func TestTheServerInstructionsNameNothingThePrototypeHad(t *testing.T) {
	t.Parallel()

	text := loaded(t).ServerInstructions()
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

	// So are the topics a topic builds on, handed out with the topic or alone.
	knights, _ := c.Topic("logic.knights_liars")
	knights.BuildsOn[0] = "edited.by.a.caller"
	if again, _ := c.Topic("logic.knights_liars"); again.BuildsOn[0] == "edited.by.a.caller" {
		t.Error("editing the bases of a topic handed out changed the catalog")
	}
	bases := c.BasesOf("logic.knights_liars")
	bases[0] = "edited.by.a.caller"
	if again := c.BasesOf("logic.knights_liars"); again[0] == "edited.by.a.caller" {
		t.Error("editing the bases handed out alone changed the catalog")
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

	templates := c.Templates("logic.ordering")
	templates[0].Program = "edited by a caller"
	if again := c.Templates("logic.ordering"); again[0].Program == "edited by a caller" {
		t.Error("editing a solver template handed out changed the content")
	}

	// An example of a kind of picture holds its topics and its picture behind
	// references, like a reference task holds its options.
	examples := c.PictureExamples()
	examples[0].Topics[0] = "edited.by.a.caller"
	examples[0].Picture[0] = 'X'
	if again := c.PictureExamples()[0]; again.Topics[0] == "edited.by.a.caller" || again.Picture[0] == 'X' {
		t.Error("editing an example of a picture handed out changed the content")
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

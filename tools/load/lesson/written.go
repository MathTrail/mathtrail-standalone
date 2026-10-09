package lesson

import "encoding/json"

// Task is a task as the chat's model writes it, with what it was asked for and
// a letter the child gets it wrong with.
type Task struct {
	// Choice is the topic, level and difficulty the task is written for.
	Choice Choice
	// Body is the task itself.
	Body Body
	// Solver is the program that proves its answer.
	Solver string
	// SelfCheck is the model's own check of the task.
	SelfCheck SelfCheck
	// Wrong is a letter of a wrong option.
	Wrong string
}

// Body is a task in the shape the model hands it in.
type Body struct {
	CoreIdea    string                `json:"core_idea"`
	Question    string                `json:"question"`
	Picture     json.RawMessage       `json:"picture,omitempty"`
	Options     map[string]string     `json:"options"`
	Correct     string                `json:"correct_answer"`
	Hint        string                `json:"hint"`
	Solution    string                `json:"solution"`
	Distractors map[string]Distractor `json:"distractors"`
}

// Distractor is a wrong option: the trap it is for, and what the child who
// chose it is told.
type Distractor struct {
	Trap string `json:"trap"`
	Text string `json:"text"`
}

// SelfCheck is the model's own check of a task: what it thinks of each option,
// the answer it arrives at, and the issues it found.
type SelfCheck struct {
	Issues      []string          `json:"issues"`
	OptionCheck map[string]string `json:"option_check"`
	FinalAnswer string            `json:"final_answer"`
}

// ordering is what every written task is asked for: ordering, at the first
// level, of a middling difficulty.
var ordering = Choice{
	Topic:      "logic.ordering",
	GradeLevel: "1-2",
	Difficulty: 2,
	Reason:     "The load hands in tasks it has written itself, all on ordering.",
}

// The traps the written tasks use, each a wrong option's.
const (
	reversed = "reversed_relation"
	stopped  = "stopped_early"
	ignored  = "ignored_condition"
	other    = "answered_other_question"
)

// Written are the tasks a lesson of the load hands in, in the order it hands
// them in. Each is a task a model could have written — its question, five
// options, the explanations behind the four wrong ones, a solver that proves
// the answer and a self-check — and each tells a story of its own, so that
// none of them is a near-copy of another or of a reference task of its level,
// and one child can be handed all of them in a row.
func Written() []Task {
	return []Task{race(), heights(), ages(), queue(), cats()}
}

func race() Task {
	return Task{
		Choice: ordering,
		Body: Body{
			CoreIdea: "Order three runners from two comparisons.",
			Question: "Ann, Ben and Kim ran a race. Ben finished before Kim. Ann finished after Kim. Who finished first?",
			Picture:  json.RawMessage(`{"kind":"row","items":[{"label":"1"},{"label":"2"},{"label":"3"}]}`),
			Options:  map[string]string{"A": "Ann", "B": "Kim", "C": "Ben", "D": "Nobody", "E": "All at once"},
			Correct:  "C",
			Hint:     "Who finished before Kim?",
			Solution: "Ben is ahead of Kim and Kim is ahead of Ann, so Ben crosses the line first.",
			Distractors: map[string]Distractor{
				"A": {Trap: reversed, Text: "Ann finished after Kim, so she came last."},
				"B": {Trap: stopped, Text: "Kim sits in the middle: Ben beat her."},
				"D": {Trap: ignored, Text: "In any race somebody always finishes first."},
				"E": {Trap: other, Text: "The three finished one after another."},
			},
		},
		Solver: `def solve(options):
    firsts = []
    for order in permutations(["Ann", "Ben", "Kim"]):
        place = {name: i for i, name in enumerate(order)}
        if place["Ben"] < place["Kim"] and place["Kim"] < place["Ann"]:
            firsts.append(order[0])
    return match(options, firsts[0])
`,
		SelfCheck: SelfCheck{
			Issues: []string{},
			OptionCheck: map[string]string{
				"A": "Ann is last.", "B": "Kim is second.", "C": "Ben is first.", "D": "Someone was first.", "E": "Nobody tied.",
			},
			FinalAnswer: "C",
		},
		Wrong: "A",
	}
}

func heights() Task {
	return Task{
		Choice: ordering,
		Body: Body{
			CoreIdea: "Find the tallest of three children from two comparisons.",
			Question: "Bob, Gus and Hal stand in a row. Bob is taller than Gus. Hal is shorter than Gus. Who is the tallest?",
			Options:  map[string]string{"A": "Gus", "B": "Hal", "C": "Bob", "D": "All the same", "E": "Nobody"},
			Correct:  "C",
			Hint:     "Who is taller than Gus?",
			Solution: "Bob is taller than Gus, and Gus is taller than Hal. So Bob is the tallest.",
			Distractors: map[string]Distractor{
				"A": {Trap: stopped, Text: "Gus is only in the middle of the three."},
				"B": {Trap: reversed, Text: "Hal is the shortest of the three."},
				"D": {Trap: ignored, Text: "Bob and Gus are not the same height."},
				"E": {Trap: other, Text: "One of three children is always the tallest."},
			},
		},
		Solver: `def solve(options):
    tallest = []
    for order in permutations(["Bob", "Gus", "Hal"]):
        place = {name: i for i, name in enumerate(order)}
        if place["Bob"] < place["Gus"] and place["Gus"] < place["Hal"]:
            tallest.append(order[0])
    return match(options, tallest[0])
`,
		SelfCheck: SelfCheck{
			Issues: []string{},
			OptionCheck: map[string]string{
				"A": "Gus is in the middle.", "B": "Hal is the shortest.", "C": "Bob is the tallest.",
				"D": "They differ in height.", "E": "Someone is the tallest.",
			},
			FinalAnswer: "C",
		},
		Wrong: "B",
	}
}

func ages() Task {
	return Task{
		Choice: ordering,
		Body: Body{
			CoreIdea: "Find the youngest of three friends from two comparisons.",
			Question: "Liz is older than Pam. Pam is older than Dan. Who is the youngest?",
			Options:  map[string]string{"A": "Liz", "B": "Pam", "C": "Dan", "D": "Liz and Dan", "E": "They are twins"},
			Correct:  "C",
			Hint:     "Who is younger than Pam?",
			Solution: "Liz is older than Pam, and Pam is older than Dan. So Dan is the youngest.",
			Distractors: map[string]Distractor{
				"A": {Trap: reversed, Text: "Liz is the oldest of the three friends."},
				"B": {Trap: stopped, Text: "Pam is younger than Liz, yet Dan is younger still."},
				"D": {Trap: other, Text: "Only one of the friends can be the youngest."},
				"E": {Trap: ignored, Text: "The three friends are all of different ages."},
			},
		},
		Solver: `def solve(options):
    youngest = []
    for order in permutations(["Liz", "Pam", "Dan"]):
        age = {name: i for i, name in enumerate(order)}
        if age["Liz"] < age["Pam"] and age["Pam"] < age["Dan"]:
            youngest.append(order[2])
    return match(options, youngest[0])
`,
		SelfCheck: SelfCheck{
			Issues: []string{},
			OptionCheck: map[string]string{
				"A": "Liz is the oldest.", "B": "Pam is in the middle.", "C": "Dan is the youngest.",
				"D": "Two cannot both be youngest.", "E": "Their ages differ.",
			},
			FinalAnswer: "C",
		},
		Wrong: "A",
	}
}

func queue() Task {
	return Task{
		Choice: ordering,
		Body: Body{
			CoreIdea: "Place three children in a line from two clues.",
			Question: "Joe, Kit and Ned wait in a line. Kit stands behind Joe. Ned stands in front of Joe. Who is at the back?",
			Options:  map[string]string{"A": "Joe", "B": "Ned", "C": "Kit", "D": "Nobody", "E": "Joe and Kit"},
			Correct:  "C",
			Hint:     "Who stands behind Joe?",
			Solution: "Ned is in front of Joe, and Kit is behind Joe. So Kit is at the back.",
			Distractors: map[string]Distractor{
				"A": {Trap: stopped, Text: "Joe stands between Ned and Kit."},
				"B": {Trap: reversed, Text: "Ned is at the front of the line."},
				"D": {Trap: ignored, Text: "Someone has to be last in a line."},
				"E": {Trap: other, Text: "Only one child can stand at the very back."},
			},
		},
		Solver: `def solve(options):
    backs = []
    for order in permutations(["Joe", "Kit", "Ned"]):
        place = {name: i for i, name in enumerate(order)}
        if place["Joe"] < place["Kit"] and place["Ned"] < place["Joe"]:
            backs.append(order[2])
    return match(options, backs[0])
`,
		SelfCheck: SelfCheck{
			Issues: []string{},
			OptionCheck: map[string]string{
				"A": "Joe is in the middle.", "B": "Ned is at the front.", "C": "Kit is at the back.",
				"D": "Someone is last.", "E": "Only Kit is last.",
			},
			FinalAnswer: "C",
		},
		Wrong: "B",
	}
}

func cats() Task {
	return Task{
		Choice: ordering,
		Body: Body{
			CoreIdea: "Find the middle of three places from two clues about left and right.",
			Question: "Three cats sleep on a sofa. The black cat lies to the left of the white cat. The grey cat lies to the right of the white cat. Which cat lies in the middle?",
			Options:  map[string]string{"A": "Black", "B": "White", "C": "Grey", "D": "None of them", "E": "All three"},
			Correct:  "B",
			Hint:     "Which cat has a cat on each side?",
			Solution: "The black cat is on the left of the white cat, and the grey cat is on its right. So the white cat is in the middle.",
			Distractors: map[string]Distractor{
				"A": {Trap: reversed, Text: "The black cat lies at the left end."},
				"C": {Trap: stopped, Text: "The grey cat lies at the right end."},
				"D": {Trap: ignored, Text: "With three cats in a row, one of them is in the middle."},
				"E": {Trap: other, Text: "Only one cat can lie in the middle."},
			},
		},
		Solver: `def solve(options):
    middles = []
    for order in permutations(["Black", "White", "Grey"]):
        place = {cat: i for i, cat in enumerate(order)}
        if place["Black"] < place["White"] and place["White"] < place["Grey"]:
            middles.append(order[1])
    return match(options, middles[0])
`,
		SelfCheck: SelfCheck{
			Issues: []string{},
			OptionCheck: map[string]string{
				"A": "Black is at the left end.", "B": "White is in the middle.", "C": "Grey is at the right end.",
				"D": "One cat is in the middle.", "E": "Only one is in the middle.",
			},
			FinalAnswer: "B",
		},
		Wrong: "A",
	}
}

// Refused is a task the checks refuse every time it is handed in, as a model
// that got its own answer wrong would hand it in: the race, with the answer,
// the self-check and the explanations all pointing at an option its solver
// proves wrong.
func Refused() Task {
	task := race()
	right, wrong := task.Body.Correct, task.Wrong
	task.Body.Correct = wrong
	task.Body.Distractors[right] = task.Body.Distractors[wrong]
	delete(task.Body.Distractors, wrong)
	task.SelfCheck.FinalAnswer = wrong
	task.Wrong = right
	return task
}

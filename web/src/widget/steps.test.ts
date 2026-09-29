import { afterEach, describe, expect, test, vi } from "vitest";
import { stepsOf } from "./steps";
import { fenceSolution, fenceSolutionInRussian } from "./testing/lesson";

afterEach(() => {
	vi.unstubAllGlobals();
});

describe("a solution's steps", () => {
	test.each([
		[
			"the fence, in English",
			fenceSolution,
			"en",
			[
				"12 ÷ 3 = 4 gaps.",
				"A straight fence with posts at both ends has one more post than gaps.",
				"4 + 1 = 5 posts.",
			],
		],
		[
			"the fence, in Russian",
			fenceSolutionInRussian,
			"ru",
			[
				"12 : 3 = 4 промежутка.",
				"У прямого забора со столбами на обоих концах столбов на один больше, чем промежутков.",
				"4 + 1 = 5 столбов.",
			],
		],
		[
			"a sentence after a number with a word in lowercase after it",
			"The jug takes 7 litres. 10 - 7 = 3 litres are left in the bucket.",
			"en",
			["The jug takes 7 litres.", "10 - 7 = 3 litres are left in the bucket."],
		],
		[
			"a sum written with an ellipsis",
			"1 + 2 + ... + 12 = 78. Pair the numbers: 1 + 12, 2 + 11 and so on.",
			"en",
			["1 + 2 + ... + 12 = 78.", "Pair the numbers: 1 + 12, 2 + 11 and so on."],
		],
		[
			"a difference written with an ellipsis",
			"100 - 1 - 2 - ... - 10 = 45. So the answer is 45.",
			"en",
			["100 - 1 - 2 - ... - 10 = 45.", "So the answer is 45."],
		],
		[
			"a quotient with colons, as Russian writes one",
			"60 : 2 : ... : 3 = 1. Значит, ответ 1.",
			"ru",
			["60 : 2 : ... : 3 = 1.", "Значит, ответ 1."],
		],
		[
			"a sentence that opens with a negative number",
			"5 - 8 = -3. -3 is less than 0.",
			"en",
			["5 - 8 = -3.", "-3 is less than 0."],
		],
		[
			"a difference written with an ellipsis and the minus sign",
			"100 − 1 − 2 − ... − 10 = 45. So the answer is 45.",
			"en",
			["100 − 1 − 2 − ... − 10 = 45.", "So the answer is 45."],
		],
		[
			"a sentence that opens with a negative number in the minus sign",
			"5 − 8 = −3. −3 is less than 0.",
			"en",
			["5 − 8 = −3.", "−3 is less than 0."],
		],
		[
			"a decimal number",
			"The bag holds 2.5 kg. Then it is full.",
			"en",
			["The bag holds 2.5 kg.", "Then it is full."],
		],
		[
			"lines",
			"Fill the big jug\nPour it into the small one",
			"en",
			["Fill the big jug", "Pour it into the small one"],
		],
		[
			"Hindi, whose full stop is the danda",
			"अब 5 बचे। फिर 3 जोड़ो। उत्तर 8 है।",
			"hi",
			["अब 5 बचे।", "फिर 3 जोड़ो।", "उत्तर 8 है।"],
		],
		[
			"Arabic, in its own digits",
			"٣ + ٤ = ٧. ٧ × ٢ = ١٤.",
			"ar",
			["٣ + ٤ = ٧.", "٧ × ٢ = ١٤."],
		],
		[
			"Japanese, with no spaces",
			"答えは5です。まず3を足す。",
			"ja",
			["答えは5です。", "まず3を足す。"],
		],
		[
			"one sentence",
			"5 × 2 = 10, and 10 × 17 = 170.",
			"en",
			["5 × 2 = 10, and 10 × 17 = 170."],
		],
		["nothing", " \n ", "en", []],
	])("of %s", (_, solution, language, want) => {
		expect(stepsOf(solution, language)).toEqual(want);
	});

	test.each(["en_US", "", "not a language"])(
		"in a language written %j are cut by the platform's rules",
		(language) => {
			expect(stepsOf(fenceSolution, language)).toHaveLength(3);
		},
	);

	test("are whole lines where the platform has no sentence rules", () => {
		vi.stubGlobal("Intl", { ...Intl, Segmenter: undefined });

		expect(stepsOf("One. Two.\nThree.", "en")).toEqual(["One. Two.", "Three."]);
	});
});

// holdsTogether says what a solution's steps must be whatever it says: no step
// empty, none with space around it or a line break in it, and all of them
// together the solution itself, nothing lost and nothing added.
function holdsTogether(solution: string, steps: readonly string[]): void {
	for (const step of steps) {
		expect(step).not.toBe("");
		expect(step).toBe(step.trim());
		expect(step).not.toMatch(/[\r\n]/);
	}
	expect(steps.join("").replace(/\s/g, "")).toBe(solution.replace(/\s/g, ""));
}

describe("every reference solution", () => {
	// The solutions the reference tasks carry, which a model writing a task
	// learns its way of telling one from.
	const solutions = Object.values(
		import.meta.glob<readonly { solution: string }[]>(
			"../../../content/examples/*.json",
			{ eager: true, import: "default" },
		),
	).flatMap((tasks) => tasks.map((task) => task.solution));

	test("is there to be read", () => {
		expect(solutions.length).toBeGreaterThan(500);
	});

	test("is told in steps that hold together, none of them a piece of a sum", () => {
		for (const solution of solutions) {
			const steps = stepsOf(solution, "en");
			holdsTogether(solution, steps);
			expect(steps.length).toBeGreaterThan(0);
			for (const step of steps) {
				expect(step).not.toMatch(/^(?:[+×÷=*/:]|[-−]\s)/u);
			}
		}
	});
});

// seeded is a source of numbers from 0 to 1 that says the same every run: a
// failure found once is found again.
function seeded(seed: number): () => number {
	let state = seed;
	return () => {
		state = (state + 0x6d2b79f5) | 0;
		let mixed = Math.imul(state ^ (state >>> 15), 1 | state);
		mixed = (mixed + Math.imul(mixed ^ (mixed >>> 7), 61 | mixed)) ^ mixed;
		return ((mixed ^ (mixed >>> 14)) >>> 0) / 4294967296;
	};
}

// The pieces a solution a model writes is made of: words in several scripts,
// numbers in several digits, operations, and every kind of stop and space.
const pieces = [
	"posts",
	"Pour",
	"так",
	"बचे",
	"答え",
	"كيلو",
	"12",
	"٧",
	"2.5",
	"+",
	"−",
	"×",
	"÷",
	"=",
	"...",
	"…",
	".",
	"!",
	"?",
	"。",
	"।",
	"؟",
	" ",
	", ",
	"  ",
	"\n",
	"\r\n",
	"\t",
	"\u00a0",
	"(",
	")",
	":",
];
const languages = ["en", "ru", "hi", "ar", "ja", "zh-Hans", "en_US", "", "x-?"];

test("steps hold together whatever the solution says, in whatever language", () => {
	const random = seeded(55);
	const pick = <T>(from: readonly T[]): T =>
		from[Math.floor(random() * from.length)] as T;
	for (let run = 0; run < 1000; run++) {
		const solution = Array.from({ length: Math.floor(random() * 40) }, () =>
			pick(pieces),
		).join("");
		holdsTogether(solution, stepsOf(solution, pick(languages)));
	}
});

import { describe, expect, test } from "vitest";
import fixture from "../../../site/research/testdata/research.json";
import { letters } from "../widget/choices";
import { siteCatalog } from "./data";
import { markFrom, readResearch } from "./research";

// authors are the authors the site's data names for the page.
const authors = {
	sources: [{ id: "perelman", died: 1942, books: ["firsthundred"] }],
};

// changed is the fixture with one thing changed, as edit changes it.
function changed(edit: (file: typeof fixture) => void): unknown {
	const file = structuredClone(fixture);
	edit(file);
	return file;
}

// read reads data as the build reads the page's file.
function read(data: unknown) {
	return readResearch(siteCatalog, authors, data);
}

// rowOf is the row of the fixture with this id, to change.
function rowOf(file: typeof fixture, id: string) {
	const row = file.bench.rows.find((each) => each.id === id);
	if (row === undefined) {
		throw new Error(`the fixture has no row ${id}`);
	}
	return row;
}

describe("the data of the page Research", () => {
	test("is the file the bench writes, read whole, with the site's authors", () => {
		const research = read(fixture);

		expect(research.bench.rows.map(({ id }) => id)).toEqual(
			fixture.bench.rows.map(({ id }) => id),
		);
		expect(research.product).toEqual(fixture.product);
		expect(research.authors).toEqual(authors.sources);
	});

	test("reads the levels of the reference tasks from the lowest, with the grades each spans", () => {
		expect(read(fixture).levels).toEqual([
			{ key: "1-2", first: 1, last: 2, count: 200 },
			{ key: "3-4", first: 3, last: 4, count: 250 },
			{ key: "5-6", first: 5, last: 6, count: 153 },
		]);
	});

	test("reads the levels from the lowest in whatever order the file gives them", () => {
		const edited = changed((file) => {
			file.product.reference_tasks.by_level = {
				"5-6": 153,
				"3-4": 250,
				"1-2": 200,
			};
		});

		expect(read(edited).levels.map(({ key }) => key)).toEqual([
			"1-2",
			"3-4",
			"5-6",
		]);
	});

	test("is refused when the site's data names no authors for it", () => {
		expect(() => readResearch(siteCatalog, undefined, fixture)).toThrow(
			"site/data.json names no authors for the page Research",
		);
	});

	test.each([
		[
			"a key the bench does not write",
			(file: typeof fixture) => Object.assign(file, { more: 1 }),
		],
		[
			"a key the bench writes left out",
			(file: typeof fixture) => Reflect.deleteProperty(file, "live"),
		],
		[
			"another version of the shape",
			(file: typeof fixture) => Object.assign(file, { schema: 2 }),
		],
		[
			"a commit of another shape",
			(file: typeof fixture) =>
				Object.assign(file.built_from, { commit: "0123456" }),
		],
		[
			"the rule before called the baseline",
			(file: typeof fixture) =>
				Object.assign(rowOf(file, "jump_unsettled").values.earlier, {
					mark: "baseline",
				}),
		],
		[
			"a PDF served elsewhere than the assets",
			(file: typeof fixture) =>
				Object.assign(file.paper.files[0] ?? {}, {
					path: "/research/paper-a.en.pdf",
				}),
		],
	])("is refused for %s", (_, edit) => {
		expect(() => read(changed(edit))).toThrow("research.json:");
	});
});

describe("the marks of the page's goals", () => {
	test.each([
		["lower", 0.3, { low: 0.1, high: 0.3 }, "reached"],
		["lower", 0.3, { low: 0.2, high: 0.4 }, "on_the_edge"],
		["lower", 0.3, { low: 0.3, high: 0.4 }, "on_the_edge"],
		["lower", 0.3, { low: 0.31, high: 0.4 }, "not_reached"],
		["higher", 0.3, { low: 0.3, high: 0.4 }, "reached"],
		["higher", 0.3, { low: 0.2, high: 0.4 }, "on_the_edge"],
		["higher", 0.3, { low: 0.2, high: 0.3 }, "on_the_edge"],
		["higher", 0.3, { low: 0.1, high: 0.29 }, "not_reached"],
	] as const)(
		"are read %s than %s on [%o] as %s, the bound itself on the better side",
		(better, bound, ends, mark) => {
			expect(markFrom(better, bound, ends)).toBe(mark);
		},
	);

	test("are refused when the rule before is marked otherwise than its numbers give", () => {
		const edited = changed((file) => {
			Object.assign(rowOf(file, "jump_unsettled").values.earlier, {
				mark: "reached",
			});
		});

		expect(() => read(edited)).toThrow(
			"research.json: the rule before is marked reached on jump_unsettled, and its numbers give not_reached",
		);
	});

	test("are refused when the service is marked otherwise than its numbers give", () => {
		const edited = changed((file) => {
			Object.assign(rowOf(file, "false_mastery_static").values.service, {
				mark: "not_reached",
			});
		});

		expect(() => read(edited)).toThrow(
			"research.json: the service is marked not_reached on false_mastery_static, and its numbers give reached",
		);
	});

	test("are refused when the service is the baseline against a bound not its own", () => {
		const edited = changed((file) => {
			Object.assign(rowOf(file, "false_mastery_static").values.service, {
				mark: "baseline",
			});
		});

		expect(() => read(edited)).toThrow(
			"research.json: the service is marked baseline on false_mastery_static, and its numbers give reached",
		);
	});

	test("are refused when the service is marked against a bound read off its own number", () => {
		const edited = changed((file) => {
			Object.assign(rowOf(file, "lag").values.service, { mark: "not_reached" });
		});

		expect(() => read(edited)).toThrow(
			"research.json: the service is marked not_reached on lag, and its numbers give baseline",
		);
	});
});

describe("the rows of the page's table", () => {
	test("are refused when one is there twice", () => {
		const edited = changed((file) => {
			Object.assign(rowOf(file, "jump_unsettled"), { id: "lag" });
		});

		expect(() => read(edited)).toThrow(
			"research.json: the row lag is there twice",
		);
	});

	test("are refused when an interval's ends are the wrong way round", () => {
		const edited = changed((file) => {
			Object.assign(rowOf(file, "error_static").values.service, {
				low: 0.6,
				high: 0.4,
			});
		});

		expect(() => read(edited)).toThrow(
			"research.json: the row error_static has an interval from 0.6 down to 0.4",
		);
	});

	test.each([
		["its number", { value: -0.1 }],
		["the low end of its interval", { low: -0.1 }],
	])(
		"are refused when %s runs under zero, where a row's drawing starts",
		(_, under) => {
			const edited = changed((file) => {
				Object.assign(rowOf(file, "error_static").values.service, under);
			});

			expect(() => read(edited)).toThrow(
				"research.json: the row error_static has a number under zero, where its drawing starts",
			);
		},
	);

	test("are refused when a goal lies under zero, where a row's drawing starts", () => {
		const edited = changed((file) => {
			Object.assign(rowOf(file, "lag").bound ?? {}, { value: -0.1 });
		});

		expect(() => read(edited)).toThrow(
			"research.json: the row lag has a goal under zero, where its drawing starts",
		);
	});

	test("are refused when the rules do not play each part once", () => {
		const edited = changed((file) => {
			Object.assign(file.bench.rules[1] ?? {}, { role: "service" });
		});

		expect(() => read(edited)).toThrow(
			"research.json: the rules play service, service, ceiling",
		);
	});
});

describe("the product's counts on the page", () => {
	const { topics, traps, reference_tasks, grades } = fixture.product;
	const levels = Object.entries(reference_tasks.by_level);
	const [level = "", tasks = 0] = levels[0] ?? [];
	const [lastLevel = "", lastTasks = 0] = levels.at(-1) ?? [];
	// byLevel is a file's count of the reference tasks by level, to change.
	const byLevel = (file: typeof fixture) =>
		file.product.reference_tasks.by_level as Record<string, number>;

	test.each([
		[
			"the topics",
			(file: typeof fixture) => (file.product.topics += 1),
			`it counts ${topics + 1} topics, and the catalog ${topics}`,
		],
		[
			"the traps",
			(file: typeof fixture) => (file.product.traps -= 1),
			`it counts ${traps - 1} traps, and the catalog ${traps}`,
		],
		[
			"the reference tasks",
			(file: typeof fixture) => (file.product.reference_tasks.total += 1),
			`it counts ${reference_tasks.total + 1} reference tasks, and the catalog ${reference_tasks.total}`,
		],
		[
			"the reference tasks at a level",
			(file: typeof fixture) => (byLevel(file)[level] = tasks + 1),
			`it counts ${tasks + 1} reference tasks at ${level}, and the catalog ${tasks}`,
		],
		[
			"the levels",
			(file: typeof fixture) =>
				Reflect.deleteProperty(byLevel(file), lastLevel),
			`it counts ${levels.length - 1} levels, and the catalog ${levels.length}`,
		],
		[
			"a level the catalog does not have",
			(file: typeof fixture) => {
				Reflect.deleteProperty(byLevel(file), lastLevel);
				byLevel(file)["7-8"] = lastTasks;
			},
			`it counts ${lastTasks} reference tasks at 7-8, and the catalog undefined`,
		],
		[
			"the first grade",
			(file: typeof fixture) => (file.product.grades.first += 1),
			`it counts ${grades.first + 1} as the first grade, and the catalog ${grades.first}`,
		],
		[
			"the last grade",
			(file: typeof fixture) => (file.product.grades.last += 1),
			`it counts ${grades.last + 1} as the last grade, and the catalog ${grades.last}`,
		],
	])(
		"are refused when they disagree with the catalog on %s",
		(_, edit, said) => {
			expect(() => read(changed(edit))).toThrow(`research.json: ${said}`);
		},
	);
});

describe("the product's constants on the page", () => {
	const { options, relabel, relabelled, guess, corridor } = fixture.product;
	const swapped = [
		...relabelled.slice(0, -2),
		...relabelled.slice(-2).reverse(),
	];

	test.each([
		[
			"options other than the card's",
			(file: typeof fixture) => (file.product.options = options - 1),
			`it offers ${options - 1} options, and a card shows ${options}`,
		],
		[
			"a relabelling that is no shift by its size",
			(file: typeof fixture) => (file.product.relabelled = swapped),
			`it relabels the letters as ${swapped.join("")}, which is no shift by ${relabel}: that is ${relabelled.join("")}`,
		],
		[
			"a shift that leaves every letter where it was",
			(file: typeof fixture) => {
				file.product.relabel = options;
				file.product.relabelled = [...letters];
			},
			`it moves the letters ${options} places on, and among ${options} options a shift that moves every letter is 1 to ${options - 1}`,
		],
		[
			"a shift by more places than there are options",
			(file: typeof fixture) => (file.product.relabel = relabel + options),
			`it moves the letters ${relabel + options} places on, and among ${options} options a shift that moves every letter is 1 to ${options - 1}`,
		],
		[
			"a guess that is not one option in all of them",
			(file: typeof fixture) => (file.product.guess = guess * 2),
			`its guess is ${guess * 2}, and one option in ${options} is ${1 / options}`,
		],
		[
			"a corridor that starts at no chance at all",
			(file: typeof fixture) => (file.product.corridor.low = 0),
			`its corridor runs from 0 to ${corridor.high} around ${corridor.middle}`,
		],
		[
			"a corridor whose middle lies under its start",
			(file: typeof fixture) =>
				(file.product.corridor.middle = corridor.low / 2),
			`its corridor runs from ${corridor.low} to ${corridor.high} around ${corridor.low / 2}`,
		],
		[
			"a corridor whose middle lies over its end",
			(file: typeof fixture) =>
				(file.product.corridor.middle = (corridor.high + 1) / 2),
			`its corridor runs from ${corridor.low} to ${corridor.high} around ${(corridor.high + 1) / 2}`,
		],
		[
			"a corridor that ends at a certain answer",
			(file: typeof fixture) => (file.product.corridor.high = 1),
			`its corridor runs from ${corridor.low} to 1 around ${corridor.middle}`,
		],
	])("are refused for %s", (_, edit, said) => {
		expect(() => read(changed(edit))).toThrow(`research.json: ${said}`);
	});
});

describe("the paper's files on the page", () => {
	test("are refused when two are served at one address", () => {
		const edited = changed((file) => {
			file.paper.files.push({
				...(file.paper.files[0] ?? fixture.paper.files[0]),
				lang: "ru",
			} as (typeof file.paper.files)[number]);
		});

		expect(() => read(edited)).toThrow(
			"research.json: two files of the paper are served at one address",
		);
	});

	test("are refused when none is in English, which every language links", () => {
		const edited = changed((file) => {
			Object.assign(file.paper.files[0] ?? {}, { lang: "ru" });
		});

		expect(() => read(edited)).toThrow(
			"research.json: the paper's files hold none in English",
		);
	});

	test("may be none, until a whole paper is kept", () => {
		const edited = changed((file) => {
			file.paper.files.length = 0;
		});

		expect(read(edited).paper.files).toEqual([]);
	});
});

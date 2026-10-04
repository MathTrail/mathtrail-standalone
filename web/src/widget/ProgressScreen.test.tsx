import { render } from "preact";
import { act } from "preact/test-utils";
import { afterEach, describe, expect, test, vi } from "vitest";
import { ProgressScreen } from "./ProgressScreen";
import { readScreen } from "./payload";
import { type Drawn, drawCard, foldIn, press, takeDown } from "./testing/card";
import {
	atTheTop,
	inTrial,
	moving,
	standing,
	standingBefore,
} from "./testing/lesson";

let drawn: Drawn | undefined;

afterEach(() => {
	takeDown(drawn?.root);
	drawn = undefined;
	vi.restoreAllMocks();
});

// draw draws the progress card of payload, as it first opens.
async function draw(
	payload: object,
	options: Parameters<typeof drawCard>[1] = {},
): Promise<Drawn> {
	drawn = await drawCard(payload, options);
	return drawn;
}

// drawOpened draws the progress card of payload and opens each of its
// sections in turn, as a person reading all of it does.
async function drawOpened(payload: object): Promise<Drawn> {
	const opened = await draw(payload);
	for (const title of opened.root.querySelectorAll<HTMLButtonElement>(
		'.mt-fold-button[aria-expanded="false"]',
	)) {
		press(title);
	}
	return opened;
}

// sections are the card's sections in order, each its title, the summary
// beside it, whether it is open, and whether what it holds is shown.
function sections(root: HTMLElement): [string, string, string, boolean][] {
	return [...root.querySelectorAll(".mt-fold")].map((fold) => [
		fold.querySelector(".mt-fold-title")?.textContent ?? "",
		fold.querySelector(".mt-fold-summary")?.textContent ?? "",
		fold.querySelector(".mt-fold-button")?.getAttribute("aria-expanded") ?? "",
		!(fold.querySelector(".mt-fold-body")?.hasAttribute("hidden") ?? true),
	]);
}

const text = (root: HTMLElement, selector: string) =>
	root.querySelector(selector)?.textContent;

// fills are how far each step of the first course drawn in element is filled,
// or "open" for a step drawn open.
function fills(element: Element | null | undefined): string[] {
	return [...(element?.querySelector(".mt-segments")?.children ?? [])].flatMap(
		(step) => {
			if (step.classList.contains("mt-segment-open")) {
				return ["open"];
			}
			const fill = step.querySelector(".mt-segment-fill");
			return fill === null ? [] : [fill.getAttribute("width") ?? ""];
		},
	);
}

// eleven are eleven steps, as many as the ranks, filled as given and the rest
// empty.
function eleven(...filled: string[]): string[] {
	return [...filled, ...Array.from({ length: 11 - filled.length }, () => "0%")];
}

// topics are the card's topics, each its name with its mark, how it stands
// with its rank's name, the step of the ramp its course is drawn in, and how
// far each step of the course is filled.
function topics(root: HTMLElement): [string, string, string, string[]][] {
	return [...root.querySelectorAll(".mt-rank-row")].map((row) => [
		row.querySelector(".mt-rank-row-start")?.textContent ?? "",
		row.querySelector(".mt-rank-row-end")?.textContent ?? "",
		row.querySelector(".mt-segments")?.getAttribute("data-tone") ?? "",
		fills(row),
	]);
}

// lists are the card's lists of lines by the titles of their sections, each a
// list of its lines.
function lists(root: HTMLElement): Record<string, (string | null)[]> {
	return Object.fromEntries(
		[...root.querySelectorAll(".mt-fold")]
			.filter(
				(fold) =>
					fold.querySelector(".mt-list") !== null &&
					fold.querySelector(".mt-rank-row") === null,
			)
			.map((fold) => [
				fold.querySelector(".mt-fold-title")?.textContent ?? "",
				[...fold.querySelectorAll(".mt-row")].map((row) => row.textContent),
			]),
	);
}

// fields are the fields of the group labelled label.
function fields(root: HTMLElement, label: string): string[][] {
	return fieldsIn(
		[...root.querySelectorAll(".mt-fields")].find(
			(found) =>
				found.querySelector(".mt-fields-head .mt-section-label")
					?.textContent === label,
		),
	);
}

// profileFields are the fields of the profile, in the frame of the section
// titled title, which carries no label of its own.
function profileFields(root: HTMLElement, title = "Profile"): string[][] {
	return fieldsIn(
		foldIn(root, title).closest(".mt-fold")?.querySelector(".mt-fields"),
	);
}

// fieldsIn are the fields of group: each its name, what it holds — a list's
// names joined — and the note under it, when there is one.
function fieldsIn(group: Element | null | undefined): string[][] {
	return [...(group?.querySelectorAll("dl > div") ?? [])].map((field) => {
		const told = field.querySelector("dd");
		const names = [...(told?.querySelectorAll(".mt-chip") ?? [])].map(
			(chip) => chip.textContent ?? "",
		);
		const note = told?.querySelector(".mt-field-note")?.textContent;
		return [
			field.querySelector("dt")?.textContent ?? "",
			names.length > 0
				? names.join(", ")
				: (told?.firstChild?.textContent ?? ""),
			...(note === undefined || note === null ? [] : [note]),
		];
	});
}

describe("the progress", () => {
	test("names the rank reached over the rank out of how many, with no rating's number, its course filled as far as it has come, and the next rank", async () => {
		const { root } = await draw(standing);

		const rank = root.querySelector("section.mt-rank");
		expect(rank?.getAttribute("aria-label")).toBe("Overall rating");
		expect(text(root, ".mt-rank-name")).toBe("River crossing");
		expect(text(root, ".mt-rank-head .mt-meta")).toBe("rank 3 of 11");
		expect(fills(rank)).toEqual(eleven("100%", "100%", "43%"));
		expect(rank?.querySelector(".mt-segments")?.getAttribute("data-tone")).toBe(
			"ink",
		);
		expect(text(root, ".mt-rank .mt-vh")).toBe(
			"43% of the way to the next rank",
		);
		expect(text(root, ".mt-rank-line")).toBe(
			"Next rank — Hill. Right answers move the bar forward, mistakes a little back.",
		);
	});

	test("at the highest rank draws the whole way behind it, and no rank to come", async () => {
		const { root } = await draw(atTheTop);

		expect(text(root, ".mt-rank-name")).toBe("Above the clouds");
		expect(fills(root.querySelector(".mt-rank"))).toEqual(
			Array.from({ length: 11 }, () => "100%"),
		);
		expect(text(root, ".mt-rank-line")).toBe("This is the highest rank.");
		expect(root.querySelector(".mt-rank .mt-vh")).toBeNull();
	});

	test("says whose it is, with no way back to a task it never had", async () => {
		const { root } = await draw(standing);

		expect(text(root, ".mt-head .mt-name")).toBe("Comet");
		expect(text(root, ".mt-badge")).toBe("Grade 3");
		expect(root.querySelector(".mt-bar")).toBeNull();
	});

	test("shows no rank when it is told neither a rating nor a trial series", async () => {
		const { root } = await draw({ ...standing, overall: null });

		expect(root.querySelector(".mt-rank")).toBeNull();
		expect(root.querySelector(".mt-list")).not.toBeNull();
	});

	test("draws nothing over its sections when there is neither a rank nor what comes next", async () => {
		const { root } = await draw({
			...standing,
			overall: null,
			recommendation: null,
		});

		expect(root.querySelector(".mt-progress")).toBeNull();
		expect(root.querySelector(".mt-folds")).not.toBeNull();
	});

	test("opens with every section folded under its title, each title a button that says so", async () => {
		const { root } = await draw(standing);

		expect(sections(root)).toEqual([
			["Topics", "", "false", false],
			["Mistakes that repeat", "", "false", false],
			["Recent answers", "Wrong, Skipped, Right, Right, Wrong", "false", false],
			["Profile", "for the parent", "false", false],
		]);
		for (const fold of root.querySelectorAll(".mt-fold")) {
			const title = fold.querySelector(".mt-fold-button");
			expect(title?.tagName).toBe("BUTTON");
			expect(title?.getAttribute("type")).toBe("button");
			expect(title?.getAttribute("aria-controls")).toBe(
				fold.querySelector(".mt-fold-body")?.id,
			);
		}
	});

	test("opens a section when its title is pressed, and folds it again, keeping the focus on the title", async () => {
		const { root } = await draw(standing);
		const topicsTitle = foldIn(root, "Topics");

		press(topicsTitle);

		expect(
			sections(root).map(([title, , open, shown]) => [title, open, shown]),
		).toEqual([
			["Topics", "true", true],
			["Mistakes that repeat", "false", false],
			["Recent answers", "false", false],
			["Profile", "false", false],
		]);
		expect(document.activeElement).toBe(topicsTitle);

		press(topicsTitle);

		expect(sections(root)[0]).toEqual(["Topics", "", "false", false]);
		expect(document.activeElement).toBe(topicsTitle);
	});

	test("sums up the latest answers by their section's title: a dot for each in the colour of how it went, and the same in words for a screen reader", async () => {
		const { root } = await draw(standing);
		const summary = foldIn(root, "Recent answers").querySelector(
			".mt-fold-summary",
		);

		expect(
			[...(summary?.querySelectorAll(".mt-dot") ?? [])].map((dot) =>
				dot.getAttribute("data-tone"),
			),
		).toEqual(["wrong", "skipped", "correct", "correct", "wrong"]);
		expect(
			summary?.querySelector(".mt-dots")?.getAttribute("aria-hidden"),
		).toBe("true");
		expect(summary?.querySelector(".mt-vh")?.textContent).toBe(
			"Wrong, Skipped, Right, Right, Wrong",
		);
	});

	test("names the profile once, by its section's title: its frame carries no label of its own, and Edit at its head", async () => {
		const { root } = await drawOpened(standing);
		const frame = foldIn(root, "Profile")
			.closest(".mt-fold")
			?.querySelector(".mt-fields");

		expect(frame?.querySelector(".mt-section-label")).toBeNull();
		expect(frame?.querySelector(".mt-fields-head .mt-btn")?.textContent).toBe(
			"Edit",
		);
		expect(fields(root, "Your data")).toHaveLength(4);
	});

	test("draws no section with nothing in it", async () => {
		const { root } = await draw({
			...standing,
			topics: [],
			recent: [],
			mistakes: [],
		});

		expect(sections(root).map(([title]) => title)).toEqual(["Profile"]);
	});

	test("names what comes next, and that it comes again after a mistake", async () => {
		const { root } = await draw(standing);

		expect(text(root, ".mt-note-plain .mt-note-label")).toBe("Next up");
		expect(text(root, ".mt-note-plain p")).toBe("Enumeration, once more");
	});

	test("shows each topic's rank, how it stands to the overall one, and its course in the colour of its rank", async () => {
		const { root } = await drawOpened(standing);

		expect(text(root, ".mt-list-lead")).toBe("The rank in each topic");
		expect(topics(root)).toEqual([
			[
				"OrderingMastered",
				"aheadHill",
				"2",
				eleven("100%", "100%", "100%", "27%"),
			],
			["Enumeration", "evenRiver crossing", "1", eleven("100%", "100%", "76%")],
			[
				"Gaps and boundaries",
				"evenRiver crossing",
				"1",
				eleven("100%", "100%", "53%"),
			],
			[
				"Parity and alternation",
				"behindForest path",
				"1",
				eleven("100%", "87%"),
			],
			[
				"Pigeonhole principle",
				"no answers yet",
				"ink",
				Array.from({ length: 11 }, () => "open"),
			],
		]);
	});

	test("lists the latest five entries, and how many tasks were left without an answer in all, as the service counts them", async () => {
		const { root } = await drawOpened({ ...standing, skipped: 4 });

		expect(lists(root)["Recent answers"]).toEqual([
			"EnumerationWrong",
			"EnumerationSkipped",
			"Gaps and boundariesRight",
			"OrderingRight",
			"Parity and alternationWrong",
		]);
		expect(text(root, ".mt-list-note")).toBe(
			"In all, 4 tasks were left without an answer.",
		);
	});

	test("lists the mistakes that keep coming back before the latest answers, a dot for each time, in a frame of their own", async () => {
		const { root } = await drawOpened(standing);

		expect(Object.keys(lists(root))).toEqual([
			"Mistakes that repeat",
			"Recent answers",
		]);
		expect(lists(root)["Mistakes that repeat"]).toEqual([
			"Missed a case while listing3 times",
			"Counted the same thing twice2 times",
		]);
		const mistakes = root.querySelector("section.mt-list-framed");
		expect(
			[...(mistakes?.querySelectorAll(".mt-dots") ?? [])].map(
				(dots) => dots.children.length,
			),
		).toEqual([3, 2]);
	});

	test("shows no section of mistakes when none has come up twice", async () => {
		const { root } = await drawOpened({ ...standing, mistakes: [] });

		expect(Object.keys(lists(root))).toEqual(["Recent answers"]);
		expect(root.querySelector(".mt-list-framed")).toBeNull();
	});

	test("with no answer yet shows the series to come, and no list", async () => {
		const { root } = await drawOpened({
			...inTrial,
			trial: { answered: 0, of: 5 },
			topics: [],
			recent: [],
			mistakes: [],
		});

		expect(text(root, ".mt-rank-head .mt-meta")).toBe("0 of 5");
		expect(fills(root.querySelector(".mt-rank"))).toEqual([
			"0%",
			"0%",
			"0%",
			"0%",
			"0%",
		]);
		expect(root.querySelector(".mt-list")).toBeNull();
		expect(profileFields(root)).toHaveLength(4);
	});

	test("says nothing of skipped tasks when the service counts none", async () => {
		const { root } = await draw({ ...standing, skipped: 0 });

		expect(root.querySelector(".mt-list-note")).toBeNull();
	});

	test("from an earlier release still draws: its course with the step under way empty, its topics with no rank, the skips its topics count, and no data", async () => {
		const { root } = await drawOpened(standingBefore);

		expect(text(root, ".mt-rank-name")).toBe("River crossing");
		expect(fills(root.querySelector(".mt-rank"))).toEqual(
			eleven("100%", "100%"),
		);
		expect(root.querySelector(".mt-rank .mt-vh")).toBeNull();
		expect(
			topics(root).map(([name, standsAt, tone]) => [name, standsAt, tone]),
		).toEqual([
			["OrderingMastered", "", "ink"],
			["Enumeration", "", "ink"],
			["Gaps and boundaries", "", "ink"],
			["Parity and alternation", "", "ink"],
		]);
		expect(text(root, ".mt-list-note")).toBe(
			"In all, 1 task was left without an answer.",
		);
		expect(fields(root, "Your data")).toEqual([]);
	});

	test("shows the child's profile for the parent, each list as its names", async () => {
		const { root } = await drawOpened(standing);

		expect(profileFields(root)).toEqual([
			["Grade", "3", "Only a label: changing it moves no rating."],
			["Interests", "space, animals, football"],
			["Not at school yet", "Division with a remainder"],
			["Language of the lessons", "The chat's language"],
		]);
		expect(root.querySelectorAll(".mt-chip")).toHaveLength(4);
	});

	test("says a list the parent left empty is not set", async () => {
		const { root } = await drawOpened({
			...standing,
			profile: { ...standing.profile, interests: [], excluded_skills: [] },
		});

		expect(profileFields(root).slice(1, 3)).toEqual([
			["Interests", "Not set"],
			["Not at school yet", "Not set"],
		]);
	});

	test("says where the profile's file is and what the parent can do with the data, each under the question it answers", async () => {
		const { root } = await drawOpened(standing);

		expect(fields(root, "Your data")).toEqual([
			[
				"Where the profile is",
				"The whole profile is one file in your Google Drive: mathtrail-profile.json, in the folder MathTrail. Download or copy it there; the chat can give you its link.",
			],
			[
				"How to delete it",
				"Delete the file and empty Drive's bin: nothing of the profile stays anywhere.",
			],
			[
				"How to cut off access",
				"In your Google account, under third-party access, remove MathTrail: it can then reach nothing.",
			],
			[
				"How to remove the app",
				"Disconnect MathTrail in your chat's settings.",
			],
		]);
		// It names the file, and opens nothing.
		expect(root.querySelector("a")).toBeNull();
	});

	test("with nowhere to say the file is, shows no data", async () => {
		const { location: _, ...nowhere } = standing;
		const { root } = await draw(nowhere);

		expect(fields(root, "Your data")).toEqual([]);
		expect(profileFields(root)).toHaveLength(4);
	});

	test("in the trial series shows how far it has got, with no rank anywhere and every topic drawn open", async () => {
		const { root } = await drawOpened(inTrial);

		// The series is named once, by its name.
		expect(
			root.querySelector("section.mt-rank")?.hasAttribute("aria-label"),
		).toBe(false);
		expect(text(root, ".mt-rank-name")).toBe("Trial series");
		expect(text(root, ".mt-rank-head .mt-meta")).toBe("3 of 5");
		expect(fills(root.querySelector(".mt-rank"))).toEqual([
			"100%",
			"100%",
			"100%",
			"0%",
			"0%",
		]);
		expect(text(root, ".mt-rank-line")).toBe("Ranks come after 5 tasks.");
		// No topic has a rank yet, so nothing says what their ranks are.
		expect(root.querySelector(".mt-list-lead")).toBeNull();
		const open = Array.from({ length: 11 }, () => "open");
		expect(topics(root)).toEqual([
			["Ordering", "no rank yet", "ink", open],
			["Gaps and boundaries", "no rank yet", "ink", open],
			["Clocks", "no rank yet", "ink", open],
			["Parity and alternation", "no rank yet", "ink", open],
		]);
		expect(text(root, ".mt-note-plain p")).toBe("Parity and alternation");
	});

	test("is in the language the parent chose for the lessons", async () => {
		const { root } = await drawOpened({
			...standing,
			profile: { ...standing.profile, ui_language: "ru" },
		});

		expect(text(root, ".mt-rank-name")).toBe("Брод");
		expect(text(root, ".mt-rank-head .mt-meta")).toBe("ранг 3 из 11");
		expect(sections(root).map(([title, summary]) => [title, summary])).toEqual([
			["Темы", ""],
			["Повторяющиеся ошибки", ""],
			["Последние ответы", "Неверно, Пропущено, Верно, Верно, Неверно"],
			["Профиль", "для родителя"],
		]);
		expect(text(root, ".mt-rank-line")).toBe(
			"Следующий ранг — Холм. Верные ответы двигают полоску вперёд, ошибки — чуть назад.",
		);
		expect(text(root, ".mt-note-plain p")).toBe("Перебор, ещё раз");
		expect(topics(root).map(([, standsAt]) => standsAt)).toEqual([
			"впередиХолм",
			"вровеньБрод",
			"вровеньБрод",
			"отстаётЛесная тропа",
			"ответов пока нет",
		]);
		expect(text(root, ".mt-list-note")).toBe(
			"Всего без ответа оставили 1 задачу.",
		);
		expect(lists(root)["Повторяющиеся ошибки"]).toEqual([
			"Пропущен случай при переборе3 раза",
			"Одно и то же посчитано дважды2 раза",
		]);
		expect(profileFields(root, "Профиль")[3]).toEqual([
			"Язык занятий",
			"Русский",
		]);
		expect(fields(root, "Ваши данные").map(([term]) => term)).toEqual([
			"Где профиль",
			"Как удалить",
			"Как закрыть доступ",
			"Как убрать приложение",
		]);
	});

	test("keeps markup in the parent's words as text", async () => {
		const trick = "<img src=x onerror=alert(1)>";
		const { root } = await draw({
			...standing,
			profile: { ...standing.profile, pseudonym: trick, interests: [trick] },
		});

		expect(root.querySelector("img")).toBeNull();
		expect(text(root, ".mt-head .mt-name")).toBe(trick);
	});

	test("calls no tool of its own", async () => {
		const { heard } = await draw(standing);

		expect(heard.calls).toEqual([]);
	});
});

// periods are the whiles the switch offers, each its value and whether it is
// the one chosen.
function periods(root: HTMLElement): [string, boolean][] {
	return [...root.querySelectorAll<HTMLInputElement>(".mt-switch input")].map(
		(radio) => [radio.value, radio.checked],
	);
}

// stripes are how far the stripes of a move reach on each step of the first
// course drawn in element, or "" on a step no move is drawn across.
function stripes(element: Element | null | undefined): string[] {
	return [
		...(element?.querySelector(".mt-segments")?.querySelectorAll("svg") ?? []),
	].map(
		(step) =>
			step.querySelector(".mt-segment-stripes")?.getAttribute("width") ?? "",
	);
}

// counted is the topics' count of moves by their title, each way that moved
// with its arrow, and the same in words for a screen reader.
function counted(root: HTMLElement): [string[], string] {
	const summary = foldIn(root, "Topics");
	return [
		[...summary.querySelectorAll(".mt-move-counted [data-way]")].map(
			(count) => count.textContent ?? "",
		),
		summary.querySelector(".mt-move-counts .mt-vh")?.textContent ?? "",
	];
}

// chosen presses the while named on the switch.
function chosen(root: HTMLElement, period: string): void {
	const radio = root.querySelector<HTMLInputElement>(
		`.mt-switch input[value="${period}"]`,
	);
	if (radio === null) {
		throw new Error(`no ${period} on the switch`);
	}
	press(radio);
}

// withMoves is moving with the moves of the overall rank and of every topic
// told over the week as week says of each.
function withWeek(week: (change: { week: unknown }) => unknown) {
	return {
		...moving,
		overall: {
			...moving.overall,
			change: { ...moving.overall.change, week: week(moving.overall.change) },
		},
		topics: moving.topics.map((topic) =>
			topic.change === undefined
				? topic
				: { ...topic, change: { ...topic.change, week: week(topic.change) } },
		),
	};
}

describe("the moves", () => {
	test("are drawn over the week at first: the rank's line and stripes, each topic's word and stripes, the topics' count and a legend", async () => {
		const { root } = await drawOpened(moving);

		expect(periods(root)).toEqual([
			["last_task", false],
			["week", true],
		]);
		expect(text(root, ".mt-rank-move")).toBe(
			"↑Over the past week — a new rank: Forest path → River crossing",
		);
		expect(root.querySelector(".mt-rank-move")?.getAttribute("data-way")).toBe(
			"gain",
		);
		const rank = root.querySelector(".mt-rank");
		expect(fills(rank)).toEqual(eleven("100%", "80%", "0%"));
		expect(stripes(rank)).toEqual([
			"",
			"100%",
			"43%",
			"",
			"",
			"",
			"",
			"",
			"",
			"",
			"",
		]);
		expect(topics(root).map(([name, end]) => [name, end])).toEqual([
			["OrderingMastered", "↑new rankHill"],
			["Enumeration", "↑forwardRiver crossing"],
			["Gaps and boundaries", "evenRiver crossing"],
			["Parity and alternation", "↓rank belowForest path"],
			["Pigeonhole principle", "no answers yet"],
		]);
		// Ordering reached a new rank over the week: striped from where it stood
		// in the rank before, across the step, to where it stands.
		expect(stripes(root.querySelector(".mt-rank-row"))).toEqual([
			"",
			"",
			"100%",
			"27%",
			...Array.from({ length: 7 }, () => ""),
		]);
		expect(counted(root)).toEqual([
			["↑ 2", "↓ 1"],
			"2 topics moved forward, 1 topic slipped back",
		]);
		expect(text(root, ".mt-move-legend")).toBe("gainloss· over the past week");
	});

	test("are drawn since the last task once it is chosen", async () => {
		const { root } = await drawOpened(moving);

		chosen(root, "last_task");

		expect(periods(root)).toEqual([
			["last_task", true],
			["week", false],
		]);
		expect(text(root, ".mt-rank-move")).toBe(
			"↓The last task moved the bar back a little",
		);
		// A step back of four points is drawn over a seventh of the step.
		expect(stripes(root.querySelector(".mt-rank"))[2]).toBe("57%");
		expect(topics(root).map(([, end]) => end)).toEqual([
			"aheadHill",
			"↓backRiver crossing",
			"evenRiver crossing",
			"behindForest path",
			"no answers yet",
		]);
		expect(counted(root)).toEqual([["↓ 1"], "1 topic slipped back"]);
		expect(text(root, ".mt-move-legend")).toBe("gainloss· since the last task");
	});

	test("are drawn since the last task at first when the week cannot be told, and say there is nothing to compare when it is chosen", async () => {
		const { root } = await drawOpened(withWeek(() => null));

		expect(periods(root)).toEqual([
			["last_task", true],
			["week", false],
		]);
		chosen(root, "week");

		expect(text(root, ".mt-rank-move")).toBe(
			"Nothing to compare the past week with yet",
		);
		expect(root.querySelector(".mt-rank-move")?.hasAttribute("data-way")).toBe(
			false,
		);
		expect(root.querySelector(".mt-segment-stripes")).toBeNull();
		expect(topics(root).map(([, end]) => end)).toEqual([
			"aheadHill",
			"evenRiver crossing",
			"evenRiver crossing",
			"behindForest path",
			"no answers yet",
		]);
		expect(foldIn(root, "Topics").querySelector(".mt-move-counts")).toBeNull();
	});

	test("say the bar stayed where nothing moved, with no stripes and no count", async () => {
		const { root } = await drawOpened(
			withWeek((change) => ({ ...(change.week as object), moved: "same" })),
		);

		expect(text(root, ".mt-rank-move")).toBe(
			"Over the past week the bar stayed where it was",
		);
		expect(root.querySelector(".mt-segment-stripes")).toBeNull();
		expect(foldIn(root, "Topics").querySelector(".mt-move-counts")).toBeNull();
	});

	test("count a topic new to the while among those that moved up, with a word of its own", async () => {
		const { root } = await drawOpened({
			...moving,
			topics: [
				{
					...moving.topics[2],
					change: { last_task: null, week: { moved: "new" } },
				},
			],
		});

		expect(topics(root).map(([, end]) => end)).toEqual([
			"↑new topicRiver crossing",
		]);
		expect(counted(root)).toEqual([["↑ 1"], "1 topic moved forward"]);
	});

	test("draw nothing of a move a later release names, and the progress all the same", async () => {
		const { root } = await drawOpened({
			...moving,
			overall: {
				...moving.overall,
				change: {
					...moving.overall.change,
					week: { rank: 3, share: 40, moved: "leap" },
				},
			},
		});

		expect(text(root, ".mt-rank-name")).toBe("River crossing");
		expect(text(root, ".mt-rank-move")).toBe("");
		expect(stripes(root.querySelector(".mt-rank")).join("")).toBe("");
	});

	test("are not drawn during the trial series, nor from a progress that tells none", async () => {
		for (const payload of [inTrial, standing]) {
			const { root } = await drawOpened(payload);

			expect(root.querySelector(".mt-switch")).toBeNull();
			expect(root.querySelector(".mt-rank-move")).toBeNull();
			expect(root.querySelector(".mt-move-legend")).toBeNull();
			expect(root.querySelector(".mt-segment-stripes")).toBeNull();
			takeDown(root);
		}
	});
});

describe("a progress screen nobody outlives", () => {
	// stillHost is a host that answers nothing: a screen drawn on its own,
	// over no card, has nobody to call.
	const stillHost = {
		callTool: () => Promise.reject(new Error("no host")),
		sendMessage: () => Promise.reject(new Error("no host")),
		tellModel: () => Promise.reject(new Error("no host")),
	};

	test("keeps the while chosen on its switch itself", () => {
		const root = document.createElement("div");
		document.body.append(root);
		const report = readScreen(moving);
		if (report?.screen !== "progress") {
			throw new Error(`the progress reads as ${report?.screen}`);
		}
		act(() =>
			render(
				<ProgressScreen
					report={report.report}
					wide={false}
					host={stillHost}
					folds={{ open: new Set(), toggle: () => {} }}
				/>,
				root,
			),
		);

		chosen(root, "last_task");

		expect(periods(root)).toEqual([
			["last_task", true],
			["week", false],
		]);
		expect(text(root, ".mt-rank-move")).toBe(
			"↓The last task moved the bar back a little",
		);
		act(() => render(null, root));
		root.remove();
	});
});

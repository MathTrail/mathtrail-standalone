import { render } from "preact";
import { act } from "preact/test-utils";
import { afterEach, describe, expect, test, vi } from "vitest";
import { ProgressScreen } from "./ProgressScreen";
import { readScreen } from "./payload";
import {
	type Drawn,
	drawCard,
	foldIn,
	press,
	takeDown,
	unfold,
} from "./testing/card";
import {
	atTheTop,
	failure,
	fence,
	inTrial,
	moving,
	progress,
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

// drawOpened draws the progress card of payload, on a host as options say,
// and opens each of its sections in turn, as a person reading all of it does.
async function drawOpened(
	payload: object,
	options: Parameters<typeof drawCard>[1] = {},
): Promise<Drawn> {
	const opened = await draw(payload, options);
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

// parts are the parts of the review, in order, each its heading and the tone
// it is headed in.
function parts(root: HTMLElement): [string, string][] {
	return [...root.querySelectorAll(".mt-review-part")].map((part) => {
		const label = part.querySelector(".mt-review-label");
		return [label?.textContent ?? "", label?.getAttribute("data-tone") ?? ""];
	});
}

// partOf is the part of the review headed heading.
function partOf(root: HTMLElement, heading: string): Element | undefined {
	return [...root.querySelectorAll(".mt-review-part")].find(
		(part) => part.querySelector(".mt-review-label")?.textContent === heading,
	);
}

// judged are the topics of the part of the review headed heading, each its
// name and why.
function judged(root: HTMLElement, heading: string): [string, string][] {
	return [
		...(partOf(root, heading)?.querySelectorAll(".mt-judged-row") ?? []),
	].map((row) => [
		row.querySelector(".mt-judged-name")?.textContent ?? "",
		row.querySelector(".mt-review-note")?.textContent ?? "",
	]);
}

// steps are the steps the review advises, each its number, the topic it is
// for — nothing for the step that holds for every topic — and what to do.
function steps(root: HTMLElement): [string, string, string][] {
	return [...root.querySelectorAll(".mt-advice-step")].map((step) => [
		step.querySelector(".mt-step-num")?.textContent ?? "",
		step.querySelector(".mt-advice-topic")?.textContent ?? "",
		step.querySelector(".mt-review-text > span:last-child")?.textContent ?? "",
	]);
}

// withReview is Comet's progress with a review of nothing but what review
// gives, and the rest of the progress as more changes it.
function withReview(review: object, more: object = {}) {
	return {
		...standing,
		review: { strong: [], develop: [], early: [], steps: [], ...review },
		...more,
	};
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

	test("opens with the topics open and every other section folded under its title, each title a button that says so", async () => {
		const { root } = await draw(standing);

		expect(sections(root)).toEqual([
			["Topics", "", "true", true],
			["Review", "strengths and plan", "false", false],
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
		const recentTitle = foldIn(root, "Recent answers");

		press(recentTitle);

		expect(
			sections(root).map(([title, , open, shown]) => [title, open, shown]),
		).toEqual([
			["Topics", "true", true],
			["Review", "false", false],
			["Recent answers", "true", true],
			["Profile", "false", false],
		]);
		expect(document.activeElement).toBe(recentTitle);

		press(recentTitle);

		expect(sections(root)[2]).toEqual([
			"Recent answers",
			"Wrong, Skipped, Right, Right, Wrong",
			"false",
			false,
		]);
		expect(document.activeElement).toBe(recentTitle);
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
			review: { strong: [], develop: [], early: [], steps: [] },
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

	test("holds the mistakes that keep coming back in its review once the trial series is over, a dot for each time, in a frame of their own", async () => {
		const { root } = await drawOpened(standing);

		expect(Object.keys(lists(root))).toEqual(["Review", "Recent answers"]);
		expect(lists(root).Review).toEqual([
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

	test("lists the mistakes that keep coming back in a section of their own while the trial series runs, before the latest answers", async () => {
		const { root } = await drawOpened({
			...inTrial,
			mistakes: [{ trap: "off_by_one", times: 2 }],
		});

		expect(Object.keys(lists(root))).toEqual([
			"Mistakes that repeat",
			"Recent answers",
		]);
		expect(lists(root)["Mistakes that repeat"]).toEqual([
			"Off by one when counting gaps2 times",
		]);
		expect(root.querySelector(".mt-review-part")).toBeNull();
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
		expect(profileFields(root)).toHaveLength(5);
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
			[
				"Country",
				"Not set",
				"Optional. Used only to count, without names, how many families each country has.",
			],
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
		const { root } = await drawOpened(standing, { links: "open" });

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
		// It names the file, and opens nothing: the file is in Drive, no page of
		// the site, though the host opens pages.
		const data = [...root.querySelectorAll(".mt-fields")].find(
			(group) =>
				group.querySelector(".mt-fields-head .mt-section-label")
					?.textContent === "Your data",
		);
		expect(data).toBeDefined();
		expect(data?.querySelector("a")).toBeNull();
	});

	test("with nowhere to say the file is, shows no data", async () => {
		const { location: _, ...nowhere } = standing;
		const { root } = await draw(nowhere);

		expect(fields(root, "Your data")).toEqual([]);
		expect(profileFields(root)).toHaveLength(5);
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
			["Разбор", "сильные стороны и план"],
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
		expect(lists(root).Разбор).toEqual([
			"Пропущен случай при переборе3 раза",
			"Одно и то же посчитано дважды2 раза",
		]);
		expect(judged(root, "Сильные стороны")).toEqual([
			[
				"Упорядочивание",
				"Тема освоена. Заметно выше общего уровня. За неделю\u00a0— заметный рост.",
			],
		]);
		expect(steps(root)).toEqual([
			[
				"1",
				"Перебор",
				"Выписывать случаи в одном порядке, начиная с меньшего, и отмечать каждый, чтобы ни один не потерялся.",
			],
			[
				"2",
				"Чётность и чередование",
				"По 2–3 короткие задачи в день, пока полоска не вернётся.",
			],
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

// withWeek is moving with the week of the overall rank and of every topic
// told as week says of each.
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

describe("the review", () => {
	test("is a section after the topics, with the strengths and the plan by its title, and its parts in the order the adult reads them, each headed in its tone", async () => {
		const { root } = await drawOpened(standing);

		expect(sections(root).map(([title, summary]) => [title, summary])).toEqual([
			["Topics", ""],
			["Review", "strengths and plan"],
			["Recent answers", "Wrong, Skipped, Right, Right, Wrong"],
			["Profile", "for the parent"],
		]);
		expect(parts(root)).toEqual([
			["Strengths", "correct"],
			["To develop", "wrong"],
			["Too early to judge", "muted"],
			["Mistakes that repeat", "wrong"],
			["What to do next", "accent"],
		]);
	});

	test("names the strong topics and the ones to develop, each after a dot in its side's colour hidden from a screen reader, with a sentence for each reason", async () => {
		const { root } = await drawOpened(standing);

		expect(judged(root, "Strengths")).toEqual([
			[
				"Ordering",
				"Mastered. Well above the overall level. A clear step up over the past week.",
			],
		]);
		expect(judged(root, "To develop")).toEqual([
			["Enumeration", "Missed a case while listing\u00a0— again and again."],
			[
				"Parity and alternation",
				"Well below the overall level. A clear step back over the past week.",
			],
		]);
		const dots = [...root.querySelectorAll(".mt-judged-dot")];
		expect(dots.map((dot) => dot.getAttribute("data-tone"))).toEqual([
			"correct",
			"wrong",
			"wrong",
		]);
		expect(dots.map((dot) => dot.getAttribute("aria-hidden"))).toEqual([
			"true",
			"true",
			"true",
		]);
	});

	test("says why a topic is named for wrong answers in a row and the hint, and that a move has begun when its last answer was right", async () => {
		const { root } = await drawOpened(
			withReview({
				develop: [
					{
						topic: "counting.gaps",
						reasons: ["failures", "hints"],
						moving: true,
					},
				],
			}),
		);

		expect(judged(root, "To develop")).toEqual([
			[
				"Gaps and boundaries",
				"Several wrong answers in a row. Often needs the hint. The last answer was right\u00a0— a move has begun.",
			],
		]);
	});

	test("lists the topics too early to judge on a line, as the card's language writes a list, with why there is no verdict", async () => {
		const { root } = await drawOpened(
			withReview({ early: ["counting.gaps", "time.clocks"] }),
		);
		const early = partOf(root, "Too early to judge");

		expect(early?.querySelector(".mt-judged-name")?.textContent).toBe(
			"Gaps and boundaries, Clocks",
		);
		expect(early?.querySelector(".mt-review-note")?.textContent).toBe(
			"Too few answers so far to say how these stand.",
		);
	});

	test("numbers the steps in order, each the topic it is for over what to do, and the step for every topic with what to do alone", async () => {
		const { root } = await drawOpened(
			withReview({
				steps: [
					{
						kind: "trap",
						topic: "combinatorics.enumeration",
						trap: "missed_case",
					},
					{ kind: "rhythm", topic: "parity.alternation" },
					{ kind: "practice", topic: "counting.gaps" },
					{ kind: "unaided", topic: "logic.ordering" },
					{ kind: "trap", trap: "double_count" },
				],
			}),
		);

		expect(steps(root)).toEqual([
			[
				"1",
				"Enumeration",
				"List the cases in a fixed order, smallest first, and tick each one off so that none is left out.",
			],
			[
				"2",
				"Parity and alternation",
				"Two or three short tasks a day, until the bar comes back.",
			],
			[
				"3",
				"Gaps and boundaries",
				"A few more tasks in this topic\u00a0— MathTrail sets them at the child's level.",
			],
			["4", "Ordering", "Try each task without the hint first."],
			[
				"5",
				"",
				"Write each option down once, in a fixed order, and look for repeats before counting them.",
			],
		]);
		expect(root.querySelector(".mt-advice")?.tagName).toBe("OL");
	});

	test("suggests a topic to begin last, the topic over the strong one it builds on, linked to the top of its page", async () => {
		const calendar = {
			topic: "time.calendar",
			rating: null,
			rank: null,
			share: null,
			compared: null,
			answers: 0,
			correct: 0,
			mastered: false,
			skipped: 0,
			slug: "calendar-and-age",
			site_page: true,
		};
		const { root } = await drawOpened(
			withReview(
				{
					strong: [{ topic: "counting.gaps", reasons: ["mastered"] }],
					steps: [
						{ kind: "trap", trap: "double_count" },
						{ kind: "begin", topic: "time.calendar", base: "counting.gaps" },
					],
				},
				{ topics: [...standing.topics, calendar] },
			),
			{ links: "open" },
		);

		expect(steps(root)).toEqual([
			[
				"1",
				"",
				"Write each option down once, in a fixed order, and look for repeats before counting them.",
			],
			[
				"2",
				"Calendar and age",
				"Gaps and boundaries\u00a0— a strength and a good base to start this new topic.",
			],
		]);
		expect(
			root
				.querySelector(".mt-advice-step:last-child a.mt-link")
				?.getAttribute("href"),
		).toBe("https://mathtrail.app/en/topics/calendar-and-age/");
	});

	test("leaves out what it has no words for — a reason, a kind of step, a mistake's name in a sentence and its advice, the base of a topic to begin — and keeps the name of a topic named for nothing it can say", async () => {
		const { root } = await drawOpened(
			withReview({
				strong: [{ topic: "logic.ordering", reasons: ["shining"] }],
				develop: [
					{
						topic: "counting.gaps",
						reasons: ["low", "toString", "trap"],
						trap: "counted_the_cat",
					},
				],
				steps: [
					{ kind: "trap", topic: "counting.gaps", trap: "counted_the_cat" },
					{ kind: "dance", topic: "counting.gaps" },
					{ kind: "unaided", topic: "counting.gaps" },
					{ kind: "begin", topic: "time.calendar", base: "counting.cats" },
					{ kind: "begin", topic: "time.calendar" },
					{ kind: "begin", base: "logic.ordering" },
				],
			}),
		);

		expect(judged(root, "Strengths")).toEqual([["Ordering", ""]]);
		expect(judged(root, "To develop")).toEqual([
			["Gaps and boundaries", "Well below the overall level."],
		]);
		expect(steps(root)).toEqual([
			["1", "Gaps and boundaries", "Try each task without the hint first."],
		]);
	});

	test("with nothing to say of the topics still holds the mistakes that repeat", async () => {
		const { root } = await drawOpened(withReview({}));

		expect(parts(root)).toEqual([["Mistakes that repeat", "wrong"]]);
	});

	test("is not drawn during the trial series, from a progress before the review, or when it says nothing at all", async () => {
		for (const payload of [
			inTrial,
			standingBefore,
			withReview({}, { mistakes: [] }),
		]) {
			takeDown(drawn?.root);
			const { root } = await drawOpened(payload);

			expect(root.querySelector(".mt-review-part")).toBeNull();
			expect(sections(root).map(([title]) => title)).not.toContain("Review");
		}
	});

	test("that does not read is none, and the progress draws all the same, the mistakes in a section of their own", async () => {
		const { root } = await drawOpened({
			...standing,
			review: { strong: "Ordering" },
		});

		expect(sections(root).map(([title]) => title)).toEqual([
			"Topics",
			"Mistakes that repeat",
			"Recent answers",
			"Profile",
		]);
	});
});

describe("the links to the topics' pages", () => {
	// links are the card's links to pages, each the name it is under and where
	// it leads.
	const links = (root: HTMLElement) =>
		[...root.querySelectorAll<HTMLAnchorElement>("a.mt-link")].map((link) => [
			link.textContent,
			link.getAttribute("href"),
		]);

	test("name each topic and each step in a topic by a link to its page, a step's to the part it is about, when the host opens pages", async () => {
		const { root } = await drawOpened(standing, { links: "open" });

		expect(links(root)).toEqual([
			["Ordering", "https://mathtrail.app/en/topics/ordering/"],
			["Enumeration", "https://mathtrail.app/en/topics/enumeration/"],
			[
				"Gaps and boundaries",
				"https://mathtrail.app/en/topics/gaps-and-boundaries/",
			],
			[
				"Parity and alternation",
				"https://mathtrail.app/en/topics/parity-and-alternation/",
			],
			[
				"Pigeonhole principle",
				"https://mathtrail.app/en/topics/pigeonhole-principle/",
			],
			["Enumeration", "https://mathtrail.app/en/topics/enumeration/#traps"],
			[
				"Parity and alternation",
				"https://mathtrail.app/en/topics/parity-and-alternation/#home",
			],
		]);
	});

	test("lead to the pages in the card's language when the site is written in it, and in English when it is not", async () => {
		const inRussian = await drawOpened(
			{ ...standing, profile: { ...standing.profile, ui_language: "ru" } },
			{ links: "open" },
		);
		expect(links(inRussian.root)[0]).toEqual([
			"Упорядочивание",
			"https://mathtrail.app/ru/topics/ordering/",
		]);
		takeDown(inRussian.root);

		const inFrench = await drawOpened(
			{ ...standing, profile: { ...standing.profile, ui_language: "fr" } },
			{ links: "open" },
		);
		expect(links(inFrench.root)[0]?.[1]).toBe(
			"https://mathtrail.app/en/topics/ordering/",
		);
	});

	test("are none for a host that does not say it opens pages, nor from a progress before the pages", async () => {
		const quiet = await drawOpened(standing);
		expect(quiet.root.querySelector("a")).toBeNull();
		takeDown(quiet.root);

		const before = await drawOpened(standingBefore, { links: "open" });
		expect(before.root.querySelector("a")).toBeNull();
	});

	test("ask the host to open the page when pressed, and nothing more", async () => {
		const { root, heard } = await drawOpened(standing, { links: "open" });
		const ordering = root.querySelector<HTMLAnchorElement>("a.mt-link");
		const click = new MouseEvent("click", { bubbles: true, cancelable: true });

		act(() => {
			ordering?.dispatchEvent(click);
		});

		// The card itself goes nowhere: the host is what opens a page.
		expect(click.defaultPrevented).toBe(true);
		await vi.waitFor(() =>
			expect(heard.pages).toEqual([
				"https://mathtrail.app/en/topics/ordering/",
			]),
		);
		expect(root.querySelector(".mt-link-refused:not(:empty)")).toBeNull();
	});

	test("ask the host for the middle button too, and leave the browser's menu to the browser", async () => {
		const { root, heard } = await drawOpened(standing, { links: "open" });
		const ordering = root.querySelector<HTMLAnchorElement>("a.mt-link");
		const middle = new MouseEvent("auxclick", {
			bubbles: true,
			cancelable: true,
			button: 1,
		});
		const menu = new MouseEvent("auxclick", {
			bubbles: true,
			cancelable: true,
			button: 2,
		});

		act(() => {
			ordering?.dispatchEvent(menu);
			ordering?.dispatchEvent(middle);
		});

		expect(middle.defaultPrevented).toBe(true);
		expect(menu.defaultPrevented).toBe(false);
		await vi.waitFor(() =>
			expect(heard.pages).toEqual([
				"https://mathtrail.app/en/topics/ordering/",
			]),
		);
	});

	test("ask once for presses made while the host is still answering", async () => {
		const { root, heard } = await drawOpened(standing, { links: "refuse" });
		const ordering = root.querySelector<HTMLAnchorElement>("a.mt-link");

		act(() => {
			ordering?.click();
			ordering?.click();
		});

		await vi.waitFor(() =>
			expect(root.querySelector(".mt-link-refused:not(:empty)")).not.toBeNull(),
		);
		expect(heard.pages).toEqual(["https://mathtrail.app/en/topics/ordering/"]);
	});

	test.each([
		["refuses", "refuse" as const],
		["fails to answer", "fail" as const],
	])(
		"show the address under the name when the host %s, and keep the link to try again",
		async (_, links) => {
			const { root, heard } = await drawOpened(standing, { links });
			const ordering = root.querySelector<HTMLAnchorElement>("a.mt-link");
			// The line is in its place, empty, before anything is said in it:
			// a screen reader hears what a place it knows is filled with.
			const place = ordering?.parentElement?.querySelector(".mt-link-refused");
			expect(place?.getAttribute("aria-live")).toBe("polite");
			expect(place?.textContent).toBe("");

			press(ordering ?? root);

			const refused = await vi.waitFor(() => {
				const line = root.querySelector(".mt-link-refused:not(:empty)");
				expect(line).not.toBeNull();
				return line;
			});
			expect(refused).toBe(place);
			expect(refused?.firstElementChild?.textContent).toBe(
				"The chat didn't open the page. Its address:",
			);
			const address = refused?.querySelector(".mt-link-address");
			expect(address?.textContent).toBe(
				"https://mathtrail.app/en/topics/ordering/",
			);
			expect(address?.getAttribute("dir")).toBe("ltr");
			expect(document.activeElement).toBe(ordering);

			press(ordering ?? root);

			await vi.waitFor(() => expect(heard.pages).toHaveLength(2));
		},
	);

	test("are drawn in the progress over a task too", async () => {
		const { root } = await draw(fence, {
			tools: ({ name }) => (name === "read_progress" ? progress : failure),
			links: "open",
		});

		press(root.querySelector<HTMLElement>(".mt-bar:not(.mt-bar-back)") ?? root);
		await vi.waitFor(() => unfold(root, "Topics"));

		expect(links(root)[0]).toEqual([
			"Ordering",
			"https://mathtrail.app/en/topics/ordering/",
		]);
	});
});

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
			"↓The last task moved the bar back",
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
		canOpenLinks: () => false,
		openLink: () => Promise.reject(new Error("no host")),
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
			"↓The last task moved the bar back",
		);
		act(() => render(null, root));
		root.remove();
	});
});

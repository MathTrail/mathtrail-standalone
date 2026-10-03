import { act } from "preact/test-utils";
import { afterEach, describe, expect, test, vi } from "vitest";
import {
	buttonIn,
	type Drawn,
	drawCard,
	press,
	takeDown,
} from "./testing/card";
import { atTheTop, inTrial, standing, standingBefore } from "./testing/lesson";

let drawn: Drawn | undefined;

afterEach(() => {
	takeDown(drawn?.root);
	drawn = undefined;
	vi.restoreAllMocks();
});

// draw draws the progress card of payload.
async function draw(
	payload: object,
	options: Parameters<typeof drawCard>[1] = {},
): Promise<Drawn> {
	drawn = await drawCard(payload, options);
	return drawn;
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

// lists are the card's lists of lines by their labels, each a list of its
// lines.
function lists(root: HTMLElement): Record<string, (string | null)[]> {
	return Object.fromEntries(
		[...root.querySelectorAll(".mt-list")]
			.filter((list) => list.querySelector(".mt-rank-row") === null)
			.map((list) => [
				list.querySelector(".mt-section-label")?.textContent ?? "",
				[...list.querySelectorAll(".mt-row")].map((row) => row.textContent),
			]),
	);
}

// fields are the fields of the group labelled label: each its name, what it
// holds — a list's names joined — and the note under it, when there is one.
function fields(root: HTMLElement, label: string): string[][] {
	const group = [...root.querySelectorAll(".mt-fields")].find(
		(found) =>
			found.querySelector(".mt-fields-head .mt-section-label")?.textContent ===
			label,
	);
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
	test("names the rank reached over the rank out of how many and the rating, its course filled as far as it has come, and the next rank", async () => {
		const { root } = await draw(standing);

		const rank = root.querySelector("section.mt-rank");
		expect(rank?.getAttribute("aria-label")).toBe("Overall rating");
		expect(text(root, ".mt-rank-name")).toBe("River crossing");
		expect(text(root, ".mt-rank-head .mt-meta")).toBe(
			"rank 3 of 11 · rating 1573",
		);
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

	test("names what comes next, and that it comes again after a mistake", async () => {
		const { root } = await draw(standing);

		expect(text(root, ".mt-note-plain .mt-note-label")).toBe("Next up");
		expect(text(root, ".mt-note-plain p")).toBe("Enumeration, once more");
	});

	test("shows each topic's rank, how it stands to the overall one, and its course in the colour of its rank", async () => {
		const { root } = await draw(standing);

		expect(text(root, ".mt-list-lead")).toBe("The rank in each topic");
		expect(topics(root)).toEqual([
			[
				"OrderingMastered",
				"aheadHill",
				"2",
				eleven("100%", "100%", "100%", "27%"),
			],
			["Enumeration", "River crossing", "1", eleven("100%", "100%", "76%")],
			[
				"Gaps and boundaries",
				"River crossing",
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
		const { root } = await draw({ ...standing, skipped: 4 });

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

	test("lists the mistakes that keep coming back after the latest answers, a dot for each time, in a frame of their own", async () => {
		const { root } = await draw(standing);

		expect(Object.keys(lists(root))).toEqual([
			"Recent answers",
			"Mistakes that repeat",
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

	test("shows no list of mistakes when none has come up twice", async () => {
		const { root } = await draw({ ...standing, mistakes: [] });

		expect(Object.keys(lists(root))).toEqual(["Recent answers"]);
		expect(root.querySelector(".mt-dots")).toBeNull();
	});

	test("with no answer yet shows the series to come, and no list", async () => {
		const { root } = await draw({
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
		expect(fields(root, "Profile · for the parent")).toHaveLength(4);
	});

	test("says nothing of skipped tasks when the service counts none", async () => {
		const { root } = await draw({ ...standing, skipped: 0 });

		expect(root.querySelector(".mt-list-note")).toBeNull();
	});

	test("from an earlier release still draws: its course with the step under way empty, its topics with no rank, the skips its topics count, and no data", async () => {
		const { root } = await draw(standingBefore);

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
		const { root } = await draw(standing);

		expect(fields(root, "Profile · for the parent")).toEqual([
			["Grade", "3", "Only a label: changing it moves no rating."],
			["Interests", "space, animals, football"],
			["Not at school yet", "Division with a remainder"],
			["Language of the lessons", "The chat's language"],
		]);
		expect(root.querySelectorAll(".mt-chip")).toHaveLength(4);
	});

	test("says a list the parent left empty is not set", async () => {
		const { root } = await draw({
			...standing,
			profile: { ...standing.profile, interests: [], excluded_skills: [] },
		});

		expect(fields(root, "Profile · for the parent").slice(1, 3)).toEqual([
			["Interests", "Not set"],
			["Not at school yet", "Not set"],
		]);
	});

	test("says where the profile's file is and what the parent can do with the data, each under the question it answers", async () => {
		const { root } = await draw(standing);

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
		expect(fields(root, "Profile · for the parent")).toHaveLength(4);
	});

	test("in the trial series shows how far it has got, with no rank anywhere and every topic drawn open", async () => {
		const { root } = await draw(inTrial);

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
		const { root } = await draw({
			...standing,
			profile: { ...standing.profile, ui_language: "ru" },
		});

		expect(text(root, ".mt-rank-name")).toBe("Брод");
		expect(text(root, ".mt-rank-head .mt-meta")).toBe(
			"ранг 3 из 11 · рейтинг 1573",
		);
		expect(text(root, ".mt-rank-line")).toBe(
			"Следующий ранг — Холм. Верные ответы двигают полоску вперёд, ошибки — чуть назад.",
		);
		expect(text(root, ".mt-note-plain p")).toBe("Перебор, ещё раз");
		expect(topics(root).map(([, standsAt]) => standsAt)).toEqual([
			"впередиХолм",
			"Брод",
			"Брод",
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
		expect(fields(root, "Профиль · для родителя")[3]).toEqual([
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

describe("editing the profile", () => {
	test("is asked for in the chat, in the words of the button at the profile's head", async () => {
		const { root, heard } = await draw(standing);

		const edit = buttonIn(root, "Edit profile");
		expect(edit.closest(".mt-fields-head")).not.toBeNull();
		press(edit);

		await vi.waitFor(() => expect(heard.messages).toEqual(["Edit profile"]));
		await vi.waitFor(() =>
			expect(text(root, ".mt-action-note")).toBe("Sent to the chat"),
		);
		expect(heard.calls).toEqual([]);
	});

	test("is asked for once when two presses come before the card redraws", async () => {
		const { root, heard } = await draw(standing);
		const edit = buttonIn(root, "Edit profile");

		act(() => {
			edit.click();
			edit.click();
		});

		await vi.waitFor(() => expect(heard.messages).toHaveLength(1));
		await vi.waitFor(() =>
			expect(text(root, ".mt-action-note")).toBe("Sent to the chat"),
		);
		expect(heard.messages).toHaveLength(1);
	});

	test("that the chat took is not asked again from the same button", async () => {
		const { root, heard } = await draw(standing);
		const edit = buttonIn(root, "Edit profile");
		press(edit);
		await vi.waitFor(() =>
			expect(text(root, ".mt-action-note")).toBe("Sent to the chat"),
		);

		press(edit);

		expect(edit.getAttribute("aria-disabled")).toBe("true");
		expect(document.activeElement).toBe(edit);
		await Promise.resolve();
		expect(heard.messages).toEqual(["Edit profile"]);
	});

	test("that the chat does not take says so, and can be asked again", async () => {
		vi.spyOn(console, "error").mockImplementation(() => {});
		const { root, heard } = await draw(standing, { refuseMessages: true });

		press(buttonIn(root, "Edit profile"));

		await vi.waitFor(() =>
			expect(text(root, ".mt-action-note")).toBe("Not sent — try again"),
		);
		press(buttonIn(root, "Edit profile"));
		await vi.waitFor(() => expect(heard.messages).toHaveLength(2));
	});

	test("has its outcome heard once it comes", async () => {
		const { root } = await draw(standing);

		const note = root.querySelector(".mt-action-note");
		expect(note?.getAttribute("aria-live")).toBe("polite");
		expect(note?.textContent).toBe("");
	});
});

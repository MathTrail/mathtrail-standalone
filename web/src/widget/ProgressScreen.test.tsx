import { act } from "preact/test-utils";
import { afterEach, describe, expect, test, vi } from "vitest";
import {
	buttonIn,
	type Drawn,
	drawCard,
	press,
	takeDown,
} from "./testing/card";
import { inTrial, standing } from "./testing/lesson";

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

// lists are the card's lists by their labels, each a list of its lines.
function lists(root: HTMLElement): Record<string, (string | null)[]> {
	return Object.fromEntries(
		[...root.querySelectorAll(".mt-list")].map((list) => [
			list.querySelector(".mt-section-label")?.textContent ?? "",
			[...list.querySelectorAll(".mt-row")].map((row) => row.textContent),
		]),
	);
}

// bars are how far the bars of the card's lists are filled, in order.
function bars(root: HTMLElement): (string | null)[] {
	return [...root.querySelectorAll(".mt-bar-fill")].map((fill) =>
		fill.getAttribute("width"),
	);
}

// fields are the fields of the profile's section labelled label.
function fields(root: HTMLElement, label: string): string[][] {
	const section = [...root.querySelectorAll(".mt-fields")].find(
		(found) => found.querySelector(".mt-section-label")?.textContent === label,
	);
	return [...(section?.querySelectorAll("dl > div") ?? [])].map((field) =>
		[...field.children].map((part) => part.textContent ?? ""),
	);
}

describe("the progress", () => {
	test("puts the overall rating under its rank and the rank's name", async () => {
		const { root } = await draw(standing);

		const rating = root.querySelector("section.mt-rating");
		expect(rating?.getAttribute("aria-label")).toBe("Overall rating");
		expect(text(root, ".mt-rating-rank")).toBe("Rank 3 of 11 · River crossing");
		expect(text(root, ".mt-rating-num")).toBe("1573");
		expect(text(root, ".mt-rating .mt-meta")).toBe("overall rating");
		expect(
			[...root.querySelectorAll(".mt-pip")].map((pip) =>
				pip.getAttribute("data-on"),
			),
		).toEqual([
			...Array.from({ length: 3 }, () => "true"),
			...Array.from({ length: 8 }, () => "false"),
		]);
	});

	test("says whose it is, with no way back to a task it never had", async () => {
		const { root } = await draw(standing);

		expect(text(root, ".mt-head .mt-name")).toBe("Comet");
		expect(text(root, ".mt-badge")).toBe("Grade 3");
		expect(root.querySelector(".mt-bar")).toBeNull();
	});

	test("shows no number when it is told neither a rating nor a trial series", async () => {
		const { root } = await draw({ ...standing, overall: null });

		expect(root.querySelector(".mt-rating")).toBeNull();
		expect(root.querySelector(".mt-list")).not.toBeNull();
	});

	test("names what comes next, and that it comes again after a mistake", async () => {
		const { root } = await draw(standing);

		expect(text(root, ".mt-note-plain .mt-note-label")).toBe("Next up");
		expect(text(root, ".mt-note-plain p")).toBe("Enumeration, once more");
	});

	test("lists the topics met with their ratings, and the latest five entries", async () => {
		const { root } = await draw(standing);

		const shown = lists(root);
		expect(shown.Topics).toEqual([
			"OrderingMastered1712",
			"Enumeration1627",
			"Gaps and boundaries1588",
			"Parity and alternation1541",
		]);
		expect(shown["Recent answers"]).toEqual([
			"EnumerationWrong",
			"EnumerationSkipped",
			"Gaps and boundariesRight",
			"OrderingRight",
			"Parity and alternationWrong",
		]);
		expect(text(root, ".mt-list-note")).toBe(
			"In all, 1 task was left without an answer.",
		);
	});

	test("lists the mistakes that keep coming back after the latest answers, each with a bar as long as its share of the most frequent", async () => {
		const { root } = await draw(standing);

		expect(Object.keys(lists(root))).toEqual([
			"Topics",
			"Recent answers",
			"Mistakes that repeat",
		]);
		expect(lists(root)["Mistakes that repeat"]).toEqual([
			"Missed a case while listing3 times",
			"Counted the same thing twice2 times",
		]);
		expect(bars(root)).toEqual(["100%", "67%"]);
	});

	test("shows no list of mistakes when none has come up twice", async () => {
		const { root } = await draw({ ...standing, mistakes: [] });

		expect(Object.keys(lists(root))).toEqual(["Topics", "Recent answers"]);
		expect(root.querySelector(".mt-bar-chart")).toBeNull();
	});

	test("with no answer yet shows the series to come, and no list", async () => {
		const { root } = await draw({
			...inTrial,
			trial: { answered: 0, of: 5 },
			topics: [],
			recent: [],
			mistakes: [],
		});

		expect(text(root, ".mt-rating-num")).toBe("0 of 5");
		expect(root.querySelector(".mt-list")).toBeNull();
		expect(fields(root, "Profile · for the parent")).toHaveLength(4);
	});

	test("says nothing of skipped tasks when none was skipped", async () => {
		const { root } = await draw({
			...standing,
			topics: standing.topics.map((topic) => ({ ...topic, skipped: 0 })),
		});

		expect(root.querySelector(".mt-list-note")).toBeNull();
	});

	test("shows the child's profile for the parent", async () => {
		const { root } = await draw(standing);

		expect(fields(root, "Profile · for the parent")).toEqual([
			["Grade", "3", "Only a label: changing it moves no rating."],
			["Interests", "space, animals, football"],
			["Not at school yet", "Division with a remainder"],
			["Language of the cards", "The chat's language"],
		]);
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

	test("in the trial series shows how far it has got, and no rating nor rank", async () => {
		const { root } = await draw(inTrial);

		// The series is named once, by the line above its number.
		expect(
			root.querySelector("section.mt-rating")?.hasAttribute("aria-label"),
		).toBe(false);
		expect(text(root, ".mt-rating-rank")).toBe("Trial series");
		expect(text(root, ".mt-rating-num")).toBe("3 of 5");
		expect(text(root, ".mt-rating .mt-meta")).toBe(
			"the rating comes after 5 tasks",
		);
		expect(root.querySelectorAll(".mt-pip")).toHaveLength(5);
		expect(root.textContent).not.toContain("Rank");
		expect(lists(root).Topics).toEqual([
			"Ordering",
			"Gaps and boundaries",
			"Clocks",
		]);
		expect(text(root, ".mt-note-plain p")).toBe("Parity and alternation");
	});

	test("is in the language the parent chose for the cards", async () => {
		const { root } = await draw({
			...standing,
			profile: { ...standing.profile, ui_language: "ru" },
		});

		expect(text(root, ".mt-rating-rank")).toBe("Ранг 3 из 11 · Брод");
		expect(text(root, ".mt-note-plain p")).toBe("Перебор, ещё раз");
		expect(text(root, ".mt-list-note")).toBe(
			"Всего без ответа оставили 1 задачу.",
		);
		expect(lists(root)["Повторяющиеся ошибки"]).toEqual([
			"Пропущен случай при переборе3 раза",
			"Одно и то же посчитано дважды2 раза",
		]);
		expect(fields(root, "Профиль · для родителя")[3]).toEqual([
			"Язык карточек",
			"Русский",
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
	test("is asked for in the chat, in the button's own words", async () => {
		const { root, heard } = await draw(standing);

		press(buttonIn(root, "Edit profile"));

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

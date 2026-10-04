import { Window } from "happy-dom";
import { renderToString } from "preact-render-to-string";
import { afterAll, afterEach, describe, expect, test, vi } from "vitest";
import type { Section } from "../widget/folds";
import { readAnswer, readHandedTask, readScreen } from "../widget/payload";
import {
	type Drawn,
	drawCard,
	press,
	takeDown,
	unfold,
} from "../widget/testing/card";
import {
	answered,
	fenceInRussian,
	fenceSolutionInRussian,
	standing,
} from "../widget/testing/lesson";
import { StaticAnswer, StaticProgress } from "./StaticCard";

// browser reads the static card the way a page holds it, apart from the card
// the widget draws in this test's own document.
const browser = new Window();

afterAll(async () => {
	await browser.happyDOM.close();
});

let drawn: Drawn | undefined;

afterEach(() => {
	takeDown(drawn?.root);
	drawn = undefined;
});

// The widget's own example of a progress, in Russian, so that the language is
// the page's rather than the default one.
const payload = {
	...standing,
	profile: { ...standing.profile, ui_language: "ru" },
};

// still is the static card drawn for payload with sections open, read back
// as a page holds it.
function still(open: Section[]) {
	const screen = readScreen(payload);
	if (screen?.screen !== "progress") {
		throw new Error("the example is no progress");
	}
	const html = renderToString(
		<StaticProgress report={screen.report} locale="ru" open={new Set(open)} />,
	);
	return new browser.DOMParser().parseFromString(html, "text/html");
}

// Part is an element of a card, whichever document it was read into.
type Part = {
	tagName: string;
	getAttribute(name: string): string | null;
	querySelectorAll(selector: string): Iterable<Part>;
};

// shape is every element under a card, in order, as its tag and its classes:
// what the card is made of, apart from the ids each drawing numbers anew.
function shape(card: Part | null | undefined): string[] {
	return [...(card?.querySelectorAll("*") ?? [])].map(
		(element) =>
			`${element.tagName.toLowerCase()}.${element.getAttribute("class") ?? ""}`,
	);
}

describe("a progress card drawn on a page", () => {
	test("is the card the widget draws in a chat for the same progress, with the same section open", async () => {
		drawn = await drawCard(payload);
		unfold(drawn.root, "Темы");
		const page = still(["topics"]);
		const chat = drawn.root.querySelector(".mt-widget");

		expect(page.querySelector(".mt-widget")?.textContent).toBe(
			chat?.textContent,
		);
		expect(shape(page.querySelector(".mt-widget"))).toEqual(shape(chat));
	});

	test("is inert, since nothing on a page answers its buttons", () => {
		const page = still(["topics"]);

		expect(page.querySelector(".s-card")?.hasAttribute("inert")).toBe(true);
		expect(page.querySelector(".s-card .mt-widget")).not.toBeNull();
	});

	test("is drawn at the narrow width, which a page cannot measure ahead", () => {
		expect(still(["topics"]).querySelector(".mt-wide")).toBeNull();
	});

	test("opens the sections the page asks for, and only those", () => {
		const open = [
			...still(["topics"]).querySelectorAll(
				'.mt-fold-button[aria-expanded="true"]',
			),
		].map((button) => button.querySelector(".mt-fold-title")?.textContent);

		expect(open).toEqual(["Темы"]);
	});
});

// russianFence is the fence handed out in a lesson in Russian, which the
// service names beside the task, as it does for every task it hands out.
const russianFence = { ...fenceInRussian, language: "ru" };

// wrong is the fence answered with the wrong B, as the service records it for
// a card in Russian.
const wrong = answered({
	trap: {
		id: "off_by_one",
		text: "Посчитаны промежутки, а не столбы.",
		repeated: false,
	},
	solution: fenceSolutionInRussian,
});

// answeredStill is the static card of the fence once wrong is recorded, read
// back as a page holds it.
function answeredStill() {
	const handed = readHandedTask(russianFence);
	const told = readAnswer(wrong, russianFence.task.id);
	if (handed === undefined || told.kind !== "answered") {
		throw new Error("the example is no task answered");
	}
	const html = renderToString(
		<StaticAnswer handed={handed} result={told.result} locale="ru" />,
	);
	return new browser.DOMParser().parseFromString(html, "text/html");
}

describe("a card of a wrong answer drawn on a page", () => {
	test("is the card the widget draws in a chat once the same answer is recorded", async () => {
		drawn = await drawCard(russianFence, { tools: () => wrong });
		const pressed = [
			...drawn.root.querySelectorAll<HTMLButtonElement>(".mt-option"),
		].find(
			(row) => row.querySelector(".mt-option-letter")?.textContent === "B",
		);
		if (pressed === undefined) {
			throw new Error("the card has no option B");
		}
		press(pressed);
		const root = drawn.root;
		await vi.waitFor(() =>
			expect(root.querySelector(".mt-replies .mt-reply")).not.toBeNull(),
		);
		const page = answeredStill();
		const chat = root.querySelector(".mt-widget");

		expect(page.querySelector(".mt-widget")?.textContent).toBe(
			chat?.textContent,
		);
		expect(shape(page.querySelector(".mt-widget"))).toEqual(shape(chat));
	});

	test("is inert, and drawn at the narrow width", () => {
		const page = answeredStill();

		expect(page.querySelector(".s-card")?.hasAttribute("inert")).toBe(true);
		expect(page.querySelector(".s-card .mt-widget")).not.toBeNull();
		expect(page.querySelector(".mt-wide")).toBeNull();
	});

	test("marks the option picked wrong and the right one right, by their letters", () => {
		const states = [...answeredStill().querySelectorAll(".mt-option")].map(
			(row) =>
				`${row.querySelector(".mt-option-letter")?.textContent} ${row.getAttribute("data-state")}`,
		);

		expect(states).toEqual([
			"A muted",
			"B wrong",
			"C correct",
			"D muted",
			"E muted",
		]);
	});
});

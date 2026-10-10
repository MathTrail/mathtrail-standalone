import { Window } from "happy-dom";
import type { VNode } from "preact";
import { renderToString } from "preact-render-to-string";
import {
	afterAll,
	afterEach,
	beforeEach,
	describe,
	expect,
	test,
	vi,
} from "vitest";
import type { Section } from "../widget/folds";
import { lessonStart } from "../widget/lesson";
import { readAnswer, readHandedTask, readScreen } from "../widget/payload";
import {
	buttonIn,
	type Drawn,
	drawCard,
	press,
	takeDown,
	unfold,
} from "../widget/testing/card";
import {
	answered,
	coming,
	fenceInRussian,
	fenceRow,
	fenceSolutionInRussian,
	fenceSolutionRow,
	fenceTotal,
	shown,
	standing,
} from "../widget/testing/lesson";
import {
	StaticComing,
	StaticProgress,
	StaticResult,
	StaticTask,
} from "./StaticCard";
import { answerOf, handedOf } from "./taskcard";

// browser reads the static card the way a page holds it, apart from the card
// the widget draws in this test's own document.
const browser = new Window();

afterAll(async () => {
	await browser.happyDOM.close();
});

let drawn: Drawn | undefined;

// The cards are drawn as a site built from a release draws them: knowing the
// build, and naming it as the cards in a chat do.
beforeEach(() => {
	vi.stubEnv("VITE_VERSION", "v0.2.1");
});

afterEach(() => {
	takeDown(drawn?.root);
	drawn = undefined;
	vi.unstubAllEnvs();
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

// wrongTold is how the fence answered with the wrong B went, as the service
// records it for a card in Russian.
const wrongTold = {
	trap: {
		id: "off_by_one",
		text: "Посчитаны промежутки, а не столбы.",
		repeated: false,
	},
	solution: fenceSolutionInRussian,
};
const wrong = answered(wrongTold);

// answeredStill is the static card of how the fence went once wrong is
// recorded, read back as a page holds it.
function answeredStill() {
	const handed = readHandedTask(russianFence);
	const told = readAnswer(wrong, russianFence.task.id);
	if (handed === undefined || told.kind !== "answered") {
		throw new Error("the example is no task answered");
	}
	const html = renderToString(
		<StaticResult handed={handed} result={told.result} locale="ru" />,
	);
	return new browser.DOMParser().parseFromString(html, "text/html");
}

describe("a card of how a wrong answer went, drawn on a page", () => {
	test("is the card the widget draws in a chat for the same answer", async () => {
		drawn = await drawCard(shown(russianFence, wrongTold));
		const root = drawn.root;
		await vi.waitFor(() =>
			expect(root.querySelector(".mt-verdict-line")).not.toBeNull(),
		);
		sameCard(answeredStill(), root);
	});

	test("names the release the site was built from, as a chat names it", () => {
		const page = answeredStill();

		expect(page.querySelector(".mt-version-label")?.textContent).toBe("версия");
		expect(page.querySelector(".mt-version-number")?.textContent).toBe("0.2.1");
	});

	test("names no build when the site was built from no release, rather than dev", () => {
		vi.stubEnv("VITE_VERSION", "");

		const page = answeredStill();

		expect(page.querySelector(".mt-head")).not.toBeNull();
		expect(page.querySelector(".mt-version")).toBeNull();
		expect(page.querySelector(".mt-head-versioned")).toBeNull();
	});

	test("is inert, and drawn at the narrow width", () => {
		const page = answeredStill();

		expect(page.querySelector(".s-card")?.hasAttribute("inert")).toBe(true);
		expect(page.querySelector(".s-card .mt-widget")).not.toBeNull();
		expect(page.querySelector(".mt-wide")).toBeNull();
	});

	test("tells the verdict, the trap and the solution step by step, with the rating, and names the topic with no link", () => {
		const page = answeredStill();

		expect(page.querySelector(".mt-verdict-line")?.textContent).toBe(
			"Не совсем: ответ 5, а не 4.",
		);
		expect(page.querySelector(".mt-note-trap p")?.textContent).toBe(
			"Посчитаны промежутки, а не столбы.",
		);
		expect(page.querySelectorAll(".mt-steps li")).toHaveLength(3);
		expect(page.querySelector(".mt-rating-move")?.textContent).toBe(
			"1502 → 1480",
		);
		expect(page.querySelector(".mt-fact a")).toBeNull();
		expect(page.querySelector(".mt-option")).toBeNull();
	});
});

// never is an answer that never comes: a card waiting on it stays as it is.
const never = () => new Promise<never>(() => {});

// pageOf reads a card drawn on a page back as a page holds it.
function pageOf(card: VNode) {
	return new browser.DOMParser().parseFromString(
		renderToString(card),
		"text/html",
	);
}

// handedFence is the fence in Russian as a card reads a task handed out.
function handedFence() {
	const handed = readHandedTask(russianFence);
	if (handed === undefined) {
		throw new Error("the example is no task handed out");
	}
	return handed;
}

// optionIn is the option of root whose letter is letter.
function optionIn(root: HTMLElement, letter: string): HTMLButtonElement {
	const found = [
		...root.querySelectorAll<HTMLButtonElement>(".mt-option"),
	].find(
		(row) => row.querySelector(".mt-option-letter")?.textContent === letter,
	);
	if (found === undefined) {
		throw new Error(`the card has no option ${letter}`);
	}
	return found;
}

// sameCard holds a card drawn on a page to the card a chat drew, the build it
// names included, but for the line at its top: a page keeps no profile, so
// that line says whose card it is and leads nowhere.
function sameCard(page: ReturnType<typeof pageOf>, root: HTMLElement) {
	const chat = root.querySelector(".mt-widget");
	expect(belowTheLine(page.querySelector(".mt-widget"))).toEqual(
		belowTheLine(chat),
	);
	expect(topLine(page)).toEqual({
		tag: "div",
		text: chat?.querySelector(".mt-bar-name")?.textContent,
		action: false,
	});
}

// Copied is a part of a card that can be copied and taken apart, whichever
// document it was read into.
type Copied = Part & {
	cloneNode(deep: boolean): unknown;
	textContent: string | null;
};

// belowTheLine is a card's text and what it is made of, with the line at its
// top taken out.
function belowTheLine(card: Copied | null | undefined) {
	const copy = card?.cloneNode(true) as
		| (Copied & { querySelector(selector: string): { remove(): void } | null })
		| undefined;
	copy?.querySelector(".mt-bar")?.remove();
	return { text: copy?.textContent, shape: shape(copy) };
}

// topLine is the line at the top of a card drawn on a page: the element it is,
// what it says, and whether it offers the way to the progress.
function topLine(page: ReturnType<typeof pageOf>) {
	const line = page.querySelector(".mt-widget .mt-bar");
	return {
		tag: line?.tagName.toLowerCase(),
		text: line?.textContent,
		action: line?.querySelector(".mt-bar-action, .mt-chevron") != null,
	};
}

describe("a card of a task drawn on a page from where its lesson stands", () => {
	test("with an option picked and being checked, is the card a chat draws while the answer is on its way", async () => {
		drawn = await drawCard(russianFence, { tools: never });
		press(optionIn(drawn.root, "B"));
		const root = drawn.root;
		await vi.waitFor(() =>
			expect(optionIn(root, "B").dataset.state).toBe("selected"),
		);

		sameCard(
			pageOf(
				<StaticTask
					handed={handedFence()}
					start={{
						hint: { open: false, used: false },
						answer: { state: "checking", choice: "B" },
					}}
					locale="ru"
				/>,
			),
			root,
		);
	});

	test("with its hint open, is the card a chat draws once the hint is pressed", async () => {
		drawn = await drawCard(russianFence, { tools: never });
		press(buttonIn(drawn.root, "Подсказка"));
		const root = drawn.root;
		await vi.waitFor(() =>
			expect(root.querySelector(".mt-note-hint")).not.toBeNull(),
		);

		sameCard(
			pageOf(
				<StaticTask
					handed={handedFence()}
					start={{
						hint: { open: true, used: true },
						answer: { state: "open" },
					}}
					locale="ru"
				/>,
			),
			root,
		);
	});
});

describe("a card of a task being written, drawn on a page", () => {
	// comingInRussian is a task asked for in a lesson in Russian.
	const comingInRussian = {
		...coming,
		child: fenceInRussian.child,
		language: "ru",
	};

	test("is the card a chat draws before the service has said anything of the task", async () => {
		drawn = await drawCard(comingInRussian, { tools: never });

		sameCard(
			pageOf(
				<StaticComing
					coming={{
						requestId: comingInRussian.request_id,
						child: comingInRussian.child,
					}}
					locale="ru"
				/>,
			),
			drawn.root,
		);
	});

	test("is inert, and names no build when the site was built from no release", () => {
		vi.stubEnv("VITE_VERSION", "");

		const page = pageOf(
			<StaticComing
				coming={{ requestId: "req", child: fenceInRussian.child }}
				locale="ru"
			/>,
		);

		expect(page.querySelector(".s-card")?.hasAttribute("inert")).toBe(true);
		expect(page.querySelector(".mt-version")).toBeNull();
	});
});

describe("a card of a task the site draws", () => {
	// facts are the facts of a card of the site's, and said its words.
	const facts = {
		topic: "counting.gaps",
		grade: 3,
		picture: fenceRow,
		options: { A: "3", B: "4", C: "5", D: "6", E: "12" },
		solution_picture: fenceSolutionRow,
		solution_total: fenceTotal,
		choice: "B",
		correct: "C",
		trap: "fence_gaps",
		rating: { before: 1502, after: 1480 },
	};
	const said = { language: "ru", child: "Комета", question: "?", hint: "?" };

	// A picture the card would leave out would leave the site's card without
	// it, and the task would not read as the page means it to.
	test("whose picture the card cannot draw stops the build", () => {
		const unreadable = { ...facts, picture: { kind: "pie" } as never };

		expect(() =>
			handedOf(unreadable, said, "site_card", "the card of this test"),
		).toThrow("the card of this test has a picture the widget cannot draw");
	});

	// The picture of the solution and the equality under it are the card of
	// how its answer went, as a task's are: one the card would leave out stops
	// the build, as the task's own picture does.
	test("whose picture of the solution, or its total, the card would leave out stops the build", () => {
		const words = { ...said, trap: "?", solution: "?" };
		const where = "the card of this test";

		expect(
			answerOf(facts, words, "site_card", where).result.solution_total,
		).toBe(fenceTotal);
		expect(() =>
			answerOf(
				{ ...facts, solution_picture: { kind: "pie" } as never },
				words,
				"site_card",
				where,
			),
		).toThrow(
			"the card of this test has a picture of its solution the widget cannot draw",
		);
		expect(() =>
			answerOf(
				{ ...facts, solution_total: "1".repeat(33) },
				words,
				"site_card",
				where,
			),
		).toThrow(
			"the card of this test writes under the picture of its solution a total the widget does not read",
		);
	});

	// A card handed no choice of the topic, as the card of the page Why is,
	// shows none, as a card of the trial series does.
	test("handed no choice of the topic, offers none", () => {
		const handed = handedOf(facts, said, "site_card", "the card of this test");
		const page = pageOf(
			<StaticTask handed={handed} start={lessonStart} locale="ru" />,
		);

		expect(handed.topic_choice).toBeUndefined();
		expect(page.querySelector(".mt-btns .mt-btn")).not.toBeNull();
		expect(page.querySelector(".mt-topic-button")).toBeNull();
		expect(page.querySelector(".mt-topic-panel")).toBeNull();
	});

	// A page keeps no profile to save a choice in, so a card handed the coach's
	// choice, as the home page's are, shows the button and takes no press on it.
	test("handed the coach's choice, shows the button of the topic locked", () => {
		const handed = handedOf(facts, said, "site_card", "the card of this test", {
			chosen: null,
			recommended: [],
		});
		const page = pageOf(
			<StaticTask handed={handed} start={lessonStart} locale="ru" />,
		);

		expect(handed.topic_choice).toEqual({ chosen: null, recommended: [] });
		expect(
			page.querySelector(".mt-topic-button")?.getAttribute("aria-disabled"),
		).toBe("true");
		expect(page.querySelector(".mt-topic-panel")?.hasAttribute("hidden")).toBe(
			true,
		);
	});

	test("handed the coach's choice, is the card a chat draws for it, but for the lock on its button of the topic", async () => {
		const payload = {
			...russianFence,
			topic_choice: { chosen: null, recommended: [] },
		};
		drawn = await drawCard(payload, { tools: never });
		const handed = readHandedTask(payload);
		if (handed === undefined) {
			throw new Error("the example is no task handed out");
		}
		const page = pageOf(
			<StaticTask handed={handed} start={lessonStart} locale="ru" />,
		);

		sameCard(page, drawn.root);
		expect(
			drawn.root
				.querySelector(".mt-topic-button")
				?.getAttribute("aria-disabled"),
		).toBeNull();
		expect(
			page.querySelector(".mt-topic-button")?.getAttribute("aria-disabled"),
		).toBe("true");
	});
});

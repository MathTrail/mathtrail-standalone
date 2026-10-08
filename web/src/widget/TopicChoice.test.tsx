import type { CallToolResult } from "@modelcontextprotocol/client";
import { act } from "preact/test-utils";
import { afterEach, describe, expect, test, vi } from "vitest";
import { Icon } from "../design/icons";
import { drawingAlone, drawingOf } from "../design/testing/drawing";
import { cardWords } from "./dictionaries";
import { buttonIn, drawCard, openCard, press, takeDown } from "./testing/card";
import { deliver, type Opening, type ToolCall } from "./testing/host";
import {
	answered,
	editGone,
	editRefused,
	failure,
	fence,
	fenceInRussian,
	type Handed,
	progress,
	staleAnswer,
	topicSaved,
	topicWords,
	withTopicChoice,
} from "./testing/lesson";

let root: HTMLElement;

afterEach(() => {
	takeDown(root);
});

const inEnglish = cardWords("en", undefined);
const inRussian = cardWords("ru", undefined);

// service answers the card's calls as the service would: a topic chosen saved
// with the words for the model, an answer recorded as the wrong B, and the
// progress read.
function service({ name, arguments: args }: ToolCall): CallToolResult {
	switch (name) {
		case "edit_profile":
			return topicSaved(
				args.lesson_topic === "" ? null : String(args.lesson_topic),
			);
		case "read_progress":
			return progress;
		default:
			return answered();
	}
}

// draw draws the card a host hands payload to — by default the fence with the
// choice of the topic the trial series over offers — the host answering the
// card's calls with tools, and returns what the host hears.
async function draw(
	payload: Handed = withTopicChoice(fence),
	options: {
		tools?: (call: ToolCall) => CallToolResult | Promise<CallToolResult>;
		links?: Opening;
		refuseMessages?: boolean | number;
		refuseModelLines?: boolean;
	} = {},
) {
	const drawn = await drawCard(payload, { tools: service, ...options });
	root = drawn.root;
	await vi.waitFor(() =>
		expect(root.querySelector(".mt-option")).not.toBeNull(),
	);
	return drawn.heard;
}

// pending is a tool's result that arrives when the test says.
function pending() {
	let arrive: (result: CallToolResult) => void = () => {};
	const result = new Promise<CallToolResult>((resolve) => {
		arrive = resolve;
	});
	return { result, arrive };
}

const button = (label: string) => buttonIn(root, label);

function topicButton(): HTMLButtonElement {
	const found = root.querySelector<HTMLButtonElement>(".mt-topic-button");
	if (found === null) {
		throw new Error("the card has no button for the topic");
	}
	return found;
}

function panel(): HTMLElement {
	const found = root.querySelector<HTMLElement>(".mt-topic-panel");
	if (found === null) {
		throw new Error("the card has no panel of topics");
	}
	return found;
}

// choice is the choice of the panel named label, the first where it is
// offered twice.
function choice(label: string): HTMLButtonElement {
	const found = [
		...root.querySelectorAll<HTMLButtonElement>(".mt-topic-option"),
	].find(
		(option) =>
			option.querySelector(".mt-topic-option-name")?.textContent === label,
	);
	if (found === undefined) {
		throw new Error(`the panel has no choice ${label}`);
	}
	return found;
}

function option(letter: string): HTMLButtonElement {
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

const text = (selector: string) => root.querySelector(selector)?.textContent;
const shownButtons = () =>
	[...root.querySelectorAll<HTMLButtonElement>(".mt-btns .mt-btn")].map(
		(shown) => shown.textContent,
	);
// rows are the rows of the panel, each its name and the names of its choices.
const rows = () =>
	[...root.querySelectorAll(".mt-topic-row")].map((row) => [
		row.querySelector(".mt-topic-row-name a")?.textContent ??
			row.querySelector(".mt-topic-row-name")?.textContent,
		[...row.querySelectorAll(".mt-topic-option-name")].map(
			(name) => name.textContent,
		),
	]);
const pressedIn = () =>
	[
		...root.querySelectorAll(
			".mt-topic-option[aria-pressed='true'] .mt-topic-option-name",
		),
	].map((name) => name.textContent);
const topicNote = () => text(".mt-topic-note");
// askNote is what the card says of its ask for another task.
const askNote = () =>
	root.querySelector(".mt-foot > .mt-action-note:not(.mt-topic-note)");
const requestNote = () => askNote()?.textContent;
const anotherComing =
	"Once the ask reaches the chat, the new task will come below, in a new card.";
const notSent = "Not sent — try again";

// expectDoneWith checks that the chat took the card's ask for another task, and
// the card is done with: it keeps its task, its options are locked, its
// buttons and its choice of the topic are gone, and what it says of the task
// to come has the focus.
async function expectDoneWith() {
	await vi.waitFor(() => expect(requestNote()).toBe(anotherComing));
	expect(text(".mt-task-text")).toBe(fence.task.question);
	expect(option("A").getAttribute("aria-disabled")).toBe("true");
	expect(root.querySelector(".mt-btns")).toBeNull();
	expect(root.querySelector(".mt-topic-panel")).toBeNull();
	expect(askNote()?.getAttribute("tabindex")).toBe("-1");
	expect(document.activeElement).toBe(askNote());
}

// open opens the panel as a person does.
function open(): void {
	press(topicButton());
}

// The parts of the topic's button: the icon before its words, its label and
// the value after the label.
const topicIcon = () => drawingOf(topicButton().querySelector("svg"));
const topicLabel = () =>
	topicButton().querySelector(".mt-topic-button-text")?.firstElementChild;
const topicValue = () =>
	topicButton().querySelector(".mt-topic-button-value")?.textContent;

// asWide has a card measure as a wide one does: the design's tokens give the
// width a wide card starts at, and the room the card is observed to have is
// more than that. Every other style reads as it is.
function asWide(): void {
	vi.stubGlobal(
		"ResizeObserver",
		class {
			readonly tell: ResizeObserverCallback;
			constructor(tell: ResizeObserverCallback) {
				this.tell = tell;
			}
			observe() {
				const room = {
					borderBoxSize: [{ inlineSize: 736, blockSize: 0 }],
					contentRect: { width: 736 },
				} as unknown as ResizeObserverEntry;
				this.tell([room], this as unknown as ResizeObserver);
			}
			disconnect() {}
		},
	);
	const read = CSSStyleDeclaration.prototype.getPropertyValue;
	vi.spyOn(
		CSSStyleDeclaration.prototype,
		"getPropertyValue",
	).mockImplementation(function (this: CSSStyleDeclaration, name: string) {
		return name === "--widget-wide" ? "640px" : read.call(this, name);
	});
}

describe("the button that chooses the topic", () => {
	test("stands at the end of the row, saying the coach chooses, and names the panel it opens", async () => {
		await draw();

		expect(shownButtons()).toEqual(["Hint", "Another task", "Topic: Coach"]);
		expect(topicButton().getAttribute("aria-expanded")).toBe("false");
		expect(topicButton().getAttribute("aria-controls")).toBe(panel().id);
		expect(panel().hidden).toBe(true);
	});

	test("names the topic the lessons are kept to", async () => {
		await draw(withTopicChoice(fence, { chosen: "time.clocks" }));

		expect(topicButton().textContent).toBe("Topic: Clocks");
	});

	test("on a narrow card marks the coach's choice by a spark and a word, its label left to a screen reader", async () => {
		await draw();

		expect(topicIcon()).toEqual(drawingAlone(<Icon name="sparkle" />));
		expect(topicValue()).toBe("Coach");
		expect(topicLabel()?.textContent).toBe("Topic:");
		expect(topicLabel()?.className).toBe("mt-vh");
	});

	test("on a narrow card marks a topic chosen by its tag and its name, its label left to a screen reader", async () => {
		await draw(withTopicChoice(fence, { chosen: "time.clocks" }));

		expect(topicIcon()).toEqual(drawingAlone(<Icon name="tag" />));
		expect(topicValue()).toBe("Clocks");
		expect(topicLabel()?.className).toBe("mt-vh");
	});

	describe("on a wide card", () => {
		afterEach(() => {
			vi.restoreAllMocks();
			vi.unstubAllGlobals();
		});

		test("says the coach chooses, under its label, by a spark", async () => {
			asWide();
			await draw();

			expect(root.querySelector(".mt-wide")).not.toBeNull();
			expect(topicButton().textContent).toBe("Topic: the coach chooses");
			expect(topicLabel()?.className).toBe("mt-topic-button-label");
			expect(topicIcon()).toEqual(drawingAlone(<Icon name="sparkle" />));
		});

		test("names the topic chosen under its label, by its tag", async () => {
			asWide();
			await draw(withTopicChoice(fence, { chosen: "time.clocks" }));

			expect(topicButton().textContent).toBe("Topic: Clocks");
			expect(topicLabel()?.className).toBe("mt-topic-button-label");
			expect(topicIcon()).toEqual(drawingAlone(<Icon name="tag" />));
		});
	});

	test.each([
		["in the trial series, or from an earlier release", fence],
		[
			"whose choice does not read",
			{ ...fence, topic_choice: { chosen: 5 } } as unknown as Handed,
		],
	])("is not on a card that offers no choice: %s", async (_, payload) => {
		await draw(payload);

		expect(root.querySelector(".mt-topic-button")).toBeNull();
		expect(root.querySelector(".mt-topic-panel")).toBeNull();
		expect(root.querySelector(".mt-topic-chip")).toBeNull();
		expect(shownButtons()).toEqual(["Hint", "Another task"]);
	});

	test("is not on a card whose task is closed", async () => {
		await draw(withTopicChoice(fence), { tools: () => staleAnswer });

		press(option("B"));

		await vi.waitFor(() => expect(root.querySelector(".mt-foot")).toBeNull());
		expect(root.querySelector(".mt-topic-button")).toBeNull();
	});

	test("stays once the answer is in, after the next task", async () => {
		await draw();

		press(option("B"));

		await vi.waitFor(() =>
			expect(shownButtons()).toEqual(["Another task", "Topic: Coach"]),
		);
	});

	test("takes no press while an answer is checked", async () => {
		const answer = pending();
		await draw(withTopicChoice(fence), {
			tools: ({ name }) =>
				name === "submit_answer" ? answer.result : answered(),
		});

		press(option("B"));
		press(topicButton());

		expect(topicButton().getAttribute("aria-disabled")).toBe("true");
		expect(topicButton().getAttribute("aria-expanded")).toBe("false");
		expect(panel().hidden).toBe(true);
		answer.arrive(answered());
		await vi.waitFor(() =>
			expect(topicButton().getAttribute("aria-disabled")).toBeNull(),
		);
	});
});

describe("the panel of topics", () => {
	test("opens under the row with the focus on the choice in force, and closes with its cross, the focus back on the button", async () => {
		await draw();

		open();

		expect(topicButton().getAttribute("aria-expanded")).toBe("true");
		expect(panel().hidden).toBe(false);
		expect(panel().getAttribute("aria-labelledby")).toBe(
			panel().querySelector("h2")?.id,
		);
		expect(text(".mt-topic-title")).toBe("Which topic next?");
		expect(document.activeElement).toBe(choice("The coach chooses"));

		press(button("Close"));

		expect(panel().hidden).toBe(true);
		expect(topicButton().getAttribute("aria-expanded")).toBe("false");
		expect(document.activeElement).toBe(topicButton());
	});

	test("closes with Escape, the focus back on the button", async () => {
		await draw();
		open();

		act(() => {
			document.activeElement?.dispatchEvent(
				new KeyboardEvent("keydown", { key: "Escape", bubbles: true }),
			);
		});

		expect(panel().hidden).toBe(true);
		expect(document.activeElement).toBe(topicButton());
	});

	test("opens and closes with two presses that come before the card redraws", async () => {
		await draw();

		act(() => {
			topicButton().click();
			topicButton().click();
		});

		expect(panel().hidden).toBe(true);
		expect(topicButton().getAttribute("aria-expanded")).toBe("false");
	});

	test("holds no choices while it is closed", async () => {
		await draw();

		expect(panel().childElementCount).toBe(0);
		open();
		expect(panel().querySelectorAll(".mt-topic-option").length).toBeGreaterThan(
			17,
		);
	});

	test("closes with the button that opened it", async () => {
		await draw();
		open();

		press(topicButton());

		expect(panel().hidden).toBe(true);
		expect(topicButton().getAttribute("aria-expanded")).toBe("false");
	});

	test("offers the coach's choice, the topics the review suggests, and every topic in the groups of the site's page, in its order", async () => {
		await draw();
		open();

		expect(text(".mt-topic-coach")).toBe(
			"The coach choosesbased on the child's progress",
		);
		expect(rows()).toEqual([
			[
				"Suggested",
				["Enumeration", "Parity and alternation", "Pigeonhole principle"],
			],
			["Logic", ["Ordering", "Knights and liars", "Overlapping groups"]],
			[
				"Counting and enumeration",
				[
					"Enumeration",
					"Gaps and boundaries",
					"Pigeonhole principle",
					"Figures on a square grid",
				],
			],
			["Time", ["Clocks", "Calendar and age"]],
			[
				"Numbers with a trick",
				[
					"Arithmetic with a trick",
					"Parity and alternation",
					"Divisibility and remainders",
				],
			],
			[
				"Parts of a whole",
				["Parts and shares", "Percentages", "Ratios and sharing"],
			],
			[
				"Strategies and games",
				["Games with a winning strategy", "Weighing and pouring"],
			],
		]);
		expect(
			[...root.querySelectorAll(".mt-topic-row")].map((row) =>
				row.getAttribute("aria-label"),
			),
		).toEqual(rows().map(([name]) => name));
	});

	test.each([
		["suggests nothing", []],
		["suggests only topics the card does not know", ["logic.unknown"]],
	])(
		"has no row of suggestions where the review %s",
		async (_, recommended) => {
			await draw(withTopicChoice(fence, { recommended }));
			open();

			expect(rows()[0]?.[0]).toBe("Logic");
		},
	);

	test("sets apart a topic taught from a grade above the child's, and chooses it all the same", async () => {
		const heard = await draw();
		open();

		const older = choice("Overlapping groups");
		expect(older.dataset.older).toBe("true");
		expect(older.textContent).toBe(
			`Overlapping groups ${inEnglish.text("topic_choice.from_grade", { grade: 5 })}`,
		);
		expect(choice("Knights and liars").dataset.older).toBeUndefined();
		expect(
			choice("Knights and liars").querySelector(".mt-topic-from"),
		).toBeNull();

		press(older);

		await vi.waitFor(() =>
			expect(heard.calls).toEqual([
				{ name: "edit_profile", arguments: { lesson_topic: "logic.sets" } },
			]),
		);
	});

	test("presses in the topic the lessons are kept to wherever it is offered, and takes the focus to it", async () => {
		await draw(withTopicChoice(fence, { chosen: "combinatorics.enumeration" }));

		open();

		expect(pressedIn()).toEqual(["Enumeration", "Enumeration"]);
		expect(choice("The coach chooses").getAttribute("aria-pressed")).toBe(
			"false",
		);
		expect(document.activeElement).toBe(choice("Enumeration"));
		expect(choice("Enumeration").querySelector("svg")).not.toBeNull();
		expect(choice("Clocks").querySelector("svg")).toBeNull();
	});

	test("names each group by a link to its part of the site's page, which the chat opens", async () => {
		const heard = await draw(withTopicChoice(fence), { links: "open" });
		open();
		const links = [
			...panel().querySelectorAll<HTMLAnchorElement>(".mt-topic-row-name a"),
		];

		expect(links.map((link) => [link.textContent, link.href])).toEqual([
			["Logic", "https://mathtrail.app/en/topics/#logic"],
			["Counting and enumeration", "https://mathtrail.app/en/topics/#counting"],
			["Time", "https://mathtrail.app/en/topics/#time"],
			["Numbers with a trick", "https://mathtrail.app/en/topics/#numbers"],
			["Parts of a whole", "https://mathtrail.app/en/topics/#parts"],
			["Strategies and games", "https://mathtrail.app/en/topics/#games"],
		]);
		act(() => {
			links[2]?.click();
		});

		await vi.waitFor(() =>
			expect(heard.pages).toEqual(["https://mathtrail.app/en/topics/#time"]),
		);
		expect(heard.calls).toEqual([]);
	});

	test("shows the address of a group the chat did not open, to copy", async () => {
		await draw(withTopicChoice(fence), { links: "refuse" });
		open();

		act(() => {
			panel().querySelector<HTMLAnchorElement>(".mt-topic-row-name a")?.click();
		});

		await vi.waitFor(() =>
			expect(text(".mt-topic-panel .mt-link-address")).toBe(
				"https://mathtrail.app/en/topics/#logic",
			),
		);
	});

	test.each([
		["a chat that opens no links", withTopicChoice(fence), undefined],
		[
			"a site whose address is not one the card links to",
			withTopicChoice(fence, {
				site: { url: "http://mathtrail.app", languages: ["en"] },
			}),
			"open" as const,
		],
	])("names its groups as text on %s", async (_, payload, links) => {
		await draw(payload, { links });
		open();

		expect(panel().querySelector("a")).toBeNull();
		expect(rows()[1]?.[0]).toBe("Logic");
	});
});

describe("a topic chosen", () => {
	test("is saved, told to the model, and then asked for in the adult's words, which the card ends on", async () => {
		const heard = await draw();
		open();

		press(choice("Clocks"));

		await vi.waitFor(() =>
			expect(heard.messages).toEqual([
				inEnglish.text("topic_choice.ask", { topic: "Clocks" }),
			]),
		);
		expect(heard.messages[0]).toMatch(/^Another task.— on the topic “Clocks”$/);
		expect(heard.calls).toEqual([
			{ name: "edit_profile", arguments: { lesson_topic: "time.clocks" } },
		]);
		expect(heard.modelLines).toEqual([topicWords]);
		expect(heard.order).toEqual(["call", "model line", "message"]);
		// The new task comes below, in a card of its own.
		await expectDoneWith();
	});

	test("whose ask the chat did not take stays chosen, named on its button, and is asked for again", async () => {
		vi.spyOn(console, "error").mockImplementation(() => {});
		const heard = await draw(withTopicChoice(fence), { refuseMessages: 1 });
		open();

		press(choice("Clocks"));

		await vi.waitFor(() => expect(requestNote()).toBe(notSent));
		expect(heard.messages).toHaveLength(1);
		expect(panel().hidden).toBe(true);
		expect(topicButton().textContent).toBe("Topic: Clocks");
		expect(topicButton().getAttribute("aria-disabled")).toBeNull();
		expect(option("A").getAttribute("aria-disabled")).toBeNull();
		expect(topicNote()).toBe("");

		press(button("Another task"));

		await vi.waitFor(() => expect(heard.messages).toHaveLength(2));
		expect(heard.messages[1]).toBe("Another task");
		expect(heard.calls.map((call) => call.name)).toEqual(["edit_profile"]);
		await expectDoneWith();
	});

	test("given back to the coach asks for a task the coach chooses", async () => {
		const heard = await draw(withTopicChoice(fence, { chosen: "time.clocks" }));
		open();

		press(choice("The coach chooses"));

		await vi.waitFor(() =>
			expect(heard.messages).toEqual([
				inEnglish.text("topic_choice.ask_coach"),
			]),
		);
		expect(heard.calls).toEqual([
			{ name: "edit_profile", arguments: { lesson_topic: "" } },
		]);
		await expectDoneWith();
	});

	test("again, with nothing changed, tells the model nothing and asks for the task all the same", async () => {
		const heard = await draw(
			withTopicChoice(fence, { chosen: "time.clocks" }),
			{
				tools: () => ({
					content: [{ type: "text", text: topicWords }],
					structuredContent: {
						screen: "profile",
						changed: false,
						profile: { ...fence.child, interests: [], excluded_skills: [] },
					},
				}),
			},
		);
		open();

		press(choice("Clocks"));

		await vi.waitFor(() => expect(heard.messages).toHaveLength(1));
		expect(heard.modelLines).toEqual([]);
	});

	test("is asked for on a host that never answers the line for the model, and leaves the card free once the ask is refused", async () => {
		vi.spyOn(console, "error").mockImplementation(() => {});
		const opened = await openCard({ tools: service, refuseMessages: true });
		root = opened.root;
		opened.host.onupdatemodelcontext = () => new Promise(() => {});
		await deliver(opened.host, withTopicChoice(fence));
		await vi.waitFor(() =>
			expect(root.querySelector(".mt-option")).not.toBeNull(),
		);
		open();

		press(choice("Clocks"));

		await vi.waitFor(() => expect(opened.heard.messages).toHaveLength(1));
		// The chat did not take the ask, so the card is given back, no longer
		// saving the topic.
		await vi.waitFor(() =>
			expect(topicButton().getAttribute("aria-disabled")).toBeNull(),
		);
		expect(option("A").getAttribute("aria-disabled")).toBeNull();
		expect(topicNote()).toBe("");
	});

	test("is asked for once the host has taken the line for the model", async () => {
		const opened = await openCard({ tools: service });
		root = opened.root;
		opened.host.onupdatemodelcontext = async () => {
			await new Promise((taken) => setTimeout(taken, 100));
			opened.heard.order.push("model line");
			return {};
		};
		await deliver(opened.host, withTopicChoice(fence));
		await vi.waitFor(() =>
			expect(root.querySelector(".mt-option")).not.toBeNull(),
		);
		open();

		press(choice("Clocks"));

		await vi.waitFor(() => expect(opened.heard.messages).toHaveLength(1));
		expect(opened.heard.order).toEqual(["call", "model line", "message"]);
	});

	test("carries to the model the line of an answer no message has carried yet", async () => {
		const heard = await draw();
		press(option("B"));
		await vi.waitFor(() => expect(heard.modelLines).toHaveLength(1));
		open();

		press(choice("Clocks"));

		await vi.waitFor(() => expect(heard.messages).toHaveLength(1));
		expect(heard.modelLines[1]).toBe(`${heard.modelLines[0]}\n\n${topicWords}`);
	});

	test("carries to the model the line of an answer an ask the chat refused did not carry", async () => {
		vi.spyOn(console, "error").mockImplementation(() => {});
		const heard = await draw(withTopicChoice(fence), { refuseMessages: 1 });
		press(option("B"));
		await vi.waitFor(() => expect(heard.modelLines).toHaveLength(1));
		press(button("Another task"));
		await vi.waitFor(() => expect(requestNote()).toBe(notSent));
		open();

		press(choice("Clocks"));

		await vi.waitFor(() => expect(heard.messages).toHaveLength(2));
		expect(heard.modelLines[1]).toBe(`${heard.modelLines[0]}\n\n${topicWords}`);
	});

	test("is asked for on a host that keeps no line for the model", async () => {
		const heard = await draw(withTopicChoice(fence), {
			refuseModelLines: true,
		});
		open();

		press(choice("Clocks"));

		await vi.waitFor(() => expect(heard.messages).toHaveLength(1));
	});

	test.each<[string, () => CallToolResult | Promise<CallToolResult>, string]>([
		[
			"the service failed",
			() => failure,
			"The topic wasn't saved — try again.",
		],
		[
			"the service refused it",
			() => editRefused,
			"The topic wasn't saved — try again.",
		],
		[
			"the call never came back",
			() => Promise.reject(new Error("lost")),
			"The topic wasn't saved — try again.",
		],
		[
			"the profile is gone",
			() => editGone,
			"This profile is no longer there. Ask in the chat for a new one.",
		],
	])(
		"that was not saved because %s says so under the panel, asks nothing, and keeps the panel open",
		async (_, tools, said) => {
			const heard = await draw(withTopicChoice(fence), { tools });
			open();

			press(choice("Clocks"));

			await vi.waitFor(() =>
				expect(topicNote()?.replace(/\u00a0/g, " ")).toBe(said),
			);
			expect(heard.messages).toEqual([]);
			expect(heard.modelLines).toEqual([]);
			expect(panel().hidden).toBe(false);
			expect(topicButton().textContent).toBe("Topic: Coach");
			expect(document.activeElement).toBe(choice("Clocks"));

			press(button("Close"));

			expect(topicNote()).toBe("");
		},
	);

	test("is sent once, and locks the card while it is saved", async () => {
		const saved = pending();
		const heard = await draw(withTopicChoice(fence), {
			tools: ({ name }) =>
				name === "edit_profile" ? saved.result : answered(),
		});
		open();

		act(() => {
			choice("Clocks").click();
			choice("Percentages").click();
		});
		press(option("B"));
		press(button("Another task"));

		expect(topicNote()).toBe("Saving the topic…");
		expect(choice("Clocks").getAttribute("aria-disabled")).toBe("true");
		expect(option("A").getAttribute("aria-disabled")).toBe("true");
		expect(button("Another task").getAttribute("aria-disabled")).toBe("true");
		expect(button("Hint").getAttribute("aria-disabled")).toBe("true");
		expect(topicButton().getAttribute("aria-disabled")).toBe("true");
		await vi.waitFor(() =>
			expect(heard.calls.map((call) => call.name)).toEqual(["edit_profile"]),
		);
		saved.arrive(topicSaved("time.clocks"));

		await vi.waitFor(() => expect(heard.messages).toHaveLength(1));
		expect(heard.calls.map((call) => call.name)).toEqual(["edit_profile"]);
		expect(heard.messages[0]).toContain("Clocks");
	});

	test("is not chosen while an answer pressed in the same moment is on its way", async () => {
		const heard = await draw();
		open();

		act(() => {
			option("B").click();
			choice("Clocks").click();
		});

		await vi.waitFor(() =>
			expect(root.querySelector(".mt-verdict-line")).not.toBeNull(),
		);
		expect(heard.calls.map((call) => call.name)).toEqual(["submit_answer"]);
		expect(heard.messages).toEqual([]);
	});

	test("on its way turns away an answer and another task pressed in the same moment", async () => {
		const heard = await draw();
		open();

		act(() => {
			choice("Clocks").click();
			option("B").click();
			button("Another task").click();
		});

		await vi.waitFor(() => expect(heard.messages).toHaveLength(1));
		expect(heard.calls.map((call) => call.name)).toEqual(["edit_profile"]);
		expect(heard.messages[0]).toContain("Clocks");
		expect(root.querySelector(".mt-verdict-line")).toBeNull();
	});

	test("takes no choice while an ask for another task is on its way", async () => {
		const heard = await draw();
		open();

		press(button("Another task"));
		expect(choice("Clocks").getAttribute("aria-disabled")).toBe("true");
		press(choice("Clocks"));

		await vi.waitFor(() => expect(heard.messages).toEqual(["Another task"]));
		expect(heard.calls).toEqual([]);
	});

	test("once the answer is in, locks the next task while it is saved", async () => {
		const saved = pending();
		const heard = await draw(withTopicChoice(fence), {
			tools: ({ name }) =>
				name === "edit_profile" ? saved.result : answered(),
		});
		press(option("B"));
		await vi.waitFor(() =>
			expect(root.querySelector(".mt-verdict-line")).not.toBeNull(),
		);
		open();

		press(choice("Percentages"));
		press(button("Another task"));

		expect(button("Another task").getAttribute("aria-disabled")).toBe("true");
		saved.arrive(topicSaved("percent.basic"));
		await vi.waitFor(() => expect(heard.messages).toHaveLength(1));
		expect(heard.messages[0]).toContain("Percentages");
	});

	test("is not chosen while an ask for another task pressed in the same moment is on its way", async () => {
		const heard = await draw();
		open();

		act(() => {
			button("Another task").click();
			choice("Clocks").click();
		});

		await vi.waitFor(() => expect(heard.messages).toEqual(["Another task"]));
		expect(heard.calls).toEqual([]);
		await expectDoneWith();
	});

	test("can be chosen once the answer is in", async () => {
		const heard = await draw();
		press(option("B"));
		await vi.waitFor(() =>
			expect(root.querySelector(".mt-verdict-line")).not.toBeNull(),
		);
		open();

		press(choice("Percentages"));

		await vi.waitFor(() => expect(heard.messages).toHaveLength(1));
		expect(heard.calls.at(-1)).toEqual({
			name: "edit_profile",
			arguments: { lesson_topic: "percent.basic" },
		});
	});
});

describe("the mark of the topic chosen", () => {
	const onGaps = withTopicChoice(fence, { chosen: "counting.gaps" });

	test("stands above a task on the topic the lessons are kept to", async () => {
		await draw(onGaps);

		const mark = root.querySelector(".mt-body > .mt-topic-chip");
		expect(mark?.textContent).toBe("Topic: Gaps and boundaries");
		expect(mark?.nextElementSibling?.className).toBe("mt-task-text");
	});

	test("is not above a task the rule gave, once its own topic is chosen on it", async () => {
		vi.spyOn(console, "error").mockImplementation(() => {});
		// The chat does not take the ask, so the card is given back, and with it
		// the place the mark would stand.
		const heard = await draw(withTopicChoice(fence), { refuseMessages: true });
		open();

		press(choice("Gaps and boundaries"));

		await vi.waitFor(() => expect(heard.messages).toHaveLength(1));
		await vi.waitFor(() => expect(requestNote()).toBe(notSent));
		expect(topicButton().textContent).toBe("Topic: Gaps and boundaries");
		expect(root.querySelector(".mt-topic-chip")).toBeNull();
	});

	test("is not above a task that is closed", async () => {
		await draw(onGaps, { tools: () => staleAnswer });
		expect(root.querySelector(".mt-topic-chip")).not.toBeNull();

		press(option("B"));

		await vi.waitFor(() => expect(root.querySelector(".mt-foot")).toBeNull());
		expect(root.querySelector(".mt-topic-chip")).toBeNull();
	});

	test.each([
		["on another topic", withTopicChoice(fence, { chosen: "time.clocks" })],
		["while the coach chooses", withTopicChoice(fence)],
	])("is not above a task %s", async (_, payload) => {
		await draw(payload);

		expect(root.querySelector(".mt-topic-chip")).toBeNull();
	});

	test("gives the choice back to the coach with its cross, and asks for a task", async () => {
		const heard = await draw(onGaps);

		press(button("Give the choice of the topic back to the coach"));

		await vi.waitFor(() =>
			expect(heard.messages).toEqual([
				inEnglish.text("topic_choice.ask_coach"),
			]),
		);
		expect(heard.calls).toEqual([
			{ name: "edit_profile", arguments: { lesson_topic: "" } },
		]);
		expect(root.querySelector(".mt-topic-chip")).toBeNull();
		await expectDoneWith();
	});

	test("given back with its cross, whose ask the chat did not take, names the coach on the button", async () => {
		vi.spyOn(console, "error").mockImplementation(() => {});
		const heard = await draw(onGaps, { refuseMessages: true });

		press(button("Give the choice of the topic back to the coach"));

		await vi.waitFor(() => expect(heard.messages).toHaveLength(1));
		await vi.waitFor(() => expect(requestNote()).toBe(notSent));
		expect(root.querySelector(".mt-topic-chip")).toBeNull();
		expect(topicButton().textContent).toBe("Topic: Coach");
	});

	test("takes no press of its cross while an answer is checked", async () => {
		const answer = pending();
		const heard = await draw(onGaps, {
			tools: ({ name }) =>
				name === "submit_answer" ? answer.result : answered(),
		});
		const cross = button("Give the choice of the topic back to the coach");

		press(option("B"));
		press(cross);

		expect(cross.getAttribute("aria-disabled")).toBe("true");
		answer.arrive(answered());
		await vi.waitFor(() =>
			expect(root.querySelector(".mt-verdict-line")).not.toBeNull(),
		);
		expect(heard.calls.map((call) => call.name)).toEqual(["submit_answer"]);
		expect(heard.messages).toEqual([]);
	});

	test("takes no press of its cross while an ask for another task is on its way, and is gone once the chat has it", async () => {
		const heard = await draw(onGaps);
		const cross = button("Give the choice of the topic back to the coach");

		press(button("Another task"));
		expect(cross.getAttribute("aria-disabled")).toBe("true");
		press(cross);

		await vi.waitFor(() => expect(heard.messages).toEqual(["Another task"]));
		expect(heard.calls).toEqual([]);
		await expectDoneWith();
		expect(root.querySelector(".mt-topic-chip")).toBeNull();
	});

	test("that is not saved stays, and says so", async () => {
		const heard = await draw(onGaps, { tools: () => failure });

		press(button("Give the choice of the topic back to the coach"));

		await vi.waitFor(() =>
			expect(topicNote()?.replace(/\u00a0/g, " ")).toBe(
				"The topic wasn't saved — try again.",
			),
		);
		expect(root.querySelector(".mt-topic-chip")).not.toBeNull();
		expect(heard.messages).toEqual([]);
	});
});

describe("the choice of the topic in the card's language", () => {
	test("is offered and asked for in Russian", async () => {
		const heard = await draw(
			withTopicChoice({
				...fenceInRussian,
				child: { ...fenceInRussian.child, ui_language: "ru" },
			}),
			{ links: "open" },
		);

		expect(topicButton().textContent).toBe("Тема: Тренер");
		open();
		expect(choice("Проценты").textContent).toBe(
			`Проценты ${inRussian.text("topic_choice.from_grade", { grade: 5 })}`,
		);
		expect(
			panel().querySelector<HTMLAnchorElement>(".mt-topic-row-name a")?.href,
		).toBe("https://mathtrail.app/ru/topics/#logic");

		press(choice("Часы"));

		await vi.waitFor(() =>
			expect(heard.messages).toEqual([
				inRussian.text("topic_choice.ask", { topic: "Часы" }),
			]),
		);
		expect(heard.messages[0]).toMatch(/^Другая задача.— на тему «Часы»$/);
	});
});

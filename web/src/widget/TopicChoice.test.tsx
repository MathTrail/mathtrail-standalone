import type { CallToolResult } from "@modelcontextprotocol/client";
import { act } from "preact/test-utils";
import { afterEach, describe, expect, test, vi } from "vitest";
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

// takeFence is the card asking for the next task after the fence. The service
// of these cases takes none, so the card asks twice — tookNone — and then the
// chat, as it always did.
const takeFence = { name: "take_task", arguments: { task_id: "task_fence" } };
const tookNone = [takeFence, takeFence];
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
const requestNote = () =>
	text(".mt-foot > .mt-action-note:not(.mt-topic-note)");
const anotherComing =
	"Once the ask reaches the chat, the new task will come below, in a new card.";

// open opens the panel as a person does.
function open(): void {
	press(topicButton());
}

describe("the button that chooses the topic", () => {
	test("stands at the end of the row, saying the coach chooses, and names the panel it opens", async () => {
		await draw();

		expect(shownButtons()).toEqual([
			"Hint",
			"Another task",
			"Topic: the coach chooses",
		]);
		expect(topicButton().getAttribute("aria-expanded")).toBe("false");
		expect(topicButton().getAttribute("aria-controls")).toBe(panel().id);
		expect(panel().hidden).toBe(true);
	});

	test("names the topic the lessons are kept to", async () => {
		await draw(withTopicChoice(fence, { chosen: "time.clocks" }));

		expect(topicButton().textContent).toBe("Topic: Clocks");
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
			expect(shownButtons()).toEqual([
				"Another task",
				"Topic: the coach chooses",
			]),
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
			expect(heard.calls[0]).toEqual({
				name: "edit_profile",
				arguments: { lesson_topic: "logic.sets" },
			}),
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
	test("is saved, told to the model, and then asked for in the child's words, and the button names it", async () => {
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
			...tookNone,
		]);
		expect(heard.modelLines).toEqual([topicWords]);
		expect(heard.order).toEqual([
			"call",
			"model line",
			"call",
			"call",
			"message",
		]);
		expect(panel().hidden).toBe(true);
		expect(topicButton().textContent).toBe("Topic: Clocks");
		expect(document.activeElement).toBe(topicButton());
		await vi.waitFor(() => expect(requestNote()).toBe(anotherComing));
		expect(topicNote()).toBe("");
		// The card keeps its task: the new one comes below, in a card of its own.
		expect(text(".mt-task-text")).toBe(fence.task.question);
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
			...tookNone,
		]);
		expect(topicButton().textContent).toBe("Topic: the coach chooses");
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

	test("is asked for on a host that never answers the line for the model, and the card is free again", async () => {
		const opened = await openCard({ tools: service });
		root = opened.root;
		opened.host.onupdatemodelcontext = () => new Promise(() => {});
		await deliver(opened.host, withTopicChoice(fence));
		await vi.waitFor(() =>
			expect(root.querySelector(".mt-option")).not.toBeNull(),
		);
		open();

		press(choice("Clocks"));

		await vi.waitFor(() => expect(opened.heard.messages).toHaveLength(1));
		await vi.waitFor(() =>
			expect(topicButton().getAttribute("aria-disabled")).toBeNull(),
		);
		expect(option("A").getAttribute("aria-disabled")).toBeNull();
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
		expect(opened.heard.order).toEqual([
			"call",
			"model line",
			"call",
			"call",
			"message",
		]);
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

	test("carries no line a message has carried already", async () => {
		const heard = await draw();
		press(option("B"));
		await vi.waitFor(() => expect(heard.modelLines).toHaveLength(1));
		press(button("Another task"));
		await vi.waitFor(() => expect(heard.messages).toHaveLength(1));
		open();

		press(choice("Clocks"));

		await vi.waitFor(() => expect(heard.messages).toHaveLength(2));
		expect(heard.modelLines[1]).toBe(topicWords);
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
			expect(topicButton().textContent).toBe("Topic: the coach chooses");
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
		expect(heard.calls.map((call) => call.name)).toEqual([
			"edit_profile",
			"take_task",
			"take_task",
		]);
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
		expect(heard.calls.map((call) => call.name)).toEqual([
			"edit_profile",
			"take_task",
			"take_task",
		]);
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
		expect(heard.calls).toEqual(tookNone);
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
		expect(heard.calls).toEqual(tookNone);
		expect(topicButton().textContent).toBe("Topic: the coach chooses");
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
		expect(heard.calls.slice(-3)).toEqual([
			{ name: "edit_profile", arguments: { lesson_topic: "percent.basic" } },
			...tookNone,
		]);
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
		const heard = await draw();
		open();

		press(choice("Gaps and boundaries"));

		await vi.waitFor(() => expect(heard.messages).toHaveLength(1));
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
			...tookNone,
		]);
		expect(root.querySelector(".mt-topic-chip")).toBeNull();
		expect(topicButton().textContent).toBe("Topic: the coach chooses");
		expect(document.activeElement).toBe(topicButton());
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

	test("takes no press of its cross while an ask for another task is on its way", async () => {
		const heard = await draw(onGaps);
		const cross = button("Give the choice of the topic back to the coach");

		press(button("Another task"));
		expect(cross.getAttribute("aria-disabled")).toBe("true");
		press(cross);

		await vi.waitFor(() => expect(heard.messages).toEqual(["Another task"]));
		expect(heard.calls).toEqual(tookNone);
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

		expect(topicButton().textContent).toBe("Тема: выбирает тренер");
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

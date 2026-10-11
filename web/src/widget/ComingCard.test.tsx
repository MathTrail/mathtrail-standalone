import type { CallToolResult } from "@modelcontextprotocol/client";
import { act } from "preact/test-utils";
import { afterEach, beforeEach, describe, expect, test, vi } from "vitest";
import { firstAskIn } from "./coming";
import { buttonIn, foldIn, openCard, press, takeDown } from "./testing/card";
import type { ToolCall } from "./testing/host";
import {
	coming,
	failure,
	fence,
	notComing,
	onTheCard,
	progress,
	rightAnswer,
	writing,
} from "./testing/lesson";
import { moments } from "./waiting";

let root: HTMLElement;
let seen: "visible" | "hidden" = "visible";

beforeEach(() => {
	// The card's clock is the test's, and a page is looked at unless a case
	// puts it out of sight.
	vi.useFakeTimers({ toFake: ["setTimeout", "clearTimeout", "Date"] });
	seen = "visible";
	vi.spyOn(document, "visibilityState", "get").mockImplementation(() => seen);
});

afterEach(() => {
	takeDown(root);
	vi.restoreAllMocks();
	vi.useRealTimers();
});

// first is when the card of the request in the fixtures first asks, counted
// from the moment it is drawn.
const first = firstAskIn(coming.request_id);

// late is how far either side of its moment a question may be seen: the clock
// passes a frame at a time, and the card sets its next question going in the
// frame that brought the answer to the one before.
const late = 300;

// finished is how long a card holds its course done, every step ticked at
// once, before the task that has come takes the course's place.
const finished = moments.held;

// Answer is how a question is answered: at once, when a promise settles, or
// by a promise made as the question comes — an answer the service holds.
type Answer =
	| CallToolResult
	| Promise<CallToolResult>
	| (() => Promise<CallToolResult>);

// answering answers the card's questions with the statuses given, one for each
// question, the last again for every question after it; and the other tools a
// card calls as the service would.
function answering(...statuses: Answer[]) {
	let asked = 0;
	return (call: ToolCall) => {
		switch (call.name) {
			case "read_task": {
				const status = statuses[Math.min(asked, statuses.length - 1)];
				asked += 1;
				if (status === undefined) {
					throw new Error("no status to answer with");
				}
				return typeof status === "function" ? status() : status;
			}
			case "read_progress":
				return progress;
			case "submit_answer":
				return rightAnswer;
			default:
				throw new Error(`no ${call.name} here`);
		}
	};
}

// heldFor is an answer the service holds for milliseconds of the card's clock
// from the moment the question comes, as it holds a question for news.
function heldFor(milliseconds: number, status: CallToolResult): Answer {
	return () =>
		new Promise((resolve) => setTimeout(() => resolve(status), milliseconds));
}

// drawn is the card a task asked for comes to, drawn on a host that answers
// it as given, and the questions it asks of the service, as the host heard
// them. The card's effects have run, so that its clock starts now.
async function drawn(tools: ReturnType<typeof answering>) {
	const opened = await openCard({ context: { locale: "en-US" }, tools });
	root = opened.root;
	await opened.host.sendToolResult({ content: [], structuredContent: coming });
	await pass(100);
	return {
		...opened,
		asked: () => opened.heard.calls.filter((call) => call.name === "read_task"),
	};
}

// step is how much of the card's clock passes between two of its frames.
const step = 50;

// pass lets milliseconds of the card's own clock pass, a frame at a time, and
// whatever was set going in them come back. A frame ends with the card's
// effects run, as a page runs them, so that what an answer sets going — the
// next question — is set going within the frame.
async function pass(milliseconds: number): Promise<void> {
	let left = milliseconds;
	do {
		const now = Math.min(step, left);
		await act(async () => {
			await vi.advanceTimersByTimeAsync(now);
		});
		left -= now;
	} while (left > 0);
}

const text = (selector: string) => root.querySelector(selector)?.textContent;
const news = () => root.querySelector(".mt-news")?.textContent ?? "";
const labels = () =>
	[...root.querySelectorAll(".mt-gen li")].map((step) => step.textContent);

describe("a card a task asked for comes to", () => {
	test("asks how the task stands a moment after it is drawn, then every few seconds while it is written", async () => {
		const card = await drawn(answering(writing()));

		await pass(first);
		expect(card.asked()).toEqual([
			{ name: "read_task", arguments: { request_id: coming.request_id } },
		]);
		expect(labels()).toEqual([
			"Done: Picked topic and difficulty",
			"In progress: Writing the task",
			"Waiting: Checking every answer",
			"Waiting: Ready",
		]);
		expect(text(".mt-badge")).toBe("Olympiad coach · Grade 3");
		expect(text(".mt-bar-action")).toBe("Profile & progress");

		await pass(moments.ask - late);
		expect(card.asked()).toHaveLength(1);
		await pass(2 * late);
		expect(card.asked()).toHaveLength(2);
		await pass(3 * (moments.ask + late));
		expect(card.asked()).toHaveLength(5);
	});

	test("ticks off the checks and the task ready at once when the task is on the card, then turns into it, asks no more, and turns into how its answer went", async () => {
		const card = await drawn(answering(writing(), onTheCard));

		await pass(first + moments.ask);
		expect(labels()).toEqual([
			"Done: Picked topic and difficulty",
			"Done: Writing the task",
			"Done: Checking every answer",
			"Done: Ready",
		]);
		expect(root.querySelector(".mt-task-text")).toBeNull();

		await pass(moments.held);
		expect(text(".mt-task-text")).toBe(fence.task.question);
		expect(root.querySelector(".mt-gen")).toBeNull();
		await pass(moments.askSlowly * 4);
		expect(card.asked()).toHaveLength(2);

		press(buttonIn(root, "C 5"));
		await pass(0);
		expect(card.heard.calls.at(-1)).toEqual({
			name: "submit_answer",
			arguments: {
				task_id: fence.task.id,
				answer: "C",
				hint_used: false,
				utc_offset: expect.any(Number),
			},
		});
		expect(text(".mt-verdict-line")).toBe("Correct! It's 5.");
		expect(root.querySelector(".mt-option")).toBeNull();
	});

	test("ticks off the course for a task that comes after questions that went unanswered", async () => {
		await drawn(answering(failure, onTheCard));

		await pass(first + moments.askSlowly + first);
		expect(labels()).toEqual([
			"Done: Picked topic and difficulty",
			"Done: Writing the task",
			"Done: Checking every answer",
			"Done: Ready",
		]);

		await pass(finished);
		expect(text(".mt-task-text")).toBe(fence.task.question);
	});

	test("shows its task at once when the task is there at its first question, as on a card drawn again with an earlier chat", async () => {
		await drawn(answering(onTheCard));

		await pass(first);

		expect(text(".mt-task-text")).toBe(fence.task.question);
		expect(root.querySelector(".mt-gen")).toBeNull();
	});

	test("says a try was turned down while a new one is written", async () => {
		await drawn(answering(writing(1)));

		await pass(first);

		expect(news()).toBe(
			"A try didn't pass the checks; a new one is being written.",
		);
		expect(root.querySelector(".mt-gen")).not.toBeNull();
	});

	test("drops the word of a try turned down once the task has come", async () => {
		await drawn(answering(writing(1), onTheCard));

		await pass(first + moments.ask);

		expect(news()).toBe("");
		expect(root.querySelector(".mt-gen")).not.toBeNull();
	});

	test("says the task is taking long two minutes after its last news, and asks seldom from then on", async () => {
		// The news comes half a pace after the first question, so that the
		// wait goes long halfway between two questions at the usual pace.
		const card = await drawn(
			answering(heldFor(moments.ask / 2, writing()), writing()),
		);
		await pass(first + moments.ask / 2);

		await pass(moments.slow - late);
		expect(news()).toBe("");
		await pass(2 * late);
		expect(text(".mt-verdict-line")).toBe(
			"The task is taking longer than usual.",
		);
		expect(text(".mt-verdict-detail")).toBe(
			"If it doesn't appear here, ask for a task in the chat.",
		);
		expect(root.querySelector(".mt-gen")).not.toBeNull();

		// The question set going before the wait went long is asked at the
		// usual pace, half a pace later; each one after it waits the longer
		// pause, counted from the start of the one before, and put off by the
		// card's own moment as well.
		await pass(moments.ask / 2);
		const slowly = card.asked().length;
		await pass(moments.askSlowly + first - 2 * late);
		expect(card.asked()).toHaveLength(slowly);
		await pass(2 * late);
		expect(card.asked()).toHaveLength(slowly + 1);
		await pass(moments.askSlowly + first - 2 * late);
		expect(card.asked()).toHaveLength(slowly + 1);
		await pass(2 * late);
		expect(card.asked()).toHaveLength(slowly + 2);
	});

	test("counts the wait afresh from each try turned down", async () => {
		await drawn(answering(writing(), writing(1)));
		await pass(first + moments.ask + late);
		expect(news()).toBe(
			"A try didn't pass the checks; a new one is being written.",
		);

		await pass(moments.slow - moments.ask);
		expect(news()).not.toContain("longer than usual");
	});

	test("says no task is here once its request is over, and asks no more once that is sure", async () => {
		const card = await drawn(answering(writing(), notComing, notComing));

		await pass(first + moments.ask + late);
		expect(text(".mt-verdict-line")).toBe("No task here.");
		expect(text(".mt-verdict-detail")).toBe(
			"The next one comes below; if none does, ask for a task in the chat.",
		);
		expect(root.querySelector(".mt-gen")).toBeNull();

		await pass(moments.ask + late);
		expect(card.asked()).toHaveLength(3);
		await pass(moments.askSlowly * 4);
		expect(card.asked()).toHaveLength(3);
	});

	test("says no task is here when it is drawn again long after, and makes sure of it once more", async () => {
		const card = await drawn(answering(notComing));

		await pass(first);
		expect(text(".mt-verdict-line")).toBe("No task here.");
		expect(root.querySelector(".mt-gen")).toBeNull();

		await pass(moments.ask - 2 * late);
		expect(card.asked()).toHaveLength(1);
		await pass(2 * late);
		expect(card.asked()).toHaveLength(2);
		await pass(moments.askSlowly * 4);
		expect(card.asked()).toHaveLength(2);
	});

	test("goes back to the wait when the word that no task is coming was a read made too early", async () => {
		await drawn(answering(notComing, writing()));

		await pass(first + moments.ask + late);

		expect(root.querySelector(".mt-gen")).not.toBeNull();
		expect(root.querySelector(".mt-verdict")).toBeNull();
	});

	test("keeps the answer on its way when the wait goes long meanwhile", async () => {
		let deliver: (status: CallToolResult) => void = () => {};
		const onItsWay = new Promise<CallToolResult>((resolve) => {
			deliver = resolve;
		});
		// Every question is told the task is being written, until the one set
		// going a pause before the wait goes long, whose answer comes after.
		const before = moments.slow / moments.ask - 1;
		const card = await drawn(
			answering(
				...Array.from({ length: before }, () => writing()),
				onItsWay,
				writing(),
			),
		);
		await pass(first + moments.slow - late);
		expect(card.asked()).toHaveLength(before + 1);

		await pass(2 * late);
		expect(text(".mt-verdict-line")).toBe(
			"The task is taking longer than usual.",
		);
		deliver(onTheCard);
		await pass(late);
		expect(news()).toBe("");
		await pass(finished);

		expect(text(".mt-task-text")).toBe(fence.task.question);
		expect(card.asked()).toHaveLength(before + 1);
	});

	test("says in each question after it has heard the task is being written how many tries it has heard were turned down, so that the service may hold the question for news", async () => {
		const card = await drawn(answering(writing(), writing(1), writing(1)));

		await pass(first + 2 * moments.ask + late);

		expect(card.asked().map((call) => call.arguments)).toEqual([
			{ request_id: coming.request_id },
			{ request_id: coming.request_id, refused: 0 },
			{ request_id: coming.request_id, refused: 1 },
		]);
	});

	test("asks again at once after a question the service held for news, and after the rest of its pace after one answered sooner", async () => {
		const card = await drawn(
			answering(
				writing(),
				heldFor(moments.ask, writing()),
				heldFor(1_000, writing()),
				writing(),
			),
		);
		await pass(first);
		expect(card.asked()).toHaveLength(1);

		// The second question starts a pace after the first, and is held for
		// a pace: the third follows its answer at once.
		await pass(2 * moments.ask - late);
		expect(card.asked()).toHaveLength(2);
		await pass(2 * late);
		expect(card.asked()).toHaveLength(3);

		// The third is answered a second after it started, and the fourth
		// waits out the rest of the pace from the third's start.
		await pass(moments.ask - 3 * late);
		expect(card.asked()).toHaveLength(3);
		await pass(2 * late);
		expect(card.asked()).toHaveLength(4);
	});

	test("asks as it did before, saying nothing of what it heard, once two questions held for news in a row went unanswered", async () => {
		const card = await drawn(
			answering(writing(), failure, failure, writing(), writing()),
		);

		await pass(
			first + 2 * moments.ask + 2 * (moments.askSlowly + first) + late,
		);

		expect(card.asked().map((call) => call.arguments)).toEqual([
			{ request_id: coming.request_id },
			{ request_id: coming.request_id, refused: 0 },
			{ request_id: coming.request_id, refused: 0 },
			{ request_id: coming.request_id },
			{ request_id: coming.request_id },
		]);
	});

	test("asks one question at a time: a question unanswered holds back the next", async () => {
		const card = await drawn(answering(new Promise<CallToolResult>(() => {})));

		// Well within the minute the host's library waits for an answer before
		// it gives the question up.
		await pass(first + moments.askSlowly * 3);

		expect(card.asked()).toHaveLength(1);
	});

	test("takes a failed question for no news, and asks again seldom", async () => {
		const card = await drawn(answering(writing(), failure, writing()));
		await pass(first + moments.ask + late);
		expect(card.asked()).toHaveLength(2);
		expect(root.querySelector(".mt-gen")).not.toBeNull();

		await pass(moments.askSlowly + first - 2 * late);
		expect(card.asked()).toHaveLength(2);
		await pass(2 * late);
		expect(card.asked()).toHaveLength(3);
	});

	// A host may lose a call on its way to the service, or the answer on its
	// way back: the card has heard nothing of the task, and goes on waiting.
	test("takes a question whose answer never came back for no news, goes on showing the task being written, and asks again seldom", async () => {
		vi.spyOn(console, "error").mockImplementation(() => {});
		const card = await drawn(
			answering(() => Promise.reject(new Error("the host lost the call"))),
		);

		await pass(first + late);
		expect(card.asked()).toHaveLength(1);
		expect(console.error).toHaveBeenCalledOnce();
		expect(labels()).toEqual([
			"Done: Picked topic and difficulty",
			"In progress: Writing the task",
			"Waiting: Checking every answer",
			"Waiting: Ready",
		]);
		expect(news()).toBe("");
		expect(root.querySelector(".mt-verdict")).toBeNull();

		await pass(moments.askSlowly + first - 2 * late);
		expect(card.asked()).toHaveLength(1);
		await pass(2 * late);
		expect(card.asked()).toHaveLength(2);
	});

	test("asks nothing while the page is out of sight, and asks at once when it is looked at again", async () => {
		seen = "hidden";
		const card = await drawn(answering(writing()));

		await pass(first + moments.askSlowly * 2);
		expect(card.asked()).toHaveLength(0);

		seen = "visible";
		document.dispatchEvent(new Event("visibilitychange"));
		await pass(0);
		expect(card.asked()).toHaveLength(1);
	});

	test("asks nothing once it is taken off the page", async () => {
		const card = await drawn(answering(writing()));
		await pass(first);

		takeDown(root);
		await pass(moments.askSlowly * 4);

		expect(card.asked()).toHaveLength(1);
	});

	test("keeps the progress open over it when the task comes, and leads back to the task", async () => {
		await drawn(answering(writing(), onTheCard));
		await pass(first);

		press(buttonIn(root, "Comet Profile & progress"));
		await pass(late);
		expect(buttonIn(root, "Back")).toBeTruthy();

		await pass(moments.ask + late + finished);
		press(buttonIn(root, "Back to task"));
		await pass(0);

		expect(text(".mt-task-text")).toBe(fence.task.question);
		// One frame holds the wait and the task: one line at its top.
		expect(root.querySelectorAll(".mt-bar")).toHaveLength(1);
	});

	test("keeps a section of the progress opened over the wait open over the task that came", async () => {
		await drawn(answering(writing(), onTheCard));
		await pass(first);
		press(buttonIn(root, "Comet Profile & progress"));
		await pass(late);
		press(foldIn(root, "Profile"));

		await pass(moments.ask + late + finished);
		press(buttonIn(root, "Back to task"));
		await pass(0);
		expect(text(".mt-task-text")).toBe(fence.task.question);
		press(buttonIn(root, "Comet Profile & progress"));
		await pass(late);

		expect(foldIn(root, "Profile").getAttribute("aria-expanded")).toBe("true");
	});
});

import { afterEach, describe, expect, test, vi } from "vitest";
import { type Box, repliesFrom, shareTheCard, stepAtMiddle } from "./stage";
import { openHome } from "./testing/home";

afterEach(() => {
	document.body.innerHTML = "";
	vi.restoreAllMocks();
});

// box is a step standing from top to bottom in the window.
const box = (top: number, bottom: number): Box => ({
	top,
	bottom,
	height: bottom - top,
});

describe("the step at the middle of the window", () => {
	test.each([
		["the one the middle falls in", [box(0, 300), box(350, 600)], 1],
		["the one whose edge is nearest", [box(0, 300), box(450, 700)], 1],
		["the first of two as near", [box(0, 350), box(450, 700)], 0],
		["never one that takes no room", [box(400, 400), box(500, 700)], 1],
	])("is %s", (_, boxes, want) => {
		expect(stepAtMiddle(boxes, 800)).toBe(want);
	});

	test("is none when no step takes room", () => {
		expect(stepAtMiddle([box(0, 0)], 800)).toBeUndefined();
		expect(stepAtMiddle([], 800)).toBeUndefined();
	});
});

// windowOf is a window as tall as height, wide enough for the steps to share
// a card while wide says so, whose frames run when the test says.
function windowOf(wide: boolean, height = 800) {
	let matches = wide;
	const changes = new Set<() => void>();
	const heard = new Map<string, Set<() => void>>();
	const frames: FrameRequestCallback[] = [];
	const cancelled: number[] = [];
	const fire = (type: string) => {
		for (const listener of [...(heard.get(type) ?? [])]) {
			listener();
		}
	};
	const runFrames = () => {
		for (const run of frames.splice(0)) {
			run(0);
		}
	};
	const window = {
		innerHeight: height,
		matchMedia: () => ({
			get matches() {
				return matches;
			},
			addEventListener: (_: string, change: () => void) => changes.add(change),
			removeEventListener: (_: string, change: () => void) =>
				changes.delete(change),
		}),
		requestAnimationFrame: (run: FrameRequestCallback) => frames.push(run),
		cancelAnimationFrame: (frame: number) => cancelled.push(frame),
		addEventListener: (type: string, listener: () => void) => {
			const listeners = heard.get(type) ?? new Set();
			listeners.add(listener);
			heard.set(type, listeners);
		},
		removeEventListener: (type: string, listener: () => void) =>
			heard.get(type)?.delete(listener),
	} as unknown as Window;
	return {
		window,
		resize(toWide: boolean) {
			matches = toWide;
			for (const change of [...changes]) {
				change();
			}
		},
		scroll() {
			fire("scroll");
			runFrames();
		},
		changeSize() {
			fire("resize");
			runFrames();
		},
		scrollWithFrameDue() {
			fire("scroll");
		},
		cancelled,
	};
}

// steps are the steps of the lesson, in the order the page shows them.
const steps = () => [
	...document.querySelectorAll<HTMLElement>("#lesson .s-walk-step"),
];

// standing puts the step at in the middle of a window 800 tall, the others
// above and below it, 400 tall each and 100 apart.
function standing(at: number): void {
	steps().forEach((step, index) => {
		const top = 200 + (index - at) * 500;
		vi.spyOn(step, "getBoundingClientRect").mockReturnValue({
			top,
			bottom: top + 400,
			height: 400,
		} as DOMRect);
	});
}

// shown is the card the steps share, as shown.
const shown = () =>
	document.querySelector<HTMLElement>(
		".s-walk-stage > .s-walk-card[data-shown]",
	);

describe("the steps of the lesson on a wide window", () => {
	test("share one card beside them, in a column of its own: the card of the step at the middle", () => {
		openHome("en");
		standing(3);

		shareTheCard(document, windowOf(true).window);

		expect(document.querySelector(".s-walk-staged > .s-walk")).not.toBeNull();
		expect(
			document.querySelectorAll(
				".s-walk-staged > .s-walk-stage > .s-walk-card",
			),
		).toHaveLength(6);
		expect(document.querySelector(".s-walk-step .s-walk-card")).toBeNull();
		expect(document.querySelectorAll(".s-walk-card[data-shown]")).toHaveLength(
			1,
		);
		expect(shown()?.querySelector(".mt-note-trap")).not.toBeNull();
	});

	test("follow the step at the middle as the page scrolls", () => {
		openHome("en");
		standing(0);
		const window = windowOf(true);
		shareTheCard(document, window.window);
		expect(shown()?.querySelector(".mt-gen")).not.toBeNull();

		standing(5);
		window.scroll();

		expect(shown()?.querySelector(".mt-rank")).not.toBeNull();
		expect(document.querySelectorAll(".s-walk-card[data-shown]")).toHaveLength(
			1,
		);
	});

	test("show in each card's frame what its step speaks of: the question under a card, the replies of an answer told", () => {
		openHome("en");
		standing(0);
		const [question, wrong] = [4, 3].map(
			(at) =>
				steps()[at]?.querySelector<HTMLElement>(".s-frame-body") ?? undefined,
		);
		if (question === undefined || wrong === undefined) {
			throw new Error("the lesson has no frame for its answer or question");
		}
		Object.defineProperty(question, "scrollHeight", { value: 900 });
		vi.spyOn(wrong, "getBoundingClientRect").mockReturnValue({
			top: 100,
		} as DOMRect);
		const replies = wrong.querySelector<HTMLElement>(".mt-replies .mt-reply");
		if (replies === null) {
			throw new Error("the answer's card has no replies");
		}
		vi.spyOn(replies, "getBoundingClientRect").mockReturnValue({
			top: 700,
		} as DOMRect);

		shareTheCard(document, windowOf(true).window);

		expect(question.scrollTop).toBe(900);
		expect(wrong.scrollTop).toBe(700 - 100 - repliesFrom);
	});
});

describe("the steps of the lesson on a narrow window", () => {
	test("keep a card each, as the page was built", () => {
		openHome("en");
		const built = document.querySelector("#lesson")?.innerHTML;

		shareTheCard(document, windowOf(false).window);

		expect(document.querySelector(".s-walk-staged")).toBeNull();
		expect(document.querySelector("#lesson")?.innerHTML).toBe(built);
	});

	test("get their cards back as the window narrows, the page as it was built", () => {
		openHome("en");
		standing(2);
		const built = document.querySelector("#lesson")?.innerHTML;
		const window = windowOf(true);
		shareTheCard(document, window.window);

		window.resize(false);

		expect(document.querySelector(".s-walk-staged")).toBeNull();
		expect(document.querySelector("#lesson")?.innerHTML).toBe(built);

		window.resize(true);
		expect(
			document.querySelectorAll(".s-walk-stage > .s-walk-card"),
		).toHaveLength(6);
	});
});

test("the shared card, stopped, leaves the page as it was built and follows nothing", () => {
	openHome("en");
	standing(1);
	const built = document.querySelector("#lesson")?.innerHTML;
	const window = windowOf(true);
	const stop = shareTheCard(document, window.window);

	stop();
	window.scroll();
	window.resize(true);

	expect(document.querySelector("#lesson")?.innerHTML).toBe(built);
});

describe("the shared card", () => {
	test("is lined up again in its frame and follows the step at the middle as the window changes size", () => {
		openHome("en");
		standing(0);
		const window = windowOf(true);
		shareTheCard(document, window.window);
		const question = document.querySelectorAll<HTMLElement>(
			".s-walk-stage .s-frame-body",
		)[4];
		if (question === undefined) {
			throw new Error("the stage has no frame for the question");
		}
		Object.defineProperty(question, "scrollHeight", { value: 640 });

		standing(4);
		window.changeSize();

		expect(question.scrollTop).toBe(640);
		expect(shown()?.querySelector(".s-chat")).not.toBeNull();
	});

	test("stopped while a frame is due, lets the frame go", () => {
		openHome("en");
		standing(0);
		const window = windowOf(true);
		const stop = shareTheCard(document, window.window);
		window.scrollWithFrameDue();

		stop();

		expect(window.cancelled).toEqual([1]);
	});

	test("leaves a page with no lesson as it is", () => {
		document.body.innerHTML = "<main><p>No lesson here.</p></main>";

		shareTheCard(document, windowOf(true).window)();

		expect(document.body.innerHTML).toBe("<main><p>No lesson here.</p></main>");
	});
});

import { afterEach, describe, expect, test, vi } from "vitest";
import { type Box, shareTheCard, stepAtMiddle } from "./stage";
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
		changeSizeWithFrameDue() {
			fire("resize");
		},
		runFrames,
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

// questionFrame is the frame of the chat the question under the answer's card
// stands in, once the steps share one card.
function questionFrame(): HTMLElement {
	const frame = document.querySelectorAll<HTMLElement>(
		".s-walk-stage .s-frame-body",
	)[4];
	if (frame === undefined) {
		throw new Error("the stage has no frame for the question");
	}
	return frame;
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
		// The steps' words say what the cards show: the column that repeats them
		// is no second reading for a screen reader.
		expect(
			document.querySelector(".s-walk-stage")?.getAttribute("aria-hidden"),
		).toBe("true");
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

	test("show in each card's frame what its step speaks of: the question under a card, the top of the card of how an answer went", () => {
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
		wrong.scrollTop = 300;

		shareTheCard(document, windowOf(true).window);

		expect(question.scrollTop).toBe(900);
		expect(wrong.scrollTop).toBe(0);
		expect(wrong.querySelector(".mt-verdict-line")).not.toBeNull();
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

	// The window narrows between a scroll and the frame it asked for: by the
	// frame, there is no shared card left to show.
	test("get their cards back as the window narrows, and keep them through a frame that was due", () => {
		openHome("en");
		standing(2);
		const built = document.querySelector("#lesson")?.innerHTML;
		const window = windowOf(true);
		shareTheCard(document, window.window);
		window.scrollWithFrameDue();
		window.changeSizeWithFrameDue();

		window.resize(false);
		window.runFrames();

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

	test("is lined up in the window's next frame, once however often the window changes size before it", () => {
		openHome("en");
		standing(4);
		const window = windowOf(true);
		shareTheCard(document, window.window);
		const question = questionFrame();
		Object.defineProperty(question, "scrollHeight", {
			value: 640,
			configurable: true,
		});

		window.changeSizeWithFrameDue();
		window.changeSizeWithFrameDue();
		expect(question.scrollTop).toBe(0);
		window.runFrames();

		expect(question.scrollTop).toBe(640);
	});

	test("is lined up again, and the step at the middle found again, once the page's own typeface has come", async () => {
		openHome("en");
		standing(0);
		let typeface = () => {};
		Object.defineProperty(document, "fonts", {
			value: {
				ready: new Promise<void>((resolve) => {
					typeface = resolve;
				}),
			},
			configurable: true,
		});
		const window = windowOf(true);
		shareTheCard(document, window.window);
		const question = questionFrame();
		Object.defineProperty(question, "scrollHeight", {
			value: 720,
			configurable: true,
		});
		standing(4);

		typeface();
		await Promise.resolve();
		await Promise.resolve();
		window.runFrames();

		expect(question.scrollTop).toBe(720);
		expect(shown()?.querySelector(".s-chat")).not.toBeNull();
		Reflect.deleteProperty(document, "fonts");
	});

	test("shows a step's own card when a step before it has none", () => {
		openHome("en");
		steps()[2]?.querySelector(".s-walk-card")?.remove();
		standing(3);

		shareTheCard(document, windowOf(true).window);

		// The answer told, and not the question asked under it, which tells
		// the same answer.
		expect(shown()?.querySelector(".mt-note-trap")).not.toBeNull();
		expect(shown()?.querySelector(".s-chat")).toBeNull();
		expect(
			document.querySelectorAll(".s-walk-stage > .s-walk-card"),
		).toHaveLength(5);
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

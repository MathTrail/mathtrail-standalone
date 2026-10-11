import { type ComponentChildren, render } from "preact";
import { act } from "preact/test-utils";
import { afterEach, beforeEach, describe, expect, test, vi } from "vitest";
import { CardIsTheDocument } from "./focus";
import {
	bringsBack,
	heldFor,
	type Observed,
	type TopSight,
	timeLeft,
	topSightOf,
	useTopHeldInSight,
} from "./sight";

// Watched is a stand-in for the browser's intersection observer, which the
// test page does not lay out for: the test says where the band along the
// card's top stands in the chat's sight, and whether any of the card is.
class Watched {
	static last: Watched | undefined;
	targets: Element[] = [];
	readonly tell: IntersectionObserverCallback;
	readonly options: IntersectionObserverInit | undefined;
	constructor(
		tell: IntersectionObserverCallback,
		options?: IntersectionObserverInit,
	) {
		this.tell = tell;
		this.options = options;
		Watched.last = this;
	}
	observe(target: Element) {
		this.targets.push(target);
	}
	unobserve() {}
	disconnect() {
		this.targets = [];
	}
	takeRecords() {
		return [];
	}
}

// The band's box, and the part of it in sight where it stands so, as a frame
// that is the whole page measures them: the edge above cuts its top ten
// pixels, the edge below its foot as many.
const box = new DOMRect(0, 0, 320, 24);
const inSight: Record<TopSight, DOMRect> = {
	in: new DOMRect(0, 0, 320, 24),
	above: new DOMRect(0, 10, 320, 14),
	below: new DOMRect(0, 0, 320, 14),
	out: new DOMRect(0, 0, 0, 0),
};

// observed is an observation of the band standing where top says.
function observed(top: TopSight): Observed {
	return {
		isIntersecting: top !== "out",
		intersectionRatio: inSight[top].height / box.height,
		boundingClientRect: box,
		intersectionRect: inSight[top],
	};
}

const band = () => document.querySelector(".mt-top-band");

// see tells the card where its top and the card stand, as the browser tells
// whatever still watches them: a hold that has ended hears nothing.
function see(top: TopSight, card = true) {
	const watched = Watched.last;
	const watching = band();
	if (watched === undefined || watching === null) {
		return;
	}
	if (!watched.targets.includes(watching)) {
		return;
	}
	const entries = [
		{ target: watching, ...observed(top) },
		{ target: document.documentElement, isIntersecting: card },
	] as unknown as IntersectionObserverEntry[];
	act(() => watched.tell(entries, watched as unknown as IntersectionObserver));
}

function Card({ since }: { since: number | undefined }) {
	useTopHeldInSight(since);
	return <p>A task</p>;
}

const root = document.createElement("div");
document.body.append(root);

// draw draws a card whose task became new at since, inside around.
function draw(
	since: number | undefined,
	around: (card: ComponentChildren) => ComponentChildren = (card) => card,
) {
	act(() => render(around(<Card since={since} />), root));
}

let scrolled: ReturnType<typeof vi.fn>;
const scrollIntoView = Element.prototype.scrollIntoView;

beforeEach(() => {
	Watched.last = undefined;
	vi.stubGlobal("IntersectionObserver", Watched);
	scrolled = vi.fn();
	Element.prototype.scrollIntoView = scrolled as never;
});

afterEach(() => {
	act(() => render(null, root));
	vi.unstubAllGlobals();
	vi.useRealTimers();
	Element.prototype.scrollIntoView = scrollIntoView;
});

// press presses key, on target or else on the page.
function press(key: string, target: EventTarget = document) {
	act(() => {
		target.dispatchEvent(new KeyboardEvent("keydown", { key, bubbles: true }));
	});
}

describe("a card whose task was handed out a moment ago", () => {
	test("brings its top back once the chat moves it up out of sight", () => {
		draw(Date.now());
		see("in");
		const top = band();

		see("above");

		expect(scrolled).toHaveBeenCalledTimes(1);
		expect(scrolled).toHaveBeenCalledWith({
			block: "start",
			behavior: "instant",
		});
		expect(scrolled.mock.contexts[0]).toBe(top);
	});

	test("brings it back when the top is gone while the rest of the card is still in sight", () => {
		draw(Date.now());
		see("in");

		see("out", true);

		expect(scrolled).toHaveBeenCalledTimes(1);
	});

	// The chat moved once has been told, and turns no more pages for that
	// reply: a second move is the reader's, even over a new task on the card.
	test("brings it back once, and never again", () => {
		draw(Date.now());
		see("in");
		see("above");

		draw(Date.now() + 1);
		see("in");
		see("above");

		expect(scrolled).toHaveBeenCalledTimes(1);
		expect(band()).toBeNull();
	});

	test("leaves a top it has never seen whole where the chat keeps it", () => {
		draw(Date.now());

		see("above");
		see("out", true);

		expect(scrolled).not.toHaveBeenCalled();
	});

	test("leaves a card wholly out of sight where the chat keeps it", () => {
		draw(Date.now());
		see("in");

		see("out", false);

		expect(scrolled).not.toHaveBeenCalled();
	});

	// A top cut by the edge below is a card coming into sight, or leaving it
	// as the reader scrolls back up the chat.
	test("leaves a top cut by the edge below alone, and still holds it", () => {
		draw(Date.now());
		see("in");

		see("below");
		expect(scrolled).not.toHaveBeenCalled();

		see("above");
		expect(scrolled).toHaveBeenCalledTimes(1);
	});

	test("stops holding at the reader's first scroll over the card", () => {
		draw(Date.now());
		see("in");

		act(() => {
			document.dispatchEvent(new Event("wheel"));
		});
		see("above");

		expect(scrolled).not.toHaveBeenCalled();
		expect(band()).toBeNull();
	});

	// A reader who has taken the chat over has it for the rest of the card's
	// life, the task that comes to a card that waited for it included. A touch
	// the browser takes over to scroll with ends for the card.
	test("stops holding for good at the reader's first scroll, whatever new it shows after", () => {
		draw(Date.now());
		see("in");
		act(() => {
			document.dispatchEvent(new Event("pointercancel"));
		});

		draw(Date.now() + 1);

		expect(band()).toBeNull();
	});

	test.each(["PageDown", "PageUp", "End", "Home", "ArrowDown", "ArrowUp"])(
		"stops holding for good at %s, which scrolls the chat from the card",
		(key) => {
			draw(Date.now());
			see("in");

			press(key);
			see("above");

			expect(scrolled).not.toHaveBeenCalled();
			expect(band()).toBeNull();
		},
	);

	test("stops holding at Space, which scrolls the chat where no control takes it", () => {
		draw(Date.now());

		press(" ");

		expect(band()).toBeNull();
	});

	// Space presses the option the focus is on; Enter presses it too; a finger
	// drifts a little as it taps one. None of them scrolls the chat.
	test("holds on through an option pressed by key or by touch", () => {
		const option = document.createElement("button");
		document.body.append(option);
		draw(Date.now());
		see("in");

		press(" ", option);
		press("Enter", option);
		act(() => {
			option.dispatchEvent(new Event("touchmove", { bubbles: true }));
		});
		see("above");

		expect(scrolled).toHaveBeenCalledTimes(1);
		option.remove();
	});

	// The card a task is on its way to watches its top while the task is
	// written, and on for the task that comes: what it saw of its top holds.
	test("keeps what it has seen when its moment moves on", () => {
		draw(Date.now());
		see("in");

		draw(Date.now() + 1);
		see("above");

		expect(scrolled).toHaveBeenCalledTimes(1);
	});

	test("holds on for the whole time from a later moment", () => {
		vi.useFakeTimers();
		draw(Date.now());
		act(() => {
			vi.advanceTimersByTime(heldFor / 2);
		});

		draw(Date.now());
		act(() => {
			vi.advanceTimersByTime(heldFor / 2 + 1);
		});
		expect(band()).not.toBeNull();

		act(() => {
			vi.advanceTimersByTime(heldFor / 2);
		});
		expect(band()).toBeNull();
	});

	// A band whose sight starts a fraction short of whole is still told as the
	// edge above cuts into it, and told again at each quarter of it.
	test("is told where its band stands just short of whole, and at each quarter of it", () => {
		draw(Date.now());

		const told = [Watched.last?.options?.threshold ?? []].flat();
		expect(told).toEqual(expect.arrayContaining([0, 0.25, 0.5, 0.75, 1]));
		expect(told.some((ratio) => ratio >= 0.99 && ratio < 1)).toBe(true);
	});

	test("stops holding once its time is up", () => {
		vi.useFakeTimers();
		draw(Date.now());
		see("in");

		act(() => {
			vi.advanceTimersByTime(heldFor);
		});
		see("above");

		expect(scrolled).not.toHaveBeenCalled();
		expect(band()).toBeNull();
	});

	test("holds only what is left of its time", () => {
		vi.useFakeTimers();
		draw(Date.now() - heldFor / 2);
		see("in");

		act(() => {
			vi.advanceTimersByTime(heldFor / 2);
		});

		expect(band()).toBeNull();
	});

	test("takes its band away with it", () => {
		draw(Date.now());
		expect(band()).not.toBeNull();

		act(() => render(null, root));

		expect(band()).toBeNull();
	});
});

describe("a card holds nothing", () => {
	test("when its task was handed out long ago, as in a chat drawn again", () => {
		draw(Date.now() - heldFor);

		expect(band()).toBeNull();
		expect(Watched.last).toBeUndefined();
	});

	test("when it does not know when", () => {
		draw(undefined);

		expect(band()).toBeNull();
	});

	test("when it is a part of a page", () => {
		draw(Date.now(), (card) => (
			<CardIsTheDocument.Provider value={false}>
				{card}
			</CardIsTheDocument.Provider>
		));

		expect(band()).toBeNull();
	});

	test("where nothing can watch it", () => {
		vi.unstubAllGlobals();
		vi.stubGlobal("IntersectionObserver", undefined);

		draw(Date.now());

		expect(band()).toBeNull();
	});
});

describe("the time a hold has left", () => {
	const now = Date.UTC(2026, 9, 11, 12);

	test.each([
		["begun now", now, heldFor],
		["begun a minute ago", now - 60_000, heldFor - 60_000],
		// The device's clock may run behind the service's: a moment a little
		// ahead is the start of a whole hold.
		["begun a little ahead of the device's clock", now + 30_000, heldFor],
		["begun just inside its time", now - heldFor + 1, 1],
	])("%s", (_, since, left) => {
		expect(timeLeft(since, now)).toBe(left);
	});

	test.each([
		["begun its whole time ago", now - heldFor],
		["begun a day ago", now - 86_400_000],
		["begun further ahead than any two clocks part", now + heldFor + 1],
		["begun at no moment", Number.NaN],
	])("is none when %s", (_, since) => {
		expect(timeLeft(since, now)).toBeUndefined();
	});
});

describe("where a card's top stands", () => {
	test.each([
		["in", "whole in sight"],
		["above", "cut by the edge above"],
		["below", "cut by the edge below"],
		["out", "out of sight"],
	] as const)("is %s when the band is %s", (top, _) => {
		expect(topSightOf(observed(top))).toBe(top);
	});

	// The edge above is seen from the first fraction of a pixel it cuts, and
	// the edge below may cut a fraction of a band that is in sight all the same.
	test("is above from the first fraction of a pixel the edge above cuts", () => {
		expect(
			topSightOf({
				isIntersecting: true,
				intersectionRatio: 0.99,
				boundingClientRect: box,
				intersectionRect: new DOMRect(0, 0.24, 320, 23.76),
			}),
		).toBe("above");
	});

	test("is in sight when the edge below cuts a fraction of it", () => {
		expect(
			topSightOf({
				isIntersecting: true,
				intersectionRatio: 0.995,
				boundingClientRect: box,
				intersectionRect: new DOMRect(0, 0, 320, 23.88),
			}),
		).toBe("in");
	});
});

describe("a card holding its top brings it back", () => {
	test.each([
		[{ seen: true, top: "above", card: true }, true],
		[{ seen: true, top: "out", card: true }, true],
		[{ seen: true, top: "out", card: false }, false],
		[{ seen: true, top: "in", card: true }, false],
		[{ seen: true, top: "below", card: true }, false],
		[{ seen: false, top: "above", card: true }, false],
		[{ seen: false, top: "out", card: true }, false],
	] as const)("given %o: %s", (hold, brings) => {
		expect(bringsBack(hold)).toBe(brings);
	});
});

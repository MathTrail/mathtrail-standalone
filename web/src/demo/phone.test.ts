import { readFileSync } from "node:fs";
import { join } from "node:path";
import { afterEach, describe, expect, test, vi } from "vitest";
import { chatAt, heldFrom, holdTheFirstScreen, idleFor } from "./phone";
import { openHome } from "./testing/pages";

afterEach(() => {
	document.body.innerHTML = "";
	vi.restoreAllMocks();
});

describe("the idle scroll after the chat's end", () => {
	test.each([
		["half the window", 900, 450],
		["never less than 320 pixels", 500, 320],
	])("is %s", (_, height, want) => {
		expect(idleFor(height)).toBe(want);
	});
});

describe("the chat in the phone", () => {
	test.each([
		["as far as the page has scrolled", 150, 400, 150],
		["at its end once the page has scrolled past it", 600, 400, 400],
		["at its top before the page has scrolled", -20, 400, 0],
		["at its top when it fits its screen", 150, 0, 0],
	])("stands %s", (_, held, overflow, want) => {
		expect(chatAt(held, overflow)).toBe(want);
	});
});

test("the stylesheet holds the first screen by the query the script asks", () => {
	const sheet = readFileSync(
		join(import.meta.dirname, "..", "site", "style.css"),
		"utf8",
	);

	expect(sheet).toContain(`@media ${heldFrom} {`);
});

// Layout is how the home page stands in a window: how far the page has
// scrolled, how tall the first screen is, the room the stylesheet gives it
// held, and how tall the chat's screen and what the chat holds are.
type Layout = {
	scrolled: number;
	hero: number;
	room: number;
	screen: number;
	content: number;
};

// the is the one element selector finds on the page.
function the(selector: string): HTMLElement {
	const element = document.querySelector<HTMLElement>(selector);
	if (element === null) {
		throw new Error(`the home page has nothing ${selector} finds`);
	}
	return element;
}

// lay sets the home page out as layout says, as a browser would: the track
// at the top of the page, under the menu, 64 pixels down, as tall as the room
// the script gives it once the first screen is held, and otherwise as tall
// as the first screen; the chat scrolls no further than what it holds runs
// past its screen.
function lay(layout: Layout): void {
	const track = the(".s-hero-track");
	const hero = the(".s-hero");
	const chat = the(".s-phone .s-frame-body");
	const end = () => Math.max(0, layout.content - layout.screen);
	let top = 0;
	Object.defineProperties(hero, {
		offsetHeight: { get: () => layout.hero, configurable: true },
	});
	Object.defineProperties(track, {
		offsetHeight: {
			get: () => {
				const room = Number.parseFloat(
					track.style.getPropertyValue("--s-hero-room"),
				);
				return track.dataset.pinned === undefined || Number.isNaN(room)
					? layout.hero
					: room;
			},
			configurable: true,
		},
	});
	Object.defineProperties(chat, {
		scrollHeight: { get: () => layout.content, configurable: true },
		clientHeight: { get: () => layout.screen, configurable: true },
		scrollTop: {
			get: () => Math.min(top, end()),
			set: (to: number) => {
				top = Math.min(Math.max(0, to), end());
			},
			configurable: true,
		},
	});
	vi.spyOn(track, "getBoundingClientRect").mockImplementation(
		() => ({ top: 64 - layout.scrolled }) as DOMRect,
	);
}

// windowOf is a window over layout, as heldFrom asks while wide says so: its
// stylesheet holds the first screen 64 pixels down, under the menu, in the
// room layout says, the page scrolls by layout, and its frames run and the
// sizes it watches change when the test says; a frame cancelled before then
// never runs. A window that refuses listeners throws at the first one asked
// of its media query.
function windowOf(layout: Layout, wide = true, refuses = false) {
	let matches = wide;
	const queries: string[] = [];
	const changes = new Set<() => void>();
	const heard = new Map<string, Set<() => void>>();
	const frames = new Map<number, FrameRequestCallback>();
	let framesAsked = 0;
	const watchers = new Set<() => void>();
	const watched: Element[] = [];
	const fire = (type: string) => {
		for (const listener of [...(heard.get(type) ?? [])]) {
			listener();
		}
	};
	const runFrames = () => {
		const due = [...frames.values()];
		frames.clear();
		for (const run of due) {
			run(0);
		}
	};
	class Sizes {
		readonly #changed: () => void;
		constructor(changed: () => void) {
			this.#changed = changed;
		}
		observe(element: Element) {
			watched.push(element);
			watchers.add(this.#changed);
		}
		disconnect() {
			watchers.delete(this.#changed);
		}
	}
	const window = {
		get innerHeight() {
			return 900;
		},
		matchMedia: (query: string) => {
			queries.push(query);
			return {
				get matches() {
					return matches;
				},
				addEventListener: (_: string, change: () => void) => {
					if (refuses) {
						throw new Error("this window's media query takes no listener");
					}
					changes.add(change);
				},
				removeEventListener: (_: string, change: () => void) =>
					changes.delete(change),
			};
		},
		getComputedStyle: () => ({ top: "64px", minHeight: `${layout.room}px` }),
		scrollBy: ({ top = 0 }: ScrollToOptions) => {
			layout.scrolled += top;
		},
		requestAnimationFrame: (run: FrameRequestCallback) => {
			framesAsked += 1;
			frames.set(framesAsked, run);
			return framesAsked;
		},
		cancelAnimationFrame: (frame: number) => frames.delete(frame),
		addEventListener: (type: string, listener: () => void) => {
			const listeners = heard.get(type) ?? new Set();
			listeners.add(listener);
			heard.set(type, listeners);
		},
		removeEventListener: (type: string, listener: () => void) =>
			heard.get(type)?.delete(listener),
		ResizeObserver: Sizes,
	} as unknown as Window & typeof globalThis;
	return {
		window,
		queries,
		watched,
		scroll(to: number) {
			layout.scrolled = to;
			fire("scroll");
			runFrames();
		},
		scrollWithFrameDue(to: number) {
			layout.scrolled = to;
			fire("scroll");
		},
		framesDue: () => frames.size,
		resize(toWide: boolean) {
			matches = toWide;
			for (const change of [...changes]) {
				change();
			}
			fire("resize");
		},
		sizesChange() {
			for (const changed of [...watchers]) {
				changed();
			}
		},
		runFrames,
	};
}

// tall is a window 900 pixels tall over the first screen held: 836 tall, as
// the room under the menu is, beside a chat that runs 199 pixels past its
// screen, the page at its top.
const tall = (): Layout => ({
	scrolled: 0,
	hero: 836,
	room: 836,
	screen: 561,
	content: 760,
});

const track = () => the(".s-hero-track");
const chat = () => the(".s-phone .s-frame-body");
const room = () => track().style.getPropertyValue("--s-hero-room");

describe("the first screen on a wide window", () => {
	test("is held, the page given room to scroll through the chat and then on, idle, for half the window", () => {
		openHome("en");
		const layout = tall();
		lay(layout);
		const window = windowOf(layout);

		holdTheFirstScreen(document, window.window);

		expect(window.queries).toEqual([heldFrom]);
		expect(track().dataset.pinned).toBe("");
		expect(room()).toBe(`${836 + 199 + 450}px`);
	});

	test("moves the chat as far as the page has scrolled, and no further than its end", () => {
		openHome("en");
		const layout = tall();
		lay(layout);
		const window = windowOf(layout);
		holdTheFirstScreen(document, window.window);

		window.scroll(150);
		expect(chat().scrollTop).toBe(150);

		window.scroll(400);
		expect(chat().scrollTop).toBe(199);
	});

	test("keeps the chat where it stands when an answer makes it longer, the page brought back to it", () => {
		openHome("en");
		const layout = tall();
		lay(layout);
		const window = windowOf(layout);
		holdTheFirstScreen(document, window.window);
		window.scroll(250);

		layout.content = 760 + 465;
		window.sizesChange();
		window.runFrames();

		expect(chat().scrollTop).toBe(199);
		expect(layout.scrolled).toBe(199);
		expect(room()).toBe(`${836 + 664 + 450}px`);
	});

	test("brings the page back to the chat's top when another task leaves it short, the first screen still held", () => {
		openHome("en");
		const layout = tall();
		layout.content = 1225;
		lay(layout);
		const window = windowOf(layout);
		holdTheFirstScreen(document, window.window);
		window.scroll(900);
		expect(chat().scrollTop).toBe(664);

		layout.content = 400;
		window.sizesChange();
		window.runFrames();

		expect(layout.scrolled).toBe(0);
		expect(chat().scrollTop).toBe(0);
		expect(room()).toBe(`${836 + 0 + 450}px`);
	});

	test("leaves the page where it is when the chat changes size with the first screen gone on up the page", () => {
		openHome("en");
		const layout = tall();
		lay(layout);
		const window = windowOf(layout);
		holdTheFirstScreen(document, window.window);
		window.scroll(3000);

		layout.content = 860;
		window.sizesChange();
		window.runFrames();

		expect(layout.scrolled).toBe(3000);
		expect(chat().scrollTop).toBe(299);
	});

	test("brings the page back to the chat's top when another task cuts the chat short with the first screen partly gone up the page, the cut chat's scroll heard first", () => {
		openHome("en");
		const layout = tall();
		layout.content = 1225;
		lay(layout);
		const window = windowOf(layout);
		holdTheFirstScreen(document, window.window);
		window.scroll(1300);
		expect(chat().scrollTop).toBe(664);

		layout.content = 400;
		chat().dispatchEvent(new Event("scroll"));
		window.runFrames();
		window.sizesChange();
		window.runFrames();

		expect(layout.scrolled).toBe(0);
		expect(chat().scrollTop).toBe(0);
	});

	test("leaves the page where it is when another task cuts the chat short with the first screen gone on up the page, the cut chat's scroll heard first", () => {
		openHome("en");
		const layout = tall();
		layout.content = 1225;
		lay(layout);
		const window = windowOf(layout);
		holdTheFirstScreen(document, window.window);
		window.scroll(3000);
		expect(chat().scrollTop).toBe(664);

		layout.content = 400;
		chat().dispatchEvent(new Event("scroll"));
		window.runFrames();
		window.sizesChange();
		window.runFrames();

		expect(layout.scrolled).toBe(3000);
		expect(chat().scrollTop).toBe(0);
	});

	test("scrolls the page to the chat when something else scrolls the chat, as the focus on a button below its edge does", () => {
		openHome("en");
		const layout = tall();
		lay(layout);
		const window = windowOf(layout);
		holdTheFirstScreen(document, window.window);

		chat().scrollTop = 120;
		chat().dispatchEvent(new Event("scroll"));
		window.runFrames();

		expect(layout.scrolled).toBe(120);
		expect(chat().scrollTop).toBe(120);
	});

	test("brings the page back from far below the first screen when the focus scrolls the chat", () => {
		openHome("en");
		const layout = tall();
		lay(layout);
		const window = windowOf(layout);
		holdTheFirstScreen(document, window.window);
		window.scroll(3000);
		expect(chat().scrollTop).toBe(199);

		chat().scrollTop = 50;
		chat().dispatchEvent(new Event("scroll"));
		window.runFrames();

		expect(layout.scrolled).toBe(50);
		expect(chat().scrollTop).toBe(50);
	});

	test("watches the first screen and what the chat holds for a change of size", () => {
		openHome("en");
		const layout = tall();
		lay(layout);
		const window = windowOf(layout);

		holdTheFirstScreen(document, window.window);

		expect(window.watched).toEqual([the(".s-hero"), ...chat().children]);
		expect(window.watched).toContain(the(".s-phone .s-card"));
	});

	test("is measured again in the window's next frame, once however many sizes change before it", () => {
		openHome("en");
		const layout = tall();
		lay(layout);
		const window = windowOf(layout);
		holdTheFirstScreen(document, window.window);
		const measured = vi.spyOn(track().style, "setProperty");

		layout.content = 860;
		window.sizesChange();
		window.resize(true);
		expect(measured).not.toHaveBeenCalled();

		window.runFrames();
		expect(measured).toHaveBeenCalledOnce();
		expect(room()).toBe(`${836 + 299 + 450}px`);
	});

	// A page scrolled by a wheel or a finger says so many times in a frame,
	// and the chat is set once for the frame the window draws.
	test("follows the page's scroll in the window's next frame, once however often the page scrolls before it", () => {
		openHome("en");
		const layout = tall();
		lay(layout);
		const window = windowOf(layout);
		holdTheFirstScreen(document, window.window);

		window.scrollWithFrameDue(100);
		window.scrollWithFrameDue(150);
		expect(window.framesDue()).toBe(1);
		expect(chat().scrollTop).toBe(0);

		window.runFrames();
		expect(chat().scrollTop).toBe(150);
	});

	test("is not held where, held, it would not stand whole under the menu, and the chat scrolls by itself", () => {
		openHome("en");
		const layout = tall();
		layout.hero = 900;
		lay(layout);
		const window = windowOf(layout);

		holdTheFirstScreen(document, window.window);
		chat().scrollTop = 80;
		window.scroll(300);

		expect(track().dataset.pinned).toBeUndefined();
		expect(room()).toBe("");
		expect(chat().scrollTop).toBe(80);
	});
});

describe("the first screen on a narrower window", () => {
	test("is not held, and the chat scrolls by itself", () => {
		openHome("en");
		const layout = tall();
		lay(layout);
		const window = windowOf(layout, false);

		holdTheFirstScreen(document, window.window);
		chat().scrollTop = 80;
		window.scroll(300);

		expect(track().dataset.pinned).toBeUndefined();
		expect(room()).toBe("");
		expect(chat().scrollTop).toBe(80);
		expect(layout.scrolled).toBe(300);
	});

	test("is let go as the window narrows, and held again as it widens", () => {
		openHome("en");
		const layout = tall();
		lay(layout);
		const window = windowOf(layout);
		holdTheFirstScreen(document, window.window);

		window.resize(false);
		window.runFrames();
		expect(track().dataset.pinned).toBeUndefined();
		expect(room()).toBe("");

		window.resize(true);
		window.runFrames();
		expect(track().dataset.pinned).toBe("");
		expect(room()).toBe(`${836 + 199 + 450}px`);
	});
});

test("the first screen, let go, is as the page was built and follows nothing", () => {
	openHome("en");
	const layout = tall();
	lay(layout);
	const window = windowOf(layout);
	const stop = holdTheFirstScreen(document, window.window);

	stop();
	window.scroll(150);
	window.sizesChange();
	window.runFrames();

	expect(track().dataset.pinned).toBeUndefined();
	expect(room()).toBe("");
	expect(chat().scrollTop).toBe(0);
});

test("the first screen, let go while frames are due, is as the page was built once they would have come", () => {
	openHome("en");
	const layout = tall();
	lay(layout);
	const window = windowOf(layout);
	const stop = holdTheFirstScreen(document, window.window);
	window.scrollWithFrameDue(150);
	window.sizesChange();
	expect(window.framesDue()).toBe(2);

	stop();
	window.runFrames();

	expect(track().dataset.pinned).toBeUndefined();
	expect(room()).toBe("");
	expect(chat().scrollTop).toBe(0);
});

test("a window that refuses a listener leaves the first screen as the page was built, whatever sizes change", () => {
	openHome("en");
	const layout = tall();
	lay(layout);
	const window = windowOf(layout, true, true);

	expect(() => holdTheFirstScreen(document, window.window)).toThrow(
		"takes no listener",
	);
	window.sizesChange();
	window.runFrames();

	expect(track().dataset.pinned).toBeUndefined();
	expect(room()).toBe("");
});

test("a page with no phone on its first screen is left as it is", () => {
	document.body.innerHTML = '<section class="s-hero"></section>';
	const layout = tall();
	const window = windowOf(layout);

	const stop = holdTheFirstScreen(document, window.window);
	stop();

	expect(document.body.innerHTML).toBe('<section class="s-hero"></section>');
	expect(window.queries).toEqual([]);
});

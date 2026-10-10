import { afterEach, beforeEach, describe, expect, test, vi } from "vitest";
import { nextOf, playTheHistory, waitFor } from "./history";
import { openWhy } from "./testing/pages";

beforeEach(() => {
	vi.useFakeTimers();
	openWhy();
});

afterEach(() => {
	vi.useRealTimers();
	document.body.innerHTML = "";
});

// reduced is the query a browser answers for a reader who asks for less
// motion.
const reduced = "(prefers-reduced-motion: reduce)";

// windowAsking is a window on the test's clock, whose reader asks for less
// motion where still says so.
function windowAsking(still = false): Window {
	return {
		matchMedia: (query: string) => ({ matches: still && query === reduced }),
		setTimeout: (run: () => void, wait: number) => setTimeout(run, wait),
		clearTimeout: (timer?: number) => clearTimeout(timer),
	} as unknown as Window;
}

// the is the one element selector finds on the page.
function the(selector: string): HTMLElement {
	const element = document.querySelector<HTMLElement>(selector);
	if (element === null) {
		throw new Error(`the page Why has nothing ${selector} finds`);
	}
	return element;
}

// picks are the chips of the history's pictures, in the page's order.
const picks = () => [
	...document.querySelectorAll<HTMLInputElement>(".s-history-pick"),
];

// names are the names of the history's pictures, in the page's order.
const names = () => picks().map((pick) => pick.value);

// shown is the name of the picture whose chip is checked.
const shown = () => picks().find((pick) => pick.checked)?.value;

// following is the name of the picture that comes after the one called name.
const following = (name: string | undefined) =>
	names()[names().indexOf(name ?? "") + 1];

// era is the heading of the era at, counted from 0, whose words pick the
// era's first picture.
const era = (at: number) =>
	the(`.s-history-era:nth-child(${at + 1}) .s-history-era-text h3`);

// firstOf is the name of the first picture of the era at.
const firstOf = (at: number) =>
	the(`.s-history-era:nth-child(${at + 1}) [data-first-picture]`).dataset
		.firstPicture;

// held is whether the history says it waits.
const held = () => the("#history").hasAttribute("data-held");

describe("the picture after another", () => {
	test.each([
		["the next one", 0, 3, 1],
		["the first again after the last", 2, 3, 0],
		["the first where none is shown", -1, 3, 0],
		["the first of none", 0, 0, 0],
	])("is %s", (_, at, count, want) => {
		expect(nextOf(at, count)).toBe(want);
	});
});

describe("the pictures of the history", () => {
	test("move on by themselves, one every few seconds, and after the last the first again", () => {
		playTheHistory(document, windowAsking());
		const order = names();

		expect(shown()).toBe(order[0]);
		for (const name of [...order.slice(1), order[0]]) {
			vi.advanceTimersByTime(waitFor - 1);
			expect(shown()).not.toBe(name);
			vi.advanceTimersByTime(1);
			expect(shown()).toBe(name);
		}
	});

	test("mark the history live and playing, with the time a picture waits, and the stop puts the page back as it was", () => {
		const stop = playTheHistory(document, windowAsking());
		const history = the("#history");

		expect(history.hasAttribute("data-live")).toBe(true);
		expect(history.hasAttribute("data-playing")).toBe(true);
		expect(history.style.getPropertyValue("--s-history-wait")).toBe("5s");

		the(".s-history-split").dispatchEvent(new Event("pointerenter"));
		stop();

		expect(history.hasAttribute("data-live")).toBe(false);
		expect(history.hasAttribute("data-playing")).toBe(false);
		expect(held()).toBe(false);
		expect(history.style.getPropertyValue("--s-history-wait")).toBe("");
		vi.advanceTimersByTime(waitFor * 3);
		era(2).click();
		expect(shown()).toBe(names()[0]);
	});

	test("wait while a pointer rests on them, and go on with the time that was left", () => {
		playTheHistory(document, windowAsking());
		const split = the(".s-history-split");
		const first = shown();

		vi.advanceTimersByTime(2000);
		split.dispatchEvent(new Event("pointerenter"));
		expect(held()).toBe(true);
		vi.advanceTimersByTime(waitFor * 4);
		expect(shown()).toBe(first);

		split.dispatchEvent(new Event("pointerleave"));
		expect(held()).toBe(false);
		vi.advanceTimersByTime(waitFor - 2000 - 1);
		expect(shown()).toBe(first);
		vi.advanceTimersByTime(1);
		expect(shown()).toBe(following(first));
	});

	test("give a picture a moment more when the pointer leaves it about to go", () => {
		playTheHistory(document, windowAsking());
		const split = the(".s-history-split");
		const first = shown();

		vi.advanceTimersByTime(waitFor - 100);
		split.dispatchEvent(new Event("pointerenter"));
		split.dispatchEvent(new Event("pointerleave"));
		vi.advanceTimersByTime(399);
		expect(shown()).toBe(first);
		vi.advanceTimersByTime(1);
		expect(shown()).toBe(following(first));
	});

	test("wait while the keyboard's focus is in them, a pointer leaving or not, and go on once the focus leaves", () => {
		playTheHistory(document, windowAsking());
		const split = the(".s-history-split");
		const chip = picks()[0];
		const first = shown();

		vi.advanceTimersByTime(1000);
		split.dispatchEvent(new Event("pointerenter"));
		chip?.focus();
		split.dispatchEvent(new Event("pointerleave"));
		expect(held()).toBe(true);
		vi.advanceTimersByTime(waitFor * 2);
		expect(shown()).toBe(first);

		chip?.blur();
		expect(held()).toBe(false);
		vi.advanceTimersByTime(waitFor - 1000);
		expect(shown()).toBe(following(first));
	});

	test("hold nothing for a focus the page does not show, as a chip pressed with a pointer has", () => {
		playTheHistory(document, windowAsking());
		const first = shown();

		picks()[0]?.dispatchEvent(new FocusEvent("focusin", { bubbles: true }));
		expect(held()).toBe(false);
		vi.advanceTimersByTime(waitFor);
		expect(shown()).toBe(following(first));
	});

	test("start the wait anew when a chip is picked by hand", () => {
		playTheHistory(document, windowAsking());
		const chip = picks()[2];

		vi.advanceTimersByTime(3000);
		chip?.click();
		expect(shown()).toBe(chip?.value);
		vi.advanceTimersByTime(waitFor - 1);
		expect(shown()).toBe(chip?.value);
		vi.advanceTimersByTime(1);
		expect(shown()).toBe(following(chip?.value));
	});

	test("pick an era's first picture when its words are pressed, and start the wait anew", () => {
		playTheHistory(document, windowAsking());

		vi.advanceTimersByTime(3000);
		era(2).click();
		expect(shown()).toBe(firstOf(2));
		vi.advanceTimersByTime(waitFor - 1);
		expect(shown()).toBe(firstOf(2));
		vi.advanceTimersByTime(1);
		expect(shown()).toBe(following(firstOf(2)));
	});

	test("keep the picture shown when the words of its own era are pressed", () => {
		playTheHistory(document, windowAsking());

		vi.advanceTimersByTime(waitFor);
		const second = shown();
		expect(second).toBe(names()[1]);
		era(0).click();
		expect(shown()).toBe(second);
	});

	test("move on not at all for a reader who asks for less motion, whose eras' words still pick", () => {
		playTheHistory(document, windowAsking(true));
		const history = the("#history");

		expect(history.hasAttribute("data-live")).toBe(true);
		expect(history.hasAttribute("data-playing")).toBe(false);
		vi.advanceTimersByTime(waitFor * 5);
		expect(shown()).toBe(names()[0]);

		era(1).click();
		expect(shown()).toBe(firstOf(1));
		vi.advanceTimersByTime(waitFor * 5);
		expect(shown()).toBe(firstOf(1));
	});

	test.each([
		["no history", "<main><p>A page.</p></main>"],
		["a history with no pictures", '<section id="history"></section>'],
		[
			"a history with no chips",
			'<section id="history"><div class="s-history-split"></div></section>',
		],
	])("leave a page with %s as it is", (_, body) => {
		document.body.innerHTML = body;

		const stop = playTheHistory(document, windowAsking());

		expect(document.body.innerHTML).toBe(body);
		expect(vi.getTimerCount()).toBe(0);
		stop();
	});
});

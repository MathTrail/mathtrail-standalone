// @vitest-environment node
import { readdir } from "node:fs/promises";
import { tmpdir } from "node:os";
import { errors, type Frame, type Page } from "playwright-core";
import type { Plugin } from "vite";
import { afterEach, describe, expect, test, vi } from "vitest";
import {
	addressOf,
	againIfLate,
	builtPreview,
	Late,
	lateIn,
	markupOf,
	previewFolder,
	readWait,
	served,
	settled,
	settleWait,
	unsettled,
	within,
} from "./drive.ts";

afterEach(() => {
	vi.useRealTimers();
});

// never is a wait that never ends: a frame that has stopped answering.
const never = () => new Promise<never>(() => {});

describe("a wait given a time", () => {
	test("comes to what it waited for, and leaves no timer behind", async () => {
		vi.useFakeTimers();

		await expect(
			within(1000, "a card", Promise.resolve("drawn")),
		).resolves.toBe("drawn");
		expect(vi.getTimerCount()).toBe(0);
	});

	test("is refused once its time is up, naming what it waited for, and not before", async () => {
		vi.useFakeTimers();
		let ended = false;
		const waited = within(15_000, "reading the card task", never());
		const refused = expect(waited).rejects.toThrow(
			"drive: reading the card task took longer than 15 s",
		);
		waited
			.catch(() => {})
			.finally(() => {
				ended = true;
			});

		await vi.advanceTimersByTimeAsync(14_999);
		expect(ended).toBe(false);
		await vi.advanceTimersByTimeAsync(1);
		await refused;
	});
});

describe("a browser that stopped answering", () => {
	// late is the refusal of a wait on a browser that stopped answering.
	const late = new Late("drive: finding the cards took longer than 15 s");

	test("is told by its refusal, and by a refusal it caused", () => {
		const opening = new errors.TimeoutError("page.goto: Timeout 30000ms");

		expect(lateIn(late)).toBe(late);
		expect(lateIn(opening)).toBe(opening);
		expect(lateIn(new Error("layout: webkit en 320px", { cause: late }))).toBe(
			late,
		);
		expect(lateIn(new Error("drive: the cards never settled"))).toBeUndefined();
		expect(lateIn("late")).toBeUndefined();
	});

	test("is renewed, and the attempt made once more", async () => {
		const attempts = [() => Promise.reject(late), () => Promise.resolve(2)];
		const renewed: Late[] = [];

		await expect(
			againIfLate(
				() => attempts.shift()?.() ?? Promise.resolve(0),
				async (refusal) => {
					renewed.push(refusal);
				},
			),
		).resolves.toBe(2);
		expect(renewed).toEqual([late]);
	});

	test("is renewed when Playwright's own time for a wait runs out", async () => {
		const opening = new errors.TimeoutError("page.goto: Timeout 30000ms");
		const attempts = [() => Promise.reject(opening), () => Promise.resolve(2)];
		let renewed = 0;

		await expect(
			againIfLate(
				() => attempts.shift()?.() ?? Promise.resolve(0),
				async () => {
					renewed++;
				},
			),
		).resolves.toBe(2);
		expect(renewed).toBe(1);
	});

	test("stops the run when the second one stops answering too", async () => {
		const second = new Late("drive: finding the cards took longer than 15 s");
		const attempts = [() => Promise.reject(late), () => Promise.reject(second)];

		await expect(
			againIfLate(
				() => attempts.shift()?.() ?? Promise.resolve(0),
				async () => {},
			),
		).rejects.toBe(second);
	});

	test("is not what a card that never settles is, and nothing is tried again", async () => {
		const unsettledCards = new Error("drive: the cards never settled");
		let tried = 0;
		let renewed = 0;

		await expect(
			againIfLate(
				async () => {
					tried++;
					throw unsettledCards;
				},
				async () => {
					renewed++;
				},
			),
		).rejects.toBe(unsettledCards);
		expect([tried, renewed]).toEqual([1, 0]);
	});

	test("is not met by an attempt that comes to its end", async () => {
		let renewed = 0;

		await expect(
			againIfLate(
				() => Promise.resolve(1),
				async () => {
					renewed++;
				},
			),
		).resolves.toBe(1);
		expect(renewed).toBe(0);
	});
});

describe("what a card holds", () => {
	// cardAnswering is the card of the scene task, whose frame answers with
	// answer.
	const cardAnswering = (answer: () => Promise<string>) => ({
		scene: "task",
		frame: { evaluate: answer } as unknown as Frame,
	});

	test("is its markup", async () => {
		await expect(
			markupOf(cardAnswering(() => Promise.resolve("<p>1</p>"))),
		).resolves.toBe("<p>1</p>");
	});

	test("is nothing while its frame is between pages", async () => {
		await expect(
			markupOf(
				cardAnswering(() =>
					Promise.reject(new Error("Execution context was destroyed")),
				),
			),
		).resolves.toBe("");
	});

	test("is refused by its scene's name once the card stops answering", async () => {
		vi.useFakeTimers();
		const refused = expect(markupOf(cardAnswering(never))).rejects.toThrow(
			`drive: reading the card task took longer than ${readWait / 1000} s`,
		);

		await vi.advanceTimersByTimeAsync(readWait);
		await refused;
	});
});

// kept is the text of the file at path of the preview served at base, which
// the browser is let keep for the run.
async function kept(base: string, path: string): Promise<string> {
	const response = await fetch(new URL(path, base));
	expect(response.status, path).toBe(200);
	expect(response.headers.get("cache-control"), path).toBe(
		"max-age=31536000, immutable",
	);
	return response.text();
}

// filesOf are the built files a text names: the scripts and the styles a
// page loads, and the parts of a script it imports.
function filesOf(text: string): string[] {
	return [
		...text.matchAll(/(?:src|href)="(\/assets\/[^"]+)"|from"\.\/([^"]+\.js)"/g),
	].map(([, named, imported]) => named ?? `/assets/${imported}`);
}

// previewFolders are the folders built previews are served from.
const previewFolders = async () =>
	(await readdir(tmpdir())).filter((name) => name.startsWith(previewFolder));

describe("the preview the cards are driven in", () => {
	test("is built with the environment given back as it was", async () => {
		const environment = process.env.NODE_ENV;

		await builtPreview({});

		expect(process.env.NODE_ENV).toBe(environment);
	}, 60_000);

	test("leaves no folder behind when its build fails", async () => {
		const before = await previewFolders();
		const failing: Plugin = {
			name: "failing",
			buildStart() {
				throw new Error("the build failed");
			},
		};

		await expect(served({ plugins: [failing] })).rejects.toThrow(
			"the build failed",
		);
		expect(await previewFolders()).toEqual(before);
	}, 60_000);

	test("is built, each of its files kept by the browser for the run, and speaks the pseudo-language", async () => {
		const preview = await served();
		try {
			const read = new Map<string, string>();
			const toRead = ["preview.html", "widget.html"];
			for (
				let path = toRead.shift();
				path !== undefined;
				path = toRead.shift()
			) {
				if (!read.has(path)) {
					const text = await kept(preview.base, path);
					read.set(path, text);
					toRead.push(...filesOf(text));
				}
			}

			expect(
				[...read.keys()].filter((path) => path.endsWith(".js")).length,
			).toBeGreaterThan(1);
			expect([...read.values()].some((text) => text.includes("en-XA"))).toBe(
				true,
			);
		} finally {
			await preview.close();
		}
	}, 60_000);
});

describe("cards that never settle", () => {
	test("are given up on once their time is up, however few ticks that took, and named", async () => {
		vi.useFakeTimers();
		let ticks = 0;
		// A page of one card, task, that is never drawn, on a machine whose
		// every tick takes a second.
		const card = {
			contentFrame: async () => ({ evaluate: () => Promise.resolve("") }),
			getAttribute: async () => "task",
			dispose: async () => {},
		};
		const page = {
			clock: {
				runFor: async () => {
					ticks++;
				},
			},
			locator: () => ({ elementHandles: async () => [card] }),
			waitForTimeout: () =>
				new Promise<void>((done) => {
					setTimeout(done, 1000);
				}),
		} as unknown as Page;
		const refused = expect(settled(page)).rejects.toThrow(
			"drive: the cards never settled: nothing drawn in task",
		);

		await vi.advanceTimersByTimeAsync(settleWait + 2000);
		await refused;
		expect(ticks).toBeLessThan(settleWait / 1000 + 2);
	});
});

describe("what kept the cards from settling", () => {
	test("names by scene the cards with nothing drawn and those still changing", () => {
		expect(
			unsettled(
				["task", "wrong answer", "progress"],
				["", "<p>2</p>", "<p>1</p>"],
				["", "<p>1</p>", "<p>1</p>"],
			),
		).toBe("nothing drawn in task; still changing: wrong answer");
	});

	test("names only the cards still changing when every card is drawn", () => {
		expect(
			unsettled(
				["task", "progress"],
				["<p>2</p>", "<p>1</p>"],
				["<p>1</p>", "<p>1</p>"],
			),
		).toBe("still changing: task");
	});

	test("counts the cards when some came or went in between", () => {
		expect(
			unsettled(["task", "progress"], ["<p>1</p>", "<p>1</p>"], ["<p>1</p>"]),
		).toBe("cards came or went: 1 a tick before, 2 now");
	});

	test("says so when the page holds no card", () => {
		expect(unsettled([], [], [])).toBe("no card on the page");
	});

	test("is nothing once every card is drawn and holds what it held", () => {
		expect(
			unsettled(
				["task", "progress"],
				["<p>1</p>", "<p>2</p>"],
				["<p>1</p>", "<p>2</p>"],
			),
		).toBe("");
	});

	test("keeps a card drawn for the first time from counting as settled", () => {
		expect(unsettled(["task"], ["<p>1</p>"], [""])).toBe(
			"still changing: task",
		);
	});
});

describe("the address of a picture", () => {
	test("asks the preview for its scene alone, in its theme, in English, at its width", () => {
		const address = new URL(
			addressOf("http://127.0.0.1:5173/", {
				scene: "wrong after the trial series",
				theme: "light",
				width: 640,
				path: "wrong.png",
			}),
		);
		expect(address.pathname).toBe("/preview.html");
		expect(Object.fromEntries(address.searchParams)).toEqual({
			scene: "wrong after the trial series",
			theme: "light",
			lang: "en",
			widths: "640",
		});
	});
});

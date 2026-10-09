import { describe, expect, test } from "vitest";
import { loadInTurn, turns } from "./turns";

describe("turns", () => {
	test("lets as many take a turn at once as there are turns, and no more", async () => {
		const loading = turns(2);
		const taken: string[] = [];
		for (const card of ["first", "second", "third"]) {
			void loading.take().then(() => taken.push(card));
		}
		await settledPromises();
		expect(taken).toEqual(["first", "second"]);
	});

	test("hands a turn that ends to whoever has waited longest", async () => {
		const loading = turns(1);
		const taken: string[] = [];
		const end = await loading.take();
		void loading.take().then((ended) => {
			taken.push("second");
			ended();
		});
		void loading.take().then(() => taken.push("third"));
		end();
		await settledPromises();
		expect(taken).toEqual(["second", "third"]);
	});

	test("ends a turn once, however often its end is called", async () => {
		const loading = turns(1);
		const taken: string[] = [];
		const end = await loading.take();
		void loading.take().then(() => taken.push("second"));
		void loading.take().then(() => taken.push("third"));
		end();
		end();
		await settledPromises();
		expect(taken).toEqual(["second"]);
	});
});

describe("loading a frame in turn", () => {
	test("points no more frames at the page at once than there are turns", async () => {
		const loading = turns(2);
		const frames = [aFrame(), aFrame(), aFrame()];
		for (const frame of frames) {
			loadInTurn(frame, "/widget.html", loading);
		}
		await settledPromises();
		expect(pointed(frames)).toEqual([true, true, false]);
		expect(frames[0]?.getAttribute("src")).toBe("/widget.html");
	});

	test("points the next frame at the page once one has loaded", async () => {
		const loading = turns(1);
		const [first, second] = [aFrame(), aFrame()];
		loadInTurn(first, "/widget.html", loading);
		loadInTurn(second, "/widget.html", loading);
		await settledPromises();
		first.dispatchEvent(new Event("load"));
		await settledPromises();
		expect(pointed([first, second])).toEqual([true, true]);
	});

	test("gives back the turn of a frame stopped while it loads", async () => {
		const loading = turns(1);
		const [first, second] = [aFrame(), aFrame()];
		const stop = loadInTurn(first, "/widget.html", loading);
		loadInTurn(second, "/widget.html", loading);
		await settledPromises();
		stop();
		await settledPromises();
		expect(pointed([first, second])).toEqual([true, true]);
	});

	test("never points a frame stopped before its turn came, and hands the turn on", async () => {
		const loading = turns(1);
		const [first, second, third] = [aFrame(), aFrame(), aFrame()];
		loadInTurn(first, "/widget.html", loading);
		const stop = loadInTurn(second, "/widget.html", loading);
		loadInTurn(third, "/widget.html", loading);
		await settledPromises();
		stop();
		first.dispatchEvent(new Event("load"));
		await settledPromises();
		expect(pointed([first, second, third])).toEqual([true, false, true]);
	});

	test("gives back no second turn when a frame that has loaded is stopped", async () => {
		const loading = turns(1);
		const [first, second, third] = [aFrame(), aFrame(), aFrame()];
		const stop = loadInTurn(first, "/widget.html", loading);
		loadInTurn(second, "/widget.html", loading);
		loadInTurn(third, "/widget.html", loading);
		await settledPromises();
		first.dispatchEvent(new Event("load"));
		await settledPromises();
		stop();
		await settledPromises();
		expect(pointed([first, second, third])).toEqual([true, true, false]);
	});
});

// aFrame is a frame standing apart from any page, which loads nothing when it
// is pointed at one.
function aFrame(): HTMLIFrameElement {
	return document.createElement("iframe");
}

// pointed says of each frame whether it has been pointed at a page.
function pointed(frames: HTMLIFrameElement[]): boolean[] {
	return frames.map((frame) => frame.hasAttribute("src"));
}

// settledPromises waits until every promise settled so far has run what it
// was waiting to run.
function settledPromises(): Promise<void> {
	return new Promise((resolve) => setTimeout(resolve, 0));
}

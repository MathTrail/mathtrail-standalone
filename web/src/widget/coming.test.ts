import { describe, expect, test } from "vitest";
import {
	askIn,
	firstAskIn,
	heldOn,
	isSettled,
	shows,
	type Wait,
	type WaitEvent,
	waitAfter,
	waitStart,
} from "./coming";
import type { TaskStatus } from "./payload";
import { fence } from "./testing/lesson";
import { moments } from "./waiting";

// told is the wait once each status given has been answered, in order.
function told(...statuses: TaskStatus[]): Wait {
	return statuses.reduce(
		(wait, status) => waitAfter(wait, { type: "answered", status }),
		waitStart,
	);
}

const writing = (refused = 0): TaskStatus => ({ kind: "writing", refused });
const onTheCard: TaskStatus = { kind: "task", handed: fence };
const over: TaskStatus = { kind: "over" };
const unknown: TaskStatus = { kind: "unknown" };

describe("a card waiting for a task", () => {
	test("shows it being written, and asks again soon", () => {
		const wait = told(writing());

		expect(shows(wait)).toBe("writing");
		expect(wait.seenWriting).toBe(true);
		expect(askIn(wait)).toBe(moments.ask);
	});

	test("counts the tries turned down, and never fewer than it was told", () => {
		expect(told(writing(), writing(1)).refused).toBe(1);
		expect(told(writing(2), writing(1)).refused).toBe(2);
	});

	test("turns into the task once it is on the card, and asks no more", () => {
		const wait = told(writing(), onTheCard);

		expect(wait.task).toBe(fence);
		expect(isSettled(wait)).toBe(true);
		expect(askIn(wait)).toBeUndefined();
	});

	test("says no task is coming at the first word of it, and settles at the second", () => {
		const once = told(writing(), over);
		expect(shows(once)).toBe("not coming");
		expect(isSettled(once)).toBe(false);
		expect(askIn(once)).toBe(moments.ask);

		const twice = waitAfter(once, { type: "answered", status: over });
		expect(isSettled(twice)).toBe(true);
		expect(askIn(twice)).toBeUndefined();
	});

	test("goes back to the wait when the word that none is coming was a read too early", () => {
		const wait = told(over, writing());

		expect(shows(wait)).toBe("writing");
		expect(wait.notComing).toBe(0);
		expect(told(over, onTheCard).task).toBe(fence);
	});

	test("asks again seldom after a question that went unanswered, and as often as before once one is answered", () => {
		expect(askIn(told(writing(), unknown))).toBe(moments.askSlowly);
		expect(askIn(told(writing(), unknown, writing()))).toBe(moments.ask);
	});

	test("takes an unanswered question for no news: nothing it showed changes", () => {
		const before = told(writing(1));
		const after = waitAfter(before, { type: "answered", status: unknown });

		expect({ ...after, answers: 0, unanswered: 0 }).toEqual({
			...before,
			answers: 0,
		});
	});

	test("makes sure no task is coming at the usual pace, whether or not it heard of the task being written", () => {
		expect(askIn(told(over))).toBe(moments.ask);
		expect(askIn(told(writing(), over))).toBe(moments.ask);
	});

	test("puts a seldom question off by its own moment as well, and an often one not", () => {
		expect(askIn(told(writing(), unknown), 700)).toBe(moments.askSlowly + 700);
		expect(askIn(told(writing()), 700)).toBe(moments.ask);
	});

	test("stops asking once twenty questions in a row went unanswered, and says the wait is long", () => {
		const unanswered = Array.from({ length: 20 }, () => unknown);
		const nearly = told(writing(), ...unanswered.slice(1));

		expect(isSettled(nearly)).toBe(false);
		expect(askIn(nearly)).toBe(moments.askSlowly);
		const given = waitAfter(nearly, { type: "answered", status: unknown });
		expect(isSettled(given)).toBe(true);
		expect(given.slow).toBe(true);
		expect(shows(given)).toBe("writing");
		expect(isSettled(told(writing(), ...unanswered.slice(1), writing()))).toBe(
			false,
		);
	});
});

describe("a wait that goes long", () => {
	// slowed is the wait once the clock of its latest news ran out.
	const slowed = (wait: Wait): Wait =>
		waitAfter(wait, { type: "slowed", news: wait.news });

	test("says so once the clock of its latest news runs out, and asks seldom", () => {
		const wait = slowed(told(writing()));

		expect(wait.slow).toBe(true);
		expect(askIn(wait)).toBe(moments.askSlowly);
	});

	test("is not long while news keeps coming: a clock of older news is spent", () => {
		const fresh = told(writing(), writing(1));

		expect(waitAfter(fresh, { type: "slowed", news: fresh.news - 1 })).toBe(
			fresh,
		);
	});

	test("is long no more once there is news: another try turned down", () => {
		const wait = waitAfter(slowed(told(writing())), {
			type: "answered",
			status: writing(1),
		});

		expect(wait.slow).toBe(false);
		expect(askIn(wait)).toBe(moments.ask);
	});

	test("hears nothing new in the same word told again", () => {
		const wait = slowed(told(writing(1)));

		expect(waitAfter(wait, { type: "answered", status: writing(1) }).slow).toBe(
			true,
		);
	});
});

describe("a wait that is settled", () => {
	test.each<[string, Wait]>([
		["into its task", told(onTheCard)],
		["into no task coming", told(over, over)],
	])("moves no more, %s", (_, settled) => {
		for (const event of [
			{ type: "answered", status: writing(2) },
			{ type: "answered", status: over },
			{ type: "slowed", news: settled.news },
		] satisfies WaitEvent[]) {
			expect(waitAfter(settled, event)).toBe(settled);
		}
	});
});

describe("the first question of a card", () => {
	test("waits a moment of its request's own, under the spread, the same each time", () => {
		const ids = Array.from({ length: 50 }, (_, at) => `req_${at}-${at * 7919}`);
		const firsts = ids.map(firstAskIn);

		for (const first of firsts) {
			expect(first).toBeGreaterThanOrEqual(0);
			expect(first).toBeLessThan(moments.spread);
		}
		expect(new Set(firsts).size).toBeGreaterThan(25);
		expect(firstAskIn(ids[0] ?? "")).toBe(firsts[0]);
	});
});

describe("what a card's question says it has heard", () => {
	test.each<[string, Wait, number | undefined]>([
		["nothing before its first answer", waitStart, undefined],
		["no try turned down once it hears of the task", told(writing()), 0],
		["the tries turned down it has heard of", told(writing(), writing(2)), 2],
		[
			"nothing once a word says none is coming",
			told(writing(), over),
			undefined,
		],
		["nothing while it has not heard of the task", told(over), undefined],
		[
			"the tries still after one question held for news went unanswered",
			told(writing(), unknown),
			0,
		],
		[
			"nothing once two questions held for news in a row went unanswered",
			told(writing(), unknown, unknown),
			undefined,
		],
		[
			"nothing for good after that",
			told(writing(), unknown, unknown, writing()),
			undefined,
		],
		[
			"the tries again after a question not held went unanswered",
			told(unknown, writing()),
			0,
		],
	])("says %s", (_, wait, heard) => {
		expect(heldOn(wait)).toBe(heard);
	});
});

describe("the pace of a card's questions", () => {
	test("runs from the start of each question: one held for news is followed at once, one answered at once waits out the rest", () => {
		const wait = told(writing());

		expect(askIn(wait, 0, 1_000)).toBe(moments.ask - 1_000);
		expect(askIn(wait, 0, moments.ask + 500)).toBe(0);
	});

	test("waits the whole pace when the clock was set back since the question started", () => {
		expect(askIn(told(writing()), 0, -60_000)).toBe(moments.ask);
	});

	test("runs the seldom pace of a wait gone long from the start of the question too", () => {
		const fresh = told(writing());
		const long = waitAfter(fresh, { type: "slowed", news: fresh.news });

		expect(askIn(long, 700, 1_000)).toBe(moments.askSlowly + 700 - 1_000);
	});

	test("runs the pause after a question that went unanswered from its end, however long it took to fail", () => {
		expect(askIn(told(writing(), unknown), 700, 20_000)).toBe(
			moments.askSlowly + 700,
		);
	});
});

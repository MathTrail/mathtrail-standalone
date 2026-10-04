import { useState } from "preact/hooks";
import type { MoveWay } from "../design/progress";
import type { Move, Moves } from "./payload";

/**
 * Period is a while the progress tells a move over: since the last task, or
 * over the week.
 */
export type Period = "last_task" | "week";

/**
 * PeriodChoice is the while chosen on the progress's switch, undefined until
 * one is, and how to choose one.
 */
export type PeriodChoice = {
	chosen: Period | undefined;
	choose: (period: Period) => void;
};

/**
 * usePeriod is the while chosen on the progress's switch, kept for as long as
 * the component that calls it is drawn: a card keeps it while the progress
 * over it is closed and opened again, as it keeps the sections that fold.
 */
export function usePeriod(): PeriodChoice {
	const [chosen, choose] = useState<Period | undefined>(undefined);
	return { chosen, choose };
}

/**
 * periodShown is the while the progress draws: the one chosen, or else the
 * week when it can be told, the last task when only that can, and the week
 * when neither can.
 */
export function periodShown(
	chosen: Period | undefined,
	moves: Moves | undefined,
): Period {
	if (chosen !== undefined) {
		return chosen;
	}
	return moves?.week === null && moves.last_task !== null
		? "last_task"
		: "week";
}

/** moveOver is the move of the while period, or null where none is told. */
export function moveOver(moves: Moves, period: Period): Move | null {
	return period === "week" ? moves.week : moves.last_task;
}

/**
 * Reading is a move as the card draws it: a while that cannot be told; no
 * move; a topic first answered in it; a move one way, from the rank and the
 * share it stood at, and whether the rank itself changed; or a word this card
 * does not know, or a move it cannot draw, which draws nothing.
 */
export type Reading =
	| { kind: "untold" }
	| { kind: "same" }
	| { kind: "new" }
	| {
			kind: "moved";
			way: MoveWay;
			ranked: boolean;
			rank: number;
			share: number;
	  }
	| { kind: "unknown" };

/** readingOf is a move as the card draws it. */
export function readingOf(move: Move | null): Reading {
	if (move === null) {
		return { kind: "untold" };
	}
	switch (move.moved) {
		case "same":
			return { kind: "same" };
		case "new":
			return { kind: "new" };
		case "rank_up":
		case "forward":
		case "back":
		case "rank_down":
			if (move.rank === undefined || move.share === undefined) {
				return { kind: "unknown" };
			}
			return {
				kind: "moved",
				way:
					move.moved === "rank_up" || move.moved === "forward"
						? "gain"
						: "loss",
				ranked: move.moved === "rank_up" || move.moved === "rank_down",
				rank: move.rank,
				share: move.share,
			};
		default:
			return { kind: "unknown" };
	}
}

/**
 * movesCounted is how many of the readings moved up — a topic first answered
 * among them, since it gained a rank — and how many moved back.
 */
export function movesCounted(readings: readonly Reading[]): {
	up: number;
	down: number;
} {
	let up = 0;
	let down = 0;
	for (const reading of readings) {
		if (
			reading.kind === "new" ||
			(reading.kind === "moved" && reading.way === "gain")
		) {
			up++;
		} else if (reading.kind === "moved") {
			down++;
		}
	}
	return { up, down };
}

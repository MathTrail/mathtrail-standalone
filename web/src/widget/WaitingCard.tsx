import { Verdict } from "../design/blocks";
import type { Host } from "./bridge";
import { CardFrame } from "./CardFrame";
import { CardHeader } from "./CardRoot";
import type { Waiting } from "./payload";
import { type Key, useWords } from "./words";

/**
 * WaitingCard is a card a task did not come to: the model's attempt failed its
 * checks, the request it was handed in for is over, its last attempt failed
 * too, or the day has no room for another. It says what happened and what
 * comes next, and nothing on it moves or asks: it cannot see the next attempt
 * being written, which comes in a card of its own, and a host may hold a
 * card's message for the person to send.
 */
export function WaitingCard({
	waiting,
	host,
}: {
	waiting: Waiting;
	host: Host;
}) {
	const words = useWords();
	const [happened, next] = linesOf[waiting.kind];
	return (
		<CardFrame
			child={waiting.child}
			back={words.text("progress.back_plain")}
			host={host}
		>
			{(wide) => (
				<article aria-label={words.text("waiting.label")}>
					<CardHeader grade={waiting.child?.grade} wide={wide} />
					<div class="mt-body">
						<Verdict detail={words.text(next)}>{words.text(happened)}</Verdict>
					</div>
				</article>
			)}
		</CardFrame>
	);
}

// The words each card says: what happened, then what comes next. A task
// handed in again comes below, in a card of its own, when the model writes
// one, as it is told to; a model stopped short writes none, so the words say
// where to ask then.
const linesOf: Readonly<Record<Waiting["kind"], readonly [Key, Key]>> = {
	refused: ["waiting.refused", "waiting.next_below"],
	stale: ["waiting.stale", "waiting.next_below"],
	exhausted: ["waiting.exhausted", "waiting.next_below"],
	limit: ["waiting.limit", "waiting.limit_detail"],
};

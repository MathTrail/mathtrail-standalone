import { useEffect, useReducer, useRef, useState } from "preact/hooks";
import type { Host } from "./bridge";
import { CardFrame } from "./CardFrame";
import { CardHeader } from "./CardRoot";
import {
	askIn,
	firstAskIn,
	isSettled,
	shows,
	type Wait,
	waitAfter,
	waitStart,
} from "./coming";
import type { Coming } from "./payload";
import { useService } from "./service";
import { TaskInCard } from "./TaskCard";
import { TaskWait } from "./TaskWait";
import { moments, type Phase } from "./waiting";
import { useWords } from "./words";

/**
 * ComingCard is the card a task asked for comes to. It is drawn as soon as the
 * task is asked for, and waits for it: it asks the service how the task
 * stands, and shows the task being written — and a try the checks turned
 * down, with a new one being written, and the wait gone long — until the task
 * is on the card. It then ticks off the rest of the course, the checks the
 * task passed and the task ready, and turns into the task, in the same frame,
 * the progress opened over it left open. A task that is not coming is said
 * so, in words true whatever ended its request — the tries spent, the request
 * replaced, or the task gone from the card long since, for a card drawn again
 * with an earlier chat: no task is here, and where the next one comes.
 */
export function ComingCard({ coming, host }: { coming: Coming; host: Host }) {
	const words = useWords();
	const wait = useWaitFor(coming.requestId);
	const finishing = useFinishing(wait);
	const task = finishing === undefined ? wait.task : undefined;
	return (
		<CardFrame
			child={coming.child}
			back={words.text(
				task === undefined ? "progress.back_plain" : "progress.back",
			)}
			host={host}
		>
			{(wide, whose) =>
				task === undefined ? (
					<article aria-label={words.text("waiting.label")}>
						<CardHeader grade={whose.grade} wide={wide} />
						<TaskWait
							phase={finishing ?? "writing"}
							rewriting={finishing === undefined && wait.refused > 0}
							slow={finishing === undefined && wait.slow}
							ended={
								shows(wait) === "not coming"
									? { said: "waiting.stale", next: "waiting.next_below" }
									: undefined
							}
						/>
					</article>
				) : (
					<TaskInCard
						key={task.task.id}
						handed={task}
						host={host}
						wide={wide}
						grade={whose.grade}
					/>
				)
			}
		</CardFrame>
	);
}

// useWaitFor is the wait for the task of the request given, kept up by asking
// the service how it stands: first after a moment of the request's own, then
// each time the question before was answered, as often as the wait says, until
// it is settled. A page out of sight asks nothing, and asks at once when it is
// looked at again.
function useWaitFor(requestId: string): Wait {
	const service = useService();
	const [wait, dispatch] = useReducer(waitAfter, waitStart);
	// latest is the wait as it stands, which the next question is timed by. A
	// change of pace alone — the wait gone long — sets no question going afresh,
	// and drops no answer already on its way.
	const latest = useRef(wait);
	latest.current = wait;
	const settled = isSettled(wait);

	useEffect(() => {
		if (settled) {
			return;
		}
		let gone = false;
		const ask = async () => {
			if (document.visibilityState === "hidden") {
				document.addEventListener("visibilitychange", whenSeen);
				return;
			}
			const status = await service.taskStatus(requestId);
			if (!gone) {
				dispatch({ type: "answered", status });
			}
		};
		const whenSeen = () => {
			if (document.visibilityState !== "hidden") {
				document.removeEventListener("visibilitychange", whenSeen);
				void ask();
			}
		};
		const pause =
			wait.answers === 0
				? firstAskIn(requestId)
				: (askIn(latest.current, firstAskIn(requestId)) ?? moments.askSlowly);
		const timer = setTimeout(() => void ask(), pause);
		return () => {
			gone = true;
			clearTimeout(timer);
			document.removeEventListener("visibilitychange", whenSeen);
		};
	}, [service, requestId, wait.answers, settled]);

	useEffect(() => {
		if (settled) {
			return;
		}
		const timer = setTimeout(
			() => dispatch({ type: "slowed", news: wait.news }),
			moments.slow,
		);
		return () => clearTimeout(timer);
	}, [wait.news, settled]);

	return wait;
}

// Ticked is how far a card has ticked off the rest of the course since its
// task came: the checks, the task ready, and the course done.
type Ticked = "checked" | "ready" | "done";

// useFinishing is the phase of the course a card shows between its task coming
// and the task taking the course's place: the checks passed at once, the task
// ready a beat later, held a moment so that it is seen. It is nothing while the
// task is awaited and once the course is done, and nothing at all for a card
// that never saw the task being written — one drawn again with an earlier
// chat, its task there at the first question —, which shows its task at once.
function useFinishing(wait: Wait): Phase | undefined {
	const come = wait.task !== undefined && wait.seenWriting;
	const [ticked, setTicked] = useState<Ticked>("checked");

	useEffect(() => {
		if (!come) {
			return;
		}
		const ready = setTimeout(() => setTicked("ready"), moments.beat);
		const done = setTimeout(
			() => setTicked("done"),
			moments.beat + moments.held,
		);
		return () => {
			clearTimeout(ready);
			clearTimeout(done);
		};
	}, [come]);

	return come && ticked !== "done" ? ticked : undefined;
}

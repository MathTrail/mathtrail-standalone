import { useEffect, useReducer, useRef } from "preact/hooks";
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
import { type Coming, readTaskStatus, type TaskStatus } from "./payload";
import { TaskInCard } from "./TaskCard";
import { TaskWait } from "./TaskWait";
import { moments } from "./waiting";
import { useWords } from "./words";

/**
 * ComingCard is the card a task asked for comes to. It is drawn as soon as the
 * task is asked for, and waits for it: it asks the service how the task
 * stands, and shows the task being written — and a try the checks turned
 * down, with a new one being written, and the wait gone long — until the task
 * is on the card, and then turns into it, in the same frame, the progress
 * opened over it left open. A task that is not coming is said so, in words
 * true whatever ended its request — the tries spent, the request replaced, or
 * the task gone from the card long since, for a card drawn again with an
 * earlier chat: no task is here, and where the next one comes.
 */
export function ComingCard({ coming, host }: { coming: Coming; host: Host }) {
	const words = useWords();
	const wait = useWaitFor(host, coming.requestId);
	const { task } = wait;
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
							phase="writing"
							rewriting={wait.refused > 0}
							slow={wait.slow}
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
function useWaitFor(host: Host, requestId: string): Wait {
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
			const status = await statusOf(host, requestId);
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
	}, [host, requestId, wait.answers, settled]);

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

// statusOf asks the service how the task of the request stands. A question
// whose answer never came leaves it unknown, and is asked again later.
async function statusOf(host: Host, requestId: string): Promise<TaskStatus> {
	try {
		return readTaskStatus(
			await host.callTool("read_task", { request_id: requestId }),
		);
	} catch (error: unknown) {
		console.error("widget: how the task stands did not arrive", error);
		return { kind: "unknown" };
	}
}

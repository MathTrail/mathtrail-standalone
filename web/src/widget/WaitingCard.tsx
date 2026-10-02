import { useEffect, useReducer, useRef, useState } from "preact/hooks";
import { GeneratingSteps, Verdict } from "../design/blocks";
import { Button } from "../design/controls";
import type { Host } from "./bridge";
import { CardFrame } from "./CardFrame";
import { CardHeader } from "./CardRoot";
import { useFocusKeptOnTheCard } from "./focus";
import type { Waiting } from "./payload";
import {
	isLate,
	moments,
	stepsAt,
	type Wait,
	waitAfter,
	waitStart,
} from "./waiting";
import { useWords } from "./words";

/**
 * WaitingCard is a card drawn while the next task is written — after the model
 * handed one in that failed its checks — or once no task is coming: the
 * model's last attempt failed too, or the day has no room for another.
 */
export function WaitingCard({
	waiting,
	host,
}: {
	waiting: Waiting;
	host: Host;
}) {
	const words = useWords();
	const asking = useAsking(host);
	return (
		<CardFrame
			child={waiting.child}
			back={words.text("progress.back_plain")}
			host={host}
		>
			{(wide) =>
				waiting.kind === "working" ? (
					<WaitingScreen
						grade={waiting.child.grade}
						wide={wide}
						asking={asking}
					/>
				) : (
					<NoTaskScreen
						grade={waiting.child?.grade}
						wide={wide}
						why={waiting.kind}
					/>
				)
			}
		</CardFrame>
	);
}

/**
 * Asking is a card's asking for the next task: where its wait stands, and the
 * ask that starts a new round of it.
 */
type Asking = { wait: Wait; ask: () => void };

/**
 * useAsking is a card's asking for the next task. Each ask sends the label of
 * "Another task", in the card's language, to the chat as the child's message —
 * only the model can write a task, and those are the words it takes as the
 * ask — and starts a round of the wait afresh; one the chat does not take
 * makes its round late at once.
 */
function useAsking(host: Host): Asking {
	const words = useWords();
	const [wait, dispatch] = useReducer(waitAfter, waitStart);
	const asks = useRef(0);
	function ask() {
		// A second press that comes before the card has redrawn is the same ask.
		if (wait.round !== asks.current) {
			return;
		}
		asks.current += 1;
		const round = asks.current;
		dispatch({ type: "asked", round });
		host.sendMessage(words.text("task.another")).catch((error: unknown) => {
			console.error(
				"widget: the ask for another task did not reach the chat",
				error,
			);
			dispatch({ type: "lost", round });
		});
	}
	return { wait, ask };
}

/**
 * WaitingScreen is what a card shows while the next task is written: the usual
 * course of a task, moving by the clock and claiming no numbers, since the card
 * cannot see the work; and — once it has lasted well past the time a task
 * takes, or the ask never reached the chat — the plain word that no task is
 * coming unless a new one is below, and a way to ask again. A screen reader
 * hears each change once, as it comes: the steps are heard from their own
 * list, and the word that no task is coming from a place beside it, on the
 * page from the start.
 */
function WaitingScreen({
	grade,
	wide,
	asking,
}: {
	grade: number;
	wide: boolean;
	asking: Asking;
}) {
	const words = useWords();
	const seconds = useSecondsWaited(asking.wait.round);
	const late = isLate(asking.wait, seconds);
	const title = useRef<HTMLParagraphElement>(null);
	const askAgain = useRef<HTMLButtonElement>(null);
	// A focus lost to a button gone goes to what took its place: the title as
	// a round starts, the way to ask again as one is given up on.
	useFocusKeptOnTheCard(!late, title);
	useFocusKeptOnTheCard(late, askAgain);

	return (
		<>
			<article aria-label={words.text("waiting.label")}>
				<CardHeader grade={grade} wide={wide} />
				<div class="mt-body">
					<div class="mt-wait">
						{!late && (
							<GeneratingSteps
								title={words.text("waiting.title")}
								steps={stepsAt(seconds).map(({ key, status }) => ({
									label: words.text(key, { grade }),
									status,
								}))}
								statusLabels={{
									done: words.text("waiting.done"),
									active: words.text("waiting.active"),
									waiting: words.text("waiting.waiting"),
								}}
								titleRef={title}
							/>
						)}
						<div class="mt-news" aria-live="polite">
							{late && (
								<Verdict detail={words.text("waiting.late_detail")}>
									{words.text("waiting.late")}
								</Verdict>
							)}
						</div>
					</div>
				</div>
			</article>
			{late && (
				<div class="mt-foot">
					<div class="mt-btns">
						<Button
							key="again"
							variant="primary"
							onClick={asking.ask}
							buttonRef={askAgain}
						>
							{words.text("waiting.ask_again")}
						</Button>
					</div>
				</div>
			)}
		</>
	);
}

// NoTaskScreen is a card no task is coming to: the model's last attempt at one
// failed its checks, or the day has no room for another. It says what happened
// and what comes next — a new task below, which the model is told to ask for
// at once, or more tomorrow. No course of a task runs on it, since it cannot
// see the new one being written, and there is nothing to press.
function NoTaskScreen({
	grade,
	wide,
	why,
}: {
	grade: number | undefined;
	wide: boolean;
	why: "exhausted" | "limit";
}) {
	const words = useWords();
	return (
		<article aria-label={words.text("waiting.label")}>
			<CardHeader grade={grade} wide={wide} />
			<div class="mt-body">
				{why === "exhausted" ? (
					<Verdict detail={words.text("waiting.exhausted_detail")}>
						{words.text("waiting.exhausted")}
					</Verdict>
				) : (
					<Verdict detail={words.text("waiting.limit_detail")}>
						{words.text("waiting.limit")}
					</Verdict>
				)}
			</div>
		</article>
	);
}

// useSecondsWaited is how long the round has lasted, as of the last moment of
// it reached: the card is drawn anew at each moment and no oftener. A new
// round starts from nothing, and never shows a moment of the one before.
function useSecondsWaited(round: number): number {
	const [reached, setReached] = useState({ round, seconds: 0 });
	useEffect(() => {
		const timers = Object.values(moments).map((seconds) =>
			setTimeout(() => setReached({ round, seconds }), seconds * 1000),
		);
		return () => {
			for (const timer of timers) {
				clearTimeout(timer);
			}
		};
	}, [round]);
	return reached.round === round ? reached.seconds : 0;
}

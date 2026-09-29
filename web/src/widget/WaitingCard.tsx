import { useEffect, useReducer, useRef, useState } from "preact/hooks";
import { GeneratingSteps, Note, Verdict } from "../design/blocks";
import { Button } from "../design/controls";
import type { Host } from "./bridge";
import { CardFrame, CardHeader } from "./CardFrame";
import { useFocusKeptOnTheCard } from "./focus";
import { LessonButtons, QuestionField } from "./LessonFoot";
import type { Waiting } from "./payload";
import {
	isLate,
	moments,
	stepsAt,
	type Wait,
	waitAfter,
	waitStart,
	warmUpAt,
} from "./waiting";
import { useWords } from "./words";

/**
 * WaitingCard is a card drawn while the next task is written — after the model
 * handed one in that failed its checks, its last attempt included — or once
 * the day has no room for another.
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
				waiting.kind === "limit" ? (
					<LimitScreen grade={waiting.child?.grade} wide={wide} />
				) : (
					<WaitingScreen
						grade={waiting.child.grade}
						exhausted={waiting.kind === "exhausted"}
						wide={wide}
						asking={asking}
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
export type Asking = { wait: Wait; ask: () => void };

/**
 * useAsking is a card's asking for the next task. Each ask sends the label of
 * "Another task", in the card's language, to the chat as the child's message —
 * only the model can write a task, and those are the words it takes as the
 * ask — and starts a round of the wait afresh; one the chat does not take
 * makes its round late at once.
 */
export function useAsking(host: Host): Asking {
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
 * cannot see the work; a warm-up once the wait drags on; and — once it has
 * lasted well past the time a task takes, or the ask never reached the chat —
 * the plain word that no task is coming unless a new one is below, and a way
 * to ask again. A card whose task failed its checks on the model's last
 * attempt says so above the course. A screen reader hears each change once,
 * as it comes: the steps are heard from their own list, and the warm-up and
 * the word that no task is coming from a place beside it, on the page from
 * the start.
 */
export function WaitingScreen({
	grade,
	exhausted = false,
	wide,
	asking,
}: {
	grade: number;
	exhausted?: boolean;
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
	// The failure told is the one the card was drawn for: an ask made from the
	// card starts a wait of its own.
	const failed = exhausted && asking.wait.round === 0;

	return (
		<>
			<article aria-label={words.text("waiting.label")}>
				<CardHeader grade={grade} wide={wide} />
				<div class="mt-body">
					<div class="mt-wait">
						{!late && failed && (
							<Verdict>{words.text("waiting.exhausted")}</Verdict>
						)}
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
							{late ? (
								<Verdict detail={words.text("waiting.late_detail")}>
									{words.text("waiting.late")}
								</Verdict>
							) : (
								seconds >= moments.warmUp && <WarmUp />
							)}
						</div>
					</div>
				</div>
			</article>
			<div class="mt-foot">
				<QuestionField off />
				<div class="mt-btns">
					{late ? (
						<Button
							key="again"
							variant="primary"
							onClick={asking.ask}
							buttonRef={askAgain}
						>
							{words.text("waiting.ask_again")}
						</Button>
					) : (
						<LessonButtons off />
					)}
				</div>
			</div>
		</>
	);
}

// LimitScreen is a card refused for the day: what happened, when it clears,
// and what can be done meanwhile. No task is coming, so there is nothing to
// ask about and nothing to press.
function LimitScreen({
	grade,
	wide,
}: {
	grade: number | undefined;
	wide: boolean;
}) {
	const words = useWords();
	return (
		<article aria-label={words.text("waiting.label")}>
			<CardHeader grade={grade} wide={wide} />
			<div class="mt-body">
				<Verdict detail={words.text("waiting.limit_detail")}>
					{words.text("waiting.limit")}
				</Verdict>
			</div>
		</article>
	);
}

// WarmUp is something for the child to do while the task is written, one at a
// time; the button brings the next. The first is chosen by the second the
// warm-up is offered at, so that a child who waits often does not meet the
// same one first each time.
function WarmUp() {
	const words = useWords();
	const [turns, setTurns] = useState(() => Math.floor(Date.now() / 1000));
	return (
		<div class="mt-warmup">
			<Note label={words.text("waiting.warmup_label")}>
				{words.text(warmUpAt(turns))}
			</Note>
			<div class="mt-btns">
				<Button onClick={() => setTurns((turned) => turned + 1)}>
					{words.text("waiting.warmup_next")}
				</Button>
			</div>
		</div>
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

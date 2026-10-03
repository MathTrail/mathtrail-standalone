import { useEffect, useState } from "preact/hooks";
import { GeneratingSteps, Verdict } from "../design/blocks";
import type { CallStage } from "./bridge";
import { CardHeader, CardRoot } from "./CardRoot";
import { type CourseStep, moments, stepsOf } from "./waiting";
import { type Key, useWords } from "./words";

/**
 * ComingCard is the card of a task being handed in, drawn before its result:
 * the usual course of a task, each step taken up as the host tells of the
 * call — the task written while its arguments come, checked once they have
 * come whole — and, when the call is cancelled, the word that the task was not
 * finished, or, when no result has come long after the call last moved on, the
 * word that none is coming. It shows nothing for its first moment, so that a
 * result that comes with the call does not make a wait flash past, and it has
 * nothing to press. Before its result it knows no child and no lesson: it has
 * no line at its top and no grade, and it speaks the host's language. A screen
 * reader hears each step from the list, and the word that ends the wait from a
 * place beside it, on the page from the start.
 */
export function ComingCard({ stage }: { stage: CallStage }) {
	const words = useWords();
	const shown = useShown();
	const late = useLate(stage);
	if (!shown) {
		return null;
	}
	const wait = waitOf(stage, late);
	return (
		<CardRoot>
			{(wide) => (
				<article aria-label={words.text("waiting.label")}>
					<CardHeader grade={undefined} wide={wide} />
					<div class="mt-body">
						<div class="mt-wait">
							{wait.steps !== undefined && (
								<GeneratingSteps
									title={words.text("waiting.title")}
									steps={wait.steps.map(({ key, status }) => ({
										label: words.text(key),
										status,
									}))}
									statusLabels={{
										done: words.text("waiting.done"),
										active: words.text("waiting.active"),
										waiting: words.text("waiting.waiting"),
									}}
								/>
							)}
							<div class="mt-news" aria-live="polite">
								{wait.over !== undefined && (
									<Verdict detail={words.text("waiting.ask_in_chat")}>
										{words.text(wait.over)}
									</Verdict>
								)}
							</div>
						</div>
					</div>
				</article>
			)}
		</CardRoot>
	);
}

// Wait is what the card shows of the wait: the steps of the task as far as its
// call has got, or, once the wait is over, the word that ends it.
type Wait =
	| { steps: CourseStep[]; over?: undefined }
	| { over: Key; steps?: undefined };

// waitOf is the wait as the call stands: over once the call is cancelled, or
// once it is late, and its steps otherwise.
function waitOf(stage: CallStage, late: boolean): Wait {
	if (stage === "cancelled") {
		return { over: "waiting.cancelled" };
	}
	if (late) {
		return { over: "waiting.late" };
	}
	return { steps: stepsOf(stage) };
}

// useShown says whether the card has been up long enough to show its wait.
function useShown(): boolean {
	const [shown, setShown] = useState(false);
	useEffect(() => {
		const timer = setTimeout(() => setShown(true), moments.shown);
		return () => clearTimeout(timer);
	}, []);
	return shown;
}

// useLate says whether no result has come long after the call last moved on:
// the clock starts with the card, and again once the task has come whole.
function useLate(stage: CallStage): boolean {
	const [late, setLate] = useState({ stage, late: false });
	useEffect(() => {
		const timer = setTimeout(
			() => setLate({ stage, late: true }),
			moments.givenUp,
		);
		return () => clearTimeout(timer);
	}, [stage]);
	return late.stage === stage && late.late;
}

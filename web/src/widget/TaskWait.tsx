import { GeneratingSteps, Verdict } from "../design/blocks";
import { type Phase, stepsOf } from "./waiting";
import { type Key, useWords } from "./words";

/**
 * Ending is what a card says when its wait is over with no task: what
 * happened, and what can be done.
 */
export type Ending = { said: Key; next: Key };

/**
 * TaskWait is the wait for a task as a card shows it: the usual course of a
 * task as far as the card knows it — the steps it saw taken, and, once the
 * task has come, the checks it passed and the task ready —, and the news of
 * the wait — a try the checks turned down, with a new one being written, or
 * the wait gone long — or, once the wait is over with no task, the word that
 * ends it in the course's place. Nothing on it moves but the step under way,
 * when one is, and there is nothing to press. A screen reader hears each step
 * from the list, and the news from a place beside it, on the page from the
 * start, so that each is heard once.
 */
export function TaskWait({
	phase,
	rewriting = false,
	slow = false,
	ended,
}: {
	phase: Phase;
	rewriting?: boolean;
	slow?: boolean;
	ended?: Ending;
}) {
	const words = useWords();
	return (
		<div class="mt-body">
			<div class="mt-wait">
				{ended === undefined && (
					<GeneratingSteps
						title={words.text("waiting.title")}
						steps={stepsOf(phase).map(({ key, status }) => ({
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
					{ended !== undefined ? (
						<Verdict detail={words.text(ended.next)}>
							{words.text(ended.said)}
						</Verdict>
					) : (
						<>
							{rewriting && (
								<Verdict>{words.text("waiting.rewriting")}</Verdict>
							)}
							{slow && (
								<Verdict detail={words.text("waiting.slow_detail")}>
									{words.text("waiting.slow")}
								</Verdict>
							)}
						</>
					)}
				</div>
			</div>
		</div>
	);
}

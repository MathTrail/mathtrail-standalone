import { useMemo } from "preact/hooks";
import {
	Note,
	SolutionSteps,
	Verdict,
	type VerdictTone,
} from "../design/blocks";
import type { Said } from "../design/controls";
import { ReplyCard } from "../design/thread";
import { dontKnow, type Letter } from "./choices";
import type { AnswerResult, HandedTask } from "./payload";
import { stepsOf } from "./steps";
import { type Key, ratingText, useWords } from "./words";

/**
 * TaskResult is MathTrail's reply once an answer is recorded, told below the
 * task: how the answer went, the trap behind a wrong option, the solution step
 * by step, and how the rating in the topic moved — or, while the trial series
 * runs, how far it has got. "I don't know" is told the solution with no
 * verdict, and the rating all the same, since it counts as a wrong answer.
 * Everything comes from what the service recorded, which is the answer a
 * second tab or an earlier press gave when it was not this card's.
 */
export function TaskResult({
	result,
	task,
	inTask,
}: {
	result: AnswerResult;
	task: HandedTask["task"];
	inTask: Said;
}) {
	const words = useWords();
	const optionText = (letter: Letter) => task.options[letter];
	const [tone, verdict] = verdictOf(result, optionText);
	// The card redraws as an ask for another task goes to the chat; the steps
	// are cut once for each solution.
	const steps = useMemo(
		() => stepsOf(result.solution, task.language),
		[result.solution, task.language],
	);
	return (
		<ReplyCard name={words.text("app.name")}>
			{result.already_answered && <p>{words.text("result.told_again")}</p>}
			<Verdict tone={tone}>{words.text(verdict.key, verdict.slots)}</Verdict>
			{result.trap !== null && (
				<Note
					tone="trap"
					label={words.text("result.trap")}
					said={inTask}
					detail={
						result.trap.repeated ? words.text("result.trap_again") : undefined
					}
				>
					{result.trap.text}
				</Note>
			)}
			<SolutionSteps
				label={words.text("result.solution")}
				steps={steps}
				said={inTask}
			/>
			{result.trial !== null ? (
				<Note tone="plain" label={words.text("result.trial")}>
					{words.text("result.trial_progress", {
						answered: result.trial.answered,
						total: result.trial.of,
					})}
				</Note>
			) : (
				result.rating !== null && (
					<Note tone="plain" label={words.text("result.rating")}>
						{words.text("result.rating_change", {
							before: ratingText(words, result.rating.before),
							after: ratingText(words, result.rating.after),
						})}
					</Note>
				)
			)}
		</ReplyCard>
	);
}

// verdictOf is how the verdict line reads for result: its tone, and the words
// that say it, the options named by their texts.
function verdictOf(
	result: AnswerResult,
	optionText: (letter: Letter) => string,
): [VerdictTone, { key: Key; slots?: Record<string, string> }] {
	if (result.choice === dontKnow) {
		return ["none", { key: "result.idk" }];
	}
	const correct = optionText(result.correct_answer);
	if (result.correct) {
		return ["correct", { key: "result.right", slots: { correct } }];
	}
	return [
		"wrong",
		{
			key: "result.wrong",
			slots: { correct, picked: optionText(result.choice) },
		},
	];
}

import type { AnswerResult, HandedTask, ResultShown } from "./payload";

/**
 * resultShownOf is how result went for the task handed, as a card shows it,
 * made from what the card of the task holds: whose card it is, the task's
 * options and topic, and the choice of the topic the task offered. It stands
 * for what the service hands over whole wherever only the result is there to
 * go on — on a page that answers for the service, and on a card whose service
 * told an answer's result alone. It names no site, so it opens no topic's page.
 */
export function resultShownOf(
	handed: HandedTask,
	result: AnswerResult,
): ResultShown {
	const { task } = handed;
	return {
		screen: "result",
		child: handed.child,
		task: {
			id: result.task_id,
			topic: task.topic,
			language: task.language,
			options: task.options,
		},
		result,
		topic_choice: handed.topic_choice,
		site: undefined,
		language: handed.language,
	};
}

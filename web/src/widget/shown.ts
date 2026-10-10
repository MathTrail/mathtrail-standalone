import type { AnswerResult, HandedTask, ResultShown } from "./payload";

/**
 * resultShownOf is the card of how result went for the task handed, as the
 * service hands it once asked, for a page that answers for the service: whose
 * card it is, the task's options and topic, and the choice of the topic the
 * task offered. Nothing on such a page opens a topic's page, so it names no
 * site.
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

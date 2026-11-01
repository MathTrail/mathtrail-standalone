import type { AnswerResult, HandedTask, ResultShown } from "./payload";

/**
 * resultShownOf is how result went for the task handed, as a card shows it,
 * made from what the card of the task holds: whose card it is, the task's
 * question, picture, options and topic, and the choice of the topic the task
 * offered. It stands for what the service hands over whole wherever only the
 * result is there to go on — on a page that answers for the service, and on a
 * card whose service told an answer's result alone. It names no site, so it
 * opens no topic's page.
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
			question: task.question,
			picture: task.picture,
			options: task.options,
		},
		result,
		topic_choice: handed.topic_choice,
		site: undefined,
		language: handed.language,
	};
}

/**
 * withTheTaskOf is how an answer to the task handed went, as shown tells it,
 * with the task's question and picture as the card of the task holds them
 * where shown has no question: a service before the task travelled with the
 * result told the rest whole, and the card keeps the task it showed.
 */
export function withTheTaskOf(
	shown: ResultShown,
	handed: HandedTask,
): ResultShown {
	if (shown.task.question !== undefined) {
		return shown;
	}
	return {
		...shown,
		task: {
			...shown.task,
			question: handed.task.question,
			picture: handed.task.picture,
		},
	};
}

import type { Host } from "../widget/bridge";
import type { AnswerResult, HandedTask } from "../widget/payload";
import type { Service } from "../widget/service";
import type { DemoData } from "./data";

/**
 * checkingTakes is how long the demo takes to record an answer, in
 * milliseconds: about as long as a chat takes, so that the card is seen
 * checking it.
 */
export const checkingTakes = 1100;

/**
 * demoService answers for MathTrail's service on the home page, from the
 * page's own data, and calls nobody. An answer is recorded after a moment as
 * the service would record it: the result the data holds for the choice, with
 * the hint as the card says it was used, and the rating moved on from where
 * the answer before left it, as far as the data moves it. The result alone is
 * told, and the card makes how it went from the task it holds. A task asked
 * for is still being written the first time its card asks, and handed out the
 * next time: the lesson's task again, under an id of its own. The page keeps
 * no profile, so no progress is read and no change to one is ever saved.
 */
export function demoService(data: DemoData): Service {
	const asked = new Map<string, number>();
	let rating: number | undefined;
	// movedOn is the rating move of an answer, taken on from where the answer
	// before left the rating.
	const movedOn = (move: AnswerResult["rating"]): AnswerResult["rating"] => {
		if (move === null) {
			return null;
		}
		const before = rating ?? move.before;
		rating = before + move.after - move.before;
		return { before, after: rating };
	};
	return {
		recordAnswer: (taskId, choice, hintUsed) =>
			new Promise((resolve) => {
				setTimeout(() => {
					const told = data.results[choice];
					resolve({
						kind: "answered",
						result: {
							...told,
							task_id: taskId,
							hint_used: hintUsed,
							rating: movedOn(told.rating),
						},
					});
				}, checkingTakes);
			}),
		taskStatus: (requestId) => {
			const times = (asked.get(requestId) ?? 0) + 1;
			asked.set(requestId, times);
			return Promise.resolve(
				times === 1
					? { kind: "writing", refused: 0 }
					: { kind: "task", handed: handedFor(data.handed, requestId) },
			);
		},
		readProgress: () => Promise.resolve({ kind: "failed" }),
		saveEdit: () => Promise.resolve({ kind: "failed" }),
	};
}

// handedFor is the lesson's task handed out again for the request requestId,
// under an id of its own, as every task the service hands out has.
function handedFor(handed: HandedTask, requestId: string): HandedTask {
	return { ...handed, task: { ...handed.task, id: `${requestId}_task` } };
}

/**
 * demoHost is the chat around the card on the home page. A message the card
 * puts in the chat is handed to sent, which shows it as the parent's and
 * brings what it asks for, the next task. A line for the model goes nowhere,
 * since no model reads the page, and no tool is called and no page opened: the
 * card on the page asks the service through the page, and shows no link.
 */
export function demoHost(sent: (text: string) => void): Host {
	return {
		callTool: (name) =>
			Promise.reject(new Error(`demo: the page calls no tool, not ${name}`)),
		sendMessage: (text) => {
			sent(text);
			return Promise.resolve();
		},
		tellModel: () => Promise.resolve(),
		canOpenLinks: () => false,
		openLink: () => Promise.resolve(false),
	};
}

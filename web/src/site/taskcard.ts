import { type Letter, letters } from "../widget/choices";
import {
	type AnswerResult,
	type HandedTask,
	readAnswer,
	readHandedTask,
	type TopicChoice,
} from "../widget/payload";
import { type CatalogTopic, gradesOf } from "./topics";
import type { CatalogTrap } from "./traps";

/**
 * CardFacts are a card of a task the site draws, as far as no language changes
 * it: the topic and the grade it is set in, its drawing, its five options, the
 * option a child picks on it — a wrong one — and the right one, the catalog's
 * trap behind the option picked, and the rating in the topic before and after
 * the answer.
 */
export type CardFacts = {
	readonly topic: string;
	readonly grade: number;
	readonly drawing: string;
	readonly options: Readonly<Record<string, string>>;
	readonly choice: string;
	readonly correct: string;
	readonly trap: string;
	readonly rating: { readonly before: number; readonly after: number };
};

/**
 * TaskWords are the words of a card's task in the page's language: whose card
 * it is, the task, and its hint.
 */
export type TaskWords = {
	readonly language: string;
	readonly child: string;
	readonly question: string;
	readonly hint: string;
};

/**
 * CardWords are the words of a card whose answer is in: whose card it is, the
 * task, the trap behind the option picked, and the solution.
 */
export type CardWords = {
	readonly language: string;
	readonly child: string;
	readonly question: string;
	readonly trap: string;
	readonly solution: string;
};

/**
 * WrongAnswer is a card as the widget reads one once its answer is in: the
 * task handed to the child, and the wrong answer recorded for it.
 */
export type WrongAnswer = {
	readonly handed: HandedTask;
	readonly result: AnswerResult;
};

/**
 * CardCatalog is what of the service's catalog a card is held to: its topics
 * and its traps.
 */
export type CardCatalog = {
	readonly topics: readonly CatalogTopic[];
	readonly traps: readonly CatalogTrap[];
};

/**
 * checkCard refuses a card the catalog would not have set or the widget could
 * not draw: of a topic or a trap the catalog does not have, in a grade the
 * topic is not taught in, or one the widget's own readers of a task and of an
 * answer refuse. It refuses a right answer too, and "I don't know", which
 * names no trap: neither is the card a page talks about. where names the card
 * in what it says, as "the card on the page Why".
 */
export function checkCard(
	catalog: CardCatalog,
	card: CardFacts,
	where: string,
): void {
	if (!letters.includes(card.choice as Letter)) {
		throw new Error(
			`${where} picks ${card.choice}, and the page shows an option picked: a letter A to E`,
		);
	}
	if (card.choice === card.correct) {
		throw new Error(
			`${where} answers right, and the page shows a wrong answer`,
		);
	}
	const topic = catalog.topics.find(({ id }) => id === card.topic);
	if (topic === undefined) {
		throw new Error(
			`${where} is of ${card.topic}, a topic the catalog does not have`,
		);
	}
	const [first, last] = gradesOf(topic.grade_levels);
	if (card.grade < first || card.grade > last) {
		throw new Error(
			`${where} is set in grade ${card.grade}, which ${card.topic} is not taught in`,
		);
	}
	if (!catalog.traps.some(({ id }) => id === card.trap)) {
		throw new Error(
			`${where} names the trap ${card.trap}, which the catalog does not have`,
		);
	}
	answerOf(card, anyWords, checkedTask, where);
}

// anyWords are words any card can be read with while it is checked, and
// checkedTask the id of its task then: what the widget's readers refuse is in
// the card's facts, and the page's words come with the page.
const anyWords: CardWords = {
	language: "en",
	child: "Comet",
	question: "?",
	trap: "?",
	solution: "?",
};
const checkedTask = "site_card";

/**
 * handedOf is card's task with the words said, as the widget reads a task
 * handed out, under the task's id: no service handed it out, and nothing shows
 * the id. A task handed out with offered has the button of the topic, as a
 * task after the trial series has. A task the widget's own reader refuses
 * stops the build.
 */
export function handedOf(
	card: CardFacts,
	said: TaskWords,
	id: string,
	where: string,
	offered?: TopicChoice,
): HandedTask {
	const handed = readHandedTask({
		screen: "task",
		child: { pseudonym: said.child, grade: card.grade, ui_language: null },
		task: {
			id,
			topic: card.topic,
			language: said.language,
			question: said.question,
			drawing: card.drawing,
			options: card.options,
			hint: said.hint,
		},
		language: said.language,
		topic_choice: offered,
	});
	if (handed === undefined) {
		throw new Error(
			`${where} is no task the widget can draw: it needs the five options A to E`,
		);
	}
	return handed;
}

/**
 * answerOf is card with the words said, as the widget reads one: the task
 * handed to the child, with the button of the topic when offered, and the
 * wrong answer recorded for it. The task has no hint, which a card no longer
 * shows once its answer is in. A card the widget's own readers refuse stops
 * the build.
 */
export function answerOf(
	card: CardFacts,
	said: CardWords,
	id: string,
	where: string,
	offered?: TopicChoice,
): WrongAnswer {
	const handed = handedOf(card, { ...said, hint: "" }, id, where, offered);
	const result = recordedOf(
		{
			task_id: id,
			topic: card.topic,
			choice: card.choice,
			correct: false,
			correct_answer: card.correct,
			trap: { id: card.trap, text: said.trap, repeated: false },
			solution: said.solution,
			hint_used: false,
			rating: card.rating,
			trial: null,
			already_answered: false,
		},
		where,
	);
	return { handed, result };
}

/**
 * recordedOf is result as the widget reads an answer the service recorded for
 * the task it names. A result the widget's own reader refuses stops the
 * build; where names the card in what it says.
 */
export function recordedOf(
	result: { readonly task_id: string } & Readonly<Record<string, unknown>>,
	where: string,
): AnswerResult {
	const told = readAnswer(
		{ content: [], structuredContent: { screen: "result", result } },
		result.task_id,
	);
	if (told.kind !== "answered") {
		throw new Error(
			`${where} holds no answer the widget can draw: its choice and its right option are letters A to E`,
		);
	}
	return told.result;
}

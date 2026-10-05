import type { Dictionary } from "../i18n/words";
import { type Choice, dontKnow, letters } from "../widget/choices";
import type { AnswerResult, HandedTask } from "../widget/payload";

/**
 * DemoData is what the home page hands its demo, written into the page when
 * the site is built: the language the page is in, the widget's words in that
 * language alone, the lesson's task as a card is handed it, and what the
 * service would record of every answer the card can be given.
 */
export type DemoData = {
	readonly locale: string;
	readonly words: Dictionary;
	readonly handed: HandedTask;
	readonly results: Readonly<Record<Choice, AnswerResult>>;
};

/**
 * DemoDataScript is the element the home page carries the demo's data in: a
 * script of JSON, which a browser never runs, each "<" in it written as its
 * escape, so that no word of the data can end the element or open a comment.
 */
export function DemoDataScript({ data }: { data: DemoData }) {
	return (
		<script
			type="application/json"
			data-demo=""
			dangerouslySetInnerHTML={{
				__html: JSON.stringify(data).replace(/</g, "\\u003c"),
			}}
		/>
	);
}

/**
 * readDemoData is the demo's data the page carries, or undefined when the page
 * carries none or none of the shape the demo knows: a page and a script from
 * two builds — one of them still in a cache — leave the page as it was built,
 * with nothing on it alive. The data is the site's own, checked when it was
 * built, so only its shape is looked at here.
 */
export function readDemoData(document: Document): DemoData | undefined {
	const element = document.querySelector("script[data-demo]");
	if (element === null) {
		return undefined;
	}
	let data: unknown;
	try {
		data = JSON.parse(element.textContent ?? "");
	} catch {
		return undefined;
	}
	return isDemoData(data) ? data : undefined;
}

// choices are every answer the card on the first screen can be given.
const choices: readonly Choice[] = [...letters, dontKnow];

// isDemoData says whether data has the shape of the demo's data: a language,
// words, a task handed out with its id, and a result for every answer.
function isDemoData(data: unknown): data is DemoData {
	if (!isObject(data)) {
		return false;
	}
	const { locale, words, handed, results } = data;
	return (
		typeof locale === "string" &&
		isObject(words) &&
		isObject(handed) &&
		isObject(handed.task) &&
		typeof handed.task.id === "string" &&
		isObject(results) &&
		choices.every((choice) => isObject(results[choice]))
	);
}

// isObject says whether value is an object whose fields can be read.
function isObject(value: unknown): value is Record<string, unknown> {
	return typeof value === "object" && value !== null;
}

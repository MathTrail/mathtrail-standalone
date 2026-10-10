import { useContext, useMemo, useRef } from "preact/hooks";
import {
	Note,
	SolutionSteps,
	Verdict,
	type VerdictTone,
} from "../design/blocks";
import { classes } from "../design/classes";
import { Button, type Said } from "../design/controls";
import { PageLink } from "../design/links";
import { Diagram } from "../design/picture/diagram";
import { directionOf } from "../i18n/lookup";
import type { Words } from "../i18n/words";
import type { Host } from "./bridge";
import { CardFrame } from "./CardFrame";
import { CardHeader } from "./CardRoot";
import { RequestNote, useChatRequest, useModelLines } from "./ChatRequest";
import { dontKnow, type Letter } from "./choices";
import { useFocusKeptOnTheCard } from "./focus";
import { useLinking } from "./linking";
import { pageAddress } from "./links";
import { topicName } from "./names";
import type { AnswerResult, ResultShown, Shown } from "./payload";
import { stepsOf } from "./steps";
import {
	ChoosesTopic,
	TopicButton,
	TopicNote,
	TopicPanel,
	useTopicChoice,
} from "./TopicChoice";
import { type Key, ratingText, useWords } from "./words";

/**
 * ResultCard is the card of how an answer went, drawn below the task's card
 * when the model shows how an answer given in the chat went: whether it was
 * right, the trap behind a wrong option, the solution step by step, the topic
 * the task was on, linked to its page where the chat opens links, and how the
 * rating in the topic moved — or how far the trial series has got. Under it
 * are the next task to ask for and the topic to choose, as the task's card
 * had them before the answer; once the chat has the ask, this card is done
 * with, as a task's card is. A card with nothing to show says why: the task
 * has no answer yet, or is no longer on the card.
 */
export function ResultCard({ shown, host }: { shown: Shown; host: Host }) {
	const words = useWords();
	const child = shown.kind === "shown" ? shown.result.child : shown.child;
	return (
		<CardFrame
			child={child}
			back={words.text("progress.back_plain")}
			host={host}
		>
			{(wide, whose) =>
				shown.kind === "shown" ? (
					<ResultInCard
						shown={shown.result}
						host={host}
						wide={wide}
						grade={whose.grade}
					/>
				) : (
					<article aria-label={words.text("result.label")}>
						<CardHeader grade={whose.grade} wide={wide} />
						<div class="mt-body">
							<Verdict>{words.text(`result.${shown.kind}`)}</Verdict>
						</div>
					</article>
				)
			}
		</CardFrame>
	);
}

// ResultInCard is how the answer went, inside the frame of its card, under the
// card's header with the grade given, and what the card offers next.
function ResultInCard({
	shown,
	host,
	wide,
	grade,
}: {
	shown: ResultShown;
	host: Host;
	wide: boolean;
	grade: number;
}) {
	const words = useWords();
	const tell = useModelLines(host);
	return (
		<>
			<article aria-label={words.text("result.label")}>
				<CardHeader grade={grade} wide={wide} />
				<ResultBody shown={shown} host={host} />
			</article>
			<div class="mt-foot">
				<ResultFoot
					shown={shown}
					host={host}
					wide={wide}
					grade={grade}
					tell={tell}
					answeredHere={false}
				/>
			</div>
		</>
	);
}

/**
 * ResultBody is how an answer went, as a card shows it under its header: the
 * verdict, the trap behind a wrong option, the picture of the solution with
 * its total, the solution step by step, and the topic the task was on with how
 * the rating in it moved — or how far the trial series has got.
 */
export function ResultBody({
	shown,
	host,
}: {
	shown: ResultShown;
	host: Host;
}) {
	const words = useWords();
	const { task, result } = shown;
	// The texts of the task are in the language it was written in, which need
	// not be the card's: a screen reader reads them in their own voice, and
	// they run their own way.
	const inTask: Said = { lang: task.language, dir: directionOf(task.language) };
	const [tone, verdict] = verdictOf(result, (letter) => task.options[letter]);
	// The card redraws as an ask goes to the chat; the steps are cut once for
	// each solution.
	const steps = useMemo(
		() => stepsOf(result.solution, task.language),
		[result.solution, task.language],
	);
	return (
		<div class="mt-body">
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
			<SolutionPicture result={result} task={task} />
			<SolutionSteps
				label={words.text("result.solution")}
				steps={steps}
				said={inTask}
			/>
			<Facts shown={shown} host={host} />
		</div>
	);
}

/**
 * ResultFoot is what a card offers under how an answer went: the next task to
 * ask for and the topic to choose, which ask the chat as a task's card does,
 * and once the chat has the ask, where the next task comes, the card done
 * with, the focus on it if the card has lost its own. answeredHere says the
 * answer was given on this very card, whose option pressed went with the task:
 * a focus lost with it goes to the next task to ask for. A card drawn for an
 * answer given elsewhere takes no focus as it is drawn. tell is how the card
 * tells the model a line, so that the line of a topic chosen here goes with
 * those the card told before it.
 */
export function ResultFoot({
	shown,
	host,
	wide,
	grade,
	tell,
	answeredHere,
}: {
	shown: ResultShown;
	host: Host;
	wide: boolean;
	grade: number;
	tell: (line: string) => Promise<void>;
	answeredHere: boolean;
}) {
	const words = useWords();
	const another = useChatRequest(host);
	const anotherButton = useRef<HTMLButtonElement>(null);
	const note = useRef<HTMLParagraphElement>(null);
	// A choice asks for a task, so it waits for an ask already on its way,
	// even one pressed in the same moment. Where a choice cannot be acted on,
	// none is ever made.
	const choosesTopic = useContext(ChoosesTopic);
	const choosing = useTopicChoice(shown.topic_choice, {
		busy: () => !choosesTopic || another.busy(),
		tell,
		ask: another.send,
	});
	const saving = choosing.said === "saving";
	const asking = another.state === "sending";
	const asked = another.state === "sent";
	useFocusKeptOnTheCard(answeredHere && !asked, anotherButton);
	useFocusKeptOnTheCard(asked, note);

	const offered = shown.topic_choice;
	const topicLocked = !choosesTopic || asking || saving;
	return (
		<>
			{!asked && (
				<>
					<div class="mt-btns">
						<Button
							variant="primary"
							locked={asking || saving}
							onClick={() => {
								if (!choosing.busy()) {
									another.send(words.text("task.another"));
								}
							}}
							buttonRef={anotherButton}
						>
							{words.text("task.another")}
						</Button>
						{offered !== undefined && (
							<TopicButton
								choosing={choosing}
								wide={wide}
								locked={topicLocked}
							/>
						)}
					</div>
					{offered !== undefined && (
						<>
							<TopicPanel
								choosing={choosing}
								offered={offered}
								grade={grade}
								host={host}
								locked={topicLocked}
							/>
							<TopicNote said={choosing.said} />
						</>
					)}
				</>
			)}
			<RequestNote
				state={another.state}
				taken="task.another_coming"
				noteRef={note}
			/>
		</>
	);
}

/**
 * verdictText is the verdict on result in the words given, the options named
 * by their texts: what a screen reader hears once a card turns into how the
 * answer went.
 */
export function verdictText(
	words: Words<Key>,
	result: AnswerResult,
	options: Readonly<Record<Letter, string>>,
): string {
	const [, verdict] = verdictOf(result, (letter) => options[letter]);
	return words.text(verdict.key, verdict.slots);
}

// SolutionPicture is the picture of the solution, laid out as the task's own
// picture is, with the equality the solution comes to written large under
// it — and, after a wrong option, the option it was not, in the words of the
// card's language.
function SolutionPicture({
	result,
	task,
}: {
	result: AnswerResult;
	task: ResultShown["task"];
}) {
	const words = useWords();
	const drawn = result.solution_picture;
	if (drawn === undefined) {
		return null;
	}
	const total = result.solution_total;
	const picked = pickedWrong(result, (letter) => task.options[letter]);
	return (
		<figure class="mt-solution-picture">
			<Diagram
				picture={drawn}
				label={words.text(`picture.${drawn.kind}`)}
				locale={task.language}
				keyLabel={words.text("picture.colors")}
				said={{ lang: task.language, dir: directionOf(task.language) }}
			/>
			{total !== undefined && (
				<figcaption class="mt-total">
					{picked === undefined ? (
						<TotalSum total={total} />
					) : (
						pieces(
							words.text("result.total_not", { total: firstMark, picked }),
						).map((piece) =>
							piece === firstMark ? (
								<TotalSum key="total" total={total} />
							) : (
								piece
							),
						)
					)}
				</figcaption>
			)}
		</figure>
	);
}

// TotalSum is the equality itself, which reads left to right in a card of any
// language, as arithmetic does.
function TotalSum({ total }: { total: string }) {
	return (
		<bdi class="mt-total-sum" dir="ltr">
			{total}
		</bdi>
	);
}

// pickedWrong is the text of the option a wrong answer chose, and nothing for
// an answer that was right or was "I don't know".
function pickedWrong(
	result: AnswerResult,
	optionText: (letter: Letter) => string,
): string | undefined {
	if (result.correct || result.choice === dontKnow) {
		return undefined;
	}
	return optionText(result.choice);
}

// Facts are the topic the task was on, which opens the topic's page through
// the chat where the chat opens links and is its name alone otherwise, and how
// the rating in the topic moved — or, while the trial series runs, how far it
// has got.
function Facts({ shown, host }: { shown: ResultShown; host: Host }) {
	const words = useWords();
	const { task, result } = shown;
	const linking = useLinking(host, words.text("progress.link_refused"));
	const name = topicName(words, task.topic);
	const href = pageAddress(shown.site, words.locale, task, "");
	return (
		<dl class="mt-facts">
			<div class="mt-fact">
				<dt>{words.text("result.topic")}</dt>
				<dd>
					{href !== undefined && linking !== undefined ? (
						<PageLink label={name} href={href} linking={linking} />
					) : (
						name
					)}
				</dd>
			</div>
			{result.trial !== null ? (
				<div class="mt-fact">
					<dt>{words.text("result.trial")}</dt>
					<dd>
						{words.text("result.trial_progress", {
							answered: result.trial.answered,
							total: result.trial.of,
						})}
					</dd>
				</div>
			) : (
				result.rating !== null && (
					<div class="mt-fact">
						<dt>{words.text("result.rating")}</dt>
						<dd>
							<RatingMove
								before={result.rating.before}
								after={result.rating.after}
							/>
						</dd>
					</div>
				)
			)}
		</dl>
	);
}

// A mark no wording holds stands in a wording for a value the card draws in a
// style of its own, while the wording is cut where the value is written: from
// the Private Use Area, so it is no letter of any language.
const firstMark = "\uE000";
const secondMark = "\uE001";

// RatingMove is how the rating moved, in the words the dictionary writes it
// with — the arrow pointing the way the language reads — the rating before
// muted and the rating after in the colour of the move.
function RatingMove({ before, after }: { before: number; after: number }) {
	const words = useWords();
	const line = words.text("result.rating_change", {
		before: firstMark,
		after: secondMark,
	});
	return (
		<span class="mt-rating-move">
			{pieces(line).map((piece) => {
				switch (piece) {
					case firstMark:
						return (
							<span key="before" class="mt-rating-before">
								{ratingText(words, before)}
							</span>
						);
					case secondMark:
						return (
							<span
								key="after"
								class={classes(
									"mt-rating-after",
									after > before && "mt-rating-gain",
									after < before && "mt-rating-loss",
								)}
							>
								{ratingText(words, after)}
							</span>
						);
					default:
						return piece;
				}
			})}
		</span>
	);
}

// pieces are a wording cut where its marks stand, each mark a piece of its
// own, with no empty piece where a mark opens or closes it.
function pieces(line: string): string[] {
	return line.split(/([\uE000\uE001])/).filter((piece) => piece !== "");
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

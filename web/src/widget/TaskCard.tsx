import type { ComponentChildren } from "preact";
import { useContext, useReducer, useRef } from "preact/hooks";
import { Note } from "../design/blocks";
import {
	type Option,
	OptionList,
	type OptionState,
	type Said,
} from "../design/controls";
import { Diagram } from "../design/picture/diagram";
import { directionOf } from "../i18n/lookup";
import type { Host } from "./bridge";
import { CardFrame } from "./CardFrame";
import { CardHeader } from "./CardRoot";
import { RequestNote, useChatRequest, useModelLines } from "./ChatRequest";
import { type Letter, letters } from "./choices";
import { useFocusKeptOnTheCard } from "./focus";
import { LessonButtons } from "./LessonFoot";
import {
	type Answer,
	canAnswer,
	isSettled,
	type Lesson,
	lessonStart,
	modelLineOf,
	next,
	optionStateOf,
} from "./lesson";
import type { HandedTask } from "./payload";
import { ResultBody, ResultFoot, verdictText } from "./ResultCard";
import { useService } from "./service";
import { resultShownOf } from "./shown";
import {
	ChoosesTopic,
	TopicButton,
	TopicChip,
	TopicNote,
	TopicPanel,
	useTopicChoice,
} from "./TopicChoice";
import { type Key, useWords } from "./words";

/**
 * TaskCard is the card a task is handed to the child on: the task, in the
 * frame every card of a lesson has, which opens the progress in the card and
 * comes back to the task as it was left. The lesson on it begins at start: the
 * task just handed out, with no answer given, unless start says how far it
 * has got.
 */
export function TaskCard({
	handed,
	host,
	start,
}: {
	handed: HandedTask;
	host: Host;
	start?: Lesson;
}) {
	const words = useWords();
	return (
		<CardFrame
			child={handed.child}
			back={words.text("progress.back")}
			host={host}
		>
			{(wide, whose) => (
				<TaskInCard
					handed={handed}
					host={host}
					wide={wide}
					grade={whose.grade}
					start={start}
				/>
			)}
		</CardFrame>
	);
}

/**
 * TaskInCard is a task inside the frame of the card it is on, under the
 * card's header with the grade given. The child answers by pressing an option,
 * which records the answer straight away: the card turns into how the answer
 * went, in the task's place and under the same header, tells the model in a
 * line, and asks the chat for nothing. Under how it went are the next task to
 * ask for and the topic to choose, as on the card of how an answer went. An
 * answer recorded before is shown as it was recorded, and says so. Before an
 * answer the child opens the hint, and asks for another task, which goes to
 * the chat for the model to write, the new task in a card of its own; once the
 * chat has that ask, the card is done with. A task handed out with the choice
 * of the topic offers it before an answer: a topic to keep the lessons to, or
 * the coach's choice, saved and then asked for as another task is. The lesson
 * begins at start, the task just handed out unless it says otherwise.
 */
export function TaskInCard({
	handed,
	host,
	wide,
	grade,
	start = lessonStart,
}: {
	handed: HandedTask;
	host: Host;
	wide: boolean;
	grade: number;
	start?: Lesson;
}) {
	const { task } = handed;
	const words = useWords();
	const [lesson, dispatch] = useReducer(next, start);
	const service = useService();
	const another = useChatRequest(host);
	const tell = useModelLines(host);
	const note = useRef<HTMLParagraphElement>(null);
	const said = useRef<HTMLParagraphElement>(null);
	// An answer is sent once. From the moment it is on its way the options are
	// locked, and a second press that comes before the card redraws to lock
	// them is turned away here.
	const answering = useRef(false);
	// A choice asks for a task, so it waits for an answer or an ask already on
	// its way, even one pressed in the same moment, before the card redraws.
	// Where a choice cannot be acted on, none is ever made.
	const choosesTopic = useContext(ChoosesTopic);
	const choosing = useTopicChoice(handed.topic_choice, {
		busy: () => !choosesTopic || answering.current || another.busy(),
		tell,
		ask: another.send,
	});
	// A choice of the topic on its way locks the card as an answer does: the
	// task it asks for would race the answer.
	const saving = choosing.said === "saving";
	// An ask for another task on its way locks the options as well: an answer
	// then would race the ask that sets the task aside. An ask the chat has
	// taken leaves the card done with; one it refuses gives the card back.
	const asking = another.state === "sending";
	const asked = another.state === "sent";

	async function answer(choice: Letter) {
		if (answering.current || choosing.busy() || another.busy()) {
			return;
		}
		answering.current = true;
		dispatch({ type: "picked", choice });
		const told = await service.recordAnswer(task.id, choice, lesson.hint.used);
		answering.current = false;
		dispatch({ type: "told", outcome: told });
		// The card shows how the answer went itself; the model reads the line
		// with the adult's next message, and nothing goes to the chat.
		if (told.kind === "answered") {
			tell(modelLineOf(told.result)).catch((error: unknown) => {
				console.error("widget: the model was not told of the answer", error);
			});
		}
	}

	function askForAnother() {
		// Not while an answer is on its way: the ask would race the answer to a
		// task the model is about to set aside. The label of the button, in the
		// card's language, goes to the chat as the adult's message: only the
		// model can write a task, and those are the words it takes as the ask.
		if (answering.current || choosing.busy()) {
			return;
		}
		another.send(words.text("task.another"));
	}

	// A focus lost with the button pressed goes to what the card says: why
	// the answer was not recorded, or where the next task comes. Once the card
	// shows how the answer went, what it offers next keeps the focus itself.
	useFocusKeptOnTheCard(
		lesson.answer.state === "closed" || lesson.answer.state === "failed",
		said,
	);
	useFocusKeptOnTheCard(asked, note);

	// The task is in the language it was written in, which need not be the
	// card's: a screen reader reads it in its own voice, and it runs its own way.
	const inTask: Said = { lang: task.language, dir: directionOf(task.language) };
	const offered = handed.topic_choice;
	// The choice of the topic asks for a task, so it waits as another task does;
	// where a choice cannot be acted on, its button is shown and stays locked.
	const topicLocked =
		!choosesTopic || lesson.answer.state === "checking" || asking || saving;
	// How the answer went, as the service told it — or, from a service that
	// told the result alone, as the task on the card names it.
	const went =
		lesson.answer.state === "answered"
			? (lesson.answer.shown ?? resultShownOf(handed, lesson.answer.result))
			: undefined;
	const saysOfTheAnswer = saidOf(lesson.answer);

	return (
		<>
			<article
				aria-label={words.text(
					went === undefined ? "task.label" : "result.label",
				)}
			>
				<CardHeader grade={grade} wide={wide} />
				{went === undefined ? (
					<TaskBody
						handed={handed}
						lesson={lesson}
						inTask={inTask}
						locked={!canAnswer(lesson) || saving || asking || asked}
						doneWith={asked}
						onAnswer={answer}
						chip={
							givenOnTheChoice(handed, choosing.chosen) &&
							!isSettled(lesson.answer) &&
							!asked && (
								<TopicChip
									choosing={choosing}
									topic={task.topic}
									locked={topicLocked}
								/>
							)
						}
					/>
				) : (
					<ResultBody shown={went} host={host} />
				)}
			</article>
			<div class="mt-foot">
				{went !== undefined && (
					<ResultFoot
						shown={went}
						host={host}
						wide={wide}
						grade={grade}
						tell={tell}
						answeredHere
					/>
				)}
				{went === undefined && !asked && !isSettled(lesson.answer) && (
					<>
						<div class="mt-btns">
							<LessonButtons
								locked={lesson.answer.state === "checking" || saving}
								anotherSending={asking || saving}
								hintOpen={lesson.hint.open}
								onHint={() => dispatch({ type: "hint toggled" })}
								onAnother={askForAnother}
							/>
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
				{went === undefined && (
					<RequestNote
						state={another.state}
						taken="task.another_coming"
						noteRef={note}
					/>
				)}
				<p
					class="mt-action-note mt-answer-note"
					aria-live="polite"
					ref={said}
					tabIndex={saysOfTheAnswer === undefined ? undefined : -1}
				>
					{saysOfTheAnswer === undefined ? "" : words.text(saysOfTheAnswer)}
				</p>
				<output class="mt-vh">
					{went === undefined
						? ""
						: verdictText(words, went.result, went.task.options)}
				</output>
			</div>
		</>
	);
}

// saidOf is what the card says of the answer, or nothing while it has nothing
// to say: why the answer was not recorded — the task is closed, or the answer
// could not be checked — or, of an answer recorded, that it was recorded
// before, which is shown as it was recorded then.
function saidOf(answer: Answer): Key | undefined {
	switch (answer.state) {
		case "closed":
			return "task.closed";
		case "failed":
			return "task.answer_failed";
		case "answered":
			return answer.result.already_answered
				? "task.answered_before"
				: undefined;
		default:
			return undefined;
	}
}

// givenOnTheChoice says whether the task was handed out on the topic the
// lessons are kept to, and the choice still stands: chosen is the choice in
// force on the card. A task the rule gave is not, even once its own topic is
// chosen on it.
function givenOnTheChoice(handed: HandedTask, chosen: string | null): boolean {
	const topic = handed.task.topic;
	return handed.topic_choice?.chosen === topic && chosen === topic;
}

// TaskBody is the task itself, until the card turns into how its answer
// went: the mark of the topic the lessons are kept to, when the task is on it,
// its question, its picture, the hint while it is shown and the task is not
// done with, and the options — to press unless locked. A card done with, its
// next task asked of the chat, no longer asks for a pick.
function TaskBody({
	handed,
	lesson,
	inTask,
	locked,
	doneWith,
	onAnswer,
	chip,
}: {
	handed: HandedTask;
	lesson: Lesson;
	inTask: Said;
	locked: boolean;
	doneWith: boolean;
	onAnswer: (choice: Letter) => void;
	chip: ComponentChildren;
}) {
	const words = useWords();
	const { task } = handed;
	return (
		<div class="mt-body">
			{chip}
			<p class="mt-task-text" lang={inTask.lang} dir={inTask.dir}>
				{task.question}
			</p>
			{task.picture !== undefined && (
				<Diagram
					picture={task.picture}
					label={words.text(`picture.${task.picture.kind}`)}
					locale={task.language}
					keyLabel={words.text("picture.colors")}
					said={inTask}
				/>
			)}
			{lesson.hint.open && !isSettled(lesson.answer) && (
				<Note tone="hint" label={words.text("task.hint")} said={inTask}>
					{task.hint}
				</Note>
			)}
			<OptionList
				legend={words.text(doneWith ? "result.answers" : "task.pick")}
				options={optionsOf(handed, lesson.answer, (key) => words.text(key))}
				locked={locked}
				onSelect={onAnswer}
				said={inTask}
			/>
		</div>
	);
}

// optionsOf are the task's five options as the card shows them, in the order
// of their letters, each with its state and the words its state is said in.
function optionsOf(
	handed: HandedTask,
	answer: Answer,
	text: (key: Key) => string,
): Option<Letter>[] {
	return letters.map((letter) => {
		const state = optionStateOf(answer, letter);
		const status = statusKeys[state];
		return {
			letter,
			value: handed.task.options[letter],
			state,
			status: status === undefined ? undefined : text(status),
		};
	});
}

// The words beside an option in each state that has any.
const statusKeys: Partial<Record<OptionState, Key>> = {
	selected: "task.checking",
};

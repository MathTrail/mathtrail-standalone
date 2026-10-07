import type { ComponentChildren, Ref } from "preact";
import { useReducer, useRef, useState } from "preact/hooks";
import { Diagram, Note, Verdict } from "../design/blocks";
import {
	Button,
	type Option,
	OptionList,
	type OptionState,
	type Said,
} from "../design/controls";
import { ReplyCard } from "../design/thread";
import { directionOf } from "../i18n/lookup";
import type { Host } from "./bridge";
import { CardFrame } from "./CardFrame";
import { CardHeader } from "./CardRoot";
import {
	lineWait,
	type Request,
	RequestNote,
	useChatRequest,
	within,
} from "./ChatRequest";
import { type Letter, letters } from "./choices";
import { useFocusKeptOnTheCard } from "./focus";
import type { Shown } from "./LessonCard";
import { LessonButtons } from "./LessonFoot";
import {
	type Answer,
	canAnswer,
	isOver,
	isSettled,
	type Lesson,
	lessonStart,
	modelLineOf,
	next,
	optionStateOf,
	recordedResult,
	takenLineOf,
} from "./lesson";
import type { HandedTask } from "./payload";
import { useService } from "./service";
import { TaskResult } from "./TaskResult";
import {
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
 * has got. onTaken shows in the card's place what the card took when the child
 * asked for another task on it; a card without it asks the chat instead.
 */
export function TaskCard({
	handed,
	host,
	start,
	onTaken,
}: {
	handed: HandedTask;
	host: Host;
	start?: Lesson;
	onTaken?: (shown: Shown) => void;
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
					onTaken={onTaken}
				/>
			)}
		</CardFrame>
	);
}

/**
 * TaskInCard is a task inside the frame of the card it is on, under the
 * card's header with the grade given. The child answers by pressing an option,
 * which records the answer straight away, and reads the result below the
 * task, in the same card; opens the hint; and asks for another task, which
 * the card takes itself: the one written ahead, in this card's place at once,
 * or the one being written, waited for here — the model told which, and the
 * chat asked to get the next one ready. Where the card can take none it goes
 * to the chat for the model to write, as it always did: the new task comes in
 * a card of its own, below, and this one says so and keeps its task.
 * A task handed out with the choice of the topic offers it too: a topic to
 * keep the lessons to, or the coach's choice, saved and then taken as another
 * task is. The lesson begins at start, the task just handed out unless it says
 * otherwise.
 */
export function TaskInCard({
	handed,
	host,
	wide,
	grade,
	start = lessonStart,
	onTaken,
}: {
	handed: HandedTask;
	host: Host;
	wide: boolean;
	grade: number;
	start?: Lesson;
	onTaken?: (shown: Shown) => void;
}) {
	const { task } = handed;
	const words = useWords();
	const [lesson, dispatch] = useReducer(next, start);
	const service = useService();
	const another = useChatRequest(host);
	const outcome = useRef<HTMLDivElement>(null);
	const nextTask = useRef<HTMLButtonElement>(null);
	const article = useRef<HTMLElement>(null);
	// taking is the next task asked of the service and not yet in the card's
	// place: it locks the card as a choice on its way does, and a second press
	// meanwhile, even one before the card has redrawn, is turned away.
	const [taking, setTaking] = useState(false);
	const takingNow = useRef(false);
	// dayOver is the day found to have no room for another task. The card
	// keeps its task, to be answered or gone back over, and says so.
	const [dayOver, setDayOver] = useState(false);
	// An answer is sent once. From the moment it is on its way the options are
	// locked, and a second press that comes before the card has redrawn to lock
	// them is turned away here.
	const answering = useRef(false);
	// held is the line the model was given on this card that no message has
	// carried to it yet. The host keeps one line and reads it with the next
	// message, so a later line carries it along, and a message lets it go.
	const held = useRef<string | undefined>(undefined);
	function tell(line: string): Promise<void> {
		held.current =
			held.current === undefined ? line : `${held.current}\n\n${line}`;
		return host.tellModel(held.current);
	}
	function ask(message: string) {
		held.current = undefined;
		another.send(message);
	}
	// A choice asks for a task, so it waits for an answer or an ask already on
	// its way, even one pressed in the same moment, before the card redraws;
	// once saved, the card takes the task on the topic chosen.
	const choosing = useTopicChoice(handed.topic_choice, {
		busy: () => answering.current || another.busy() || takingNow.current,
		tell,
		ask: (fallback) => void takeNext(fallback),
	});
	// A choice of the topic on its way locks the card as an answer does: the
	// task it asks for would race the answer. So does a task being taken.
	const saving = choosing.said === "saving" || taking;

	async function answer(choice: Letter) {
		if (answering.current || choosing.busy()) {
			return;
		}
		answering.current = true;
		dispatch({ type: "picked", choice });
		const told = await service.recordAnswer(task.id, choice, lesson.hint.used);
		answering.current = false;
		dispatch({ type: "told", outcome: told });
		if (told.kind === "answered") {
			tell(modelLineOf(told.result)).catch((error: unknown) => {
				console.error("widget: the model was not told of the answer", error);
			});
		}
	}

	function askForAnother() {
		// Not while an answer is on its way: the ask would race the answer to a
		// task about to be set aside. Where the card can take no task, the label
		// of the button, in the card's language, goes to the chat as the child's
		// message: those are the words the model takes as the ask.
		if (answering.current || choosing.busy()) {
			return;
		}
		void takeNext(words.text("task.another"));
	}

	// takeNext takes the next task for this card. One written ahead, or one
	// still being written, takes the card's place: the model is told first which
	// it is, in a line read with the child's message, and then the message asks
	// the chat to get the next one ready. A day with no room for another is said
	// under the task, which stays, and a task no longer the one being solved
	// shows itself closed, with what was recorded of it. A task the service could
	// not take — and any, where the card has no place to show one in — is asked
	// of the chat in fallback's words.
	async function takeNext(fallback: string) {
		if (takingNow.current || another.busy()) {
			return;
		}
		if (onTaken === undefined) {
			ask(fallback);
			return;
		}
		takingNow.current = true;
		setTaking(true);
		const taken = await service.takeTask(task.id);
		if (taken.kind === "task" || taken.kind === "coming") {
			// The card stays locked until what it took takes its place: a press
			// meanwhile would take again, and ask the chat a second time.
			await within(
				lineWait,
				tell(takenLineOf(taken)).catch((error: unknown) => {
					console.error(
						"widget: the model was not told of the task taken",
						error,
					);
				}),
			);
			ask(words.text("task.ready_next"));
			onTaken(
				taken.kind === "task"
					? { kind: "task", handed: taken.handed }
					: { kind: "coming", coming: taken.coming },
			);
			return;
		}
		takingNow.current = false;
		setTaking(false);
		switch (taken.kind) {
			case "limit":
				setDayOver(true);
				return;
			case "over":
				dispatch({ type: "closed" });
				return;
			case "failed":
				ask(fallback);
		}
	}

	// Once the answer is in, a focus that was lost goes to the one thing left
	// to do, while the replies read the result out; with nothing left to do,
	// to what the card says.
	useFocusKeptOnTheCard(isOver(lesson.answer), nextTask, outcome);
	// A task that took the place of another — the button pressed for it gone —
	// takes the focus, so that a screen reader reads it next.
	useFocusKeptOnTheCard(true, article);

	// The task is in the language it was written in, which need not be the
	// card's: a screen reader reads it in its own voice, and it runs its own way.
	const inTask: Said = { lang: task.language, dir: directionOf(task.language) };
	const offered = handed.topic_choice;
	// The choice of the topic asks for a task, so it waits as another task does.
	const topicLocked =
		lesson.answer.state === "checking" || another.state === "sending" || saving;

	return (
		<>
			<article
				aria-label={words.text("task.label")}
				ref={article}
				tabIndex={-1}
			>
				<CardHeader grade={grade} wide={wide} />
				<TaskBody
					handed={handed}
					lesson={lesson}
					inTask={inTask}
					locked={!canAnswer(lesson) || saving}
					onAnswer={answer}
					chip={
						givenOnTheChoice(handed, choosing.chosen) &&
						lesson.answer.state !== "closed" && (
							<TopicChip
								choosing={choosing}
								topic={task.topic}
								locked={topicLocked}
							/>
						)
					}
				/>
			</article>
			<section
				class="mt-replies"
				aria-label={words.text("task.replies")}
				aria-live="polite"
			>
				{isOver(lesson.answer) && (
					<div ref={outcome} tabIndex={-1}>
						<Outcome answer={lesson.answer} task={task} inTask={inTask} />
					</div>
				)}
				{dayOver && <DayOver />}
			</section>
			{lesson.answer.state !== "closed" && (
				<div class="mt-foot">
					<div class="mt-btns">
						<TaskActions
							lesson={lesson}
							another={another.state}
							saving={saving}
							nextTask={nextTask}
							onHint={() => dispatch({ type: "hint toggled" })}
							onAnother={askForAnother}
						/>
						{offered !== undefined && (
							<TopicButton choosing={choosing} locked={topicLocked} />
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
					<RequestNote state={another.state} taken="task.another_coming" />
				</div>
			)}
		</>
	);
}

// givenOnTheChoice says whether the task was handed out on the topic the
// lessons are kept to, and the choice still stands: chosen is the choice in
// force on the card. A task the rule gave is not, even once its own topic is
// chosen on it.
function givenOnTheChoice(handed: HandedTask, chosen: string | null): boolean {
	const topic = handed.task.topic;
	return handed.topic_choice?.chosen === topic && chosen === topic;
}

// TaskBody is the task itself: the mark of the topic the lessons are kept to,
// when the task is on it, its question, its drawing, the hint while it is
// shown and the task is not done with, and the options — to press unless
// locked, or marked once the answer is in.
function TaskBody({
	handed,
	lesson,
	inTask,
	locked,
	onAnswer,
	chip,
}: {
	handed: HandedTask;
	lesson: Lesson;
	inTask: Said;
	locked: boolean;
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
			{task.drawing !== "" && (
				<Diagram drawing={task.drawing} label={words.text("task.drawing")} />
			)}
			{lesson.hint.open && !isSettled(lesson.answer) && (
				<Note tone="hint" label={words.text("task.hint")} said={inTask}>
					{task.hint}
				</Note>
			)}
			<OptionList
				legend={words.text(
					recordedResult(lesson.answer) === undefined
						? "task.pick"
						: "result.answers",
				)}
				options={optionsOf(handed, lesson.answer, (key) => words.text(key))}
				locked={locked}
				onSelect={onAnswer}
				said={inTask}
			/>
		</div>
	);
}

// TaskActions are the buttons under a task. A task answered has one thing
// left to do, the next task. A task closed has none, and its card no foot: the
// task being solved is on a newer card, and asking for another here would
// skip it.
// The next task's button is its own: the one pressed is not turned into it
// under the focus, which a screen reader would not tell. While an ask for
// another task is on its way to the chat its button takes no second press;
// once the chat has it, the button may be pressed again, since a host may hold
// the message for the person to send and the card cannot see whether it went,
// and a second ask while a task is being written is handed the one open.
// While a choice of the topic is on its way, every button is locked.
function TaskActions({
	lesson,
	another,
	saving,
	nextTask,
	onHint,
	onAnother,
}: {
	lesson: Lesson;
	another: Request;
	saving: boolean;
	nextTask: Ref<HTMLButtonElement>;
	onHint: () => void;
	onAnother: () => void;
}) {
	const words = useWords();
	const anotherSending = another === "sending" || saving;
	if (lesson.answer.state === "answered") {
		return (
			<Button
				key="next"
				variant="primary"
				locked={anotherSending}
				onClick={onAnother}
				buttonRef={nextTask}
			>
				{words.text("task.another")}
			</Button>
		);
	}
	return (
		<LessonButtons
			locked={lesson.answer.state === "checking" || saving}
			anotherSending={anotherSending}
			hintOpen={lesson.hint.open}
			onHint={onHint}
			onAnother={onAnother}
		/>
	);
}

// Outcome is MathTrail's reply to an answer: the result of one recorded, a
// line saying why none was, or both — the result of a task closed since, and
// that it is closed.
function Outcome({
	answer,
	task,
	inTask,
}: {
	answer: Answer;
	task: HandedTask["task"];
	inTask: Said;
}) {
	const words = useWords();
	const result = recordedResult(answer);
	return (
		<>
			{result !== undefined && (
				<TaskResult result={result} task={task} inTask={inTask} />
			)}
			{answer.state !== "answered" && (
				<ReplyCard name={words.text("app.name")}>
					<Verdict>
						{words.text(
							answer.state === "closed" ? "task.closed" : "task.answer_failed",
						)}
					</Verdict>
				</ReplyCard>
			)}
		</>
	);
}

// DayOver is MathTrail's reply to another task asked for on a day with no room
// for one: the card keeps its task, and the child can still answer it, read it
// again or look at the progress.
function DayOver() {
	const words = useWords();
	return (
		<ReplyCard name={words.text("app.name")}>
			<Verdict detail={words.text("waiting.limit_detail")}>
				{words.text("waiting.limit")}
			</Verdict>
		</ReplyCard>
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
	correct: "result.correct_answer",
	wrong: "result.your_answer",
};

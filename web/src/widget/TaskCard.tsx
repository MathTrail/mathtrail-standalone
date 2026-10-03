import type { Ref } from "preact";
import { useReducer, useRef } from "preact/hooks";
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
import { type Request, RequestNote, useChatRequest } from "./ChatRequest";
import { useFocusKeptOnTheCard } from "./focus";
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
} from "./lesson";
import {
	type AnswerOutcome,
	type Choice,
	dontKnow,
	type HandedTask,
	type Letter,
	letters,
	readAnswer,
} from "./payload";
import { TaskResult } from "./TaskResult";
import { type Key, useWords } from "./words";

/**
 * TaskCard is the card a task is handed to the child on. The child answers by
 * pressing an option or "I don't know", which records the answer straight
 * away, and reads the result below the task, in the same card; opens the hint;
 * asks for another task, which goes to the chat for the model to write — the
 * new task comes in a card of its own, below, and this one says so and keeps
 * its task; and opens the progress in the card and comes back to the task as
 * it was left.
 */
export function TaskCard({ handed, host }: { handed: HandedTask; host: Host }) {
	const { task, child } = handed;
	const words = useWords();
	const [lesson, dispatch] = useReducer(next, lessonStart);
	const another = useChatRequest(host);
	const outcome = useRef<HTMLDivElement>(null);
	const nextTask = useRef<HTMLButtonElement>(null);
	// An answer is sent once. From the moment it is on its way the options and
	// the buttons that could give another are locked, and a second press that
	// comes before the card has redrawn to lock them is turned away here.
	const answering = useRef(false);

	async function answer(choice: Choice) {
		if (answering.current) {
			return;
		}
		answering.current = true;
		dispatch({ type: "picked", choice });
		const told = await recordAnswer(host, task.id, choice, lesson.hint.used);
		answering.current = false;
		dispatch({ type: "told", outcome: told });
		if (told.kind === "answered") {
			host.tellModel(modelLineOf(told.result)).catch((error: unknown) => {
				console.error("widget: the model was not told of the answer", error);
			});
		}
	}

	function askForAnother() {
		// Not while an answer is on its way: the ask would race the answer to a
		// task the model is about to set aside. The label of the button, in the
		// card's language, goes to the chat as the child's message: only the
		// model can write a task, and those are the words it takes as the ask.
		if (answering.current) {
			return;
		}
		another.send(words.text("task.another"));
	}

	// Once the answer is in, a focus that was lost goes to the one thing left
	// to do, while the replies read the result out; with nothing left to do,
	// to what the card says.
	useFocusKeptOnTheCard(isOver(lesson.answer), nextTask, outcome);

	// The task is in the language it was written in, which need not be the
	// card's: a screen reader reads it in its own voice, and it runs its own way.
	const inTask: Said = { lang: task.language, dir: directionOf(task.language) };

	return (
		<CardFrame child={child} back={words.text("progress.back")} host={host}>
			{(wide, whose) => (
				<>
					<article aria-label={words.text("task.label")}>
						<CardHeader grade={whose?.grade} wide={wide} />
						<TaskBody
							handed={handed}
							lesson={lesson}
							inTask={inTask}
							onAnswer={answer}
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
					</section>
					{lesson.answer.state !== "closed" && (
						<div class="mt-foot">
							<div class="mt-btns">
								<TaskActions
									lesson={lesson}
									another={another.state}
									nextTask={nextTask}
									onDontKnow={() => answer(dontKnow)}
									onHint={() => dispatch({ type: "hint toggled" })}
									onAnother={askForAnother}
								/>
							</div>
							<RequestNote state={another.state} taken="task.another_coming" />
						</div>
					)}
				</>
			)}
		</CardFrame>
	);
}

// TaskBody is the task itself: its question, its drawing, the hint while it
// is shown and the task is not done with, and the options — to press, or
// marked once the answer is in.
function TaskBody({
	handed,
	lesson,
	inTask,
	onAnswer,
}: {
	handed: HandedTask;
	lesson: Lesson;
	inTask: Said;
	onAnswer: (choice: Choice) => void;
}) {
	const words = useWords();
	const { task } = handed;
	return (
		<div class="mt-body">
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
					lesson.answer.state === "answered" ? "result.answers" : "task.pick",
				)}
				options={optionsOf(handed, lesson.answer, (key) => words.text(key))}
				locked={!canAnswer(lesson)}
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
function TaskActions({
	lesson,
	another,
	nextTask,
	onDontKnow,
	onHint,
	onAnother,
}: {
	lesson: Lesson;
	another: Request;
	nextTask: Ref<HTMLButtonElement>;
	onDontKnow: () => void;
	onHint: () => void;
	onAnother: () => void;
}) {
	const words = useWords();
	const anotherSending = another === "sending";
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
			locked={lesson.answer.state === "checking"}
			anotherSending={anotherSending}
			hintOpen={lesson.hint.open}
			onDontKnow={onDontKnow}
			onHint={onHint}
			onAnother={onAnother}
		/>
	);
}

// Outcome is MathTrail's reply to an answer: the result of one recorded, or a
// line saying why none was.
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
	if (answer.state === "answered") {
		return <TaskResult result={answer.result} task={task} inTask={inTask} />;
	}
	return (
		<ReplyCard name={words.text("app.name")}>
			<Verdict>
				{words.text(
					answer.state === "closed" ? "task.closed" : "task.answer_failed",
				)}
			</Verdict>
		</ReplyCard>
	);
}

// recordAnswer sends the child's answer to the service and reads how it went.
// An answer whose reply never came is one the card cannot call recorded; sent
// again, it is either recorded then or told as the service recorded it.
async function recordAnswer(
	host: Host,
	taskId: string,
	choice: Choice,
	hintUsed: boolean,
): Promise<AnswerOutcome> {
	try {
		const result = await host.callTool("submit_answer", {
			task_id: taskId,
			answer: choice,
			hint_used: hintUsed,
		});
		return readAnswer(result, taskId);
	} catch (error: unknown) {
		console.error("widget: the answer did not reach the service", error);
		return { kind: "failed" };
	}
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

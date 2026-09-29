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
import { CardFrame, CardHeader } from "./CardFrame";
import { useFocusKeptOnTheCard } from "./focus";
import { LessonButtons, QuestionField } from "./LessonFoot";
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
	type Question,
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
import { useAsking, WaitingScreen } from "./WaitingCard";
import { type Key, useWords } from "./words";

/**
 * TaskCard is the card a task is handed to the child on. The child answers by
 * pressing an option or "I don't know", which records the answer straight
 * away, and reads the result below the task, in the same card; opens the hint;
 * asks a question, which goes to the chat for the model to answer there; asks
 * for another task, which the model writes while the card waits for it; and
 * opens the progress in the card and comes back to the task as it was left.
 */
export function TaskCard({ handed, host }: { handed: HandedTask; host: Host }) {
	const { task, child } = handed;
	const words = useWords();
	const [lesson, dispatch] = useReducer(next, lessonStart);
	const asking = useAsking(host);
	const outcome = useRef<HTMLDivElement>(null);
	const nextTask = useRef<HTMLButtonElement>(null);
	const questionsAsked = useRef(0);
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
		// Not while an answer is on its way: its reply would come to a card no
		// longer showing the task. Otherwise the card turns to the wait at once,
		// and the ask reaches the model after it.
		if (answering.current) {
			return;
		}
		dispatch({ type: "another asked" });
		asking.ask();
	}

	async function ask(question: string) {
		const id = questionsAsked.current++;
		dispatch({ type: "asked", id, words: question });
		try {
			await host.sendMessage(question);
			dispatch({ type: "question sent", id });
		} catch (error: unknown) {
			console.error("widget: the question did not reach the chat", error);
			dispatch({ type: "question lost", id });
		}
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
			{(wide) =>
				lesson.stage === "waiting" ? (
					<WaitingScreen grade={child.grade} wide={wide} asking={asking} />
				) : (
					<>
						<article aria-label={words.text("task.label")}>
							<CardHeader grade={child.grade} wide={wide} />
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
							{lesson.questions.map((question) => (
								<ReplyCard
									key={question.id}
									author="person"
									name={words.text("child.you")}
									meta={questionMeta(question, (key) => words.text(key))}
								>
									<p class="mt-reply-lead">{question.words}</p>
								</ReplyCard>
							))}
						</section>
						<div class="mt-foot">
							<QuestionField
								value={lesson.draft}
								onInput={(typed) => dispatch({ type: "typed", words: typed })}
								onSend={ask}
							/>
							<div class="mt-btns">
								<TaskActions
									lesson={lesson}
									nextTask={nextTask}
									onDontKnow={() => answer(dontKnow)}
									onHint={() => dispatch({ type: "hint toggled" })}
									onAnother={askForAnother}
								/>
							</div>
						</div>
					</>
				)
			}
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
// left to do, the next task. A task closed has none on this card: the task
// being solved is on a newer one, and asking for another here would skip it.
// The next task's button is its own: the one pressed is not turned into it
// under the focus, which a screen reader would not tell.
function TaskActions({
	lesson,
	nextTask,
	onDontKnow,
	onHint,
	onAnother,
}: {
	lesson: Lesson;
	nextTask: Ref<HTMLButtonElement>;
	onDontKnow: () => void;
	onHint: () => void;
	onAnother: () => void;
}) {
	const words = useWords();
	if (lesson.answer.state === "answered") {
		return (
			<Button
				key="next"
				variant="primary"
				onClick={onAnother}
				buttonRef={nextTask}
			>
				{words.text("task.another")}
			</Button>
		);
	}
	if (lesson.answer.state === "closed") {
		return null;
	}
	return (
		<LessonButtons
			locked={lesson.answer.state === "checking"}
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
		<ReplyCard author="app" name={words.text("app.name")}>
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

// questionMeta is the note beside a question the child asked: whether it got
// to the chat, and nothing while it is on its way.
function questionMeta(
	question: Question,
	text: (key: Key) => string,
): string | undefined {
	switch (question.state) {
		case "sending":
			return undefined;
		case "sent":
			return text("task.sent");
		case "lost":
			return text("task.not_sent");
	}
}

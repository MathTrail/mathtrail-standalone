import { useEffect, useReducer, useRef } from "preact/hooks";
import { Diagram, Note, Verdict } from "../design/blocks";
import { classes } from "../design/classes";
import {
	Button,
	type Option,
	OptionList,
	type OptionState,
	ReplyField,
	type Said,
} from "../design/controls";
import { MessageHeader, ReplyCard, ThreadBar } from "../design/thread";
import { useWide } from "../design/wide";
import { directionOf } from "../i18n/lookup";
import type { Host } from "./bridge";
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
import { StubCard } from "./StubCard";
import { TaskResult } from "./TaskResult";
import { type Key, useWords } from "./words";

/**
 * TaskCard is the card a task is handed to the child on. The child answers by
 * pressing an option or "I don't know", which records the answer straight
 * away, and reads the result below the task, in the same card; opens the hint;
 * asks a question, which goes to the chat for the model to answer there; asks
 * for another task, which the model writes; and opens the progress in the card
 * and comes back to the task as it was left.
 */
export function TaskCard({ handed, host }: { handed: HandedTask; host: Host }) {
	const { task, child } = handed;
	const words = useWords();
	const [lesson, dispatch] = useReducer(next, lessonStart);
	const root = useRef<HTMLDivElement>(null);
	const wide = useWide(root);
	const topBar = useRef<HTMLButtonElement>(null);
	const backBar = useRef<HTMLButtonElement>(null);
	const outcome = useRef<HTMLDivElement>(null);
	const nextTask = useRef<HTMLButtonElement>(null);
	const waitingTitle = useRef<HTMLParagraphElement>(null);
	const questionsAsked = useRef(0);
	const progressOpenings = useRef(0);
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
		// The card turns to the wait at once; the ask reaches the model after it.
		// If the host drops it, the adult asks in the chat, and the wait is the
		// same.
		dispatch({ type: "another asked" });
		host.sendMessage(words.text("task.another")).catch((error: unknown) => {
			console.error(
				"widget: the ask for another task did not reach the chat",
				error,
			);
		});
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

	async function openProgress() {
		progressOpenings.current += 1;
		const opening = progressOpenings.current;
		dispatch({ type: "progress opened", opening });
		try {
			const read = await host.callTool("read_progress", {});
			dispatch(
				read.isError === true
					? { type: "progress failed", opening }
					: {
							type: "progress read",
							opening,
							payload: read.structuredContent,
						},
			);
		} catch (error: unknown) {
			console.error("widget: the progress did not arrive", error);
			dispatch({ type: "progress failed", opening });
		}
	}

	useFocusFollowsProgress(lesson, topBar, backBar);
	// Once the answer is in, a focus that was lost goes to the one thing left
	// to do, while the replies read the result out; with nothing left to do,
	// to what the card says.
	useFocusKeptOnTheCard(isOver(lesson.answer), nextTask, outcome);
	useFocusKeptOnTheCard(lesson.stage === "waiting", waitingTitle);

	// The task is in the language it was written in, which need not be the
	// card's: a screen reader reads it in its own voice, and it runs its own way.
	const inTask: Said = { lang: task.language, dir: directionOf(task.language) };

	const waiting = lesson.stage === "waiting";
	return (
		<div
			ref={root}
			class={classes(
				"mt",
				"mt-widget",
				wide && "mt-wide",
				words.dir === "rtl" && "mt-rtl",
			)}
		>
			<div hidden={lesson.progress !== undefined}>
				<ThreadBar
					name={child.pseudonym}
					action={words.text("task.profile_action")}
					onClick={openProgress}
					buttonRef={topBar}
				/>
				<article aria-label={words.text("task.label")}>
					<MessageHeader
						author="app"
						name={words.text("app.name")}
						badge={words.text("app.badge", { grade: child.grade })}
						wide={wide}
					/>
					{waiting ? (
						<div class="mt-body">
							<div class="mt-gen">
								<p class="mt-gen-title" ref={waitingTitle} tabIndex={-1}>
									{words.text("waiting.title")}
								</p>
							</div>
						</div>
					) : (
						<div class="mt-body">
							<p class="mt-task-text" lang={inTask.lang} dir={inTask.dir}>
								{task.question}
							</p>
							{task.drawing !== "" && (
								<Diagram
									drawing={task.drawing}
									label={words.text("task.drawing")}
								/>
							)}
							{lesson.hint.open && !isSettled(lesson.answer) && (
								<Note tone="hint" label={words.text("task.hint")} said={inTask}>
									{task.hint}
								</Note>
							)}
							<OptionList
								legend={words.text(
									lesson.answer.state === "answered"
										? "result.answers"
										: "task.pick",
								)}
								options={optionsOf(handed, lesson.answer, (key) =>
									words.text(key),
								)}
								locked={!canAnswer(lesson)}
								onSelect={answer}
								said={inTask}
							/>
						</div>
					)}
				</article>
				<section
					class="mt-replies"
					aria-label={words.text("task.replies")}
					aria-live="polite"
				>
					{!waiting && (
						<>
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
						</>
					)}
				</section>
				<div class="mt-foot">
					<ReplyField
						value={lesson.draft}
						placeholder={words.text("task.ask")}
						label={words.text("task.ask_label")}
						sendLabel={words.text("task.send")}
						disabled={waiting}
						onInput={(typed) => dispatch({ type: "typed", words: typed })}
						onSend={ask}
					/>
					<div class="mt-btns">
						{/* A task answered has one thing left to do; a task closed has
						    none on this card, since the task being solved is on a newer
						    one, and asking for another here would skip it. */}
						{!waiting && lesson.answer.state === "answered" ? (
							<Button
								variant="primary"
								onClick={askForAnother}
								buttonRef={nextTask}
							>
								{words.text("task.another")}
							</Button>
						) : !waiting && lesson.answer.state === "closed" ? null : (
							<>
								<Button
									disabled={waiting}
									locked={lesson.answer.state === "checking"}
									onClick={() => answer(dontKnow)}
								>
									{words.text("task.idk")}
								</Button>
								<Button
									disabled={waiting}
									locked={lesson.answer.state === "checking"}
									expanded={waiting ? undefined : lesson.hint.open}
									onClick={() => dispatch({ type: "hint toggled" })}
								>
									{words.text(
										lesson.hint.open && !waiting
											? "task.hide_hint"
											: "task.hint",
									)}
								</Button>
								<Button
									disabled={waiting}
									locked={lesson.answer.state === "checking"}
									onClick={askForAnother}
								>
									{words.text("task.another")}
								</Button>
							</>
						)}
					</div>
				</div>
			</div>
			{lesson.progress !== undefined && (
				<>
					<ThreadBar
						variant="back"
						label={words.text("progress.back")}
						onClick={() => dispatch({ type: "back to task" })}
						buttonRef={backBar}
					/>
					<article
						aria-label={words.text("progress.label")}
						aria-busy={lesson.progress.state === "reading"}
					>
						<MessageHeader
							author="person"
							name={child.pseudonym}
							badge={words.text("child.grade", { grade: child.grade })}
							wide={wide}
						/>
						<div class="mt-progress">
							{lesson.progress.state === "read" && (
								<StubCard payload={lesson.progress.payload} />
							)}
							{lesson.progress.state === "failed" && (
								<Verdict>{words.text("progress.failed")}</Verdict>
							)}
						</div>
					</article>
				</>
			)}
		</div>
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

// useFocusFollowsProgress moves the focus with the card between the task and
// the progress: to the way back as the progress opens, to the way there as it
// closes. The first draw moves nothing.
function useFocusFollowsProgress(
	lesson: Lesson,
	topBar: { current: HTMLButtonElement | null },
	backBar: { current: HTMLButtonElement | null },
): void {
	const open = lesson.progress !== undefined;
	const wasOpen = useRef(open);
	useEffect(() => {
		if (wasOpen.current === open) {
			return;
		}
		wasOpen.current = open;
		(open ? backBar : topBar).current?.focus({ preventScroll: true });
	}, [open, topBar, backBar]);
}

// useFocusKeptOnTheCard gives the focus to target — or, when there is none on
// the card, to fallback — when shown comes true and the focus has been lost:
// the button pressed has gone or been switched off. A focus still on something
// is left where it is, and the page is not scrolled.
function useFocusKeptOnTheCard(
	shown: boolean,
	target: { current: HTMLElement | null },
	fallback?: { current: HTMLElement | null },
): void {
	useEffect(() => {
		if (shown && focusIsLost()) {
			(target.current ?? fallback?.current)?.focus({ preventScroll: true });
		}
	}, [shown, target, fallback]);
}

// focusIsLost says whether the focus is on nothing a child could act on: the
// page itself, or an element that has left the page or been switched off.
function focusIsLost(): boolean {
	const focused = document.activeElement;
	return (
		focused === null ||
		focused === document.body ||
		!focused.isConnected ||
		(focused instanceof HTMLButtonElement && focused.disabled)
	);
}

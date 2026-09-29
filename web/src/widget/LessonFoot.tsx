import { Button, ReplyField } from "../design/controls";
import { useWords } from "./words";

/**
 * QuestionField is the field a question about the task is typed into. Switched
 * off, while a card has no task to ask about, it is empty and takes nothing.
 */
export function QuestionField(
	props:
		| {
				off?: false;
				value: string;
				onInput: (typed: string) => void;
				onSend: (question: string) => void;
		  }
		| { off: true },
) {
	const words = useWords();
	const said = {
		placeholder: words.text("task.ask"),
		label: words.text("task.ask_label"),
		sendLabel: words.text("task.send"),
	};
	if (props.off === true) {
		return (
			<ReplyField
				{...said}
				value=""
				disabled
				onInput={nothing}
				onSend={nothing}
			/>
		);
	}
	return (
		<ReplyField
			{...said}
			value={props.value}
			onInput={props.onInput}
			onSend={props.onSend}
		/>
	);
}

/**
 * LessonButtons are the three buttons under a task: "I don't know", the hint,
 * and another task. While an answer is checked they are locked; while a card
 * waits for a task they are switched off — the same three in the same places,
 * so that a card does not jump as it turns to the wait.
 */
export function LessonButtons(
	props:
		| {
				off?: false;
				locked: boolean;
				hintOpen: boolean;
				onDontKnow: () => void;
				onHint: () => void;
				onAnother: () => void;
		  }
		| { off: true },
) {
	const words = useWords();
	const on = props.off === true ? undefined : props;
	const off = on === undefined;
	return (
		<>
			<Button
				key="dont-know"
				disabled={off}
				locked={on?.locked}
				onClick={on?.onDontKnow}
			>
				{words.text("task.idk")}
			</Button>
			<Button
				key="hint"
				disabled={off}
				locked={on?.locked}
				expanded={on?.hintOpen}
				onClick={on?.onHint}
			>
				{words.text(on?.hintOpen ? "task.hide_hint" : "task.hint")}
			</Button>
			<Button
				key="another"
				disabled={off}
				locked={on?.locked}
				onClick={on?.onAnother}
			>
				{words.text("task.another")}
			</Button>
		</>
	);
}

// nothing is what a switched-off field does with what it is given.
function nothing(): void {
	// A card with no task has nothing to ask about.
}

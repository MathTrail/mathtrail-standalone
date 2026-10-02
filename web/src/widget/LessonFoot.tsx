import { Button } from "../design/controls";
import { useWords } from "./words";

/**
 * LessonButtons are the three buttons under a task: "I don't know", the hint,
 * and another task. While an answer is checked they are locked, and while an
 * ask for another task is on its way to the chat its button is too.
 */
export function LessonButtons({
	locked,
	anotherSending,
	hintOpen,
	onDontKnow,
	onHint,
	onAnother,
}: {
	locked: boolean;
	anotherSending: boolean;
	hintOpen: boolean;
	onDontKnow: () => void;
	onHint: () => void;
	onAnother: () => void;
}) {
	const words = useWords();
	return (
		<>
			<Button key="dont-know" locked={locked} onClick={onDontKnow}>
				{words.text("task.idk")}
			</Button>
			<Button key="hint" locked={locked} expanded={hintOpen} onClick={onHint}>
				{words.text(hintOpen ? "task.hide_hint" : "task.hint")}
			</Button>
			<Button
				key="another"
				locked={locked || anotherSending}
				onClick={onAnother}
			>
				{words.text("task.another")}
			</Button>
		</>
	);
}

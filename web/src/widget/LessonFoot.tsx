import { Button } from "../design/controls";
import { useWords } from "./words";

/**
 * LessonButtons are the buttons under a task not yet answered: the hint, and
 * another task. While an answer is checked they are locked, and while an ask
 * for another task is on its way to the chat its button is too.
 */
export function LessonButtons({
	locked,
	anotherSending,
	hintOpen,
	onHint,
	onAnother,
}: {
	locked: boolean;
	anotherSending: boolean;
	hintOpen: boolean;
	onHint: () => void;
	onAnother: () => void;
}) {
	const words = useWords();
	return (
		<>
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

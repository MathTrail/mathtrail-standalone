import { IconButton } from "../design/controls";
import { useWords } from "./words";

/**
 * LessonButtons are the buttons under a task not yet answered, each drawn by
 * its icon: the hint, a bulb, and another task, two arrows in a circle. While
 * an answer is checked they are locked, and while an ask for another task is
 * on its way to the chat its button is too.
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
			<IconButton
				key="hint"
				icon="hint"
				label={words.text(hintOpen ? "task.hide_hint" : "task.hint")}
				className="mt-btn-hint"
				locked={locked}
				expanded={hintOpen}
				onClick={onHint}
			/>
			<IconButton
				key="another"
				icon="renew"
				label={words.text("task.another")}
				className="mt-btn-another"
				locked={locked || anotherSending}
				onClick={onAnother}
			/>
		</>
	);
}

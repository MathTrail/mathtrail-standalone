import { useEffect, useState } from "preact/hooks";
import type { CallStage } from "./bridge";
import { CardHeader, CardRoot } from "./CardRoot";
import { TaskWait } from "./TaskWait";
import { moments } from "./waiting";
import { useWords } from "./words";

/**
 * ChoosingCard is the card of a task asked for, drawn before the service has
 * answered the ask: the usual course of a task, its topic and difficulty being
 * picked; the word that the task was not finished when the call is cancelled;
 * and, when no answer has come long after, the word that the task is taking
 * long. It shows nothing for its first moment, so that an answer that comes at
 * once does not make a wait flash past, and it has nothing to press. Before
 * the answer it knows no child and no lesson: it has no line at its top and no
 * grade, and it speaks the host's language.
 */
export function ChoosingCard({ stage }: { stage: CallStage }) {
	const words = useWords();
	const shown = useLater(moments.shown);
	const slow = useLater(moments.slow);
	if (!shown) {
		return null;
	}
	return (
		<CardRoot>
			{(wide) => (
				<article aria-label={words.text("waiting.label")}>
					<CardHeader grade={undefined} wide={wide} />
					<TaskWait
						phase="choosing"
						slow={slow}
						ended={
							stage === "cancelled"
								? { said: "waiting.cancelled", next: "waiting.ask_in_chat" }
								: undefined
						}
					/>
				</article>
			)}
		</CardRoot>
	);
}

// useLater says whether the card has been up for the milliseconds given.
function useLater(milliseconds: number): boolean {
	const [past, setPast] = useState(false);
	useEffect(() => {
		const timer = setTimeout(() => setPast(true), milliseconds);
		return () => clearTimeout(timer);
	}, [milliseconds]);
	return past;
}

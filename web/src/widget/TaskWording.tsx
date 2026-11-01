import type { Said } from "../design/controls";
import { Diagram } from "../design/picture/diagram";
import type { Picture } from "../design/picture/model";
import { useWords } from "./words";

/**
 * TaskWording is a task as the child reads it: its question, and under it its
 * picture where it has one, laid out for locale, both in the language said. It
 * is drawn the same on the task's card and at the head of how its answer went.
 */
export function TaskWording({
	question,
	picture,
	locale,
	said,
}: {
	question: string;
	picture: Picture | undefined;
	locale: string;
	said: Said;
}) {
	const words = useWords();
	return (
		<>
			<p class="mt-task-text" lang={said.lang} dir={said.dir}>
				{question}
			</p>
			{picture !== undefined && (
				<Diagram
					picture={picture}
					label={words.text(`picture.${picture.kind}`)}
					locale={locale}
					keyLabel={words.text("picture.colors")}
					said={said}
				/>
			)}
		</>
	);
}

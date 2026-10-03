import { Note, Verdict } from "../design/blocks";
import { CardHeader, CardRoot } from "./CardRoot";
import type { FirstRun } from "./payload";
import { useWords } from "./words";

/**
 * FirstRunCard is the card of an account with no profile yet, drawn when the
 * model finds none.
 */
export function FirstRunCard({ firstRun }: { firstRun: FirstRun }) {
	return (
		<CardRoot>
			{(wide) => <FirstRunScreen firstRun={firstRun} wide={wide} />}
		</CardRoot>
	);
}

/**
 * FirstRunScreen tells the adult what MathTrail is, the rule of the
 * pseudonym, what the chat will ask, where the answers are kept and how to
 * start: by asking in the chat, where the model asks the adult to say they are
 * the child's parent or tutor and makes the profile. A card makes none. A
 * profile the model failed to make, in a card of an earlier chat, is said to
 * be not made.
 */
export function FirstRunScreen({
	firstRun,
	wide,
}: {
	firstRun: FirstRun;
	wide: boolean;
}) {
	const words = useWords();
	return (
		<article aria-label={words.text("first_run.label")}>
			<CardHeader grade={undefined} wide={wide} />
			<div class="mt-body">
				{firstRun.refused && (
					<Verdict detail={words.text("profile.not_saved_detail")}>
						{words.text("first_run.not_created")}
					</Verdict>
				)}
				<p class="mt-lead">{words.text("first_run.lead")}</p>
				<Note label={words.text("first_run.pseudonym_label")}>
					{words.text("first_run.pseudonym")}
				</Note>
				<Note label={words.text("first_run.fill_label")}>
					{words.text("first_run.fill")}
				</Note>
				<Note label={words.text("first_run.data_label")}>
					{words.text("first_run.data")}
				</Note>
				<Note label={words.text("first_run.start_label")}>
					{words.text("first_run.start")}
				</Note>
			</div>
		</article>
	);
}

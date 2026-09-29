import { useState } from "preact/hooks";
import { Note, Verdict } from "../design/blocks";
import { Button, Checkbox } from "../design/controls";
import type { Host } from "./bridge";
import { CardHeader, CardRoot } from "./CardRoot";
import { RequestNote, useChatRequest } from "./ChatRequest";
import type { FirstRun } from "./payload";
import { useWords } from "./words";

/**
 * FirstRunCard is the card of an account with no profile yet, drawn when the
 * model finds none.
 */
export function FirstRunCard({
	firstRun,
	host,
}: {
	firstRun: FirstRun;
	host: Host;
}) {
	return (
		<CardRoot>
			{(wide) => <FirstRunScreen firstRun={firstRun} wide={wide} host={host} />}
		</CardRoot>
	);
}

/**
 * FirstRunScreen tells the adult what MathTrail is, the rule of the
 * pseudonym, what the chat will ask and where the answers are kept, and lets
 * them ask for the profile — the model asks for its details and makes it —
 * once they have said they are the child's parent or tutor. A profile the
 * model just failed to make is said to be not made.
 */
export function FirstRunScreen({
	firstRun,
	wide,
	host,
}: {
	firstRun: FirstRun;
	wide: boolean;
	host: Host;
}) {
	const words = useWords();
	const [adult, setAdult] = useState(false);
	const request = useChatRequest(host);
	return (
		<>
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
				</div>
			</article>
			<div class="mt-foot">
				<Checkbox
					label={words.text("first_run.confirm")}
					checked={adult}
					onChange={setAdult}
				/>
				<div class="mt-btns">
					<Button
						variant="primary"
						disabled={!adult}
						locked={request.state === "sent"}
						onClick={() => request.send(words.text("first_run.create"))}
					>
						{words.text("first_run.create")}
					</Button>
				</div>
				<RequestNote state={request.state} />
			</div>
		</>
	);
}

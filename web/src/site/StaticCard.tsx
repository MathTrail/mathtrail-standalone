import type { ComponentChildren } from "preact";
import type { Host } from "../widget/bridge";
import { CardRoot } from "../widget/CardRoot";
import type { Open } from "../widget/folds";
import { ProgressScreen } from "../widget/ProgressScreen";
import type {
	AnswerResult,
	HandedTask,
	ProgressReport,
} from "../widget/payload";
import { TaskCard } from "../widget/TaskCard";
import { NamesBuild } from "../widget/version";
import { cardWords, WordsContext } from "../widget/words";

// stillHost answers nothing: a card drawn on a page has no chat behind it,
// and being inert, it never asks.
const stillHost: Host = {
	callTool: () => Promise.reject(new Error("a card on a page has no host")),
	sendMessage: () => Promise.reject(new Error("a card on a page has no host")),
	tellModel: () => Promise.reject(new Error("a card on a page has no host")),
};

/**
 * StaticProgress draws the widget's progress card on a page that runs no
 * script, as a chat draws it for report: the widget's own components, in the
 * page's language as the widget's dictionaries say it, with the sections open
 * that open names. A page is drawn before anybody reads it, so it cannot
 * measure the room the card has, and the card is drawn at the narrow width a
 * phone's chat gives it. Nothing on the page answers its buttons, so the card
 * is inert — it cannot be pressed, focused or selected — and is shown rather
 * than used.
 */
export function StaticProgress({
	report,
	locale,
	open,
}: {
	report: ProgressReport;
	locale: string;
	open: Open;
}) {
	return (
		<Still locale={locale}>
			<CardRoot>
				{(wide) => (
					<ProgressScreen
						report={report}
						wide={wide}
						host={stillHost}
						folds={{ open, toggle: () => {} }}
					/>
				)}
			</CardRoot>
		</Still>
	);
}

/**
 * StaticAnswer draws the widget's card of a task once its answer is in, on a
 * page that runs no script, as a chat draws it when the service has recorded
 * result: the options marked, the trap, the solution step by step and the
 * rating, in the page's language as the widget's dictionaries say it. Like the
 * progress, it is drawn at the narrow width and is inert.
 */
export function StaticAnswer({
	handed,
	result,
	locale,
}: {
	handed: HandedTask;
	result: AnswerResult;
	locale: string;
}) {
	return (
		<Still locale={locale}>
			<TaskCard
				handed={handed}
				host={stillHost}
				start={{
					hint: { open: false, used: result.hint_used },
					answer: { state: "answered", result },
				}}
			/>
		</Still>
	);
}

// Still is a card drawn on a page: shown rather than used, speaking the
// widget's words in the page's language, and naming no build, since the site
// is published before the release its commit becomes is tagged.
function Still({
	locale,
	children,
}: {
	locale: string;
	children: ComponentChildren;
}) {
	return (
		<div class="s-card" inert>
			<NamesBuild.Provider value={false}>
				<WordsContext.Provider value={cardWords(locale, undefined)}>
					{children}
				</WordsContext.Provider>
			</NamesBuild.Provider>
		</div>
	);
}

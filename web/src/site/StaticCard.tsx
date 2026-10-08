import type { ComponentChildren } from "preact";
import type { Host } from "../widget/bridge";
import { CardRoot } from "../widget/CardRoot";
import { ComingCard } from "../widget/ComingCard";
import { cardWords } from "../widget/dictionaries";
import type { Open } from "../widget/folds";
import type { Lesson } from "../widget/lesson";
import { ProgressScreen } from "../widget/ProgressScreen";
import type {
	AnswerResult,
	Coming,
	HandedTask,
	ProgressReport,
} from "../widget/payload";
import { TaskCard } from "../widget/TaskCard";
import { ChoosesTopic } from "../widget/TopicChoice";
import { NamesBuild, versionGiven } from "../widget/version";
import { WordsContext } from "../widget/words";

// stillHost answers nothing and opens nothing: a card drawn on a page has no
// chat behind it, and being inert, it never asks.
const stillHost: Host = {
	callTool: () => Promise.reject(new Error("a card on a page has no host")),
	sendMessage: () => Promise.reject(new Error("a card on a page has no host")),
	tellModel: () => Promise.reject(new Error("a card on a page has no host")),
	canOpenLinks: () => false,
	openLink: () => Promise.resolve(false),
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
 * StaticTask draws the widget's card of a task on a page that runs no script,
 * as a chat draws it once the lesson on it has got as far as start: the task
 * as it is handed out, an option picked and being checked, the hint open, or
 * the answer in. Like the progress, it is drawn at the narrow width and is
 * inert.
 */
export function StaticTask({
	handed,
	start,
	locale,
}: {
	handed: HandedTask;
	start: Lesson;
	locale: string;
}) {
	return (
		<Still locale={locale}>
			<TaskCard handed={handed} host={stillHost} start={start} />
		</Still>
	);
}

/**
 * StaticAnswer draws the widget's card of a task once its answer is in, as a
 * chat draws it when the service has recorded result: the options marked, the
 * trap, the solution step by step and the rating, in the page's language as
 * the widget's dictionaries say it.
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
		<StaticTask
			handed={handed}
			start={{
				hint: { open: false, used: result.hint_used },
				answer: { state: "answered", result },
			}}
			locale={locale}
		/>
	);
}

/**
 * StaticComing draws the widget's card of a task asked for and still being
 * written, as a chat draws it before the service has said anything of it: the
 * course of a task, its writing under way. The card asks nothing on a page —
 * it is drawn once, and nothing runs on it afterwards.
 */
export function StaticComing({
	coming,
	locale,
}: {
	coming: Coming;
	locale: string;
}) {
	return (
		<Still locale={locale}>
			<ComingCard coming={coming} host={stillHost} />
		</Still>
	);
}

// Still is a card drawn on a page: shown rather than used, speaking the
// widget's words in the page's language, and naming the build only when the
// site was built from a release, which is the one the chats run. A site built
// from anything else names none, rather than a "dev" nobody runs. A page lets
// no topic be chosen, so a task's button of the topic is drawn locked.
function Still({
	locale,
	children,
}: {
	locale: string;
	children: ComponentChildren;
}) {
	return (
		<div class="s-card" inert>
			<NamesBuild.Provider value={versionGiven()}>
				<WordsContext.Provider value={cardWords(locale, undefined)}>
					<ChoosesTopic.Provider value={false}>
						{children}
					</ChoosesTopic.Provider>
				</WordsContext.Provider>
			</NamesBuild.Provider>
		</div>
	);
}

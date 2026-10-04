import type { Host } from "../widget/bridge";
import { CardRoot } from "../widget/CardRoot";
import type { Open } from "../widget/folds";
import { ProgressScreen } from "../widget/ProgressScreen";
import type { ProgressReport } from "../widget/payload";
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
		<div class="s-card" inert>
			<WordsContext.Provider value={cardWords(locale, undefined)}>
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
			</WordsContext.Provider>
		</div>
	);
}

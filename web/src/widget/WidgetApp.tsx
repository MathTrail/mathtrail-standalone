import { useEffect, useLayoutEffect, useMemo, useState } from "preact/hooks";
import type { Words } from "../i18n/words";
import type { Bridge, Host } from "./bridge";
import { FirstRunCard } from "./FirstRunScreen";
import { ProfileCard } from "./ProfileScreen";
import { ProgressCard } from "./ProgressScreen";
import { readScreen } from "./payload";
import { TaskCard } from "./TaskCard";
import { UnreadableCard } from "./UnreadableCard";
import { WaitingCard } from "./WaitingCard";
import { cardWords, type Key, languageChosenIn, WordsContext } from "./words";

/**
 * WidgetApp is the card a tool's result is drawn as, in the words of the
 * language the parent chose for the cards or, when they chose none, of the
 * host's. Until the first result arrives it draws nothing.
 */
export function WidgetApp({ bridge, host }: { bridge: Bridge; host: Host }) {
	const result = useBridge(bridge, latestResult);
	const hostLocale = useBridge(bridge, localeOfHost);
	const words = useMemo(
		() => cardWords(languageChosenIn(result?.structuredContent), hostLocale),
		[result, hostLocale],
	);
	useDocumentLanguage(words);
	if (result === undefined) {
		return null;
	}
	return (
		<WordsContext.Provider value={words}>
			<Screen payload={result.structuredContent} host={host} />
		</WordsContext.Provider>
	);
}

// Screen is the card a payload draws: a task handed to the child, a new card
// for each task; a wait for the next task, a new wait for each payload that
// asks for one; the progress, the profile or the first sign-in. A payload that
// names none of them, or does not read as the one it names, draws a card that
// says so.
function Screen({ payload, host }: { payload: unknown; host: Host }) {
	const shown = useMemo(() => readScreen(payload), [payload]);
	// A payload told again is the same card; another one — the next refusal of
	// a task, say — starts the card afresh.
	const said = useMemo(() => JSON.stringify(payload ?? null), [payload]);
	switch (shown?.screen) {
		case "task":
			return (
				<TaskCard
					key={shown.handed.task.id}
					handed={shown.handed}
					host={host}
				/>
			);
		case "waiting":
			return <WaitingCard key={said} waiting={shown.waiting} host={host} />;
		case "progress":
			return <ProgressCard key={said} report={shown.report} host={host} />;
		case "profile":
			return <ProfileCard key={said} profile={shown.profile} host={host} />;
		case "first_run":
			return <FirstRunCard key={said} firstRun={shown.firstRun} host={host} />;
		case undefined:
			return <UnreadableCard />;
	}
}

const latestResult = (bridge: Bridge) => bridge.result();
const localeOfHost = (bridge: Bridge) => bridge.locale();

// useBridge is what read finds on the bridge, kept current as the host tells
// the widget more.
function useBridge<T>(bridge: Bridge, read: (bridge: Bridge) => T): T {
	const [value, setValue] = useState(() => read(bridge));
	useEffect(() => {
		// The host may have told more between the first render and this
		// subscription; it is read once more rather than missed.
		setValue(() => read(bridge));
		return bridge.subscribe(() => setValue(() => read(bridge)));
	}, [bridge, read]);
	return value;
}

// useDocumentLanguage names on the page the language its words are in and the
// way they run: a screen reader reads them in that language, and a language
// written right to left lays the card out mirrored. Both are set before the
// card is painted, so that it never shows a frame the wrong way round.
function useDocumentLanguage(words: Words<Key>): void {
	useLayoutEffect(() => {
		document.documentElement.lang = words.locale;
		document.documentElement.dir = words.dir;
	}, [words]);
}

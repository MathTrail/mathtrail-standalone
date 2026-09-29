import { useEffect, useLayoutEffect, useMemo, useState } from "preact/hooks";
import type { Words } from "../i18n/words";
import type { Bridge, Host } from "./bridge";
import { readHandedTask, readWaiting } from "./payload";
import { StubCard } from "./StubCard";
import { TaskCard } from "./TaskCard";
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

// Screen is the screen a payload draws. A task handed to the child draws its
// card, a new card for each task; a wait for the next task draws the waiting
// card, a new wait for each payload that asks for one; a screen not drawn yet
// shows its payload as it arrived.
function Screen({ payload, host }: { payload: unknown; host: Host }) {
	const handed = useMemo(() => readHandedTask(payload), [payload]);
	const waiting = useMemo(() => readWaiting(payload), [payload]);
	// A payload told again is the same wait; another one — the next refusal,
	// say — starts the card afresh.
	const said = useMemo(() => JSON.stringify(payload ?? null), [payload]);
	if (handed !== undefined) {
		return <TaskCard key={handed.task.id} handed={handed} host={host} />;
	}
	if (waiting !== undefined) {
		return <WaitingCard key={said} waiting={waiting} host={host} />;
	}
	return <StubCard payload={payload} />;
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

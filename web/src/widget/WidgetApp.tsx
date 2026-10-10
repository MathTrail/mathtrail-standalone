import { useEffect, useLayoutEffect, useMemo, useState } from "preact/hooks";
import type { Words } from "../i18n/words";
import type { Bridge, Call, Host } from "./bridge";
import { OpensProgress } from "./CardFrame";
import { ChoosingCard } from "./ChoosingCard";
import { ComingCard } from "./ComingCard";
import { cardWords } from "./dictionaries";
import { FirstRunCard } from "./FirstRunScreen";
import { serviceThrough } from "./hostedService";
import { ProfileCard } from "./ProfileScreen";
import { ProgressOverCard } from "./ProgressOverCard";
import { ProgressCard } from "./ProgressScreen";
import { readScreen } from "./payload";
import { ResultCard } from "./ResultCard";
import { ServiceContext } from "./service";
import { TaskCard } from "./TaskCard";
import { UnreadableCard } from "./UnreadableCard";
import { WaitingCard } from "./WaitingCard";
import { type Key, languageIn, WordsContext } from "./words";

/**
 * WidgetApp is the card a tool's result is drawn as, in the words of the
 * language its result names — the lesson's, or the one the parent chose — or,
 * when it names none, of the one a task was asked for in, on the card of that
 * ask, and else of the host's. Until the first result arrives it draws the
 * wait for a task asked for, when the host says that is the call it was drawn
 * for, and nothing otherwise. A result with nothing for a card in it, and no
 * failure either, draws nothing: a tool that once answered the model in words
 * alone, drawn again with an earlier chat, has no card to show.
 */
export function WidgetApp({ bridge, host }: { bridge: Bridge; host: Host }) {
	const result = useBridge(bridge, latestResult);
	const call = useBridge(bridge, callOf);
	const hostLocale = useBridge(bridge, localeOfHost);
	const asked = asksForATask(call);
	// The language of a task's ask is the lesson's as the chat has it; the
	// arguments of any other call say nothing of the lesson.
	const askedIn = asked ? call.language : undefined;
	const words = useMemo(
		() =>
			cardWords(languageIn(result?.structuredContent) ?? askedIn, hostLocale),
		[result, askedIn, hostLocale],
	);
	// One service for as long as the host is the same: a card's wait asks it
	// afresh whenever it changes.
	const service = useMemo(() => serviceThrough(host), [host]);
	useDocumentLanguage(words);
	if (result === undefined) {
		return asked ? (
			<WordsContext.Provider value={words}>
				<ChoosingCard stage={call.stage} languageTold={askedIn !== undefined} />
			</WordsContext.Provider>
		) : null;
	}
	if (result.structuredContent === undefined && result.isError !== true) {
		return null;
	}
	return (
		<WordsContext.Provider value={words}>
			<ServiceContext.Provider value={service}>
				<OpensProgress.Provider value={ProgressOverCard}>
					<Screen payload={result.structuredContent} host={host} />
				</OpensProgress.Provider>
			</ServiceContext.Provider>
		</WordsContext.Provider>
	);
}

// Screen is the card a payload draws: a task handed to the child, a new card
// for each task; how an answer went; a task on its way, a new card for each
// request; a card a task did not come to, a new one for each payload that says
// so; the progress, the profile or the first sign-in. A payload that names none of them, or does not
// read as the one it names, draws a card that says so.
function Screen({ payload, host }: { payload: unknown; host: Host }) {
	const shown = useMemo(() => readScreen(payload), [payload]);
	// A task's card stays the same card for the same task, and the card a task
	// is on its way to for the same request, whatever else a payload told again
	// says: what the child did on it stays. Any other payload told again is the
	// same card, and another one starts it afresh.
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
		case "result":
			return <ResultCard key={said} shown={shown.shown} host={host} />;
		case "coming":
			return (
				<ComingCard
					key={shown.coming.requestId}
					coming={shown.coming}
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
			return <FirstRunCard key={said} firstRun={shown.firstRun} />;
		case undefined:
			return <UnreadableCard />;
	}
}

const latestResult = (bridge: Bridge) => bridge.result();
const callOf = (bridge: Bridge) => bridge.call();
const localeOfHost = (bridge: Bridge) => bridge.locale();

// asksForATask says whether the call that drew the card asks for a task. A
// host may put a prefix of its own before the tool's name, and each host
// writes it its own way; the name is matched by its end, which no other tool
// of the service shares.
function asksForATask(call: Call): boolean {
	return call.tool?.endsWith("next_task") ?? false;
}

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

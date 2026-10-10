import { render } from "preact";
import { useMemo, useRef, useState } from "preact/hooks";
import { IdScope } from "../design/ids";
import { openWords } from "../i18n/words";
import { ComingCard } from "../widget/ComingCard";
import { CardIsTheDocument } from "../widget/focus";
import type { AnswerResult } from "../widget/payload";
import { ResultCard } from "../widget/ResultCard";
import { ServiceContext } from "../widget/service";
import { resultShownOf } from "../widget/shown";
import { TaskCard } from "../widget/TaskCard";
import { ChoosesTopic } from "../widget/TopicChoice";
import { NamesBuild, versionGiven } from "../widget/version";
import { type Key, WordsContext } from "../widget/words";
import type { DemoData } from "./data";
import { demoHost, demoService } from "./service";

/**
 * liveScope is what the ids of the live card begin with, apart from those of
 * the cards the page was built with.
 */
export const liveScope = "live-";

// Scene is what the card on the first screen shows: the lesson's task; how an
// answer to it went, once its card asked the chat to go over it; or the task
// asked for in the chat, on its way to a card of its own.
type Scene =
	| { kind: "task" }
	| { kind: "result"; result: AnswerResult }
	| { kind: "coming"; requestId: string };

/**
 * HeroDemo is the card on the first screen come alive: the widget's own card
 * of the lesson's task, in the page's language, answering presses as a chat's
 * does — an option checked and marked, the hint opened. The answer is gone
 * over as a chat goes over it: the card's ask in the chat's frame, as the
 * parent's message, and under it the card of how the answer went. The page
 * answers for the service, and the card, a part of the page, takes no focus
 * back from the reader. Its button of the topic, the coach choosing, takes no
 * press: the page lets no topic be chosen. Another task is asked for as a chat
 * asks for one: the parent's message, under it the card the task is written
 * on, and then the lesson's task once more. frame is the body of the chat's
 * frame the card stands in, and bubble the message above the card.
 */
export function HeroDemo({
	data,
	frame,
	bubble,
}: {
	data: DemoData;
	frame: HTMLElement;
	bubble: HTMLElement | null;
}) {
	const words = useMemo(
		() => openWords<Key>(data.locale, new Map([[data.locale, data.words]])),
		[data],
	);
	const service = useMemo(() => demoService(data), [data]);
	const [scene, setScene] = useState<Scene>({ kind: "task" });
	const asks = useRef(0);
	const host = useMemo(
		() =>
			demoHost((text) => {
				// The card asked for is drawn in place of the one that asked, so the
				// focus on its button would be lost with it: the chat's frame keeps
				// it instead, as a chat keeps it in the conversation.
				const focused = frame.contains(document.activeElement);
				if (bubble !== null) {
					bubble.textContent = text;
				}
				const told = service.told();
				if (text === words.text("task.review") && told !== undefined) {
					setScene({ kind: "result", result: told });
				} else {
					asks.current += 1;
					setScene({ kind: "coming", requestId: `demo_ask_${asks.current}` });
				}
				if (focused) {
					frame.focus({ preventScroll: true });
				}
			}),
		[frame, bubble, service, words],
	);
	return (
		<NamesBuild.Provider value={versionGiven()}>
			<WordsContext.Provider value={words}>
				<IdScope.Provider value={liveScope}>
					<CardIsTheDocument.Provider value={false}>
						<ChoosesTopic.Provider value={false}>
							<ServiceContext.Provider value={service}>
								{scene.kind === "task" && (
									<TaskCard handed={data.handed} host={host} />
								)}
								{scene.kind === "result" && (
									<ResultCard
										key={scene.result.task_id}
										shown={{
											kind: "shown",
											result: resultShownOf(data.handed, scene.result),
										}}
										host={host}
									/>
								)}
								{scene.kind === "coming" && (
									<ComingCard
										key={scene.requestId}
										coming={{
											requestId: scene.requestId,
											child: data.handed.child,
										}}
										host={host}
									/>
								)}
							</ServiceContext.Provider>
						</ChoosesTopic.Provider>
					</CardIsTheDocument.Provider>
				</IdScope.Provider>
			</WordsContext.Provider>
		</NamesBuild.Provider>
	);
}

/**
 * bringHeroAlive puts the live card in place of the still one on the first
 * screen, and says whether it did; a page whose first screen has no card
 * keeps what it has. The live card's place is taken and the card drawn in it
 * in one step, which the page is not painted in the middle of: nobody sees a
 * card half drawn, and the card, drawn in the page, measures the room it has.
 * A card that cannot be drawn gives the still one its place back.
 */
export function bringHeroAlive(document: Document, data: DemoData): boolean {
	const still = document.querySelector(".s-hero-card .s-card");
	const frame = still?.closest<HTMLElement>(".s-frame-body");
	if (!still || !frame) {
		return false;
	}
	const bubble = frame.querySelector<HTMLElement>(".s-bubble");
	const live = document.createElement("div");
	live.className = "s-card";
	still.replaceWith(live);
	try {
		render(<HeroDemo data={data} frame={frame} bubble={bubble} />, live);
	} catch (error: unknown) {
		live.replaceWith(still);
		throw error;
	}
	// The frame takes the focus when a card it holds gives way to the next.
	frame.tabIndex = -1;
	return true;
}

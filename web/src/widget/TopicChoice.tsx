import { type ComponentChildren, createContext, type RefObject } from "preact";
import { useEffect, useRef, useState } from "preact/hooks";
import { classes } from "../design/classes";
import { Button } from "../design/controls";
import { Icon } from "../design/icons";
import { useScopedId } from "../design/ids";
import { type Linking, PageLink } from "../design/links";
import type { Host } from "./bridge";
import { lineWait, within } from "./ChatRequest";
import { useLinking } from "./linking";
import { groupAddress } from "./links";
import { groupName, topicName } from "./names";
import type { TopicChoice } from "./payload";
import { useService } from "./service";
import { topicGroups } from "./topicGroups";
import { type Key, useWords } from "./words";

/**
 * ChoosesTopic says whether a choice of the topic can be acted on where a card
 * is drawn: saved to the child's profile and followed by a task on the topic.
 * Where it cannot, the card shows the topic's button locked and makes no
 * choice, whatever is pressed.
 */
export const ChoosesTopic = createContext(true);

/**
 * Said is what the card says of the last choice sent: that it is on its way,
 * that it was not saved, or that it was not saved because the profile is not
 * there any more.
 */
type Said = "saving" | "not_saved" | "gone";

/**
 * Choosing is the choice of the topic as a card of a task holds it: the topic
 * the lessons are kept to, null while the coach chooses; whether the panel is
 * open; what became of the last choice sent; and what moves it — the panel
 * opened or closed, and a topic chosen, or the coach's choice. busy says, at
 * the moment it is asked, whether a choice is on its way to the service.
 */
export type Choosing = {
	chosen: string | null;
	open: boolean;
	said: Said | undefined;
	busy: () => boolean;
	toggle: () => void;
	close: () => void;
	pick: (topic: string | null) => void;
	button: RefObject<HTMLButtonElement | null>;
	panel: RefObject<HTMLElement | null>;
	panelId: string;
};

/**
 * useTopicChoice is the choice of the topic on a card that offers it: the
 * topic the lessons are kept to as the service last said, until the card
 * changes it. A choice is saved to the profile, the model is told of it in
 * the service's words through tell, and only then is the next task asked of
 * the chat in the adult's words, through ask, as another task is: a choice not
 * saved asks nothing, and says so. Nothing is chosen while busy says the card is in the
 * middle of something else, nor while a choice is on its way: a press of a
 * choice that looks locked is turned away here. The panel takes
 * the focus to the choice pressed in as it opens, and gives it back to the
 * button as it closes.
 */
export function useTopicChoice(
	offered: TopicChoice | undefined,
	{
		busy,
		tell,
		ask,
	}: {
		busy: () => boolean;
		tell: (line: string) => Promise<void>;
		ask: (words: string) => void;
	},
): Choosing {
	const words = useWords();
	const service = useService();
	const [state, setState] = useState<{
		chosen: string | null;
		open: boolean;
		said: Said | undefined;
	}>({ chosen: offered?.chosen ?? null, open: false, said: undefined });
	const saving = useRef(false);
	const button = useRef<HTMLButtonElement>(null);
	const panel = useRef<HTMLElement>(null);
	const panelId = useScopedId();

	useEffect(() => {
		if (state.open) {
			panel.current
				?.querySelector<HTMLElement>('[aria-pressed="true"]')
				?.focus({ preventScroll: true });
		}
	}, [state.open]);

	// What was said of a choice is said until the panel opens or closes, unless
	// the choice is still on its way. open is worked out from whether it is
	// open, so that two presses before the card redraws open it and close it.
	const opened = (open: (was: boolean) => boolean) =>
		setState((was) => ({
			...was,
			open: open(was.open),
			said: saving.current ? was.said : undefined,
		}));

	function close() {
		opened(() => false);
		button.current?.focus({ preventScroll: true });
	}

	async function pick(topic: string | null) {
		if (saving.current || busy()) {
			return;
		}
		saving.current = true;
		setState((was) => ({ ...was, said: "saving" }));
		const outcome = await service.saveEdit({ lesson_topic: topic ?? "" });
		if (outcome.kind !== "saved") {
			saving.current = false;
			setState((was) => ({
				...was,
				said: outcome.kind === "gone" ? "gone" : "not_saved",
			}));
			return;
		}
		// The model reads the line with the card's message, so the line goes to
		// the host first, and the message once the host has taken it — or a
		// moment later, since a host that keeps no line, or never says so, still
		// takes the message.
		if (outcome.told !== undefined) {
			await within(
				lineWait,
				tell(outcome.told).catch((error: unknown) => {
					console.error("widget: the model was not told of the topic", error);
				}),
			);
		}
		saving.current = false;
		setState({ chosen: topic, open: false, said: undefined });
		button.current?.focus({ preventScroll: true });
		ask(
			topic === null
				? words.text("topic_choice.ask_coach")
				: words.text("topic_choice.ask", { topic: topicName(words, topic) }),
		);
	}

	return {
		...state,
		busy: () => saving.current,
		toggle: () => opened((open) => !open),
		close,
		pick: (topic) => void pick(topic),
		button,
		panel,
		panelId,
	};
}

/**
 * TopicButton is the button at the end of a task's row that opens the choice
 * of the topic: a tag and the topic the lessons are kept to, or a spark and
 * that the coach chooses it, and whether the choice is open. A narrow card
 * shows the topic alone, or the coach in a word, and leaves "Topic:" to a
 * screen reader. Its words wrap rather than run out of a narrow card. It takes
 * no press while locked.
 */
export function TopicButton({
	choosing,
	wide,
	locked,
}: {
	choosing: Choosing;
	wide: boolean;
	locked: boolean;
}) {
	const words = useWords();
	const chosen = choosing.chosen;
	const coach = wide ? "topic_choice.coach" : "topic_choice.coach_short";
	return (
		<Button
			className="mt-topic-button"
			expanded={choosing.open}
			controls={choosing.panelId}
			locked={locked}
			onClick={choosing.toggle}
			buttonRef={choosing.button}
		>
			<Icon name={chosen === null ? "sparkle" : "tag"} size={20} />
			<span class="mt-topic-button-text">
				<span class={wide ? "mt-topic-button-label" : "mt-vh"}>
					{words.text("topic_choice.label")}
				</span>{" "}
				<span class="mt-topic-button-value">
					{chosen === null ? words.text(coach) : topicName(words, chosen)}
				</span>
			</span>
			<Icon name="chevron-right" size={16} className="mt-chevron" />
		</Button>
	);
}

// taughtFrom is the first grade each topic the card offers is taught from.
const taughtFrom: ReadonlyMap<string, number> = new Map(
	topicGroups.flatMap((group) =>
		group.topics.map((topic) => [topic.id, topic.fromGrade] as const),
	),
);

/**
 * TopicPanel is the choice of the topic, under the buttons of a task: the
 * coach's choice, the topics the review suggests, and every topic the card
 * knows in the groups of the site's page of topics, each group named by a
 * link to its part of that page where the chat opens links. A topic taught
 * from a grade above the child's is set apart, with the grade it is taught
 * from, and can be chosen all the same: it is a hint, not a bar. The choice
 * in force is pressed in, and none takes a press while locked. It is on the
 * page while closed, hidden and empty, so that the button that opens it can
 * always name it; Escape and its cross close it.
 */
export function TopicPanel({
	choosing,
	offered,
	grade,
	host,
	locked,
}: {
	choosing: Choosing;
	offered: TopicChoice;
	grade: number;
	host: Host;
	locked: boolean;
}) {
	const words = useWords();
	const linking = useLinking(host, words.text("progress.link_refused"));
	const title = useScopedId();
	const option = (topic: string) => {
		const from = taughtFrom.get(topic) ?? 0;
		return (
			<TopicOption
				key={topic}
				pressed={choosing.chosen === topic}
				locked={locked}
				older={
					from > grade
						? words.text("topic_choice.from_grade", { grade: from })
						: undefined
				}
				onPick={() => choosing.pick(topic)}
			>
				{topicName(words, topic)}
			</TopicOption>
		);
	};
	// A topic the card does not know is one it cannot place: it is not offered.
	const suggested = offered.recommended.filter((topic) =>
		taughtFrom.has(topic),
	);
	return (
		<section
			id={choosing.panelId}
			ref={choosing.panel}
			class="mt-topic-panel"
			aria-labelledby={title}
			hidden={!choosing.open}
			onKeyDown={(event) => {
				if (event.key === "Escape") {
					event.preventDefault();
					choosing.close();
				}
			}}
		>
			{choosing.open && (
				<>
					<div class="mt-topic-head">
						<div class="mt-topic-head-text">
							<h2 id={title} class="mt-topic-title">
								{words.text("topic_choice.title")}
							</h2>
							<p class="mt-topic-lead">{words.text("topic_choice.lead")}</p>
						</div>
						<button
							type="button"
							class="mt-topic-close"
							aria-label={words.text("topic_choice.close")}
							onClick={choosing.close}
						>
							<Icon name="cross" size={16} />
						</button>
					</div>
					<div class="mt-topic-coach">
						<TopicOption
							pressed={choosing.chosen === null}
							locked={locked}
							onPick={() => choosing.pick(null)}
						>
							{words.text("topic_choice.coach_option")}
						</TopicOption>
						<span class="mt-topic-coach-note">
							{words.text("topic_choice.coach_note")}
						</span>
					</div>
					<div class="mt-topic-rows">
						{suggested.length > 0 && (
							<TopicRow
								name={words.text("topic_choice.recommended")}
								className="mt-topic-row-suggested"
							>
								{suggested.map(option)}
							</TopicRow>
						)}
						{topicGroups.map((group) => (
							<TopicRow
								key={group.id}
								name={groupName(words, group.id)}
								href={groupAddress(offered.site, words.locale, group.id)}
								linking={linking}
							>
								{group.topics.map((topic) => option(topic.id))}
							</TopicRow>
						))}
					</div>
				</>
			)}
		</section>
	);
}

// TopicRow is a row of the panel's topics under its name: the topics the
// review suggests, or a group of the site's page of topics, its name a link
// to its part of the page where the card has an address for it and the chat
// opens links. A screen reader hears the topics as a group of that name.
function TopicRow({
	name,
	href,
	linking,
	className,
	children,
}: {
	name: string;
	href?: string;
	linking?: Linking;
	className?: string;
	children: ComponentChildren;
}) {
	return (
		// biome-ignore lint/a11y/useSemanticElements: a fieldset lays its legend out above its content, and a row on a wide card has its name in a column beside its topics
		<div
			class={classes("mt-topic-row", className)}
			role="group"
			aria-label={name}
		>
			<div class="mt-topic-row-name">
				{href !== undefined && linking !== undefined ? (
					<PageLink label={name} href={href} linking={linking} />
				) : (
					name
				)}
			</div>
			<div class="mt-topic-row-options">{children}</div>
		</div>
	);
}

// TopicOption is one choice of the panel, a button pressed in while it is the
// choice in force, and ticked so that it is seen to be; a topic taught from a
// grade above the child's says from which. A press chooses it, and asks for a
// task on it at once, so the options are buttons rather than a group of radio
// buttons, whose arrow keys would choose at every step. A locked option says
// so and keeps the focus it has; the choice turns its press away while the
// card is busy.
function TopicOption({
	pressed,
	locked,
	older,
	onPick,
	children,
}: {
	pressed: boolean;
	locked: boolean;
	older?: string;
	onPick: () => void;
	children: ComponentChildren;
}) {
	return (
		<button
			type="button"
			class="mt-topic-option"
			aria-pressed={pressed ? "true" : "false"}
			aria-disabled={locked ? "true" : undefined}
			data-older={older === undefined ? undefined : "true"}
			onClick={onPick}
		>
			{pressed && <Icon name="check" size={14} />}
			<span class="mt-topic-option-text">
				<span class="mt-topic-option-name">{children}</span>
				{older !== undefined && (
					<>
						{" "}
						<span class="mt-topic-from">{older}</span>
					</>
				)}
			</span>
		</button>
	);
}

/**
 * TopicChip is the mark above a task given on the topic the lessons are kept
 * to: the topic, and a cross that gives the choice back to the coach and asks
 * for a new task, as the coach's choice in the panel does. A locked cross says
 * so, and the choice turns its press away while the card is busy.
 */
export function TopicChip({
	choosing,
	topic,
	locked,
}: {
	choosing: Choosing;
	topic: string;
	locked: boolean;
}) {
	const words = useWords();
	return (
		<div class="mt-chip mt-chip-removable mt-topic-chip">
			<span class="mt-chip-text">
				<span class="mt-topic-chip-label">
					{words.text("topic_choice.label")}
				</span>{" "}
				<span class="mt-topic-chip-value">{topicName(words, topic)}</span>
			</span>
			<button
				type="button"
				class="mt-chip-remove"
				aria-label={words.text("topic_choice.clear")}
				aria-disabled={locked ? "true" : undefined}
				onClick={() => choosing.pick(null)}
			>
				<Icon name="cross" size={14} />
			</button>
		</div>
	);
}

// The words for what became of the last choice sent.
const saidKeys: Record<Said, Key> = {
	saving: "topic_choice.saving",
	not_saved: "topic_choice.not_saved",
	gone: "profile.gone",
};

/**
 * TopicNote says what became of the last choice sent: on its way, or not
 * saved. It is on the page before it says anything, so that a screen reader
 * hears it when it does.
 */
export function TopicNote({ said }: { said: Said | undefined }) {
	const words = useWords();
	return (
		<p class="mt-action-note mt-topic-note" aria-live="polite">
			{said === undefined ? "" : words.text(saidKeys[said])}
		</p>
	);
}

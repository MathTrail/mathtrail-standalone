import type { ComponentChildren } from "preact";
import type { Section } from "../widget/folds";
import { lessonStart } from "../widget/lesson";
import { trapName } from "../widget/names";
import type { HandedTask, ProgressReport } from "../widget/payload";
import { cardWords } from "../widget/words";
import { sectionAddress } from "./addresses";
import {
	chatGPTDeveloperModeURL,
	claudeConnectorsURL,
	connectorURL,
	sourceURL,
} from "./brand";
import { ChatFrame, ChatLines } from "./Chat";
import { frontPage } from "./content";
import { progressOf } from "./data";
import {
	connectAddress,
	connectSection,
	type Home,
	homeAnswerOf,
	homeTaskOf,
	lessonSection,
} from "./home";
import type { PageProps } from "./pages";
import type { PageReader } from "./reader";
import {
	StaticAnswer,
	StaticComing,
	StaticProgress,
	StaticTask,
} from "./StaticCard";
import type { WrongAnswer } from "./taskcard";
import { gradesOfAll } from "./topics";
import { gradesRange, gradesText, useSiteWords } from "./words";

// progressOpen are the sections of the progress the lesson's last step shows
// open, which its words tell of: where the child stands in each topic, and the
// mistakes that repeat.
const progressOpen: ReadonlySet<Section> = new Set(["topics", "mistakes"]);

/**
 * HomePage is the site's front page, for a parent: what MathTrail is, beside
 * the card of a task as a chat draws it; why asking the chat alone is not
 * enough; how a lesson goes, step by step, each step beside the card as it
 * stands at that step; how to connect MathTrail; and a last call to start.
 * Nothing on it runs: every card is drawn once, when the site is built. The
 * grades are the catalog's, the names of the traps and the cards' words the
 * widget's, and the lesson's task the site's data said in the page's words.
 */
export function HomePage({ page, data }: PageProps) {
	const home = data.home;
	if (home === undefined) {
		throw new Error(
			"site/data.json has nothing for the home page: the card of its lesson and the traps it names",
		);
	}
	const lesson = lessonOf(page, home);
	return (
		<>
			<Hero
				page={page}
				grades={gradesOfAll(data.topics)}
				handed={lesson.handed}
			/>
			<Pillars page={page} home={home} />
			<Steps
				page={page}
				home={home}
				lesson={lesson}
				progress={progressOf(data, lesson.handed.child.pseudonym)}
			/>
			<Connect page={page} />
			<Ask page={page} />
		</>
	);
}

// LessonCards are the lesson's task as its cards draw it: handed out, and
// answered with the option the steps pick.
type LessonCards = {
	readonly handed: HandedTask;
	readonly answered: WrongAnswer;
};

// lessonOf is the lesson's task of the site's data, said in the page's words.
function lessonOf(page: PageReader, home: Home): LessonCards {
	const said = {
		language: page.locale,
		child: page.plain("lesson.child"),
		question: page.plain("task.question"),
		hint: page.plain("task.hint"),
		trap: page.plain("task.trap"),
		solution: page.plain("task.solution"),
	};
	return {
		handed: homeTaskOf(home.card, said),
		answered: homeAnswerOf(home.card, said),
	};
}

// Hero is the first screen: what MathTrail is and for which grades, the way to
// add it and the way to see a lesson first, where it works, and beside them
// the card of a task as a chat draws it once the task has arrived.
function Hero({
	page,
	grades,
	handed,
}: {
	page: PageReader;
	grades: readonly [number, number];
	handed: HandedTask;
}) {
	const words = useSiteWords();
	return (
		<section class="s-wrap s-hero">
			<div class="s-hero-copy">
				<p class="s-chips">
					<span class="s-chip">{page.text("hero.free")}</span>
					<span class="s-chip">{page.text("hero.open")}</span>
					<span class="s-chip">{gradesText(words, grades)}</span>
				</p>
				<h1>{page.text("hero.title")}</h1>
				<p class="s-lead">
					{page.text("hero.lead", { range: gradesRange(page.locale, grades) })}
				</p>
				<p class="s-choices s-hero-actions">
					<a class="s-btn s-btn-filled" href={connectAddress(page.locale)}>
						{words.text("nav.add")}
					</a>
					<a
						class="s-btn"
						href={sectionAddress(page.locale, frontPage, lessonSection)}
					>
						{page.text("hero.see")}
					</a>
				</p>
				<p class="s-hero-note">{page.text("hero.note")}</p>
			</div>
			<figure class="s-panel s-hero-card">
				<Chat page={page}>
					<StaticTask
						handed={handed}
						start={lessonStart}
						locale={page.locale}
					/>
				</Chat>
				<figcaption>{page.text("hero.caption")}</figcaption>
			</figure>
		</section>
	);
}

// Chat is a card in the frame of the chat a parent already uses, under the
// parent's message, which is asking for a task unless ask says otherwise.
function Chat({
	page,
	ask = "chat.ask",
	children,
}: {
	page: PageReader;
	ask?: string;
	children: ComponentChildren;
}) {
	return (
		<ChatFrame title={page.text("chat.title")} ask={page.text(ask)}>
			{children}
		</ChatFrame>
	);
}

// Pillars are why asking the chat alone is not enough, and what MathTrail puts
// behind the chat's model, pillar by pillar. The traps a pillar names are the
// catalog's, under the names the card gives them.
function Pillars({ page, home }: { page: PageReader; home: Home }) {
	const card = cardWords(page.locale, undefined);
	const [first, second, third] = home.traps;
	const traps = {
		first: trapName(card, first),
		second: trapName(card, second),
		third: trapName(card, third),
	};
	const numbers = new Intl.NumberFormat(page.locale, {
		minimumIntegerDigits: 2,
	});
	return (
		<section class="s-wrap s-section">
			<div class="s-intro">
				<h2>{page.text("pillars.title")}</h2>
				<p class="s-intro-line">{page.text("pillars.lead")}</p>
			</div>
			<ol class="s-tiles">
				{page.list("pillars.items").map((key, at) => (
					<li key={key} class="s-tile">
						<span class="s-tip-number" aria-hidden="true">
							{numbers.format(at + 1)}
						</span>
						<h3>{page.text(`${key}.title`)}</h3>
						<p class="s-tile-text">{page.text(`${key}.text`, traps)}</p>
					</li>
				))}
			</ol>
		</section>
	);
}

// Steps is how a lesson goes, a step at a time, each beside the card as a chat
// draws it at that step: the task being written; an option picked and being
// checked; the hint open, a step within answering; the wrong answer told, with
// its trap and the solution; the question the child asks in the chat once the
// answer is in, a step within that; and the progress. Then the one thing a
// parent should know while the child solves.
function Steps({
	page,
	home,
	lesson,
	progress,
}: {
	page: PageReader;
	home: Home;
	lesson: LessonCards;
	progress: ProgressReport;
}) {
	const { handed, answered } = lesson;
	const { locale } = page;
	return (
		<section id={lessonSection} class="s-wrap s-section">
			<div class="s-intro">
				<p class="s-eyebrow">{page.text("lesson.eyebrow")}</p>
				<h2>{page.text("lesson.title")}</h2>
				<p class="s-intro-line">{page.text("lesson.lead")}</p>
			</div>
			<ol class="s-walk">
				<Step page={page} name="generating" number={1}>
					<Chat page={page}>
						<StaticComing
							coming={{ requestId: handed.task.id, child: handed.child }}
							locale={locale}
						/>
					</Chat>
				</Step>
				<Step page={page} name="selected" number={2}>
					<Chat page={page}>
						<StaticTask
							handed={handed}
							start={{
								...lessonStart,
								answer: { state: "checking", choice: home.card.choice },
							}}
							locale={locale}
						/>
					</Chat>
				</Step>
				<Step page={page} name="hint">
					<Chat page={page}>
						<StaticTask
							handed={handed}
							start={{ ...lessonStart, hint: { open: true, used: true } }}
							locale={locale}
						/>
					</Chat>
				</Step>
				<Step page={page} name="wrong" number={3}>
					<Chat page={page}>
						<StaticAnswer
							handed={answered.handed}
							result={answered.result}
							locale={locale}
						/>
					</Chat>
				</Step>
				<Step page={page} name="question">
					<Chat page={page}>
						<StaticAnswer
							handed={answered.handed}
							result={answered.result}
							locale={locale}
						/>
						<ChatLines
							label={page.text("lesson.chat.label")}
							lines={[
								{
									from: "child",
									speaker: page.text("lesson.chat.child"),
									said: page.text("lesson.chat.question"),
								},
								{
									from: "model",
									speaker: page.text("lesson.chat.model"),
									said: page.text("lesson.chat.reply"),
								},
							]}
						/>
					</Chat>
				</Step>
				<Step page={page} name="progress" number={4}>
					<Chat page={page} ask="chat.progress">
						<StaticProgress
							report={progress}
							locale={locale}
							open={progressOpen}
						/>
					</Chat>
				</Step>
			</ol>
			<aside class="s-note s-walk-note">
				<p class="s-note-text">{page.text("lesson.log")}</p>
			</aside>
		</section>
	);
}

// Step is one step of the lesson, by the name its words have, beside its card:
// numbered, or, with no number, a step within the one before it.
function Step({
	page,
	name,
	number,
	children,
}: {
	page: PageReader;
	name: string;
	number?: number;
	children: ComponentChildren;
}) {
	const at = `lesson.steps.${name}`;
	return (
		<li
			class={number === undefined ? "s-walk-step s-walk-within" : "s-walk-step"}
		>
			<div class="s-walk-copy">
				<span class="s-walk-number" aria-hidden="true">
					{number === undefined
						? ""
						: new Intl.NumberFormat(page.locale).format(number)}
				</span>
				<h3>{page.text(`${at}.title`)}</h3>
				<p class="s-tile-text">{page.text(`${at}.text`)}</p>
			</div>
			<div class="s-walk-card">{children}</div>
		</li>
	);
}

// Connect is how to connect MathTrail: to Claude step by step, the address to
// paste among the steps, in full and selected whole with one click; to
// ChatGPT, where it is not listed yet; and what accounts it takes.
function Connect({ page }: { page: PageReader }) {
	const numbers = new Intl.NumberFormat(page.locale);
	return (
		<section id={connectSection} class="s-wrap s-section">
			<div class="s-intro">
				<p class="s-eyebrow">{page.text("connect.eyebrow")}</p>
				<h2>{page.text("connect.title")}</h2>
			</div>
			<div class="s-connect">
				<div class="s-connect-panel">
					<div class="s-connect-head">
						<h3>Claude</h3>
						<span class="s-pill s-pill-ok">
							{page.text("connect.claude.pill")}
						</span>
					</div>
					<ol class="s-connect-steps">
						{claudeSteps.map((name, at) => (
							<li key={name}>
								<span class="s-connect-number" aria-hidden="true">
									{numbers.format(at + 1)}
								</span>
								<div class="s-connect-step">
									<p class="s-connect-text">
										{page.text(`connect.claude.steps.${name}`, {
											connectors: (
												<a href={claudeConnectorsURL}>
													{page.text("connect.claude.connectors")}
												</a>
											),
										})}
									</p>
									{name === "paste" && (
										<code class="s-address" dir="ltr">
											{connectorURL}
										</code>
									)}
								</div>
							</li>
						))}
					</ol>
					<p class="s-connect-note">{page.text("connect.claude.note")}</p>
				</div>
				<div class="s-connect-panel s-connect-muted">
					<div class="s-connect-head">
						<h3>ChatGPT</h3>
						<span class="s-pill">{page.text("connect.chatgpt.pill")}</span>
					</div>
					<p class="s-tile-text">{page.text("connect.chatgpt.text")}</p>
					<p class="s-tile-text">
						{page.text("connect.chatgpt.dev", {
							openai: (
								<a href={chatGPTDeveloperModeURL}>
									{page.text("connect.chatgpt.openai")}
								</a>
							),
						})}
					</p>
				</div>
			</div>
			<p class="s-connect-need">{page.text("connect.need")}</p>
		</section>
	);
}

// claudeSteps are the steps of connecting to Claude, by the names their words
// have, in their order: open the connectors, add one, paste the address —
// which follows the words of that step — sign in, and switch it on in a chat.
const claudeSteps = ["open", "add", "paste", "sign", "chat"] as const;

// Ask is the last call: add MathTrail to Claude, or read its code first.
function Ask({ page }: { page: PageReader }) {
	const words = useSiteWords();
	return (
		<section class="s-section s-ask-wrap">
			<div class="s-ask">
				<h2 class="s-ask-title">{page.text("ask.title")}</h2>
				<p class="s-ask-lead">{page.text("ask.lead")}</p>
				<p class="s-choices">
					<a class="s-btn s-btn-filled" href={connectAddress(page.locale)}>
						{words.text("nav.add")}
					</a>
					<a class="s-btn" href={sourceURL}>
						{page.text("ask.code")}
					</a>
				</p>
			</div>
		</section>
	);
}

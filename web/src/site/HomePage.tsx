import type { ComponentChildren } from "preact";
import { type DemoData, DemoDataScript } from "../demo/data";
import { dictionaries } from "../widget/dictionaries";
import type { Section } from "../widget/folds";
import { lessonStart } from "../widget/lesson";
import type { HandedTask, ProgressReport } from "../widget/payload";
import { sectionAddress } from "./addresses";
import {
	chatGPTDeveloperModeURL,
	claudeConnectorsURL,
	connectorURL,
	sourceURL,
} from "./brand";
import { AdultAsks, ChatFrame } from "./Chat";
import { frontPage } from "./content";
import { progressOf } from "./data";
import { Pick } from "./HomePick";
import { Why } from "./HomeWhy";
import {
	connectAddress,
	connectSection,
	type Home,
	type HomeResults,
	type HomeWords,
	homeAnswerOf,
	homeResultsOf,
	homeTaskOf,
	lessonSection,
} from "./home";
import { Phone } from "./Phone";
import type { PageProps } from "./pages";
import type { PageReader } from "./reader";
import { GitHubMark, SourceChip } from "./SourceChip";
import {
	StaticComing,
	StaticProgress,
	StaticResult,
	StaticTask,
} from "./StaticCard";
import type { WrongAnswer } from "./taskcard";
import { gradesOfAll } from "./topics";
import { gradesRange, gradesText, useSiteWords } from "./words";

// progressOpen are the sections of the progress the lesson's last step shows
// open, which its words tell of: where the child stands in each topic, and the
// mistakes that repeat.
const progressOpen: ReadonlySet<Section> = new Set(["topics", "mistakes"]);

// demoScript is where the home page's demo is served: one module, which the
// site's build makes of web/src/demo.
const demoScript = "/assets/demo.js";

/**
 * HomePage is the site's front page, for a parent: what MathTrail is and how
 * it picked the task, beside the card of that task as a chat draws it, in a
 * phone; how a lesson goes, step by step, each step beside the card as it
 * stands at that step; why asking the chat alone is not enough; how to
 * connect MathTrail; and a last call to start. Every card is drawn when the
 * site is built, and the page reads in
 * full as it is drawn. Its demo, when it runs, brings the card on the first
 * screen alive and, on a wide window, moves the chat in the phone with the
 * page's scroll; it gives the steps one card on a wide window and the
 * connector's address a button that copies it; the page carries what the
 * demo needs, so that it loads nothing else. The grades are the catalog's,
 * the names of the traps and the cards' words the widget's, and the lesson's
 * task the site's data said in the page's words.
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
				home={home}
			/>
			<Steps
				page={page}
				home={home}
				lesson={lesson}
				progress={progressOf(data, lesson.handed.child.pseudonym)}
			/>
			<Why page={page} home={home} />
			<Connect page={page} />
			<Ask page={page} />
			<DemoDataScript data={demoOf(page, lesson)} />
			<script type="module" src={demoScript} />
		</>
	);
}

// demoOf is what the demo needs to bring the card on the first screen alive:
// the widget's words in the page's language alone, the lesson's task, and
// what the service would record of every answer to it.
function demoOf(page: PageReader, lesson: LessonCards): DemoData {
	const words = dictionaries.get(page.locale);
	if (words === undefined) {
		throw new Error(
			`the widget has no words in ${page.locale}, the language of the home page its demo speaks`,
		);
	}
	return {
		locale: page.locale,
		words,
		handed: lesson.handed,
		results: lesson.results,
	};
}

// LessonCards are the lesson's task as its cards draw it: handed out, and
// answered with the option the steps pick; and what the service would record
// of every answer to it, which the demo answers with.
type LessonCards = {
	readonly handed: HandedTask;
	readonly answered: WrongAnswer;
	readonly results: HomeResults;
};

// lessonOf is the lesson's task of the site's data, said in the page's words.
function lessonOf(page: PageReader, home: Home): LessonCards {
	const said: HomeWords = {
		language: page.locale,
		child: page.plain("lesson.child"),
		question: page.plain("task.question"),
		hint: page.plain("task.hint"),
		traps: Object.fromEntries(
			Object.keys(home.card.traps).map((letter) => [
				letter,
				page.plain(`task.traps.${letter.toLowerCase()}`),
			]),
		),
		solution: page.plain("task.solution"),
	};
	return {
		handed: homeTaskOf(home.card, said),
		answered: homeAnswerOf(home.card, said),
		results: homeResultsOf(home.card, said),
	};
}

// Hero is the first screen: what MathTrail is and for which grades, how it
// picked the task on the card, the way to add it and the way to see a lesson
// first, and a line for the parent while the child solves; beside them the
// card of that task as a chat draws it once the task has arrived, in a phone.
// The track around it is the room the page's scroll runs through while the
// demo holds the first screen in place and moves the chat in the phone.
function Hero({
	page,
	grades,
	handed,
	home,
}: {
	page: PageReader;
	grades: readonly [number, number];
	handed: HandedTask;
	home: Home;
}) {
	const words = useSiteWords();
	return (
		<div class="s-hero-track">
			<section class="s-wrap s-hero">
				<div class="s-hero-copy">
					<p class="s-chips">
						<span class="s-chip s-chip-free">{page.text("hero.free")}</span>
						<SourceChip label={page.text("hero.open")} />
						<span class="s-chip">{gradesText(words, grades)}</span>
					</p>
					<h1>{page.text("hero.title")}</h1>
					<p class="s-lead">
						{page.text("hero.lead", {
							range: gradesRange(page.locale, grades),
						})}
					</p>
					<Pick page={page} pick={home.pick} picked={home.card.topic} />
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
					<p class="s-hero-note s-tea">
						<span class="s-cup" aria-hidden="true" />
						{page.text("hero.note")}
					</p>
				</div>
				<figure class="s-hero-card s-hero-phone">
					<Phone>
						<Chat page={page}>
							<StaticTask
								handed={handed}
								start={lessonStart}
								locale={page.locale}
							/>
						</Chat>
					</Phone>
				</figure>
			</section>
		</div>
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

// Steps is how a lesson goes, a step at a time, each beside the card as a chat
// draws it at that step: the task being written; an option picked and being
// checked; the hint open, a step within answering; the card turned into how
// the wrong answer went, with its trap and the solution; the question the
// child asks in the chat once the answer is in, a step within that; and the
// progress. Then the one thing a parent should know while the child solves.
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
						<StaticResult
							handed={answered.handed}
							result={answered.result}
							locale={locale}
						/>
					</Chat>
				</Step>
				<Step page={page} name="question">
					<Chat page={page}>
						<StaticResult
							handed={answered.handed}
							result={answered.result}
							locale={locale}
						/>
						<AdultAsks page={page} at="lesson.chat" />
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
										<>
											<code class="s-address" dir="ltr">
												{connectorURL}
											</code>
											<CopyButton />
										</>
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

// CopyButton is the button that copies the address before it, kept in a
// template: the demo puts it beside the address where the browser lets a page
// copy, and a page read without the demo shows no button that does nothing.
function CopyButton() {
	const words = useSiteWords();
	return (
		<template data-copy="">
			<button type="button" class="s-copy" data-done={words.text("copy.done")}>
				{words.text("copy.action")}
			</button>
			<span class="s-hidden" aria-live="polite" />
		</template>
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
			<div class="s-ask s-ask-row">
				<div class="s-ask-copy">
					<h2 class="s-ask-title">{page.text("ask.title")}</h2>
					<p class="s-ask-lead">{page.text("ask.lead")}</p>
				</div>
				<p class="s-choices">
					<a class="s-btn s-btn-filled" href={connectAddress(page.locale)}>
						{words.text("nav.add")}
					</a>
					<a class="s-btn s-btn-github" href={sourceURL}>
						<GitHubMark size={20} />
						{page.text("ask.code")}
					</a>
				</p>
			</div>
		</section>
	);
}

import type { ComponentChildren, VNode } from "preact";
import { cardWords } from "../widget/dictionaries";
import { topicName, trapName } from "../widget/names";
import type { Home, HomeAlone } from "./home";
import type { Fill, PageReader } from "./reader";

/**
 * Why is why asking the chat alone is not enough: the words beside a chat
 * asked alone for a task, which writes one with two right options; the chat
 * alone against the chat with MathTrail, point against point; what MathTrail
 * puts behind the chat's model, tile by tile; and what stands behind every
 * task. The traps a tile names are the catalog's, under the names the card
 * gives them, and so are the topics of the map's tile; the languages are
 * named each in itself.
 */
export function Why({ page, home }: { page: PageReader; home: Home }) {
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
	const [base, otherBase] = home.map.bases;
	return (
		<section class="s-wrap s-section">
			<div class="s-pillars-head">
				<div class="s-intro">
					<h2>{page.text("pillars.title")}</h2>
					<p class="s-intro-line">{page.text("pillars.lead")}</p>
				</div>
				<ChatAlone page={page} alone={home.alone} />
			</div>
			<Against page={page} />
			<ol class="s-pillars">
				<Tile
					page={page}
					at="pillars.tiles.endless"
					icon="endless"
					number={numbers.format(1)}
				>
					{languagesFirst(page.locale, home.languages).map((tag) => (
						<span key={tag} class="s-mini" lang={tag}>
							{nameOf(tag)}
						</span>
					))}
					<span class="s-mini">{page.text("pillars.tiles.endless.more")}</span>
				</Tile>
				<Tile
					page={page}
					at="pillars.tiles.checked"
					icon="checked"
					number={numbers.format(2)}
				>
					{page.list("pillars.tiles.checked.checks").map((key) => (
						<span key={key} class="s-mini">
							<span aria-hidden="true">✓ </span>
							{page.text(key)}
						</span>
					))}
				</Tile>
				<Tile
					page={page}
					at="pillars.tiles.diagnosis"
					icon="diagnosis"
					number={numbers.format(3)}
					slots={traps}
				>
					<span class="s-mini s-mini-trap">
						{page.text("pillars.tiles.diagnosis.trap")}
					</span>
				</Tile>
			</ol>
			<h3 class="s-pillars-band">{page.text("pillars.behind.title")}</h3>
			<ul class="s-pillars s-pillars-plain">
				<Tile page={page} at="pillars.behind.coach" icon="coach" heading="h4">
					<span class="s-mini-said">{page.text("chat.ask")}</span>
					<Arrow />
					<span class="s-mini">{page.text("pillars.behind.coach.picked")}</span>
				</Tile>
				<Tile
					page={page}
					at="pillars.behind.research"
					icon="research"
					heading="h4"
				>
					<code class="s-mini s-mini-formula">
						{page.text("pillars.behind.research.formula")}
					</code>
				</Tile>
				<Tile page={page} at="pillars.behind.map" icon="map" heading="h4">
					<span class="s-mini">{topicName(card, base)}</span>
					<Arrow />
					<span class="s-mini s-mini-topic">
						{topicName(card, home.map.topic)}
					</span>
					<Arrow back />
					<span class="s-mini">{topicName(card, otherBase)}</span>
				</Tile>
			</ul>
		</section>
	);
}

// ChatAlone is a chat asked alone for a task: the parent's message, the task
// it writes, with its options and those that are right — more than one, which
// nothing caught — ringed, and a flag that says so.
function ChatAlone({ page, alone }: { page: PageReader; alone: HomeAlone }) {
	return (
		<figure class="s-alone">
			<figcaption class="s-alone-bar">
				{page.text("pillars.alone.name")}
			</figcaption>
			<div class="s-alone-body">
				<p class="s-alone-ask">{page.text("pillars.alone.ask")}</p>
				<div class="s-alone-task">
					<p>{page.text("pillars.alone.task")}</p>
					<ul class="s-alone-options">
						{Object.entries(alone.options).map(([letter, value]) => (
							<li
								key={letter}
								data-right={alone.right.includes(letter) ? "" : undefined}
							>
								<b>{letter}</b> {value}
							</li>
						))}
					</ul>
				</div>
				<p class="s-alone-flag">
					<span aria-hidden="true">✕ </span>
					{page.text("pillars.alone.flag")}
				</p>
			</div>
		</figure>
	);
}

// Against is the chat alone against the chat with MathTrail, both asked for the
// same task, point against point: what goes wrong without MathTrail, crossed
// out, and what MathTrail does about it, ticked.
function Against({ page }: { page: PageReader }) {
	const asked = () =>
		page.text("pillars.versus.asked", {
			ask: page.plain("pillars.alone.ask"),
		});
	return (
		<div class="s-against">
			<Side
				page={page}
				kind="alone"
				name={page.text("pillars.alone.name")}
				asked={asked()}
				mark="✕"
				points="pillars.versus.alone"
			/>
			<Side
				page={page}
				kind="ours"
				name={page.text("pillars.versus.ours.name")}
				asked={asked()}
				mark="✓"
				points="pillars.versus.ours.points"
			/>
		</div>
	);
}

// Side is one side of the comparison: its name, what it was asked, and its
// points, the items of the page's list under the key given, each after a mark
// that says whether it counts against or for.
function Side({
	page,
	kind,
	name,
	asked,
	mark,
	points,
}: {
	page: PageReader;
	kind: "alone" | "ours";
	name: ComponentChildren;
	asked: ComponentChildren;
	mark: string;
	points: string;
}) {
	return (
		<div class={`s-against-side s-against-${kind}`}>
			<p class="s-against-head">
				<span class="s-against-tag">{name}</span>
				<span class="s-against-ask">{asked}</span>
			</p>
			<ul class="s-against-points">
				{page.list(points).map((point) => (
					<li key={point}>
						<span class="s-against-mark" aria-hidden="true">
							{mark}
						</span>
						<span>{page.text(point)}</span>
					</li>
				))}
			</ul>
		</div>
	);
}

// Tile is one tile of the section, by the place of its words: its mark and,
// when it has one, its number; its heading and its text, the slots of the
// text filled; and at its foot, the short things that show what it says.
function Tile({
	page,
	at,
	icon,
	number,
	heading: Heading = "h3",
	slots,
	children,
}: {
	page: PageReader;
	at: string;
	icon: TileIcon;
	number?: string;
	heading?: "h3" | "h4";
	slots?: Readonly<Record<string, Fill>>;
	children: ComponentChildren;
}) {
	return (
		<li class="s-pillar">
			<div class="s-pillar-top">
				<span class="s-pillar-icon">
					<svg
						viewBox="0 0 24 24"
						width="22"
						height="22"
						fill="none"
						stroke="currentColor"
						stroke-width="1.8"
						stroke-linecap="round"
						stroke-linejoin="round"
						aria-hidden="true"
					>
						{tileIcons[icon]()}
					</svg>
				</span>
				{number !== undefined && (
					<span class="s-pillar-number" aria-hidden="true">
						{number}
					</span>
				)}
			</div>
			<Heading>{page.text(`${at}.title`)}</Heading>
			<p class="s-pillar-text">{page.text(`${at}.text`, slots)}</p>
			<p class="s-pillar-mini">{children}</p>
		</li>
	);
}

// Arrow points from one short thing to the next, the way the language reads,
// or back against it.
function Arrow({ back = false }: { back?: boolean }) {
	return (
		<span class="s-mini-arrow" aria-hidden="true">
			{back ? "←" : "→"}
		</span>
	);
}

// TileIcon is the name of a tile's mark.
type TileIcon =
	| "endless"
	| "checked"
	| "diagnosis"
	| "coach"
	| "research"
	| "map";

// tileIcons are the tiles' marks, drawn in strokes on a 24-unit square: a loop
// with no end, a shield with a tick, a warning sign; a person, a flask, and
// three topics joined, two below the one above them.
const tileIcons: Readonly<Record<TileIcon, () => VNode>> = {
	endless: () => (
		<path d="M5 12c0-2.2 1.8-4 4-4 3 0 4 8 7 8 2.2 0 4-1.8 4-4s-1.8-4-4-4c-3 0-4 8-7 8-2.2 0-4-1.8-4-4z" />
	),
	checked: () => (
		<>
			<path d="M9 12l2 2 4-4" />
			<path d="M12 3l7 3v6c0 4.5-3 7.5-7 9-4-1.5-7-4.5-7-9V6z" />
		</>
	),
	diagnosis: () => (
		<>
			<path d="M12 3 2.5 20h19z" />
			<path d="M12 10v4M12 17h.01" />
		</>
	),
	coach: () => (
		<>
			<path d="M12 12a4 4 0 1 0 0-8 4 4 0 0 0 0 8z" />
			<path d="M4 21c1.5-4 4.5-6 8-6s6.5 2 8 6" />
		</>
	),
	research: () => (
		<>
			<path d="M9 3h6M10 3v6l-5 9a2 2 0 0 0 1.7 3h10.6a2 2 0 0 0 1.7-3l-5-9V3" />
			<path d="M7.5 15h9" />
		</>
	),
	map: () => (
		<>
			<circle cx="6" cy="6" r="2.5" />
			<circle cx="18" cy="6" r="2.5" />
			<circle cx="12" cy="18" r="2.5" />
			<path d="M7.8 7.8 10.6 16M16.2 7.8 13.4 16" />
		</>
	),
};

// languagesFirst are the languages a tile names, the page's own first and the
// rest in the order the site's data gives them.
function languagesFirst(
	locale: string,
	languages: readonly string[],
): string[] {
	const own = languages.filter((tag) => tag === locale);
	return [...own, ...languages.filter((tag) => tag !== locale)];
}

// nameOf is a language's name in the language itself, begun with a capital,
// as a list of languages to choose from writes it.
function nameOf(tag: string): string {
	const name =
		new Intl.DisplayNames([tag], { type: "language" }).of(tag) ?? tag;
	return name.charAt(0).toLocaleUpperCase(tag) + name.slice(1);
}

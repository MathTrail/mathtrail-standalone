import type { ComponentChildren } from "preact";
import { profileFileName } from "./brand";
import type { PageReader } from "./reader";
import { Cite, Stroke } from "./WhyParts";
import type { Hue, Source } from "./why";

// appCards are the cards of how MathTrail differs from an ordinary app, by
// the names their words have, each with the colour it is lit in: a program
// checks every problem, it is free and plainly why, the data stays with the
// family, and nothing keeps a child at the screen.
const appCards: readonly (readonly [string, Hue])[] = [
	["checked", "blue"],
	["free", "green"],
	["data", "violet"],
	["screen", "amber"],
];

/**
 * App is why MathTrail is no ordinary app: what research found of the apps
 * children play, beside a make-believe app of that kind with every trick it
 * plays and the line that MathTrail has none of them; then card by card how
 * MathTrail differs, each with the short things at its foot that show it.
 */
export function App({ page, apps }: { page: PageReader; apps: Source }) {
	return (
		<section class="s-wrap s-section">
			<div class="s-why-app-head">
				<div class="s-intro">
					<h2>{page.text("app.title")}</h2>
					<p class="s-intro-line">
						{page.text("app.lead", { source: <Cite source={apps} /> })}
					</p>
				</div>
				<Bait page={page} />
			</div>
			<ul class="s-why-app-cards">
				{appCards.map(([card, hue]) => (
					<li key={card} class="s-why-app-card" data-hue={hue}>
						<span class="s-why-badge">
							<Stroke size={22}>{cardMarks[card]?.()}</Stroke>
						</span>
						<h3>{page.text(`app.cards.${card}.title`)}</h3>
						<p class="s-why-tile-text">{page.text(`app.cards.${card}.text`)}</p>
						<p class="s-why-app-foot">
							<CardFoot page={page} card={card} />
						</p>
					</li>
				))}
			</ul>
		</section>
	);
}

// CardFoot is what stands at the foot of a card: the checks a problem passes
// and what becomes of one that fails; nothing to pay and all that is not
// charged for; the file a profile is and what the server keeps; and all that
// does not hold a child, then what a solved problem leads to.
function CardFoot({ page, card }: { page: PageReader; card: string }) {
	const at = `app.cards.${card}`;
	switch (card) {
		case "checked":
			return (
				<>
					{page.list(`${at}.checks`).map((key) => (
						<span key={key} class="s-why-foot-chip s-why-foot-yes">
							<span aria-hidden="true">✓ </span>
							{page.text(key)}
						</span>
					))}
					<span class="s-why-foot-chip s-why-foot-no">
						{page.text(`${at}.fails`)}
					</span>
				</>
			);
		case "free":
			return (
				<>
					<span class="s-why-foot-price">{page.text(`${at}.price`)}</span>
					<Struck page={page} at={`${at}.none`} />
				</>
			);
		case "data":
			return (
				<>
					<span class="s-why-foot-chip">
						<code>{profileFileName}</code> · {page.text(`${at}.where`)}
					</span>
					<span class="s-why-foot-chip s-why-foot-strong">
						{page.text(`${at}.server`)}
					</span>
				</>
			);
		case "screen":
			return (
				<>
					<Struck page={page} at={`${at}.none`} />
					<span class="s-why-foot-chip s-why-foot-strong">
						{page.text(`${at}.after`)}
					</span>
				</>
			);
		default:
			return null;
	}
}

// Struck is every thing of the list under at, each crossed out.
function Struck({ page, at }: { page: PageReader; at: string }) {
	return (
		<>
			{page.list(at).map((key) => (
				<span key={key} class="s-why-foot-chip">
					<s>{page.text(key)}</s>
				</span>
			))}
		</>
	);
}

// cardMarks are the cards' marks: a shield with a tick, a dollar, a padlock
// and a lamp.
const cardMarks: Readonly<Record<string, () => ComponentChildren>> = {
	checked: () => (
		<>
			<path d="M12 3l7 3v6c0 4.5-3 7.5-7 9-4-1.5-7-4.5-7-9V6z" />
			<path d="M9 12l2 2 4-4" />
		</>
	),
	free: () => (
		<path d="M12 2v20M17 5H9.5a3.5 3.5 0 0 0 0 7h5a3.5 3.5 0 0 1 0 7H6" />
	),
	data: () => (
		<>
			<rect x="4" y="11" width="16" height="10" rx="2" />
			<path d="M8 11V7a4 4 0 0 1 8 0v4" />
		</>
	),
	screen: () => <path d="M3 11h18M5 11V8a7 7 0 0 1 14 0v3M12 15v4M8 21h8" />,
};

// Bait is a make-believe app of the kind that keeps a child at the screen:
// coins, gems and a timer to a bonus, a trophy for one sum solved, a chest to
// open, a hat for sale and premium to buy. It is a drawing a screen reader
// skips; the line under it says what it is for.
function Bait({ page }: { page: PageReader }) {
	const at = "app.bait";
	return (
		<figure class="s-why-bait">
			<div class="s-why-bait-app" aria-hidden="true">
				<p class="s-why-bait-row">
					<span class="s-why-bait-pill">
						<span class="s-why-bait-coin" />
						{page.text(`${at}.coins`)}
					</span>
					<span class="s-why-bait-pill">
						<span class="s-why-bait-gem" />
						{page.text(`${at}.gems`)}
					</span>
					<span class="s-why-bait-pill s-why-bait-timer">
						⏱ {page.text(`${at}.timer`)}
					</span>
				</p>
				<div class="s-why-bait-prize">
					<span class="s-why-bait-cup">
						<span />
						<span />
						<span />
						<span />
						<span />
					</span>
					<span class="s-why-bait-cheer">{page.text(`${at}.cheer`)}</span>
					<span class="s-why-bait-praise">{page.text(`${at}.praise`)}</span>
				</div>
				<p class="s-why-bait-shop">
					<span class="s-why-bait-item s-why-bait-new">
						<span class="s-why-bait-chest">
							<span />
							<span />
							<span />
							<span />
						</span>
						{page.text(`${at}.chest`)}
					</span>
					<span class="s-why-bait-item">
						<span class="s-why-bait-hat">
							<span />
							<span />
						</span>
						{page.text(`${at}.hat`)}
					</span>
				</p>
				<p class="s-why-bait-buy">{page.text(`${at}.premium`)}</p>
			</div>
			<figcaption>{page.text(`${at}.caption`)}</figcaption>
		</figure>
	);
}

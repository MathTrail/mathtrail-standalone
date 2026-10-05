import { address } from "./addresses";
import { coachPrototypePath } from "./brand";
import type { SiteData } from "./data";
import type { PageProps } from "./pages";
import type { PageReader } from "./reader";
import { allTechniques } from "./techniques";
import { useSiteWords } from "./words";

/**
 * prototypeSection is the anchor of the part of the page that frames the
 * prototype, which the first screen's button leads to.
 */
export const prototypeSection = "prototype";

/**
 * CoachPage is the page of the coach: a product of its own, still in the
 * making, that leads a child through a problem on a tablet. It says what the
 * coach is, frames its prototype for a reader to try, and says what the
 * prototype only pretends to do yet. The page is the site's own and reads in
 * full without a script; the prototype in its frame is the export of the tool
 * it was drawn in, and speaks English alone in every language of the page.
 */
export function CoachPage({ page, data }: PageProps) {
	return (
		<>
			<Hero page={page} techniques={techniqueCount(data)} />
			<Ideas page={page} />
			<Prototype page={page} />
			<Notes page={page} />
		</>
	);
}

// techniqueCount is how many techniques the page of the techniques teaches,
// which the first screen's second button names.
function techniqueCount(data: SiteData): number {
	if (data.techniques === undefined) {
		throw new Error("site/data.json gives the page of the coach no techniques");
	}
	return allTechniques(data.techniques).length;
}

// Hero is the first screen: a badge that says the coach is in the making, the
// heading and what the coach does, then the way to its prototype further down
// and to the techniques it plans with.
function Hero({ page, techniques }: { page: PageReader; techniques: number }) {
	const words = useSiteWords();
	return (
		<section class="s-wrap s-hero s-coach-hero">
			<div class="s-hero-copy">
				<p class="s-badge">{page.text("hero.badge")}</p>
				<h1>{page.text("hero.title")}</h1>
				<p class="s-lead">{page.text("hero.lead")}</p>
				<p class="s-choices s-hero-actions">
					<a class="s-btn s-btn-filled" href={`#${prototypeSection}`}>
						{page.text("hero.try")}
					</a>
					<a class="s-btn" href={address(page.locale, "techniques")}>
						{words.text("techniques.count", { count: techniques })}
					</a>
				</p>
			</div>
		</section>
	);
}

// Ideas is what the coach is about, a card for each idea in the order the
// words give. The section's heading is for a screen reader, which moves from
// card to card by the headings under it.
function Ideas({ page }: { page: PageReader }) {
	return (
		<section class="s-wrap" aria-labelledby="ideas">
			<h2 id="ideas" class="s-hidden">
				{page.text("ideas.title")}
			</h2>
			<ul class="s-tiles s-coach-ideas">
				{page.list("ideas.items").map((key) => (
					<li key={key} class="s-tile">
						<h3>{page.text(`${key}.title`)}</h3>
						<p class="s-tile-text">{page.text(`${key}.text`)}</p>
					</li>
				))}
			</ul>
		</section>
	);
}

// Prototype frames the coach's prototype on the paper it is drawn on, with a
// line under it and a link that opens it alone, larger, in a tab of its own.
// The frame loads only once a reader comes near it, since the prototype
// weighs more than the rest of the page together.
function Prototype({ page }: { page: PageReader }) {
	return (
		<section id={prototypeSection} class="s-wrap s-section s-prototype-section">
			<div class="s-paper">
				<div class="s-prototype">
					<iframe
						src={coachPrototypePath}
						title={page.plain("prototype.frame")}
						loading="lazy"
					/>
				</div>
				<p class="s-prototype-caption">
					<span>{page.text("prototype.caption")}</span>
					<a href={coachPrototypePath} target="_blank" rel="noopener">
						{page.text("prototype.open")}
					</a>
				</p>
			</div>
		</section>
	);
}

// Notes says what the prototype is and what it only pretends to do yet, so
// that nobody takes a pretence for a promise.
function Notes({ page }: { page: PageReader }) {
	return (
		<section class="s-wrap s-section s-prototype-notes">
			<div class="s-code s-coach-notes">
				<div class="s-coach-note">
					<h2 class="s-contact-title">{page.text("notes.title")}</h2>
					<p class="s-tile-text">{page.text("notes.text")}</p>
				</div>
				<div class="s-coach-note">
					<h3 class="s-label">{page.text("notes.imitated.title")}</h3>
					<ul class="s-chips s-coach-imitated">
						{page.list("notes.imitated.items").map((key) => (
							<li key={key} class="s-chip">
								{page.text(key)}
							</li>
						))}
					</ul>
				</div>
			</div>
		</section>
	);
}

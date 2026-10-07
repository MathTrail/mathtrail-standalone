import { letters } from "../widget/choices";
import { sourceURL } from "./brand";
import { Counted, countSlots, Given, Num } from "./ResearchNumbers";
import type { PageReader } from "./reader";
import type { PaperFile, Research } from "./research";
import { useSiteWords } from "./words";
import { answer, example, fence, fencePosts } from "./worked";

/**
 * paperFor is the file of the paper a page in locale offers: the one in its
 * language, or else the English one, which every language links until there
 * is one in its own; none while the site ships none.
 */
export function paperFor(
	research: Research,
	locale: string,
): PaperFile | undefined {
	const { files } = research.paper;
	return (
		files.find(({ lang }) => lang === locale) ??
		files.find(({ lang }) => lang === "en")
	);
}

/**
 * Hero is the first screen: a badge that says the page is written for
 * technical readers, the heading, what the page shows, the way to the paper
 * and to the code, beside a drawing of the way a task goes from the chat's
 * model through the checks to the child; and under them the counts the page
 * rests on.
 */
export function Hero({
	page,
	research,
	paper,
}: {
	page: PageReader;
	research: Research;
	paper: PaperFile | undefined;
}) {
	const words = useSiteWords();
	const { product } = research;
	return (
		<section class="s-wrap s-research-wrap s-research-top">
			<div class="s-research-hero">
				<div class="s-research-hero-copy">
					<p class="s-badge s-badge-tech">{page.text("hero.badge")}</p>
					<h1>
						<span>{page.text("hero.title")}</span>{" "}
						<span>{page.text("hero.title_end")}</span>
					</h1>
					<p class="s-research-lead">
						{page.text(
							"hero.lead",
							countSlots(words, product.checks, "research.passed"),
						)}
					</p>
					<p class="s-research-actions">
						<PaperButton page={page} paper={paper} at="hero.paper" />
						<a class="s-btn" href={sourceURL}>
							{page.text("hero.code")}
						</a>
					</p>
				</div>
				<Flow page={page} checks={product.checks} />
			</div>
			<ul class="s-research-facts">
				<li>
					<Counted
						value={product.reference_tasks.total}
						noun="research.tasks"
					/>
				</li>
				<li>
					<Counted value={product.checks} noun="research.checks" />
				</li>
				<li>
					<Counted value={product.topics} noun="research.topics" />
				</li>
				<li>
					<Counted value={product.traps} noun="research.traps" />
				</li>
				<li class="s-research-plain">
					{page.text("hero.grades", {
						first: <Num value={product.grades.first} />,
						last: <Num value={product.grades.last} />,
					})}
				</li>
				<li>{page.text("hero.licence")}</li>
			</ul>
		</section>
	);
}

// Flow draws a task's way: the chat's model writes it with its solver, which
// MathTrail runs through every check, sending a task that fails one back,
// and a task that passes them all reaches the child, its answer sealed. The
// worked task is the fence, whose options are those of the solver's two runs
// further down.
function Flow({ page, checks }: { page: PageReader; checks: number }) {
	const at = "hero.flow";
	return (
		<div class="s-research-flow">
			<p class="s-research-flow-title s-research-flow-head s-research-flow-model">
				<span class="s-research-flow-who">
					{page.text("theses.writes.model")}
				</span>
				<span>{page.text("theses.writes.model_does")}</span>
			</p>
			<div class="s-research-flow-body s-research-flow-written">
				<p>
					<span class="s-research-flow-label">{page.text(`${at}.task`)}</span>
					<span>
						{page.text(`${at}.example`, {
							length: <Given value={fence.length} />,
							gap: <Given value={fence.gap} />,
						})}
					</span>
				</p>
				<p>
					<span class="s-research-flow-label">{page.text(`${at}.solver`)}</span>
					<span class="s-research-flow-code">
						{page.text(`${at}.sum`, {
							length: <Given value={fence.length} />,
							gap: <Given value={fence.gap} />,
							one: <Given value={1} />,
							posts: <Given value={fencePosts} />,
							letter: letters[answer],
						})}
					</span>
				</p>
			</div>
			<p class="s-research-flow-step s-research-flow-handed">
				<span class="s-research-flow-arrow" aria-hidden="true" />
				<span>{page.text(`${at}.hands_in`)}</span>
				<span class="s-research-flow-back">{page.text(`${at}.back`)}</span>
			</p>
			<span class="s-research-flow-return" aria-hidden="true" />
			<p class="s-research-flow-title s-research-flow-head s-research-flow-service">
				<span class="s-research-flow-who">
					{page.text("theses.writes.service")}
				</span>
				<span>{page.text("theses.writes.service_does")}</span>
			</p>
			<div class="s-research-flow-body s-research-flow-checked">
				<p class="s-research-flow-count">
					<span>
						<Counted value={checks} noun="research.checks" />
					</span>
					<span>{page.text(`${at}.twice`)}</span>
				</p>
				<ol class="s-research-flow-checks" aria-hidden="true">
					{Array.from({ length: checks }, (_, place) => (
						<li key={place}>✓</li>
					))}
				</ol>
				<p>{page.text(`${at}.asks`)}</p>
			</div>
			<p class="s-research-flow-step s-research-flow-passed">
				<span class="s-research-flow-arrow" aria-hidden="true" />
				<span>{page.text(`${at}.passed`)}</span>
			</p>
			<div class="s-research-flow-child">
				<p class="s-research-flow-title">
					<span class="s-research-flow-who">
						{page.text("theses.writes.child")}
					</span>
					<span>{page.text("theses.writes.child_does")}</span>
				</p>
				<ol class="s-research-flow-options" dir="ltr">
					{example.map((value, place) => (
						<li key={letters[place]}>
							<span>{letters[place]}</span>
							<Given value={value} />
						</li>
					))}
				</ol>
				<p class="s-research-flow-sealed">{page.text(`${at}.sealed`)}</p>
			</div>
		</div>
	);
}

/**
 * PaperButton is the button to the paper's PDF, its pages in its words when
 * they name them, while the site ships one; while it ships none the button's
 * words are left out.
 */
export function PaperButton({
	page,
	paper,
	at,
}: {
	page: PageReader;
	paper: PaperFile | undefined;
	at: string;
}) {
	const words = useSiteWords();
	if (paper === undefined) {
		page.leaveOut(at);
		return null;
	}
	return (
		<a
			class="s-btn s-btn-filled"
			href={paper.path}
			hreflang={paper.lang}
			type="application/pdf"
		>
			{page.text(at, countSlots(words, paper.pages, "research.pages"))}
		</a>
	);
}

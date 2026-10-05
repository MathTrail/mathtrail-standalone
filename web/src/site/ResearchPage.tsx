import { sourceURL } from "./brand";
import type { PageProps } from "./pages";
import { StudentModel } from "./ResearchModel";
import { Counted, countSlots, Num } from "./ResearchNumbers";
import { Theses } from "./ResearchTheses";
import type { PageReader } from "./reader";
import type { PaperFile, Research } from "./research";
import { useSiteWords } from "./words";

/**
 * researchSections are the ids of the parts of the page "Research" that other
 * pages and links lead to: the student model, the live numbers, where the
 * reference tasks came from, and the whole paper.
 */
export const researchSections = {
	model: "student-model",
	live: "live",
	sources: "sources",
	paper: "paper",
} as const;

/**
 * ResearchPage is the page of the numbers behind the product: the paper's
 * title and its PDF; four things the paper shows; the student model against
 * its goals; the live numbers, once there are enough of them; where the
 * reference tasks came from; and the whole paper. Every number it shows is
 * its data's, computed from the commit the site is built from, or marked as
 * one its text chooses.
 */
export function ResearchPage({ page, data }: PageProps) {
	const research = data.research;
	if (research === undefined) {
		throw new Error(
			"the build was given no numbers for the page Research: its table, its counts and its paper",
		);
	}
	const paper = paperFor(research, page.locale);
	return (
		<>
			<Hero page={page} research={research} paper={paper} />
			<Theses page={page} product={research.product} />
			<StudentModel
				id={researchSections.model}
				page={page}
				research={research}
			/>
			<Live page={page} />
			<Sources page={page} research={research} />
			<Paper page={page} paper={paper} />
		</>
	);
}

// paperFor is the file of the paper a page in locale offers: the one in its
// language, or else the English one, which every language links until there
// is one in its own; none while the site ships none.
function paperFor(research: Research, locale: string): PaperFile | undefined {
	const { files } = research.paper;
	return (
		files.find(({ lang }) => lang === locale) ??
		files.find(({ lang }) => lang === "en")
	);
}

// Hero is the first screen: the paper's title, what it shows, the way to the
// paper and to the code, and the counts it rests on.
function Hero({
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
		<section class="s-wrap s-hero s-research-hero">
			<div class="s-hero-copy">
				<h1>
					{page.text("hero.title")}{" "}
					<span class="s-research-subtitle">{page.text("hero.subtitle")}</span>
				</h1>
				<p class="s-lead">
					{page.text(
						"hero.lead",
						countSlots(words, product.checks, "research.passed"),
					)}
				</p>
				<p class="s-choices s-hero-actions">
					<PaperButton page={page} paper={paper} at="hero.paper" />
					<a class="s-btn" href={sourceURL}>
						{page.text("hero.code")}
					</a>
				</p>
				<ul class="s-chips">
					<li class="s-chip">
						<Counted
							value={product.reference_tasks.total}
							noun="research.tasks"
						/>
					</li>
					<li class="s-chip">
						<Counted value={product.checks} noun="research.checks" />
					</li>
					<li class="s-chip">
						<Counted value={product.topics} noun="research.topics" />
					</li>
					<li class="s-chip">
						<Counted value={product.traps} noun="research.traps" />
					</li>
					<li class="s-chip">
						{page.text("hero.grades", {
							first: <Num value={product.grades.first} />,
							last: <Num value={product.grades.last} />,
						})}
					</li>
					<li class="s-chip">{page.text("hero.licence")}</li>
				</ul>
			</div>
		</section>
	);
}

// PaperButton is the button to the paper's PDF, with its pages, when the site
// ships one; the words of the button are left out while it ships none.
function PaperButton({
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

// Live is the block of the live numbers: what it will show, once there are
// enough of them to show without telling a child apart.
function Live({ page }: { page: PageReader }) {
	return (
		<section class="s-wrap s-section" id={researchSections.live}>
			<div class="s-intro">
				<h2>{page.text("live.title")}</h2>
				<p class="s-intro-line">{page.text("live.lead")}</p>
			</div>
			<div class="s-panel s-research-live">
				<p class="s-badge">{page.text("live.badge")}</p>
				<ul>
					{page.list("live.measures").map((key) => (
						<li key={key}>{page.text(key)}</li>
					))}
				</ul>
				<p>{page.text("live.when")}</p>
			</div>
		</section>
	);
}

// Sources is where the reference tasks came from: the rule that takes an idea
// and not a text, the way from a book to a task checked, the authors of the
// books, and the reference tasks by level.
function Sources({ page, research }: { page: PageReader; research: Research }) {
	const words = useSiteWords();
	const total = research.product.reference_tasks.total;
	return (
		<section class="s-wrap s-section" id={researchSections.sources}>
			<div class="s-intro">
				<h2>{page.text("sources.title")}</h2>
				<p class="s-intro-line">{page.text("sources.lead")}</p>
			</div>
			<ol class="s-moves">
				{page.list("sources.flow").map((key) => (
					<li key={key}>{page.text(key)}</li>
				))}
			</ol>
			<ul class="s-research-authors">
				{research.authors.map((author) => (
					<li key={author.id}>
						<p class="s-research-author">
							{page.text(`sources.authors.${author.id}.name`)}{" "}
							<span class="s-research-died">
								{page.text("sources.died", {
									year: <Num value={author.died} form="year" />,
								})}
							</span>
						</p>
						<ul class="s-research-books">
							{author.books.map((book) => (
								<li key={book}>
									<cite>
										{page.text(`sources.authors.${author.id}.books.${book}`)}
									</cite>
								</li>
							))}
						</ul>
					</li>
				))}
			</ul>
			<aside class="s-note">
				<p class="s-note-title">{page.text("sources.public")}</p>
				<p class="s-note-text">{page.text("sources.public_note")}</p>
			</aside>
			<div class="s-research-total">
				<p>
					{page.text(
						"sources.total",
						countSlots(words, total, "research.tasks"),
					)}
				</p>
				<p>{page.text("sources.each")}</p>
				<ol class="s-research-levels">
					{research.levels.map((level) => (
						<li key={level.key} style={{ flexGrow: level.count }}>
							{page.text("sources.level", {
								count: <Num value={level.count} />,
								first: <Num value={level.first} />,
								last: <Num value={level.last} />,
							})}
						</li>
					))}
				</ol>
			</div>
			<p class="s-research-note">{page.text("sources.note")}</p>
		</section>
	);
}

// Paper is the whole paper: what it holds, its PDF when the site ships one, or
// when it comes, and the code.
function Paper({
	page,
	paper,
}: {
	page: PageReader;
	paper: PaperFile | undefined;
}) {
	if (paper === undefined) {
		page.leaveOut("paper.ready");
	} else {
		page.leaveOut("paper.waiting");
	}
	return (
		<section class="s-section s-ask-wrap" id={researchSections.paper}>
			<div class="s-ask">
				<h2 class="s-ask-title">{page.text("paper.title")}</h2>
				<p class="s-ask-lead">
					{page.text(paper === undefined ? "paper.waiting" : "paper.ready")}
				</p>
				<p class="s-choices">
					<PaperButton page={page} paper={paper} at="paper.open" />
					<a class="s-btn" href={sourceURL}>
						{page.text("paper.code")}
					</a>
				</p>
			</div>
		</section>
	);
}

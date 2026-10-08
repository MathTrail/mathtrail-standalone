import { Fragment } from "preact";
import { sourceURL } from "./brand";
import type { PageProps } from "./pages";
import { Hero, PaperButton, paperFor } from "./ResearchHero";
import { LiveNumbers } from "./ResearchLive";
import { StudentModel } from "./ResearchModel";
import { countSlots, Num } from "./ResearchNumbers";
import { Theses } from "./ResearchTheses";
import type { PageReader } from "./reader";
import type { PaperFile, Research } from "./research";
import { GitHubMark } from "./SourceChip";
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
 * ResearchPage is the page of the numbers behind the product: the way a task
 * goes from the chat's model through the checks to the child, and the way to
 * the paper; four things the paper shows; the student model against its
 * goals; the live numbers of the latest month counted whole, once there are
 * enough of them; where the reference tasks came from; and the whole paper.
 * Every number it shows is its data's, computed from the commit the site is
 * built from or taken from the snapshot committed with it, or marked as one
 * its text chooses.
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
			<LiveNumbers id={researchSections.live} page={page} research={research} />
			<Sources page={page} research={research} />
			<Paper page={page} paper={paper} />
		</>
	);
}

// Sources is where the reference tasks came from: the rule that takes an idea
// and not a text, the authors of the books with the year each died, that the
// books belong to everyone, and the reference tasks by level.
function Sources({ page, research }: { page: PageReader; research: Research }) {
	const words = useSiteWords();
	const total = research.product.reference_tasks.total;
	return (
		<section
			class="s-wrap s-research-wrap s-research-section"
			id={researchSections.sources}
		>
			<h2>{page.text("sources.title")}</h2>
			<p class="s-research-intro">{page.text("sources.lead")}</p>
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
						<p class="s-research-books">
							{author.books.map((book, at) => (
								<Fragment key={book}>
									{at === 0 ? null : ", "}
									<cite>
										{page.text(`sources.authors.${author.id}.books.${book}`)}
									</cite>
								</Fragment>
							))}
						</p>
					</li>
				))}
			</ul>
			<p class="s-research-public">
				<span class="s-research-public-name">
					{page.text("sources.public")}
				</span>{" "}
				<span>{page.text("sources.public_note")}</span>
			</p>
			<div class="s-research-total">
				<p class="s-research-total-head">
					<span class="s-research-total-count">
						{page.text(
							"sources.total",
							countSlots(words, total, "research.tasks"),
						)}
					</span>{" "}
					<span>{page.text("sources.each")}</span>
				</p>
				<ol class="s-research-levels">
					{research.levels.map((level) => (
						<li key={level.key} style={{ flexGrow: level.count }}>
							<span>
								{page.text("sources.level", {
									count: <Num value={level.count} />,
									first: <Num value={level.first} />,
									last: <Num value={level.last} />,
								})}
							</span>
						</li>
					))}
				</ol>
				<p class="s-research-note">{page.text("sources.note")}</p>
			</div>
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
		<section
			class="s-wrap s-research-wrap s-research-section s-research-end"
			id={researchSections.paper}
		>
			<div class="s-ask s-ask-row">
				<div class="s-ask-copy">
					<h2 class="s-ask-title">{page.text("paper.title")}</h2>
					<p class="s-ask-lead">
						{page.text(paper === undefined ? "paper.waiting" : "paper.ready")}
					</p>
				</div>
				<p class="s-choices">
					<PaperButton page={page} paper={paper} at="paper.open" />
					<a class="s-btn s-btn-github" href={sourceURL}>
						<GitHubMark size={20} />
						{page.text("paper.code")}
					</a>
				</p>
			</div>
		</section>
	);
}

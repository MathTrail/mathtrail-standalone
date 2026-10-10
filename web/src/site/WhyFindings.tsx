import { useSiteWords } from "./words";
import type { PageReader } from "./reader";
import { doiAddress, type Finding } from "./why";
import { citeText } from "./WhyParts";

/**
 * Research is what studies have found about olympiad maths: a card for each
 * work the site's data names, in its order, each lit in a colour of its own —
 * where and whom the work studied, what it found in a few large words, the
 * finding said in full, a small drawing of it and the work itself.
 */
export function Research({
	page,
	findings,
}: {
	page: PageReader;
	findings: readonly Finding[];
}) {
	return (
		<section class="s-wrap s-section">
			<div class="s-intro">
				<h2>{page.text("research.title")}</h2>
				<p class="s-intro-line">{page.text("research.lead")}</p>
			</div>
			<ul class="s-why-findings">
				{findings.map((finding) => (
					<FindingCard key={finding.id} page={page} finding={finding} />
				))}
			</ul>
		</section>
	);
}

// FindingCard is one finding: a tag of where and whom, a few large words and
// a line under them, the finding's heading and its words, its drawing and the
// work it comes from, which leaves the site.
function FindingCard({ page, finding }: { page: PageReader; finding: Finding }) {
	const words = useSiteWords();
	const at = `research.findings.${finding.id}`;
	return (
		<li class="s-why-finding" data-hue={finding.hue}>
			<span class="s-why-finding-tag">{page.text(`${at}.tag`)}</span>
			<p class="s-why-finding-lead">
				<span class="s-why-finding-big">{page.text(`${at}.big`)}</span>
				<span class="s-why-finding-small">{page.text(`${at}.small`)}</span>
			</p>
			<h3>{page.text(`${at}.title`)}</h3>
			<p class="s-why-finding-text">{page.text(`${at}.text`)}</p>
			<FindingDrawing page={page} at={`${at}.mark`} mark={finding.mark} />
			<a class="s-why-finding-source" href={doiAddress(finding)}>
				{citeText(words, finding)}
				<span aria-hidden="true"> ↗</span>
			</a>
		</li>
	);
}

// FindingDrawing is the small drawing under a finding, by its kind, from the
// words under at: what rose; a row of rising bars, then what they say; a
// chain of steps; or a span of time from one point to another.
function FindingDrawing({
	page,
	at,
	mark,
}: {
	page: PageReader;
	at: string;
	mark: Finding["mark"];
}) {
	switch (mark) {
		case "rise":
			return (
				<p class="s-why-mark">
					{page.list(at).map((key) => (
						<span key={key} class="s-why-mark-chip">
							<b aria-hidden="true">↑</b>
							{page.text(key)}
						</span>
					))}
				</p>
			);
		case "bars":
			return (
				<p class="s-why-mark">
					<span class="s-why-bars" aria-hidden="true">
						<span />
						<span />
						<span />
						<span />
						<span />
					</span>
					<span class="s-why-mark-words">{page.text(at)}</span>
				</p>
			);
		case "chain":
			return (
				<ol class="s-why-mark s-why-chain">
					{page.list(at).map((key) => (
						<li key={key}>
							<span class="s-why-mark-chip">{page.text(key)}</span>
						</li>
					))}
				</ol>
			);
		case "span":
			return (
				<p class="s-why-mark s-why-span">
					<span class="s-why-mark-chip">{page.text(`${at}.from`)}</span>
					<span class="s-why-span-line" aria-hidden="true" />
					<span class="s-why-mark-chip">{page.text(`${at}.to`)}</span>
				</p>
			);
	}
}

import { contactAddress, issuesURL, sourceURL } from "./brand";
import type { PageProps } from "./pages";
import type { PageReader } from "./reader";

/**
 * AboutPage is the page that says who makes MathTrail — one family, each of
 * them under a role, the children by their place in it rather than by name —
 * and where to write about a mistake or an idea: the address the privacy
 * policy names and the code on GitHub, the very ones the footer gives, and
 * the issues the code is discussed in.
 */
export function AboutPage({ page }: PageProps) {
	return (
		<>
			<section class="s-wrap s-hero s-about-hero">
				<div class="s-hero-copy">
					<h1>{page.text("hero.title")}</h1>
					<p class="s-lead">{page.text("hero.lead")}</p>
				</div>
			</section>
			<Team page={page} />
			<Contact page={page} />
		</>
	);
}

// Team is the family, a card for each of them in the order the words give:
// an empty frame where a photograph would stand, for the page shows none of
// the family's faces; the role they play, who they are in the family, and
// what they do. The
// section's heading is for a screen reader, which moves from card to card by
// the names under it.
function Team({ page }: { page: PageReader }) {
	return (
		<section class="s-wrap s-section s-team" aria-labelledby="team">
			<h2 id="team" class="s-hidden">
				{page.text("team.title")}
			</h2>
			<ul class="s-tiles s-tiles-four">
				{page.list("team.members").map((key) => (
					<li key={key} class="s-tile s-member">
						<div class="s-member-frame" />
						<p class="s-chip s-chip-group s-member-role">
							{page.text(`${key}.role`)}
						</p>
						<h3 class="s-member-name">{page.text(`${key}.name`)}</h3>
						<p class="s-tile-text">{page.text(`${key}.text`)}</p>
					</li>
				))}
			</ul>
		</section>
	);
}

// Contact is where to write about a mistake or an idea: the address, written
// out so that it can be copied with no mail program at hand and offered as a
// button, the issues of the code on GitHub, and the code itself.
function Contact({ page }: { page: PageReader }) {
	const mail = `mailto:${contactAddress}`;
	return (
		<section class="s-wrap s-section">
			<div class="s-code s-contact">
				<div class="s-contact-copy">
					<h2 class="s-contact-title">{page.text("contact.title")}</h2>
					<p class="s-tile-text">
						{page.text("contact.lead", {
							address: <a href={mail}>{contactAddress}</a>,
							issue: <a href={issuesURL}>{page.text("contact.issue")}</a>,
						})}
					</p>
				</div>
				<p class="s-choices s-contact-actions">
					<a class="s-btn s-btn-filled" href={mail}>
						{page.text("contact.write")}
					</a>
					<a class="s-btn" href={sourceURL}>
						{page.text("contact.github")}
					</a>
				</p>
			</div>
		</section>
	);
}

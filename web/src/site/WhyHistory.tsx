import { address } from "./addresses";
import { historyPicturePath, historyPictureSize } from "./brand";
import type { PageReader } from "./reader";
import type { HistoryPicture } from "./why";

/**
 * History is how long people have learned to find a way with no example to
 * follow: a saying of Euclid's beside the words, then the pictures of four
 * eras beside the eras themselves. One picture shows at a time, with its place
 * over it and its caption under it, and a chip for each in its era picks it:
 * the chips are radio buttons, and the page's rules show the picture the
 * checked one names, so the pictures can be gone through with no script. The
 * page's script, where it runs, goes on to the next picture every few
 * seconds, a bar under the era filling as it waits; it waits while a pointer
 * rests on the pictures or the eras, and does not move on at all for a reader
 * who asks for less motion.
 */
export function History({
	page,
	eras,
}: {
	page: PageReader;
	eras: readonly (readonly HistoryPicture[])[];
}) {
	const keys = page.list("history.eras");
	if (keys.length !== eras.length) {
		throw new Error(
			`the page Why has the words of ${keys.length} eras of its history and the pictures of ${eras.length}`,
		);
	}
	const pictures = eras.flat();
	return (
		<section id={historySection} class="s-wrap s-section s-history">
			<div class="s-history-head">
				<div class="s-intro">
					<h2>{page.text("history.title")}</h2>
					<p class="s-intro-line">{page.text("history.lead")}</p>
				</div>
				<figure class="s-history-quote">
					<blockquote>{page.text("history.quote")}</blockquote>
					<figcaption>{page.text("history.quoted")}</figcaption>
				</figure>
			</div>
			<div class="s-history-split">
				<figure class="s-history-show">
					<div class="s-history-frame">
						{pictures.map((picture, at) => (
							<img
								key={picture.name}
								class="s-history-picture"
								data-picture={picture.name}
								src={historyPicturePath(picture.name)}
								alt={page.plain(`history.pictures.${picture.name}.alt`)}
								width={historyPictureSize.width}
								height={historyPictureSize.height}
								loading={at === 0 ? "eager" : "lazy"}
								decoding="async"
								style={{ transformOrigin: picture.focus }}
							/>
						))}
						{pictures.map((picture) => (
							<span
								key={picture.name}
								class="s-history-place"
								data-picture={picture.name}
								aria-hidden="true"
							>
								{page.text(`history.pictures.${picture.name}.place`)}
							</span>
						))}
					</div>
					<figcaption class="s-history-captions">
						{pictures.map((picture) => (
							<span
								key={picture.name}
								class="s-history-caption"
								data-picture={picture.name}
							>
								{page.text(`history.pictures.${picture.name}.caption`)}
							</span>
						))}
					</figcaption>
				</figure>
				<fieldset class="s-history-eras">
					<legend class="s-hidden">{page.text("history.choice")}</legend>
					<ol>
						{eras.map((era, at) => (
							<Era
								key={keys[at]}
								page={page}
								at={keys[at] ?? ""}
								era={era}
								first={at === 0}
							/>
						))}
					</ol>
				</fieldset>
			</div>
		</section>
	);
}

/**
 * historySection is the id of the page's history, which its rules and its
 * script find it by.
 */
export const historySection = "history";

/**
 * historyChoice is the name of the radio buttons that pick a picture of the
 * history.
 */
export const historyChoice = "history-picture";

/** historyPickId is the id of the radio button that picks the picture name. */
export function historyPickId(name: string): string {
	return `history-${name}`;
}

/**
 * historyRules are the rules of style that mark as shown the picture of the
 * history the checked chip names, its place and its caption, which the
 * stylesheet then shows, the picture slowly growing and the caption rising
 * into place; and that fill the bar of that picture in its era as the page's
 * script waits on it, the bars of the pictures before it in its era standing
 * full.
 */
export function historyRules(
	eras: readonly (readonly HistoryPicture[])[],
): string {
	const chosen = (name: string) =>
		`#${historySection}:has(#${historyPickId(name)}:checked)`;
	const shown = eras
		.flat()
		.map(({ name }) => `${chosen(name)} [data-picture="${name}"]`);
	const filling = eras
		.flat()
		.map(({ name }) => `${chosen(name)} [data-bar="${name}"]`);
	const filled = eras.flatMap((era) =>
		era.flatMap((picture, at) =>
			era
				.slice(0, at)
				.map(({ name }) => `${chosen(picture.name)} [data-bar="${name}"]`),
		),
	);
	return [
		`${shown.join(",")}{--s-shown:1;--s-seen:visible;--s-rise:s-history-rise}`,
		`${filling.join(",")}{animation-name:s-history-fill}`,
		filled.length > 0 ? `${filled.join(",")}{transform:none}` : "",
	]
		.filter((written) => written !== "")
		.join("\n");
}

// Era is one era of the history: its time, its heading and what happened in
// it, then a chip for each of its pictures and, under them, a bar for each
// that the page's script fills as it waits. The era of today leads to the
// coach's page.
function Era({
	page,
	at,
	era,
	first,
}: {
	page: PageReader;
	at: string;
	era: readonly HistoryPicture[];
	first: boolean;
}) {
	return (
		<li class="s-history-era">
			<span class="s-history-line" aria-hidden="true" />
			<span class="s-history-dot" aria-hidden="true" />
			<div class="s-history-era-text" data-first-picture={era[0]?.name}>
				<p class="s-history-when">{page.text(`${at}.label`)}</p>
				<h3>{page.text(`${at}.title`)}</h3>
				<p class="s-history-what">{page.text(`${at}.text`)}</p>
			</div>
			<p class="s-history-chips">
				{era.map((picture, index) => (
					<label key={picture.name} class="s-history-chip">
						<input
							type="radio"
							class="s-history-pick s-hidden"
							name={historyChoice}
							id={historyPickId(picture.name)}
							value={picture.name}
							checked={first && index === 0}
						/>
						<span>
							{page.has(`history.pictures.${picture.name}.chip`)
								? page.text(`history.pictures.${picture.name}.chip`)
								: page.text(`history.pictures.${picture.name}.place`)}
						</span>
					</label>
				))}
			</p>
			{page.has(`${at}.link`) && (
				<p class="s-history-link">
					<a href={address(page.locale, "coach")}>{page.text(`${at}.link`)}</a>
				</p>
			)}
			<span class="s-history-bars" aria-hidden="true">
				{era.map((picture) => (
					<span key={picture.name}>
						<span data-bar={picture.name} />
					</span>
				))}
			</span>
		</li>
	);
}

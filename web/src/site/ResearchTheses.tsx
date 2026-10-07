import type { ComponentChildren } from "preact";
import { letters } from "../widget/choices";
import { Counted, countSlots, Given, Num } from "./ResearchNumbers";
import type { PageReader } from "./reader";
import type { ResearchFile } from "./research";
import { useSiteWords } from "./words";

/** Product is what the page says of the product: its counts and constants. */
type Product = ResearchFile["product"];

// example are the values of the options of the worked example of the solver's
// two runs, which the text chooses, and answer the place of the right one.
const example = [3, 4, 5, 6, 12] as const;
const answer = 2;

/**
 * secondRun is the options of a task as the solver's second run sees them:
 * the value of each option moved to the letter the relabelling gives its
 * letter, so the value under a letter is the one that letter did not have.
 */
export function secondRun(
	values: readonly number[],
	relabelled: readonly string[],
): number[] {
	const moved = [...values];
	relabelled.forEach((letter, from) => {
		const to = letters.indexOf(letter as (typeof letters)[number]);
		moved[to] = values[from] ?? 0;
	});
	return moved;
}

/**
 * Theses are four things the paper shows: the chat's model writes every task,
 * the solver it hands in runs twice, the answer is sealed, and the next task
 * is a little harder than the child, but within reach. Each says its numbers
 * through the data, beside a drawing of what it says.
 */
export function Theses({
	page,
	product,
}: {
	page: PageReader;
	product: Product;
}) {
	const words = useSiteWords();
	return (
		<section class="s-wrap s-section" aria-labelledby="theses">
			<h2 id="theses" class="s-hidden">
				{page.text("theses.title")}
			</h2>
			<ol class="s-research-theses">
				<Thesis place={1} page={page} at="theses.writes">
					<Writes page={page} />
				</Thesis>
				<Thesis
					place={2}
					page={page}
					at="theses.twice"
					slots={{
						...countSlots(words, product.relabel, "research.places"),
						from: letters[0],
						to: product.relabelled[0] ?? "",
					}}
				>
					<Twice page={page} product={product} />
				</Thesis>
				<Thesis place={3} page={page} at="theses.sealed">
					<Sealed page={page} />
				</Thesis>
				<Thesis
					place={4}
					page={page}
					at="theses.reach"
					slots={{
						low: <Num value={product.corridor.low} form="share" />,
						high: <Num value={product.corridor.high} form="share" />,
						middle: <Num value={product.corridor.middle} form="share" />,
						trial: (
							<Counted value={product.trial_answers} noun="research.trial" />
						),
					}}
				>
					<Reach page={page} product={product} />
				</Thesis>
			</ol>
		</section>
	);
}

// Thesis is one of the four: its place, its title, what it says, and its
// drawing.
function Thesis({
	place,
	page,
	at,
	slots = {},
	children,
}: {
	place: number;
	page: PageReader;
	at: string;
	slots?: NonNullable<Parameters<PageReader["text"]>[1]>;
	children: ComponentChildren;
}) {
	return (
		<li class="s-research-thesis">
			<p class="s-research-place" aria-hidden="true">
				<Given value={place} digits={2} />
			</p>
			<h3>{page.text(`${at}.title`)}</h3>
			<p>{page.text(`${at}.text`, slots)}</p>
			{children}
		</li>
	);
}

// Writes draws who does what: the chat's model writes, MathTrail checks, and
// the child solves.
function Writes({ page }: { page: PageReader }) {
	return (
		<ol class="s-moves">
			{(["model", "service", "child"] as const).map((who) => (
				<li key={who}>
					<strong>{page.text(`theses.writes.${who}`)}</strong>{" "}
					{page.text(`theses.writes.${who}_does`)}
				</li>
			))}
		</ol>
	);
}

// Twice draws the solver's two runs of a worked example: the options under
// their letters in the first run and in the second, the letter each run's
// program picks, and what a program that wrote the letter by hand points at.
function Twice({ page, product }: { page: PageReader; product: Product }) {
	const second = secondRun(example, product.relabelled);
	const right = letters[answer];
	const moved = product.relabelled[answer] ?? "";
	return (
		<>
			<table class="s-research-runs" dir="ltr">
				<thead>
					<tr>
						<td />
						{letters.map((letter) => (
							<th key={letter} scope="col">
								{letter}
							</th>
						))}
						<th scope="col">{page.text("theses.twice.picks")}</th>
					</tr>
				</thead>
				<tbody>
					<tr>
						<th scope="row">{page.text("theses.twice.first")}</th>
						{example.map((value, at) => (
							<td
								key={letters[at]}
								class={at === answer ? "s-research-right" : ""}
							>
								<Given value={value} />
							</td>
						))}
						<td>{right}</td>
					</tr>
					<tr>
						<th scope="row">{page.text("theses.twice.second")}</th>
						{second.map((value, at) => (
							<td
								key={letters[at]}
								class={letters[at] === moved ? "s-research-right" : ""}
							>
								<Given value={value} />
							</td>
						))}
						<td>{moved}</td>
					</tr>
				</tbody>
			</table>
			<p class="s-research-caption">
				{page.text("theses.twice.caption", {
					letter: right,
					value: <Given value={second[answer] ?? 0} />,
				})}
			</p>
		</>
	);
}

// Sealed draws what the child sees beside what stays sealed until the answer,
// and says what the seal does not reach.
function Sealed({ page }: { page: PageReader }) {
	return (
		<>
			<div class="s-research-sealed">
				<div>
					<p class="s-research-label">{page.text("theses.sealed.sees")}</p>
					<p>{page.text("theses.sealed.seen")}</p>
				</div>
				<div class="s-research-locked">
					<p class="s-research-label">{page.text("theses.sealed.hidden")}</p>
					<p>{page.text("theses.sealed.kept")}</p>
				</div>
			</div>
			<p class="s-research-caption">{page.text("theses.sealed.cipher")}</p>
			<p class="s-research-caption">{page.text("theses.sealed.outside")}</p>
		</>
	);
}

// Reach draws the chance of a right answer from nothing to certain, with the
// guess a child has among the options and the corridor the next task is
// chosen in, under the formula of the chance.
function Reach({ page, product }: { page: PageReader; product: Product }) {
	const { guess, corridor, options } = product;
	const at = (chance: number) => `${chance * 100}%`;
	return (
		<>
			<p class="s-research-formula" dir="ltr">
				P = c + (<Given value={1} /> − c) · σ(θ + δ<sub>t</sub> − β)
			</p>
			<div class="s-research-axis" dir="ltr" aria-hidden="true">
				<span
					class="s-research-band"
					style={{
						insetInlineStart: at(corridor.low),
						inlineSize: at(corridor.high - corridor.low),
					}}
				/>
				<span class="s-research-tick" style={{ insetInlineStart: at(0) }}>
					<Given value={0} />
				</span>
				<span
					class="s-research-tick s-research-guess"
					style={{ insetInlineStart: at(guess) }}
				>
					<Num value={guess} form="chance" />
				</span>
				<span
					class="s-research-tick"
					style={{ insetInlineStart: at(corridor.low) }}
				>
					<Num value={corridor.low} form="chance" />
				</span>
				<span
					class="s-research-tick"
					style={{ insetInlineStart: at(corridor.high) }}
				>
					<Num value={corridor.high} form="chance" />
				</span>
				<span class="s-research-tick" style={{ insetInlineStart: at(1) }}>
					<Given value={1} />
				</span>
			</div>
			<p class="s-research-caption">
				{page.text("theses.reach.guess", {
					guess: <Num value={guess} form="chance" />,
					options: <Num value={options} />,
				})}{" "}
				{page.text("theses.reach.corridor")}
			</p>
		</>
	);
}

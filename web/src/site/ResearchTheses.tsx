import type { ComponentChildren } from "preact";
import { letters } from "../widget/choices";
import { Counted, countSlots, Given, Num } from "./ResearchNumbers";
import type { PageReader } from "./reader";
import type { ResearchFile } from "./research";
import { useSiteWords } from "./words";
import { answer, example } from "./worked";

/** Product is what the page says of the product: its counts and constants. */
type Product = ResearchFile["product"];

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
 * is a little harder than the child, but within reach. Each is a row: its
 * place, what it says with its numbers through the data, and a drawing of it.
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
		<section class="s-wrap s-research-wrap" aria-labelledby="theses">
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

// Thesis is one of the four: its place, its title and what it says, and its
// drawing beside them, or under them on a narrow screen.
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
			<div class="s-research-thesis-body">
				<div class="s-research-thesis-words">
					<h3>{page.text(`${at}.title`)}</h3>
					<p>{page.text(`${at}.text`, slots)}</p>
				</div>
				<div class="s-research-thesis-drawing">{children}</div>
			</div>
		</li>
	);
}

// Writes draws who does what: the chat's model writes, MathTrail checks, and
// the child solves.
function Writes({ page }: { page: PageReader }) {
	return (
		<ol class="s-research-writes">
			{(["model", "service", "child"] as const).map((who) => (
				<li key={who} class={who === "service" ? "s-research-us" : undefined}>
					<span class="s-research-box">
						<span class="s-research-who">
							{page.text(`theses.writes.${who}`)}
						</span>
						<span class="s-research-does">
							{page.text(`theses.writes.${who}_does`)}
						</span>
					</span>
				</li>
			))}
		</ol>
	);
}

// Twice draws the solver's two runs of the worked task: in each, the options
// under their letters, the right one filled, and the letter the run's program
// picks; under them, what a program that wrote the letter by hand points at.
function Twice({ page, product }: { page: PageReader; product: Product }) {
	const second = secondRun(example, product.relabelled);
	const right = letters[answer];
	const moved = product.relabelled[answer] ?? "";
	return (
		<>
			<div class="s-research-runs" dir="ltr">
				<Run
					label={page.text("theses.twice.first")}
					values={example}
					picked={right}
					picks={page.text("theses.twice.picks", { letter: right })}
				/>
				<Run
					label={page.text("theses.twice.second")}
					values={second}
					picked={moved}
					picks={page.text("theses.twice.picks", { letter: moved })}
				/>
			</div>
			<p class="s-research-caption">
				{page.text("theses.twice.caption", {
					letter: right,
					value: <Given value={second[answer] ?? 0} />,
				})}
			</p>
		</>
	);
}

// Run is one run of the solver: its name, the options under their letters,
// the one picked filled, and the letter picked.
function Run({
	label,
	values,
	picked,
	picks,
}: {
	label: ComponentChildren;
	values: readonly number[];
	picked: string;
	picks: ComponentChildren;
}) {
	return (
		<div class="s-research-run">
			<p class="s-research-run-name">{label}</p>
			<ol class="s-research-cells">
				{values.map((value, at) => (
					<li
						key={letters[at]}
						class={letters[at] === picked ? "s-research-right" : undefined}
					>
						<span class="s-research-letter">{letters[at]}</span>
						<span class="s-research-cell">
							<Given value={value} />
						</span>
					</li>
				))}
			</ol>
			<p class="s-research-picks">{picks}</p>
		</div>
	);
}

// Sealed draws what the child sees beside what stays sealed until the answer,
// names the cipher, and says what the seal does not reach.
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
			<p class="s-research-cipher">{page.text("theses.sealed.cipher")}</p>
			<p class="s-research-caption">{page.text("theses.sealed.outside")}</p>
		</>
	);
}

// Reach draws the chance of a right answer from nothing to certain under the
// formula of the chance: the guess a child has among the options hatched, the
// corridor the next task is chosen in filled, each named under the scale.
function Reach({ page, product }: { page: PageReader; product: Product }) {
	const { guess, corridor, options } = product;
	const at = (chance: number) => `${chance * 100}%`;
	return (
		<>
			<p class="s-research-formula" dir="ltr">
				P = c + (<Given value={1} /> − c) · σ(θ + δ<sub>t</sub> − β)
			</p>
			<div class="s-research-scale" dir="ltr" aria-hidden="true">
				<span class="s-research-scale-track">
					<span
						class="s-research-scale-guess"
						style={{ inlineSize: at(guess) }}
					/>
					<span
						class="s-research-scale-corridor"
						style={{
							insetInlineStart: at(corridor.low),
							inlineSize: at(corridor.high - corridor.low),
						}}
					/>
				</span>
				<span class="s-research-tick s-research-tick-first">
					<Given value={0} />
				</span>
				<span class="s-research-tick" style={{ insetInlineStart: at(guess) }}>
					<Num value={guess} form="chance" />
				</span>
				<span
					class="s-research-tick s-research-tick-ours"
					style={{ insetInlineStart: at(corridor.low) }}
				>
					<Num value={corridor.low} form="hundredths" />
				</span>
				<span
					class="s-research-tick s-research-tick-ours"
					style={{ insetInlineStart: at(corridor.high) }}
				>
					<Num value={corridor.high} form="hundredths" />
				</span>
				<span class="s-research-tick s-research-tick-last">
					<Given value={1} />
				</span>
			</div>
			<ul class="s-research-swatches">
				<li>
					<span class="s-research-swatch s-research-swatch-guess" />
					<span>
						{page.text("theses.reach.guess", {
							guess: <Num value={guess} form="chance" />,
							options: <Num value={options} />,
						})}
					</span>
				</li>
				<li>
					<span class="s-research-swatch s-research-swatch-corridor" />
					<span>{page.text("theses.reach.corridor")}</span>
				</li>
			</ul>
		</>
	);
}

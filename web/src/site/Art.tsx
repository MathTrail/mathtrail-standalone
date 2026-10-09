import { type ComponentChildren, Fragment } from "preact";
import type { PageReader } from "./reader";

/**
 * DrawingProps are what a drawing of the site's own is drawn from: the page,
 * the key of the drawing's words in the page's words, and the way the page's
 * language writes numbers. A drawing's numbers and its shape are its
 * problem's, the same in every language; only its words are the page's.
 */
export type DrawingProps = {
	readonly page: PageReader;
	readonly at: string;
	readonly numbers: Intl.NumberFormat;
};

/**
 * DigitTree draws each digit as a branch, with the two-digit numbers it
 * begins: the digit followed by every other digit, in the order the digits
 * are given, each written the way the page's language writes numbers.
 */
export function DigitTree({
	digits,
	numbers,
}: {
	digits: readonly number[];
	numbers: Intl.NumberFormat;
}) {
	return (
		<div class="s-tree">
			{digits.map((first) => (
				<div key={first} class="s-tree-branch">
					<span class="s-art-dot">{numbers.format(first)}</span>
					<span class="s-tree-leaves">
						{digits
							.filter((second) => second !== first)
							.map((second) => (
								<span key={second} class="s-art-chip">
									{numbers.format(first * 10 + second)}
								</span>
							))}
					</span>
				</div>
			))}
		</div>
	);
}

/**
 * Step is a step of a chain: the words on its arrow, if it has any, and the
 * value it ends at.
 */
export type Step = { readonly by?: ComponentChildren; readonly to: string };

/**
 * Chain draws a value and the values the steps from it reach, one after
 * another, each step's words beside its arrow. A chain told forward points on
 * from each value to the next; one told back points from each value to the
 * one before it, a step undone. A step with no words is its arrow alone.
 */
export function Chain({
	start,
	steps,
	back = false,
	accent = false,
}: {
	start: string;
	steps: readonly Step[];
	back?: boolean;
	accent?: boolean;
}) {
	const dot = accent ? "s-art-dot s-art-dot-accent" : "s-art-dot";
	return (
		<div class="s-chain">
			<span class={dot}>{start}</span>
			{steps.map(({ by, to }, at) => (
				<Fragment key={at}>
					<span class="s-chain-step">
						<Arrow by={by} back={back} />
					</span>
					<span class={dot}>{to}</span>
				</Fragment>
			))}
		</div>
	);
}

// Arrow is a step's arrow with its words: after them where the chain is told
// forward, before them where it is told back.
function Arrow({ by, back }: { by?: ComponentChildren; back: boolean }) {
	if (by === undefined) {
		return <>{back ? "←" : "→"}</>;
	}
	return back ? <>← {by}</> : <>{by} →</>;
}

import type { Words } from "../i18n/words";
import { type SiteKey, useSiteWords } from "./words";

// The page "Research" types no number of its own. A number it reports comes
// from its data and is written as data, its value as the file holds it, so
// that a check of the page can tell it from a number typed into the words; a
// number the text chooses rather than reports, such as the options of a worked
// example, is marked as given, and the same check lets it through by its mark.

/**
 * NumberForm is how a number of the page is written: a count; a share, as a
 * percentage; a measure in logits, in answers, in points or in changes a
 * hundred answers; a chance; and a year or a seed, which are names more than
 * amounts and keep their digits together.
 */
export type NumberForm =
	| "count"
	| "share"
	| "logit"
	| "answers"
	| "points"
	| "per_100_answers"
	| "chance"
	| "year"
	| "seed";

// formats are the options each form is written with.
const formats: Readonly<Record<NumberForm, Intl.NumberFormatOptions>> = {
	count: {},
	share: { style: "percent", maximumFractionDigits: 1 },
	logit: { minimumFractionDigits: 2, maximumFractionDigits: 2 },
	answers: { maximumFractionDigits: 1 },
	points: { maximumFractionDigits: 0 },
	per_100_answers: { maximumFractionDigits: 1 },
	chance: { minimumFractionDigits: 1, maximumFractionDigits: 3 },
	year: { useGrouping: false },
	seed: { useGrouping: false },
};

/** written is a number as the language of locale writes it in form. */
export function written(
	locale: string,
	value: number,
	form: NumberForm = "count",
	digits?: number,
): string {
	return new Intl.NumberFormat(locale, {
		...formats[form],
		minimumIntegerDigits: digits,
	}).format(value);
}

/**
 * Num writes a number of the page's data: its value as the data holds it, in
 * the element's value, and its text as the page's language writes it.
 */
export function Num({
	value,
	form = "count",
}: {
	readonly value: number;
	readonly form?: NumberForm;
}) {
	const { locale } = useSiteWords();
	return <data value={String(value)}>{written(locale, value, form)}</data>;
}

/**
 * Given writes a number the page's text chooses rather than reports: a value
 * of a worked example, an end of an axis, the place of a thesis, written with
 * at least digits digits. It is never data.
 */
export function Given({
	value,
	form = "count",
	digits,
}: {
	readonly value: number;
	readonly form?: NumberForm;
	readonly digits?: number;
}) {
	const { locale } = useSiteWords();
	return <span data-given="">{written(locale, value, form, digits)}</span>;
}

/** Commit writes a commit of the page's data, cut to the length people read it at. */
export function Commit({ hash }: { readonly hash: string }) {
	return <code>{hash.slice(0, 12)}</code>;
}

/** CommitDate writes the date of a commit, as the page's language writes dates. */
export function CommitDate({ date }: { readonly date: string }) {
	const { locale } = useSiteWords();
	return (
		<time dateTime={date}>
			{new Intl.DateTimeFormat(locale, {
				dateStyle: "long",
				timeZone: "UTC",
			}).format(new Date(date))}
		</time>
	);
}

/**
 * countSlots are the slots count and noun that a sentence of the page's words
 * says a count of its data in: the count written as data, and its noun in the
 * form the count asks for, which the words place as their language orders
 * them.
 */
export function countSlots(
	words: Words<SiteKey>,
	value: number,
	noun: SiteKey,
) {
	return {
		count: <Num value={value} />,
		noun: words.text(noun, { count: value }),
	};
}

/**
 * Counted writes a count of the page's data with its noun, which the site's
 * words give in the form the count asks for: the count picks the form, and is
 * written apart, as data.
 */
export function Counted({
	value,
	noun,
}: {
	readonly value: number;
	readonly noun: SiteKey;
}) {
	const words = useSiteWords();
	return (
		<>
			<Num value={value} /> {words.text(noun, { count: value })}
		</>
	);
}

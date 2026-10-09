import type { Calendar } from "./model";
import {
	type Drawn,
	Label,
	lineHeight,
	r1,
	Stack,
	sizes,
	stackOf,
} from "./shapes";
import { room, smallest, widthOf, written } from "./text";

// A month's page, in the card's pixels: how wide a day's column is at most,
// how tall the line of a week's days is, the room around a weekday's name in
// its column, and how large the square under a marked day is.
const widestColumn = 40;
const weekTall = 24;
const nameAside = 1;
const markedWide = 26;
const markedTall = 20;

// monday is a Monday, as the reference the names of the weekdays are read
// from: the names do not depend on the year.
const monday = Date.UTC(2024, 0, 1);

// weekdays is how the names of the weekdays are written in a language: in
// the language the tag names where the browser knows it, and as the browser
// writes them where the tag is none it can read.
function weekdays(
	locale: string,
	form: "short" | "narrow",
): Intl.DateTimeFormat {
	const options = { weekday: form, timeZone: "UTC" } as const;
	try {
		return new Intl.DateTimeFormat(locale, options);
	} catch {
		return new Intl.DateTimeFormat(undefined, options);
	}
}

// latin says whether a text is written in Latin letters alone.
function latin(text: string): boolean {
	return /^[A-Za-z .]+$/.test(text);
}

/**
 * weekdayNames are the names of the days of a week in a language, from the
 * day the week starts on, and the size they are written at: the short names
 * where every one fits its column, at the larger size where it can, else the
 * narrow ones. Where a language's narrow names are Latin letters and its short
 * ones are not — the data of some languages has them so — the narrow ones are
 * no names in that language, and the short ones stand at the smaller size.
 */
export function weekdayNames(
	locale: string,
	startsOnSunday: boolean,
	column: number,
): { names: string[]; size: number } {
	const named = (form: "short" | "narrow") => {
		const format = weekdays(locale, form);
		return Array.from({ length: 7 }, (_, after) =>
			format.format(
				monday + ((after + (startsOnSunday ? 6 : 0)) % 7) * 86_400_000,
			),
		);
	};
	const fits = (names: readonly string[], size: number) =>
		names.every((name) => widthOf(name, size) + 2 * nameAside <= column);
	const short = named("short");
	for (const size of [sizes.number, smallest]) {
		if (fits(short, size)) {
			return { names: short, size };
		}
	}
	const narrow = named("narrow");
	if (narrow.every(latin) && !short.every(latin)) {
		return { names: short, size: smallest };
	}
	return { names: narrow, size: sizes.number };
}

/**
 * drawCalendar draws a month's page: the names of the weekdays in the card's
 * language over seven columns, from the day the week starts on; the days of
 * the month in Latin digits, the 1st under its weekday; each marked day on a
 * square in the shading tone, and its label under its week's days, on as many
 * lines as keep the labels of a week apart.
 */
export function drawCalendar(calendar: Calendar, locale: string): Drawn {
	const column = Math.min(widestColumn, room / 7);
	const width = 7 * column;
	const sunday = calendar.week_starts === "sunday";
	const { names, size } = weekdayNames(locale, sunday, column);
	const header = names.map((name, place) => ({ name, column: place }));
	const offset = sunday ? calendar.first % 7 : calendar.first - 1;
	const weeks = Math.ceil((offset + calendar.days) / 7);
	const marks = new Map(
		Object.entries(calendar.marks ?? {}).map(([day, label]) => [
			Number(day),
			written(label),
		]),
	);
	const days = Array.from({ length: calendar.days }, (_, before) => {
		const day = before + 1;
		const cell = offset + before;
		return { day, week: Math.floor(cell / 7), x: ((cell % 7) + 0.5) * column };
	});
	const headerTall = lineHeight(size) + 4;
	let top = headerTall + 4;
	const laid = Array.from({ length: weeks }, (_, week) => {
		const labels = stackOf(
			days
				.filter((one) => one.week === week && marks.has(one.day))
				.map((one) => ({
					id: `mark-${one.day}`,
					x: one.x,
					text: marks.get(one.day) ?? "",
				})),
			0,
			width,
		);
		const at = { week, top, labels };
		top += weekTall + labels.lines * lineHeight();
		return at;
	});
	const tops = new Map(laid.map((one) => [one.week, one.top]));
	return {
		width,
		height: top + 2,
		body: (
			<>
				{header.map((one) => (
					<Label
						key={one.column}
						x={(one.column + 0.5) * column}
						y={headerTall / 2}
						text={one.name}
						size={size}
						strong={false}
					/>
				))}
				<path
					d={`M0 ${r1(headerTall + 1)}H${r1(width)}`}
					class="mt-pic-line"
					stroke-width={1}
				/>
				{days.map((one) => {
					const middle = (tops.get(one.week) ?? 0) + weekTall / 2;
					return (
						<g key={one.day}>
							{marks.has(one.day) && (
								<rect
									x={r1(one.x - markedWide / 2)}
									y={r1(middle - markedTall / 2)}
									width={markedWide}
									height={markedTall}
									rx={3}
									class="mt-pic-fill"
								/>
							)}
							<Label
								x={one.x}
								y={middle}
								text={String(one.day)}
								size={sizes.day}
								strong={false}
							/>
						</g>
					);
				})}
				{laid.map((one) => (
					<Stack
						key={one.week}
						placed={one.labels.placed}
						first={one.top + weekTall + lineHeight() / 2 - 2}
					/>
				))}
			</>
		),
	};
}

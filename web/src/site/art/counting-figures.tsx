import type { Lattice, LatticeLine } from "../../design/picture/model";
import { Say, Stack } from "../Sketch";
import type { ArtProps, TopicArt } from "./art";
import { Kind } from "./pictures";

// grid is every line of a rectangle of squared paper, width cells across and
// height down, drawn as long lines from side to side.
function grid(width: number, height: number): LatticeLine[] {
	return [
		...Array.from(
			{ length: height + 1 },
			(_, y): LatticeLine => [0, y, width, y],
		),
		...Array.from(
			{ length: width + 1 },
			(_, x): LatticeLine => [x, 0, x, height],
		),
	];
}

// matches are the sides of the cells of a rectangle, each a line of its own:
// one match a side, the sides two cells share drawn once.
function matches(width: number, height: number): LatticeLine[] {
	const along = Array.from({ length: height + 1 }, (_, y) =>
		Array.from({ length: width }, (_, x): LatticeLine => [x, y, x + 1, y]),
	).flat();
	const down = Array.from({ length: width + 1 }, (_, x) =>
		Array.from({ length: height }, (_, y): LatticeLine => [x, y, x, y + 1]),
	).flat();
	return [...along, ...down];
}

// Drawn is a figure on squared paper with the words under it.
function Drawn({
	page,
	at,
	picture,
	words,
}: ArtProps & { picture: Lattice; words: string }) {
	return (
		<Stack gap={8}>
			<Kind locale={page.locale} picture={picture} />
			<Say>{page.text(`${at}.${words}`)}</Say>
		</Stack>
	);
}

// The matches the third example takes away: the two inside matches that stand
// up and down, and one of the top.
const taken: LatticeLine[] = [
	[1, 0, 1, 1],
	[1, 1, 1, 2],
	[0, 0, 1, 0],
];
const left = matches(2, 2).filter(
	(line) =>
		!taken.some((gone) => gone.every((value, at) => value === line[at])),
);

const diagonals: LatticeLine[] = [
	[0, 0, 2, 0],
	[2, 0, 2, 2],
	[2, 2, 0, 2],
	[0, 2, 0, 0],
	[0, 0, 2, 2],
	[2, 0, 0, 2],
];
const rays: LatticeLine[] = [
	[0, 0, 0, 4],
	[0, 4, 4, 4],
	[0, 0, 4, 4],
	[0, 0, 1, 4],
	[0, 0, 2, 4],
];

/**
 * art are the drawings of the page of counting figures, each a figure on
 * squared paper drawn as the card draws one: a square of four cells, a square
 * cut by its diagonals, a row of squares of matches, and the figures of the
 * worked examples.
 */
export const art: TopicArt = {
	pictures: true,
	hero: (props) => (
		<Drawn
			{...props}
			picture={{ kind: "lattice", width: 2, height: 2, lines: grid(2, 2) }}
			words="count"
		/>
	),
	basis: [
		(props) => (
			<Drawn
				{...props}
				picture={{ kind: "lattice", width: 2, height: 2, lines: grid(2, 2) }}
				words="count"
			/>
		),
		(props) => (
			<Drawn
				{...props}
				picture={{ kind: "lattice", width: 2, height: 2, lines: diagonals }}
				words="count"
			/>
		),
		(props) => (
			<Drawn
				{...props}
				picture={{ kind: "lattice", width: 3, height: 1, lines: matches(3, 1) }}
				words="count"
			/>
		),
	],
	examples: [
		{
			task: (props) => (
				<Drawn
					{...props}
					picture={{ kind: "lattice", width: 4, height: 4, lines: rays }}
					words="count"
				/>
			),
		},
		{
			task: (props) => (
				<Drawn
					{...props}
					picture={{ kind: "lattice", width: 3, height: 3, lines: grid(3, 3) }}
					words="sizes"
				/>
			),
		},
		{
			task: (props) => (
				<Drawn
					{...props}
					picture={{
						kind: "lattice",
						width: 2,
						height: 2,
						lines: left,
						dashed: taken,
					}}
					words="taken"
				/>
			),
		},
	],
};

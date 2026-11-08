import type { Corner, Lattice, LatticeLabel } from "../../design/picture/model";
import { Say, Stack } from "../Sketch";
import type { ArtProps, TopicArt } from "./art";
import { Kind } from "./pictures";

// net is the cells of a net and the label of each, the cells and the marks of
// a figure on squared paper.
function net(
	width: number,
	height: number,
	labelled: readonly [Corner, string][],
): Lattice {
	return {
		kind: "lattice",
		width,
		height,
		cells: labelled.map(([cell]) => cell),
		marks: labelled.map(([at, label]): LatticeLabel => ({ at, label })),
	};
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

// The cross the page folds: a column of four with one cell on each side of
// the second, numbered from the top of the column and then left to right.
const cross = net(3, 4, [
	[[1, 0], "1"],
	[[0, 1], "2"],
	[[1, 1], "3"],
	[[2, 1], "4"],
	[[1, 2], "5"],
	[[1, 3], "6"],
]);

// The cross of the first example: the column numbered 1 to 4 from the top,
// 5 and 6 beside its second square.
const column = net(3, 4, [
	[[1, 0], "1"],
	[[1, 1], "2"],
	[[1, 2], "3"],
	[[1, 3], "4"],
	[[0, 1], "5"],
	[[2, 1], "6"],
]);

// A sheet of 4 by 4 cells opened after two folds, its fold lines dashed and
// one hole in each quarter.
const opened: Lattice = {
	kind: "lattice",
	width: 4,
	height: 4,
	cells: Array.from(
		{ length: 16 },
		(_, at): Corner => [at % 4, Math.floor(at / 4)],
	),
	dashed: [
		[2, 0, 2, 4],
		[0, 2, 4, 2],
	],
	marks: [
		{ at: [0, 0], label: "O" },
		{ at: [3, 0], label: "O" },
		{ at: [0, 3], label: "O" },
		{ at: [3, 3], label: "O" },
	],
};

/**
 * art are the drawings of the page of cubes and paper, each a figure on
 * squared paper drawn as the card draws one: a net of a cube, a sheet opened
 * after its folds, and the cells a die stands on.
 */
export const art: TopicArt = {
	pictures: true,
	hero: (props) => <Drawn {...props} picture={cross} words="opposite" />,
	basis: [
		(props) => <Drawn {...props} picture={cross} words="opposite" />,
		(props) => <Drawn {...props} picture={opened} words="holes" />,
		(props) => (
			<Drawn {...props} picture={net(1, 1, [[[0, 0], "2"]])} words="bottom" />
		),
	],
	examples: [
		{ task: (props) => <Drawn {...props} picture={column} words="opposite" /> },
		{ task: (props) => <Drawn {...props} picture={opened} words="holes" /> },
		{
			task: (props) => (
				<Drawn
					{...props}
					picture={{
						kind: "lattice",
						width: 3,
						height: 1,
						cells: [
							[0, 0],
							[1, 0],
							[2, 0],
						],
						marks: [
							{ at: [0, 0], label: "2" },
							{ at: [2, 0], label: "5" },
						],
					}}
					words="top"
				/>
			),
		},
	],
};

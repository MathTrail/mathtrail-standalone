import type { Kind, Picture } from "../model";

/**
 * Ahead is a kind the card draws before the service hands it out: the names
 * of its limits, which the service's list does not hold yet, and an example
 * of it, which the content does not show the model yet. A card of an earlier
 * release reads a kind it does not know as no picture, so the card that draws
 * a kind goes out before the service that hands it out, and until then the
 * card's tests and its preview take the kind and its example from here.
 */
export type Ahead = {
	kind: Kind;
	limits: readonly string[];
	example: { purpose: string; picture: Picture };
};

/** ahead are the kinds the card draws before the service hands them out. */
export const ahead: readonly Ahead[] = [
	{
		kind: "lattice",
		limits: [
			"minLatticeSide",
			"maxLatticeSide",
			"maxLatticeCells",
			"maxLatticeLines",
			"maxLatticeDashed",
			"maxLatticePoints",
			"maxLatticeMarks",
			"maxShortLabelCharacters",
		],
		example: {
			purpose:
				"A figure on squared paper: the walls of a maze from S to E, drawn as lines from corner to corner.",
			picture: {
				kind: "lattice",
				width: 3,
				height: 3,
				lines: [
					[1, 0, 1, 2],
					[2, 1, 2, 3],
				],
				marks: [
					{ at: [0, 0], label: "S" },
					{ at: [2, 2], label: "E" },
				],
			},
		},
	},
	{
		kind: "tiles",
		limits: [
			"minTileSides",
			"maxTileSides",
			"minTiles",
			"maxTiles",
			"maxTilesInRow",
			"maxHexagonsInRow",
			"maxTileDigits",
		],
		example: {
			purpose:
				"Identical tiles in a row, each cut into triangles with a number in some, and ? in the one asked about.",
			picture: {
				kind: "tiles",
				sides: 4,
				layout: "row",
				tiles: [
					["", "1", "2", "3"],
					["2", "", "", ""],
					["", "", "?", ""],
				],
			},
		},
	},
	{
		kind: "shapes",
		limits: ["minShapeRows", "maxShapeRows", "maxShapesInRow"],
		example: {
			purpose:
				"Rows of shapes with a total beside each, the same shape the same number, and ? for the total asked about.",
			picture: {
				kind: "shapes",
				rows: [
					{ shapes: ["triangle", "triangle"], total: "6" },
					{ shapes: ["triangle", "square"], total: "7" },
					{ shapes: ["square", "circle"], total: "?" },
				],
			},
		},
	},
	{
		kind: "stairs",
		limits: ["minSteps", "maxSteps", "maxStairMarks", "maxStairJumps"],
		example: {
			purpose:
				"A staircase whose steps the card does not number, A on the floor and B at the top, each with a jump.",
			picture: {
				kind: "stairs",
				steps: 8,
				marks: [
					{ step: 0, label: "A" },
					{ step: 8, label: "B" },
				],
				jumps: [
					{ from: 0, to: 2 },
					{ from: 8, to: 5 },
				],
			},
		},
	},
];

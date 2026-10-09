import type { Balance } from "./model";
import { type Drawn, Label, r1 } from "./shapes";
import { room, widthOf, written } from "./text";

// A balance, in the card's pixels: the size the labels of the weights are
// written at, the room around a label in its weight, how tall a weight is and
// how narrow it may be, the room between weights, how wide a pan may be at
// most and at least, how far its lip turns up, the room between the two pans,
// and the heights of the parts under the pans: the post from a pan to the
// beam, the stand from the beam to its foot.
const weightSize = 13;
const weightAside = 4;
const weightTall = 24;
const narrowestWeight = 26;
const between = 4;
const apart = 26;
const widestPan = (room - apart - 2) / 2;
const narrowestPan = 76;
const lip = 6;
const post = 14;
const stand = 34;

// Weight is one weight on a pan: what it says, how wide it is, and where it
// stands: its row from the pan up, and its middle across.
type Weight = {
	at: number;
	text: string;
	width: number;
	row: number;
	x: number;
};

// rowsOf are the weights on a pan in rows from the pan up, as many to a row
// as fit the widest pan, each row centred on the pan's middle at 0, and how
// wide the widest row is.
function rowsOf(labels: readonly string[]): {
	weights: Weight[];
	width: number;
} {
	const sized = labels.map((label, at) => {
		const text = written(label);
		return {
			at,
			text,
			width: Math.max(
				narrowestWeight,
				widthOf(text, weightSize) + 2 * weightAside,
			),
		};
	});
	const rowWidth = (row: readonly { width: number }[]) =>
		row.reduce((sum, one) => sum + one.width, 0) + between * (row.length - 1);
	let perRow = sized.length;
	while (
		perRow > 1 &&
		chunks(sized, perRow).some((row) => rowWidth(row) > widestPan - 2 * lip - 2)
	) {
		perRow--;
	}
	const rows = chunks(sized, Math.max(1, perRow));
	const weights = rows.flatMap((row, line) => {
		let x = -rowWidth(row) / 2;
		return row.map((one) => {
			const placed = { ...one, row: line, x: x + one.width / 2 };
			x += one.width + between;
			return placed;
		});
	});
	return { weights, width: Math.max(0, ...rows.map(rowWidth)) };
}

// chunks are the items of a list in runs of a size, the last run shorter.
function chunks<T>(items: readonly T[], size: number): T[][] {
	const runs: T[][] = [];
	for (let at = 0; at < items.length; at += size) {
		runs.push(items.slice(at, at + size));
	}
	return runs;
}

/**
 * drawBalance draws a pan balance at rest, its beam level, since the picture
 * is not told which side is heavier: a stand on a foot, the beam across its
 * top, a pan on a post at each end of the beam, and on each pan the weights
 * the description names, each a box with its label, in rows from the pan up.
 */
export function drawBalance(balance: Balance): Drawn {
	const left = rowsOf(balance.left);
	const right = rowsOf(balance.right);
	const pan = Math.min(
		widestPan,
		Math.max(narrowestPan, left.width + 2 * lip + 2, right.width + 2 * lip + 2),
	);
	const sides = [
		{ name: "left", weights: left.weights, centre: pan / 2 + 1 },
		{ name: "right", weights: right.weights, centre: pan * 1.5 + 1 + apart },
	];
	const rows = Math.max(
		1,
		...[...left.weights, ...right.weights].map((one) => one.row + 1),
	);
	const panAt = 2 + rows * (weightTall + between);
	const beamAt = panAt + post;
	const footAt = beamAt + stand;
	const width = 2 * pan + apart + 2;
	const pivot = width / 2;
	return {
		width,
		height: footAt + 2,
		body: (
			<>
				<path
					d={`M${r1(pivot)} ${r1(beamAt)}L${r1(pivot - 14)} ${r1(footAt)}H${r1(pivot + 14)}Z`}
					class="mt-pic-line"
					stroke-width={2}
					stroke-linejoin="round"
				/>
				<path
					d={`M${r1(pivot - 34)} ${r1(footAt)}H${r1(pivot + 34)}`}
					class="mt-pic-line"
					stroke-width={2.5}
				/>
				<path
					d={`M${r1(pan / 2 + 1)} ${r1(beamAt)}H${r1(pan * 1.5 + 1 + apart)}`}
					class="mt-pic-line"
					stroke-width={3}
					stroke-linecap="round"
				/>
				{sides.map(({ name, weights, centre }) => {
					return (
						<g key={name}>
							<path
								d={`M${r1(centre)} ${r1(beamAt)}V${r1(panAt)}`}
								class="mt-pic-line"
								stroke-width={2}
							/>
							<path
								d={
									`M${r1(centre - pan / 2)} ${r1(panAt - lip)}L${r1(centre - pan / 2 + lip)} ${r1(panAt)}` +
									`H${r1(centre + pan / 2 - lip)}L${r1(centre + pan / 2)} ${r1(panAt - lip)}`
								}
								class="mt-pic-line"
								stroke-width={2.5}
								stroke-linecap="round"
								stroke-linejoin="round"
							/>
							{weights.map((weight) => (
								<g key={weight.at}>
									<rect
										x={r1(centre + weight.x - weight.width / 2)}
										y={r1(
											panAt -
												1.25 -
												(weight.row + 1) * (weightTall + between) +
												between,
										)}
										width={r1(weight.width)}
										height={weightTall}
										rx={3}
										class="mt-pic-ring"
										stroke-width={1.5}
									/>
									<Label
										x={centre + weight.x}
										y={
											panAt -
											1.25 -
											(weight.row + 0.5) * (weightTall + between) +
											between / 2
										}
										text={weight.text}
										size={weightSize}
									/>
								</g>
							))}
						</g>
					);
				})}
			</>
		),
	};
}

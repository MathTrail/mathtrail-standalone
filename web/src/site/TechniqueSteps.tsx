import type { ComponentChildren } from "preact";
import { drawPicture, PictureFrame } from "../design/picture/diagram";
import type { Picture } from "../design/picture/model";
import type { Tone, Tones } from "../design/picture/tones";
import type { PageReader } from "./reader";

/**
 * StepDrawing is the drawing beside one step of a technique's solution, the
 * step counted from 0. Where a kind of picture the card draws shows what the
 * step does, it is that picture, its parts lit in the step's tones, and the
 * words a picture holds none of stand under it; elsewhere it is a drawing of
 * the site's own in markup. Its numbers are in its code and its words in the
 * page's YAML.
 */
export function StepDrawing({
	page,
	technique,
	step,
}: {
	page: PageReader;
	technique: string;
	step: number;
}) {
	const draw = stepPictures[technique];
	if (draw === undefined) {
		throw new Error(`the page of the techniques has no pictures for ${technique}`);
	}
	return <>{draw(page, `techniques.${technique}.pictures`, step)}</>;
}

/**
 * TaskLegend is what a problem's pictures mean, under the problem, for a
 * technique whose pictures need a key: the counters of hens and rabbits, who
 * of the two islanders said what, and the colours of the two players. A
 * technique whose pictures need none has none, and a screen reader passes it
 * over with the pictures it is the key to.
 */
export function TaskLegend({
	page,
	technique,
}: {
	page: PageReader;
	technique: string;
}) {
	const at = `techniques.${technique}.legend`;
	switch (technique) {
		case "all-the-same":
			return (
				<p class="s-technique-legend" aria-hidden="true">
					<span class="s-step-row">
						<Leg legs={2} />
						{page.text(`${at}.hen`)}
					</span>
					<span class="s-step-row">
						<Leg legs={4} />
						{page.text(`${at}.rabbit`)}
					</span>
				</p>
			);
		case "opposite":
			return (
				<p class="s-technique-legend" aria-hidden="true">
					<span class="s-step-row">
						<Who letter={page.plain(`${at}.anya`)} />
						<span class="s-step-said">{page.text(`${at}.anya-said`)}</span>
					</span>
					<span class="s-step-row">
						<Who letter={page.plain(`${at}.borya`)} />
						<span class="s-step-said">{page.text(`${at}.borya-said`)}</span>
					</span>
				</p>
			);
		case "mirror":
			return (
				<p class="s-technique-legend" aria-hidden="true">
					<span class="s-step-row">
						<span class="s-step-swatch" data-tone="cool" />
						{page.text(`${at}.first`)}
					</span>
					<span class="s-step-row">
						<span class="s-step-swatch" data-tone="warm" />
						{page.text(`${at}.second`)}
					</span>
					<span>{page.text(`${at}.moves`)}</span>
				</p>
			);
		default:
			return null;
	}
}

// Draw is how a technique's steps are pictured: from the page, the place of
// the technique's words for its pictures and the step, counted from 0.
type Draw = (page: PageReader, at: string, step: number) => ComponentChildren;

// Kind is a picture of a kind the card draws, its parts lit in tones, as the
// card draws it in the page's language.
function Kind({
	page,
	picture,
	tones,
}: {
	page: PageReader;
	picture: Picture;
	tones: Tones;
}) {
	return <PictureFrame drawn={drawPicture(picture, page.locale, tones)} />;
}

// Under is the line under a picture that says what it shows.
function Under({ children }: { children: ComponentChildren }) {
	return <span class="s-step-under">{children}</span>;
}

// lit lights every part named in the same tones.
function lit(names: readonly string[], ...tones: Tone[]): Tones {
	return Object.fromEntries(names.map((name) => [name, tones]));
}

// --- Draw the problem: the sister's age and the brother's, as bars ---------

// ages are the two ages as bars to one scale, the sister's over the
// brother's, which is as long and four years more, each piece under its
// label.
function ages(sister: string, brother: string, extra: string): Picture {
	return {
		kind: "bars",
		bars: [
			{ segments: [{ size: 6, label: sister }] },
			{
				segments: [
					{ size: 6, label: brother },
					{ size: 4, label: extra },
				],
			},
		],
	};
}

const drawDraw: Draw = (page, at, step) => {
	const equal = lit(["piece 0.0", "piece 1.0"], "cool");
	switch (step) {
		case 0:
			return (
				<>
					<Kind
						page={page}
						picture={ages("?", "?", "4")}
						tones={{ ...equal, "piece 1.1": ["warm"] }}
					/>
					<Under>{page.text(`${at}.sum`)}</Under>
				</>
			);
		case 1:
			return (
				<>
					<Kind
						page={page}
						picture={ages("?", "?", "−4")}
						tones={{ ...equal, "piece 1.1": ["struck"] }}
					/>
					<Under>{page.text(`${at}.less`)}</Under>
				</>
			);
		default:
			return (
				<>
					<Kind
						page={page}
						picture={ages("6", "6", "4")}
						tones={{
							"piece 0.0": ["cool"],
							"piece 1.0": ["cool", "picked"],
							"piece 1.1": ["warm", "picked"],
						}}
					/>
					<Under>{page.text(`${at}.found`)}</Under>
				</>
			);
	}
};

// --- Make a table: children against pets, a cross or a tick a step ---------

// children and pets are the rows and the columns of the table, by the names
// their words have.
const children = ["anya", "borya", "vera"] as const;
const pets = ["cat", "dog", "parrot"] as const;

// PetMark is a mark of the table: whose row and which pet's column it stands
// in, whether it rules the pet out or in, and the step of the solution that
// puts it there, counted from 1.
type PetMark = {
	child: (typeof children)[number];
	pet: (typeof pets)[number];
	yes: boolean;
	step: number;
};

// petMarks are the table's marks, step by step.
const petMarks: readonly PetMark[] = [
	{ child: "anya", pet: "cat", yes: false, step: 1 },
	{ child: "borya", pet: "cat", yes: false, step: 1 },
	{ child: "borya", pet: "dog", yes: false, step: 1 },
	{ child: "borya", pet: "parrot", yes: true, step: 2 },
	{ child: "anya", pet: "parrot", yes: false, step: 2 },
	{ child: "vera", pet: "parrot", yes: false, step: 2 },
	{ child: "anya", pet: "dog", yes: true, step: 3 },
	{ child: "vera", pet: "cat", yes: true, step: 3 },
	{ child: "vera", pet: "dog", yes: false, step: 3 },
];

// PetCell is one cell of the table as it stands after a step, counted from
// 1: empty, or its mark with the number of the step that put it there,
// outlined where that step is this one.
function PetCell({ mark, step }: { mark: PetMark | undefined; step: number }) {
	if (mark === undefined || mark.step > step) {
		return <span class="s-step-pets-cell" />;
	}
	return (
		<span
			class="s-step-pets-cell"
			data-mark={mark.yes ? "yes" : "no"}
			data-new={mark.step === step ? "" : undefined}
		>
			{mark.yes ? "✓" : "✕"}
			<sup>{mark.step}</sup>
		</span>
	);
}

const drawTable: Draw = (page, at, step) => (
	<span class="s-step-pets">
		<span />
		{pets.map((pet) => (
			<span key={pet} class="s-step-pets-head">
				{page.text(`${at}.pets.${pet}`)}
			</span>
		))}
		{children.map((child) => {
			const name = page.plain(`${at}.children.${child}`);
			return [
				<span key={child} class="s-step-pets-child">
					<span class="s-step-pets-face" data-child={child}>
						{name.charAt(0)}
					</span>
					{name}
				</span>,
				...pets.map((pet) => (
					<PetCell
						key={`${child}-${pet}`}
						mark={petMarks.find((one) => one.child === child && one.pet === pet)}
						step={step + 1}
					/>
				)),
			];
		})}
	</span>
);

// --- Start with small numbers: roads, then a table of them -----------------

// road is a road with a post every 5 metres, as long as length.
function road(length: number): Picture {
	return {
		kind: "row",
		items: Array.from({ length: length / 5 + 1 }, () => ({})),
		gaps: "5",
		span: String(length),
	};
}

const drawSmallNumbers: Draw = (page, at, step) => {
	const line: Tones = { line: ["cool"] };
	switch (step) {
		case 0:
			return (
				<>
					<Kind page={page} picture={road(10)} tones={line} />
					<Under>{page.text(`${at}.ten`)}</Under>
				</>
			);
		case 1:
			return (
				<>
					<Kind page={page} picture={road(15)} tones={line} />
					<Under>{page.text(`${at}.fifteen`)}</Under>
					<Kind page={page} picture={road(20)} tones={line} />
					<Under>{page.text(`${at}.twenty`)}</Under>
				</>
			);
		default:
			return (
				<>
					<Kind
						page={page}
						picture={{
							kind: "table",
							rows: [
								["10", "2", "3"],
								["15", "3", "4"],
								["20", "4", "5"],
								["100", "20", "21"],
							],
						}}
						tones={{ "cell 3.2": ["picked"] }}
					/>
					<Under>{page.text(`${at}.table`)}</Under>
				</>
			);
	}
};

// --- Enumerate in order: the tree of the four even digits ------------------

// digits are the digits of the problem; each keeps its colour in every
// number.
const digits = ["2", "4", "6", "8"] as const;

// Digit is one digit on a tile of its colour, at a size.
function Digit({
	digit,
	size,
}: {
	digit: string;
	size: "large" | "middle" | "small";
}) {
	return (
		<span class="s-step-digit" data-digit={digit} data-size={size}>
			{digit}
		</span>
	);
}

// Pair is a two-digit number as two tiles side by side.
function Pair({
	first,
	second,
	size,
}: {
	first: string;
	second: string;
	size: "middle" | "small";
}) {
	return (
		<span class="s-step-pair">
			<Digit digit={first} size={size} />
			<Digit digit={second} size={size} />
		</span>
	);
}

// others are the digits a number goes on with after first.
function others(first: string): string[] {
	return digits.filter((digit) => digit !== first);
}

const drawSystematic: Draw = (page, at, step) => {
	switch (step) {
		case 0:
			return (
				<>
					<span class="s-step-row">
						{digits.map((digit) => (
							<Digit key={digit} digit={digit} size="large" />
						))}
					</span>
					<Under>{page.text(`${at}.first`)}</Under>
				</>
			);
		case 1:
			return (
				<span class="s-step-branch">
					<Digit digit="2" size="large" />
					<span class="s-step-branch-stem" />
					<span class="s-step-branch-arms" />
					<span class="s-step-row">
						{others("2").map((second) => (
							<Pair key={second} first="2" second={second} size="middle" />
						))}
					</span>
				</span>
			);
		default:
			return (
				<>
					<span class="s-step-branches">
						{digits.map((first) => (
							<span key={first} class="s-step-column">
								{others(first).map((second) => (
									<Pair key={second} first={first} second={second} size="small" />
								))}
							</span>
						))}
					</span>
					<Under>{page.text(`${at}.total`)}</Under>
				</>
			);
	}
};

// --- Work backwards: apples, eaten forwards and given back -----------------

// Apples are a heap of apples, so many a row, the last of them ringed as the
// ones eaten from it.
function Apples({
	count,
	eaten = 0,
	across,
}: {
	count: number;
	eaten?: number;
	across: number;
}) {
	return (
		<span class="s-step-apples" style={{ "--s-across": String(across) }}>
			{Array.from({ length: count }, (_, at) => (
				<span
					key={at}
					class="s-step-apple"
					data-eaten={at >= count - eaten ? "" : undefined}
				/>
			))}
		</span>
	);
}

// Heap is a heap of apples with its count under it.
function Heap(props: { count: number; eaten?: number; across: number }) {
	return (
		<span class="s-step-heap">
			<Apples {...props} />
			<span class="s-step-count">{props.count}</span>
		</span>
	);
}

// Back is the arrow back from one heap to the one before it, under what is
// done to go back.
function Back({ children }: { children: ComponentChildren }) {
	return (
		<span class="s-step-back">
			{children}
			<span class="s-step-back-arrow">←</span>
		</span>
	);
}

// Box is a number in a box, ringed where it is the answer, or framed where
// it is the end a check comes back to.
function Box({
	children,
	picked = false,
	end = false,
}: {
	children: ComponentChildren;
	picked?: boolean;
	end?: boolean;
}) {
	return (
		<span
			class="s-step-box"
			data-picked={picked ? "" : undefined}
			data-end={end ? "" : undefined}
		>
			{children}
		</span>
	);
}

// Sign is a sign between the boxes of a sum or of a check.
function Sign({ children }: { children: ComponentChildren }) {
	return <span class="s-step-sign">{children}</span>;
}

const drawBackwards: Draw = (page, at, step) => {
	switch (step) {
		case 0:
			return (
				<span class="s-step-row">
					<Apples count={2} across={2} />
					<Under>{page.text(`${at}.end`)}</Under>
				</span>
			);
		case 1:
			return (
				<span class="s-step-row">
					<Heap count={6} eaten={4} across={3} />
					<Back>{page.text(`${at}.undo-petya`)}</Back>
					<Heap count={2} across={3} />
				</span>
			);
		default:
			return (
				<>
					<span class="s-step-row">
						<Heap count={12} eaten={6} across={3} />
						<Back>{page.text(`${at}.undo-tanya`)}</Back>
						<Heap count={6} across={3} />
						<Back>{page.text(`${at}.undo-petya`)}</Back>
						<Heap count={2} across={3} />
					</span>
					<span class="s-step-row s-step-check">
						<Under>{page.text(`${at}.check`)}</Under>
						<Box>12</Box>
						<Sign>→</Sign>
						<Box>6</Box>
						<Sign>→</Sign>
						<Box end>2</Box>
					</span>
				</>
			);
	}
};

// --- Suppose all the same: legs, two to a hen and four to a rabbit ---------

// Leg is the counter of one animal: a round one with the 2 legs of a hen, or
// a square one with the 4 of a rabbit, ringed where it is picked.
function Leg({ legs, picked = false }: { legs: 2 | 4; picked?: boolean }) {
	return (
		<span
			class="s-step-leg"
			data-legs={legs}
			data-picked={picked ? "" : undefined}
		>
			{legs}
		</span>
	);
}

// Legs are the counters of the yard: so many rabbits, ringed, then so many
// hens.
function Legs({ rabbits, hens }: { rabbits: number; hens: number }) {
	return (
		<span class="s-step-legs">
			{Array.from({ length: rabbits }, (_, at) => (
				<Leg key={`rabbit-${at}`} legs={4} picked />
			))}
			{Array.from({ length: hens }, (_, at) => (
				<Leg key={`hen-${at}`} legs={2} />
			))}
		</span>
	);
}

const drawAllTheSame: Draw = (page, at, step) => {
	switch (step) {
		case 0:
			return (
				<span class="s-step-row">
					<Legs rabbits={0} hens={10} />
					<Under>{page.text(`${at}.guess`)}</Under>
				</span>
			);
		case 1:
			return (
				<span class="s-step-row">
					<Box>28</Box>
					<Sign>−</Sign>
					<Box>20</Box>
					<Sign>=</Sign>
					<Box picked>8</Box>
					<Under>{page.text(`${at}.short`)}</Under>
				</span>
			);
		default:
			return (
				<>
					<Legs rabbits={4} hens={6} />
					<Under>{page.text(`${at}.fits`)}</Under>
				</>
			);
	}
};

// --- Look for a pattern: flags in fours ------------------------------------

// flagColours are the colours of the flags in the order they hang.
const flagColours = ["red", "yellow", "green", "blue"] as const;

// Flag is one flag on its pole, of a colour, framed where it is picked.
function Flag({
	colour,
	small = false,
	picked = false,
}: {
	colour: (typeof flagColours)[number];
	small?: boolean;
	picked?: boolean;
}) {
	return (
		<span
			class="s-step-flag"
			data-colour={colour}
			data-small={small ? "" : undefined}
			data-picked={picked ? "" : undefined}
		/>
	);
}

// Four is the four flags that repeat, framed in the accent where picked.
function Four({
	small = false,
	picked = false,
}: {
	small?: boolean;
	picked?: boolean;
}) {
	return (
		<span class="s-step-four" data-picked={picked ? "" : undefined}>
			{flagColours.map((colour) => (
				<Flag key={colour} colour={colour} small={small} />
			))}
		</span>
	);
}

const drawPattern: Draw = (page, at, step) => {
	switch (step) {
		case 0:
			return (
				<span class="s-step-row">
					<Four picked />
					<Four />
					<Under>…</Under>
				</span>
			);
		case 1:
			return (
				<>
					<span class="s-step-row">
						{Array.from({ length: 5 }, (_, at) => (
							<Four key={at} small />
						))}
						<span class="s-step-flags">
							{flagColours.slice(0, 3).map((colour) => (
								<Flag key={colour} colour={colour} small />
							))}
						</span>
					</span>
					<Under>{page.text(`${at}.sum`)}</Under>
				</>
			);
		default:
			return (
				<span class="s-step-row">
					<span class="s-step-flags">
						{flagColours.slice(0, 3).map((colour) => (
							<Flag key={colour} colour={colour} picked={colour === "green"} />
						))}
					</span>
					<Under>{page.text(`${at}.last`)}</Under>
				</span>
			);
	}
};

// --- Suppose the opposite: a chain of what follows -------------------------

// Who is an islander by the first letter of their name: plain where the
// problem names them, green as a knight and red as a liar.
function Who({
	letter,
	role,
}: {
	letter: string;
	role?: "knight" | "liar";
}) {
	return (
		<span class="s-step-who" data-role={role}>
			{letter}
		</span>
	);
}

// Then is the arrow from one link of a chain to the next.
function Then() {
	return <span class="s-step-then" />;
}

// Chain is a chain of what follows from a supposition, tinted by where it
// ends: in a contradiction, crossed, or in what fits, ticked.
function Chain({
	fits,
	children,
}: {
	fits: boolean;
	children: ComponentChildren;
}) {
	return (
		<span class="s-step-chain" data-fits={fits ? "" : undefined}>
			{children}
			<span class="s-step-end">{fits ? "✓" : "✕"}</span>
		</span>
	);
}

const drawOpposite: Draw = (page, at, step) => {
	const anya = page.plain("techniques.opposite.legend.anya");
	const borya = page.plain("techniques.opposite.legend.borya");
	switch (step) {
		case 0:
			return (
				<Chain fits={false}>
					<Who letter={borya} role="knight" />
					<Then />
					<Who letter={anya} role="knight" />
					<Then />
					<span class="s-step-said">{page.text(`${at}.said`)}</span>
				</Chain>
			);
		case 1:
			return (
				<Chain fits>
					<Who letter={borya} role="liar" />
					<Then />
					<span class="s-step-said">{page.text(`${at}.false`)}</span>
				</Chain>
			);
		default:
			return (
				<Chain fits>
					<span class="s-step-said">{page.text(`${at}.true`)}</span>
					<Then />
					<Who letter={anya} role="knight" />
				</Chain>
			);
	}
};

// --- Watch the parity: the line coloured, and the jumps --------------------

// parityLine is the grasshopper's number line, from −4 to 4, with a pointer
// over each of the numbers given.
function parityLine(marks: readonly number[]): Picture {
	return {
		kind: "number_line",
		from: -4,
		to: 4,
		marks: marks.map((at) => ({ at })),
	};
}

// parityTones light the line's numbers, the even cool and the odd warm, the
// pointer of the start, and the pointers of the cells the grasshopper can
// be on.
function parityTones(odd: readonly number[]): Tones {
	const numbers = Array.from({ length: 9 }, (_, place) => place - 4);
	const ticks = (even: boolean) =>
		numbers
			.filter((number) => (number % 2 === 0) === even)
			.map((number) => `tick ${number}`);
	return {
		...lit(ticks(true), "cool"),
		...lit(ticks(false), "warm"),
		"mark 0": ["start"],
		...lit(
			odd.map((number) => `mark ${number}`),
			"picked",
		),
	};
}

const drawParity: Draw = (page, at, step) => {
	switch (step) {
		case 0:
			return (
				<>
					<Kind page={page} picture={parityLine([0])} tones={parityTones([])} />
					<Under>{page.text(`${at}.colours`)}</Under>
				</>
			);
		case 1:
			return (
				<>
					<Kind
						page={page}
						picture={{
							kind: "table",
							header: ["1", "2", "3", "4", "5"],
							rows: [["", "", "", "", ""]],
						}}
						tones={{
							...lit(["cell 0.0", "cell 0.2", "cell 0.4"], "warm"),
							...lit(["cell 0.1", "cell 0.3"], "cool"),
						}}
					/>
					<Under>{page.text(`${at}.jumps`)}</Under>
				</>
			);
		default:
			return (
				<>
					<Kind
						page={page}
						picture={parityLine([-3, -1, 0, 1, 3])}
						tones={parityTones([-3, -1, 1, 3])}
					/>
					<Under>{page.text(`${at}.after`)}</Under>
				</>
			);
	}
};

// --- Rabbits and hutches: the faces of a die, and the throws ---------------

// pips are where the pips of each face of a die stand, on a grid of 3 by 3
// counted row by row from the top left.
const pips: readonly (readonly number[])[] = [
	[4],
	[0, 8],
	[0, 4, 8],
	[0, 2, 6, 8],
	[0, 2, 4, 6, 8],
	[0, 2, 3, 5, 6, 8],
];

// Die is one face of a die, its pips where they stand.
function Die({ face }: { face: readonly number[] }) {
	return (
		<span class="s-step-die">
			{Array.from({ length: 9 }, (_, at) => (
				<span key={at} data-pip={face.includes(at) ? "" : undefined} />
			))}
		</span>
	);
}

// Dice are the six faces of a die, each over the throws that land on it, if
// throws are given: a dot in the accent each, and a ringed red one for a
// throw that lands where one already has.
function Dice({ throws }: { throws?: readonly number[] }) {
	return (
		<span class="s-step-row">
			{pips.map((face, at) => (
				<span key={face.length} class="s-step-column">
					<Die face={face} />
					{throws !== undefined && (
						<span class="s-step-throws">
							{Array.from({ length: throws[at] ?? 0 }, (_, one) => (
								<span
									key={one}
									class="s-step-throw"
									data-again={one > 0 ? "" : undefined}
								/>
							))}
						</span>
					)}
				</span>
			))}
		</span>
	);
}

const drawPigeonhole: Draw = (page, at, step) => {
	switch (step) {
		case 0:
			return <Dice />;
		case 1:
			return (
				<>
					<Dice throws={[1, 1, 1, 1, 1, 1]} />
					<Under>{page.text(`${at}.throw`)}</Under>
				</>
			);
		default:
			return (
				<>
					<Dice throws={[1, 1, 1, 2, 1, 1]} />
					<Under>{page.text(`${at}.seventh`)}</Under>
				</>
			);
	}
};

// --- Mirror the moves: the strip of nine cells -----------------------------

// strip is the strip of nine cells, numbered over them, with the number of
// each move made in the cell it was made in.
function strip(moves: Readonly<Record<string, string>>): Picture {
	return {
		kind: "grid",
		rows: [""],
		cols: ["1", "2", "3", "4", "5", "6", "7", "8", "9"],
		marks: moves,
	};
}

const drawMirror: Draw = (page, at, step) => {
	switch (step) {
		case 0:
			return (
				<Kind
					page={page}
					picture={strip({ "5": "1" })}
					tones={{ "cell 5": ["cool", "picked"] }}
				/>
			);
		case 1:
			return (
				<>
					<Kind
						page={page}
						picture={strip({ "1": "2", "5": "1", "9": "3" })}
						tones={{
							"cell 1": ["warm"],
							"cell 5": ["cool"],
							"cell 9": ["cool", "picked"],
						}}
					/>
					<Under>{page.text(`${at}.answer`)}</Under>
				</>
			);
		default:
			return (
				<>
					<Kind
						page={page}
						picture={strip({ "1": "2", "3": "4", "5": "1", "7": "5", "9": "3" })}
						tones={{
							...lit(["cell 2", "cell 4", "cell 6", "cell 8"], "struck"),
							...lit(["cell 1", "cell 3"], "warm"),
							...lit(["cell 5", "cell 9"], "cool"),
							"cell 7": ["cool", "picked"],
						}}
					/>
					<Under>{page.text(`${at}.stuck`)}</Under>
				</>
			);
	}
};

// stepPictures are how each technique's steps are pictured, by the
// technique's name.
const stepPictures: Readonly<Record<string, Draw>> = {
	draw: drawDraw,
	table: drawTable,
	"small-numbers": drawSmallNumbers,
	systematic: drawSystematic,
	backwards: drawBackwards,
	"all-the-same": drawAllTheSame,
	pattern: drawPattern,
	opposite: drawOpposite,
	parity: drawParity,
	pigeonhole: drawPigeonhole,
	mirror: drawMirror,
};

import type { VNode } from "preact";
import { Chain, DigitTree, type DrawingProps } from "./Art";

// Labels are where a drawing's labels are: the page, and their key in its
// words.
type Labels = Pick<DrawingProps, "page" | "at">;

/**
 * TechniqueDrawing is the drawing of the technique called id, laid out from
 * its problem's numbers, its labels the page's words under at. A drawing is
 * no text to translate: where a cross stands and how many posts a road has
 * are the problem's, the same in every language. A technique the page has no
 * drawing for stops the build.
 */
export function TechniqueDrawing({ id, page, at }: Labels & { id: string }) {
	const Drawing = drawings[id];
	if (Drawing === undefined) {
		throw new Error(`the page of the techniques has no drawing for ${id}`);
	}
	return (
		<Drawing page={page} at={at} numbers={new Intl.NumberFormat(page.locale)} />
	);
}

// drawings are the techniques' drawings, by the technique's name.
const drawings: Readonly<Record<string, (props: DrawingProps) => VNode>> = {
	draw: Ages,
	table: PetsTable,
	"small-numbers": Posts,
	systematic: Tree,
	backwards: Backwards,
	"all-the-same": HeadsAndLegs,
	pattern: Flags,
	opposite: TwoPaths,
	parity: NumberLine,
	pigeonhole: Dice,
	mirror: Strip,
};

// Title is the label a drawing opens with.
function Title({ page, at }: Labels) {
	return <p class="s-art-title">{page.text(`${at}.title`)}</p>;
}

// Ages draws the sister's age as a bar and the brother's as the same bar and
// four years more, the two together sixteen; then, without the extra four
// years, two equal bars of six.
function Ages({ page, at, numbers }: DrawingProps) {
	return (
		<>
			<Title page={page} at={at} />
			<div class="s-ages">
				<span class="s-ages-name">{page.text(`${at}.sister`)}</span>
				<span class="s-ages-bars">
					<span class="s-ages-bar">?</span>
					<span class="s-ages-rest" />
				</span>
				<span class="s-ages-name">{page.text(`${at}.brother`)}</span>
				<span class="s-ages-bars">
					<span class="s-ages-bar">?</span>
					<span class="s-ages-extra">{numbers.format(4)}</span>
				</span>
				<span class="s-ages-total">{numbers.format(16)}</span>
			</div>
			<div class="s-art-part">
				<p class="s-art-title">{page.text(`${at}.without`)}</p>
				<span class="s-ages-bars s-ages-halves">
					<span class="s-ages-bar">{numbers.format(6)}</span>
					<span class="s-ages-bar">{numbers.format(6)}</span>
				</span>
				<p class="s-art-sum">{page.text(`${at}.count`)}</p>
			</div>
		</>
	);
}

// children and pets are the rows and the columns of the table, by the keys of
// their names in the page's words: a mark belongs to a child and a pet, not to
// a place in a list a translation might write in another order.
const children = ["anya", "borya", "vera"] as const;
const pets = ["cat", "dog", "parrot"] as const;

// petMarks are the table as the solution fills it in: for each child and pet,
// whether the pet is the child's, and the step that marked it so.
const petMarks: Readonly<
	Record<
		(typeof children)[number],
		Readonly<Record<(typeof pets)[number], readonly [boolean, number]>>
	>
> = {
	anya: { cat: [false, 1], dog: [true, 3], parrot: [false, 2] },
	borya: { cat: [false, 1], dog: [false, 1], parrot: [true, 2] },
	vera: { cat: [true, 3], dog: [false, 3], parrot: [false, 2] },
};

// PetsTable draws the children against the pets, every cell crossed out or
// ticked, with the step of the solution that marked it.
function PetsTable({ page, at, numbers }: DrawingProps) {
	return (
		<>
			<Title page={page} at={at} />
			<div class="s-pets">
				<span />
				{pets.map((pet) => (
					<span key={pet} class="s-pets-head">
						{page.text(`${at}.pets.${pet}`)}
					</span>
				))}
				{children.map((child) => [
					<span key={child} class="s-pets-name">
						{page.text(`${at}.children.${child}`)}
					</span>,
					...pets.map((pet) => {
						const [yes, step] = petMarks[child][pet];
						return (
							<span
								key={`${child}-${pet}`}
								class={yes ? "s-pets-cell s-pets-yes" : "s-pets-cell s-pets-no"}
							>
								{yes ? "✓" : "✕"}
								<span class="s-pets-step">{numbers.format(step)}</span>
							</span>
						);
					}),
				])}
			</div>
		</>
	);
}

// roads are the short roads the solution tries first, in metres, then the
// long one it is asked about.
const roads = [10, 15, 20] as const;
const road = 100;
const apart = 5;

// Posts draws the short roads with their posts, the gaps and the posts each
// has, and the long road with the rule they show: one post more than gaps.
function Posts({ page, at, numbers }: DrawingProps) {
	return (
		<div class="s-posts">
			<span />
			<span />
			<span class="s-posts-head">{page.text(`${at}.gaps`)}</span>
			<span class="s-posts-head">{page.text(`${at}.posts`)}</span>
			{roads.map((metres) => {
				const gaps = metres / apart;
				return [
					<span key={`${metres}-name`} class="s-posts-name">
						{page.text(`${at}.road`, { length: metres })}
					</span>,
					<span key={`${metres}-road`} class="s-posts-road">
						<span class="s-posts-line">
							{Array.from({ length: gaps + 1 }, (_, post) => (
								<span key={post} class="s-posts-post" />
							))}
						</span>
					</span>,
					<span key={`${metres}-gaps`} class="s-posts-count">
						{numbers.format(gaps)}
					</span>,
					<span key={`${metres}-posts`} class="s-posts-count s-art-accent">
						{numbers.format(gaps + 1)}
					</span>,
				];
			})}
			<span class="s-posts-name s-posts-last">
				{page.text(`${at}.road`, { length: road })}
			</span>
			<span class="s-posts-rule s-posts-last">{page.text(`${at}.rule`)}</span>
			<span class="s-posts-count s-posts-last">
				<span class="s-art-chip">{numbers.format(road / apart)}</span>
			</span>
			<span class="s-posts-count s-posts-last">
				<span class="s-art-chip s-art-chip-accent">
					{numbers.format(road / apart + 1)}
				</span>
			</span>
		</div>
	);
}

// digits are the digits the numbers of the tree are made of.
const digits = [2, 4, 6, 8] as const;

// Tree draws each first digit as a branch, with the numbers it begins.
function Tree({ page, at, numbers }: DrawingProps) {
	return (
		<>
			<Title page={page} at={at} />
			<DigitTree digits={digits} numbers={numbers} />
			<p class="s-art-sum">{page.text(`${at}.total`)}</p>
		</>
	);
}

// Backwards draws the story as it went, apples left after each eater, and the
// way back, each step undone.
function Backwards({ page, at, numbers }: DrawingProps) {
	return (
		<>
			<p class="s-art-title">{page.text(`${at}.forward`)}</p>
			<Chain
				start={numbers.format(12)}
				steps={[
					{ by: page.text(`${at}.tanya`), to: numbers.format(6) },
					{ by: page.text(`${at}.petya`), to: numbers.format(2) },
				]}
			/>
			<div class="s-art-part">
				<p class="s-art-title">{page.text(`${at}.backward`)}</p>
				<Chain
					start={numbers.format(12)}
					steps={[
						{ by: page.text(`${at}.undo-tanya`), to: numbers.format(6) },
						{ by: page.text(`${at}.undo-petya`), to: numbers.format(2) },
					]}
					back
					accent
				/>
			</div>
		</>
	);
}

// heads are the animals of the yard, the four rabbits first.
const heads = 10;
const rabbits = 4;

// HeadsAndLegs draws every head as a hen's first, two legs each, then the
// four that turn out to be rabbits, two legs more each.
function HeadsAndLegs({ page, at, numbers }: DrawingProps) {
	return (
		<>
			<p class="s-art-title">{page.text(`${at}.guess`)}</p>
			<div class="s-heads">
				{Array.from({ length: heads }, (_, head) => (
					<span key={head} class="s-heads-head">
						{numbers.format(2)}
					</span>
				))}
				<span class="s-heads-total">{page.text(`${at}.guess-legs`)}</span>
			</div>
			<div class="s-art-part">
				<p class="s-art-title">{page.text(`${at}.short`)}</p>
				<div class="s-heads">
					{Array.from({ length: heads }, (_, head) => (
						<span
							key={head}
							class={
								head < rabbits ? "s-heads-head s-heads-rabbit" : "s-heads-head"
							}
						>
							{numbers.format(head < rabbits ? 4 : 2)}
						</span>
					))}
					<span class="s-heads-total">{page.text(`${at}.real-legs`)}</span>
				</div>
				<p class="s-art-chips">
					<span class="s-art-chip s-art-chip-accent">
						{page.text(`${at}.rabbits`)} {numbers.format(rabbits)}
					</span>
					<span class="s-art-chip">
						{page.text(`${at}.hens`)} {numbers.format(heads - rabbits)}
					</span>
				</p>
			</div>
		</>
	);
}

// colours are the flags' colours in the order they repeat in, and asked is
// the flag the problem asks about.
const colours = ["red", "yellow", "green", "blue"] as const;
const asked = 23;

// Flags draws the flags in fours, up to the one asked about, which stands out.
function Flags({ page, at }: Labels) {
	const fours = Math.ceil(asked / colours.length);
	return (
		<>
			<Title page={page} at={at} />
			<div class="s-flags">
				{Array.from({ length: fours }, (_, four) => (
					<span key={four} class="s-flags-four">
						{colours.map((colour, place) => {
							const flag = four * colours.length + place + 1;
							return flag > asked ? null : (
								<span
									key={colour}
									class={`s-flag s-flag-${colour}${flag === asked ? " s-flag-asked" : ""}`}
								/>
							);
						})}
					</span>
				))}
			</div>
			<p class="s-art-sum">{page.text(`${at}.sum`)}</p>
			<p class="s-art-note">{page.text(`${at}.note`)}</p>
		</>
	);
}

// TwoPaths draws the two ways of supposing: the one that runs into a
// contradiction, and the one that fits.
function TwoPaths({ page, at }: Labels) {
	return (
		<div class="s-ways">
			<ol class="s-way s-way-dead">
				<li class="s-way-start">{page.text(`${at}.knight`)}</li>
				<li>{page.text(`${at}.knight-then`)}</li>
				<li>{page.text(`${at}.knight-but`)}</li>
				<li class="s-way-end">✕ {page.text(`${at}.contradiction`)}</li>
			</ol>
			<ol class="s-way s-way-fits">
				<li class="s-way-start">{page.text(`${at}.liar`)}</li>
				<li>{page.text(`${at}.liar-then`)}</li>
				<li class="s-way-end">✓ {page.text(`${at}.fits`)}</li>
			</ol>
		</div>
	);
}

// points are the cells of the number line drawn around zero, and jumps the
// grasshopper makes.
const points = [-4, -3, -2, -1, 0, 1, 2, 3, 4] as const;
const jumps = 5;

// NumberLine draws the line coloured even and odd, the grasshopper's start,
// and the colour of its cell after each jump.
function NumberLine({ page, at, numbers }: DrawingProps) {
	return (
		<>
			<Title page={page} at={at} />
			<div class="s-line">
				{points.map((point) => (
					<span
						key={point}
						class={`s-line-cell ${point % 2 === 0 ? "s-line-even" : "s-line-odd"}${point === 0 ? " s-line-start" : ""}`}
					>
						{numbers.format(point)}
					</span>
				))}
			</div>
			<p class="s-art-chips">
				<span class="s-art-chip s-line-even">{page.text(`${at}.even`)}</span>
				<span class="s-art-chip s-line-odd">{page.text(`${at}.odd`)}</span>
				<span class="s-art-chip s-line-start">{page.text(`${at}.start`)}</span>
			</p>
			<div class="s-art-part">
				<p class="s-art-title">{page.text(`${at}.after`)}</p>
				<div class="s-jumps">
					{Array.from({ length: jumps }, (_, jump) => (
						<span
							key={jump}
							class={`s-jumps-jump ${jump % 2 === 0 ? "s-line-odd" : "s-line-even"}`}
						>
							<span class="s-jumps-number">{numbers.format(jump + 1)}</span>
							{page.text(
								jump % 2 === 0 ? `${at}.odd-short` : `${at}.even-short`,
							)}
						</span>
					))}
				</div>
				<p class="s-art-note">{page.text(`${at}.end`)}</p>
			</div>
		</>
	);
}

// faces are the numbers a die shows, and repeated the one the seventh throw
// happens to bring again.
const faces = [1, 2, 3, 4, 5, 6] as const;
const repeated = 3;

// Dice draws each number of the die as a hole holding one of the first six
// throws, and the seventh throw landing in a hole already taken.
function Dice({ page, at, numbers }: DrawingProps) {
	return (
		<>
			<Title page={page} at={at} />
			<div class="s-holes">
				{faces.map((face) => (
					<span
						key={face}
						class={face === repeated ? "s-hole s-hole-twice" : "s-hole"}
					>
						<span class="s-hole-face">{numbers.format(face)}</span>
						<span class="s-hole-throw" />
						{face === repeated && <span class="s-hole-throw s-hole-seventh" />}
					</span>
				))}
			</div>
			<p class="s-art-chips">
				<span class="s-art-chip">{page.text(`${at}.first`)}</span>
				<span class="s-art-chip s-art-chip-accent">
					{page.text(`${at}.seventh`)}
				</span>
			</p>
		</>
	);
}

// game is a game on the strip of nine cells played to its end: the cell of
// each move in turn, the first player's first, every reply mirrored.
const game: readonly number[] = [5, 1, 9, 3, 7];
const cells = 9;

// playerOf is the class of whoever made the move counted from 0: the first
// player makes every other move, starting with the first.
function playerOf(move: number): string {
	return move % 2 === 0 ? "s-strip-first" : "s-strip-second";
}

// Strip draws the strip with each counter's move, the first player's and the
// opponent's apart.
function Strip({ page, at, numbers }: DrawingProps) {
	return (
		<>
			<Title page={page} at={at} />
			<div class="s-strip">
				{Array.from({ length: cells }, (_, cell) => {
					const move = game.indexOf(cell + 1);
					return move < 0 ? (
						<span key={cell} class="s-strip-cell" />
					) : (
						<span key={cell} class={`s-strip-cell ${playerOf(move)}`}>
							{numbers.format(move + 1)}
						</span>
					);
				})}
			</div>
			<p class="s-art-chips">
				<span class="s-art-chip s-strip-first">{page.text(`${at}.first`)}</span>
				<span class="s-art-chip s-strip-second">
					{page.text(`${at}.second`)}
				</span>
				<span class="s-art-chip">{page.text(`${at}.moves`)}</span>
			</p>
		</>
	);
}

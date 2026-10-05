/**
 * sharedFrom is the media query of the windows wide enough for the steps of a
 * lesson to share one card: below it every step keeps its own, as the page was
 * built.
 */
export const sharedFrom = "(min-width: 960px)";

/**
 * repliesFrom is how far below the top of its frame the card of a wrong answer
 * shows its replies, in pixels: the verdict and the trap in sight, and the
 * options they speak of just above them.
 */
export const repliesFrom = 180;

/** Box is how a step stands in the window: its top, bottom and height. */
export type Box = { top: number; bottom: number; height: number };

/**
 * stepAtMiddle is the index of the step nearest the middle of the window, a
 * window viewportHeight tall: the one the middle falls in, or else the one
 * whose edge is nearest it, the first of two as near. A step that takes no
 * room is never it, and with none that takes room there is none.
 */
export function stepAtMiddle(
	boxes: readonly Box[],
	viewportHeight: number,
): number | undefined {
	const middle = viewportHeight / 2;
	let nearest: number | undefined;
	let distance = Number.POSITIVE_INFINITY;
	boxes.forEach((box, at) => {
		if (box.height === 0) {
			return;
		}
		const off =
			box.top > middle
				? box.top - middle
				: box.bottom < middle
					? middle - box.bottom
					: 0;
		if (off < distance) {
			distance = off;
			nearest = at;
		}
	});
	return nearest;
}

// Staged is a lesson whose steps share one card: the steps, the card of each
// in its layer, and how to put them all back.
type Staged = {
	readonly steps: readonly HTMLElement[];
	readonly layers: readonly HTMLElement[];
	readonly undo: () => void;
};

/**
 * shareTheCard gives the steps of the lesson one card on a wide window. The
 * steps run down the start column, and beside them, in a column that stays in
 * sight, stands the card of the step at the middle of the window, in place of
 * the one before it as the page scrolls. The cards are the ones the page was
 * built with, each moved into a layer of that column and back again, never
 * copied: still, as on the page, and silent to a screen reader, which reads
 * the steps. On a narrower window, or once the window narrows, every step
 * keeps its own card, where the page put it. It returns what stops it and puts
 * the page back as it was.
 */
export function shareTheCard(document: Document, window: Window): () => void {
	const list = document.querySelector<HTMLElement>("#lesson .s-walk");
	if (list === null) {
		return () => {};
	}
	const wide = window.matchMedia(sharedFrom);
	let staged: Staged | undefined;
	let frame = 0;

	const follow = () => {
		if (staged === undefined || frame !== 0) {
			return;
		}
		frame = window.requestAnimationFrame(() => {
			frame = 0;
			if (staged !== undefined) {
				show(staged, window);
			}
		});
	};
	const fitTheWidth = () => {
		if (wide.matches && staged === undefined) {
			staged = stage(document, list);
			show(staged, window);
			alignAll(staged.layers);
		} else if (!wide.matches && staged !== undefined) {
			staged.undo();
			staged = undefined;
		}
	};
	const resized = () => {
		if (staged !== undefined) {
			alignAll(staged.layers);
		}
		follow();
	};

	fitTheWidth();
	wide.addEventListener("change", fitTheWidth);
	window.addEventListener("scroll", follow, { passive: true });
	window.addEventListener("resize", resized, { passive: true });
	return () => {
		wide.removeEventListener("change", fitTheWidth);
		window.removeEventListener("scroll", follow);
		window.removeEventListener("resize", resized);
		if (frame !== 0) {
			window.cancelAnimationFrame(frame);
		}
		staged?.undo();
		staged = undefined;
	};
}

// stage moves the card of every step of list into a layer of a column beside
// the steps, and returns the staged lesson: a list may hold nothing but its
// steps, so the steps and the column go into a box of their own.
function stage(document: Document, list: HTMLElement): Staged {
	const steps = [
		...list.querySelectorAll<HTMLElement>(":scope > .s-walk-step"),
	];
	const box = document.createElement("div");
	box.className = "s-walk-staged";
	const column = document.createElement("div");
	column.className = "s-walk-stage";
	list.before(box);
	box.append(list, column);
	const layers: HTMLElement[] = [];
	const homes: { card: HTMLElement; step: HTMLElement }[] = [];
	for (const step of steps) {
		const card = step.querySelector<HTMLElement>(":scope > .s-walk-card");
		if (card === null) {
			continue;
		}
		homes.push({ card, step });
		layers.push(card);
		column.append(card);
	}
	return {
		steps,
		layers,
		undo: () => {
			for (const { card, step } of homes) {
				card.removeAttribute("data-shown");
				step.append(card);
			}
			box.before(list);
			box.remove();
		},
	};
}

// show shows the layer of the step at the middle of the window, and hides the
// rest.
function show(staged: Staged, window: Window): void {
	const at = stepAtMiddle(
		staged.steps.map((step) => step.getBoundingClientRect()),
		window.innerHeight,
	);
	staged.layers.forEach((layer, index) => {
		layer.toggleAttribute("data-shown", index === (at ?? 0));
	});
}

// alignAll shows in each layer the part of its card that its step speaks of.
function alignAll(layers: readonly HTMLElement[]): void {
	for (const layer of layers) {
		align(layer);
	}
}

// align shows the part of a layer's card its step speaks of, the frame of the
// chat being shorter than the card: the question and the reply under a card,
// the replies of a card whose answer is told, and the top of any other.
function align(layer: HTMLElement): void {
	const body = layer.querySelector<HTMLElement>(".s-frame-body");
	if (body === null) {
		return;
	}
	if (layer.querySelector(".s-chat") !== null) {
		body.scrollTop = body.scrollHeight;
		return;
	}
	const replies = layer.querySelector<HTMLElement>(".mt-replies .mt-reply");
	if (replies !== null) {
		const from =
			replies.getBoundingClientRect().top -
			body.getBoundingClientRect().top +
			body.scrollTop;
		body.scrollTop = Math.max(0, from - repliesFrom);
		return;
	}
	body.scrollTop = 0;
}

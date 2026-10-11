import { useContext, useEffect, useRef } from "preact/hooks";
import { CardIsTheDocument } from "./focus";

/**
 * heldFor is how long a card keeps its top in the chat's sight, in
 * milliseconds from the moment what it shows became new: longer than nine
 * replies in ten take to bring a task and write the next one ahead, a try
 * turned down among them.
 */
export const heldFor = 120_000;

/**
 * useTopHeldInSight keeps the top of a card that is the whole page in the
 * chat's sight while the reply that brought what it shows may still be
 * coming: for heldFor from since, the moment it became new, and on from a
 * later moment the card is given, without losing what it has seen. A host may
 * scroll its chat on under a reply still at work, once the reply outgrows the
 * room under the message it answers, until the reply's end stands high on the
 * screen; a card taller than that room is then left showing the child its
 * foot. A card whose top has been in sight and is then moved up out of it
 * brings the top back, once: a chat moved by something other than its own
 * scrolling stops scrolling for that reply, and a later move is the reader's.
 * A reader who scrolls over the card ends the hold for good, and a card that
 * is part of a page, or whose moment is unknown or long past, holds nothing.
 */
export function useTopHeldInSight(since: number | undefined): void {
	const whole = useContext(CardIsTheDocument);
	// The hold under way, which a later moment lengthens rather than starts
	// again, and whether the card holds no more: its top brought back, or its
	// reader scrolling over it, whatever new it shows after.
	const holding = useRef<Holding | undefined>(undefined);
	const over = useRef(false);
	useEffect(() => {
		const left =
			whole && !over.current && since !== undefined
				? timeLeft(since, Date.now())
				: undefined;
		if (left === undefined || typeof IntersectionObserver === "undefined") {
			holding.current?.end();
			return;
		}
		holding.current ??= holdTop(document, (forGood) => {
			holding.current = undefined;
			over.current ||= forGood;
		});
		holding.current.lasts(left);
	}, [whole, since]);
	useEffect(() => () => holding.current?.end(), []);
}

/**
 * timeLeft is how long a hold that began at since has left at now, both in
 * milliseconds since the epoch, or undefined when it has none: the moment
 * long past, unknown, or so far ahead that the device's clock and the
 * service's cannot both be right.
 */
export function timeLeft(since: number, now: number): number | undefined {
	const age = now - since;
	if (!Number.isFinite(age) || age >= heldFor || age < -heldFor) {
		return undefined;
	}
	return heldFor - Math.max(0, age);
}

/**
 * TopSight is where a card's top stands in the chat's sight: whole in it,
 * cut by its edge above, cut by its edge below, or out of it.
 */
export type TopSight = "in" | "above" | "below" | "out";

/**
 * Observed is what an observation of the band along a card's top tells: the
 * band's box, and the part of it in sight, in the card's own frame.
 */
export type Observed = Pick<
	IntersectionObserverEntry,
	| "isIntersecting"
	| "intersectionRatio"
	| "intersectionRect"
	| "boundingClientRect"
>;

// cutAtLeast is how far the part of the band in sight has to start below the
// band for the edge above to cut it: a frame measures its boxes in fractions
// of a pixel, and a band whole in sight has its part in sight start where it
// does. allInSight is how much of the band has to be in sight for all of it
// to be, a band cut by a fraction at its foot among them.
const cutAtLeast = 0.01;
const allInSight = 0.99;

/**
 * topSightOf is where the band along a card's top stands, from an
 * observation of it. The part in sight starts lower than the band itself only
 * where the edge above cuts it: a card framed by another site's page learns
 * nothing more of the chat than that, and nothing at all of where in the
 * chat's sight it stands.
 */
export function topSightOf(observed: Observed): TopSight {
	if (!observed.isIntersecting) {
		return "out";
	}
	if (
		observed.intersectionRect.top >
		observed.boundingClientRect.top + cutAtLeast
	) {
		return "above";
	}
	return observed.intersectionRatio >= allInSight ? "in" : "below";
}

/**
 * Hold is what a card holding its top has seen: whether its top has been in
 * sight since the hold began, where it stands now, and whether any of the
 * card is in sight.
 */
export type Hold = { seen: boolean; top: TopSight; card: boolean };

/**
 * bringsBack says whether a card holding its top brings it back now: the top
 * has been in sight, and has since been moved up out of it — cut by the edge
 * above, or gone while the rest of the card is still in sight, which it can
 * only be above. A card never in sight with its top whole is left where the
 * chat keeps it, and so is a card wholly out of sight.
 */
export function bringsBack(hold: Hold): boolean {
	return (
		hold.seen && (hold.top === "above" || (hold.top === "out" && hold.card))
	);
}

/**
 * bandSight is where a band's sight is told: as it leaves and enters sight,
 * as it comes and goes whole, and at each quarter of it between, a band whose
 * sight starts a fraction short of whole among them, so that the edge above
 * is seen as soon as it cuts in.
 */
export const bandSight = [0, 0.25, 0.5, 0.75, 0.99, 1];

// scrollKeys are the keys that scroll a chat whatever has the focus in the
// card; Space scrolls it too, unless a control takes it. A card's frame does
// not scroll, so a key pressed in it scrolls the chat.
const scrollKeys = new Set([
	"PageUp",
	"PageDown",
	"Home",
	"End",
	"ArrowUp",
	"ArrowDown",
]);

// controls are what takes Space for itself rather than scrolling.
const controls = "button, input, select, textarea, [contenteditable]";

/**
 * isReaderScroll says whether event is the reader scrolling the chat over the
 * card: a wheel turned, a key that scrolls, or a touch the browser has taken
 * over to scroll with, which it then ends for the card. A tap, its finger
 * drifting as it presses an option, is no scroll.
 */
export function isReaderScroll(event: Event): boolean {
	switch (event.type) {
		case "wheel":
		case "pointercancel":
			return true;
		case "keydown": {
			const { key, target } = event as KeyboardEvent;
			const control = target instanceof Element && target.closest(controls);
			return scrollKeys.has(key) || (key === " " && !control);
		}
		default:
			return false;
	}
}

// readerEvents are the events a reader may scroll the chat with over the card;
// isReaderScroll says which of them did.
const readerEvents = ["wheel", "pointercancel", "keydown"] as const;

// Holding is a hold under way: lasts sets how many milliseconds from now it
// lasts, and end ends it.
type Holding = { lasts: (ms: number) => void; end: () => void };

// holdTop holds the top of page in the chat's sight until lasts runs out, and
// calls ended once the hold is over: for good when the top was brought back or
// the reader scrolled over the card, and not for good when its time ran out or
// end was called. A band along the page's top edge tells where the top stands,
// and the page's own edge whether any of the card is in sight.
function holdTop(page: Document, ended: (forGood: boolean) => void): Holding {
	const band = page.createElement("div");
	band.className = "mt-top-band";
	band.setAttribute("aria-hidden", "true");
	page.body.prepend(band);
	const hold: Hold = { seen: false, top: "out", card: false };
	let timer: ReturnType<typeof setTimeout> | undefined;
	let stopped = false;
	const observer = new IntersectionObserver(
		(entries) => {
			for (const entry of entries) {
				if (entry.target === band) {
					hold.top = topSightOf(entry);
					hold.seen ||= hold.top === "in";
				} else {
					hold.card = entry.isIntersecting;
				}
			}
			if (bringsBack(hold)) {
				// At once rather than smoothly: a host gliding the chat sets where
				// it stands frame by frame, which would cut a smooth scroll short,
				// and stops gliding once it finds the chat where it did not put it.
				band.scrollIntoView({ block: "start", behavior: "instant" });
				stop(true);
			}
		},
		{ threshold: bandSight },
	);
	const scrolled = (event: Event) => {
		if (isReaderScroll(event)) {
			stop(true);
		}
	};
	function stop(forGood: boolean) {
		if (stopped) {
			return;
		}
		stopped = true;
		observer.disconnect();
		clearTimeout(timer);
		band.remove();
		for (const type of readerEvents) {
			page.removeEventListener(type, scrolled);
		}
		ended(forGood);
	}
	observer.observe(band);
	observer.observe(page.documentElement);
	for (const type of readerEvents) {
		page.addEventListener(type, scrolled, { passive: true });
	}
	return {
		lasts(ms) {
			clearTimeout(timer);
			timer = setTimeout(() => stop(false), ms);
		},
		end: () => stop(false),
	};
}

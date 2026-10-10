/**
 * waitFor is how long a picture of the page's history stays before the next
 * one comes, in milliseconds.
 */
export const waitFor = 5000;

/**
 * nextOf is the picture that comes after the one at, of count: the next, and
 * the first again after the last.
 */
export function nextOf(at: number, count: number): number {
	return count === 0 ? 0 : (at + 1) % count;
}

/**
 * playTheHistory moves the pictures of the page's history on by themselves:
 * every waitFor it checks the chip of the next picture, after the last the
 * first, and the page's rules show it. It waits while a pointer rests on the
 * pictures or the eras, or the keyboard's focus is in them, and goes on with
 * the time that was left. A chip picked by hand, or an era's words pressed,
 * which pick the era's first picture, start the wait anew. For a reader who asks for
 * less motion nothing moves on by itself, and the era's words still pick. It
 * marks the history as live, its eras' words pressable, and as playing, which
 * shows the bars that fill as a picture waits; and it returns what stops it
 * and puts the page back as it was.
 */
export function playTheHistory(
	document: Document,
	window: Window,
): () => void {
	const root = document.getElementById("history");
	const split = root?.querySelector<HTMLElement>(".s-history-split");
	const picks = [
		...(root?.querySelectorAll<HTMLInputElement>(".s-history-pick") ?? []),
	];
	if (!root || !split || picks.length === 0) {
		return () => {};
	}
	const still = window.matchMedia("(prefers-reduced-motion: reduce)").matches;
	let timer: number | undefined;
	let left = waitFor;
	let since = 0;
	let held = false;

	const wait = () => {
		window.clearTimeout(timer);
		timer = undefined;
		if (!still && !held) {
			since = Date.now();
			timer = window.setTimeout(next, Math.max(400, left));
		}
	};
	const restart = () => {
		left = waitFor;
		wait();
	};
	const next = () => {
		const at = picks.findIndex((pick) => pick.checked);
		const pick = picks[nextOf(at, picks.length)];
		if (pick !== undefined) {
			pick.checked = true;
		}
		restart();
	};
	const hold = () => {
		if (held) {
			return;
		}
		held = true;
		if (timer !== undefined) {
			left -= Date.now() - since;
		}
		window.clearTimeout(timer);
		timer = undefined;
		root.toggleAttribute("data-held", true);
	};
	const release = () => {
		if (!held) {
			return;
		}
		held = false;
		root.toggleAttribute("data-held", false);
		wait();
	};
	// focused is whether the keyboard's focus is in the pictures or the eras:
	// a chip pressed with a pointer holds nothing once the pointer leaves.
	const focused = () => split.querySelector(":focus-visible") !== null;
	const focus = () => {
		if (focused()) {
			hold();
		}
	};
	const leave = () => {
		if (!focused()) {
			release();
		}
	};
	const blur = () => {
		if (!split.matches(":hover")) {
			release();
		}
	};
	// pickEra picks the first picture of the era whose words were pressed,
	// unless a picture of that era is already shown.
	const pickEra = (event: Event) => {
		const words = (event.target as Element | null)?.closest<HTMLElement>(
			"[data-first-picture]",
		);
		const era = words?.closest(".s-history-era");
		if (!words || !era || era.querySelector(".s-history-pick:checked")) {
			return;
		}
		const first = picks.find((pick) => pick.value === words.dataset.firstPicture);
		if (first !== undefined) {
			first.checked = true;
			restart();
		}
	};

	split.addEventListener("pointerenter", hold);
	split.addEventListener("pointerleave", leave);
	split.addEventListener("focusin", focus);
	split.addEventListener("focusout", blur);
	split.addEventListener("change", restart);
	split.addEventListener("click", pickEra);
	root.style.setProperty("--s-history-wait", `${waitFor / 1000}s`);
	root.toggleAttribute("data-live", true);
	root.toggleAttribute("data-playing", !still);
	wait();
	return () => {
		window.clearTimeout(timer);
		split.removeEventListener("pointerenter", hold);
		split.removeEventListener("pointerleave", leave);
		split.removeEventListener("focusin", focus);
		split.removeEventListener("focusout", blur);
		split.removeEventListener("change", restart);
		split.removeEventListener("click", pickEra);
		root.style.removeProperty("--s-history-wait");
		root.toggleAttribute("data-live", false);
		root.toggleAttribute("data-playing", false);
		root.toggleAttribute("data-held", false);
	};
}

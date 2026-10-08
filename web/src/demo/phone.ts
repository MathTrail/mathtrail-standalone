/**
 * heldFrom is the media query of the windows the first screen may be held in
 * place on while the page's scroll moves the chat in its phone: wide enough
 * for the phone to stand beside the words, and tall enough for a phone. The
 * stylesheet holds the first screen by the same query.
 */
export const heldFrom = "(min-width: 992px) and (min-height: 700px)";

/**
 * idleFor is how far the page scrolls, idle, once the chat in the phone has
 * reached its end and before the page scrolls on, on a window viewportHeight
 * tall: half the window, and never less than 320 pixels.
 */
export function idleFor(viewportHeight: number): number {
	return Math.max(320, Math.round(viewportHeight / 2));
}

/**
 * chatAt is how far the chat in the phone stands scrolled once the page has
 * scrolled scrolled pixels into the first screen's track, the chat running
 * past its screen by overflow: as far as the page has scrolled, and never
 * past the chat's end.
 */
export function chatAt(scrolled: number, overflow: number): number {
	return Math.min(Math.max(0, overflow), Math.max(0, scrolled));
}

/**
 * holdTheFirstScreen moves the chat in the phone on the first screen with the
 * page's scroll. On a window heldFrom matches, when the first screen held
 * stands whole under the menu, the stylesheet holds it in place while the
 * page scrolls through the track around it, and the chat scrolls as far as
 * the page has, to its end; then the page scrolls on, idle, for idleFor of
 * the window, and the first screen goes on up with the page. The track is
 * given that room, measured again in the window's next frame whenever the
 * window, the first screen or what the chat holds changes size, once however
 * often it changes before then. The page's scroll is where the chat stands,
 * both ways: when the chat is scrolled by anything else — the focus reaching
 * a button below its edge, the browser finding a word in it — the page
 * scrolls to where the chat now is; and when the chat changes size while the
 * first screen is in sight — an answer told, another task asked for — the
 * page scrolls so that the chat stays where it stands, the first screen held
 * around it. Otherwise nothing is held and the chat scrolls by itself; a window
 * that changes size is followed either way. It returns what stops it and
 * puts the page back as it was.
 */
export function holdTheFirstScreen(
	document: Document,
	window: Window & Pick<typeof globalThis, "ResizeObserver">,
): () => void {
	const track = document.querySelector<HTMLElement>(".s-hero-track");
	const hero = track?.querySelector<HTMLElement>(":scope > .s-hero");
	const chat = track?.querySelector<HTMLElement>(".s-phone .s-frame-body");
	if (!track || !hero || !chat) {
		return () => {};
	}
	const wide = window.matchMedia(heldFrom);
	// given is where the script last set the chat: a chat that stands
	// anywhere else was scrolled by something else.
	let given = chat.scrollTop;
	// following and sizing are the window's frames the chat is followed and
	// the first screen measured in, while one is due.
	let following = 0;
	let sizing = 0;

	const held = () => track.dataset.pinned !== undefined;
	// scrolled is how far the page has scrolled into the track, once the
	// first screen is held: how far the track's top has gone up past the line
	// the stylesheet holds the first screen at.
	const scrolled = () =>
		Number.parseFloat(window.getComputedStyle(hero).top) -
		track.getBoundingClientRect().top;
	const release = () => {
		track.toggleAttribute("data-pinned", false);
		track.style.removeProperty("--s-hero-room");
	};
	const give = (top: number) => {
		chat.scrollTop = top;
		given = chat.scrollTop;
	};
	// standAt scrolls the page to at into the track, and the chat to at.
	const standAt = (at: number) => {
		const by = at - scrolled();
		if (Math.abs(by) >= 1) {
			window.scrollBy({ top: by, behavior: "instant" });
		}
		give(at);
	};
	const follow = () => {
		if (!held()) {
			return;
		}
		const overflow = overflowOf(chat);
		// A chat that stands at its end below where it was set was cut short
		// by what it holds rather than scrolled, which measure answers for.
		const cut = given > overflow && chat.scrollTop >= overflow - 1;
		if (!cut && Math.abs(chat.scrollTop - given) >= 1) {
			standAt(chat.scrollTop);
			return;
		}
		give(chatAt(scrolled(), overflow));
	};
	const measure = () => {
		// The first screen is in sight while the track's end is still below
		// the line the first screen is held at.
		const seen = held() && scrolled() < track.offsetHeight;
		// The first screen is laid out held to be measured held, and let go
		// again in the same step when it does not stand whole.
		track.toggleAttribute("data-pinned", wide.matches);
		if (!wide.matches || !standsWhole(hero, window)) {
			release();
			return;
		}
		const into = scrolled();
		const overflow = overflowOf(chat);
		const idle = idleFor(window.innerHeight);
		track.style.setProperty(
			"--s-hero-room",
			`${hero.offsetHeight + overflow + idle}px`,
		);
		// Where the chat stands, past its end no more, against where the page
		// would set it: a chat grown under the page's hand would jump, and a
		// first screen past its new room would go up with the page, the chat
		// out of sight.
		const stands = Math.min(chat.scrollTop, overflow);
		const jumps = Math.abs(chatAt(into, overflow) - stands) >= 1;
		if (seen && (jumps || into > overflow + idle)) {
			standAt(stands);
		} else {
			give(chatAt(into, overflow));
		}
	};
	const followSoon = () => {
		if (following !== 0) {
			return;
		}
		following = window.requestAnimationFrame(() => {
			following = 0;
			follow();
		});
	};
	// The first screen is measured in the window's next frame rather than in
	// the observer's own call: letting it go there would change a size the
	// observer has just reported, which the observer takes for a loop.
	const measureSoon = () => {
		if (sizing !== 0) {
			return;
		}
		sizing = window.requestAnimationFrame(() => {
			sizing = 0;
			measure();
		});
	};

	// The sizes are watched once everything that follows the page is in
	// place: a browser that refuses a listener leaves the page as built.
	const sizes = new window.ResizeObserver(measureSoon);
	wide.addEventListener("change", measureSoon);
	window.addEventListener("resize", measureSoon, { passive: true });
	window.addEventListener("scroll", followSoon, { passive: true });
	chat.addEventListener("scroll", followSoon, { passive: true });
	sizes.observe(hero);
	for (const part of chat.children) {
		sizes.observe(part);
	}
	measure();
	return () => {
		sizes.disconnect();
		wide.removeEventListener("change", measureSoon);
		window.removeEventListener("resize", measureSoon);
		window.removeEventListener("scroll", followSoon);
		chat.removeEventListener("scroll", followSoon);
		for (const due of [following, sizing]) {
			if (due !== 0) {
				window.cancelAnimationFrame(due);
			}
		}
		release();
	};
}

// standsWhole says whether the first screen, laid out held, stands whole under
// the menu: no taller than the room the stylesheet gives it, the window's
// height below the menu, which taller words than that would push past.
function standsWhole(hero: HTMLElement, window: Window): boolean {
	const room = Number.parseFloat(window.getComputedStyle(hero).minHeight);
	return hero.offsetHeight <= room + 1;
}

// overflowOf is how far the chat runs past its screen.
function overflowOf(chat: HTMLElement): number {
	return Math.max(0, chat.scrollHeight - chat.clientHeight);
}

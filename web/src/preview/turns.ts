/**
 * Turns lets a few of many do something at once: take waits until one of the
 * turns is free and hands it over, with the end of the turn, which hands it on
 * to whoever has waited longest. A turn ends once, however often its end is
 * called.
 */
export type Turns = { take(): Promise<() => void> };

/** turns are count turns, none of them taken. */
export function turns(count: number): Turns {
	let free = count;
	const waiting: (() => void)[] = [];
	const ended = () => {
		free++;
		waiting.shift()?.();
	};
	return {
		take: () =>
			new Promise((given) => {
				const give = () => {
					free--;
					let over = false;
					given(() => {
						if (!over) {
							over = true;
							ended();
						}
					});
				};
				if (free > 0) {
					give();
				} else {
					waiting.push(give);
				}
			}),
	};
}

/**
 * loadInTurn points frame at page once loading gives it a turn, and gives the
 * turn back once the frame has loaded or once the stop it returns is called,
 * whichever comes first. A frame stopped before its turn came is never
 * pointed at the page, and hands the turn on as soon as it is given it.
 */
export function loadInTurn(
	frame: HTMLIFrameElement,
	page: string,
	loading: Turns,
): () => void {
	let stopped = false;
	let ended = () => {};
	void loading.take().then((end) => {
		ended = end;
		if (stopped) {
			end();
			return;
		}
		frame.addEventListener("load", end, { once: true });
		frame.src = page;
	});
	return () => {
		stopped = true;
		frame.removeEventListener("load", ended);
		ended();
	};
}

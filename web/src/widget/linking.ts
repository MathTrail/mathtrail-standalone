import { useRef, useState } from "preact/hooks";
import type { Linking } from "../design/links";
import type { Host } from "./bridge";

/**
 * useLinking is how a card opens a page of the site through the chat, or
 * undefined when the chat did not say it opens pages. A press asks the chat,
 * once at a time for an address, and a page it did not open is kept among the
 * refused for as long as the card is drawn, its address shown to copy under
 * note, until a later press opens it.
 */
export function useLinking(host: Host, note: string): Linking | undefined {
	const [refused, setRefused] = useState<ReadonlySet<string>>(new Set());
	const asking = useRef(new Set<string>());
	if (!host.canOpenLinks()) {
		return undefined;
	}
	return {
		refused,
		note,
		open: (href) => {
			if (asking.current.has(href)) {
				return;
			}
			asking.current.add(href);
			void host.openLink(href).then((opened) => {
				asking.current.delete(href);
				setRefused((was) => {
					const next = new Set(was);
					if (opened) {
						next.delete(href);
					} else {
						next.add(href);
					}
					return next;
				});
			});
		},
	};
}

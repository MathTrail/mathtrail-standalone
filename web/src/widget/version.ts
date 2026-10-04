import { createContext } from "preact";

// dev is what a build that was told no version calls itself.
const dev = "dev";

/**
 * buildVersion is the version of the build the widget came with — a release,
 * or how far past one the build is — or "dev" when the build did not say. The
 * widget gives the host this name in the handshake and shows it in a card's
 * header, so that what a host is running can be read off the card itself.
 */
export function buildVersion(): string {
	return import.meta.env.VITE_VERSION || dev;
}

/**
 * versionGiven says whether the build was told a version of its own, as the
 * build of a release always is. A build told none calls itself "dev", which a
 * card in a chat may show and a page of the site may not.
 */
export function versionGiven(): boolean {
	return buildVersion() !== dev;
}

/**
 * versionNumber is the number a card's header shows for a version: the
 * version without the "v" a release's tag begins with, since the word beside
 * the number already says what it is. A version that does not begin with "v"
 * and a digit — "dev", or a bare commit — is shown as it is.
 */
export function versionNumber(version: string): string {
	return version.replace(/^v(?=\d)/, "");
}

/**
 * NamesBuild says whether a card's header names the build it came with. A card
 * in a host does. A card drawn where no host runs it — on a page of the site —
 * names the release the page was built from, and nothing when the page was
 * built from none, rather than a "dev" no chat runs.
 */
export const NamesBuild = createContext(true);

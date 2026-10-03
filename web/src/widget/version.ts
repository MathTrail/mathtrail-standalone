/**
 * buildVersion is the version of the build the widget came with — a release,
 * or how far past one the build is — or "dev" when the build did not say. The
 * widget gives the host this name in the handshake and shows it in a card's
 * header, so that what a host is running can be read off the card itself.
 */
export function buildVersion(): string {
	return import.meta.env.VITE_VERSION || "dev";
}

import type { Transport } from "@modelcontextprotocol/client";
import { render } from "preact";
import { openBridge } from "./bridge";
import { WidgetApp } from "./WidgetApp";

/**
 * start draws the card into root and connects it to the host: over transport
 * when one is given, with the page that framed the widget otherwise.
 */
export async function start(
	root: Element,
	transport?: Transport,
): Promise<void> {
	const bridge = openBridge();
	render(<WidgetApp bridge={bridge} host={bridge} />, root);
	await bridge.connect(transport);
}

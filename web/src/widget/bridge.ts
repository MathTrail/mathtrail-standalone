import type { Transport } from "@modelcontextprotocol/client";
import {
	App,
	type AppEventMap,
	applyDocumentTheme,
	type McpUiHostContext,
} from "@modelcontextprotocol/ext-apps";

/** ToolResult is a tool's result as the host hands it to the widget. */
export type ToolResult = AppEventMap["toolresult"];

/**
 * Bridge is the widget's side of its conversation with the chat host: the
 * result of the tool call that drew the card, and the handshake that starts it.
 */
export type Bridge = {
	/** result is the latest result the host delivered, undefined before one. */
	result(): ToolResult | undefined;
	/** subscribe calls listener after each new result, until it is stopped. */
	subscribe(listener: () => void): () => void;
	/**
	 * connect performs the handshake with the host: over transport when one is
	 * given, with the page that framed the widget otherwise.
	 */
	connect(transport?: Transport): Promise<void>;
};

/** openBridge opens the widget's bridge to its host, ready to connect. */
export function openBridge(): Bridge {
	// autoResize is written out because the library only assumes it when no
	// options are passed at all: without it the card never tells the host its
	// height. strict makes a misuse — a call before the handshake, a one-time
	// notification handled too late — an error rather than a console warning.
	const app = new App(
		{ name: "mathtrail", version: import.meta.env.VITE_VERSION ?? "dev" },
		{},
		{ autoResize: true, strict: true },
	);
	let latest: ToolResult | undefined;
	const listeners = new Set<() => void>();

	// Both are registered before the handshake: the host sends the result once,
	// right after it, and a listener added later would never hear it. The result
	// is kept here, so a card drawn after it arrived still shows it.
	app.addEventListener("toolresult", (result) => {
		latest = result;
		for (const listener of listeners) {
			listener();
		}
	});
	app.addEventListener("hostcontextchanged", followTheme);

	return {
		result: () => latest,
		subscribe(listener) {
			listeners.add(listener);
			return () => {
				listeners.delete(listener);
			};
		},
		async connect(transport) {
			await app.connect(transport);
			followTheme(app.getHostContext());
		},
	};
}

// followTheme gives the page the host's light or dark theme, so the card reads
// on either. A change of context names the theme only when the theme changed.
function followTheme(context: McpUiHostContext | undefined): void {
	if (context?.theme !== undefined) {
		applyDocumentTheme(context.theme);
	}
}

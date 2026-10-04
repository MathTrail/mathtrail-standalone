import type { CallToolResult, Transport } from "@modelcontextprotocol/client";
import {
	App,
	type AppEventMap,
	applyDocumentTheme,
	type McpUiHostContext,
} from "@modelcontextprotocol/ext-apps";
import { buildVersion } from "./version";

/** ToolResult is a tool's result as the host hands it to the widget. */
export type ToolResult = AppEventMap["toolresult"];

/**
 * CallStage is how far the tool call that drew the card has got before its
 * result: under way, or cancelled.
 */
export type CallStage = "started" | "cancelled";

/**
 * Call is the tool call that drew the card, as far as the host has told of
 * it: the tool, by the name the host gives it, or undefined while the host
 * has named none; how far the call has got; and the language its arguments
 * name, or undefined until the host has told them whole and when they name
 * none.
 */
export type Call = {
	tool: string | undefined;
	stage: CallStage;
	language: string | undefined;
};

/**
 * Bridge is the widget's side of its conversation with the chat host: the
 * tool call that drew the card and its result, the language the host reads,
 * and the handshake that starts it.
 */
export type Bridge = {
	/** result is the latest result the host delivered, undefined before one. */
	result(): ToolResult | undefined;
	/**
	 * call is the tool call that drew the card. Of its arguments only the
	 * language they name is kept, which a card asked for a task speaks before
	 * its result: a task handed in carries its answer in them, and a host with
	 * an earlier list of the tools still draws a card for that call.
	 */
	call(): Call;
	/**
	 * locale is the language and region the host says its user reads, as a
	 * BCP 47 tag, or undefined while it has named none.
	 */
	locale(): string | undefined;
	/**
	 * subscribe calls listener after each new result, each step of the call,
	 * the language its arguments name, and each time the host may have named
	 * another locale, until it is stopped.
	 */
	subscribe(listener: () => void): () => void;
	/**
	 * connect performs the handshake with the host: over transport when one is
	 * given, with the page that framed the widget otherwise.
	 */
	connect(transport?: Transport): Promise<void>;
};

/**
 * Host is what a card asks of the chat host it is drawn in: to call one of
 * the service's tools, to put the child's words in the chat, and to tell the
 * model what happened on the card.
 */
export type Host = {
	/**
	 * callTool calls the service's tool name with args, through the host. The
	 * tool's own failure comes back as a result that says so; a call that never
	 * reached the tool, or whose answer never came back, rejects.
	 */
	callTool(
		name: string,
		args: Record<string, unknown>,
	): Promise<CallToolResult>;
	/**
	 * sendMessage puts text in the chat as the child's message, which the model
	 * answers. It rejects when the host refuses the message or it is lost.
	 */
	sendMessage(text: string): Promise<void>;
	/**
	 * tellModel gives the model a line to read with the next message in the
	 * chat, spending no turn of the conversation. Each line replaces the one
	 * before.
	 */
	tellModel(text: string): Promise<void>;
	/**
	 * canOpenLinks says whether the host told the card, as it connected, that
	 * it opens a page a card asks it to. A host that did not say so is asked
	 * nothing: a link it would not open is a button that does nothing.
	 */
	canOpenLinks(): boolean;
	/**
	 * openLink asks the host to open the page at address, and says whether it
	 * did: a host may refuse, fail, or give up before it answers, and each of
	 * those is told as a page not opened, never as an error.
	 */
	openLink(address: string): Promise<boolean>;
};

/** openBridge opens the widget's bridge to its host, ready to connect. */
export function openBridge(): Bridge & Host {
	// autoResize is written out because the library only assumes it when no
	// options are passed at all: without it the card never tells the host its
	// height. strict makes a misuse — a call before the handshake, a one-time
	// notification handled too late — an error rather than a console warning.
	const app = new App(
		{ name: "mathtrail", version: buildVersion() },
		{},
		{ autoResize: true, strict: true },
	);
	let latest: ToolResult | undefined;
	// The call is replaced, never changed in place, and only when something of
	// it changed: a card reading it redraws when it is another object.
	let call: Call = { tool: undefined, stage: "started", language: undefined };
	const listeners = new Set<() => void>();
	const notify = () => {
		for (const listener of listeners) {
			listener();
		}
	};
	const callNow = (next: Call) => {
		if (
			next.tool !== call.tool ||
			next.stage !== call.stage ||
			next.language !== call.language
		) {
			call = next;
			notify();
		}
	};
	// The tool is kept here rather than read from the context each time: the
	// handshake's answer replaces the context, and with it whatever an earlier
	// change of it had told.
	const toolNamed = (context: McpUiHostContext | undefined) => {
		const tool = context?.toolInfo?.tool.name;
		if (tool !== undefined) {
			callNow({ ...call, tool });
		}
	};

	// All are registered before the handshake: the host sends each of them once,
	// right after it, and a listener added later would never hear it. The result
	// is kept here, so a card drawn after it arrived still shows it.
	app.addEventListener("toolresult", (result) => {
		latest = result;
		notify();
	});
	// Of the arguments told whole, the language they name and nothing else: a
	// task handed in carries its answer in them. The arguments still coming are
	// not listened for, since a language in them may be cut short, and the
	// library drops what nobody listens for. A host tells the arguments before
	// the result; told after it, they would turn a card already drawn to
	// another language, so they are not taken.
	app.addEventListener("toolinput", ({ arguments: told }) => {
		const language = told?.language;
		if (
			latest === undefined &&
			typeof language === "string" &&
			language !== ""
		) {
			callNow({ ...call, language });
		}
	});
	app.addEventListener("toolcancelled", () => {
		callNow({ ...call, stage: "cancelled" });
	});
	// A change of context names only what changed; the library has merged it
	// into the context by the time this runs.
	app.addEventListener("hostcontextchanged", (changed) => {
		followTheme(changed);
		followInsets(changed);
		toolNamed(changed);
		if (changed.locale !== undefined) {
			notify();
		}
	});

	return {
		result: () => latest,
		call: () => call,
		locale: () => app.getHostContext()?.locale,
		subscribe(listener) {
			listeners.add(listener);
			return () => {
				listeners.delete(listener);
			};
		},
		async connect(transport) {
			await app.connect(transport);
			followTheme(app.getHostContext());
			followInsets(app.getHostContext());
			toolNamed(app.getHostContext());
			// The handshake brings the host's first context, its locale among it.
			notify();
		},
		callTool: (name, args) => app.callServerTool({ name, arguments: args }),
		async sendMessage(text) {
			const answer = await app.sendMessage({
				role: "user",
				content: [{ type: "text", text }],
			});
			if (answer.isError === true) {
				throw new Error("widget: the host did not take the message");
			}
		},
		canOpenLinks: () => app.getHostCapabilities()?.openLinks !== undefined,
		async openLink(address) {
			try {
				const answer = await app.openLink({ url: address });
				return answer.isError !== true;
			} catch {
				// A host that failed, or never answered, did not open the page.
				return false;
			}
		},
		async tellModel(text) {
			await app.updateModelContext({ content: [{ type: "text", text }] });
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

// followInsets gives the page the room the host keeps at each edge of the
// screen, as the custom properties --safe-area-top, -right, -bottom and -left
// the page's padding reads. A change of context names them only when they
// changed.
function followInsets(context: McpUiHostContext | undefined): void {
	const insets = context?.safeAreaInsets;
	if (insets === undefined) {
		return;
	}
	for (const side of ["top", "right", "bottom", "left"] as const) {
		document.documentElement.style.setProperty(
			`--safe-area-${side}`,
			`${insets[side]}px`,
		);
	}
}

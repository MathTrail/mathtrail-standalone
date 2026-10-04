import type { CallToolResult } from "@modelcontextprotocol/client";
import { InMemoryTransport } from "@modelcontextprotocol/client";
import type {
	McpUiHostCapabilities,
	McpUiHostContext,
} from "@modelcontextprotocol/ext-apps";
import { AppBridge } from "@modelcontextprotocol/ext-apps/app-bridge";

/**
 * openTestHost connects the library's own host side to a transport the widget
 * can be given, as a chat host would, and returns both. Nothing in between is
 * faked: what the widget receives went through the protocol. The host tells
 * the widget what it can do as capabilities says, and by default that it can
 * do nothing a widget may ask of a host.
 */
export async function openTestHost(
	context: McpUiHostContext = {},
	capabilities: McpUiHostCapabilities = {},
) {
	const [hostSide, widgetSide] = InMemoryTransport.createLinkedPair();
	const host = new AppBridge(
		null,
		{ name: "test-host", version: "1.0.0" },
		capabilities,
		{ hostContext: context },
	);
	await host.connect(hostSide);
	return { host, widgetSide };
}

/**
 * toolInfoOf is what a host tells a card of the call that drew it: the tool,
 * by the name given.
 */
export function toolInfoOf(
	name: string,
): NonNullable<McpUiHostContext["toolInfo"]> {
	return { tool: { name, inputSchema: { type: "object" } } };
}

/**
 * deliver hands the widget a tool's result the way a host does once a call it
 * drew a card for has ended: the call's arguments first, then its result.
 */
export async function deliver(
	host: AppBridge,
	structuredContent: object,
	args: Record<string, unknown> = {},
): Promise<void> {
	await host.sendToolInput({ arguments: args });
	await host.sendToolResult({
		content: [],
		structuredContent: structuredContent as Record<string, unknown>,
	});
}

/** ToolCall is a call of one of the service's tools that a widget asked for. */
export type ToolCall = { name: string; arguments: Record<string, unknown> };

/**
 * Opening is how a host answers a page the widget asks it to open: it opens
 * it, it refuses it, or it fails to answer at all.
 */
export type Opening = "open" | "refuse" | "fail";

/**
 * Asked is a kind of thing a widget asks of its host: a tool called, a message
 * put in the chat, a line given to the model, or a page opened.
 */
export type Asked = "call" | "message" | "model line" | "page";

/**
 * listenAsHost has host answer the widget as a chat host would: each tool call
 * the widget asks for is answered by tools, each message for the chat and each
 * line for the model is taken, and each page it asks to open is answered as
 * opening says. What was asked is kept, in order, for a test to read: each
 * kind on its own, and the kinds as they came, one after another.
 */
export function listenAsHost(
	host: AppBridge,
	tools: (call: ToolCall) => CallToolResult | Promise<CallToolResult>,
	{
		refuseMessages = false,
		refuseModelLines = false,
		opening = "open",
	}: {
		refuseMessages?: boolean;
		refuseModelLines?: boolean;
		opening?: Opening;
	} = {},
) {
	const heard = {
		calls: [] as ToolCall[],
		messages: [] as string[],
		modelLines: [] as string[],
		pages: [] as string[],
		order: [] as Asked[],
	};
	host.onopenlink = async ({ url }) => {
		heard.pages.push(url);
		heard.order.push("page");
		if (opening === "fail") {
			throw new Error("the host could not open the page");
		}
		return opening === "refuse" ? { isError: true } : {};
	};
	host.oncalltool = async (params) => {
		const call = { name: params.name, arguments: params.arguments ?? {} };
		heard.calls.push(call);
		heard.order.push("call");
		return tools(call);
	};
	host.onmessage = async (params) => {
		heard.messages.push(textOf(params.content));
		heard.order.push("message");
		return refuseMessages ? { isError: true } : {};
	};
	host.onupdatemodelcontext = async (params) => {
		heard.modelLines.push(textOf(params.content ?? []));
		heard.order.push("model line");
		if (refuseModelLines) {
			throw new Error("the host keeps no lines for the model");
		}
		return {};
	};
	return heard;
}

// textOf is the text of a message's blocks, one after another.
function textOf(blocks: readonly { type: string; text?: string }[]): string {
	return blocks.map((block) => block.text ?? "").join("");
}

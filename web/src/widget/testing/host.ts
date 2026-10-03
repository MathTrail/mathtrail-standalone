import type { CallToolResult } from "@modelcontextprotocol/client";
import { InMemoryTransport } from "@modelcontextprotocol/client";
import type { McpUiHostContext } from "@modelcontextprotocol/ext-apps";
import { AppBridge } from "@modelcontextprotocol/ext-apps/app-bridge";

/**
 * openTestHost connects the library's own host side to a transport the widget
 * can be given, as a chat host would, and returns both. Nothing in between is
 * faked: what the widget receives went through the protocol.
 */
export async function openTestHost(context: McpUiHostContext = {}) {
	const [hostSide, widgetSide] = InMemoryTransport.createLinkedPair();
	const host = new AppBridge(
		null,
		{ name: "test-host", version: "1.0.0" },
		{},
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
 * listenAsHost has host answer the widget as a chat host would: each tool call
 * the widget asks for is answered by tools, and each message for the chat and
 * each line for the model is taken. What was asked is kept, in order, for a
 * test to read.
 */
export function listenAsHost(
	host: AppBridge,
	tools: (call: ToolCall) => CallToolResult | Promise<CallToolResult>,
	{
		refuseMessages = false,
		refuseModelLines = false,
	}: { refuseMessages?: boolean; refuseModelLines?: boolean } = {},
) {
	const heard = {
		calls: [] as ToolCall[],
		messages: [] as string[],
		modelLines: [] as string[],
	};
	host.oncalltool = async (params) => {
		const call = { name: params.name, arguments: params.arguments ?? {} };
		heard.calls.push(call);
		return tools(call);
	};
	host.onmessage = async (params) => {
		heard.messages.push(textOf(params.content));
		return refuseMessages ? { isError: true } : {};
	};
	host.onupdatemodelcontext = async (params) => {
		heard.modelLines.push(textOf(params.content ?? []));
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

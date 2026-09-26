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
 * deliver hands the widget a tool's result the way a host does once a call it
 * drew a card for has ended: the call's arguments first, then its result.
 */
export async function deliver(
	host: AppBridge,
	structuredContent: Record<string, unknown>,
): Promise<void> {
	await host.sendToolInput({ arguments: {} });
	await host.sendToolResult({ content: [], structuredContent });
}

import type { Host } from "./bridge";
import { type EditOutcome, readEdited } from "./payload";

/**
 * sendEdit sends changes to the child's profile to the service, named as
 * edit_profile takes them, and says how it ended. A call that never reached
 * the service, or whose answer never came back, is a change not saved.
 */
export async function sendEdit(
	host: Host,
	changes: Record<string, unknown>,
): Promise<EditOutcome> {
	try {
		return readEdited(await host.callTool("edit_profile", changes));
	} catch (error: unknown) {
		console.error("widget: the change did not reach the service", error);
		return { kind: "failed" };
	}
}

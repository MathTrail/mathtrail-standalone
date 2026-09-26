import { expect, test, vi } from "vitest";
import { start } from "./start";
import { deliver, openTestHost } from "./testing/host";

test("start draws the card and connects it, so a delivered result shows", async () => {
	const { host, widgetSide } = await openTestHost();
	const root = document.createElement("div");

	await start(root, widgetSide);
	await deliver(host, { screen: "task" });

	await vi.waitFor(() =>
		expect(root.textContent).toContain('"screen": "task"'),
	);
});

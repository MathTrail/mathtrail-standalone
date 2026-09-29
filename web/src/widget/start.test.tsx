import { expect, test, vi } from "vitest";
import { start } from "./start";
import { deliver, openTestHost } from "./testing/host";
import { fence } from "./testing/lesson";

test("start draws the card and connects it, so a delivered task shows", async () => {
	const { host, widgetSide } = await openTestHost();
	const root = document.createElement("div");

	await start(root, widgetSide);
	await deliver(host, fence);

	await vi.waitFor(() =>
		expect(root.querySelector(".mt-task-text")?.textContent).toBe(
			fence.task.question,
		),
	);
});

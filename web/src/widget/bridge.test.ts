import { afterEach, describe, expect, test, vi } from "vitest";
import { openBridge } from "./bridge";
import { deliver, openTestHost } from "./testing/host";

afterEach(() => {
	document.documentElement.removeAttribute("data-theme");
	document.documentElement.style.colorScheme = "";
});

describe("the bridge", () => {
	test("hands on the result the host delivers", async () => {
		const { host, widgetSide } = await openTestHost();
		const bridge = openBridge();
		const heard = vi.fn();
		bridge.subscribe(heard);
		await bridge.connect(widgetSide);

		await deliver(host, { screen: "task" });

		await vi.waitFor(() => expect(heard).toHaveBeenCalledOnce());
		expect(bridge.result()?.structuredContent).toEqual({ screen: "task" });
	});

	test("keeps a result that arrived before anyone listened", async () => {
		const { host, widgetSide } = await openTestHost();
		const bridge = openBridge();
		await bridge.connect(widgetSide);

		await deliver(host, { screen: "progress" });

		await vi.waitFor(() =>
			expect(bridge.result()?.structuredContent).toEqual({
				screen: "progress",
			}),
		);
	});

	test("stops calling a listener that stopped listening", async () => {
		const { host, widgetSide } = await openTestHost();
		const bridge = openBridge();
		const heard = vi.fn();
		const stop = bridge.subscribe(heard);
		await bridge.connect(widgetSide);

		stop();
		await deliver(host, { screen: "task" });

		await vi.waitFor(() => expect(bridge.result()).toBeDefined());
		expect(heard).not.toHaveBeenCalled();
	});

	test("gives the page the host's theme, and follows it when it changes", async () => {
		const { host, widgetSide } = await openTestHost({ theme: "dark" });
		const bridge = openBridge();
		await bridge.connect(widgetSide);

		expect(document.documentElement.dataset.theme).toBe("dark");
		expect(document.documentElement.style.colorScheme).toBe("dark");

		await host.sendHostContextChange({ theme: "light" });

		await vi.waitFor(() =>
			expect(document.documentElement.dataset.theme).toBe("light"),
		);
	});

	test("tells the host how big the card is", async () => {
		const { host, widgetSide } = await openTestHost();
		const sizes = vi.fn();
		host.addEventListener("sizechange", sizes);
		const bridge = openBridge();

		await bridge.connect(widgetSide);

		await vi.waitFor(() => expect(sizes).toHaveBeenCalled());
	});
});

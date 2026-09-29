import { afterEach, describe, expect, test, vi } from "vitest";
import { openBridge } from "./bridge";
import { deliver, listenAsHost, openTestHost } from "./testing/host";
import { progress } from "./testing/lesson";

afterEach(() => {
	document.documentElement.removeAttribute("data-theme");
	document.documentElement.removeAttribute("style");
});

describe("the bridge", () => {
	// The card subscribes as it is drawn, before the handshake, and hears the
	// handshake itself and then the result.
	test("hands on the result the host delivers", async () => {
		const { host, widgetSide } = await openTestHost();
		const bridge = openBridge();
		const heard = vi.fn();
		bridge.subscribe(heard);
		await bridge.connect(widgetSide);

		await deliver(host, { screen: "task" });

		await vi.waitFor(() => expect(heard).toHaveBeenCalledTimes(2));
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

		stop();
		await bridge.connect(widgetSide);
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

	test("tells the card the host's locale, and each locale it names after", async () => {
		const { host, widgetSide } = await openTestHost({ locale: "ru-RU" });
		const bridge = openBridge();
		const heard = vi.fn();
		bridge.subscribe(heard);

		await bridge.connect(widgetSide);

		expect(bridge.locale()).toBe("ru-RU");
		expect(heard).toHaveBeenCalledOnce();

		await host.sendHostContextChange({ locale: "en-GB" });

		await vi.waitFor(() => expect(bridge.locale()).toBe("en-GB"));
		expect(heard).toHaveBeenCalledTimes(2);
	});

	test("leaves the card be when a change of context names no locale", async () => {
		const { host, widgetSide } = await openTestHost({ locale: "ru-RU" });
		const bridge = openBridge();
		await bridge.connect(widgetSide);
		const heard = vi.fn();
		bridge.subscribe(heard);

		await host.sendHostContextChange({ theme: "dark" });

		await vi.waitFor(() =>
			expect(document.documentElement.dataset.theme).toBe("dark"),
		);
		expect(heard).not.toHaveBeenCalled();
		expect(bridge.locale()).toBe("ru-RU");
	});

	test("tells the host how big the card is", async () => {
		const { host, widgetSide } = await openTestHost();
		const sizes = vi.fn();
		host.addEventListener("sizechange", sizes);
		const bridge = openBridge();

		await bridge.connect(widgetSide);

		await vi.waitFor(() => expect(sizes).toHaveBeenCalled());
	});

	test("gives the page the room the host keeps at the screen's edges, and follows it", async () => {
		const { host, widgetSide } = await openTestHost({
			safeAreaInsets: { top: 12, right: 0, bottom: 34, left: 4 },
		});
		const bridge = openBridge();
		await bridge.connect(widgetSide);
		const inset = (side: string) =>
			document.documentElement.style.getPropertyValue(`--safe-area-${side}`);

		expect(["top", "right", "bottom", "left"].map(inset)).toEqual([
			"12px",
			"0px",
			"34px",
			"4px",
		]);

		await host.sendHostContextChange({
			safeAreaInsets: { top: 0, right: 0, bottom: 0, left: 0 },
		});

		await vi.waitFor(() => expect(inset("bottom")).toBe("0px"));
	});
});

describe("what the card asks of the host", () => {
	test("a tool is called through the host, and its result comes back", async () => {
		const { host, widgetSide } = await openTestHost();
		const heard = listenAsHost(host, () => progress);
		const bridge = openBridge();
		await bridge.connect(widgetSide);

		const result = await bridge.callTool("read_progress", {});

		expect(heard.calls).toEqual([{ name: "read_progress", arguments: {} }]);
		expect(result.structuredContent).toEqual(progress.structuredContent);
	});

	test("a message is put in the chat as the child's", async () => {
		const { host, widgetSide } = await openTestHost();
		const heard = listenAsHost(host, () => progress);
		const bridge = openBridge();
		await bridge.connect(widgetSide);

		await bridge.sendMessage("why isn't it 6?");

		expect(heard.messages).toEqual(["why isn't it 6?"]);
	});

	test("a message the host does not take is a failure", async () => {
		const { host, widgetSide } = await openTestHost();
		listenAsHost(host, () => progress, { refuseMessages: true });
		const bridge = openBridge();
		await bridge.connect(widgetSide);

		await expect(bridge.sendMessage("why?")).rejects.toThrow(
			"the host did not take the message",
		);
	});

	test("a line reaches the model", async () => {
		const { host, widgetSide } = await openTestHost();
		const heard = listenAsHost(host, () => progress);
		const bridge = openBridge();
		await bridge.connect(widgetSide);

		await bridge.tellModel("The child answered.");

		expect(heard.modelLines).toEqual(["The child answered."]);
	});
});

import { afterEach, beforeEach, describe, expect, test, vi } from "vitest";
import { connectorURL } from "../site/brand";
import { addCopyButtons, copiedFor } from "./copy";
import { openHome } from "./testing/home";

beforeEach(() => {
	vi.useFakeTimers({ toFake: ["setTimeout", "clearTimeout"] });
	openHome("en");
});

afterEach(() => {
	document.body.innerHTML = "";
	vi.useRealTimers();
});

// copyButton is the button that copies the address, once the demo put it.
function copyButton(): HTMLButtonElement {
	const button = document.querySelector<HTMLButtonElement>(".s-copy");
	if (button === null) {
		throw new Error("the page has no button that copies the address");
	}
	return button;
}

// said is what a screen reader is told of the copy.
const said = () =>
	document.querySelector(".s-address-row [aria-live]")?.textContent;

describe("the button that copies the connector's address", () => {
	test("stands with the address in one row where the browser can copy, and the template that held it is gone", () => {
		addCopyButtons(document, { writeText: () => Promise.resolve() });

		const row = document.querySelector(".s-address-row");
		expect([...(row?.children ?? [])].map((part) => part.className)).toEqual([
			"s-address",
			"s-copy",
			"s-hidden",
		]);
		expect(row?.querySelector(".s-address")?.textContent).toBe(connectorURL);
		expect(copyButton().textContent).toBe("Copy");
		expect(document.querySelector("template[data-copy]")).toBeNull();
	});

	test("is not shown where the browser cannot copy, and the page is as it was built", () => {
		const built = document.body.innerHTML;

		addCopyButtons(document, undefined);

		expect(document.querySelector(".s-copy")).toBeNull();
		expect(document.body.innerHTML).toBe(built);
	});

	test("copies the address, and says so for a moment, to the eye and to a screen reader", async () => {
		const copied: string[] = [];
		addCopyButtons(document, {
			writeText: (text) => {
				copied.push(text);
				return Promise.resolve();
			},
		});

		copyButton().click();
		await vi.advanceTimersByTimeAsync(0);

		expect(copied).toEqual([connectorURL]);
		expect(copyButton().textContent).toBe("Copied");
		expect(said()).toBe("Copied");
		await vi.advanceTimersByTimeAsync(copiedFor);
		expect(copyButton().textContent).toBe("Copy");
		expect(said()).toBe("");
	});

	test("selects the whole address when the browser refuses to copy, to be copied by hand", async () => {
		addCopyButtons(document, {
			writeText: () => Promise.reject(new Error("not allowed")),
		});

		copyButton().click();
		await vi.advanceTimersByTimeAsync(0);

		expect(document.getSelection()?.toString()).toBe(connectorURL);
		expect(copyButton().textContent).toBe("Copy");
	});
});

test("a template that follows no address is left as it is", () => {
	document.body.innerHTML =
		'<p>Nothing to copy</p><template data-copy=""><button type="button" class="s-copy">Copy</button><span aria-live="polite"></span></template>';

	addCopyButtons(document, { writeText: () => Promise.resolve() });

	expect(document.querySelector(".s-copy")).toBeNull();
	expect(document.querySelector("template[data-copy]")).not.toBeNull();
});

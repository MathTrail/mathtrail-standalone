import { render } from "preact";
import { act } from "preact/test-utils";
import { afterEach, describe, expect, test, vi } from "vitest";
import { Icon, Mark } from "./icons";
import { drawingAlone, drawingOf, drawnAlone } from "./testing/drawing";
import { MessageHeader, NameBar, ThreadBar } from "./thread";

const root = document.createElement("div");

afterEach(() => {
	act(() => render(null, root));
});

function draw(element: preact.JSX.Element): void {
	act(() => render(element, root));
}

describe("the line at the top", () => {
	test("over a task names the child and leads to the progress", () => {
		const opened = vi.fn();
		draw(
			<ThreadBar name="Comet" action="Profile & progress" onClick={opened} />,
		);

		const bar = root.querySelector("button.mt-bar");
		expect(bar?.classList.contains("mt-bar-back")).toBe(false);
		expect(bar?.querySelector(".mt-bar-name")?.textContent).toBe("Comet");
		expect(bar?.querySelector(".mt-bar-action")?.textContent).toBe(
			"Profile & progress",
		);
		act(() => (bar as HTMLButtonElement).click());
		expect(opened).toHaveBeenCalledOnce();
	});

	test("over the progress leads back to the task", () => {
		draw(<ThreadBar variant="back" label="Back to task" onClick={() => {}} />);

		const bar = root.querySelector("button.mt-bar");
		expect(bar?.classList.contains("mt-bar-back")).toBe(true);
		expect(bar?.textContent).toBe("Back to task");
	});

	test("where there is no progress to open names the child and offers nothing to press", () => {
		draw(<NameBar name="Comet" />);

		const bar = root.querySelector(".mt-bar");
		expect(bar?.tagName.toLowerCase()).toBe("div");
		expect(bar?.textContent).toBe("Comet");
		expect(root.querySelector("button")).toBeNull();
		expect(root.querySelector(".mt-bar-action, .mt-chevron")).toBeNull();
	});
});

describe("a header", () => {
	test("under the name carries the badge on a narrow card", () => {
		draw(<MessageHeader name="MathTrail" badge="Olympiad coach · Grade 3" />);

		expect(root.querySelector(".mt-head")?.className).toBe("mt-head");
		expect(root.querySelector(".mt-head-line .mt-name")?.textContent).toBe(
			"MathTrail",
		);
		expect(root.querySelector(".mt-head-text > .mt-badge")?.textContent).toBe(
			"Olympiad coach · Grade 3",
		);
	});

	test("beside the name carries the badge on a wide card", () => {
		draw(
			<MessageHeader name="MathTrail" badge="Olympiad coach · Grade 3" wide />,
		);

		expect(root.querySelector(".mt-head")?.className).toBe(
			"mt-head mt-head-wide",
		);
		expect(root.querySelector(".mt-head-line")).toBeNull();
		expect(
			[...(root.querySelector(".mt-head-text")?.children ?? [])].map(
				(part) => part.className,
			),
		).toEqual(["mt-name", "mt-badge"]);
	});

	test.each([
		["narrow", false],
		["wide", true],
	])("closes its line with the version on a %s card", (_, wide) => {
		draw(
			<MessageHeader
				name="MathTrail"
				badge="Olympiad coach · Grade 3"
				version={{ label: "version", number: "0.2.1" }}
				wide={wide}
			/>,
		);

		const header = root.querySelector(".mt-head");
		expect(header?.classList.contains("mt-head-versioned")).toBe(true);
		const last = header?.lastElementChild;
		expect(last?.className).toBe("mt-version");
		expect(last?.getAttribute("aria-hidden")).toBe("true");
		expect(last?.hasAttribute("dir")).toBe(false);
	});

	test("draws the version as the word, then the number read left to right", () => {
		draw(
			<MessageHeader
				name="MathTrail"
				version={{ label: "версия", number: "0.2.1" }}
			/>,
		);

		const parts = [...(root.querySelector(".mt-version")?.children ?? [])];
		expect(
			parts.map((part) => [
				part.className,
				part.textContent,
				part.getAttribute("dir"),
			]),
		).toEqual([
			["mt-version-label", "версия", null],
			["mt-version-number", "0.2.1", "ltr"],
		]);
		// Read without the stylesheet — in a reader mode, a text browser — the
		// word and the number stay two words.
		expect(root.querySelector(".mt-version")?.textContent).toBe("версия 0.2.1");
	});

	test("shows no version it was not given", () => {
		draw(<MessageHeader name="MathTrail" badge="Olympiad coach · Grade 3" />);

		expect(root.querySelector(".mt-version")).toBeNull();
		expect(root.querySelector(".mt-head-versioned")).toBeNull();
	});

	test("says who speaks: MathTrail by its logo, the child by the avatar", () => {
		draw(
			<>
				<MessageHeader author="app" name="MathTrail" />
				<MessageHeader author="person" name="Comet" />
			</>,
		);

		const [app, person] = [...root.querySelectorAll(".mt-head")].map((head) =>
			head.querySelector("svg"),
		);
		expect(drawingOf(app ?? null)).toEqual(drawingAlone(<Mark size={32} />));
		expect(person?.outerHTML).toBe(
			drawnAlone(<Icon name="avatar" size={32} />),
		);
	});

	test("has no button that opens nothing", () => {
		draw(<MessageHeader name="MathTrail" badge="Olympiad coach · Grade 3" />);

		expect(root.querySelector("button")).toBeNull();
		expect(root.querySelector("[aria-haspopup]")).toBeNull();
	});
});

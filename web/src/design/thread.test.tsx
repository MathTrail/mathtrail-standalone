import { render } from "preact";
import { act } from "preact/test-utils";
import { afterEach, describe, expect, test, vi } from "vitest";
import { MessageHeader, ReplyCard, ThreadBar } from "./thread";

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

	test("says who speaks: MathTrail by its mark, the child by the avatar", () => {
		draw(
			<>
				<MessageHeader author="app" name="MathTrail" />
				<MessageHeader author="person" name="Comet" />
			</>,
		);

		const [app, person] = root.querySelectorAll(".mt-head");
		expect(app?.querySelector("svg circle")?.getAttribute("fill")).toBe(
			"var(--mark-fill)",
		);
		expect(person?.querySelector("svg circle")?.getAttribute("fill")).toBe(
			"var(--avatar-fill)",
		);
	});

	test("has no button that opens nothing", () => {
		draw(<MessageHeader name="MathTrail" badge="Olympiad coach · Grade 3" />);

		expect(root.querySelector("button")).toBeNull();
		expect(root.querySelector("[aria-haspopup]")).toBeNull();
	});
});

describe("a reply", () => {
	test("is headed compactly, with a note beside the name", () => {
		draw(
			<ReplyCard author="person" name="You" meta="Sent to the chat">
				<p class="mt-reply-lead">why isn't it 6?</p>
			</ReplyCard>,
		);

		const reply = root.querySelector("article.mt-reply");
		expect(reply?.querySelector(".mt-head")?.className).toBe(
			"mt-head mt-head-compact",
		);
		expect(reply?.querySelector(".mt-name")?.textContent).toBe("You");
		expect(reply?.querySelector(".mt-meta")?.textContent).toBe(
			"Sent to the chat",
		);
		expect(reply?.querySelector(".mt-reply-body")?.textContent).toBe(
			"why isn't it 6?",
		);
	});

	test("with nothing to note shows only the name", () => {
		draw(
			<ReplyCard author="app" name="MathTrail">
				<p>Here's how to solve it.</p>
			</ReplyCard>,
		);

		expect(root.querySelector(".mt-meta")).toBeNull();
	});
});

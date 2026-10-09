import { render } from "preact";
import { useRef } from "preact/hooks";
import { act } from "preact/test-utils";
import { afterEach, beforeEach, describe, expect, test, vi } from "vitest";
import { useWide } from "./wide";

// Observed is a stand-in for the browser's resize observer, which the test
// page does not lay out for: the test says how wide the element has become.
class Observed {
	static last: Observed | undefined;
	observing: Element[] = [];
	readonly tell: ResizeObserverCallback;
	constructor(tell: ResizeObserverCallback) {
		this.tell = tell;
		Observed.last = this;
	}
	observe(element: Element) {
		this.observing.push(element);
	}
	unobserve() {}
	disconnect() {
		this.observing = [];
	}
	resize(width: number) {
		const entry = {
			borderBoxSize: [{ inlineSize: width, blockSize: 0 }],
			contentRect: { width },
		} as unknown as ResizeObserverEntry;
		act(() => this.tell([entry], this as unknown as ResizeObserver));
	}
}

function Card() {
	const root = useRef<HTMLDivElement>(null);
	const wide = useWide(root);
	return (
		<div ref={root} style="--widget-wide: 640px" data-wide={String(wide)} />
	);
}

const root = document.createElement("div");
document.body.append(root);

beforeEach(() => {
	Observed.last = undefined;
	vi.stubGlobal("ResizeObserver", Observed);
});

afterEach(() => {
	act(() => render(null, root));
	vi.unstubAllGlobals();
	vi.restoreAllMocks();
});

const wideness = () =>
	root.querySelector("[data-wide]")?.getAttribute("data-wide");

// scrollbar is a scrollbar as wide as given down the side of the test page,
// which lays nothing out: the page's own width less the scrollbar's.
function scrollbar(width: number) {
	vi.spyOn(document.documentElement, "clientWidth", "get").mockReturnValue(
		window.innerWidth - width,
	);
}

describe("a card", () => {
	test("is narrow until it is measured", () => {
		act(() => render(<Card />, root));

		expect(wideness()).toBe("false");
	});

	test("already wide is drawn wide from the first frame", () => {
		vi.spyOn(HTMLElement.prototype, "getBoundingClientRect").mockReturnValue(
			new DOMRect(0, 0, 736, 400),
		);

		act(() => render(<Card />, root));

		expect(wideness()).toBe("true");
		vi.restoreAllMocks();
	});

	test("is wide from the width the tokens give, and narrow below it", () => {
		act(() => render(<Card />, root));

		Observed.last?.resize(639);
		expect(wideness()).toBe("false");
		Observed.last?.resize(640);
		expect(wideness()).toBe("true");
		Observed.last?.resize(736);
		expect(wideness()).toBe("true");
	});

	test("once wide, stays wide while its page's scrollbar takes some of its width", () => {
		act(() => render(<Card />, root));
		Observed.last?.resize(640);

		scrollbar(17);
		Observed.last?.resize(640 - 17);
		expect(wideness()).toBe("true");
		Observed.last?.resize(640 - 18);
		expect(wideness()).toBe("true");
		Observed.last?.resize(640 - 19);
		expect(wideness()).toBe("false");
	});

	test("once wide, goes narrow when its frame narrows", () => {
		act(() => render(<Card />, root));
		Observed.last?.resize(640);

		scrollbar(0);
		Observed.last?.resize(638);
		expect(wideness()).toBe("false");
	});

	test("narrow, is not made wide by its page's scrollbar", () => {
		act(() => render(<Card />, root));

		scrollbar(17);
		Observed.last?.resize(640 - 17);
		expect(wideness()).toBe("false");
		Observed.last?.resize(640);
		expect(wideness()).toBe("true");
	});

	test("stops being measured once it is taken down", () => {
		act(() => render(<Card />, root));
		const observer = Observed.last;
		expect(observer?.observing).toHaveLength(1);

		act(() => render(null, root));

		expect(observer?.observing).toHaveLength(0);
	});

	test("where nothing can measure it stays narrow", () => {
		vi.stubGlobal("ResizeObserver", undefined);

		act(() => render(<Card />, root));

		expect(wideness()).toBe("false");
	});

	test("with no width in its tokens stays narrow", () => {
		function Untokened() {
			const element = useRef<HTMLDivElement>(null);
			return <div ref={element} data-wide={String(useWide(element))} />;
		}
		act(() => render(<Untokened />, root));

		expect(Observed.last).toBeUndefined();
		expect(wideness()).toBe("false");
	});
});

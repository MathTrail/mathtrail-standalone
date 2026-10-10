import { Window } from "happy-dom";
import { renderToString } from "preact-render-to-string";
import { afterAll, describe, expect, test } from "vitest";
import { type Thumb, thumbs } from "./drawings";
import { TopicDrawing } from "./TopicThumbs";

const browser = new Window();

afterAll(async () => {
	await browser.happyDOM.close();
});

// drawn is the small drawing called name, as a page in a language draws it on
// a topic's card, as a document.
function drawn(name: Thumb, locale = "en") {
	return new browser.DOMParser().parseFromString(
		renderToString(
			<TopicDrawing
				drawing={{ markup: name }}
				numbers={new Intl.NumberFormat(locale)}
			/>,
		),
		"text/html",
	) as unknown as Document;
}

describe("a topic card's drawing", () => {
	test.each(thumbs)(
		"%s is drawn left to right, in a box a screen reader passes over",
		(name) => {
			const box = drawn(name).querySelector(".s-art");

			expect(box?.getAttribute("aria-hidden")).toBe("true");
			expect(box?.getAttribute("dir")).toBe("ltr");
			expect(box?.querySelector(".s-thumb")).not.toBeNull();
		},
	);

	test.each([
		["en", "25 × 4 = 100"],
		["fa", "۲۵ × ۴ = ۱۰۰"],
	])(
		"writes its numbers as the page's language does, in %s",
		(locale, want) => {
			expect(drawn("hundred-product", locale).body.textContent).toContain(want);
		},
	);
});

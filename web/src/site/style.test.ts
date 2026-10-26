import { readFileSync } from "node:fs";
import { dirname, join } from "node:path";
import { fileURLToPath } from "node:url";
import { Window } from "happy-dom";
import { afterAll, describe, expect, test } from "vitest";

// textOf is a file of the repository, by its path from the root.
const textOf = (...path: string[]) =>
	readFileSync(
		join(dirname(fileURLToPath(import.meta.url)), "..", "..", "..", ...path),
		"utf8",
	);

// screen is a phone's window under the stylesheets every page of the site
// loads.
const screen = new Window({ width: 320, height: 800 });
for (const sheet of [
	["internal", "widget", "tokens.css"],
	["web", "src", "site", "style.css"],
]) {
	const style = screen.document.createElement("style");
	style.textContent = textOf(...sheet);
	screen.document.head.append(style);
}

afterAll(async () => {
	await screen.happyDOM.close();
});

// styleIn is the style of the element markup draws, on a page in language.
const styleIn = (language: string, markup: string, selector: string) => {
	screen.document.documentElement.lang = language;
	screen.document.body.innerHTML = markup;
	const element = screen.document.querySelector(selector);
	if (element === null) {
		throw new Error(`the markup draws nothing ${selector} finds`);
	}
	return screen.getComputedStyle(element);
};

describe("the site's stylesheet", () => {
	test("breaks an address in a page's text where it has to, so that the page does not move sideways on a phone", () => {
		const code = styleIn(
			"en",
			'<main class="s-doc"><p><code>https://mcp.mathtrail.app/mcp</code></p></main>',
			"code",
		);

		expect(code.getPropertyValue("overflow-wrap")).toBe("anywhere");
	});

	// A script that joins its letters keeps their own spacing, which the
	// tightening of the Latin letters would break: the rule that says so is
	// read as written, since the window the tests draw in does not match a
	// language through :is().
	test.each(["ar", "hi"])(
		"keeps the letters of a page in %s, whose letters join, at their own spacing",
		(language) => {
			const own =
				/:root:is\(([^{]*)\)\s*:not\(\.s-nav-tag\)\s*\{\s*letter-spacing:\s*normal;/.exec(
					textOf("web", "src", "site", "style.css"),
				);

			expect(own?.[1]).toContain(`:lang(${language})`);
		},
	);
});

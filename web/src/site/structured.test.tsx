import { renderToString } from "preact-render-to-string";
import { describe, expect, test } from "vitest";
import {
	breadcrumbList,
	type StructuredData,
	StructuredDataScript,
	webSite,
} from "./structured";

describe("structured data", () => {
	test("says the site's name for the front page it is given, in schema.org's words and in one order", () => {
		expect(JSON.stringify(webSite("https://example.test/"))).toBe(
			'{"@context":"https://schema.org","@type":"WebSite","name":"MathTrail","url":"https://example.test/"}',
		);
	});

	test("numbers a path's steps from one, each at its whole address on the site", () => {
		expect(
			breadcrumbList("https://example.test", [
				{ name: "Topics", href: "/en/topics/" },
				{ name: "Logic", href: "/en/topics/#logic" },
			]).itemListElement,
		).toEqual([
			{
				"@type": "ListItem",
				position: 1,
				name: "Topics",
				item: "https://example.test/en/topics/",
			},
			{
				"@type": "ListItem",
				position: 2,
				name: "Logic",
				item: "https://example.test/en/topics/#logic",
			},
		]);
	});

	// No word of the data may end the element or open a comment, whatever a
	// translation writes.
	test("is carried in a script of JSON that no word of it can end, and reads back as it was", () => {
		const data: StructuredData = breadcrumbList("https://example.test", [
			{ name: "A </script><!-- B", href: "/en/" },
		]);
		const html = renderToString(<StructuredDataScript data={data} />);
		const json = html.match(
			/^<script type="application\/ld\+json">([^<]*)<\/script>$/,
		)?.[1];

		expect(json).toBeDefined();
		expect(JSON.parse(json ?? "")).toEqual(data);
	});
});

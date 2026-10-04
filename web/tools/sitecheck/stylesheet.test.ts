import { describe, expect, test } from "vitest";
import { stylesheetReferences } from "./stylesheet.ts";

describe("a stylesheet's references", () => {
	test("are every address it loads, quoted or not, in its order", () => {
		expect(
			stylesheetReferences(
				[
					'@import "base.css";',
					"@import 'print.css' print;",
					'@import url("theme.css");',
					'@font-face{src:url(/assets/a.woff2)format("woff2")}',
					".mark{background:url( 'mark.svg' )}",
					'.icon{mask:URL("icon.svg#tick")}',
				].join("\n"),
			),
		).toEqual([
			"base.css",
			"print.css",
			"theme.css",
			"/assets/a.woff2",
			"mark.svg",
			"icon.svg#tick",
		]);
	});

	test("leave out what a comment says", () => {
		expect(
			stylesheetReferences(
				'/* url(old.woff2) and @import "old.css"; */ body{background:url(new.png)}',
			),
		).toEqual(["new.png"]);
	});

	test("are none for a stylesheet that loads nothing", () => {
		expect(stylesheetReferences("body{color:#000}")).toEqual([]);
	});

	test("are read in time that grows with the stylesheet, however its spaces and quotes fall", () => {
		const spaces = " ".repeat(200_000);
		const unclosed = 'url("a"'.repeat(50_000);

		expect(stylesheetReferences(`a{background:url(${spaces}`)).toEqual([]);
		expect(
			stylesheetReferences(`a{background:url(${spaces}x.png${spaces})}`),
		).toEqual(["x.png"]);
		expect(stylesheetReferences(`a{b:${unclosed}}`)).toEqual([]);
	});

	test("read past a url() that never closes to the one after it", () => {
		expect(stylesheetReferences('a{b:url("a) ; c:url("c")}')).toEqual(["c"]);
	});

	test("are one for an @import whose string holds a url()", () => {
		expect(stylesheetReferences('@import "url(x.css)";')).toEqual([
			"url(x.css)",
		]);
	});
});

import type { ComponentChildren } from "preact";
import { renderToString } from "preact-render-to-string";
import { describe, expect, test } from "vitest";
import { parsePageWords } from "./pagewords";
import { openReader } from "./reader";

// reader opens words written as YAML, as the page en/sample.yaml.
function reader(source: string, locale = "en") {
	return openReader(parsePageWords(source), `${locale}/sample.yaml`, locale);
}

// html is what a text draws, as markup.
function html(children: ComponentChildren): string {
	return renderToString(<p>{children}</p>);
}

describe("a page's text", () => {
	test("is its words, with its emphasis drawn", () => {
		const { page } = reader("lead: One *idea*, **checked**, \\*no star\\*.\n");

		expect(html(page.text("lead"))).toBe(
			"<p>One <em>idea</em>, <strong>checked</strong>, *no star*.</p>",
		);
	});

	test("has its slots filled with words, numbers in the page's language and elements", () => {
		const { page } = reader(
			"lead: '{count} topics for **grade {grade}**, see {link}.'\n",
			"ru",
		);

		expect(
			html(
				page.text("lead", {
					count: 12345,
					grade: "3",
					link: <a href="/ru/topics/">темы</a>,
				}),
			),
		).toBe(
			'<p>12\u00a0345 topics for <strong>grade 3</strong>, see <a href="/ru/topics/">темы</a>.</p>',
		);
	});

	test("is escaped as text, whatever characters it holds", () => {
		const { page } = reader("lead: 'Tom & Jerry say 2 < 3'\n");

		expect(html(page.text("lead"))).toBe("<p>Tom &amp; Jerry say 2 &lt; 3</p>");
	});

	test.each([
		["a link", "lead: See [the topics](topics/).\n", '"[the topics](topics/)"'],
		["an image", "lead: '![a fence](fence.png)'\n", '"![a fence](fence.png)"'],
		["markup", "lead: A <b>bold</b> word.\n", '"<b>"'],
		["code", "lead: Run `just site`.\n", '"`just site`"'],
		["a break", 'lead: "One  \\nTwo"\n', '"  \\n"'],
	])("is refused when it holds %s", (_, source, raw) => {
		const { page } = reader(source);

		expect(() => page.text("lead")).toThrow(
			`en/sample.yaml: lead holds ${raw}, and a page's text may hold emphasis alone`,
		);
	});

	test("is refused when it writes a character as a reference", () => {
		const { page } = reader("lead: Tom &amp; Jerry.\n");

		expect(() => page.text("lead")).toThrow(
			"en/sample.yaml: lead writes a character as a reference",
		);
	});

	test("is refused when the page gives one of its slots nothing", () => {
		const { page } = reader("lead: Grade {grade}.\n");

		expect(() => page.text("lead", {})).toThrow(
			"en/sample.yaml: lead names {grade}, and the page gives it nothing",
		);
	});
});

describe("a page's plain text", () => {
	test("is its words as they are written, stars and all, with its slots filled", () => {
		const { page } = reader("drawing: |\n  * * *\n  {count} stars\n", "ru");

		expect(page.plain("drawing", { count: 1000 })).toBe(
			"* * *\n1\u00a0000 stars",
		);
	});

	test("is refused when the page gives one of its slots nothing", () => {
		const { page } = reader("title: Grade {grade}\n");

		expect(() => page.plain("title")).toThrow("title names {grade}");
	});
});

describe("a page's list", () => {
	const source = [
		"tips:",
		"  - First",
		"  - Second",
		"examples:",
		"  - question: Q1",
		"    answer: A1",
		"  - question: Q2",
		"    answer: A2",
		"hero:",
		"  lead: L",
	].join("\n");

	test("is the keys of its items, in order, whatever each item holds", () => {
		const { page } = reader(source);

		expect(page.list("tips")).toEqual(["tips.1", "tips.2"]);
		expect(page.list("examples")).toEqual(["examples.1", "examples.2"]);
		expect(
			page.list("examples").map((item) => page.plain(`${item}.answer`)),
		).toEqual(["A1", "A2"]);
	});

	test.each([
		[
			"one the file does not have",
			"steps",
			"reads the list steps, which the file does not have",
		],
		["a section", "hero", "hero is a section, and the page reads it as a list"],
	])("is refused for %s", (_, key, want) => {
		const { page } = reader(source);

		expect(() => page.list(key)).toThrow(want);
	});
});

describe("the words a page reads", () => {
	test("are refused when the file does not have them", () => {
		const { page } = reader("lead: L\n");

		expect(() => page.text("hero.lead")).toThrow(
			"en/sample.yaml: the page reads hero.lead, which the file does not have",
		);
	});

	test("are told apart from the ones it has not read, in the file's order", () => {
		const { page, unread } = reader("title: T\nlead: L\nnote: N\ndrawing: D\n");

		expect(unread()).toEqual(["title", "lead", "note", "drawing"]);
		page.text("note");
		page.plain("title");

		expect(unread()).toEqual(["lead", "drawing"]);
	});
});

describe("whether a page's words hold a part", () => {
	const source = [
		"examples:",
		"  - question: Q1",
		"    note:",
		"      title: N",
		"  - question: Q2",
	].join("\n");

	test("is told of a text and of a section of texts, and of nothing they only begin like", () => {
		const { page } = reader(source);

		expect(page.has("examples.1.question")).toBe(true);
		expect(page.has("examples.1.note")).toBe(true);
		expect(page.has("examples.2.note")).toBe(false);
		expect(page.has("examples.1.not")).toBe(false);
	});

	test("reads nothing of what it finds", () => {
		const { page, unread } = reader(source);

		page.has("examples.1.note");

		expect(unread()).toEqual([
			"examples.1.question",
			"examples.1.note.title",
			"examples.2.question",
		]);
	});
});

describe("a text a page leaves out for its data", () => {
	test("counts as read, and draws nothing", () => {
		const { page, unread } = reader("title: T\npaper: Open the PDF\n");

		page.leaveOut("paper");

		expect(unread()).toEqual(["title"]);
	});

	test("is refused when the file does not have it, as one the page would show", () => {
		const { page } = reader("title: T\n");

		expect(() => page.leaveOut("paper")).toThrow(
			"en/sample.yaml: the page reads paper, which the file does not have",
		);
	});

	test("may be a whole section, every text of which counts as read, and none it only begins like", () => {
		const { page, unread } = reader(
			"title: T\nready:\n  month: M\n  chances:\n    title: C\nreadiness: R\n",
		);

		page.leaveOut("ready");

		expect(unread()).toEqual(["title", "readiness"]);
	});
});

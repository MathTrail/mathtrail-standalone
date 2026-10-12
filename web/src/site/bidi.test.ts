import { describe, expect, test } from "vitest";
import { leftToRightSpans, type Span } from "./bidi";

// partsOf is a text in its parts: words as they are, and each stretch that
// reads left to right set apart.
function partsOf(text: string, wholes: readonly Span[] = []) {
	const parts: (string | { leftToRight: string })[] = [];
	let from = 0;
	for (const [start, end] of leftToRightSpans(text, wholes)) {
		if (start > from) {
			parts.push(text.slice(from, start));
		}
		parts.push({ leftToRight: text.slice(start, end) });
		from = end;
	}
	if (from < text.length) {
		parts.push(text.slice(from));
	}
	return parts;
}

describe("a text written right to left", () => {
	test.each([
		[
			"a sum",
			"إذن 12 ÷ 3 = 4 فواصل.",
			["إذن ", { leftToRight: "12 ÷ 3 = 4" }, " فواصل."],
		],
		["a sum with a star", "العدد 38*6", ["العدد ", { leftToRight: "38*6" }]],
		[
			"a path and an address",
			"المسار /mcp وعنوان ui://widget",
			[
				"المسار ",
				{ leftToRight: "/mcp" },
				" وعنوان ",
				{ leftToRight: "ui://widget" },
			],
		],
		[
			"a name that opens with a dot",
			"الملف .well-known هنا",
			["الملف ", { leftToRight: ".well-known" }, " هنا"],
		],
		[
			"a formula of Greek letters",
			"الصيغة P = 0.2 + 0.8·σ(θ + δ − β) تقول",
			["الصيغة ", { leftToRight: "P = 0.2 + 0.8·σ(θ + δ − β)" }, " تقول"],
		],
		[
			"a citation in its brackets",
			"كما يبيّن (Radesky et al., 2022) ذلك",
			["كما يبيّن ", { leftToRight: "(Radesky et al., 2022)" }, " ذلك"],
		],
		[
			"a name in brackets the sentence opened",
			"(بالإنجليزية MathTrail)",
			["(بالإنجليزية MathTrail)"],
		],
		[
			"a sum the sentence's full stop follows",
			"والجواب 4 + 1 = 5. ثم",
			["والجواب ", { leftToRight: "4 + 1 = 5" }, ". ثم"],
		],
		[
			"a sum of squares",
			"مسألة: (10² + 11²) ÷ 365. والحل",
			["مسألة: ", { leftToRight: "(10² + 11²) ÷ 365" }, ". والحل"],
		],
		[
			"a sum with a fraction",
			"أخذ 0.5 + ½ = 1، وبقي 0",
			["أخذ ", { leftToRight: "0.5 + ½ = 1" }, "، وبقي 0"],
		],
		["a number after its sign", "رنا −3", ["رنا ", { leftToRight: "−3" }]],
		[
			"a button that opens with its sign",
			"اضغطوا + Add الآن",
			["اضغطوا ", { leftToRight: "+ Add" }, " الآن"],
		],
		[
			"a sum with an ellipsis",
			"الجدول: 1 + 2 + … + 10 هنا",
			["الجدول: ", { leftToRight: "1 + 2 + … + 10" }, " هنا"],
		],
		[
			"a sum after a bracket the sentence opened",
			"(بالعربية x) + (y) = z هنا",
			["(بالعربية x) ", { leftToRight: "+ (y) = z" }, " هنا"],
		],
	] as const)("keeps %s left to right", (_, text, parts) => {
		expect(partsOf(text)).toEqual(parts);
	});

	test.each([
		["a word alone", "تطبيق MathTrail المجاني"],
		["a number alone", "مسافة 12 مترًا"],
		["a range, which reads in the sentence's direction", "للصفوف 3–4"],
		["the sentence's full stop", "انتهى."],
		["no Latin at all", "مسألة جديدة، من فضلك"],
	])("leaves %s as it is", (_, text) => {
		expect(partsOf(text)).toEqual([text]);
	});

	test.each([
		"",
		"(",
		")",
		".",
		"/",
		"((x",
		"x))",
		"a.b.",
		"(a (b) c",
		"12 ÷ 3 = 4. MathTrail",
		"«MathTrail»، Claude و ChatGPT.",
		"(بالإنجليزية MathTrail) و(Radesky et al., 2022)",
	])("is put back together whole from its parts: %j", (text) => {
		const back = partsOf(text)
			.map((part) => (typeof part === "string" ? part : part.leftToRight))
			.join("");
		expect(back).toBe(text);
	});

	test("takes a stretch it holds as one whole, as an emphasis on the sum's result", () => {
		const text = "مجموع أرقامه: 3 + 8 + 1 + 6 = 18";

		expect(partsOf(text, [[text.length - 2, text.length]])).toEqual([
			"مجموع أرقامه: ",
			{ leftToRight: "3 + 8 + 1 + 6 = 18" },
		]);
	});

	test("leaves out a stretch held as one that reaches past the run, as an emphasis on words", () => {
		const text = "مجموع أرقامه: 3 + 8 + 1 + 6 = 18 فقط";
		const from = text.indexOf("18");

		expect(partsOf(text, [[from, text.length]])).toEqual([
			"مجموع أرقامه: ",
			{ leftToRight: "3 + 8 + 1 + 6" },
			" = 18 فقط",
		]);
	});
});

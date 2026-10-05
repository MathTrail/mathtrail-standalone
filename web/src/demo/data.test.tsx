import { renderToString } from "preact-render-to-string";
import { afterEach, describe, expect, test } from "vitest";
import { letters } from "../widget/choices";
import { readAnswer, readHandedTask } from "../widget/payload";
import { answered, fence } from "../widget/testing/lesson";
import { type DemoData, DemoDataScript, readDemoData } from "./data";

afterEach(() => {
	document.body.innerHTML = "";
});

// example is data of the demo for the fence, with a word in it that would end
// the element it is written into, and open a comment.
function example(): DemoData {
	const handed = readHandedTask(fence);
	const told = readAnswer(answered(), fence.task.id);
	if (handed === undefined || told.kind !== "answered") {
		throw new Error("the fence is no task answered");
	}
	return {
		locale: "en",
		words: { "task.hint": "Hint </script><!-- for nobody -->" },
		handed,
		results: Object.fromEntries(
			letters.map((choice) => [choice, { ...told.result, choice }]),
		) as DemoData["results"],
	};
}

describe("the demo's data", () => {
	test("reads back from the page as it was written, a word that would end its element and all", () => {
		const data = example();
		document.body.innerHTML = renderToString(<DemoDataScript data={data} />);

		expect(document.body.querySelectorAll("script")).toHaveLength(1);
		expect(document.body.innerHTML).not.toContain("</script><!--");
		expect(readDemoData(document)).toEqual(data);
	});

	test.each([
		["no data", ""],
		[
			"data that is no JSON",
			'<script type="application/json" data-demo="">{"locale":</script>',
		],
		[
			"data of another shape",
			'<script type="application/json" data-demo="">{"locale":"en"}</script>',
		],
	])("is none on a page with %s", (_, html) => {
		document.body.innerHTML = html;

		expect(readDemoData(document)).toBeUndefined();
	});

	test("is none when an option has no result", () => {
		const { results, ...rest } = example();
		const { E: _, ...some } = results;
		document.body.innerHTML = renderToString(
			<DemoDataScript data={{ ...rest, results: some } as DemoData} />,
		);

		expect(readDemoData(document)).toBeUndefined();
	});
});

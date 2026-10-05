import { render } from "preact";
import { act } from "preact/test-utils";
import { afterEach, describe, expect, test } from "vitest";
import { Mark } from "./icons";
import { IdScope, useScopedId } from "./ids";

const page = document.createElement("div");
const live = document.createElement("div");

afterEach(() => {
	act(() => render(null, page));
	act(() => render(null, live));
});

function Probe() {
	return <span id={useScopedId()} />;
}

// idsIn are the ids every element under root holds, in document order.
function idsIn(root: Element): string[] {
	return [...root.querySelectorAll("[id]")].map((element) => element.id);
}

describe("useScopedId", () => {
	test("in no scope gives the id Preact gives, so that a card in a chat keeps its ids", () => {
		act(() => render(<Probe />, page));

		expect(idsIn(page)).toEqual(["P0-0"]);
	});

	test("in a scope puts the scope in front of it", () => {
		act(() =>
			render(
				<IdScope.Provider value="demo-">
					<Probe />
				</IdScope.Provider>,
				live,
			),
		);

		expect(idsIn(live)).toEqual(["demo-P0-0"]);
	});

	test("keeps the ids of a root drawn in a scope apart from those of a root drawn in none", () => {
		// Each root Preact draws numbers its ids from the start: a page drawn
		// in one root and a live card drawn in another would share every id
		// but for the scope.
		act(() =>
			render(
				<>
					<Mark />
					<Mark />
				</>,
				page,
			),
		);
		act(() =>
			render(
				<IdScope.Provider value="demo-">
					<Mark />
					<Mark />
				</IdScope.Provider>,
				live,
			),
		);

		const onThePage = idsIn(page);
		const onTheCard = idsIn(live);
		expect(onThePage.length).toBeGreaterThan(0);
		expect(onTheCard.length).toBe(onThePage.length);
		expect(onTheCard.filter((id) => onThePage.includes(id))).toEqual([]);
	});
});

import { describe, expect, test } from "vitest";
import { verdictOf } from "./enumeration";

describe("a cell of the order table", () => {
	test.each([
		["Yes: AB and BA are two lines", ["Yes", "AB and BA are two lines"]],
		[
			"Oui\u00a0: AB et BA sont deux files",
			["Oui", "AB et BA sont deux files"],
		],
		["是：AB 和 BA 是两种排法", ["是", "AB 和 BA 是两种排法"]],
		[
			"Yes: the leader comes first: larger",
			["Yes", "the leader comes first: larger"],
		],
		["Ja (2:1): zwei zu eins", ["Ja (2:1)", "zwei zu eins"]],
		["Yes", ["Yes", ""]],
	])(
		"%j is parted at its first colon that ends a phrase, the spaces around it left out",
		(cell, want) => {
			expect(verdictOf(cell)).toEqual(want);
		},
	);
});

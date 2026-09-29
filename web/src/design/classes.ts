/**
 * classes is the class attribute of an element: the names given, in order,
 * with those switched off — false, undefined or empty — left out.
 */
export function classes(...names: (string | false | undefined)[]): string {
	return names
		.filter((name) => typeof name === "string" && name !== "")
		.join(" ");
}

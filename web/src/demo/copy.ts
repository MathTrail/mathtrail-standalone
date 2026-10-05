/**
 * copiedFor is how long the button says it copied the address, in
 * milliseconds.
 */
export const copiedFor = 1600;

/**
 * addCopyButtons puts beside every address the page offers to copy the button
 * that copies it, from the template the page carries right after the address,
 * and only where the browser lets a page write to the clipboard: a button that
 * cannot copy is no button to show. The address and its button stand in one
 * row. Pressed, the button says for a moment that it copied, and a screen
 * reader hears so; a copy the browser refuses selects the whole address
 * instead, to be copied by hand.
 */
export function addCopyButtons(
	document: Document,
	clipboard: Pick<Clipboard, "writeText"> | undefined,
): void {
	if (clipboard === undefined) {
		return;
	}
	for (const template of document.querySelectorAll<HTMLTemplateElement>(
		"template[data-copy]",
	)) {
		const address = template.previousElementSibling;
		const parts = template.content.cloneNode(true) as DocumentFragment;
		const button = parts.querySelector<HTMLButtonElement>("button");
		const said = parts.querySelector<HTMLElement>("[aria-live]");
		if (
			address === null ||
			!address.classList.contains("s-address") ||
			button === null ||
			said === null
		) {
			continue;
		}
		const row = document.createElement("div");
		row.className = "s-address-row";
		address.before(row);
		row.append(address, parts);
		template.remove();
		const label = button.textContent ?? "";
		const done = button.dataset.done ?? label;
		let back: ReturnType<typeof setTimeout> | undefined;
		button.addEventListener("click", () => {
			clipboard.writeText(address.textContent ?? "").then(
				() => {
					button.textContent = done;
					said.textContent = done;
					clearTimeout(back);
					back = setTimeout(() => {
						button.textContent = label;
						said.textContent = "";
					}, copiedFor);
				},
				() => selectWhole(document, address),
			);
		});
	}
}

// selectWhole selects all of element's text, as a click on it does.
function selectWhole(document: Document, element: Element): void {
	const selection = document.getSelection();
	if (selection === null) {
		return;
	}
	const range = document.createRange();
	range.selectNodeContents(element);
	selection.removeAllRanges();
	selection.addRange(range);
}

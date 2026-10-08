// The home page's demo: the card on the first screen comes alive, the page's
// scroll moves the chat in its phone on a wide window, the steps of the lesson
// share one card on a wide window, and the connector's address gets a button
// that copies it. The page reads in full without it.
import { addCopyButtons } from "./copy";
import { readDemoData } from "./data";
import { bringHeroAlive } from "./hero";
import { holdTheFirstScreen } from "./phone";
import { shareTheCard } from "./stage";

// begin starts one part of the demo. A part that fails leaves its part of the
// page as it was built, and the others alive.
function begin(part: string, run: () => void): void {
	try {
		run();
	} catch (error: unknown) {
		console.error(`demo: ${part} did not start`, error);
	}
}

const data = readDemoData(document);
if (data !== undefined) {
	begin("the card on the first screen", () => bringHeroAlive(document, data));
}
// The live card is in place by now, so that what the chat holds is measured
// once it is.
begin("the phone on the first screen", () =>
	holdTheFirstScreen(document, window),
);
begin("the card of the lesson's steps", () => shareTheCard(document, window));
begin("the button that copies the address", () =>
	addCopyButtons(document, navigator.clipboard),
);

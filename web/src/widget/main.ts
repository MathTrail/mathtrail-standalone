// The design's tokens, then its components, then the page's own rules, each
// later sheet free to build on the ones before it. The tokens live beside the
// page the build writes, where the pages the service renders read them too; an
// import the build cannot find stops the build, which is what keeps a page
// without them from shipping.
import "../../../internal/widget/tokens.css";
import "../design/mathtrail.css";
import "./style.css";
import { start } from "./start";

const root = document.getElementById("app");
if (root === null) {
	throw new Error("widget: the page has no element to draw the card in");
}
try {
	await start(root);
} catch (error: unknown) {
	console.error("widget: the host did not complete the handshake", error);
}

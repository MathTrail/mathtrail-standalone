import "./style.css";
import { start } from "./start";

const root = document.getElementById("app");
if (root === null) {
	throw new Error("widget: the page has no element to draw the card in");
}
start(root).catch((error: unknown) => {
	console.error("widget: the host did not complete the handshake", error);
});

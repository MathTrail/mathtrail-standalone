import { render } from "preact";
import "./preview.css";
import { Preview } from "./Preview";

const root = document.getElementById("preview");
if (root === null) {
	throw new Error("preview: the page has no element to draw in");
}
render(<Preview />, root);

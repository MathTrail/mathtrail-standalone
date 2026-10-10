// The page "Why"'s script: the pictures of its history move on by themselves.
// The page reads in full without it, a chip of an era picking each picture.
import { playTheHistory } from "./history";

try {
	playTheHistory(document, window);
} catch (error: unknown) {
	console.error("why: the history's pictures did not start", error);
}

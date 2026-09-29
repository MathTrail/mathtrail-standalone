/**
 * stepsOf is a solution told step by step. A solution arrives as one text,
 * and each of its sentences is a step: found by the sentence rules of the
 * task's language, which know every script's full stop, and by two rules of
 * their own for the arithmetic those rules were not written for.
 */
export function stepsOf(solution: string, language: string): string[] {
	const sentencesOf = sentenceSplitter(language);
	return solution.split(/\r?\n/).flatMap((line) => {
		const pieces = sentencesOf(line).flatMap((sentence) =>
			sentence.split(fullStopBeforeNumber),
		);
		return joinContinuations(
			pieces.map((piece) => piece.trim()).filter((piece) => piece !== ""),
		);
	});
}

// A full stop, spaces and then a number, a negative one too, start a new
// sentence. The sentence rules read a full stop followed by a lowercase word
// as an abbreviation, and in "12 ÷ 3 = 4 gaps. 4 + 1 = 5 posts." they find
// that word after the number.
const fullStopBeforeNumber = /(?<=\.)\s+(?=[-−]?\p{Nd})/u;

// A piece that opens with an operation goes on with the expression before it:
// the sentence rules end a sentence at the ellipsis of "1 + 2 + ... + 12".
// They read a hyphen or a colon after an ellipsis as going on, and never cut
// there; the minus sign they do cut before, and it is an operation when a
// space follows it — against a digit it is the sign of a negative number,
// which may well start a sentence.
const operationFirst = /^(?:[+×÷=*/]|−\s)/u;

// joinContinuations joins every piece that goes on with an expression to the
// piece before it.
function joinContinuations(pieces: readonly string[]): string[] {
	const steps: string[] = [];
	for (const piece of pieces) {
		const last = steps.length - 1;
		if (last >= 0 && operationFirst.test(piece)) {
			steps[last] = `${steps[last]} ${piece}`;
		} else {
			steps.push(piece);
		}
	}
	return steps;
}

// sentenceSplitter cuts a line into sentences by the rules of language, or of
// the platform's own language when the tag cannot be read. Where the platform
// has no sentence rules at all, a line stays one step.
function sentenceSplitter(language: string): (line: string) => string[] {
	if (typeof Intl.Segmenter !== "function") {
		return (line) => [line];
	}
	const segmenter = segmenterFor(language);
	return (line) =>
		Array.from(segmenter.segment(line), ({ segment }) => segment);
}

function segmenterFor(language: string): Intl.Segmenter {
	try {
		return new Intl.Segmenter(language, { granularity: "sentence" });
	} catch {
		return new Intl.Segmenter(undefined, { granularity: "sentence" });
	}
}

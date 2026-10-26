FIRST = [40, 70]  # the scores of the first two shooters
SHOOTER = 10  # the one asked about

def solve(options):
    # Each later shooter scores the average of all the scores before theirs.
    # The average is kept a whole number, so that no rounding can hide in it.
    scores = list(FIRST)
    while len(scores) < SHOOTER:
        total = sum(scores)
        if total % len(scores) != 0:
            fail("an average is not a whole number")
        scores.append(total // len(scores))
    return match(options, scores[SHOOTER - 1])

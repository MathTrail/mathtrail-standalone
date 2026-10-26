def solve(options):
    # Every way of putting four coins straight into two purses, side by side
    # or one inside the other; a purse holds its own coins and those of a
    # purse inside it.
    larger = set()
    for first in range(5):
        second = 4 - first
        for inside in ["none", "first in second", "second in first"]:
            held_first, held_second = first, second
            if inside == "first in second":
                held_second += first
            if inside == "second in first":
                held_first += second
            small, big = min(held_first, held_second), max(held_first, held_second)
            if big == 2 * small:
                larger.add(big)
    if len(larger) != 1:
        fail("the purses give %d answers" % len(larger))
    return match(options, list(larger)[0])

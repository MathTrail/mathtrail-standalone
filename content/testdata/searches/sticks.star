# For "lay every stick end to end along the sides of a square": sticks drawn
# as lines along the lines of squared paper, each as long as the cells it
# spans. Placing each
# stick on one of four sides asks for four to the power of the sticks, past a
# million at ten, so the sticks are counted by their length instead: sticks of
# one length are alike, and a side is how many of each length it takes. Four
# sides that together take every stick make the square.

def solve(options, picture):
    counts = {}
    for x1, y1, x2, y2 in picture["lines"]:
        if (x1 != x2 and y1 != y2) or (x1 == x2 and y1 == y2):
            fail("a stick lies along a line of the paper and is at least a cell long")
        length = abs(x2 - x1) + abs(y2 - y1)
        counts[length] = counts.get(length, 0) + 1
    lengths = sorted(counts.keys())
    total = sum([length * counts[length] for length in lengths])
    if total % 4 != 0:
        return match(options, "It is impossible")
    side = total // 4
    sides = []
    for taken in product(*[range(min(counts[length], side // length) + 1) for length in lengths]):
        if sum([taken[i] * lengths[i] for i in range(len(lengths))]) == side:
            sides.append(taken)
    for square in product(sides, repeat=4):
        if all([sum([taken[i] for taken in square]) == counts[lengths[i]] for i in range(len(lengths))]):
            return match(options, side)
    return match(options, "It is impossible")

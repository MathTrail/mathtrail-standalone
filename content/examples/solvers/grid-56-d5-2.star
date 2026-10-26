def fold(painted, line, across):
    # The cells a fold along a line of the grid paints, given as a bit for
    # each cell of the 4 by 4 square: a cell takes the paint of the cell it
    # lands on, and a cell past the edge lands on nothing.
    after = painted
    for row in range(4):
        for column in range(4):
            if not painted & (1 << (4 * row + column)):
                continue
            if across:
                row_to, column_to = 2 * line - 1 - row, column
            else:
                row_to, column_to = row, 2 * line - 1 - column
            if row_to >= 0 and row_to < 4 and column_to >= 0 and column_to < 4:
                after = after | (1 << (4 * row_to + column_to))
    return after

def solve(options):
    # Every way of folding, one fold after another, from A1 painted alone;
    # the first round that reaches the whole square gives the fewest folds.
    whole = (1 << 16) - 1
    seen = set([1])
    reached = [1]
    folds = 0
    while whole not in seen:
        folds += 1
        if folds > 16:
            fail("the folds never paint the whole square")
        nxt = []
        for painted in reached:
            for line in range(1, 4):
                for across in [False, True]:
                    after = fold(painted, line, across)
                    if after not in seen:
                        seen.add(after)
                        nxt.append(after)
        reached = nxt
    return match(options, folds)

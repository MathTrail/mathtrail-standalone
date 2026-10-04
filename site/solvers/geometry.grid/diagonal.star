# A rectangle of 3 by 5 cells and the line from one corner to the opposite
# one: keep the figure as the set of cells the line passes through inside, and
# count them. A cell counts when the line crosses its inside, not when it only
# touches one of its corners. The note says a 4 by 6 rectangle gives 8 cells,
# so that is checked too.
# Written from the template figure of geometry.grid.

def crossed(rows, columns):
    # The line goes from the corner (0, 0) to (columns, rows), so at x it is
    # at height rows * x / columns. It crosses the inside of the cell between
    # columns c and c + 1 and rows r and r + 1 when it is below the cell's top
    # at the cell's left side and above the cell's bottom at its right side,
    # worked out in whole numbers.
    return set([(r, c) for r in range(rows) for c in range(columns)
                if rows * c < columns * (r + 1) and rows * (c + 1) > columns * r])

def solve(options):
    if len(crossed(4, 6)) != 8:
        fail("the line across a 4 by 6 rectangle does not pass through 8 cells")
    return match(options, len(crossed(3, 5)))

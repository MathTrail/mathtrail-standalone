# For splitting cells into two pieces joined by whole sides: try every first
# piece of the right size, and keep it when both it and the cells left over
# hold together. When the two pieces are the same size, the first cell always
# goes into the first piece, so that each split is counted once and not once
# for each piece; pieces of different sizes are told apart by their size.
# From reference task grid-56-d4-1.

ROWS = 2  # rows of cells
COLUMNS = 3  # columns of cells
PIECE = 3  # cells in the first piece

def joined(piece):
    # Walk from one cell to its neighbours by whole sides, a queue with a head
    # index; the piece holds together when the walk reaches all of it.
    seen = [piece[0]]
    head = 0
    while head < len(seen):
        row, column = seen[head]
        head += 1
        for step_row, step_column in [(-1, 0), (1, 0), (0, -1), (0, 1)]:
            neighbour = (row + step_row, column + step_column)
            if neighbour in piece and neighbour not in seen:
                seen.append(neighbour)
    return len(seen) == len(piece)

def solve(options):
    cells = [(row, column) for row in range(ROWS) for column in range(COLUMNS)]
    if PIECE < 1 or PIECE >= len(cells):
        fail("two pieces need a first piece of 1 to %d cells, not %d" % (len(cells) - 1, PIECE))
    ways = 0
    for piece in combinations(cells, PIECE):
        if PIECE * 2 == len(cells) and cells[0] not in piece:
            continue
        rest = [cell for cell in cells if cell not in piece]
        if joined(list(piece)) and joined(rest):
            ways += 1
    return match(options, ways)

# A strip of 2 by 4 cells cut along the grid lines into two pieces of 4 cells,
# each holding together: try every first piece of 4 cells, and keep it when
# both it and the cells left over hold together. The two pieces are the same
# size, so the first cell always goes into the first piece and each cut is
# counted once. The note says that without this each cut is counted twice, so
# that is checked too.
# Written from the template pieces of geometry.grid.

ROWS = 2  # rows of cells
COLUMNS = 4  # columns of cells
PIECE = 4  # cells in the first piece

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

def cuts(once):
    # The cuts into two pieces that hold together, each counted once when
    # once is True, and once for each of its pieces when it is False.
    cells = [(row, column) for row in range(ROWS) for column in range(COLUMNS)]
    ways = 0
    for piece in combinations(cells, PIECE):
        if once and cells[0] not in piece:
            continue
        rest = [cell for cell in cells if cell not in piece]
        if joined(list(piece)) and joined(rest):
            ways += 1
    return ways

def solve(options):
    ways = cuts(True)
    if cuts(False) != 2 * ways:
        fail("listing the pieces one at a time does not count each cut twice")
    return match(options, ways)

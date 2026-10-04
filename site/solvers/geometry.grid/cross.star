# A cross of 5 cells, one in the middle of a 3 by 3 grid and one on each side
# of it: keep the figure as a set of cells and count the cell sides that face
# outside it. The note counts all 20 sides and takes the 4 shared ones away
# twice, so that way is checked too.
# Written from the template figure of geometry.grid.

CELLS = ["A2", "B1", "B2", "B3", "C2"]  # the figure: rows A to C from the top, columns 1 to 3
ROWS = "ABC"  # the row letters, from the top
SIDE = 1  # the length of one side of a cell

def cell(name):
    # The letter picks the row and the number the column, both from zero.
    return (ROWS.index(name[0]), int(name[1:]) - 1)

def solve(options):
    figure = set([cell(name) for name in CELLS])
    outside = 0
    shared = 0  # each side two cells share, met once from each of them
    for row, column in figure:
        for step_row, step_column in [(-1, 0), (1, 0), (0, -1), (0, 1)]:
            if (row + step_row, column + step_column) in figure:
                shared += 1
            else:
                outside += 1
    if shared != 2 * 4 or 4 * len(figure) - shared != outside:
        fail("20 sides less the 4 shared ones twice is not the perimeter")
    return match(options, outside * SIDE)

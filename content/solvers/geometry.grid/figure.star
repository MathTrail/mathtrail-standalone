# For the perimeter or the area of a figure made of whole cells: keep the
# figure as a set of cells and count the cell sides that face outside it. A
# side two cells of the figure share is inside and does not count.
# From reference task grid-56-d3-1.

CELLS = ["A1", "B1", "C1", "C2", "C3", "B3", "A3"]  # the figure, as the question names its cells
ROWS = "ABC"  # the row letters, from the top
SIDE = 1  # the length of one side of a cell

def cell(name):
    # The letter picks the row and the number the column, both from zero.
    return (ROWS.index(name[0]), int(name[1:]) - 1)

def solve(options):
    figure = set([cell(name) for name in CELLS])
    outside = 0
    for row, column in figure:
        for step_row, step_column in [(-1, 0), (1, 0), (0, -1), (0, 1)]:
            if (row + step_row, column + step_column) not in figure:
                outside += 1
    return match(options, outside * SIDE)  # the perimeter; the area is len(figure)

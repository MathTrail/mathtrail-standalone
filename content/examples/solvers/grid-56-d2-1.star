def cell(name):
    # A1 is the top-left cell: the letter is the row, the number the column.
    return ("ABC".index(name[0]), int(name[1:]) - 1)

def solve(options):
    shape = set([cell(name) for name in ["A1", "B1", "C1", "C2"]])
    outside = 0
    for row, column in shape:
        for step_row, step_column in [(-1, 0), (1, 0), (0, -1), (0, 1)]:
            if (row + step_row, column + step_column) not in shape:
                outside += 1
    return match(options, outside)

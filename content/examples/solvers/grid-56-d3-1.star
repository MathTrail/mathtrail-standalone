def cell(name):
    # A1 is the top-left square: the letter is the row, the number the column.
    return ("ABC".index(name[0]), int(name[1:]) - 1)

def solve(options):
    playground = set([cell(name) for name in ["A1", "B1", "C1", "C2", "C3", "B3", "A3"]])
    fence = 0
    for row, column in playground:
        for step_row, step_column in [(-1, 0), (1, 0), (0, -1), (0, 1)]:
            if (row + step_row, column + step_column) not in playground:
                fence += 1
    return match(options, fence)

def cell(name):
    # A1 is the top-left cell: the letter is the row, the number the column.
    return ("ABCDE".index(name[0]), int(name[1:]) - 1)

def solve(options):
    row, column = cell("C3")
    reached = [
        (r, c)
        for r in range(5)
        for c in range(5)
        if (r, c) != (row, column) and abs(r - row) == abs(c - column)
    ]
    return match(options, len(reached))

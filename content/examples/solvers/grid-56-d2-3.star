def solve(options):
    roses = [(1, 1), (1, 2), (2, 1), (2, 2)]  # B2, B3, C2 and C3
    sides = [(-1, 0), (1, 0), (0, -1), (0, 1)]
    touching = [
        (row, column)
        for row in range(4)
        for column in range(4)
        if (row, column) not in roses and any([(row + dr, column + dc) in roses for dr, dc in sides])
    ]
    return match(options, len(touching))

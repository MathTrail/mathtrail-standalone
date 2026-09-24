def solve(options):
    corners = [(0, 0), (0, 2), (2, 0), (2, 2)]
    left = [(row, column) for row in range(3) for column in range(3) if (row, column) not in corners]
    return match(options, len(left))

def solve(options):
    side = 2  # centimetres along one side of a piece
    cake = set([(row, column) for row in range(3) for column in range(3) if (row, column) != (0, 0)])
    outside = 0
    for row, column in cake:
        for step_row, step_column in [(-1, 0), (1, 0), (0, -1), (0, 1)]:
            if (row + step_row, column + step_column) not in cake:
                outside += 1
    return match(options, outside * side)

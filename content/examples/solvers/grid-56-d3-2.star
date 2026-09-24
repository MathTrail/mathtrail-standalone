def solve(options):
    squares = [(row, column) for row in range(3) for column in range(3)]
    placements = [
        pair
        for pair in combinations(squares, 2)
        if abs(pair[0][0] - pair[1][0]) + abs(pair[0][1] - pair[1][1]) == 1
    ]
    return match(options, len(placements))

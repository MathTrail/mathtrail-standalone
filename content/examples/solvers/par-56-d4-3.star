SIZE = 3  # squares along a side of the board

def neighbours(square):
    row, column = square
    near = [(row + 1, column), (row - 1, column), (row, column + 1), (row, column - 1)]
    return [(r, c) for r, c in near if 0 <= r and r < SIZE and 0 <= c and c < SIZE]

def solve(options):
    squares = [(row, column) for row in range(SIZE) for column in range(SIZE)]
    # Every way all the bugs can crawl at once, one square each; the squares
    # no bug lands on are the empty ones.
    fewest = len(squares)
    for landed in product(*[neighbours(square) for square in squares]):
        fewest = min([fewest, len(squares) - len(set(landed))])
    return match(options, fewest)

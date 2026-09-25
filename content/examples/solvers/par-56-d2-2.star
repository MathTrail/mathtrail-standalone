SIZE = 8  # squares along a side of the chessboard
JUMPS = [(1, 2), (2, 1), (2, -1), (1, -2), (-1, -2), (-2, -1), (-2, 1), (-1, 2)]
CANDIDATES = [7, 13, 10, 9, 15]  # the numbers of moves the options name

def landings(square):
    row, column = square
    near = [(row + down, column + across) for down, across in JUMPS]
    return [(r, c) for r, c in near if 0 <= r and r < SIZE and 0 <= c and c < SIZE]

def back_after(start, reach):
    # The numbers of moves, up to the largest option, after which the knight
    # can stand on its first square again.
    here = set([start])
    counts = []
    for moves in range(1, max(CANDIDATES) + 1):
        here = set([landing for square in here for landing in reach[square]])
        if start in here:
            counts.append(moves)
    return counts

def solve(options):
    squares = [(row, column) for row in range(SIZE) for column in range(SIZE)]
    reach = {square: landings(square) for square in squares}
    # The board looks the same from each of its corners, so the squares of one
    # quarter stand for all of them.
    quarter = [(row, column) for row, column in squares if row < SIZE // 2 and column < SIZE // 2]
    returns = [back_after(square, reach) for square in quarter]
    fitting = [n for n in CANDIDATES if all([n in counts for counts in returns])]
    if len(fitting) != 1:
        fail("%d of the numbers work from every square" % len(fitting))
    if any([n in counts for counts in returns for n in CANDIDATES if n != fitting[0]]):
        fail("another number works from some square")
    return match(options, fitting[0])

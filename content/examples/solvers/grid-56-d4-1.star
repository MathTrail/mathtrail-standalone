def joined(piece):
    # Walk from one square to its neighbours by whole sides; the piece holds
    # together when the walk reaches every square of it.
    seen = [piece[0]]
    head = 0
    while head < len(seen):
        row, column = seen[head]
        head += 1
        for step_row, step_column in [(-1, 0), (1, 0), (0, -1), (0, 1)]:
            neighbour = (row + step_row, column + step_column)
            if neighbour in piece and neighbour not in seen:
                seen.append(neighbour)
    return len(seen) == len(piece)

def solve(options):
    squares = [(row, column) for row in range(2) for column in range(3)]
    ways = 0
    for piece in combinations(squares, 3):
        # The top-left square is always in the first piece, so each split is
        # counted once rather than once for each piece.
        if squares[0] not in piece:
            continue
        rest = [square for square in squares if square not in piece]
        if joined(list(piece)) and joined(rest):
            ways += 1
    return match(options, ways)

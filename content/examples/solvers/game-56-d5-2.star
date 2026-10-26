def steps(square):
    row, column = square
    near = [(row + 1, column), (row - 1, column), (row, column + 1), (row, column - 1)]
    return [(r, c) for r, c in near if r >= 0 and r < 4 and c >= 0 and c < 4]

def solve(options):
    # Every position: the square of the dragon to move and of the other one.
    # A position is won for the mover when it can step onto the other dragon,
    # or when some move leaves the other dragon a lost position; it is lost
    # when every move leaves the other a won one.
    squares = [(r, c) for r in range(4) for c in range(4)]
    won = {}
    for mover in squares:
        for other in squares:
            if mover != other and other in steps(mover):
                won[(mover, other)] = True
    changed = True
    while changed:
        changed = False
        for mover in squares:
            for other in squares:
                if mover == other or (mover, other) in won:
                    continue
                after = [won.get((other, square)) for square in steps(mover) if square != other]
                if False in after:
                    won[(mover, other)] = True
                    changed = True
                elif after and all([value == True for value in after]):
                    won[(mover, other)] = False
                    changed = True
    start = won.get(((0, 0), (3, 3)))
    if start == None:
        return match(options, "Neither: they can run round the board for ever")
    if start:
        return match(options, "The dragon from A1, since it moves first")
    return match(options, "The dragon from D4")

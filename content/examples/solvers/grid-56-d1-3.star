def turned(cell):
    # A quarter turn clockwise about the centre of a 3 by 3 board.
    row, column = cell
    return (column, 2 - row)

def solve(options):
    # Every set of painted cells with A1 among them, kept when a quarter turn
    # leaves it as it was.
    cells = [(row, column) for row in range(3) for column in range(3)]
    others = [c for c in cells if c != (0, 0)]
    fewest = 10
    for chosen in product([False, True], repeat=len(others)):
        painted = set([(0, 0)] + [others[i] for i in range(len(others)) if chosen[i]])
        if set([turned(c) for c in painted]) == painted:
            fewest = min(fewest, len(painted))
    return match(options, fewest)

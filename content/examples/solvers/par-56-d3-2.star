SIZE = 4  # squares along a side of the board
CUT = [(0, 0), (SIZE - 1, SIZE - 1)]  # two opposite corners

def solve(options):
    # Go through the squares in order, and at each free one either leave it
    # empty or lay a domino from it to the right or down. The stack holds the
    # square reached, the squares already covered, and the dominoes laid.
    squares = SIZE * SIZE
    covered = 0
    for row, column in CUT:
        covered |= 1 << (row * SIZE + column)
    most = 0
    stack = [(0, covered, 0)]
    while stack:
        place, covered, laid = stack.pop()
        if place == squares:
            most = max([most, laid])
            continue
        stack.append((place + 1, covered, laid))
        if covered & (1 << place):
            continue
        right, down = place + 1, place + SIZE
        if place % SIZE != SIZE - 1 and not covered & (1 << right):
            stack.append((place + 1, covered | (1 << place) | (1 << right), laid + 1))
        if down < squares and not covered & (1 << down):
            stack.append((place + 1, covered | (1 << place) | (1 << down), laid + 1))
    return match(options, most)

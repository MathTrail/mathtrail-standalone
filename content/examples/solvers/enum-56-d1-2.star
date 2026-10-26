CORNERS = 8

def solve(options):
    # The corners stand in a ring; two of them are joined by a side when they
    # are next to each other round it, and by a diagonal otherwise.
    diagonals = [pair for pair in combinations(range(CORNERS), 2) if (pair[1] - pair[0]) % CORNERS not in [1, CORNERS - 1]]
    return match(options, len(diagonals))

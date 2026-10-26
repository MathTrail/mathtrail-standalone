def solve(options):
    # Every cut of the strip; Ben takes the longer piece.
    best = 0
    for cut in range(1, 12):
        best = max(best, min(cut, 12 - cut))
    return match(options, best)

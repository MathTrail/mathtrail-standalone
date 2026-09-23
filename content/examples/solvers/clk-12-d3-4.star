def solve(options):
    # The eggs boil side by side in one pot, so the pot is ready when the
    # slowest egg is.
    return match(options, max([3, 3, 3]))

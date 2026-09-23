def solve(options):
    return match(options, sum(range(1, 100, 2)) - sum(range(2, 99, 2)))

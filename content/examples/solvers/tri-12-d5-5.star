def solve(options):
    return match(options, sum([n if n % 2 == 0 else -n for n in range(100, 0, -1)]))

def solve(options):
    totals = [sum(chosen) for size in range(1, 4) for chosen in combinations([1, 2, 5], size)]
    return match(options, len(set(totals)))

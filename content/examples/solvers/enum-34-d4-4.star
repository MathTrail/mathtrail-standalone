def solve(options):
    ways = 0
    for dasha in combinations(range(5), 2):
        for pasha in range(5):
            if pasha not in dasha:
                ways = ways + 1
    return match(options, ways)

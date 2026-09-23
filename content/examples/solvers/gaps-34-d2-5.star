def solve(options):
    return match(options, sum([len(str(page)) for page in range(1, 25 + 1)]))

def solve(options):
    return match(options, sum([str(n).count("1") for n in range(1, 31)]))

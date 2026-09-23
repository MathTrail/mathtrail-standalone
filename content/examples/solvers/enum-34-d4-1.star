def solve(options):
    found = [n for n in range(1, 101) if "7" in str(n)]
    return match(options, len(found))

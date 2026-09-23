def solve(options):
    found = [n for n in range(1, 31) if "3" in str(n)]
    return match(options, len(found))

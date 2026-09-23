def solve(options):
    found = [n for n in range(10, 100) if "5" in str(n)]
    return match(options, len(found))

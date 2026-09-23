def solve(options):
    coloured = list(range(3, 50 + 1, 3))
    plain = [cell for cell in range(coloured[0], coloured[-1] + 1) if cell not in coloured]
    return match(options, len(plain))

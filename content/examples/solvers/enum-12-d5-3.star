def solve(options):
    via_birch = product(["r1", "r2", "r3"], ["s1", "s2"])
    return match(options, len(via_birch) + 1)

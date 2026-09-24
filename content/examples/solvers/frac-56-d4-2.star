def solve(options):
    chocolates = set()
    for sweets in range(21, 40):  # more than 20, fewer than 40
        if sweets % 3 != 0 or sweets * 3 % 8 != 0:
            continue
        toffees = sweets // 3
        mints = sweets * 3 // 8
        chocolates.add(sweets - toffees - mints)
    if len(chocolates) == 0:
        fail("no box between 20 and 40 splits into thirds and eighths")
    if len(chocolates) > 1:
        return match(options, "It cannot be worked out")
    return match(options, list(chocolates)[0])

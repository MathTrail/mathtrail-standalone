def solve(options):
    # Every farm on which each share is a whole number of birds gives the same
    # percentage of brown hens.
    shares = set()
    for birds in range(1, 1001):
        if birds * 60 % 100 != 0:
            continue
        hens = birds * 60 // 100
        if hens * 25 % 100 != 0:
            continue
        brown = hens * 25 // 100
        if brown * 100 % birds == 0:
            shares.add(brown * 100 // birds)
    if len(shares) != 1:
        fail("the farms give %d different percentages" % len(shares))
    return match(options, "%d%%" % list(shares)[0])

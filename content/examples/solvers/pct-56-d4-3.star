def solve(options):
    adults = set()
    for passengers in range(1, 40):  # fewer than 40
        if passengers * 35 % 100 != 0:
            continue
        children = passengers * 35 // 100
        adults.add(passengers - children)
    if len(adults) != 1:
        fail("the bus can carry %d different numbers of adults" % len(adults))
    return match(options, list(adults)[0])

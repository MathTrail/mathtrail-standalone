def solve(options):
    # Every number of glasses whose quarter litres make the jug's five half
    # litres: n/4 = 5/2, that is, 2n = 20.
    glasses = [n for n in range(0, 101) if n * 2 == 5 * 4]
    if len(glasses) != 1:
        fail("%d numbers of glasses fill the jug" % len(glasses))
    return match(options, glasses[0])

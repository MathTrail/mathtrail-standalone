def solve(options):
    # Every size of bag A in which both bags split into whole beads gives the
    # same ratio in the box, written in lowest terms as the options write it.
    ratios = set()
    for bag_a in range(1, 301):
        bag_b = 2 * bag_a
        if bag_a % 3 != 0 or bag_b % 4 != 0:
            continue
        red = bag_a // 3 + bag_b // 4 * 3
        blue = bag_a // 3 * 2 + bag_b // 4
        common = gcd(red, blue)
        ratios.add("%d : %d" % (red // common, blue // common))
    if len(ratios) != 1:
        fail("the bags give %d different ratios" % len(ratios))
    return match(options, list(ratios)[0])

def solve(options):
    # Every box of whole biscuits: what Owen ate before Zoe came, and what was
    # left, shared into two equal whole shares.
    parts = set()
    for box in range(1, 301):
        for before in range(0, box + 1):
            left = box - before
            if left % 2 != 0:
                continue
            share = left // 2
            if share > 0 and before + share == 4 * share:
                common = gcd(before, box)
                parts.add("%d/%d" % (before // common, box // common))
    if len(parts) != 1:
        fail("the boxes give %d answers" % len(parts))
    return match(options, list(parts)[0])

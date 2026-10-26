def solve(options):
    # Every jar of whole units, the peaches in eighths, so that a quarter of
    # them and half of the rest are whole.
    falls = set()
    for juice in range(1, 101):
        for peaches in range(8, 401, 8):
            # Taking out a quarter of the peaches lowers the level by a fifth.
            if peaches // 4 * 5 != juice + peaches:
                continue
            now = juice + peaches - peaches // 4
            out = (peaches - peaches // 4) // 2
            common = gcd(out, now)
            falls.add("%d/%d" % (out // common, now // common))
    if len(falls) != 1:
        fail("the jars give %d answers" % len(falls))
    return match(options, list(falls)[0])

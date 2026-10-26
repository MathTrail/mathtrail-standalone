def solve(options):
    # Spoonfuls in whole fives, Mia's spoon 1 unit and Leo's twice it.
    parts = set()
    for spoonfuls in range(5, 501, 5):
        mia = spoonfuls * 1
        leo = spoonfuls * 80 // 100 * 2
        common = gcd(leo, mia + leo)
        parts.add("%d/%d" % (leo // common, (mia + leo) // common))
    if len(parts) != 1:
        fail("the spoonfuls give %d answers" % len(parts))
    return match(options, list(parts)[0])

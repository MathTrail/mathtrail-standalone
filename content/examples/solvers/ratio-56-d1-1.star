def solve(options):
    # Every class with 2 boys for every 3 girls gives the same fraction.
    fractions = set()
    for boys in range(2, 101, 2):
        girls = boys // 2 * 3
        children = boys + girls
        common = gcd(boys, children)
        fractions.add("%d/%d" % (boys // common, children // common))
    if len(fractions) != 1:
        fail("the classes give %d fractions" % len(fractions))
    return match(options, list(fractions)[0])

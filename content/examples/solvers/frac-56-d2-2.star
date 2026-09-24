def solve(options):
    crayons = 24
    left = set()
    for ben, lucy in product(range(crayons + 1), repeat=2):
        # A quarter and a third, both of the whole pack.
        if ben * 4 == crayons and lucy * 3 == crayons:
            left.add(crayons - ben - lucy)
    if len(left) != 1:
        fail("the shares leave %d answers" % len(left))
    return match(options, list(left)[0])

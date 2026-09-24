def solve(options):
    # Every number that leaves 22 when divided by 24 leaves the same remainder when divided by 3.
    remainders = set([n % 3 for n in range(0, 1001) if n % 24 == 22])
    if len(remainders) != 1:
        fail("the numbers leave %d different remainders" % len(remainders))
    return match(options, list(remainders)[0])

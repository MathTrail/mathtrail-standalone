def solve(options):
    # Whatever the two ages are now, as long as they add up to 15.
    sums = set([(a + 3) + (15 - a + 3) for a in range(1, 15)])
    if len(sums) != 1:
        fail("the sum in three years depends on the ages now")
    return match(options, list(sums)[0])

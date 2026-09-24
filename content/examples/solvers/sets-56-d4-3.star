def solve(options):
    pupils = 50
    neither = set()
    # No part of a group can be bigger than the group it is part of.
    for comics_only, magazines_only, both in product(range(33), range(26), range(26)):
        rest = pupils - comics_only - magazines_only - both
        if rest >= 0 and comics_only + both == 32 and magazines_only + both == 25 and both == 2 * rest:
            neither.add(rest)
    if len(neither) != 1:
        fail("the numbers leave %d answers" % len(neither))
    return match(options, list(neither)[0])

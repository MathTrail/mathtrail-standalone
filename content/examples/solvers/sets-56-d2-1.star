def solve(options):
    pupils = 25
    # Every way of splitting the class into football only, chess only and
    # both, with the rest in neither club; the question keeps one of them.
    neither = set()
    for football_only, chess_only, both in product(range(pupils + 1), repeat=3):
        rest = pupils - football_only - chess_only - both
        if rest >= 0 and football_only + both == 15 and chess_only + both == 12 and both == 5:
            neither.add(rest)
    if len(neither) != 1:
        fail("the numbers leave %d answers" % len(neither))
    return match(options, list(neither)[0])

def solve(options):
    children = 30
    both = set()
    for museum_only, zoo_only, went_to_both in product(range(children + 1), repeat=3):
        stayed_away = children - museum_only - zoo_only - went_to_both
        # Every child went somewhere, so nobody stayed away from both.
        if stayed_away == 0 and museum_only + went_to_both == 20 and zoo_only + went_to_both == 16:
            both.add(went_to_both)
    if len(both) != 1:
        fail("the numbers leave %d answers" % len(both))
    return match(options, list(both)[0])

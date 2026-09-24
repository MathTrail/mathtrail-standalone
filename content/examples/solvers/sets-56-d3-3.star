def solve(options):
    pupils = 40
    exactly_one = set()
    for english_only, german_only, both in product(range(pupils + 1), repeat=3):
        # Everyone learns at least one of the languages, so these three are all.
        if english_only + german_only + both == pupils and english_only + both == 28 and german_only + both == 19:
            exactly_one.add(english_only + german_only)
    if len(exactly_one) != 1:
        fail("the numbers leave %d answers" % len(exactly_one))
    return match(options, list(exactly_one)[0])

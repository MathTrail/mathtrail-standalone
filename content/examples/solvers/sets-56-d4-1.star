def solve(options):
    pupils = 32
    # Try every size of the overlaps: in all three groups, and in exactly two.
    # Those in one group only are what is left of that group, and the rest of
    # the year is in none; a split with anyone counted below zero is no split.
    none = set()
    for all_three, choir_drama, choir_art, drama_art in product(range(13), repeat=4):
        choir_only = 15 - choir_drama - choir_art - all_three
        drama_only = 13 - choir_drama - drama_art - all_three
        art_only = 12 - choir_art - drama_art - all_three
        rest = pupils - choir_only - drama_only - art_only - choir_drama - choir_art - drama_art - all_three
        if min(choir_only, drama_only, art_only, rest) < 0:
            continue
        # The pair counts in the question include those in all three.
        if all_three == 2 and choir_drama + all_three == 5 and choir_art + all_three == 4 and drama_art + all_three == 3:
            none.add(rest)
    if len(none) != 1:
        fail("the numbers leave %d answers" % len(none))
    return match(options, list(none)[0])

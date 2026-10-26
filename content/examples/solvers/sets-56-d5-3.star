def solve(options):
    # Every kind of zoo, each animal kept or not, against the three rules.
    allowed = 0
    for lions, zebras, camels in product([False, True], repeat=3):
        if lions and not zebras:
            continue
        if camels and not zebras:
            continue
        if zebras and not lions:
            continue
        allowed += 1
    return match(options, allowed)

def solve(options):
    # Every split of the class in which a quarter of each group is whole.
    pupils = 32
    seeing = set()
    for boys in range(0, pupils + 1):
        girls = pupils - boys
        if boys % 4 != 0 or girls % 4 != 0:
            continue
        # A boy sees unless he closed his left eye too, and a girl unless she
        # closed her right eye too.
        seeing.add(boys - boys // 4 + girls - girls // 4)
    if len(seeing) != 1:
        fail("the splits give %d answers" % len(seeing))
    return match(options, list(seeing)[0])

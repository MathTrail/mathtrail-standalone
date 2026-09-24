def solve(options):
    flour, sugar = 1000, 300
    # Try every amount of flour the dough could take: its sugar, 2 for every 5,
    # has to be a whole number of grams the baker has.
    most = -1
    for used in range(0, flour + 1):
        if used * 2 % 5 == 0 and used * 2 // 5 <= sugar:
            most = max(most, used)
    if most < 0:
        fail("no dough can be made")
    return match(options, flour - most)

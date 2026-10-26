AGES = range(11, 17)  # from 11 to 16, in whole years

def solve(options):
    # The first number of members that cannot all be of different ages.
    members = 1
    while len(combinations(list(AGES), members)) > 0:
        members += 1
        if members > 50:
            fail("the ages never run out")
    return match(options, members)

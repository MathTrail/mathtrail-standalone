SEATS = 8  # islanders round the table

def says(left, right):
    return not left or not right  # 'At least one of my two neighbours is a liar.'

def fits(kinds):
    for seat in range(SEATS):
        left, right = kinds[(seat - 1) % SEATS], kinds[(seat + 1) % SEATS]
        if says(left, right) != kinds[seat]:
            return False
    return True

def solve(options):
    counts = [len([kind for kind in kinds if kind]) for kinds in product([True, False], repeat=SEATS) if fits(kinds)]
    if len(counts) == 0:
        fail("no table of knights and liars can say this")
    return match(options, max(counts))  # the largest number of knights

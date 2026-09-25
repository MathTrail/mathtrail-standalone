NUMBERS = list(range(1, 11))

def has_divisor(chosen):
    # Combinations keep the order of NUMBERS, so the first of a pair is the smaller.
    return any([larger % smaller == 0 for smaller, larger in combinations(chosen, 2)])

def solve(options):
    for count in range(1, len(NUMBERS) + 1):
        if all([has_divisor(chosen) for chosen in combinations(NUMBERS, count)]):
            return match(options, count)
    fail("no number of choices is ever enough")

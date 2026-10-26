DIGITS = ["1", "3", "5", "7"]

def solve(options):
    # Every four-digit number that uses each digit once, added up.
    return match(options, sum([int("".join(order)) for order in permutations(DIGITS)]))

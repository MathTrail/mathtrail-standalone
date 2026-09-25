FACTORS = [2, 4, 8, 125, 25, 5]

def solve(options):
    # The long way: multiply the factors in the order they are written.
    return match(options, prod(FACTORS))

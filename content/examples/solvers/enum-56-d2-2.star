NUMBERS = [0, 1, 2, 3, 4]

def solve(options):
    # A tile is two numbers in no order, and a double such as 4–4 is a tile too.
    return match(options, len(combinations_with_replacement(NUMBERS, 2)))

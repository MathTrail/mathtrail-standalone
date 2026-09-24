def solve(options):
    # The dice have colours, so red 2 with blue 6 and red 6 with blue 2 are two ways.
    ways = [(red, blue) for red, blue in product(range(1, 7), range(1, 7)) if red + blue == 8]
    return match(options, len(ways))

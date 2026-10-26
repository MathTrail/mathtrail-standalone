BALLS = {"red": 4, "blue": 3, "green": 2}
BAGS = 3
EACH = 3  # balls in a bag

def solve(options):
    colours = sorted(BALLS.keys())
    # A bag is its balls with no order, and three alike bags are three such
    # bags with no order either.
    bags = combinations_with_replacement(colours, EACH)
    ways = 0
    for split in combinations_with_replacement(bags, BAGS):
        used = {colour: 0 for colour in colours}
        for bag in split:
            for colour in bag:
                used[colour] += 1
        if used == BALLS:
            ways += 1
    return match(options, ways)

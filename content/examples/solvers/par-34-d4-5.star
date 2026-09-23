def splittable(cards):
    total = sum(cards)
    for size in range(len(cards) + 1):
        for pile in combinations(cards, size):
            if 2 * sum(pile) == total:
                return True
    return False

def solve(options):
    fitting = ["cards 1 to %d" % last for last in [5, 6, 8, 9, 10] if splittable(range(1, last + 1))]
    if len(fitting) != 1:
        fail("%d of the sets can be split" % len(fitting))
    return match(options, fitting[0])

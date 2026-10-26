def solve(options):
    # Every pair of faces Jack can put on top, and every way Jill can turn
    # the dice over.
    for first, second in product(range(1, 7), repeat=2):
        turned = [(a, b) for a in [first, 7 - first] for b in [second, 7 - second]]
        if not any([(a + b) % 4 == 0 for a, b in turned]):
            return match(options, "It depends on the faces Jack puts on top")
    return match(options, "Jill, whatever Jack does")

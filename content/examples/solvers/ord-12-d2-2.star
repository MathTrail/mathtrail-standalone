def solve(options):
    found = set()
    for order in permutations(["Leo", "f1", "f2", "f3", "f4"]):
        place = {name: i + 1 for i, name in enumerate(order)}
        if place["Leo"] == 3:
            found.add(len(order) - place["Leo"])
    if len(found) != 1:
        fail("the clues fit %d different answers" % len(found))
    return match(options, list(found)[0])

def solve(options):
    found = set()
    for order in permutations(["Nina", "c1", "c2", "c3", "c4", "c5", "c6"]):
        place = {name: i + 1 for i, name in enumerate(order)}
        if place["Nina"] == 5:
            found.add(place["Nina"] - 1)
    if len(found) != 1:
        fail("the clues fit %d different answers" % len(found))
    return match(options, list(found)[0])

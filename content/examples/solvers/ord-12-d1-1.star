def solve(options):
    found = set()
    for order in permutations(["Ann", "Ben", "Kim"]):
        place = {name: i + 1 for i, name in enumerate(order)}
        if place["Ann"] < place["Ben"] and place["Ben"] < place["Kim"]:
            found.add(order[-1])
    if len(found) != 1:
        fail("the clues fit %d different answers" % len(found))
    return match(options, list(found)[0])

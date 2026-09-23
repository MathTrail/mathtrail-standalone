def solve(options):
    found = set()
    for order in permutations(["Ann", "Ben", "Kim", "Dan"]):
        place = {name: i + 1 for i, name in enumerate(order)}
        if place["Kim"] == 1 and place["Dan"] == 4 and place["Ann"] == place["Ben"] + 1:
            found.add(order[1])
    if len(found) != 1:
        fail("the clues fit %d different answers" % len(found))
    return match(options, list(found)[0])

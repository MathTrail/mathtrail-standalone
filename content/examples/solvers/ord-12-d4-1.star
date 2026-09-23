def solve(options):
    found = set()
    for order in permutations(["Ann", "Bea", "Cleo", "Dora", "Emma"]):
        place = {name: i + 1 for i, name in enumerate(order)}
        if (place["Ann"] < place["Bea"] and place["Cleo"] < place["Ann"] and
                place["Dora"] == place["Bea"] + 1 and place["Emma"] == 5):
            found.add(order[1])
    if len(found) != 1:
        fail("the clues fit %d different answers" % len(found))
    return match(options, list(found)[0])

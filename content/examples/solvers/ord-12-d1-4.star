def solve(options):
    found = set()
    for order in permutations(["Mia", "Zoe", "Lily"]):
        place = {name: i + 1 for i, name in enumerate(order)}
        if place["Mia"] < place["Zoe"] and place["Lily"] < place["Mia"]:
            found.add(order[0])
    if len(found) != 1:
        fail("the clues fit %d different answers" % len(found))
    return match(options, list(found)[0])

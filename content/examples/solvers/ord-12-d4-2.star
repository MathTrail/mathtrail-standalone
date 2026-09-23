def solve(options):
    found = set()
    for order in permutations(["Kim", "Lily", "Mia", "Nora"]):
        place = {name: i + 1 for i, name in enumerate(order)}
        if (place["Nora"] == 1 and place["Kim"] < place["Lily"] and
                place["Kim"] < place["Mia"] and place["Lily"] != 4):
            found.add(order[-1])
    if len(found) != 1:
        fail("the clues fit %d different answers" % len(found))
    return match(options, list(found)[0])

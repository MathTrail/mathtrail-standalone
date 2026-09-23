def solve(options):
    found = set()
    for order in permutations(["Oak", "Pine", "Reed", "Stone"]):
        place = {name: i + 1 for i, name in enumerate(order)}
        between = (min([place["Oak"], place["Reed"]]) < place["Pine"] and
                   place["Pine"] < max([place["Oak"], place["Reed"]]))
        if (between and (place["Stone"] == 1 or place["Stone"] == 4) and
                abs(place["Stone"] - place["Reed"]) == 1):
            found.add(order[-1] if order[0] == "Stone" else order[0])
    if len(found) != 1:
        fail("the clues fit %d different answers" % len(found))
    return match(options, list(found)[0])

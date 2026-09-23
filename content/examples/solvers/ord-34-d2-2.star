def solve(options):
    found = set()
    for order in permutations(["Dan", "Eva", "Fay", "Gil", "Hal"]):
        place = {name: i + 1 for i, name in enumerate(order)}
        if (place["Hal"] == 1 and place["Eva"] < place["Dan"] and place["Dan"] < place["Fay"] and
                place["Gil"] == place["Fay"] + 1):
            found.add(order[2])
    if len(found) != 1:
        fail("the clues fit %d different answers" % len(found))
    return match(options, list(found)[0])

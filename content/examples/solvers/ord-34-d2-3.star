def solve(options):
    found = set()
    for order in permutations(["red", "blue", "green", "white"]):
        place = {name: i + 1 for i, name in enumerate(order)}
        if (place["red"] < place["blue"] and place["blue"] < place["green"] and
                place["white"] < place["red"]):
            found.add(len(order) - place["red"])
    if len(found) != 1:
        fail("the clues fit %d different answers" % len(found))
    return match(options, list(found)[0])

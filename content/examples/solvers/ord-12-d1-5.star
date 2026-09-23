def solve(options):
    found = set()
    for order in permutations(["the red box", "the blue box", "the green box"]):
        place = {name: i + 1 for i, name in enumerate(order)}
        if place["the red box"] < place["the blue box"] and place["the blue box"] < place["the green box"]:
            found.add(order[-1])
    if len(found) != 1:
        fail("the clues fit %d different answers" % len(found))
    return match(options, list(found)[0])

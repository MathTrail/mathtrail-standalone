def solve(options):
    found = set()
    for order in permutations(["red", "blue", "green", "white", "yellow"]):
        place = {name: i + 1 for i, name in enumerate(order)}
        if (place["green"] == 2 and place["white"] == 5 and
                abs(place["white"] - place["red"]) != 1 and
                abs(place["yellow"] - place["green"]) == 1 and
                abs(place["red"] - place["blue"]) == 1):
            found.add(order[2])
    if len(found) != 1:
        fail("the clues fit %d different answers" % len(found))
    return match(options, list(found)[0])

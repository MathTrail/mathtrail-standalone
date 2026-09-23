def solve(options):
    houses = ["the red house", "the blue house", "the green house", "the yellow house", "the white house"]
    found = set()
    for order in permutations(houses):
        place = {name: i + 1 for i, name in enumerate(order)}
        if (place["the yellow house"] == 1 and
                abs(place["the white house"] - place["the yellow house"]) == 1 and
                abs(place["the red house"] - place["the white house"]) == 1 and
                abs(place["the green house"] - place["the red house"]) == 1 and
                abs(place["the green house"] - place["the blue house"]) == 1):
            found.add(order[-1])
    if len(found) != 1:
        fail("the clues fit %d different answers" % len(found))
    return match(options, list(found)[0])

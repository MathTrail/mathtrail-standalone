def solve(options):
    found = set()
    for order in permutations(["Tom", "Sam", "Leo"]):
        place = {name: i + 1 for i, name in enumerate(order)}
        if place["Tom"] < place["Sam"] and place["Sam"] < place["Leo"]:
            found.add(len(order) - place["Tom"])
    if len(found) != 1:
        fail("the clues fit %d different answers" % len(found))
    return match(options, list(found)[0])

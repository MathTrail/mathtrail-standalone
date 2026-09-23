def solve(options):
    found = set()
    for order in permutations(["Sam", "Tom", "Ula", "Vera"]):
        place = {name: i + 1 for i, name in enumerate(order)}
        if ((place["Sam"] == 1 or place["Sam"] == 4) and
                abs(place["Tom"] - place["Sam"]) == 1 and
                abs(place["Vera"] - place["Tom"]) != 1):
            found.add(order[-1] if order[0] == "Sam" else order[0])
    if len(found) != 1:
        fail("the clues fit %d different answers" % len(found))
    return match(options, list(found)[0])

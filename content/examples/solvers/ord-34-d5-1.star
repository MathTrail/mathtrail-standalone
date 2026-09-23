def ordinal(number):
    return {1: "1st", 2: "2nd", 3: "3rd"}.get(number, "%dth" % number)

def solve(options):
    found = set()
    for order in permutations(["Ann", "Ben", "Kim", "c1", "c2", "c3"]):
        place = {name: i + 1 for i, name in enumerate(order)}
        if (place["Kim"] == 6 and
                abs(place["Ben"] - place["Kim"]) - 1 == 2 and
                abs(place["Ann"] - place["Ben"]) - 1 == 1 and
                place["Ann"] != 1 and place["Ann"] != 6):
            found.add(ordinal(place["Ann"]))
    if len(found) != 1:
        fail("the clues fit %d different answers" % len(found))
    return match(options, list(found)[0])

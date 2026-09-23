def ordinal(number):
    return {1: "1st", 2: "2nd", 3: "3rd"}.get(number, "%dth" % number)

def solve(options):
    found = set()
    for order in permutations(["Ann", "Ben", "Kim", "Dan", "Eva"]):
        place = {name: i + 1 for i, name in enumerate(order)}
        if (place["Dan"] == 1 and place["Kim"] == place["Ben"] + 1 and place["Ann"] < place["Ben"] and
                abs(place["Kim"] - place["Ann"]) - 1 == 2):
            found.add(ordinal(place["Eva"]))
    if len(found) != 1:
        fail("the clues fit %d different answers" % len(found))
    return match(options, list(found)[0])

def ordinal(number):
    return {1: "1st", 2: "2nd", 3: "3rd"}.get(number, "%dth" % number)

def solve(options):
    found = set()
    for order in permutations(["Ann", "Kim", "Ben", "Dan", "Eva"]):
        place = {name: i + 1 for i, name in enumerate(order)}
        if (place["Ann"] == 1 and place["Kim"] == place["Ann"] + 1 and
                place["Ben"] == place["Kim"] + 1 and place["Dan"] == 5):
            found.add(ordinal(place["Eva"]))
    if len(found) != 1:
        fail("the clues fit %d different answers" % len(found))
    return match(options, list(found)[0])

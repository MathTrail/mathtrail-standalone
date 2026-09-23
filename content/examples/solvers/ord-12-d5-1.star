def ordinal(number):
    return {1: "1st", 2: "2nd", 3: "3rd"}.get(number, "%dth" % number)

def solve(options):
    found = set()
    for order in permutations(["Ann", "Ben", "Kim", "Dan", "Eva"]):
        place = {name: i + 1 for i, name in enumerate(order)}
        if (place["Kim"] == 5 and place["Dan"] == 4 and
                place["Ann"] != 1 and place["Ann"] != 5 and
                place["Ben"] == place["Ann"] - 1 and place["Eva"] != 1):
            found.add(ordinal(place["Eva"]))
    if len(found) != 1:
        fail("the clues fit %d different answers" % len(found))
    return match(options, list(found)[0])

def solve(options):
    found = set()
    for order in permutations(["Dan", "Eva", "Fay", "Ann", "Ben", "Kim"]):
        place = {name: i + 1 for i, name in enumerate(order)}
        if (place["Dan"] == 1 and place["Eva"] == 6 and place["Fay"] == 2 and
                place["Ann"] == place["Ben"] + 1 and place["Kim"] == place["Ann"] + 1):
            found.add(order[4])
    if len(found) != 1:
        fail("the clues fit %d different answers" % len(found))
    return match(options, list(found)[0])

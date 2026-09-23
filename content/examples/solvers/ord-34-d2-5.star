def solve(options):
    found = set()
    for order in permutations(["Kim", "Lea", "Mo"]):
        place = {name: i + 1 for i, name in enumerate(order)}
        if place["Mo"] != 2 and place["Lea"] == place["Mo"] + 1 and place["Kim"] != 1:
            found.add(order[1])
    if len(found) != 1:
        fail("the clues fit %d different answers" % len(found))
    return match(options, list(found)[0])

def solve(options):
    found = set()
    for order in permutations(["Anna", "Boris", "Vera", "Gleb"]):
        place = {name: i + 1 for i, name in enumerate(order)}
        if place["Anna"] < place["Boris"] and place["Gleb"] < place["Vera"] and place["Boris"] < place["Gleb"]:
            found.add(order[2])
    if len(found) != 1:
        fail("the clues fit %d different answers" % len(found))
    return match(options, list(found)[0])

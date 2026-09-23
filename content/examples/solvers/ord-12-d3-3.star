def solve(options):
    found = set()
    for order in permutations(["Nina", "Olga", "Pete", "Rita"]):
        place = {name: i + 1 for i, name in enumerate(order)}
        if (place["Nina"] != 1 and place["Nina"] != 4 and place["Olga"] == place["Nina"] - 1 and
                place["Pete"] > place["Nina"] and place["Rita"] == 4):
            found.add(order[0])
    if len(found) != 1:
        fail("the clues fit %d different answers" % len(found))
    return match(options, list(found)[0])

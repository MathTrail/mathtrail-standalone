def solve(options):
    found = set()
    for order in permutations(["Lev", "Mira", "Nika", "Oleg"]):
        place = {name: i + 1 for i, name in enumerate(order)}
        shorter_than_lev = len(order) - place["Lev"]
        if shorter_than_lev == 2 and place["Lev"] < place["Mira"] and place["Mira"] < place["Nika"]:
            found.add(order[0])
    if len(found) != 1:
        fail("the clues fit %d different answers" % len(found))
    return match(options, list(found)[0])

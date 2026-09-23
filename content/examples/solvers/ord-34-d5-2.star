def solve(options):
    found = set()
    for order in permutations(["Eva", "Fedor", "Gosha", "Hanna", "Ivan"]):
        place = {name: i + 1 for i, name in enumerate(order)}
        if (place["Eva"] != 1 and place["Eva"] != 5 and
                place["Fedor"] == place["Eva"] + 1 and place["Gosha"] < place["Eva"] and
                place["Ivan"] == place["Gosha"] + 1 and place["Hanna"] > place["Fedor"]):
            found.add(order[3])
    if len(found) != 1:
        fail("the clues fit %d different answers" % len(found))
    return match(options, list(found)[0])

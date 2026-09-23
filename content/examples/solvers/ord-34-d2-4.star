def solve(options):
    found = set()
    for order in permutations(["Oleg", "Petr", "Rita", "Sasha", "Tanya"]):
        place = {name: i + 1 for i, name in enumerate(order)}
        if (place["Oleg"] < place["Petr"] and place["Rita"] < place["Oleg"] and
                place["Petr"] < place["Sasha"] and place["Tanya"] < place["Rita"]):
            found.add(order[-1])
    if len(found) != 1:
        fail("the clues fit %d different answers" % len(found))
    return match(options, list(found)[0])

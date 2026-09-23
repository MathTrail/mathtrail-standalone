def solve(options):
    found = set()
    for order in permutations(["Lena", "Masha", "Nadya", "Olya"]):
        place = {name: i + 1 for i, name in enumerate(order)}
        older_than_lena = len(order) - place["Lena"]
        if place["Masha"] == 1 and older_than_lena == 1 and place["Olya"] < place["Nadya"]:
            found.add(order[-1])
    if len(found) != 1:
        fail("the clues fit %d different answers" % len(found))
    return match(options, list(found)[0])

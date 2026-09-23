def solve(options):
    chain = ["Vika", "Pavel", "Ruslan", "Semyon", "Timur", "Uliana"]
    found = set()
    for order in permutations(chain):
        place = {name: i + 1 for i, name in enumerate(order)}
        if all([place[chain[i]] < place[chain[i + 1]] for i in range(len(chain) - 1)]):
            found.add(place["Timur"] - place["Vika"] - 1)
    if len(found) != 1:
        fail("the clues fit %d different answers" % len(found))
    return match(options, list(found)[0])

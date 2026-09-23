def solve(options):
    names = ["Ann", "Ben", "Kim", "Dan", "Eva"]
    found = set()
    for order in permutations(names):
        place = {name: i + 1 for i, name in enumerate(order)}
        if all([place[names[i]] < place[names[i + 1]] for i in range(len(names) - 1)]):
            found.add(order[2])
    if len(found) != 1:
        fail("the clues fit %d different answers" % len(found))
    return match(options, list(found)[0])

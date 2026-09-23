def solve(options):
    found = set()
    for order in permutations(["the hare", "the fox", "the mouse"]):
        place = {name: i + 1 for i, name in enumerate(order)}
        if place["the hare"] < place["the fox"] and place["the fox"] < place["the mouse"]:
            found.add(order[-1])
    if len(found) != 1:
        fail("the clues fit %d different answers" % len(found))
    return match(options, list(found)[0])

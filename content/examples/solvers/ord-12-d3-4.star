def solve(options):
    found = set()
    for order in permutations(["the red book", "the blue book", "the green book", "the yellow book"]):
        place = {name: i + 1 for i, name in enumerate(order)}
        if (place["the red book"] < place["the blue book"] and
                place["the green book"] > place["the blue book"] and
                place["the yellow book"] < place["the red book"]):
            found.add(order[-1])
    if len(found) != 1:
        fail("the clues fit %d different answers" % len(found))
    return match(options, list(found)[0])

def solve(options):
    found = set()
    for order in permutations(["Lily", "Max", "Olga", "Paul", "Rita"]):
        place = {name: i + 1 for i, name in enumerate(order)}  # place 1 is the tallest
        if (place["Max"] > place["Lily"] and place["Paul"] < place["Lily"] and
                place["Rita"] < place["Olga"] and place["Rita"] > place["Max"]):
            found.add(order[-2])  # second from the shortest
    if len(found) != 1:
        fail("the clues fit %d different answers" % len(found))
    return match(options, list(found)[0])

def solve(options):
    found = set()
    for order in permutations(["Rex", "Spot", "Max", "Bo"]):
        place = {name: i + 1 for i, name in enumerate(order)}
        if place["Rex"] < place["Spot"] and place["Spot"] < place["Max"] and place["Max"] < place["Bo"]:
            found.add(order[1])
    if len(found) != 1:
        fail("the clues fit %d different answers" % len(found))
    return match(options, list(found)[0])

def solve(options):
    found = set()
    for order in permutations(["Leo", "Max", "Nick", "Oleg"]):
        place = {name: i + 1 for i, name in enumerate(order)}
        if place["Leo"] < place["Max"] and place["Oleg"] < place["Nick"] and place["Max"] < place["Oleg"]:
            found.add(order[-1])
    if len(found) != 1:
        fail("the clues fit %d different answers" % len(found))
    return match(options, list(found)[0])

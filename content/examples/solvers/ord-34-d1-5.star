def solve(options):
    found = set()
    for order in permutations(["Tom", "Sam", "Leo", "Max"]):
        place = {name: i + 1 for i, name in enumerate(order)}
        if place["Tom"] < place["Sam"] and place["Sam"] < place["Leo"] and place["Max"] < place["Tom"]:
            found.add(order[-1])
    if len(found) != 1:
        fail("the clues fit %d different answers" % len(found))
    return match(options, list(found)[0])

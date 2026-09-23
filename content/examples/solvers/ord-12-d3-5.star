def solve(options):
    found = set()
    for order in permutations(["Leo", "Max", "Sam", "Tom"]):
        place = {name: i + 1 for i, name in enumerate(order)}
        if place["Leo"] < place["Max"] and place["Sam"] < place["Leo"] and place["Tom"] < place["Sam"]:
            found.add(place["Max"] - 1)
    if len(found) != 1:
        fail("the clues fit %d different answers" % len(found))
    return match(options, list(found)[0])

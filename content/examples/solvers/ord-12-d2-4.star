def solve(options):
    found = set()
    for order in permutations(["Eva", "Tom", "Max", "Leo"]):
        place = {name: i + 1 for i, name in enumerate(order)}
        if place["Eva"] < place["Tom"] and place["Tom"] < place["Max"] and place["Max"] < place["Leo"]:
            found.add(order[2])
    if len(found) != 1:
        fail("the clues fit %d different answers" % len(found))
    return match(options, list(found)[0])

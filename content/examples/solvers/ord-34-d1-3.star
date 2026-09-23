def solve(options):
    boxes = ["red", "green", "blue", "box4", "box5"]
    found = set()
    for masses in permutations([1, 2, 3, 4, 5]):
        mass = dict(zip(boxes, masses))
        if mass["green"] == 3 and mass["green"] < mass["red"] and mass["red"] < mass["blue"]:
            found.add(mass["red"])
    if len(found) != 1:
        fail("the clues fit %d different answers" % len(found))
    return match(options, list(found)[0])

def ordinal(number):
    return {1: "1st", 2: "2nd", 3: "3rd"}.get(number, "%dth" % number)

def solve(options):
    found = set()
    # Places from the front of a queue of seven for Nina, Owen and Pablo.
    for nina, owen, pablo in permutations(range(1, 8), 3):
        if nina == 4 and abs(owen - nina) - 1 == 2 and pablo == owen - 1:
            found.add(pablo)
    if len(found) != 1:
        fail("the clues fit %d different answers" % len(found))
    return match(options, ordinal(list(found)[0]))

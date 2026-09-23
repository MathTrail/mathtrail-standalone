def ordinal(number):
    return {1: "1st", 2: "2nd", 3: "3rd"}.get(number, "%dth" % number)

def solve(options):
    found = set()
    for place in range(1, 30):
        if place + 3 == 7:
            found.add(ordinal(place))
    if len(found) != 1:
        fail("the clues fit %d different answers" % len(found))
    return match(options, list(found)[0])

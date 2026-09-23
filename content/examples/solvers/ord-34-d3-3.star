def ordinal(number):
    return {1: "1st", 2: "2nd", 3: "3rd"}.get(number, "%dth" % number)

def solve(options):
    found = set()
    for blue in range(1, 20):
        red = blue - 2
        green = blue + 3
        if green == 6 and red >= 1:
            found.add(ordinal(red))
    if len(found) != 1:
        fail("the clues fit %d different answers" % len(found))
    return match(options, list(found)[0])

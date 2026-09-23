def ordinal(number):
    return {1: "1st", 2: "2nd", 3: "3rd"}.get(number, "%dth" % number)

def solve(options):
    found = set()
    for front in range(7):
        for behind in range(7):
            if front + behind == 6 and front == 2 * behind:
                found.add(ordinal(front + 1))
    if len(found) != 1:
        fail("the clues fit %d different answers" % len(found))
    return match(options, list(found)[0])

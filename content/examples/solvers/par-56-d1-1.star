CHILDREN = 9  # each with a red balloon and a yellow one

def solve(options):
    # Every way to share the balloons out two to a child: so many children
    # with two red ones, so many with two yellow ones, and the rest with one
    # of each.
    best = None
    for red in range(CHILDREN + 1):
        for yellow in range(CHILDREN + 1 - red):
            mixed = CHILDREN - red - yellow
            if 2 * red + mixed == CHILDREN and 2 * yellow + mixed == CHILDREN:
                if best == None or red + yellow > best:
                    best = red + yellow
    if best == None:
        fail("the balloons cannot be shared out")
    return match(options, best)

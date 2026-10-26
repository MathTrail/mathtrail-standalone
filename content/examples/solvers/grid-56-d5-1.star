def solve(options):
    # Every rectangle up to 100 cells a side, its width the shorter side, kept
    # when it is not a square and its cells are as many as the sides round it.
    found = []
    for width in range(1, 101):
        for length in range(width + 1, 101):
            if width * length == 2 * (width + length):
                found.append("%d and %d" % (width, length))
    if len(found) != 1:
        fail("%d rectangles fit" % len(found))
    return match(options, found[0])

SUM = 99  # XY + YX
DIFFERENCE = 45  # XY - YX

def solve(options):
    found = []
    for x in range(1, 10):  # XY has two digits, so X is not 0
        for y in range(10):
            xy, yx = 10 * x + y, 10 * y + x
            if xy + yx == SUM and xy - yx == DIFFERENCE:
                found.append(xy)
    if len(found) != 1:
        fail("%d numbers fit" % len(found))
    return match(options, found[0])

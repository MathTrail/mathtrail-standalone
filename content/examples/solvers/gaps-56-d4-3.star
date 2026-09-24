LAMPS = 30  # round the whole edge
SHORT = 6  # on a short side, counting its corners

def lamps_round(short, long):
    # A lamp at every place on the edge of the rectangle; a corner is one place on two sides.
    places = set()
    for x in range(short):
        for y in range(long):
            if x == 0 or x == short - 1 or y == 0 or y == long - 1:
                places.add((x, y))
    return len(places)

def solve(options):
    found = [long for long in range(2, LAMPS + 1) if lamps_round(SHORT, long) == LAMPS]
    if len(found) != 1:
        fail("%d lengths fit" % len(found))
    return match(options, found[0])

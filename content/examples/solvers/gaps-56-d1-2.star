TREES = 22  # on both sides together
APART = 5  # metres between neighbours on one side

def trees_along(length):
    # A tree every 5 metres from one end of a side to the other, both ends included.
    return len([place for place in range(length + 1) if place % APART == 0])

def solve(options):
    # A tree stands at each end of both rows, so the road is a whole number of gaps long.
    found = [length for length in range(1, 201) if length % APART == 0 and 2 * trees_along(length) == TREES]
    if len(found) != 1:
        fail("%d lengths fit" % len(found))
    return match(options, "%d m" % found[0])

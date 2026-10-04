# A spider crawls along the edges of a wire cube, one edge a minute: make every
# corner it can be at after exactly 7 minutes, and see whether the corner it
# started from is among them. A corner is three coordinates, each 0 or 1, and
# an edge changes one of them. The steps colour the corners so that every edge
# joins two colours, and the note says the spider can be back after 8 minutes;
# both are checked too.
# Written from the template reachable of parity.alternation.

MINUTES = 7
START = (0, 0, 0)

def neighbours(corner):
    return [tuple([1 - c if i == axis else c for i, c in enumerate(corner)]) for axis in range(3)]

def after(minutes):
    corners = set([START])
    for _ in range(minutes):
        corners = set([n for corner in corners for n in neighbours(corner)])
    return corners

def colour(corner):
    return sum(corner) % 2

def solve(options):
    for corner in product([0, 1], repeat=3):
        for n in neighbours(corner):
            if colour(n) == colour(corner):
                fail("an edge joins two corners of one colour")
    if START not in after(MINUTES + 1):
        fail("the spider cannot be back after %d minutes, and the note says it can" % (MINUTES + 1))
    return match(options, "Yes" if START in after(MINUTES) else "No")

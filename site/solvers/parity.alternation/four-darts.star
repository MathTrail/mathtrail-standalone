# Ben throws 4 darts at rings worth 1, 3, 5 and 7 points, and every dart hits:
# make every total 4 darts can score, and see whether 15 is among them. The
# note says 3 darts can score 15, so that is checked too.
# Written from the template reachable of parity.alternation.

RINGS = [1, 3, 5, 7]  # the points of the rings
DARTS = 4
SAID = 15  # the score Ben names

def totals(darts):
    return set([sum(hits) for hits in product(RINGS, repeat=darts)])

def solve(options):
    if SAID not in totals(3):
        fail("3 darts cannot score %d, and the note says they can" % SAID)
    return match(options, "Yes" if SAID in totals(DARTS) else "No")

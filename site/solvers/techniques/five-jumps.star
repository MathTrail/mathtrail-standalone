# A grasshopper jumps one step left or right each time: make every point the
# jumps can end on, and see whether the start is among them.
# Written from the template reachable of parity.alternation.

JUMPS = 5  # jumps the grasshopper makes, starting at 0

def solve(options):
    ends = set([sum(jumps) for jumps in product([1, -1], repeat=JUMPS)])
    # The steps say why: after an odd number of jumps every end is odd.
    if [end for end in ends if end % 2 == 0]:
        fail("an even point can be reached in %d jumps" % JUMPS)
    return match(options, "Yes" if 0 in ends else "No")

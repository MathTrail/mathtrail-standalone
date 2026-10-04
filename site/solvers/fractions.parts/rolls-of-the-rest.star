# A baker has 60 rolls, sells half of them by noon and two thirds of the rest
# by evening: take the shares in turn, each of what is left at that moment,
# and see what remains.
# Written from the template what-is-left of fractions.parts.

WHOLE = 60  # the rolls the baker has at first
STEPS = [(1, 2, "rest"), (2, 3, "rest")]  # each share: top, bottom, and "whole" or "rest" for what it is taken of

def after(whole):
    # What remains of this whole after every share, or None when a share is
    # not a whole number.
    left = whole
    for top, bottom, of in STEPS:
        base = whole if of == "whole" else left
        if base * top % bottom != 0:
            return None
        left -= base * top // bottom
        if left < 0:
            fail("the shares give away more than the whole")
    return left

def solve(options):
    left = after(WHOLE)
    if left == None:
        fail("a share of %d rolls is not a whole number" % WHOLE)
    return match(options, left)

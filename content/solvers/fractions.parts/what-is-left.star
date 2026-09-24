# For a whole that gives away shares one after another: try every whole, take
# the shares in turn, each of the whole or of what is left at that moment,
# and keep the wholes in which every share is a whole number and what remains
# is what the question says.
# From reference task frac-56-d3-1.

STEPS = [(1, 2, "rest"), (1, 4, "rest")]  # each share: top, bottom, and "whole" or "rest" for what it is taken of
LEFT = 30  # what remains after the last share
LARGEST = 1000  # the largest whole worth trying

def after(whole):
    # What remains of this whole after every share, or None when a share is
    # not a whole number. Shares that give away more than there is stop the
    # program: it would prove the answer of some other task.
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
    # A share of anything but the whole or the rest is a slip in filling in
    # STEPS, and it would quietly count as a share of the rest.
    for _, _, of in STEPS:
        if of not in ("whole", "rest"):
            fail("a share is taken of the \"whole\" or of the \"rest\", not of %r" % of)
    # When the question gives the whole and asks what remains, match
    # after(that whole) instead.
    wholes = [whole for whole in range(1, LARGEST + 1) if after(whole) == LEFT]
    if len(wholes) != 1:
        fail("%d wholes leave %d, and the question has one" % (len(wholes), LEFT))
    return match(options, wholes[0])

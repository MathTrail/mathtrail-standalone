# Tom reads 2/7 of a book on one day and 3/7 of it on the next, and 40 pages
# are left: try every number of pages, take both shares of the whole book, and
# keep the books in which every share is a whole number of pages and 40 are
# left.
# Written from the template what-is-left of fractions.parts.

STEPS = [(2, 7, "whole"), (3, 7, "whole")]  # each share: top, bottom, and "whole" or "rest" for what it is taken of
LEFT = 40  # the pages left after both days
LARGEST = 1000  # the longest book worth trying

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
    wholes = [whole for whole in range(1, LARGEST + 1) if after(whole) == LEFT]
    if len(wholes) != 1:
        fail("%d books leave %d pages, and the question has one" % (len(wholes), LEFT))
    return match(options, wholes[0])

# For "at least how many in the fullest box": try every way to share the items
# out among the boxes, and the answer is the fullest box of the most even way.
# That is ITEMS + 1 to the power BOXES ways to try, so keep the numbers small.
# From reference task pig-12-d3-3.

ITEMS = 9  # the children
BOXES = 4  # the benches they sit on

def solve(options):
    fullest = ITEMS
    for counts in product(range(ITEMS + 1), repeat=BOXES):
        if sum(counts) == ITEMS:
            fullest = min([fullest, max(counts)])
    return match(options, fullest)

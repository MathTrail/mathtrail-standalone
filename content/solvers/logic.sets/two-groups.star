# For two groups that may overlap: try every way of splitting everyone into
# the first group only, the second only, both and neither, keep the splits
# that agree with everything the question says, and read the answer off them.
# From reference task sets-56-d3-1.

TOTAL = 20  # everyone in the question
FIRST = 12  # in the first group, those in both included
SECOND = 10  # in the second group, those in both included

def fits(first_only, second_only, both, neither):
    # Everything else the question says goes here too, such as "everyone is
    # in at least one" (neither == 0) or "twice as many in both as in neither".
    return first_only + both == FIRST and second_only + both == SECOND

def only(values):
    # A question that asks for a number its data fix is answered only when
    # every split that fits gives that same number.
    if len(set(values)) != 1:
        fail("the splits that fit give %d different answers" % len(set(values)))
    return values[0]

def solve(options):
    found = []
    for first_only, second_only, both in product(range(TOTAL + 1), repeat=3):
        neither = TOTAL - first_only - second_only - both
        if neither >= 0 and fits(first_only, second_only, both, neither):
            found.append(both)  # what the question asks about
    return match(options, min(found))  # max(found) for the largest, only(found) for a fixed number

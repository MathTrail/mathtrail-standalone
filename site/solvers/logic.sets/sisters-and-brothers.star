# Two groups that overlap, with the number in both given: try every split of
# the children into those with a sister only, a brother only, both and
# neither, keep the splits the question allows, and read off how many have
# neither.
# Written from the template two-groups of logic.sets.

TOTAL = 24  # the children in the group
FIRST = 15  # with a sister, those with both included
SECOND = 11  # with a brother, those with both included
BOTH = 6  # with a sister and a brother

def fits(first_only, second_only, both, neither):
    return both == BOTH

def only(values):
    # A question that asks for a number its data fix is answered only when
    # every split that fits gives that same number.
    if len(set(values)) != 1:
        fail("the splits that fit give %d different answers" % len(set(values)))
    return values[0]

def solve(options):
    found = []
    for both in range(0, TOTAL + 1):
        first_only = FIRST - both
        second_only = SECOND - both
        neither = TOTAL - first_only - second_only - both
        if min(first_only, second_only, neither) >= 0 and fits(first_only, second_only, both, neither):
            found.append(neither)  # what the question asks about
    if len(found) == 0:
        fail("no split fits everything the question says")
    return match(options, only(found))

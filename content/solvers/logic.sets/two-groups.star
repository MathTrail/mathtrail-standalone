# For two groups that may overlap: try every number in both groups, work out
# from the sizes of the groups how many are in the first only, the second
# only and neither, keep the splits in which nobody is counted below zero and
# that agree with everything else the question says, and read the answer off
# them. A size the question gives is used as it is; one it does not give is
# None below and is tried at every value it could take. With both sizes
# given, a year group of hundreds is no slower than a class.
# From reference task sets-56-d3-1.

TOTAL = 20  # everyone in the question
FIRST = 12  # in the first group, those in both included; None when the question does not say
SECOND = 10  # in the second group, those in both included; None when the question does not say

def fits(first_only, second_only, both, neither):
    # Everything else the question says goes here, such as "everyone is in at
    # least one" (neither == 0) or "twice as many in both as in neither".
    return True

def only(values):
    # A question that asks for a number its data fix is answered only when
    # every split that fits gives that same number.
    if len(set(values)) != 1:
        fail("the splits that fit give %d different answers" % len(set(values)))
    return values[0]

def sizes(given, both):
    # The sizes a group can have: the one the question gives, or every size
    # from the number in both up to everyone.
    if given != None:
        return [given]
    return range(both, TOTAL + 1)

def solve(options):
    found = []
    for both in range(0, TOTAL + 1):
        for first in sizes(FIRST, both):
            for second in sizes(SECOND, both):
                first_only = first - both
                second_only = second - both
                neither = TOTAL - first_only - second_only - both
                if min(first_only, second_only, neither) >= 0 and fits(first_only, second_only, both, neither):
                    found.append(both)  # what the question asks about
    if len(found) == 0:
        fail("no split fits everything the question says")
    return match(options, min(found))  # max(found) for the largest, only(found) for a fixed number

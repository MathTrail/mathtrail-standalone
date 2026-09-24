# For three groups that may overlap: try every number in all three groups and
# in both of each two, work out who is in one group only from the size of that
# group, keep the splits in which nobody is counted below zero and that agree
# with everything else the question says, and read the answer off them. A
# number the question gives is used as it is; one it does not give is None
# below and is tried at every value it could take. With the pairs given, a
# school of hundreds is no slower than a class.
# From reference task sets-56-d4-2.

TOTAL = 10  # everyone in the question
GROUPS = (7, 8, 9)  # the size of each group, A, B and C, overlaps included
PAIRS = (None, None, None)  # in both A and B, both A and C, both B and C, those in all three included; None when the question does not say
ALL_THREE = None  # in all three groups; None when the question does not say

def fits(all_three, ab, ac, bc, none):
    # Everything else the question says goes here, such as "everyone is in at
    # least one group" (none == 0). ab is in A and B but not in C, and so on
    # for ac and bc.
    return True

def only(values):
    # A question that asks for a number its data fix is answered only when
    # every split that fits gives that same number.
    if len(set(values)) != 1:
        fail("the splits that fit give %d different answers" % len(set(values)))
    return values[0]

def counts(given, least, most):
    # The numbers a count can be: the one the question gives, or every number
    # from least up to most.
    if given != None:
        return [given]
    return range(least, most + 1)

def solve(options):
    a, b, c = GROUPS
    found = []
    for all_three in counts(ALL_THREE, 0, min(GROUPS)):
        for in_ab in counts(PAIRS[0], all_three, min(a, b)):
            for in_ac in counts(PAIRS[1], all_three, min(a, c)):
                a_only = a - in_ab - in_ac + all_three
                if a_only < 0:
                    break  # a larger overlap only leaves fewer in A alone
                for in_bc in counts(PAIRS[2], all_three, min(b, c)):
                    ab = in_ab - all_three
                    ac = in_ac - all_three
                    bc = in_bc - all_three
                    b_only = b - ab - bc - all_three
                    c_only = c - ac - bc - all_three
                    if min(b_only, c_only) < 0:
                        break  # a larger overlap only leaves fewer in B or C alone
                    none = TOTAL - a_only - b_only - c_only - ab - ac - bc - all_three
                    if min(ab, ac, bc, none) >= 0 and fits(all_three, ab, ac, bc, none):
                        found.append(all_three)  # what the question asks about
    if len(found) == 0:
        fail("no split fits everything the question says")
    return match(options, min(found))  # max(found) for the largest, only(found) for a fixed number

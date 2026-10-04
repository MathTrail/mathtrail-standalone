# Three groups that may overlap, with nobody in all three and everybody in at
# least one: try every number in both of each two groups, work out who is in
# one group only, keep the splits in which nobody is counted below zero and
# nobody is left out, and read off how many are in exactly two groups.
# Written from the template three-groups of logic.sets.

TOTAL = 20  # the pupils in the question
GROUPS = (12, 10, 9)  # who like maths, reading and art, overlaps included
ALL_THREE = 0  # nobody likes all three

def only(values):
    # A question that asks for a number its data fix is answered only when
    # every split that fits gives that same number.
    if len(set(values)) != 1:
        fail("the splits that fit give %d different answers" % len(set(values)))
    return values[0]

def solve(options):
    a, b, c = GROUPS
    found = []
    for ab in range(0, min(a, b) + 1):
        for ac in range(0, min(a, c) + 1):
            for bc in range(0, min(b, c) + 1):
                a_only = a - ab - ac - ALL_THREE
                b_only = b - ab - bc - ALL_THREE
                c_only = c - ac - bc - ALL_THREE
                none = TOTAL - a_only - b_only - c_only - ab - ac - bc - ALL_THREE
                if min(a_only, b_only, c_only, none) >= 0 and none == 0:
                    found.append(ab + ac + bc)  # in exactly two groups
    if len(found) == 0:
        fail("no split fits everything the question says")
    return match(options, only(found))

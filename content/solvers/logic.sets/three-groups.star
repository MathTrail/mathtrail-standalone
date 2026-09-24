# For three groups that may overlap: try every size of the overlaps, those in
# all three groups and those in exactly two, work out who is in one group
# only from the size of that group, keep the splits in which nobody is
# counted below zero, and read the answer off them.
# From reference task sets-56-d4-2.

TOTAL = 10  # everyone in the question
GROUPS = (7, 8, 9)  # the size of each group, A, B and C, overlaps included

def fits(all_three, ab, ac, bc, none):
    # Everything else the question says goes here, such as an overlap: "5 are
    # in both A and B", counting those in all three, is ab + all_three == 5.
    return True

def only(values):
    # A question that asks for a number its data fix is answered only when
    # every split that fits gives that same number.
    if len(set(values)) != 1:
        fail("the splits that fit give %d different answers" % len(set(values)))
    return values[0]

def solve(options):
    a, b, c = GROUPS
    found = []
    for all_three, ab, ac, bc in product(range(min(GROUPS) + 1), repeat=4):
        # ab is in A and B but not in C, and so on for ac and bc.
        a_only = a - ab - ac - all_three
        b_only = b - ab - bc - all_three
        c_only = c - ac - bc - all_three
        none = TOTAL - a_only - b_only - c_only - ab - ac - bc - all_three
        if min(a_only, b_only, c_only, none) >= 0 and fits(all_three, ab, ac, bc, none):
            found.append(all_three)  # what the question asks about
    return match(options, min(found))  # max(found) for the largest, only(found) for a fixed number

# Two groups that overlap, with the number in both left open: try every
# number in both, keep the splits in which nobody is counted below zero, and
# the answer is the smallest number in both that fits. The example's note
# gives the largest as 16, so that is checked too.
# Written from the template two-groups of logic.sets.

TOTAL = 25  # the children in the group
FIRST = 18  # who like apples, those who like both included
SECOND = 16  # who like pears, those who like both included
LARGEST = 16  # the most who can like both, as the note says

def solve(options):
    found = []
    for both in range(0, TOTAL + 1):
        first_only = FIRST - both
        second_only = SECOND - both
        neither = TOTAL - first_only - second_only - both
        if min(first_only, second_only, neither) >= 0:
            found.append(both)  # what the question asks about
    if len(found) == 0:
        fail("no split fits everything the question says")
    if max(found) != LARGEST:
        fail("at most %d can be in both, and the note says %d" % (max(found), LARGEST))
    return match(options, min(found))  # the smallest number in both

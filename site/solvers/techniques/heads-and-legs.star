# Chickens and rabbits, so many heads and so many legs: try every number of
# rabbits and keep the ones whose legs add up. Exactly one must be left.
# Written from the template number-thought-of of arithmetic.tricks.

HEADS = 10
LEGS = 28
SHORT = 8  # legs missing when every head is taken for a chicken's, as the steps say

def solve(options):
    if LEGS - 2 * HEADS != SHORT:
        fail("with every head a chicken's, %d legs are missing" % (LEGS - 2 * HEADS))
    fitting = [rabbits for rabbits in range(HEADS + 1) if 4 * rabbits + 2 * (HEADS - rabbits) == LEGS]
    if len(fitting) != 1:
        fail("%d numbers of rabbits fit" % len(fitting))
    return match(options, fitting[0])

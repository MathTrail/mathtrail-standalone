# How many throws make a repeat certain: the answer is the smallest number of
# throws after which no way of throwing has every number different.
# Written from the template draws of pigeonhole.basic.

FACES = 6  # the numbers a die can show

def all_different(throws):
    return permutations(list(range(1, FACES + 1)), throws)  # the ways with no number twice

def solve(options):
    # The steps say six throws can still all differ.
    if len(all_different(FACES)) == 0:
        fail("%d throws can never all differ" % FACES)
    for throws in range(1, FACES + 2):
        if len(all_different(throws)) == 0:
            return match(options, throws)
    fail("no number of throws makes a repeat certain")

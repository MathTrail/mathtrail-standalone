# Two boxes hold marbles in the ratio 3 : 1, and after 10 marbles move from the
# first box to the second they hold as many each: try every size of one part,
# make each box that part times its number in the ratio, keep the sizes for
# which the boxes are equal after the move, and count all the marbles.
# Written from the template split of ratio.sharing.

RATIO = [3, 1]  # the first box's parts, then the second's
MOVED = 10  # marbles moved from the first box to the second
LARGEST = 1000  # the largest part worth trying

def fits(shares):
    return shares[0] >= MOVED and shares[0] - MOVED == shares[1] + MOVED

def asked(shares):
    return sum(shares)  # all the marbles, before or after the move

def solve(options):
    answers = set()
    for part in range(1, LARGEST + 1):
        shares = [part * number for number in RATIO]
        if fits(shares):
            answers.add(asked(shares))
    if len(answers) != 1:
        fail("the shares that fit give %d different answers" % len(answers))
    return match(options, list(answers)[0])

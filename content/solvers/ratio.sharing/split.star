# For a whole shared in a ratio: try every size of one part, make each share
# that part times its number in the ratio, keep the sizes for which the shares
# agree with everything the question says, and read the answer off them.
# From reference task ratio-56-d3-1.

RATIO = [3, 5, 7]  # the ratio of the shares, in the order the question names them
LARGEST = 1000  # the largest part worth trying

def fits(shares):
    # Everything the question says about the shares. Here: the three baskets
    # hold 90 walnuts together.
    return sum(shares) == 90

def asked(shares):
    # What the question asks about the shares. Here: how many more walnuts
    # basket C holds than basket A.
    return shares[2] - shares[0]

def solve(options):
    answers = set()
    for part in range(1, LARGEST + 1):
        shares = [part * number for number in RATIO]
        if fits(shares):
            answers.add(asked(shares))
    if len(answers) != 1:
        fail("the shares that fit give %d different answers" % len(answers))
    return match(options, list(answers)[0])

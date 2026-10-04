# A choir has 4 boys for every 7 girls, and 9 more girls than boys: try every
# size of one part, make each share that part times its number in the ratio,
# keep the sizes for which the girls are 9 more, and count the whole choir.
# Written from the template split of ratio.sharing.

RATIO = [4, 7]  # the boys' parts, then the girls'
LARGEST = 1000  # the largest part worth trying

def fits(shares):
    return shares[1] - shares[0] == 9

def asked(shares):
    return sum(shares)  # everyone in the choir

def solve(options):
    answers = set()
    for part in range(1, LARGEST + 1):
        shares = [part * number for number in RATIO]
        if fits(shares):
            answers.add(asked(shares))
    if len(answers) != 1:
        fail("the shares that fit give %d different answers" % len(answers))
    return match(options, list(answers)[0])

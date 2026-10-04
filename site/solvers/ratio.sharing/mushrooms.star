# Max and Lena pick 45 mushrooms, 4 of Max's for every 5 of Lena's: try every
# size of one part, make each share that part times its number in the ratio,
# keep the sizes for which the shares make 45, and read off Lena's share.
# Written from the template split of ratio.sharing.

RATIO = [4, 5]  # Max's parts, then Lena's
LARGEST = 1000  # the largest part worth trying

def fits(shares):
    return sum(shares) == 45

def asked(shares):
    return shares[1]  # Lena's mushrooms

def solve(options):
    answers = set()
    for part in range(1, LARGEST + 1):
        shares = [part * number for number in RATIO]
        if fits(shares):
            answers.add(asked(shares))
    if len(answers) != 1:
        fail("the shares that fit give %d different answers" % len(answers))
    return match(options, list(answers)[0])

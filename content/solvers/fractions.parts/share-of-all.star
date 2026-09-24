# For a share of everything that is made of shares of its parts: try every
# whole the question allows, keep those in which each part is a whole number,
# and read off the part the question asks about as a fraction in lowest
# terms. Every whole that fits has to give the same fraction.
# From reference task frac-56-d4-1.

SMALLEST = 1  # the wholes the question allows, when it gives limits
LARGEST = 1000

def asked(whole):
    # The part the question asks about, or None when a part is not a whole
    # number. Here: 2/5 of the books are novels, and half of the novels and
    # 1/6 of the other books have hard covers.
    if whole * 2 % 5 != 0:
        return None
    novels = whole * 2 // 5
    others = whole - novels
    if novels % 2 != 0 or others % 6 != 0:
        return None
    return novels // 2 + others // 6

def fraction(part, whole):
    # A fraction as the options write it: in lowest terms, top over bottom.
    common = gcd(part, whole)
    return "%d/%d" % (part // common, whole // common)

def solve(options):
    answers = set()
    for whole in range(SMALLEST, LARGEST + 1):
        part = asked(whole)
        if part != None:
            answers.add(fraction(part, whole))  # or part itself, when the question asks how many
    if len(answers) != 1:
        fail("the wholes that fit give %d different answers" % len(answers))
    return match(options, list(answers)[0])

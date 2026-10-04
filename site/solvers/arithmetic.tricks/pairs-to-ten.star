# 6 + 7 + 8 + 2 + 3 + 4: work it out the long way, term by term, and not by the
# pairs that make ten, which are what the child is to find. The steps pair the
# terms as 6 + 4, 7 + 3 and 8 + 2, so that is checked too.
# Written from the template long-way of arithmetic.tricks.

TERMS = [6, 7, 8, 2, 3, 4]
PAIRS = [(6, 4), (7, 3), (8, 2)]  # the pairs the steps make

def solve(options):
    paired = sorted([term for pair in PAIRS for term in pair])
    if paired != sorted(TERMS) or [a + b for a, b in PAIRS] != [10, 10, 10]:
        fail("the steps' pairs do not use every term once and make ten each")
    total = 0
    for term in TERMS:
        total += term
    return match(options, total)

PUPILS = 50
MARKS = [1, 2, 3, 4, 5]  # in maths and in reading alike

def can_share(pairs, most):
    # Every total the pupils on the pairs can come to with at most `most` on
    # each pair, built up one pair at a time.
    totals = set([0])
    for _ in pairs:
        totals = set([total + count for total in totals for count in range(most + 1) if total + count <= PUPILS])
    return PUPILS in totals

def solve(options):
    pairs = list(product(MARKS, MARKS))
    # The fewest pupils the fullest pair can have is the first `most` that lets
    # all of them be shared out with no more than that on any pair.
    most = 1
    while not can_share(pairs, most):
        most += 1
    return match(options, most)

DIVISOR = 7  # the numbers must differ by a number of its times table
WANTED = 2  # numbers that differ that way

def solve(options):
    # Two numbers are of one kind when they differ by a number of the DIVISOR
    # times table. Sorting a stretch of numbers that way gives the kinds, and
    # from then on only how many kinds there are matters.
    stretch = range(1, DIVISOR * DIVISOR + 1)
    kinds = set([tuple([other for other in stretch if (other - number) % DIVISOR == 0]) for number in stretch])
    # Every count of different numbers that is not yet enough, with fewer than
    # WANTED of each kind, built up one kind at a time; the answer is the first
    # count not among them.
    failed = set([0])
    for _ in kinds:
        failed = set([size + count for size in failed for count in range(WANTED)])
    numbers = 0
    while numbers in failed:
        numbers += 1
    return match(options, numbers)

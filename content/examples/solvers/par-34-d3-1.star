def solve(options):
    # Every sum five odd numbers up to 49 can make, one number at a time. The
    # question asks about the sum alone, so two choices with the same sum are
    # kept once — every one of the 118,755 ways to choose would cost a quarter
    # of the step budget to say the same thing.
    sums = set([0])
    for taken in range(5):
        sums = set([total + odd for total in sums for odd in range(1, 50, 2)])
    fitting = [n for n in [20, 30, 35, 40, 50] if n in sums]
    if len(fitting) != 1:
        fail("%d of the numbers fit" % len(fitting))
    return match(options, fitting[0])

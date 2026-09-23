def solve(options):
    fitting = [masha for masha in range(1, 50) if 7 * masha + 5 == 5 * (masha + 5)]
    if len(fitting) != 1:
        fail("%d ages fit the clue" % len(fitting))
    return match(options, fitting[0])

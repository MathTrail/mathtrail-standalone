def solve(options):
    fitting = [n for n in range(1001) if 5 * n - 7 == 23]
    if len(fitting) != 1:
        fail("%d whole numbers fit the clue" % len(fitting))
    return match(options, fitting[0])

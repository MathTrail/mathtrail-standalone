def solve(options):
    fitting = [n for n in range(1001) if (n + 7) * 2 == 30]
    if len(fitting) != 1:
        fail("%d whole numbers fit the clue" % len(fitting))
    return match(options, fitting[0])

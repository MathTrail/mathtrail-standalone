def solve(options):
    fitting = [n for n in range(1001) if n + 8 - 3 == 20]
    if len(fitting) != 1:
        fail("%d whole numbers fit the clue" % len(fitting))
    return match(options, fitting[0])

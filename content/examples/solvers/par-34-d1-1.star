def solve(options):
    fitting = [n for n in [6, 10, 13, 20, 28] if (14 + n) % 2 == 1]
    if len(fitting) != 1:
        fail("%d of the numbers fit" % len(fitting))
    return match(options, fitting[0])

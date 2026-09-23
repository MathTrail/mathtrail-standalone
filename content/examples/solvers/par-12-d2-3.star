def solve(options):
    even = range(0, 100, 2)
    fitting = [n for n in [7, 9, 12, 15, 17] if n in even]
    if len(fitting) != 1:
        fail("%d of the numbers fit" % len(fitting))
    return match(options, fitting[0])

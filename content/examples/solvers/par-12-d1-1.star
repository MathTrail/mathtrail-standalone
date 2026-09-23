def solve(options):
    fitting = [n for n in [7, 9, 12, 15, 21] if n % 2 == 0]
    if len(fitting) != 1:
        fail("%d of the numbers fit" % len(fitting))
    return match(options, fitting[0])

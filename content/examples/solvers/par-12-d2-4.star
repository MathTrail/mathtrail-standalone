def solve(options):
    fitting = [n for n in [12, 15, 18, 20, 25] if n % 2 == 0 and n % 5 == 0]
    if len(fitting) != 1:
        fail("%d of the numbers fit" % len(fitting))
    return match(options, fitting[0])

def solve(options):
    sums = set([sum(four) for four in combinations_with_replacement(range(1, 26, 2), 4)])
    fitting = [n for n in [17, 19, 21, 22, 23] if n in sums]
    if len(fitting) != 1:
        fail("%d of the numbers fit" % len(fitting))
    return match(options, fitting[0])

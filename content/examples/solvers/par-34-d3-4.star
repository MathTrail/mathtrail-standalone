def solve(options):
    ends = set([sum(jumps) for jumps in product([1, -1], repeat=9)])
    fitting = [n for n in [0, 2, 4, 7, 10] if n in ends]
    if len(fitting) != 1:
        fail("%d of the numbers fit" % len(fitting))
    return match(options, fitting[0])

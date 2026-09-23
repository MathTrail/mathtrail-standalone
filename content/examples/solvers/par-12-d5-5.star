def solve(options):
    sheets = [(2 * k - 1, 2 * k) for k in range(1, 26)]
    sums = set([sum([sheet[0] + sheet[1] for sheet in three]) for three in combinations(sheets, 3)])
    fitting = [n for n in [60, 77, 80, 90, 100] if n in sums]
    if len(fitting) != 1:
        fail("%d of the numbers fit" % len(fitting))
    return match(options, fitting[0])

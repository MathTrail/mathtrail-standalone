def solve(options):
    odd_days = range(1, 32, 2)
    fitting = ["%d March" % day for day in [2, 10, 15, 20, 28] if day in odd_days]
    if len(fitting) != 1:
        fail("%d of the days fit" % len(fitting))
    return match(options, fitting[0])

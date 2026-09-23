def solve(options):
    fitting = [8 + years for years in range(50) if 8 + years == 2 * (3 + years)]
    if len(fitting) != 1:
        fail("%d ages fit the clue" % len(fitting))
    return match(options, fitting[0])

def solve(options):
    fitting = [years for years in range(1, 100) if 29 + years == 3 * (5 + years)]
    if len(fitting) != 1:
        fail("%d ages fit the clue" % len(fitting))
    return match(options, fitting[0])

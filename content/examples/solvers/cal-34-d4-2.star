def solve(options):
    fitting = [years for years in range(100) if (5 + years) + (8 + years) + (11 + years) == 60]
    if len(fitting) != 1:
        fail("%d ages fit the clue" % len(fitting))
    return match(options, fitting[0])

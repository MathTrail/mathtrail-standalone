def solve(options):
    fitting = [twin + 5 for twin in range(30) if 2 * twin + (twin + 5) == 23]
    if len(fitting) != 1:
        fail("%d ages fit the clue" % len(fitting))
    return match(options, fitting[0])
